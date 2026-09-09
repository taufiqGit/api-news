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

type websiteRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewWebsiteRepository(pool *pgxpool.Pool) domainrepo.WebsiteRepository {
	return &websiteRepository{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *websiteRepository) Create(ctx context.Context, w entity.Website) (*entity.Website, error) {
	row, err := r.q.CreateWebsite(ctx, db.CreateWebsiteParams{
		Name:        w.Name,
		Slug:        w.Slug,
		Domain:      ptrToText(w.Domain),
		Description: ptrToText(w.Description),
		LogoUrl:     ptrToText(w.LogoURL),
		IsActive:    w.IsActive,
	})
	if err != nil {
		return nil, err
	}
	return toWebsiteEntity(row), nil
}

func (r *websiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Website, error) {
	row, err := r.q.GetWebsiteByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toWebsiteEntity(row), nil
}

func (r *websiteRepository) GetBySlug(ctx context.Context, slug string) (*entity.Website, error) {
	row, err := r.q.GetWebsiteBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toWebsiteEntity(row), nil
}

func (r *websiteRepository) GetByDomain(ctx context.Context, domain string) (*entity.Website, error) {
	row, err := r.q.GetWebsiteByDomain(ctx, pgtype.Text{String: domain, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toWebsiteEntity(row), nil
}

func (r *websiteRepository) List(ctx context.Context, limit, offset int) ([]entity.Website, error) {
	rows, err := r.q.ListWebsites(ctx, db.ListWebsitesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Website, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toWebsiteEntity(row))
	}
	return items, nil
}

func (r *websiteRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountWebsites(ctx)
}

func (r *websiteRepository) Update(ctx context.Context, id uuid.UUID, w entity.WebsiteUpdate) (*entity.Website, error) {
	row, err := r.q.UpdateWebsite(ctx, db.UpdateWebsiteParams{
		ID:          id,
		Name:        ptrToText(w.Name),
		Slug:        ptrToText(w.Slug),
		Domain:      ptrToText(w.Domain),
		Description: ptrToText(w.Description),
		LogoUrl:     ptrToText(w.LogoURL),
		IsActive:    ptrToBool(w.IsActive),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toWebsiteEntity(row), nil
}

func (r *websiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteWebsite(ctx, id)
}

func toWebsiteEntity(row db.Website) *entity.Website {
	return &entity.Website{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Domain:      textToPtr(row.Domain),
		Description: textToPtr(row.Description),
		LogoURL:     textToPtr(row.LogoUrl),
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
