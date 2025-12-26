package entity

import (
	"encoding/json"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// WorkflowStepStatus represents the status of a workflow step
type WorkflowStepStatus string

const (
	WorkflowStepStatusPending   WorkflowStepStatus = "pending"   // Waiting to execute
	WorkflowStepStatusRunning   WorkflowStepStatus = "running"   // Currently executing
	WorkflowStepStatusCompleted WorkflowStepStatus = "completed" // Successfully completed
	WorkflowStepStatusFailed    WorkflowStepStatus = "failed"    // Failed with error
	WorkflowStepStatusSkipped   WorkflowStepStatus = "skipped"   // Skipped (conditional)
	WorkflowStepStatusRetrying  WorkflowStepStatus = "retrying"  // Retrying after failure
)

// WorkflowStepType represents the type of step
type WorkflowStepType string

const (
	WorkflowStepTypeActivity   WorkflowStepType = "activity"   // Execute an activity
	WorkflowStepTypeTransition WorkflowStepType = "transition" // State transition
	WorkflowStepTypeGateway    WorkflowStepType = "gateway"    // Decision point
	WorkflowStepTypeEvent      WorkflowStepType = "event"      // Event trigger
	WorkflowStepTypeTimer      WorkflowStepType = "timer"      // Delay/timer
)

// WorkflowStep represents a single step in workflow execution
// This provides granular audit trail of what happened
type WorkflowStep struct {
	ID           uuidv7.UUID        `db:"id" json:"id"`
	InstanceID   uuidv7.UUID        `db:"instance_id" json:"instance_id"`               // Parent instance
	DefinitionID uuidv7.UUID        `db:"definition_id" json:"definition_id"`           // Definition reference
	StepNumber   int                `db:"step_number" json:"step_number"`               // Sequential number
	Type         WorkflowStepType   `db:"type" json:"type"`                             // Step type
	Status       WorkflowStepStatus `db:"status" json:"status"`                         // Step status
	StateName    string             `db:"state_name" json:"state_name"`                 // State being executed
	ActivityName *string            `db:"activity_name" json:"activity_name,omitempty"` // Activity name if applicable
	Event        *string            `db:"event" json:"event,omitempty"`                 // Event that triggered transition
	Input        json.RawMessage    `db:"input" json:"input,omitempty"`                 // Step input
	Output       json.RawMessage    `db:"output" json:"output,omitempty"`               // Step output
	ErrorMessage *string            `db:"error_message" json:"error_message,omitempty"` // Error if failed
	ErrorDetails json.RawMessage    `db:"error_details" json:"error_details,omitempty"` // Detailed error
	RetryCount   int                `db:"retry_count" json:"retry_count"`               // Retry attempts
	Duration     *int64             `db:"duration" json:"duration,omitempty"`           // Execution time in milliseconds
	ExecutedBy   *uuidv7.UUID       `db:"executed_by" json:"executed_by,omitempty"`     // User if manual step
	StartedAt    *time.Time         `db:"started_at" json:"started_at,omitempty"`       // Execution start
	CompletedAt  *time.Time         `db:"completed_at" json:"completed_at,omitempty"`   // Execution end
	CreatedAt    time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `db:"updated_at" json:"updated_at"`
}

// NewWorkflowStep creates a new workflow step
func NewWorkflowStep(
	instanceID, definitionID uuidv7.UUID,
	stepNumber int,
	stepType WorkflowStepType,
	stateName string,
) *WorkflowStep {
	now := time.Now()

	return &WorkflowStep{
		ID:           uuidv7.New(),
		InstanceID:   instanceID,
		DefinitionID: definitionID,
		StepNumber:   stepNumber,
		Type:         stepType,
		Status:       WorkflowStepStatusPending,
		StateName:    stateName,
		RetryCount:   0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Start marks the step as running
func (ws *WorkflowStep) Start() {
	now := time.Now()
	ws.Status = WorkflowStepStatusRunning
	ws.StartedAt = &now
	ws.UpdatedAt = now
}

// Complete marks the step as completed
func (ws *WorkflowStep) Complete(output json.RawMessage) {
	now := time.Now()
	ws.Status = WorkflowStepStatusCompleted
	ws.Output = output
	ws.CompletedAt = &now
	ws.UpdatedAt = now

	// Calculate duration
	if ws.StartedAt != nil {
		duration := now.Sub(*ws.StartedAt).Milliseconds()
		ws.Duration = &duration
	}
}

// Fail marks the step as failed
func (ws *WorkflowStep) Fail(errorMessage string, errorDetails json.RawMessage) {
	now := time.Now()
	ws.Status = WorkflowStepStatusFailed
	ws.ErrorMessage = &errorMessage
	ws.ErrorDetails = errorDetails
	ws.CompletedAt = &now
	ws.UpdatedAt = now

	// Calculate duration
	if ws.StartedAt != nil {
		duration := now.Sub(*ws.StartedAt).Milliseconds()
		ws.Duration = &duration
	}
}

// Skip marks the step as skipped
func (ws *WorkflowStep) Skip() {
	ws.Status = WorkflowStepStatusSkipped
	ws.UpdatedAt = time.Now()
}

// Retry increments retry count and resets to pending
func (ws *WorkflowStep) Retry() {
	ws.RetryCount++
	ws.Status = WorkflowStepStatusRetrying
	ws.ErrorMessage = nil
	ws.ErrorDetails = nil
	ws.UpdatedAt = time.Now()
}

// SetActivity sets the activity name for activity steps
func (ws *WorkflowStep) SetActivity(activityName string) {
	ws.ActivityName = &activityName
	ws.UpdatedAt = time.Now()
}

// SetEvent sets the event for transition steps
func (ws *WorkflowStep) SetEvent(event string) {
	ws.Event = &event
	ws.UpdatedAt = time.Now()
}

// SetInput sets the step input
func (ws *WorkflowStep) SetInput(input json.RawMessage) {
	ws.Input = input
	ws.UpdatedAt = time.Now()
}

// SetExecutedBy sets the user who executed a manual step
func (ws *WorkflowStep) SetExecutedBy(userID uuidv7.UUID) {
	ws.ExecutedBy = &userID
	ws.UpdatedAt = time.Now()
}

// IsCompleted returns true if the step has finished
func (ws *WorkflowStep) IsCompleted() bool {
	return ws.Status == WorkflowStepStatusCompleted ||
		ws.Status == WorkflowStepStatusFailed ||
		ws.Status == WorkflowStepStatusSkipped
}

// GetDurationMs returns the duration in milliseconds
func (ws *WorkflowStep) GetDurationMs() int64 {
	if ws.Duration != nil {
		return *ws.Duration
	}

	// Calculate on the fly if not stored
	if ws.StartedAt != nil && ws.CompletedAt != nil {
		return ws.CompletedAt.Sub(*ws.StartedAt).Milliseconds()
	}

	return 0
}

// StepDurationStats represents statistics about step execution durations
type StepDurationStats struct {
	StepName    string
	MinDuration int64
	MaxDuration int64
	AvgDuration int64
	TotalSteps  int64
}
