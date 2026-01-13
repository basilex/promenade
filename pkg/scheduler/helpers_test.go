package scheduler

import (
	"context"
	"time"
)

// ============================================================================
// Test Helper Functions
// ============================================================================

// durationPtr returns a pointer to the given duration.
// Used in tests to set optional Job fields.
func durationPtr(d time.Duration) *time.Duration {
	return &d
}

// intPtr returns a pointer to the given integer.
// Used in tests to set optional Job fields.
func intPtr(i int) *int {
	return &i
}

// ============================================================================
// Mock Executor
// ============================================================================

// MockExecutor is a test double for IExecutor interface.
// It implements IExecutor and allows tests to control execution behavior.
type MockExecutor struct {
	id          string
	executeFunc func(ctx context.Context, job *Job) error
}

// Type returns the job type this executor handles
func (m *MockExecutor) Type() JobType {
	return JobTypeLUA
}

// Execute executes a job (calls executeFunc if set)
func (m *MockExecutor) Execute(ctx context.Context, job *Job) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, job)
	}
	return nil
}

// ExecuteWithRetry executes with retries
func (m *MockExecutor) ExecuteWithRetry(ctx context.Context, job *Job, executeFunc func(context.Context, *Job) error, maxRetries int, retryDelay time.Duration) error {
	if executeFunc != nil {
		return executeFunc(ctx, job)
	}
	return nil
}
