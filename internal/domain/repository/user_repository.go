package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, u entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	List(ctx context.Context, limit, offset int) ([]entity.User, error)
	Update(ctx context.Context, id uuid.UUID, u entity.UserUpdate) (*entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
