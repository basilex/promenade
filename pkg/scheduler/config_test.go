package scheduler

import (
	"errors"
	"testing"
	"time"
)

// ============================================================================
// Config Tests
// ============================================================================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("Expected Enabled=true")
	}
	if cfg.WorkerPoolSize != 10 {
		t.Errorf("Expected WorkerPoolSize=10, got %d", cfg.WorkerPoolSize)
	}
	if cfg.MaxConcurrentJobs != 5 {
		t.Errorf("Expected MaxConcurrentJobs=5, got %d", cfg.MaxConcurrentJobs)
	}
	if cfg.HealthCheckInterval != 60*time.Second {
		t.Errorf("Expected HealthCheckInterval=60s, got %v", cfg.HealthCheckInterval)
	}
	if cfg.DefaultTimeout != 300*time.Second {
		t.Errorf("Expected DefaultTimeout=300s, got %v", cfg.DefaultTimeout)
	}
	if cfg.DefaultRetryAttempts != 3 {
		t.Errorf("Expected DefaultRetryAttempts=3, got %d", cfg.DefaultRetryAttempts)
	}
	if cfg.DefaultRetryDelay != 60*time.Second {
		t.Errorf("Expected DefaultRetryDelay=60s, got %v", cfg.DefaultRetryDelay)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: nil,
		},
		{
			name: "invalid worker pool size",
			cfg: Config{
				WorkerPoolSize:       0,
				MaxConcurrentJobs:    5,
				HealthCheckInterval:  60 * time.Second,
				DefaultTimeout:       300 * time.Second,
				DefaultRetryAttempts: 3,
				DefaultRetryDelay:    60 * time.Second,
			},
			wantErr: ErrInvalidWorkerPoolSize,
		},
		{
			name: "invalid max concurrent jobs",
			cfg: Config{
				WorkerPoolSize:       10,
				MaxConcurrentJobs:    0,
				HealthCheckInterval:  60 * time.Second,
				DefaultTimeout:       300 * time.Second,
				DefaultRetryAttempts: 3,
				DefaultRetryDelay:    60 * time.Second,
			},
			wantErr: ErrInvalidMaxConcurrentJobs,
		},
		{
			name: "invalid health check interval",
			cfg: Config{
				WorkerPoolSize:       10,
				MaxConcurrentJobs:    5,
				HealthCheckInterval:  0,
				DefaultTimeout:       300 * time.Second,
				DefaultRetryAttempts: 3,
				DefaultRetryDelay:    60 * time.Second,
			},
			wantErr: ErrInvalidHealthCheckInterval,
		},
		{
			name: "invalid default timeout",
			cfg: Config{
				WorkerPoolSize:       10,
				MaxConcurrentJobs:    5,
				HealthCheckInterval:  60 * time.Second,
				DefaultTimeout:       0,
				DefaultRetryAttempts: 3,
				DefaultRetryDelay:    60 * time.Second,
			},
			wantErr: ErrInvalidDefaultTimeout,
		},
		{
			name: "invalid retry attempts",
			cfg: Config{
				WorkerPoolSize:       10,
				MaxConcurrentJobs:    5,
				HealthCheckInterval:  60 * time.Second,
				DefaultTimeout:       300 * time.Second,
				DefaultRetryAttempts: -1,
				DefaultRetryDelay:    60 * time.Second,
			},
			wantErr: ErrInvalidRetryAttempts,
		},
		{
			name: "invalid retry delay",
			cfg: Config{
				WorkerPoolSize:       10,
				MaxConcurrentJobs:    5,
				HealthCheckInterval:  60 * time.Second,
				DefaultTimeout:       300 * time.Second,
				DefaultRetryAttempts: 3,
				DefaultRetryDelay:    -1,
			},
			wantErr: ErrInvalidRetryDelay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
