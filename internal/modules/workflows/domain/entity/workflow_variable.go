package entity

import (
	"encoding/json"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// WorkflowVariableScope defines where the variable is accessible
type WorkflowVariableScope string

const (
	WorkflowVariableScopeGlobal WorkflowVariableScope = "global" // Available throughout workflow
	WorkflowVariableScopeState  WorkflowVariableScope = "state"  // Available only in specific state
	WorkflowVariableScopeLocal  WorkflowVariableScope = "local"  // Available only in current activity
)

// WorkflowVariable represents a variable in workflow execution context
// Variables can store intermediate results, decisions, user input, etc.
type WorkflowVariable struct {
	ID         uuidv7.UUID           `db:"id" json:"id"`
	InstanceID uuidv7.UUID           `db:"instance_id" json:"instance_id"`         // Parent instance
	Name       string                `db:"name" json:"name"`                       // Variable name
	Value      json.RawMessage       `db:"value" json:"value"`                     // Variable value (JSON)
	Type       string                `db:"type" json:"type"`                       // Type hint (string, number, boolean, object, array)
	Scope      WorkflowVariableScope `db:"scope" json:"scope"`                     // Variable scope
	StateName  *string               `db:"state_name" json:"state_name,omitempty"` // State if scope=state
	SetBy      *uuidv7.UUID          `db:"set_by" json:"set_by,omitempty"`         // User who set the value
	SetAt      time.Time             `db:"set_at" json:"set_at"`                   // When value was set
	CreatedAt  time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time             `db:"updated_at" json:"updated_at"`
}

// NewWorkflowVariable creates a new workflow variable
func NewWorkflowVariable(
	instanceID uuidv7.UUID,
	name string,
	value json.RawMessage,
	varType string,
	scope WorkflowVariableScope,
) *WorkflowVariable {
	now := time.Now()
	
	return &WorkflowVariable{
		ID:         uuidv7.New(),
		InstanceID: instanceID,
		Name:       name,
		Value:      value,
		Type:       varType,
		Scope:      scope,
		SetAt:      now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Update updates the variable value
func (wv *WorkflowVariable) Update(value json.RawMessage, setBy *uuidv7.UUID) {
	now := time.Now()
	wv.Value = value
	wv.SetBy = setBy
	wv.SetAt = now
	wv.UpdatedAt = now
}

// SetStateName sets the state scope
func (wv *WorkflowVariable) SetStateName(stateName string) {
	wv.StateName = &stateName
	wv.UpdatedAt = time.Now()
}

// WorkflowEventType represents the type of event
type WorkflowEventType string

const (
	WorkflowEventTypeSignal   WorkflowEventType = "signal"   // External signal
	WorkflowEventTypeMessage  WorkflowEventType = "message"  // Message received
	WorkflowEventTypeTimer    WorkflowEventType = "timer"    // Timer fired
	WorkflowEventTypeError    WorkflowEventType = "error"    // Error occurred
	WorkflowEventTypeManual   WorkflowEventType = "manual"   // Manual user action
	WorkflowEventTypeWebhook  WorkflowEventType = "webhook"  // Webhook received
	WorkflowEventTypeInternal WorkflowEventType = "internal" // Internal system event
)

// WorkflowEvent represents an event in the workflow
// Events can trigger state transitions or wake up waiting workflows
type WorkflowEvent struct {
	ID           uuidv7.UUID       `db:"id" json:"id"`
	InstanceID   uuidv7.UUID       `db:"instance_id" json:"instance_id"`                 // Parent instance
	Type         WorkflowEventType `db:"type" json:"type"`                               // Event type
	Name         string            `db:"name" json:"name"`                               // Event name (e.g., "order_paid", "approval_received")
	Payload      json.RawMessage   `db:"payload" json:"payload,omitempty"`               // Event data
	TriggeredBy  *uuidv7.UUID      `db:"triggered_by" json:"triggered_by,omitempty"`     // User who triggered
	Source       *string           `db:"source" json:"source,omitempty"`                 // Event source (webhook URL, user action, etc.)
	Processed    bool              `db:"processed" json:"processed"`                     // Whether event was processed
	ProcessedAt  *time.Time        `db:"processed_at" json:"processed_at,omitempty"`     // When event was processed
	CreatedAt    time.Time         `db:"created_at" json:"created_at"`
}

// NewWorkflowEvent creates a new workflow event
func NewWorkflowEvent(
	instanceID uuidv7.UUID,
	eventType WorkflowEventType,
	name string,
	payload json.RawMessage,
) *WorkflowEvent {
	return &WorkflowEvent{
		ID:         uuidv7.New(),
		InstanceID: instanceID,
		Type:       eventType,
		Name:       name,
		Payload:    payload,
		Processed:  false,
		CreatedAt:  time.Now(),
	}
}

// MarkAsProcessed marks the event as processed
func (we *WorkflowEvent) MarkAsProcessed() {
	now := time.Now()
	we.Processed = true
	we.ProcessedAt = &now
}

// SetTriggeredBy sets who triggered the event
func (we *WorkflowEvent) SetTriggeredBy(userID uuidv7.UUID) {
	we.TriggeredBy = &userID
}

// SetSource sets the event source
func (we *WorkflowEvent) SetSource(source string) {
	we.Source = &source
}
