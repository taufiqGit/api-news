package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/taufiqgit/news-api/internal/database/db"
	"github.com/taufiqgit/news-api/internal/domain/entity"
	domainrepo "github.com/taufiqgit/news-api/internal/domain/repository"
)

type articleRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewArticleRepository(pool *pgxpool.Pool) domainrepo.ArticleRepository {
	return &articleRepository{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *articleRepository) Create(ctx context.Context, a entity.Article) (*entity.Article, error) {
	row, err := r.q.CreateArticle(ctx, db.CreateArticleParams{
		Title:          a.Title,
		Slug:           a.Slug,
		Excerpt:        ptrToText(a.Excerpt),
		Content:        a.Content,
		CoverImage:     ptrToText(a.CoverImage),
		Status:         a.Status,
		PublishedAt:    ptrToTimestamptz(a.PublishedAt),
		CreatedBy:      a.CreatedBy,
		CategoryID:     ptrToUUID(a.CategoryID),
		WebsiteID:      ptrToUUID(a.WebsiteID),
		SourceUrl:      ptrToText(a.SourceURL),
		SourceType:     ptrToText(a.SourceType),
		SourceName:     ptrToText(a.SourceName),
		SourceHash:     ptrToText(a.SourceHash),
		IsAiGenerated:  a.IsAiGenerated,
		ImageCredit:    ptrToText(a.ImageCredit),
		ImageSourceUrl: ptrToText(a.ImageSourceURL),
		ImageLicense:   ptrToText(a.ImageLicense),
	})
	if err != nil {
		return nil, err
	}
	return toArticleEntity(row), nil
}

func (r *articleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Article, error) {
	row, err := r.q.GetArticleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toArticleWithJoinsID(row), nil
}

func (r *articleRepository) GetBySlug(ctx context.Context, slug string) (*entity.Article, error) {
	row, err := r.q.GetArticleBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toArticleWithJoinsSlug(row), nil
}

func (r *articleRepository) List(ctx context.Context, q entity.ArticleQuery) ([]entity.Article, error) {
	rows, err := r.q.ListArticles(ctx, db.ListArticlesParams{
		Column1:    q.Status,
		CategoryID: ptrToUUID(q.CategoryID),
		WebsiteID:  ptrToUUID(q.WebsiteID),
		Limit:      int32(q.Limit),
		Offset:     int32(q.Offset),
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Article, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toArticleWithJoinsList(row))
	}
	return items, nil
}

func (r *articleRepository) ListPublished(ctx context.Context, limit, offset int, websiteID *uuid.UUID) ([]entity.Article, error) {
	rows, err := r.q.ListPublishedArticles(ctx, db.ListPublishedArticlesParams{
		Limit:     int32(limit),
		Offset:    int32(offset),
		WebsiteID: ptrToUUID(websiteID),
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Article, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toArticleWithJoinsPublished(row))
	}
	return items, nil
}

func (r *articleRepository) ListByTag(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]entity.Article, error) {
	rows, err := r.q.ListArticlesByTag(ctx, db.ListArticlesByTagParams{
		TagID:  tagID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Article, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toArticleWithJoinsTag(row))
	}
	return items, nil
}

func (r *articleRepository) Update(ctx context.Context, id uuid.UUID, a entity.ArticleUpdate) (*entity.Article, error) {
	row, err := r.q.UpdateArticle(ctx, db.UpdateArticleParams{
		ID:         id,
		Title:      ptrToText(a.Title),
		Slug:       ptrToText(a.Slug),
		Excerpt:    ptrToText(a.Excerpt),
		Content:    ptrToText(a.Content),
		CoverImage: ptrToText(a.CoverImage),
		Status:     ptrToText(a.Status),
		CategoryID: ptrToUUID(a.CategoryID),
		WebsiteID:  ptrToUUID(a.WebsiteID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toArticleEntity(row), nil
}

func (r *articleRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Article, error) {
	row, err := r.q.UpdateArticleStatus(ctx, db.UpdateArticleStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toArticleEntity(row), nil
}

func (r *articleRepository) IncrementView(ctx context.Context, id uuid.UUID) error {
	_, err := r.q.IncrementViewCount(ctx, id)
	return err
}

func (r *articleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteArticle(ctx, id)
}

func (r *articleRepository) Count(ctx context.Context, status string, websiteID *uuid.UUID) (int64, error) {
	return r.q.CountArticles(ctx, db.CountArticlesParams{
		Column1:   status,
		WebsiteID: ptrToUUID(websiteID),
	})
}

// --- Dedup / metadata sumber (fitur scheduler) ---

func (r *articleRepository) GetBySourceURL(ctx context.Context, sourceURL string) (*entity.Article, error) {
	row, err := r.q.GetArticleBySourceURL(ctx, pgtype.Text{String: sourceURL, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toArticleEntity(row), nil
}

func (r *articleRepository) GetBySourceHash(ctx context.Context, sourceHash string) (*entity.Article, error) {
	row, err := r.q.GetArticleBySourceHash(ctx, pgtype.Text{String: sourceHash, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toArticleEntity(row), nil
}

func (r *articleRepository) GetExistingSourceURLs(ctx context.Context, urls []string) ([]string, error) {
	rows, err := r.q.GetExistingSourceURLs(ctx, urls)
	if err != nil {
		return nil, err
	}
	items := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Valid {
			items = append(items, row.String)
		}
	}
	return items, nil
}

// --- Tag helpers ---

func (r *articleRepository) AttachTags(ctx context.Context, articleID uuid.UUID, tagIDs []uuid.UUID) error {
	for _, tagID := range tagIDs {
		if err := r.q.CreateArticleTag(ctx, db.CreateArticleTagParams{
			ArticleID: articleID,
			TagID:     tagID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *articleRepository) ReplaceTags(ctx context.Context, articleID uuid.UUID, tagIDs []uuid.UUID) error {
	if err := r.q.DeleteArticleTags(ctx, articleID); err != nil {
		return err
	}
	return r.AttachTags(ctx, articleID, tagIDs)
}

func (r *articleRepository) ListArticleTags(ctx context.Context, articleID uuid.UUID) ([]entity.Tag, error) {
	rows, err := r.q.ListArticleTags(ctx, articleID)
	if err != nil {
		return nil, err
	}
	items := make([]entity.Tag, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toTagEntity(row))
	}
	return items, nil
}

// --- Mappers ---

// applySourceMeta mengisi metadata sumber berita & penanda AI pada entity artikel.
func applySourceMeta(a *entity.Article, sourceURL, sourceType, sourceName, sourceHash, imageCredit, imageSourceURL, imageLicense pgtype.Text, isAiGenerated bool) {
	a.SourceURL = textToPtr(sourceURL)
	a.SourceType = textToPtr(sourceType)
	a.SourceName = textToPtr(sourceName)
	a.SourceHash = textToPtr(sourceHash)
	a.IsAiGenerated = isAiGenerated
	a.ImageCredit = textToPtr(imageCredit)
	a.ImageSourceURL = textToPtr(imageSourceURL)
	a.ImageLicense = textToPtr(imageLicense)
}

func toArticleEntity(row db.Article) *entity.Article {
	article := &entity.Article{
		ID:          row.ID,
		Title:       row.Title,
		Slug:        row.Slug,
		Excerpt:     textToPtr(row.Excerpt),
		Content:     row.Content,
		CoverImage:  textToPtr(row.CoverImage),
		Status:      row.Status,
		ViewCount:   row.ViewCount,
		PublishedAt: timestamptzToPtr(row.PublishedAt),
		CreatedBy:   row.CreatedBy,
		CategoryID:  uuidToPtr(row.CategoryID),
		WebsiteID:   uuidToPtr(row.WebsiteID),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	applySourceMeta(article, row.SourceUrl, row.SourceType, row.SourceName, row.SourceHash,
		row.ImageCredit, row.ImageSourceUrl, row.ImageLicense, row.IsAiGenerated)
	return article
}

// toArticleWithJoinsID memetakan GetArticleByIDRow / GetArticleBySlugRow
func toArticleWithJoinsID(row db.GetArticleByIDRow) *entity.Article {
	article := &entity.Article{
		ID:           row.ID,
		Title:        row.Title,
		Slug:         row.Slug,
		Excerpt:      textToPtr(row.Excerpt),
		Content:      row.Content,
		CoverImage:   textToPtr(row.CoverImage),
		Status:       row.Status,
		ViewCount:    row.ViewCount,
		PublishedAt:  timestamptzToPtr(row.PublishedAt),
		CreatedBy:    row.CreatedBy,
		CategoryID:   uuidToPtr(row.CategoryID),
		WebsiteID:    uuidToPtr(row.WebsiteID),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		AuthorName:   &row.AuthorName,
		AuthorAvatar: textToPtr(row.AuthorAvatar),
		CategoryName: textToPtr(row.CategoryName),
		CategorySlug: textToPtr(row.CategorySlug),
		WebsiteName:  textToPtr(row.WebsiteName),
		WebsiteSlug:  textToPtr(row.WebsiteSlug),
	}
	applySourceMeta(article, row.SourceUrl, row.SourceType, row.SourceName, row.SourceHash,
		row.ImageCredit, row.ImageSourceUrl, row.ImageLicense, row.IsAiGenerated)
	return article
}

// toArticleWithJoinsSlug memetakan GetArticleBySlugRow
func toArticleWithJoinsSlug(row db.GetArticleBySlugRow) *entity.Article {
	article := &entity.Article{
		ID:           row.ID,
		Title:        row.Title,
		Slug:         row.Slug,
		Excerpt:      textToPtr(row.Excerpt),
		Content:      row.Content,
		CoverImage:   textToPtr(row.CoverImage),
		Status:       row.Status,
		ViewCount:    row.ViewCount,
		PublishedAt:  timestamptzToPtr(row.PublishedAt),
		CreatedBy:    row.CreatedBy,
		CategoryID:   uuidToPtr(row.CategoryID),
		WebsiteID:    uuidToPtr(row.WebsiteID),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		AuthorName:   &row.AuthorName,
		AuthorAvatar: textToPtr(row.AuthorAvatar),
		CategoryName: textToPtr(row.CategoryName),
		CategorySlug: textToPtr(row.CategorySlug),
		WebsiteName:  textToPtr(row.WebsiteName),
		WebsiteSlug:  textToPtr(row.WebsiteSlug),
	}
	applySourceMeta(article, row.SourceUrl, row.SourceType, row.SourceName, row.SourceHash,
		row.ImageCredit, row.ImageSourceUrl, row.ImageLicense, row.IsAiGenerated)
	return article
}

// toArticleWithJoinsList memetakan ListArticlesRow / ListPublishedArticlesRow
func toArticleWithJoinsList(row db.ListArticlesRow) *entity.Article {
	article := &entity.Article{
		ID:           row.ID,
		Title:        row.Title,
		Slug:         row.Slug,
		Excerpt:      textToPtr(row.Excerpt),
		Content:      row.Content,
		CoverImage:   textToPtr(row.CoverImage),
		Status:       row.Status,
		ViewCount:    row.ViewCount,
		PublishedAt:  timestamptzToPtr(row.PublishedAt),
		CreatedBy:    row.CreatedBy,
		CategoryID:   uuidToPtr(row.CategoryID),
		WebsiteID:    uuidToPtr(row.WebsiteID),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		AuthorName:   &row.AuthorName,
		AuthorAvatar: textToPtr(row.AuthorAvatar),
		CategoryName: textToPtr(row.CategoryName),
		CategorySlug: textToPtr(row.CategorySlug),
		WebsiteName:  textToPtr(row.WebsiteName),
		WebsiteSlug:  textToPtr(row.WebsiteSlug),
	}
	applySourceMeta(article, row.SourceUrl, row.SourceType, row.SourceName, row.SourceHash,
		row.ImageCredit, row.ImageSourceUrl, row.ImageLicense, row.IsAiGenerated)
	return article
}

// toArticleWithJoinsPublished memetakan ListPublishedArticlesRow
func toArticleWithJoinsPublished(row db.ListPublishedArticlesRow) *entity.Article {
	article := &entity.Article{
		ID:           row.ID,
		Title:        row.Title,
		Slug:         row.Slug,
		Excerpt:      textToPtr(row.Excerpt),
		Content:      row.Content,
		CoverImage:   textToPtr(row.CoverImage),
		Status:       row.Status,
		ViewCount:    row.ViewCount,
		PublishedAt:  timestamptzToPtr(row.PublishedAt),
		CreatedBy:    row.CreatedBy,
		CategoryID:   uuidToPtr(row.CategoryID),
		WebsiteID:    uuidToPtr(row.WebsiteID),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		AuthorName:   &row.AuthorName,
		AuthorAvatar: textToPtr(row.AuthorAvatar),
		CategoryName: textToPtr(row.CategoryName),
		CategorySlug: textToPtr(row.CategorySlug),
		WebsiteName:  textToPtr(row.WebsiteName),
		WebsiteSlug:  textToPtr(row.WebsiteSlug),
	}
	applySourceMeta(article, row.SourceUrl, row.SourceType, row.SourceName, row.SourceHash,
		row.ImageCredit, row.ImageSourceUrl, row.ImageLicense, row.IsAiGenerated)
	return article
}

// toArticleWithJoinsTag memetakan ListArticlesByTagRow
func toArticleWithJoinsTag(row db.ListArticlesByTagRow) *entity.Article {
	article := &entity.Article{
		ID:           row.ID,
		Title:        row.Title,
		Slug:         row.Slug,
		Excerpt:      textToPtr(row.Excerpt),
		Content:      row.Content,
		CoverImage:   textToPtr(row.CoverImage),
		Status:       row.Status,
		ViewCount:    row.ViewCount,
		PublishedAt:  timestamptzToPtr(row.PublishedAt),
		CreatedBy:    row.CreatedBy,
		CategoryID:   uuidToPtr(row.CategoryID),
		WebsiteID:    uuidToPtr(row.WebsiteID),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		AuthorName:   &row.AuthorName,
		AuthorAvatar: textToPtr(row.AuthorAvatar),
		CategoryName: textToPtr(row.CategoryName),
		CategorySlug: textToPtr(row.CategorySlug),
		WebsiteName:  textToPtr(row.WebsiteName),
		WebsiteSlug:  textToPtr(row.WebsiteSlug),
	}
	applySourceMeta(article, row.SourceUrl, row.SourceType, row.SourceName, row.SourceHash,
		row.ImageCredit, row.ImageSourceUrl, row.ImageLicense, row.IsAiGenerated)
	return article
}
