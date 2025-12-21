package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/robfig/cron/v3"
)

// Scheduler manages scheduled jobs using cron
type Scheduler struct {
	cron          *cron.Cron
	purgeUseCase  usecase.PurgeUseCase
	schedule      string
	dryRun        bool
	enabled       bool
	mu            sync.Mutex
	isRunning     bool
	lastRunTime   *time.Time
	lastRunStatus string
}

// NewScheduler creates a new scheduler
func NewScheduler(
	purgeUseCase usecase.PurgeUseCase,
	schedule string,
	dryRun bool,
	enabled bool,
) *Scheduler {
	return &Scheduler{
		cron:         cron.New(cron.WithSeconds()),
		purgeUseCase: purgeUseCase,
		schedule:     schedule,
		dryRun:       dryRun,
		enabled:      enabled,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	if !s.enabled {
		logger.Info("Purge scheduler is disabled")
		return nil
	}

	logger.Info("Starting purge scheduler",
		"schedule", s.schedule,
		"dry_run", s.dryRun,
	)

	// Add purge job
	_, err := s.cron.AddFunc(s.schedule, func() {
		s.runPurgeJob(ctx)
	})
	if err != nil {
		logger.Error("Failed to schedule purge job", "error", err)
		return err
	}

	s.cron.Start()
	logger.Info("Purge scheduler started successfully")

	return nil
}

// Stop stops the scheduler gracefully
func (s *Scheduler) Stop(ctx context.Context) error {
	logger.Info("Stopping purge scheduler...")

	// Stop accepting new jobs
	stopCtx := s.cron.Stop()

	// Wait for running jobs to complete
	select {
	case <-stopCtx.Done():
		logger.Info("Purge scheduler stopped gracefully")
	case <-ctx.Done():
		logger.Warn("Purge scheduler stopped due to context cancellation")
	}

	return nil
}

// runPurgeJob executes the purge job
func (s *Scheduler) runPurgeJob(ctx context.Context) {
	s.mu.Lock()
	if s.isRunning {
		logger.Warn("Purge job is already running, skipping this execution")
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isRunning = false
		now := time.Now()
		s.lastRunTime = &now
		s.mu.Unlock()
	}()

	logger.Info("Starting scheduled purge job", "dry_run", s.dryRun)

	summary, err := s.purgeUseCase.PurgeAll(ctx, s.dryRun)
	if err != nil {
		s.mu.Lock()
		s.lastRunStatus = "failed"
		s.mu.Unlock()

		logger.Error("Scheduled purge job failed", "error", err)
		return
	}

	s.mu.Lock()
	s.lastRunStatus = "success"
	s.mu.Unlock()

	logger.Info("Scheduled purge job completed",
		"total_records_purged", summary.TotalRecordsPurged,
		"duration", summary.Duration,
		"dry_run", s.dryRun,
	)
}

// GetStatus returns the current status of the scheduler
func (s *Scheduler) GetStatus() SchedulerStatus {
	s.mu.Lock()
	defer s.mu.Unlock()

	return SchedulerStatus{
		Enabled:       s.enabled,
		Schedule:      s.schedule,
		DryRun:        s.dryRun,
		IsRunning:     s.isRunning,
		LastRunTime:   s.lastRunTime,
		LastRunStatus: s.lastRunStatus,
	}
}

// TriggerNow manually triggers the purge job
func (s *Scheduler) TriggerNow(ctx context.Context, dryRun bool) error {
	logger.Info("Manually triggering purge job", "dry_run", dryRun)

	summary, err := s.purgeUseCase.PurgeAll(ctx, dryRun)
	if err != nil {
		logger.Error("Manual purge job failed", "error", err)
		return err
	}

	logger.Info("Manual purge job completed",
		"total_records_purged", summary.TotalRecordsPurged,
		"duration", summary.Duration,
		"dry_run", dryRun,
	)

	return nil
}

// SchedulerStatus represents the current state of the scheduler
type SchedulerStatus struct {
	Enabled       bool       `json:"enabled"`
	Schedule      string     `json:"schedule"`
	DryRun        bool       `json:"dry_run"`
	IsRunning     bool       `json:"is_running"`
	LastRunTime   *time.Time `json:"last_run_time,omitempty"`
	LastRunStatus string     `json:"last_run_status,omitempty"`
}
