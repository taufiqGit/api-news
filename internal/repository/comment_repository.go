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

type commentRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewCommentRepository(pool *pgxpool.Pool) domainrepo.CommentRepository {
	return &commentRepository{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *commentRepository) Create(ctx context.Context, c entity.Comment) (*entity.Comment, error) {
	row, err := r.q.CreateComment(ctx, db.CreateCommentParams{
		ArticleID: c.ArticleID,
		UserID:    c.UserID,
		ParentID:  ptrToUUID(c.ParentID),
		Content:   c.Content,
	})
	if err != nil {
		return nil, err
	}
	return toCommentEntity(row), nil
}

func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error) {
	row, err := r.q.GetCommentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toCommentEntity(row), nil
}

func (r *commentRepository) ListByArticle(ctx context.Context, articleID uuid.UUID, includeUnapproved bool) ([]entity.Comment, error) {
	rows, err := r.q.ListArticleComments(ctx, db.ListArticleCommentsParams{
		ArticleID: articleID,
		Column2:   includeUnapproved,
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.Comment, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.Comment{
			ID:         row.ID,
			ArticleID:  row.ArticleID,
			UserID:     row.UserID,
			ParentID:   uuidToPtr(row.ParentID),
			Content:    row.Content,
			IsApproved: row.IsApproved,
			CreatedAt:  row.CreatedAt,
			UserName:   row.UserName,
			UserAvatar: textToPtr(row.UserAvatar),
		})
	}
	return items, nil
}

func (r *commentRepository) Update(ctx context.Context, id uuid.UUID, content string, isApproved *bool) (*entity.Comment, error) {
	row, err := r.q.UpdateComment(ctx, db.UpdateCommentParams{
		ID:         id,
		Content:    content,
		IsApproved: ptrToBool(isApproved),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toCommentEntity(row), nil
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteComment(ctx, id)
}

func (r *commentRepository) CountByArticle(ctx context.Context, articleID uuid.UUID) (int64, error) {
	return r.q.CountArticleComments(ctx, articleID)
}

func toCommentEntity(row db.Comment) *entity.Comment {
	return &entity.Comment{
		ID:         row.ID,
		ArticleID:  row.ArticleID,
		UserID:     row.UserID,
		ParentID:   uuidToPtr(row.ParentID),
		Content:    row.Content,
		IsApproved: row.IsApproved,
		CreatedAt:  row.CreatedAt,
	}
}
