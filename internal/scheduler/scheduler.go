package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/taufiqgit/news-api/internal/config"
	"github.com/taufiqgit/news-api/internal/domain/entity"
	"github.com/taufiqgit/news-api/internal/domain/repository"
	"github.com/taufiqgit/news-api/internal/service"
)

// Scheduler menjalankan pipeline berita secara periodik (in-process goroutine).
type Scheduler struct {
	cfg      config.SchedulerConfig
	pipeline *service.Pipeline
	webRepo  repository.WebsiteRepository
	logger   *slog.Logger

	mu      sync.Mutex
	running atomic.Bool
	cancel  context.CancelFunc
	done    chan struct{}
}

// New membangun Scheduler dengan dependency pipeline & website repository.
func New(cfg config.SchedulerConfig, pipeline *service.Pipeline, webRepo repository.WebsiteRepository, logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	interval := cfg.IntervalMinutes
	if interval <= 0 {
		interval = 10
	}
	cfg.IntervalMinutes = interval

	concurrency := cfg.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 3
	}
	cfg.MaxConcurrency = concurrency

	return &Scheduler{
		cfg:      cfg,
		pipeline: pipeline,
		webRepo:  webRepo,
		logger:   logger,
		done:     make(chan struct{}),
	}
}

// Start memulai ticker. Idempoten — tidak akan memulai dua kali.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running.Load() {
		return
	}

	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.running.Store(true)

	go s.loop(runCtx)
	s.logger.Info("scheduler started", "interval_minutes", s.cfg.IntervalMinutes, "max_concurrency", s.cfg.MaxConcurrency)
}

// Stop menghentikan scheduler secara graceful dan menunggu loop selesai.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running.Load() {
		s.mu.Unlock()
		return
	}
	s.running.Store(false)
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	<-s.done
	s.logger.Info("scheduler stopped")
}

// loop adalah ticker utama.
func (s *Scheduler) loop(ctx context.Context) {
	defer close(s.done)

	interval := time.Duration(s.cfg.IntervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Jalankan sekali segera saat start (opsional tapi berguna untuk dev).
	s.runOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

// runOnce menjalankan satu siklus: ambil website aktif, proses paralel.
// Anti-overlap dijaga oleh ticker (satu goroutine) — tapi guard tambahan
// memastikan tidak ada run ganda bila dipanggil manual.
func (s *Scheduler) runOnce(ctx context.Context) {
	websites, err := s.webRepo.ListActive(ctx)
	if err != nil {
		s.logger.Error("scheduler: list active websites", "error", err)
		return
	}
	if len(websites) == 0 {
		s.logger.Info("scheduler: no active websites")
		return
	}

	s.logger.Info("scheduler: run started", "websites", len(websites))

	// Worker pool terbatas.
	concurrency := s.cfg.MaxConcurrency
	if concurrency > len(websites) {
		concurrency = len(websites)
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for _, ws := range websites {
		wg.Add(1)
		sem <- struct{}{}
		go func(w entity.Website) {
			defer wg.Done()
			defer func() { <-sem }()

			res, err := s.pipeline.ProcessWebsite(ctx, w)
			if err != nil {
				s.logger.Error("scheduler: process website", "website", w.Slug, "error", err)
				return
			}
			s.logger.Info("scheduler: website done",
				"website", w.Slug,
				"created", res.Created,
				"skipped", res.Skipped,
				"failed", res.Failed,
			)
		}(ws)
	}

	wg.Wait()
	s.logger.Info("scheduler: run finished")
}

// RunNow memicu satu siklus secara manual (untuk admin endpoint / testing).
func (s *Scheduler) RunNow(ctx context.Context) error {
	s.runOnce(ctx)
	return nil
}
