package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/jsonb"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// WorkflowInstanceStatus represents the execution status
type WorkflowInstanceStatus string

const (
	WorkflowInstanceStatusPending   WorkflowInstanceStatus = "pending"   // Waiting to start
	WorkflowInstanceStatusRunning   WorkflowInstanceStatus = "running"   // Currently executing
	WorkflowInstanceStatusWaiting   WorkflowInstanceStatus = "waiting"   // Waiting for external event
	WorkflowInstanceStatusPaused    WorkflowInstanceStatus = "paused"    // Manually paused
	WorkflowInstanceStatusCompleted WorkflowInstanceStatus = "completed" // Successfully finished
	WorkflowInstanceStatusFailed    WorkflowInstanceStatus = "failed"    // Failed with error
	WorkflowInstanceStatusCancelled WorkflowInstanceStatus = "cancelled" // Manually cancelled
	WorkflowInstanceStatusTimedOut  WorkflowInstanceStatus = "timed_out" // Exceeded timeout
)

// WorkflowInstance represents a running instance of a workflow
type WorkflowInstance struct {
	ID                uuidv7.UUID            `db:"id" json:"id"`
	DefinitionID      uuidv7.UUID            `db:"definition_id" json:"definition_id"`                     // Reference to definition
	DefinitionVersion int                    `db:"definition_version" json:"definition_version"`           // Version used
	Status            WorkflowInstanceStatus `db:"status" json:"status"`                                   // Current status
	CurrentState      string                 `db:"current_state" json:"current_state"`                     // Current state name
	PreviousState     *string                `db:"previous_state" json:"previous_state,omitempty"`         // Previous state
	Context           jsonb.Map              `db:"context" json:"context"`                                 // Workflow variables (JSONB)
	Input             jsonb.Map              `db:"input" json:"input"`                                     // Initial input data
	Output            jsonb.Map              `db:"output" json:"output,omitempty"`                         // Final output data
	ErrorMessage      *string                `db:"error_message" json:"error_message,omitempty"`           // Error if failed
	ErrorDetails      jsonb.Map              `db:"error_details" json:"error_details,omitempty"`           // Detailed error info
	StartedBy         uuidv7.UUID            `db:"started_by" json:"started_by"`                           // User who started
	AssignedTo        *uuidv7.UUID           `db:"assigned_to" json:"assigned_to,omitempty"`               // Current assignee (for manual tasks)
	Priority          int                    `db:"priority" json:"priority"`                               // Priority (1-10)
	DueDate           *time.Time             `db:"due_date" json:"due_date,omitempty"`                     // Expected completion date
	ParentInstanceID  *uuidv7.UUID           `db:"parent_instance_id" json:"parent_instance_id,omitempty"` // Parent workflow (sub-workflows)
	ExternalReference *string                `db:"external_reference" json:"external_reference,omitempty"` // External ID (order_id, ticket_id, etc.)
	Tags              []string               `db:"-" json:"tags"`                                          // Tags for filtering (handled manually in repo)
	RetryCount        int                    `db:"retry_count" json:"retry_count"`                         // Number of retries
	StartedAt         *time.Time             `db:"started_at" json:"started_at,omitempty"`                 // When execution started
	CompletedAt       *time.Time             `db:"completed_at" json:"completed_at,omitempty"`             // When execution completed
	StateEnteredAt    time.Time              `db:"state_entered_at" json:"state_entered_at"`               // When entered current state
	StateTimeoutAt    *time.Time             `db:"state_timeout_at" json:"state_timeout_at,omitempty"`     // When current state times out
	CreatedAt         time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time              `db:"updated_at" json:"updated_at"`
	DeletedAt         *time.Time             `db:"deleted_at" json:"deleted_at,omitempty"` // Soft delete
}

// Validation errors
var (
	ErrWorkflowInstanceInvalidDefinition = errors.New("workflow instance must reference a valid definition")
	ErrWorkflowInstanceInvalidState      = errors.New("workflow instance has invalid state")
	ErrWorkflowInstanceAlreadyCompleted  = errors.New("workflow instance already completed")
	ErrWorkflowInstanceNotRunning        = errors.New("workflow instance is not running")
	ErrWorkflowInstanceInvalidTransition = errors.New("invalid state transition")
)

// NewWorkflowInstance creates a new workflow instance
func NewWorkflowInstance(
	definitionID uuidv7.UUID,
	definitionVersion int,
	initialState string,
	input jsonb.Map,
	startedBy uuidv7.UUID,
) *WorkflowInstance {
	now := time.Now()

	return &WorkflowInstance{
		ID:                uuidv7.New(),
		DefinitionID:      definitionID,
		DefinitionVersion: definitionVersion,
		Status:            WorkflowInstanceStatusPending,
		CurrentState:      initialState,
		Context:           jsonb.Map{},
		Input:             input,
		StartedBy:         startedBy,
		Priority:          5, // Default medium priority
		Tags:              []string{},
		RetryCount:        0,
		StateEnteredAt:    now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// Start transitions the instance from pending to running
func (wi *WorkflowInstance) Start() error {
	if wi.Status != WorkflowInstanceStatusPending {
		return fmt.Errorf("cannot start workflow in status %s", wi.Status)
	}

	now := time.Now()
	wi.Status = WorkflowInstanceStatusRunning
	wi.StartedAt = &now
	wi.UpdatedAt = now
	return nil
}

// TransitionTo moves the workflow to a new state
func (wi *WorkflowInstance) TransitionTo(newState string, event string) error {
	if !wi.IsActive() {
		return ErrWorkflowInstanceNotRunning
	}

	previous := wi.CurrentState
	now := time.Now()

	wi.PreviousState = &previous
	wi.CurrentState = newState
	wi.StateEnteredAt = now
	wi.StateTimeoutAt = nil // Reset timeout, will be set if state has timeout
	wi.UpdatedAt = now

	return nil
}

// Complete marks the workflow as completed successfully
func (wi *WorkflowInstance) Complete(output jsonb.Map) error {
	if !wi.IsActive() {
		return ErrWorkflowInstanceNotRunning
	}

	now := time.Now()
	wi.Status = WorkflowInstanceStatusCompleted
	wi.Output = output
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// Fail marks the workflow as failed
func (wi *WorkflowInstance) Fail(errorMessage string, errorDetails jsonb.Map) error {
	if wi.IsCompleted() {
		return ErrWorkflowInstanceAlreadyCompleted
	}

	now := time.Now()
	wi.Status = WorkflowInstanceStatusFailed
	wi.ErrorMessage = &errorMessage
	wi.ErrorDetails = errorDetails
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// Cancel cancels the workflow execution
func (wi *WorkflowInstance) Cancel() error {
	if !wi.IsActive() {
		return fmt.Errorf("cannot cancel workflow in status %s", wi.Status)
	}

	now := time.Now()
	wi.Status = WorkflowInstanceStatusCancelled
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// Pause pauses the workflow execution
func (wi *WorkflowInstance) Pause() error {
	if wi.Status != WorkflowInstanceStatusRunning && wi.Status != WorkflowInstanceStatusWaiting {
		return fmt.Errorf("cannot pause workflow in status %s", wi.Status)
	}

	wi.Status = WorkflowInstanceStatusPaused
	wi.UpdatedAt = time.Now()
	return nil
}

// Resume resumes a paused workflow
func (wi *WorkflowInstance) Resume() error {
	if wi.Status != WorkflowInstanceStatusPaused {
		return fmt.Errorf("cannot resume workflow in status %s", wi.Status)
	}

	wi.Status = WorkflowInstanceStatusRunning
	wi.UpdatedAt = time.Now()
	return nil
}

// WaitForEvent transitions to waiting status
func (wi *WorkflowInstance) WaitForEvent() error {
	if wi.Status != WorkflowInstanceStatusRunning {
		return fmt.Errorf("cannot wait in status %s", wi.Status)
	}

	wi.Status = WorkflowInstanceStatusWaiting
	wi.UpdatedAt = time.Now()
	return nil
}

// Timeout marks the workflow as timed out
func (wi *WorkflowInstance) Timeout() error {
	if !wi.IsActive() {
		return fmt.Errorf("cannot timeout workflow in status %s", wi.Status)
	}

	now := time.Now()
	errorMsg := fmt.Sprintf("workflow timed out in state '%s'", wi.CurrentState)
	wi.Status = WorkflowInstanceStatusTimedOut
	wi.ErrorMessage = &errorMsg
	wi.CompletedAt = &now
	wi.UpdatedAt = now
	return nil
}

// IncrementRetry increments the retry counter
func (wi *WorkflowInstance) IncrementRetry() {
	wi.RetryCount++
	wi.UpdatedAt = time.Now()
}

// UpdateContext updates the workflow context variables
func (wi *WorkflowInstance) UpdateContext(context jsonb.Map) {
	wi.Context = context
	wi.UpdatedAt = time.Now()
}

// Assign assigns the workflow to a user
func (wi *WorkflowInstance) Assign(userID uuidv7.UUID) {
	wi.AssignedTo = &userID
	wi.UpdatedAt = time.Now()
}

// Unassign removes the assignment
func (wi *WorkflowInstance) Unassign() {
	wi.AssignedTo = nil
	wi.UpdatedAt = time.Now()
}

// SetPriority sets the workflow priority (1-10)
func (wi *WorkflowInstance) SetPriority(priority int) error {
	if priority < 1 || priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}
	wi.Priority = priority
	wi.UpdatedAt = time.Now()
	return nil
}

// SetDueDate sets the expected completion date
func (wi *WorkflowInstance) SetDueDate(dueDate time.Time) {
	wi.DueDate = &dueDate
	wi.UpdatedAt = time.Now()
}

// SetStateTimeout sets when the current state will timeout
func (wi *WorkflowInstance) SetStateTimeout(timeout time.Duration) {
	timeoutAt := time.Now().Add(timeout)
	wi.StateTimeoutAt = &timeoutAt
	wi.UpdatedAt = time.Now()
}

// IsActive returns true if the workflow is actively executing
func (wi *WorkflowInstance) IsActive() bool {
	return wi.Status == WorkflowInstanceStatusRunning ||
		wi.Status == WorkflowInstanceStatusWaiting ||
		wi.Status == WorkflowInstanceStatusPending
}

// IsCompleted returns true if the workflow has finished (success or failure)
func (wi *WorkflowInstance) IsCompleted() bool {
	return wi.Status == WorkflowInstanceStatusCompleted ||
		wi.Status == WorkflowInstanceStatusFailed ||
		wi.Status == WorkflowInstanceStatusCancelled ||
		wi.Status == WorkflowInstanceStatusTimedOut
}

// IsOverdue returns true if the workflow exceeded its due date
func (wi *WorkflowInstance) IsOverdue() bool {
	if wi.DueDate == nil || wi.IsCompleted() {
		return false
	}
	return time.Now().After(*wi.DueDate)
}

// IsStateTimedOut returns true if the current state has exceeded its timeout
func (wi *WorkflowInstance) IsStateTimedOut() bool {
	if wi.StateTimeoutAt == nil || wi.IsCompleted() {
		return false
	}
	return time.Now().After(*wi.StateTimeoutAt)
}

// GetDuration returns the total execution time
func (wi *WorkflowInstance) GetDuration() time.Duration {
	if wi.StartedAt == nil {
		return 0
	}

	endTime := time.Now()
	if wi.CompletedAt != nil {
		endTime = *wi.CompletedAt
	}

	return endTime.Sub(*wi.StartedAt)
}

// GetStateContext returns context as a map
func (wi *WorkflowInstance) GetStateContext() (map[string]interface{}, error) {
	// jsonb.Map is already map[string]interface{}, no unmarshaling needed
	return wi.Context, nil
}

// Validate validates the workflow instance
func (wi *WorkflowInstance) Validate() error {
	if wi.DefinitionID == uuidv7.Nil {
		return ErrWorkflowInstanceInvalidDefinition
	}

	if wi.CurrentState == "" {
		return ErrWorkflowInstanceInvalidState
	}

	if wi.Priority < 1 || wi.Priority > 10 {
		return fmt.Errorf("priority must be between 1 and 10")
	}

	return nil
}
