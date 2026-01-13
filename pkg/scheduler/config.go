package scheduler

import "time"

// Config defines scheduler configuration
type Config struct {
	// Enabled controls whether scheduler is active
	Enabled bool `yaml:"enabled"`

	// WorkerPoolSize is number of concurrent worker goroutines
	WorkerPoolSize int `yaml:"worker_pool_size"`

	// MaxConcurrentJobs limits concurrent job execution
	MaxConcurrentJobs int `yaml:"max_concurrent_jobs"`

	// HealthCheckInterval for monitoring scheduler health
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`

	// DefaultTimeout for job execution (can be overridden per job)
	DefaultTimeout time.Duration `yaml:"default_timeout"`

	// DefaultRetryAttempts for failed jobs (can be overridden per job)
	DefaultRetryAttempts int `yaml:"default_retry_attempts"`

	// DefaultRetryDelay between retry attempts (can be overridden per job)
	DefaultRetryDelay time.Duration `yaml:"default_retry_delay"`
}

// DefaultConfig returns default scheduler configuration
func DefaultConfig() Config {
	return Config{
		Enabled:              true,
		WorkerPoolSize:       10,
		MaxConcurrentJobs:    5,
		HealthCheckInterval:  60 * time.Second,
		DefaultTimeout:       300 * time.Second, // 5 minutes
		DefaultRetryAttempts: 3,
		DefaultRetryDelay:    60 * time.Second,
	}
}

// Validate checks configuration validity
func (c *Config) Validate() error {
	if c.WorkerPoolSize < 1 {
		return ErrInvalidWorkerPoolSize
	}
	if c.MaxConcurrentJobs < 1 {
		return ErrInvalidMaxConcurrentJobs
	}
	if c.HealthCheckInterval <= 0 {
		return ErrInvalidHealthCheckInterval
	}
	if c.DefaultTimeout <= 0 {
		return ErrInvalidDefaultTimeout
	}
	if c.DefaultRetryAttempts < 0 {
		return ErrInvalidRetryAttempts
	}
	if c.DefaultRetryDelay < 0 {
		return ErrInvalidRetryDelay
	}
	return nil
}
