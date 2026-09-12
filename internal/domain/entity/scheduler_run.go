package entity

import (
	"time"

	"github.com/google/uuid"
)

// Status scheduler run
const (
	RunStatusRunning = "running"
	RunStatusSuccess = "success"
	RunStatusPartial = "partial"
	RunStatusFailed  = "failed"
)

// SchedulerRun adalah catatan audit satu eksekusi pipeline scheduler
// untuk satu website (atau global bila WebsiteID nil).
type SchedulerRun struct {
	ID              uuid.UUID  `json:"id"`
	WebsiteID       *uuid.UUID `json:"website_id"`
	Status          string     `json:"status"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	ArticlesCreated int32      `json:"articles_created"`
	ArticlesSkipped int32      `json:"articles_skipped"`
	ArticlesFailed  int32      `json:"articles_failed"`
	Error           *string    `json:"error"`
	CreatedAt       time.Time  `json:"created_at"`
}
