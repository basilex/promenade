package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Executor Tests
// ============================================================================

func TestBaseExecutor_ExecuteWithRetry_Success(t *testing.T) {
	executor := NewBaseExecutor(JobTypeLUA)
	job := &Job{
		ID:     uuidv7.New(),
		Name:   "test-job",
		Type:   JobTypeLUA,
		Status: JobStatusPending,
	}

	executionCount := 0
	executeFunc := func(ctx context.Context, j *Job) error {
		executionCount++
		return nil // Success on first attempt
	}

	ctx := context.Background()
	err := executor.ExecuteWithRetry(ctx, job, executeFunc, 3, 10*time.Millisecond)

	if err != nil {
		t.Errorf("ExecuteWithRetry() unexpected error: %v", err)
	}
	if executionCount != 1 {
		t.Errorf("Expected 1 execution, got %d", executionCount)
	}
}

func TestBaseExecutor_ExecuteWithRetry_SuccessAfterRetry(t *testing.T) {
	executor := NewBaseExecutor(JobTypeLUA)
	job := &Job{
		ID:     uuidv7.New(),
		Name:   "test-job",
		Type:   JobTypeLUA,
		Status: JobStatusPending,
	}

	executionCount := 0
	executeFunc := func(ctx context.Context, j *Job) error {
		executionCount++
		if executionCount < 3 {
			return errors.New("temporary failure")
		}
		return nil // Success on 3rd attempt
	}

	ctx := context.Background()
	err := executor.ExecuteWithRetry(ctx, job, executeFunc, 3, 10*time.Millisecond)

	if err != nil {
		t.Errorf("ExecuteWithRetry() unexpected error: %v", err)
	}
	if executionCount != 3 {
		t.Errorf("Expected 3 executions, got %d", executionCount)
	}
}

func TestBaseExecutor_ExecuteWithRetry_AllRetriesFail(t *testing.T) {
	executor := NewBaseExecutor(JobTypeLUA)
	job := &Job{
		ID:     uuidv7.New(),
		Name:   "test-job",
		Type:   JobTypeLUA,
		Status: JobStatusPending,
	}

	executionCount := 0
	testErr := errors.New("permanent failure")
	executeFunc := func(ctx context.Context, j *Job) error {
		executionCount++
		return testErr
	}

	ctx := context.Background()
	maxRetries := 3
	err := executor.ExecuteWithRetry(ctx, job, executeFunc, maxRetries, 10*time.Millisecond)

	if err == nil {
		t.Error("ExecuteWithRetry() expected error, got nil")
	}
	if executionCount != maxRetries+1 {
		t.Errorf("Expected %d executions, got %d", maxRetries+1, executionCount)
	}
}

func TestBaseExecutor_ExecuteWithRetry_ContextCancellation(t *testing.T) {
	executor := NewBaseExecutor(JobTypeLUA)
	job := &Job{
		ID:     uuidv7.New(),
		Name:   "test-job",
		Type:   JobTypeLUA,
		Status: JobStatusPending,
	}

	executionCount := 0
	executeFunc := func(ctx context.Context, j *Job) error {
		executionCount++
		return errors.New("will retry")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	err := executor.ExecuteWithRetry(ctx, job, executeFunc, 3, 10*time.Millisecond)

	if err != context.Canceled {
		t.Errorf("ExecuteWithRetry() error = %v, want %v", err, context.Canceled)
	}
}

func TestBaseExecutor_ExecuteWithRetry_Timeout(t *testing.T) {
	executor := NewBaseExecutor(JobTypeLUA)
	shortTimeout := 50 * time.Millisecond
	job := &Job{
		ID:      uuidv7.New(),
		Name:    "test-job",
		Type:    JobTypeLUA,
		Status:  JobStatusPending,
		Timeout: &shortTimeout,
	}

	executeFunc := func(ctx context.Context, j *Job) error {
		// Wait for context to be canceled (timeout)
		select {
		case <-time.After(200 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	ctx := context.Background()
	err := executor.ExecuteWithRetry(ctx, job, executeFunc, 0, 10*time.Millisecond)

	if err == nil {
		t.Error("ExecuteWithRetry() expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("ExecuteWithRetry() expected context.DeadlineExceeded, got %v", err)
	}
}

func TestBaseExecutor_ExecuteWithRetry_PanicRecovery(t *testing.T) {
	executor := NewBaseExecutor(JobTypeLUA)
	job := &Job{
		ID:     uuidv7.New(),
		Name:   "test-job",
		Type:   JobTypeLUA,
		Status: JobStatusPending,
	}

	executeFunc := func(ctx context.Context, j *Job) error {
		panic("test panic")
	}

	ctx := context.Background()
	err := executor.ExecuteWithRetry(ctx, job, executeFunc, 0, 10*time.Millisecond)

	if err == nil {
		t.Error("ExecuteWithRetry() expected error after panic, got nil")
	}
	// Should contain panic message
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}
