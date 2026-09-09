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
	ErrSlugTaken     = errors.New("slug already used")
	ErrInvalidStatus = errors.New("invalid status")
)

// Status artikel yang valid
var validStatuses = map[string]bool{
	"draft":     true,
	"published": true,
	"archived":  true,
}

type CategoryUsecase interface {
	Create(ctx context.Context, req entity.CategoryCreate) (*entity.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Category, error)
	List(ctx context.Context, page, limit int) ([]entity.Category, int64, error)
	Update(ctx context.Context, id uuid.UUID, req entity.CategoryUpdate) (*entity.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryUsecase struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryUsecase(repo repository.CategoryRepository) CategoryUsecase {
	return &categoryUsecase{categoryRepo: repo}
}

func (u *categoryUsecase) Create(ctx context.Context, req entity.CategoryCreate) (*entity.Category, error) {
	if _, err := u.categoryRepo.GetBySlug(ctx, req.Slug); err == nil {
		return nil, ErrSlugTaken
	}

	category := entity.Category{
		Name:        req.Name,
		Slug:        strings.ToLower(req.Slug),
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    true,
	}
	return u.categoryRepo.Create(ctx, category)
}

func (u *categoryUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	return u.categoryRepo.GetByID(ctx, id)
}

func (u *categoryUsecase) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	return u.categoryRepo.GetBySlug(ctx, slug)
}

func (u *categoryUsecase) List(ctx context.Context, page, limit int) ([]entity.Category, int64, error) {
	offset := (page - 1) * limit
	items, err := u.categoryRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := u.categoryRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (u *categoryUsecase) Update(ctx context.Context, id uuid.UUID, req entity.CategoryUpdate) (*entity.Category, error) {
	return u.categoryRepo.Update(ctx, id, req)
}

func (u *categoryUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.categoryRepo.Delete(ctx, id)
}
