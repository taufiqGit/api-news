package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

type WebsiteRepository interface {
	Create(ctx context.Context, w entity.Website) (*entity.Website, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Website, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Website, error)
	GetByDomain(ctx context.Context, domain string) (*entity.Website, error)
	List(ctx context.Context, limit, offset int) ([]entity.Website, error)
	ListActive(ctx context.Context) ([]entity.Website, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, id uuid.UUID, w entity.WebsiteUpdate) (*entity.Website, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
