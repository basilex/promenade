package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Engine is the main scheduler engine
type Engine struct {
	config    Config
	cron      *cron.Cron
	jobs      map[uuidv7.UUID]*Job
	executors map[JobType]JobExecutor
	executor  *BaseExecutor
	jobQueue  chan *Job
	results   chan *ExecutionResult
	running   bool
	mu        sync.RWMutex
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewEngine creates new scheduler engine
func NewEngine(cfg Config) (*Engine, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Engine{
		config:    cfg,
		cron:      cron.New(),
		jobs:      make(map[uuidv7.UUID]*Job),
		executors: make(map[JobType]JobExecutor),
		executor:  NewBaseExecutor(JobTypeCustom),
		jobQueue:  make(chan *Job, cfg.MaxConcurrentJobs),
		results:   make(chan *ExecutionResult, 100),
		running:   false,
	}, nil
}

// Start starts the scheduler engine
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return ErrEngineAlreadyRunning
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.running = true

	// Start worker pool
	for i := 0; i < e.config.WorkerPoolSize; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}

	// Start results processor
	e.wg.Add(1)
	go e.processResults()

	// Start health monitor
	e.wg.Add(1)
	go e.healthMonitor()

	// Start cron scheduler
	e.cron.Start()

	slog.Info("Scheduler engine started",
		slog.Int("workers", e.config.WorkerPoolSize),
		slog.Int("max_concurrent", e.config.MaxConcurrentJobs),
	)

	return nil
}

// Stop stops the scheduler engine gracefully
func (e *Engine) Stop() error {
	e.mu.Lock()

	if !e.running {
		e.mu.Unlock()
		return ErrEngineNotRunning
	}

	e.running = false
	e.mu.Unlock()

	// Stop accepting new jobs
	ctx := e.cron.Stop()
	<-ctx.Done()

	// Cancel context to stop workers
	e.cancel()

	// Close job queue after workers finish
	close(e.jobQueue)

	// Wait for all workers to finish
	e.wg.Wait()

	// Close results channel
	close(e.results)

	slog.Info("Scheduler engine stopped")
	return nil
}

// AddJob adds a new job to scheduler
func (e *Engine) AddJob(job *Job) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return ErrEngineNotRunning
	}

	// Check if job already exists
	if _, exists := e.jobs[job.ID]; exists {
		return ErrJobAlreadyExists
	}

	// Validate cron expression (supports standard 5-field and @every descriptors)
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(job.CronExpression); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCronExpression, err)
	}

	// Store job
	e.jobs[job.ID] = job

	// Schedule job if enabled
	if job.Enabled {
		if err := e.scheduleJob(job); err != nil {
			delete(e.jobs, job.ID)
			return err
		}
	}

	slog.Info("Job added",
		slog.String("job_id", job.ID.String()),
		slog.String("name", job.Name),
		slog.String("cron", job.CronExpression),
	)

	return nil
}

// RemoveJob removes a job from scheduler
func (e *Engine) RemoveJob(jobID uuidv7.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return ErrEngineNotRunning
	}

	job, exists := e.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	// Unschedule from cron if was scheduled
	if job.cronEntryID != 0 {
		e.cron.Remove(job.cronEntryID)
	}

	// Remove from map
	delete(e.jobs, jobID)

	slog.Info("Job removed",
		slog.String("job_id", jobID.String()),
	)

	return nil
}

// GetJob retrieves job by ID
func (e *Engine) GetJob(jobID uuidv7.UUID) (*Job, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	job, exists := e.jobs[jobID]
	if !exists {
		return nil, ErrJobNotFound
	}

	return job, nil
}

// ListJobs returns all jobs
func (e *Engine) ListJobs() []*Job {
	e.mu.RLock()
	defer e.mu.RUnlock()

	jobs := make([]*Job, 0, len(e.jobs))
	for _, job := range e.jobs {
		jobs = append(jobs, job)
	}

	return jobs
}

// UpdateJob updates existing job
func (e *Engine) UpdateJob(job *Job) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return ErrEngineNotRunning
	}

	existing, exists := e.jobs[job.ID]
	if !exists {
		return ErrJobNotFound
	}

	// Unschedule old job if was scheduled
	if existing.cronEntryID != 0 {
		e.cron.Remove(existing.cronEntryID)
	}

	// Update job
	e.jobs[job.ID] = job
	job.UpdatedAt = time.Now()

	// Reschedule if enabled
	if job.Enabled {
		if err := e.scheduleJob(job); err != nil {
			// Restore old job on error
			e.jobs[job.ID] = existing
			return err
		}
	}

	slog.Info("Job updated",
		slog.String("job_id", job.ID.String()),
	)

	return nil
}

// EnableJob enables a job
func (e *Engine) EnableJob(jobID uuidv7.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	job, exists := e.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	if job.Enabled {
		return nil // Already enabled
	}

	job.Enabled = true
	job.UpdatedAt = time.Now()

	return e.scheduleJob(job)
}

// DisableJob disables a job
func (e *Engine) DisableJob(jobID uuidv7.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	job, exists := e.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	if !job.Enabled {
		return nil // Already disabled
	}

	job.Enabled = false
	job.UpdatedAt = time.Now()

	// Unschedule from cron if was scheduled
	if job.cronEntryID != 0 {
		e.cron.Remove(job.cronEntryID)
	}

	return nil
}

// RegisterExecutor registers job executor
func (e *Engine) RegisterExecutor(jobType JobType, executor JobExecutor) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.executors[jobType]; exists {
		return ErrExecutorAlreadyRegistered
	}

	e.executors[jobType] = executor

	slog.Info("Executor registered",
		slog.String("job_type", string(jobType)),
	)

	return nil
}

// GetStats returns scheduler statistics
func (e *Engine) GetStats() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := map[string]interface{}{
		"running":         e.running,
		"total_jobs":      len(e.jobs),
		"enabled_jobs":    0,
		"worker_pool":     e.config.WorkerPoolSize,
		"max_concurrent":  e.config.MaxConcurrentJobs,
		"queue_length":    len(e.jobQueue),
		"results_pending": len(e.results),
	}

	for _, job := range e.jobs {
		if job.Enabled {
			stats["enabled_jobs"] = stats["enabled_jobs"].(int) + 1
		}
	}

	return stats
}

// IsRunning returns whether engine is running
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// GetHealth returns health status
func (e *Engine) GetHealth() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	health := map[string]interface{}{
		"status":  "healthy",
		"running": e.running,
	}

	if !e.running {
		health["status"] = "stopped"
		return health
	}

	// Check queue saturation
	queueSaturation := float64(len(e.jobQueue)) / float64(e.config.MaxConcurrentJobs)
	if queueSaturation > 0.8 {
		health["status"] = "degraded"
		health["warning"] = "job queue near capacity"
		health["queue_saturation"] = queueSaturation
	}

	return health
}

// scheduleJob schedules job with cron (must hold lock)
func (e *Engine) scheduleJob(job *Job) error {
	entryID, err := e.cron.AddFunc(job.CronExpression, func() {
		e.enqueueJob(job)
	})

	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	// Store entry ID for later removal
	job.cronEntryID = entryID

	return nil
}

// enqueueJob adds job to execution queue
func (e *Engine) enqueueJob(job *Job) {
	select {
	case e.jobQueue <- job:
		// Job enqueued
	case <-e.ctx.Done():
		// Engine stopping
	default:
		// Queue full
		slog.Warn("Job queue full, dropping job",
			slog.String("job_id", job.ID.String()),
		)
	}
}

// worker processes jobs from queue
func (e *Engine) worker(id int) {
	defer e.wg.Done()

	slog.Debug("Worker started", slog.Int("worker_id", id))

	for {
		select {
		case job, ok := <-e.jobQueue:
			if !ok {
				slog.Debug("Worker stopped", slog.Int("worker_id", id))
				return
			}

			e.executeJob(job)

		case <-e.ctx.Done():
			slog.Debug("Worker cancelled", slog.Int("worker_id", id))
			return
		}
	}
}

// executeJob executes a single job
func (e *Engine) executeJob(job *Job) {
	startTime := time.Now()

	// Update job status
	e.mu.Lock()
	job.Status = JobStatusRunning
	job.LastRun = &startTime
	e.mu.Unlock()

	// Get executor
	e.mu.RLock()
	executor, exists := e.executors[job.Type]
	e.mu.RUnlock()

	if !exists {
		e.recordResult(&ExecutionResult{
			JobID:       job.ID,
			Status:      JobStatusFailed,
			Error:       ErrExecutorNotFound,
			Duration:    time.Since(startTime),
			StartedAt:   startTime,
			CompletedAt: time.Now(),
		})
		return
	}

	// Execute with retry
	maxRetries := job.GetRetryAttempts(e.config.DefaultRetryAttempts)
	retryDelay := job.GetRetryDelay(e.config.DefaultRetryDelay)

	err := e.executor.ExecuteWithRetry(
		e.ctx,
		job,
		executor.Execute,
		maxRetries,
		retryDelay,
	)

	// Record result
	status := JobStatusCompleted
	if err != nil {
		status = JobStatusFailed
	}

	e.recordResult(&ExecutionResult{
		JobID:       job.ID,
		Status:      status,
		Error:       err,
		Duration:    time.Since(startTime),
		StartedAt:   startTime,
		CompletedAt: time.Now(),
	})
}

// recordResult sends execution result to results channel
func (e *Engine) recordResult(result *ExecutionResult) {
	select {
	case e.results <- result:
		// Result recorded
	case <-e.ctx.Done():
		// Engine stopping
	default:
		// Results channel full
		slog.Warn("Results channel full, dropping result",
			slog.String("job_id", result.JobID.String()),
		)
	}
}

// processResults processes execution results
func (e *Engine) processResults() {
	defer e.wg.Done()

	for {
		select {
		case result, ok := <-e.results:
			if !ok {
				return
			}

			// Update job status
			e.mu.Lock()
			if job, exists := e.jobs[result.JobID]; exists {
				job.Status = result.Status
				nextRun := time.Now().Add(time.Minute) // Simplified - should use cron
				job.NextRun = &nextRun
			}
			e.mu.Unlock()

			// Log result
			if result.Error != nil {
				slog.Error("Job failed",
					slog.String("job_id", result.JobID.String()),
					slog.Duration("duration", result.Duration),
					slog.Any("error", result.Error),
				)
			} else {
				slog.Info("Job completed",
					slog.String("job_id", result.JobID.String()),
					slog.Duration("duration", result.Duration),
				)
			}

		case <-e.ctx.Done():
			return
		}
	}
}

// healthMonitor monitors engine health
func (e *Engine) healthMonitor() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			health := e.GetHealth()
			if health["status"] != "healthy" {
				slog.Warn("Scheduler health degraded",
					slog.Any("health", health),
				)
			}

		case <-e.ctx.Done():
			return
		}
	}
}
