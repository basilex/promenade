package scheduler

import (
	"testing"
)

// ============================================================================
// Error Tests
// ============================================================================

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		// Configuration errors
		{"ErrInvalidWorkerPoolSize", ErrInvalidWorkerPoolSize},
		{"ErrInvalidMaxConcurrentJobs", ErrInvalidMaxConcurrentJobs},
		{"ErrInvalidHealthCheckInterval", ErrInvalidHealthCheckInterval},
		{"ErrInvalidDefaultTimeout", ErrInvalidDefaultTimeout},
		{"ErrInvalidRetryAttempts", ErrInvalidRetryAttempts},
		{"ErrInvalidRetryDelay", ErrInvalidRetryDelay},

		// Engine errors
		{"ErrEngineNotRunning", ErrEngineNotRunning},
		{"ErrEngineAlreadyRunning", ErrEngineAlreadyRunning},
		{"ErrExecutorAlreadyRegistered", ErrExecutorAlreadyRegistered},

		// Job errors
		{"ErrJobNotFound", ErrJobNotFound},
		{"ErrJobAlreadyExists", ErrJobAlreadyExists},
		{"ErrInvalidJobID", ErrInvalidJobID},
		{"ErrInvalidJobName", ErrInvalidJobName},
		{"ErrInvalidCronExpression", ErrInvalidCronExpression},
		{"ErrExecutorNotFound", ErrExecutorNotFound},
		{"ErrJobExecutionFailed", ErrJobExecutionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("Error should not be nil")
			}
			if tt.err.Error() == "" {
				t.Error("Error message should not be empty")
			}
		})
	}
}
