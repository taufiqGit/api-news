package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/domain/repository"
)

type TagUsecase interface {
	Create(ctx context.Context, req entity.TagCreate) (*entity.Tag, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
	List(ctx context.Context) ([]entity.Tag, error)
	Update(ctx context.Context, id uuid.UUID, req entity.TagUpdate) (*entity.Tag, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type tagUsecase struct {
	tagRepo repository.TagRepository
}

func NewTagUsecase(repo repository.TagRepository) TagUsecase {
	return &tagUsecase{tagRepo: repo}
}

func (u *tagUsecase) Create(ctx context.Context, req entity.TagCreate) (*entity.Tag, error) {
	if _, err := u.tagRepo.GetBySlug(ctx, req.Slug); err == nil {
		return nil, ErrSlugTaken
	}
	tag := entity.Tag{
		Name: req.Name,
		Slug: strings.ToLower(req.Slug),
	}
	return u.tagRepo.Create(ctx, tag)
}

func (u *tagUsecase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	return u.tagRepo.GetByID(ctx, id)
}

func (u *tagUsecase) List(ctx context.Context) ([]entity.Tag, error) {
	return u.tagRepo.List(ctx)
}

func (u *tagUsecase) Update(ctx context.Context, id uuid.UUID, req entity.TagUpdate) (*entity.Tag, error) {
	return u.tagRepo.Update(ctx, id, req)
}

func (u *tagUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.tagRepo.Delete(ctx, id)
}

// ensureTags memastikan tag dengan slug yang diberikan ada, lalu mengembalikan ID-nya.
func (u *tagUsecase) ensureTags(ctx context.Context, names []string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for _, name := range names {
		slug := slugify(name)
		tag, err := u.tagRepo.GetOrCreate(ctx, name, slug)
		if err != nil {
			return nil, err
		}
		ids = append(ids, tag.ID)
	}
	return ids, nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
