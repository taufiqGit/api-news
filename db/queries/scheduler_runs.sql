-- name: CreateSchedulerRun :one
INSERT INTO scheduler_runs (website_id)
VALUES ($1)
RETURNING *;

-- name: FinishSchedulerRun :one
UPDATE scheduler_runs
SET status           = $2,
    finished_at      = NOW(),
    articles_created = $3,
    articles_skipped = $4,
    articles_failed  = $5,
    error            = $6
WHERE id = $1
RETURNING *;

-- name: GetSchedulerRunByID :one
SELECT * FROM scheduler_runs WHERE id = $1;

-- name: ListSchedulerRuns :many
SELECT * FROM scheduler_runs
ORDER BY started_at DESC
LIMIT $1 OFFSET $2;

-- name: CountSchedulerRuns :one
SELECT COUNT(*) FROM scheduler_runs;