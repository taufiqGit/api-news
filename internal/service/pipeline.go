package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/ai"
	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/domain/repository"
	"github.com/taufiqgit/news-api/internal/source"
	"github.com/taufiqgit/news-api/internal/storage"
)

// SystemUserID adalah UUID user "System" dari seed reference data.
var SystemUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// Pipeline adalah orkestrator alur satu artikel dari sumber hingga tersimpan.
type Pipeline struct {
	sources      *source.Registry
	rewriter     ai.Rewriter
	articleRepo  repository.ArticleRepository
	tagRepo      repository.TagRepository
	categoryRepo repository.CategoryRepository
	runRepo      repository.SchedulerRunRepository
	storage      storage.Storage
	imageFolder  string
	httpClient   *http.Client
	logger       *slog.Logger

	itemsPerWebsite int
	defaultStatus   string
}

// NewPipeline membangun Pipeline dengan semua dependency.
func NewPipeline(
	sources *source.Registry,
	rewriter ai.Rewriter,
	articleRepo repository.ArticleRepository,
	tagRepo repository.TagRepository,
	categoryRepo repository.CategoryRepository,
	runRepo repository.SchedulerRunRepository,
	st storage.Storage,
	logger *slog.Logger,
) *Pipeline {
	if logger == nil {
		logger = slog.Default()
	}
	return &Pipeline{
		sources:         sources,
		rewriter:        rewriter,
		articleRepo:     articleRepo,
		tagRepo:         tagRepo,
		categoryRepo:    categoryRepo,
		runRepo:         runRepo,
		storage:         st,
		imageFolder:     "articles",
		httpClient:      &http.Client{Timeout: 20 * time.Second},
		logger:          logger,
		itemsPerWebsite: 5,
		defaultStatus:   "published",
	}
}

// SetItemsPerWebsite mengatur jumlah item berita per website (default 5).
func (p *Pipeline) SetItemsPerWebsite(n int) {
	if n > 0 {
		p.itemsPerWebsite = n
	}
}

// SetDefaultStatus mengatur status default artikel (default "published").
func (p *Pipeline) SetDefaultStatus(status string) {
	if status != "" {
		p.defaultStatus = status
	}
}

// ProcessResult merangkum hasil satu eksekusi pipeline per website.
type ProcessResult struct {
	Created int
	Skipped int
	Failed  int
}

// ProcessWebsite menjalankan alur lengkap untuk satu website:
// fetch → dedup → rewrite → download/watermark/upload gambar → simpan.
// Selalu menyimpan catatan audit scheduler_runs.
func (p *Pipeline) ProcessWebsite(ctx context.Context, website entity.Website) (ProcessResult, error) {
	res := ProcessResult{}

	websiteID := website.ID
	run, err := p.runRepo.Create(ctx, &websiteID)
	if err != nil {
		return res, fmt.Errorf("service: create scheduler run: %w", err)
	}

	var runErr error
	finishStatus := entity.RunStatusSuccess
	defer func() {
		var errMsg *string
		if runErr != nil {
			errMsg = ptr(runErr.Error())
			if res.Created == 0 {
				finishStatus = entity.RunStatusFailed
			} else {
				finishStatus = entity.RunStatusPartial
			}
		}
		if _, ferr := p.runRepo.Finish(ctx, run.ID, finishStatus, res.Created, res.Skipped, res.Failed, errMsg); ferr != nil {
			p.logger.Error("finish scheduler run", "run_id", run.ID, "error", ferr)
		}
	}()

	items, err := p.fetchAndDedup(ctx, website)
	if err != nil {
		runErr = err
		return res, err
	}
	res.Skipped += items.skipped

	for _, it := range items.items {
		if err := p.processItem(ctx, website, it); err != nil {
			p.logger.Error("process item", "url", it.URL, "error", err)
			res.Failed++
			runErr = err
			continue
		}
		res.Created++
	}

	return res, runErr
}

// fetchedItems adalah hasil fetch + dedup terhadap satu website.
type fetchedItems struct {
	items   []entity.SourceItem
	skipped int
}

// fetchAndDedup mengambil kandidat dari semua sumber lalu membuang yang sudah ada.
func (p *Pipeline) fetchAndDedup(ctx context.Context, website entity.Website) (fetchedItems, error) {
	var out fetchedItems

	sources := p.selectSources()
	if len(sources) == 0 {
		return out, fmt.Errorf("service: no news sources configured")
	}

	// Kumpulkan kandidat dari semua sumber (dengan limit per website).
	// Topic default kosong (kategori bebas).
	var candidates []entity.SourceItem
	opts := source.SourceOptions{Limit: p.itemsPerWebsite}
	for _, src := range sources {
		items, err := src.Fetch(ctx, opts)
		if err != nil {
			p.logger.Warn("fetch source", "source", src.Name(), "error", err)
			continue
		}
		candidates = append(candidates, items...)
	}

	// Dedup: cek source_url yang sudah ada di DB.
	urls := make([]string, 0, len(candidates))
	for _, it := range candidates {
		if it.URL != "" {
			urls = append(urls, it.URL)
		}
	}
	existing := map[string]bool{}
	if len(urls) > 0 {
		existURLs, err := p.articleRepo.GetExistingSourceURLs(ctx, urls)
		if err != nil {
			return out, fmt.Errorf("service: dedup lookup: %w", err)
		}
		for _, u := range existURLs {
			existing[u] = true
		}
	}

	// Dedup antar-kandidat (sumber bisa overlap) + skip yang sudah ada.
	seen := map[string]bool{}
	for _, it := range candidates {
		key := it.URL
		if key == "" {
			key = sourceHash(it.Title, it.Content)
		}
		if seen[key] || existing[key] {
			out.skipped++
			continue
		}
		seen[key] = true
		out.items = append(out.items, it)
	}

	return out, nil
}

// selectSources mengembalikan semua sumber terdaftar di registry.
func (p *Pipeline) selectSources() []source.NewsSource {
	names := p.sources.List()
	srcs := make([]source.NewsSource, 0, len(names))
	for _, n := range names {
		if s, err := p.sources.Get(n); err == nil {
			srcs = append(srcs, s)
		}
	}
	return srcs
}

// processItem memproses satu item sumber menjadi artikel tersimpan.
func (p *Pipeline) processItem(ctx context.Context, website entity.Website, it entity.SourceItem) error {
	// 1) Rewrite via AI.
	rewritten, err := p.rewriter.Rewrite(ctx, entity.RewriteRequest{
		Title:       it.Title,
		Content:     firstNonEmpty(it.Content, it.Excerpt),
		SourceName:  it.SourceName,
		SourceURL:   it.URL,
		Language:    "en",
		WebsiteName: website.Name,
	})
	if err != nil {
		return fmt.Errorf("service: rewrite: %w", err)
	}

	// 2) Resolve kategori (opsional, kategori bebas).
	var categoryID *uuid.UUID
	if rewritten.Category != nil && *rewritten.Category != "" {
		if id := p.resolveCategory(ctx, *rewritten.Category); id != nil {
			categoryID = id
		}
	}

	// 3) Download & upload gambar (jika ada). Graceful degradation.
	var coverURL *string
	var imageCredit, imageSourceURL, imageLicense *string
	if it.ImageURL != "" {
		info, credit, err := p.downloadAndUpload(ctx, it.ImageURL, it.ImageCredit, website.Name)
		if err != nil {
			p.logger.Warn("image processing", "url", it.ImageURL, "error", err)
		} else {
			u := info.URL
			coverURL = &u
			imageCredit = credit
			imageSourceURL = &it.ImageURL
		}
	}

	// 4) Slug dedup.
	slug, err := p.uniqueSlug(ctx, rewritten.Slug, rewritten.Title)
	if err != nil {
		return fmt.Errorf("service: slug: %w", err)
	}

	// 5) Simpan artikel.
	now := time.Now()
	sourceType := it.SourceType
	if sourceType == "" {
		sourceType = source.TypeRSS
	}
	sourceName := it.SourceName
	hash := sourceHash(it.Title, it.Content)
	status := p.defaultStatus
	var publishedAt *time.Time
	if status == "published" {
		publishedAt = &now
	}
	article := entity.Article{
		Title:          rewritten.Title,
		Slug:           slug,
		Excerpt:        strPtr(rewritten.Excerpt),
		Content:        rewritten.Content,
		CoverImage:     coverURL,
		Status:         status,
		PublishedAt:    publishedAt,
		CreatedBy:      SystemUserID,
		CategoryID:     categoryID,
		WebsiteID:      &website.ID,
		SourceURL:      strPtr(it.URL),
		SourceType:     strPtr(sourceType),
		SourceName:     strPtr(sourceName),
		SourceHash:     strPtr(hash),
		IsAiGenerated:  true,
		ImageCredit:    imageCredit,
		ImageSourceURL: imageSourceURL,
		ImageLicense:   imageLicense,
	}

	created, err := p.articleRepo.Create(ctx, article)
	if err != nil {
		return fmt.Errorf("service: create article: %w", err)
	}

	// 6) Attach tags (jika AI menyarankan).
	if len(rewritten.Tags) > 0 {
		ids, err := p.ensureTags(ctx, rewritten.Tags)
		if err != nil {
			p.logger.Warn("attach tags", "article_id", created.ID, "error", err)
		} else if len(ids) > 0 {
			if err := p.articleRepo.AttachTags(ctx, created.ID, ids); err != nil {
				p.logger.Warn("attach tags", "article_id", created.ID, "error", err)
			}
		}
	}

	return nil
}

// resolveCategory mencari kategori yang cocok (by slug atau nama, case-insensitive).
func (p *Pipeline) resolveCategory(ctx context.Context, name string) *uuid.UUID {
	slug := cleanSlug(name)
	if cat, err := p.categoryRepo.GetBySlug(ctx, slug); err == nil {
		return &cat.ID
	}
	// fallback: cari by name dari list (kategori seed relatif sedikit).
	cats, err := p.categoryRepo.List(ctx, 100, 0)
	if err != nil {
		return nil
	}
	for _, c := range cats {
		if strings.EqualFold(c.Name, name) || strings.EqualFold(c.Slug, slug) {
			id := c.ID
			return &id
		}
	}
	return nil
}

// ensureTags memastikan tag ada lalu mengembalikan ID-nya.
func (p *Pipeline) ensureTags(ctx context.Context, names []string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		slug := slugify(name)
		if seen[slug] {
			continue
		}
		seen[slug] = true
		tag, err := p.tagRepo.GetOrCreate(ctx, name, slug)
		if err != nil {
			return nil, err
		}
		ids = append(ids, tag.ID)
	}
	return ids, nil
}

// slugify menormalkan string menjadi slug sederhana (tanpa unicode).
func slugify(s string) string {
	return cleanSlug(s)
}

// --- helpers ---

func sourceHash(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = io.WriteString(h, p)
		_, _ = io.WriteString(h, "\x00")
	}
	return hex.EncodeToString(h.Sum(nil))
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptr(s string) *string {
	return &s
}
