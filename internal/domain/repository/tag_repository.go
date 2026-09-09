package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

type TagRepository interface {
	Create(ctx context.Context, t entity.Tag) (*entity.Tag, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Tag, error)
	GetOrCreate(ctx context.Context, name, slug string) (*entity.Tag, error)
	List(ctx context.Context) ([]entity.Tag, error)
	Update(ctx context.Context, id uuid.UUID, t entity.TagUpdate) (*entity.Tag, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
