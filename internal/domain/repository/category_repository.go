package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

type CategoryRepository interface {
	Create(ctx context.Context, c entity.Category) (*entity.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Category, error)
	List(ctx context.Context, limit, offset int) ([]entity.Category, error)
	Update(ctx context.Context, id uuid.UUID, c entity.CategoryUpdate) (*entity.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}
