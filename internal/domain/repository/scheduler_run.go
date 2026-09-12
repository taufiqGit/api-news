package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

type SchedulerRunRepository interface {
	// Create memulai run baru (status awal: running)
	Create(ctx context.Context, websiteID *uuid.UUID) (*entity.SchedulerRun, error)
	// Finish menutup run dengan status akhir & ringkasan hasil
	Finish(ctx context.Context, id uuid.UUID, status string, created, skipped, failed int, errMsg *string) (*entity.SchedulerRun, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.SchedulerRun, error)
	List(ctx context.Context, limit, offset int) ([]entity.SchedulerRun, error)
	Count(ctx context.Context) (int64, error)
}
