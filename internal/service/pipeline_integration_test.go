package service

import (
	"context"
	"io"
	"log/slog"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/ai"
	"github.com/taufiqgit/news-api/internal/domain/entity"
	domainrepo "github.com/taufiqgit/news-api/internal/domain/repository"
	"github.com/taufiqgit/news-api/internal/repository"
	sqlrepo "github.com/taufiqgit/news-api/internal/repository"
	"github.com/taufiqgit/news-api/internal/source"
	"github.com/taufiqgit/news-api/internal/storage"
)

// Compile-time guarantees that all mocks satisfy their interfaces.
var (
	_ domainrepo.ArticleRepository      = (*mockArticleRepo)(nil)
	_ domainrepo.TagRepository          = (*mockTagRepo)(nil)
	_ domainrepo.CategoryRepository     = (*mockCategoryRepo)(nil)
	_ domainrepo.SchedulerRunRepository = (*mockRunRepo)(nil)
	_ ai.Rewriter                       = (*mockRewriter)(nil)
	_ source.NewsSource                 = (*mockSource)(nil)
	_ storage.Storage                   = (*mockStorage)(nil)
)

// --- mockSource (implements source.NewsSource) ---

type mockSource struct {
	name   string
	typ    string
	items  []entity.SourceItem
	err    error
	fetchN int
}

func (m *mockSource) Name() string { return m.name }
func (m *mockSource) Type() string {
	if m.typ == "" {
		return source.TypeRSS
	}
	return m.typ
}
func (m *mockSource) Fetch(ctx context.Context, opts source.SourceOptions) ([]entity.SourceItem, error) {
	m.fetchN++
	return m.items, m.err
}

// --- mockRewriter (implements ai.Rewriter) ---

type mockRewriter struct {
	result *entity.RewriteResult
	err    error
	calls  int
}

func (m *mockRewriter) Rewrite(ctx context.Context, req entity.RewriteRequest) (*entity.RewriteResult, error) {
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// --- mockArticleRepo (implements domainrepo.ArticleRepository) ---

type mockArticleRepo struct {
	existingURLs []string
	dedupErr     error
	createErr    error
	slugErr      error
	slugTaken    map[string]bool

	created  []entity.Article
	attached map[uuid.UUID][]uuid.UUID
}

func (m *mockArticleRepo) Create(ctx context.Context, a entity.Article) (*entity.Article, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	a.ID = uuid.New()
	m.created = append(m.created, a)
	return &a, nil
}
func (m *mockArticleRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Article, error) {
	return nil, repository.ErrNotFound
}
func (m *mockArticleRepo) GetBySlug(ctx context.Context, slug string) (*entity.Article, error) {
	if m.slugErr != nil {
		return nil, m.slugErr
	}
	if m.slugTaken[slug] {
		return &entity.Article{Slug: slug}, nil
	}
	return nil, repository.ErrNotFound
}
func (m *mockArticleRepo) List(ctx context.Context, q entity.ArticleQuery) ([]entity.Article, error) {
	return nil, nil
}
func (m *mockArticleRepo) ListPublished(ctx context.Context, limit, offset int, websiteID *uuid.UUID) ([]entity.Article, error) {
	return nil, nil
}
func (m *mockArticleRepo) ListByTag(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]entity.Article, error) {
	return nil, nil
}
func (m *mockArticleRepo) Update(ctx context.Context, id uuid.UUID, a entity.ArticleUpdate) (*entity.Article, error) {
	return nil, repository.ErrNotFound
}
func (m *mockArticleRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Article, error) {
	return nil, repository.ErrNotFound
}
func (m *mockArticleRepo) IncrementView(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockArticleRepo) Delete(ctx context.Context, id uuid.UUID) error        { return nil }
func (m *mockArticleRepo) Count(ctx context.Context, status string, websiteID *uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockArticleRepo) GetBySourceURL(ctx context.Context, sourceURL string) (*entity.Article, error) {
	return nil, repository.ErrNotFound
}
func (m *mockArticleRepo) GetBySourceHash(ctx context.Context, sourceHash string) (*entity.Article, error) {
	return nil, repository.ErrNotFound
}
func (m *mockArticleRepo) GetExistingSourceURLs(ctx context.Context, urls []string) ([]string, error) {
	if m.dedupErr != nil {
		return nil, m.dedupErr
	}
	return m.existingURLs, nil
}
func (m *mockArticleRepo) AttachTags(ctx context.Context, articleID uuid.UUID, tagIDs []uuid.UUID) error {
	if m.attached == nil {
		m.attached = map[uuid.UUID][]uuid.UUID{}
	}
	m.attached[articleID] = append(m.attached[articleID], tagIDs...)
	return nil
}
func (m *mockArticleRepo) ReplaceTags(ctx context.Context, articleID uuid.UUID, tagIDs []uuid.UUID) error {
	return nil
}
func (m *mockArticleRepo) ListArticleTags(ctx context.Context, articleID uuid.UUID) ([]entity.Tag, error) {
	return nil, nil
}

// --- mockTagRepo (implements domainrepo.TagRepository) ---

type tagCall struct{ name, slug string }

type mockTagRepo struct {
	err              error
	getOrCreateCalls []tagCall
}

func (m *mockTagRepo) Create(ctx context.Context, t entity.Tag) (*entity.Tag, error) {
	return &entity.Tag{ID: uuid.New(), Name: t.Name, Slug: t.Slug}, nil
}
func (m *mockTagRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	return nil, repository.ErrNotFound
}
func (m *mockTagRepo) GetBySlug(ctx context.Context, slug string) (*entity.Tag, error) {
	return nil, repository.ErrNotFound
}
func (m *mockTagRepo) GetOrCreate(ctx context.Context, name, slug string) (*entity.Tag, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.getOrCreateCalls = append(m.getOrCreateCalls, tagCall{name: name, slug: slug})
	return &entity.Tag{ID: uuid.New(), Name: name, Slug: slug}, nil
}
func (m *mockTagRepo) List(ctx context.Context) ([]entity.Tag, error) { return nil, nil }
func (m *mockTagRepo) Update(ctx context.Context, id uuid.UUID, t entity.TagUpdate) (*entity.Tag, error) {
	return nil, repository.ErrNotFound
}
func (m *mockTagRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

// --- mockCategoryRepo (implements domainrepo.CategoryRepository) ---

type mockCategoryRepo struct {
	bySlug map[string]*entity.Category
	list   []entity.Category
}

func (m *mockCategoryRepo) Create(ctx context.Context, c entity.Category) (*entity.Category, error) {
	return nil, repository.ErrNotFound
}
func (m *mockCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	return nil, repository.ErrNotFound
}
func (m *mockCategoryRepo) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	if m.bySlug != nil {
		if c, ok := m.bySlug[slug]; ok {
			return c, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (m *mockCategoryRepo) List(ctx context.Context, limit, offset int) ([]entity.Category, error) {
	return m.list, nil
}
func (m *mockCategoryRepo) Update(ctx context.Context, id uuid.UUID, c entity.CategoryUpdate) (*entity.Category, error) {
	return nil, repository.ErrNotFound
}
func (m *mockCategoryRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockCategoryRepo) Count(ctx context.Context) (int64, error)       { return 0, nil }

// --- mockRunRepo (implements domainrepo.SchedulerRunRepository) ---

type finishCall struct {
	status  string
	created int
	skipped int
	failed  int
}

type mockRunRepo struct {
	run    *entity.SchedulerRun
	finish []finishCall
}

func (m *mockRunRepo) Create(ctx context.Context, websiteID *uuid.UUID) (*entity.SchedulerRun, error) {
	m.run = &entity.SchedulerRun{
		ID:        uuid.New(),
		WebsiteID: websiteID,
		Status:    entity.RunStatusRunning,
	}
	return m.run, nil
}
func (m *mockRunRepo) Finish(ctx context.Context, id uuid.UUID, status string, created, skipped, failed int, errMsg *string) (*entity.SchedulerRun, error) {
	m.finish = append(m.finish, finishCall{status: status, created: created, skipped: skipped, failed: failed})
	return &entity.SchedulerRun{ID: id, Status: status}, nil
}
func (m *mockRunRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.SchedulerRun, error) {
	return nil, repository.ErrNotFound
}
func (m *mockRunRepo) List(ctx context.Context, limit, offset int) ([]entity.SchedulerRun, error) {
	return nil, nil
}
func (m *mockRunRepo) Count(ctx context.Context) (int64, error) { return 0, nil }

// --- mockStorage (implements storage.Storage) ---

type mockStorage struct {
	uploaded []string
}

func (m *mockStorage) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (*storage.FileInfo, error) {
	return nil, nil
}
func (m *mockStorage) UploadReader(ctx context.Context, r io.Reader, size int64, filename, contentType, folder string) (*storage.FileInfo, error) {
	m.uploaded = append(m.uploaded, folder+"/"+filename)
	return &storage.FileInfo{Key: folder + "/" + filename, URL: "http://test/" + filename, Size: size, ContentType: contentType}, nil
}
func (m *mockStorage) Delete(ctx context.Context, key string) error { return nil }
func (m *mockStorage) URL(key string) string                        { return "http://test/" + key }

// --- helper ---

func newTestPipeline(rw *mockRewriter, art *mockArticleRepo, tag *mockTagRepo, cat *mockCategoryRepo, run *mockRunRepo, st storage.Storage, srcs ...source.NewsSource) *Pipeline {
	reg := source.NewRegistry()
	for _, s := range srcs {
		reg.Add(s)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewPipeline(reg, rw, art, tag, cat, run, st, logger)
}

// --- Tests ---

func TestProcessWebsite_CreatesArticles(t *testing.T) {
	website := entity.Website{ID: uuid.New(), Name: "Example News", Slug: "example-news"}

	src := &mockSource{name: "mock-news", items: []entity.SourceItem{
		{Title: "Alpha", URL: "https://example.com/1", Content: "body-1", SourceName: "Example"},
		{Title: "Beta", URL: "https://example.com/2", Content: "body-2", SourceName: "Example"},
	}}
	rw := &mockRewriter{result: &entity.RewriteResult{
		Title: "Alpha Rewritten", Slug: "alpha-rewritten", Excerpt: "excerpt", Content: "rewritten body",
	}}
	art := &mockArticleRepo{}
	tag := &mockTagRepo{}
	cat := &mockCategoryRepo{}
	run := &mockRunRepo{}
	st := &mockStorage{}

	p := newTestPipeline(rw, art, tag, cat, run, st, src)

	res, err := p.ProcessWebsite(context.Background(), website)
	if err != nil {
		t.Fatalf("ProcessWebsite error: %v", err)
	}
	if res.Created != 2 || res.Skipped != 0 || res.Failed != 0 {
		t.Fatalf("result = %+v, want created=2 skipped=0 failed=0", res)
	}
	if len(art.created) != 2 {
		t.Fatalf("created articles = %d, want 2", len(art.created))
	}

	a := art.created[0]
	if !a.IsAiGenerated {
		t.Errorf("IsAiGenerated = %v, want true", a.IsAiGenerated)
	}
	if a.WebsiteID == nil || *a.WebsiteID != website.ID {
		t.Errorf("WebsiteID = %v, want %v", a.WebsiteID, website.ID)
	}
	if a.Status != "published" {
		t.Errorf("Status = %q, want published", a.Status)
	}
	if a.CreatedBy != SystemUserID {
		t.Errorf("CreatedBy = %v, want SystemUserID %v", a.CreatedBy, SystemUserID)
	}
	if a.SourceURL == nil || *a.SourceURL != "https://example.com/1" {
		t.Errorf("SourceURL = %v, want https://example.com/1", a.SourceURL)
	}

	if len(run.finish) != 1 {
		t.Fatalf("finish calls = %d, want 1", len(run.finish))
	}
	if f := run.finish[0]; f.status != entity.RunStatusSuccess || f.created != 2 || f.skipped != 0 || f.failed != 0 {
		t.Fatalf("finish = %+v, want success created=2", f)
	}
}

func TestProcessWebsite_DedupSkipsExisting(t *testing.T) {
	website := entity.Website{ID: uuid.New(), Name: "Example News", Slug: "example-news"}

	src := &mockSource{name: "mock-news", items: []entity.SourceItem{
		{Title: "Old", URL: "https://example.com/old", Content: "old", SourceName: "Example"},
		{Title: "New", URL: "https://example.com/new", Content: "new", SourceName: "Example"},
	}}
	rw := &mockRewriter{result: &entity.RewriteResult{Title: "New", Slug: "new", Content: "body"}}
	art := &mockArticleRepo{existingURLs: []string{"https://example.com/old"}}
	tag := &mockTagRepo{}
	cat := &mockCategoryRepo{}
	run := &mockRunRepo{}
	st := &mockStorage{}

	p := newTestPipeline(rw, art, tag, cat, run, st, src)

	res, err := p.ProcessWebsite(context.Background(), website)
	if err != nil {
		t.Fatalf("ProcessWebsite error: %v", err)
	}
	if res.Created != 1 || res.Skipped != 1 || res.Failed != 0 {
		t.Fatalf("result = %+v, want created=1 skipped=1 failed=0", res)
	}
	if len(art.created) != 1 {
		t.Fatalf("created articles = %d, want 1", len(art.created))
	}
	if *art.created[0].SourceURL != "https://example.com/new" {
		t.Fatalf("created SourceURL = %v, want new URL", *art.created[0].SourceURL)
	}
}

func TestProcessWebsite_RewriteError(t *testing.T) {
	website := entity.Website{ID: uuid.New(), Name: "Example News", Slug: "example-news"}

	src := &mockSource{name: "mock-news", items: []entity.SourceItem{
		{Title: "X", URL: "https://example.com/x", Content: "x", SourceName: "Example"},
	}}
	rw := &mockRewriter{err: io.ErrUnexpectedEOF}
	art := &mockArticleRepo{}
	tag := &mockTagRepo{}
	cat := &mockCategoryRepo{}
	run := &mockRunRepo{}
	st := &mockStorage{}

	p := newTestPipeline(rw, art, tag, cat, run, st, src)

	res, err := p.ProcessWebsite(context.Background(), website)
	if err == nil {
		t.Fatalf("ProcessWebsite harus error ketika rewrite gagal")
	}
	if res.Created != 0 || res.Skipped != 0 || res.Failed != 1 {
		t.Fatalf("result = %+v, want created=0 skipped=0 failed=1", res)
	}
	if len(art.created) != 0 {
		t.Fatalf("created articles = %d, want 0", len(art.created))
	}
	if len(run.finish) != 1 {
		t.Fatalf("finish calls = %d, want 1", len(run.finish))
	}
	if f := run.finish[0]; f.status != entity.RunStatusFailed || f.failed != 1 {
		t.Fatalf("finish = %+v, want failed", f)
	}
}

func TestProcessWebsite_NoSources(t *testing.T) {
	website := entity.Website{ID: uuid.New(), Name: "Example News", Slug: "example-news"}

	rw := &mockRewriter{}
	art := &mockArticleRepo{}
	tag := &mockTagRepo{}
	cat := &mockCategoryRepo{}
	run := &mockRunRepo{}
	st := &mockStorage{}

	// Tidak ada sumber terdaftar.
	p := newTestPipeline(rw, art, tag, cat, run, st)

	if _, err := p.ProcessWebsite(context.Background(), website); err == nil {
		t.Fatalf("ProcessWebsite harus error ketika tidak ada sumber")
	}
}

func TestProcessWebsite_DuplicateDBError(t *testing.T) {
	// Simulasi race: GetExistingSourceURLs tidak menemukan URL karena belum
	// di-commit, tapi INSERT langsung melanggar unique constraint
	// (idx_articles_source_url). Pipeline harus memperlakukannya sebagai skip.
	website := entity.Website{ID: uuid.New(), Name: "Example News", Slug: "example-news"}

	src := &mockSource{name: "mock-news", items: []entity.SourceItem{
		{Title: "Race", URL: "https://example.com/race", Content: "race body", SourceName: "X"},
	}}
	rw := &mockRewriter{result: &entity.RewriteResult{Title: "Race", Slug: "race-1", Content: "body"}}
	art := &mockArticleRepo{createErr: sqlrepo.ErrDuplicateArticle}
	tag := &mockTagRepo{}
	cat := &mockCategoryRepo{}
	run := &mockRunRepo{}
	st := &mockStorage{}

	p := newTestPipeline(rw, art, tag, cat, run, st, src)

	res, err := p.ProcessWebsite(context.Background(), website)
	if err != nil {
		t.Fatalf("ProcessWebsite harus tidak error pada duplicate (race)")
	}
	if res.Created != 0 || res.Skipped != 1 || res.Failed != 0 {
		t.Fatalf("result = %+v, want created=0 skipped=1 failed=0 (duplicate=race-skip)", res)
	}
	if len(art.created) != 0 {
		t.Fatalf("created articles = %d, want 0", len(art.created))
	}
	// Run harus success (bukan failed), karena ini dedup, bukan error.
	if len(run.finish) != 1 {
		t.Fatalf("finish calls = %d, want 1", len(run.finish))
	}
	if f := run.finish[0]; f.status != entity.RunStatusSuccess || f.skipped != 1 {
		t.Fatalf("finish = %+v, want success skipped=1", f)
	}
}

func TestProcessWebsite_AttachesTagsAndCategory(t *testing.T) {
	website := entity.Website{ID: uuid.New(), Name: "Example News", Slug: "example-news"}

	catID := uuid.New()
	src := &mockSource{name: "mock-news", items: []entity.SourceItem{
		{Title: "Tech story", URL: "https://example.com/tech", Content: "tech body", SourceName: "Example"},
	}}
	catName := "Technology"
	rw := &mockRewriter{result: &entity.RewriteResult{
		Title: "Tech story rewritten", Slug: "tech-story", Content: "body", Category: &catName, Tags: []string{"AI", "Future"},
	}}
	art := &mockArticleRepo{}
	tag := &mockTagRepo{}
	cat := &mockCategoryRepo{bySlug: map[string]*entity.Category{
		"technology": {ID: catID, Name: "Technology", Slug: "technology"},
	}}
	run := &mockRunRepo{}
	st := &mockStorage{}

	p := newTestPipeline(rw, art, tag, cat, run, st, src)

	res, err := p.ProcessWebsite(context.Background(), website)
	if err != nil {
		t.Fatalf("ProcessWebsite error: %v", err)
	}
	if res.Created != 1 {
		t.Fatalf("created = %d, want 1", res.Created)
	}
	if len(art.created) != 1 {
		t.Fatalf("created articles = %d, want 1", len(art.created))
	}
	a := art.created[0]
	if a.CategoryID == nil || *a.CategoryID != catID {
		t.Errorf("CategoryID = %v, want %v", a.CategoryID, catID)
	}
	if len(tag.getOrCreateCalls) != 2 {
		t.Fatalf("GetOrCreate calls = %d, want 2", len(tag.getOrCreateCalls))
	}
	if len(art.attached) != 1 {
		t.Fatalf("attached articles = %d, want 1", len(art.attached))
	}
	if got := art.attached[a.ID]; len(got) != 2 {
		t.Fatalf("attached tag ids = %d, want 2", len(got))
	}
}
