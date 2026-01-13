package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// BaseExecutor provides common retry logic for all executors
type BaseExecutor struct {
	jobType JobType
}

// NewBaseExecutor creates base executor
func NewBaseExecutor(jobType JobType) *BaseExecutor {
	return &BaseExecutor{jobType: jobType}
}

// Type returns job type
func (e *BaseExecutor) Type() JobType {
	return e.jobType
}

// ExecuteWithRetry executes job with retry logic
func (e *BaseExecutor) ExecuteWithRetry(
	ctx context.Context,
	job *Job,
	executeFunc func(context.Context, *Job) error,
	maxRetries int,
	retryDelay time.Duration,
) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Create timeout context
		execCtx, cancel := context.WithTimeout(ctx, job.GetTimeout(300*time.Second))

		// Execute with panic recovery
		err := e.executeOnce(execCtx, job, executeFunc)
		cancel()

		// Success
		if err == nil {
			if attempt > 0 {
				slog.Info("Job succeeded after retry",
					slog.String("job_id", job.ID.String()),
					slog.Int("attempt", attempt+1),
				)
			}
			return nil
		}

		lastErr = err

		// Check if we should retry
		if attempt < maxRetries {
			slog.Warn("Job failed, will retry",
				slog.String("job_id", job.ID.String()),
				slog.Int("attempt", attempt+1),
				slog.Int("max_retries", maxRetries),
				slog.Any("error", err),
			)

			// Wait before retry (with context cancellation support)
			select {
			case <-time.After(retryDelay):
				// Continue to next attempt
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("job failed after %d attempts: %w", maxRetries+1, lastErr)
}

// executeOnce executes job once with panic recovery
func (e *BaseExecutor) executeOnce(
	ctx context.Context,
	job *Job,
	executeFunc func(context.Context, *Job) error,
) (err error) {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("job panicked: %v", r)
			slog.Error("Job execution panic",
				slog.String("job_id", job.ID.String()),
				slog.Any("panic", r),
			)
		}
	}()

	// Execute
	return executeFunc(ctx, job)
}
