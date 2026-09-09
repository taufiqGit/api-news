package usecase

import (
	"context"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/domain/repository"
)

type CommentUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, req entity.CommentCreate) (*entity.Comment, error)
	ListByArticle(ctx context.Context, articleID uuid.UUID) ([]entity.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type commentUsecase struct {
	commentRepo repository.CommentRepository
	articleRepo repository.ArticleRepository
}

func NewCommentUsecase(commentRepo repository.CommentRepository, articleRepo repository.ArticleRepository) CommentUsecase {
	return &commentUsecase{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
	}
}

func (u *commentUsecase) Create(ctx context.Context, userID uuid.UUID, req entity.CommentCreate) (*entity.Comment, error) {
	// Validasi artikel ada
	if _, err := u.articleRepo.GetByID(ctx, req.ArticleID); err != nil {
		return nil, err
	}

	comment := entity.Comment{
		ArticleID:  req.ArticleID,
		UserID:     userID,
		ParentID:   req.ParentID,
		Content:    req.Content,
		IsApproved: true,
	}
	return u.commentRepo.Create(ctx, comment)
}

func (u *commentUsecase) ListByArticle(ctx context.Context, articleID uuid.UUID) ([]entity.Comment, error) {
	return u.commentRepo.ListByArticle(ctx, articleID, false)
}

func (u *commentUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.commentRepo.Delete(ctx, id)
}
