package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

type ArticleRepository interface {
	Create(ctx context.Context, a entity.Article) (*entity.Article, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Article, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Article, error)
	List(ctx context.Context, q entity.ArticleQuery) ([]entity.Article, error)
	ListPublished(ctx context.Context, limit, offset int, websiteID *uuid.UUID) ([]entity.Article, error)
	ListByTag(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]entity.Article, error)
	Update(ctx context.Context, id uuid.UUID, a entity.ArticleUpdate) (*entity.Article, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*entity.Article, error)
	IncrementView(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context, status string, websiteID *uuid.UUID) (int64, error)

	// Tag helpers
	AttachTags(ctx context.Context, articleID uuid.UUID, tagIDs []uuid.UUID) error
	ReplaceTags(ctx context.Context, articleID uuid.UUID, tagIDs []uuid.UUID) error
	ListArticleTags(ctx context.Context, articleID uuid.UUID) ([]entity.Tag, error)
}

type CommentRepository interface {
	Create(ctx context.Context, c entity.Comment) (*entity.Comment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error)
	ListByArticle(ctx context.Context, articleID uuid.UUID, includeUnapproved bool) ([]entity.Comment, error)
	Update(ctx context.Context, id uuid.UUID, content string, isApproved *bool) (*entity.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountByArticle(ctx context.Context, articleID uuid.UUID) (int64, error)
}
