package scheduler

import "errors"

// Configuration errors
var (
	// ErrInvalidWorkerPoolSize is returned when worker pool size is invalid
	ErrInvalidWorkerPoolSize = errors.New("worker pool size must be >= 1")

	// ErrInvalidMaxConcurrentJobs is returned when max concurrent jobs is invalid
	ErrInvalidMaxConcurrentJobs = errors.New("max concurrent jobs must be >= 1")

	// ErrInvalidHealthCheckInterval is returned when health check interval is invalid
	ErrInvalidHealthCheckInterval = errors.New("health check interval must be > 0")

	// ErrInvalidDefaultTimeout is returned when default timeout is invalid
	ErrInvalidDefaultTimeout = errors.New("default timeout must be > 0")

	// ErrInvalidRetryAttempts is returned when retry attempts is invalid
	ErrInvalidRetryAttempts = errors.New("retry attempts must be >= 0")

	// ErrInvalidRetryDelay is returned when retry delay is invalid
	ErrInvalidRetryDelay = errors.New("retry delay must be >= 0")
)

// Engine errors
var (
	// ErrEngineNotRunning is returned when operation requires running engine
	ErrEngineNotRunning = errors.New("scheduler engine is not running")

	// ErrEngineAlreadyRunning is returned when trying to start already running engine
	ErrEngineAlreadyRunning = errors.New("scheduler engine is already running")

	// ErrExecutorAlreadyRegistered is returned when executor type is already registered
	ErrExecutorAlreadyRegistered = errors.New("executor already registered for this job type")
)

// Job errors
var (
	// ErrJobNotFound is returned when job is not found
	ErrJobNotFound = errors.New("job not found")

	// ErrJobAlreadyExists is returned when job with same ID already exists
	ErrJobAlreadyExists = errors.New("job already exists")

	// ErrInvalidJobID is returned when job ID is invalid
	ErrInvalidJobID = errors.New("job ID is required")

	// ErrInvalidJobName is returned when job name is invalid
	ErrInvalidJobName = errors.New("job name is required")

	// ErrInvalidCronExpression is returned when cron expression is invalid
	ErrInvalidCronExpression = errors.New("invalid cron expression")

	// ErrExecutorNotFound is returned when executor for job type is not registered
	ErrExecutorNotFound = errors.New("executor not found for job type")

	// ErrJobExecutionFailed is returned when job execution fails
	ErrJobExecutionFailed = errors.New("job execution failed")
)
