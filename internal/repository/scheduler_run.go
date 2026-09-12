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

type schedulerRunRepository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewSchedulerRunRepository(pool *pgxpool.Pool) domainrepo.SchedulerRunRepository {
	return &schedulerRunRepository{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *schedulerRunRepository) Create(ctx context.Context, websiteID *uuid.UUID) (*entity.SchedulerRun, error) {
	row, err := r.q.CreateSchedulerRun(ctx, ptrToUUID(websiteID))
	if err != nil {
		return nil, err
	}
	return toSchedulerRunEntity(row), nil
}

func (r *schedulerRunRepository) Finish(ctx context.Context, id uuid.UUID, status string, created, skipped, failed int, errMsg *string) (*entity.SchedulerRun, error) {
	row, err := r.q.FinishSchedulerRun(ctx, db.FinishSchedulerRunParams{
		ID:              id,
		Status:          status,
		ArticlesCreated: int32(created),
		ArticlesSkipped: int32(skipped),
		ArticlesFailed:  int32(failed),
		Error:           ptrToText(errMsg),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toSchedulerRunEntity(row), nil
}

func (r *schedulerRunRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.SchedulerRun, error) {
	row, err := r.q.GetSchedulerRunByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return toSchedulerRunEntity(row), nil
}

func (r *schedulerRunRepository) List(ctx context.Context, limit, offset int) ([]entity.SchedulerRun, error) {
	rows, err := r.q.ListSchedulerRuns(ctx, db.ListSchedulerRunsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	items := make([]entity.SchedulerRun, 0, len(rows))
	for _, row := range rows {
		items = append(items, *toSchedulerRunEntity(row))
	}
	return items, nil
}

func (r *schedulerRunRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountSchedulerRuns(ctx)
}

func toSchedulerRunEntity(row db.SchedulerRun) *entity.SchedulerRun {
	return &entity.SchedulerRun{
		ID:              row.ID,
		WebsiteID:       uuidToPtr(row.WebsiteID),
		Status:          row.Status,
		StartedAt:       row.StartedAt,
		FinishedAt:      timestamptzToPtr(row.FinishedAt),
		ArticlesCreated: row.ArticlesCreated,
		ArticlesSkipped: row.ArticlesSkipped,
		ArticlesFailed:  row.ArticlesFailed,
		Error:           textToPtr(row.Error),
		CreatedAt:       row.CreatedAt,
	}
}
