package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/taufiqgit/news-api/internal/database/db"
	"github.com/taufiqgit/news-api/internal/domain/entity"
	domainrepo "github.com/taufiqgit/news-api/internal/domain/repository"
)

type categoryRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewCategoryRepository(pool *pgxpool.Pool) domainrepo.CategoryRepository {
	return &categoryRepository{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *categoryRepository) Create(ctx context.Context, c entity.Category) (*entity.Category, error) {
	row, err := r.q.CreateCategory(ctx, db.CreateCategoryParams{
		Name:        c.Name,
		Slug:        c.Slug,
		Description: ptrToText(c.Description),
		ParentID:    ptrToUUID(c.ParentID),
	})
	if err != nil {
		return nil, err
	}
	return toCategoryEntity(row), nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	row, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toCategoryEntity(row), nil
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	row, err := r.q.GetCategoryBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toCategoryEntity(row), nil
}

func (r *categoryRepository) List(ctx context.Context, limit, offset int) ([]entity.Category, error) {
	rows, err := r.q.ListCategoriesPaginated(ctx, db.ListCategoriesPaginatedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Category, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toCategoryEntity(row))
	}
	return items, nil
}

func (r *categoryRepository) Update(ctx context.Context, id uuid.UUID, c entity.CategoryUpdate) (*entity.Category, error) {
	row, err := r.q.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:          id,
		Name:        ptrToText(c.Name),
		Slug:        ptrToText(c.Slug),
		Description: ptrToText(c.Description),
		ParentID:    ptrToUUID(c.ParentID),
		IsActive:    ptrToBool(c.IsActive),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toCategoryEntity(row), nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteCategory(ctx, id)
}

func (r *categoryRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountCategories(ctx)
}

func toCategoryEntity(row db.Category) *entity.Category {
	return &entity.Category{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: textToPtr(row.Description),
		ParentID:    uuidToPtr(row.ParentID),
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
