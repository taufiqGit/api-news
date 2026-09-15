package handler

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/taufiqgit/news-api/internal/domain/repository"
	"github.com/taufiqgit/news-api/internal/handler/response"
	"github.com/taufiqgit/news-api/internal/scheduler"
)

// SchedulerHandler menangani endpoint admin untuk scheduler berita.
type SchedulerHandler struct {
	scheduler *scheduler.Scheduler // boleh nil (scheduler disabled)
	runRepo   repository.SchedulerRunRepository
}

func NewSchedulerHandler(sched *scheduler.Scheduler, runRepo repository.SchedulerRunRepository) *SchedulerHandler {
	return &SchedulerHandler{scheduler: sched, runRepo: runRepo}
}

// Run POST /api/v1/admin/scheduler/run
//
//	@Summary		Trigger scheduler manual
//	@Description	Menjalankan satu siklus scheduler secara manual (async). Role: admin.
//	@Tags			Admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		202	{object}	response.Response	"Run dipicu"
//	@Failure		403	{object}	response.Response	"Forbidden"
//	@Failure		503	{object}	response.Response	"Scheduler disabled"
//	@Router			/admin/scheduler/run [post]
func (h *SchedulerHandler) Run(c *gin.Context) {
	if h.scheduler == nil {
		response.Error(c, http.StatusServiceUnavailable, "scheduler disabled", nil)
		return
	}

	// Jalankan asinkron agar request tidak terblokir oleh proses pipeline.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		_ = h.scheduler.RunNow(ctx)
	}()

	response.Success(c, http.StatusAccepted, "scheduler run triggered", gin.H{"status": "accepted"})
}

// ListRuns GET /api/v1/admin/scheduler/runs?page=1&limit=20
//
//	@Summary		Riwayat scheduler runs
//	@Description	Menampilkan daftar riwayat eksekusi scheduler. Role: admin.
//	@Tags			Admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int	false	"Nomor halaman"	default(1)
//	@Param			limit	query		int	false	"Jumlah per halaman"	default(20)
//	@Success		200	{object}	response.Response	"Daftar run dengan meta"
//	@Failure		403	{object}	response.Response	"Forbidden"
//	@Router			/admin/scheduler/runs [get]
func (h *SchedulerHandler) ListRuns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	items, err := h.runRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, "failed to list scheduler runs")
		return
	}
	total, err := h.runRepo.Count(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to count scheduler runs")
		return
	}

	meta := &response.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
	response.SuccessWithMeta(c, http.StatusOK, "scheduler runs fetched", items, meta)
}
