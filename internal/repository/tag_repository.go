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

type tagRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewTagRepository(pool *pgxpool.Pool) domainrepo.TagRepository {
	return &tagRepository{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *tagRepository) Create(ctx context.Context, t entity.Tag) (*entity.Tag, error) {
	row, err := r.q.CreateTag(ctx, db.CreateTagParams{
		Name: t.Name,
		Slug: t.Slug,
	})
	if err != nil {
		return nil, err
	}
	return toTagEntity(row), nil
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	row, err := r.q.GetTagByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toTagEntity(row), nil
}

func (r *tagRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tag, error) {
	row, err := r.q.GetTagBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toTagEntity(row), nil
}

func (r *tagRepository) GetOrCreate(ctx context.Context, name, slug string) (*entity.Tag, error) {
	row, err := r.q.GetOrCreateTag(ctx, db.GetOrCreateTagParams{
		Name: name,
		Slug: slug,
	})
	if err != nil {
		return nil, err
	}
	return toTagEntity(row), nil
}

func (r *tagRepository) List(ctx context.Context) ([]entity.Tag, error) {
	rows, err := r.q.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]entity.Tag, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toTagEntity(row))
	}
	return items, nil
}

func (r *tagRepository) Update(ctx context.Context, id uuid.UUID, t entity.TagUpdate) (*entity.Tag, error) {
	name := t.Name
	slug := t.Slug
	if name == nil || slug == nil {
		existing, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if name == nil {
			name = &existing.Name
		}
		if slug == nil {
			slug = &existing.Slug
		}
	}
	row, err := r.q.UpdateTag(ctx, db.UpdateTagParams{
		ID:   id,
		Name: *name,
		Slug: *slug,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toTagEntity(row), nil
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteTag(ctx, id)
}

func toTagEntity(row db.Tag) *entity.Tag {
	return &entity.Tag{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
	}
}
