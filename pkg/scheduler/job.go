package scheduler

import (
	"context"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/robfig/cron/v3"
)

// JobType defines type of job
type JobType string

const (
	// JobTypeLUA executes LUA script
	JobTypeLUA JobType = "lua"

	// JobTypeEvent publishes event to event bus
	JobTypeEvent JobType = "event"

	// JobTypeHTTP makes HTTP request
	JobTypeHTTP JobType = "http"

	// JobTypeCustom for custom executor implementations
	JobTypeCustom JobType = "custom"
)

// JobStatus defines job execution status
type JobStatus string

const (
	// JobStatusPending job is waiting to execute
	JobStatusPending JobStatus = "pending"

	// JobStatusRunning job is currently executing
	JobStatusRunning JobStatus = "running"

	// JobStatusCompleted job completed successfully
	JobStatusCompleted JobStatus = "completed"

	// JobStatusFailed job failed
	JobStatusFailed JobStatus = "failed"

	// JobStatusRetrying job is retrying after failure
	JobStatusRetrying JobStatus = "retrying"

	// JobStatusCancelled job was cancelled
	JobStatusCancelled JobStatus = "cancelled"
)

// Job represents a scheduled job
type Job struct {
	// ID is unique job identifier
	ID uuidv7.UUID

	// Name is human-readable job name
	Name string

	// Type defines what kind of job this is
	Type JobType

	// CronExpression defines when job runs
	CronExpression string

	// cronEntryID stores cron scheduler entry ID (internal use)
	cronEntryID cron.EntryID

	// Payload is job-specific data (LUA code, HTTP URL, etc.)
	Payload map[string]interface{}

	// Status is current job status
	Status JobStatus

	// LastRun is timestamp of last execution
	LastRun *time.Time

	// NextRun is timestamp of next scheduled execution
	NextRun *time.Time

	// Timeout overrides default timeout for this job
	Timeout *time.Duration

	// RetryAttempts overrides default retry attempts
	RetryAttempts *int

	// RetryDelay overrides default retry delay
	RetryDelay *time.Duration

	// Enabled controls whether job is active
	Enabled bool

	// CreatedAt is job creation timestamp
	CreatedAt time.Time

	// UpdatedAt is last update timestamp
	UpdatedAt time.Time
}

// JobExecutor executes jobs of specific type
type JobExecutor interface {
	// Execute runs the job
	Execute(ctx context.Context, job *Job) error

	// Type returns job type this executor handles
	Type() JobType
}

// ExecutionResult represents job execution outcome
type ExecutionResult struct {
	// JobID is the job that was executed
	JobID uuidv7.UUID

	// Status is execution result
	Status JobStatus

	// Error if execution failed
	Error error

	// Duration of execution
	Duration time.Duration

	// Output from job execution (optional)
	Output map[string]interface{}

	// StartedAt is when execution started
	StartedAt time.Time

	// CompletedAt is when execution finished
	CompletedAt time.Time
}

// GetTimeout returns job timeout or default
func (j *Job) GetTimeout(defaultTimeout time.Duration) time.Duration {
	if j.Timeout != nil {
		return *j.Timeout
	}
	return defaultTimeout
}

// GetRetryAttempts returns job retry attempts or default
func (j *Job) GetRetryAttempts(defaultRetries int) int {
	if j.RetryAttempts != nil {
		return *j.RetryAttempts
	}
	return defaultRetries
}

// GetRetryDelay returns job retry delay or default
func (j *Job) GetRetryDelay(defaultDelay time.Duration) time.Duration {
	if j.RetryDelay != nil {
		return *j.RetryDelay
	}
	return defaultDelay
}
