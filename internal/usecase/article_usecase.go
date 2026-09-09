package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/domain/repository"
)

type ArticleUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, req entity.ArticleCreate) (*entity.Article, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Article, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Article, error)
	List(ctx context.Context, q entity.ArticleQuery) ([]entity.Article, int64, error)
	ListPublished(ctx context.Context, page, limit int, websiteID *uuid.UUID) ([]entity.Article, error)
	Update(ctx context.Context, id uuid.UUID, req entity.ArticleUpdate) (*entity.Article, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Article, error)
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementView(ctx context.Context, id uuid.UUID) error
}

type articleUsecase struct {
	articleRepo repository.ArticleRepository
	tagUsecase  TagUsecase
}

func NewArticleUsecase(articleRepo repository.ArticleRepository, tagUsecase TagUsecase) ArticleUsecase {
	return &articleUsecase{
		articleRepo: articleRepo,
		tagUsecase:  tagUsecase,
	}
}

func (u *articleUsecase) Create(ctx context.Context, userID uuid.UUID, req entity.ArticleCreate) (*entity.Article, error) {
	status := req.Status
	if status == "" {
		status = "draft"
	}
	if !validStatuses[status] {
		return nil, ErrInvalidStatus
	}

	var publishedAt *time.Time
	if status == "published" {
		t := time.Now()
		publishedAt = &t
		if req.PublishedAt != nil {
			publishedAt = req.PublishedAt
		}
	}

	article := entity.Article{
		Title:       req.Title,
		Slug:        strings.ToLower(req.Slug),
		Excerpt:     req.Excerpt,
		Content:     req.Content,
		CoverImage:  req.CoverImage,
		Status:      status,
		PublishedAt: publishedAt,
		CreatedBy:   userID,
		CategoryID:  req.CategoryID,
		WebsiteID:   req.WebsiteID,
	}

	created, err := u.articleRepo.Create(ctx, article)
	if err != nil {
		return nil, err
	}

	// Attach tags jika ada
	if len(req.TagIDs) > 0 {
		if err := u.articleRepo.AttachTags(ctx, created.ID, req.TagIDs); err != nil {
			return nil, err
		}
	}

	return u.GetByID(ctx, created.ID)
}

func (u *articleUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Article, error) {
	article, err := u.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tags, err := u.articleRepo.ListArticleTags(ctx, article.ID)
	if err != nil {
		return nil, err
	}
	article.Tags = tags
	return article, nil
}

func (u *articleUsecase) GetBySlug(ctx context.Context, slug string) (*entity.Article, error) {
	article, err := u.articleRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	tags, err := u.articleRepo.ListArticleTags(ctx, article.ID)
	if err != nil {
		return nil, err
	}
	article.Tags = tags
	return article, nil
}

func (u *articleUsecase) List(ctx context.Context, q entity.ArticleQuery) ([]entity.Article, int64, error) {
	items, err := u.articleRepo.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	total, err := u.articleRepo.Count(ctx, q.Status, q.WebsiteID)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (u *articleUsecase) ListPublished(ctx context.Context, page, limit int, websiteID *uuid.UUID) ([]entity.Article, error) {
	offset := (page - 1) * limit
	return u.articleRepo.ListPublished(ctx, limit, offset, websiteID)
}

func (u *articleUsecase) Update(ctx context.Context, id uuid.UUID, req entity.ArticleUpdate) (*entity.Article, error) {
	if req.Status != nil && !validStatuses[*req.Status] {
		return nil, ErrInvalidStatus
	}

	updated, err := u.articleRepo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Replace tags jika disediakan
	if req.TagIDs != nil {
		if err := u.articleRepo.ReplaceTags(ctx, id, req.TagIDs); err != nil {
			return nil, err
		}
	}

	return u.GetByID(ctx, updated.ID)
}

func (u *articleUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Article, error) {
	if !validStatuses[status] {
		return nil, ErrInvalidStatus
	}
	return u.articleRepo.UpdateStatus(ctx, id, status)
}

func (u *articleUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.articleRepo.Delete(ctx, id)
}

func (u *articleUsecase) IncrementView(ctx context.Context, id uuid.UUID) error {
	return u.articleRepo.IncrementView(ctx, id)
}
