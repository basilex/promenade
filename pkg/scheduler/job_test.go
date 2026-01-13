package scheduler

import (
	"testing"
	"time"
)

// ============================================================================
// Job Tests
// ============================================================================

func TestJob_GetTimeout(t *testing.T) {
	defaultTimeout := 300 * time.Second

	tests := []struct {
		name string
		job  *Job
		want time.Duration
	}{
		{
			name: "no custom timeout",
			job:  &Job{},
			want: defaultTimeout,
		},
		{
			name: "with custom timeout",
			job:  &Job{Timeout: durationPtr(120 * time.Second)},
			want: 120 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.GetTimeout(defaultTimeout)
			if got != tt.want {
				t.Errorf("GetTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_GetRetryAttempts(t *testing.T) {
	defaultRetries := 3

	tests := []struct {
		name string
		job  *Job
		want int
	}{
		{
			name: "no custom retry attempts",
			job:  &Job{},
			want: defaultRetries,
		},
		{
			name: "with custom retry attempts",
			job:  &Job{RetryAttempts: intPtr(5)},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.GetRetryAttempts(defaultRetries)
			if got != tt.want {
				t.Errorf("GetRetryAttempts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_GetRetryDelay(t *testing.T) {
	defaultDelay := 60 * time.Second

	tests := []struct {
		name string
		job  *Job
		want time.Duration
	}{
		{
			name: "no custom retry delay",
			job:  &Job{},
			want: defaultDelay,
		},
		{
			name: "with custom retry delay",
			job:  &Job{RetryDelay: durationPtr(30 * time.Second)},
			want: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.GetRetryDelay(defaultDelay)
			if got != tt.want {
				t.Errorf("GetRetryDelay() = %v, want %v", got, tt.want)
			}
		})
	}
}
