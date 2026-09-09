package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/domain/repository"
)

var (
	ErrWebsiteSlugTaken = errors.New("website slug already used")
)

type WebsiteUsecase interface {
	Create(ctx context.Context, req entity.WebsiteCreate) (*entity.Website, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Website, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Website, error)
	GetByDomain(ctx context.Context, domain string) (*entity.Website, error)
	List(ctx context.Context, page, limit int) ([]entity.Website, int64, error)
	Update(ctx context.Context, id uuid.UUID, req entity.WebsiteUpdate) (*entity.Website, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type websiteUsecase struct {
	websiteRepo repository.WebsiteRepository
}

func NewWebsiteUsecase(repo repository.WebsiteRepository) WebsiteUsecase {
	return &websiteUsecase{websiteRepo: repo}
}

func (u *websiteUsecase) Create(ctx context.Context, req entity.WebsiteCreate) (*entity.Website, error) {
	if _, err := u.websiteRepo.GetBySlug(ctx, req.Slug); err == nil {
		return nil, ErrWebsiteSlugTaken
	}
	website := entity.Website{
		Name:        req.Name,
		Slug:        strings.ToLower(req.Slug),
		Domain:      req.Domain,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		IsActive:    true,
	}
	return u.websiteRepo.Create(ctx, website)
}

func (u *websiteUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Website, error) {
	return u.websiteRepo.GetByID(ctx, id)
}

func (u *websiteUsecase) GetBySlug(ctx context.Context, slug string) (*entity.Website, error) {
	return u.websiteRepo.GetBySlug(ctx, slug)
}

func (u *websiteUsecase) GetByDomain(ctx context.Context, domain string) (*entity.Website, error) {
	return u.websiteRepo.GetByDomain(ctx, domain)
}

func (u *websiteUsecase) List(ctx context.Context, page, limit int) ([]entity.Website, int64, error) {
	offset := (page - 1) * limit
	items, err := u.websiteRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := u.websiteRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (u *websiteUsecase) Update(ctx context.Context, id uuid.UUID, req entity.WebsiteUpdate) (*entity.Website, error) {
	return u.websiteRepo.Update(ctx, id, req)
}

func (u *websiteUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.websiteRepo.Delete(ctx, id)
}
