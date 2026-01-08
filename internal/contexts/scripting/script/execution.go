package script

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// ScriptExecution represents a single execution of a script
// Note: This is an entity, not an aggregate root
type ScriptExecution struct {
	ID         uuidv7.UUID // Execution ID
	ScriptID   uuidv7.UUID // Script that was executed
	ScriptName string      // Script name (denormalized for convenience)

	InputParams  map[string]interface{} // Input parameters (stored as JSONB)
	OutputResult interface{}            // Execution result (stored as JSONB)
	Error        *string                // Error message if execution failed
	DurationMs   int                    // Execution duration in milliseconds

	ExecutedBy uuidv7.UUID // User who executed the script
	ExecutedAt time.Time   // When execution occurred
}

// NewScriptExecution creates a new script execution record
func NewScriptExecution(scriptID uuidv7.UUID, scriptName string, executedBy uuidv7.UUID) *ScriptExecution {
	return &ScriptExecution{
		ID:          uuidv7.New(),
		ScriptID:    scriptID,
		ScriptName:  scriptName,
		ExecutedBy:  executedBy,
		ExecutedAt:  time.Now(),
		InputParams: make(map[string]interface{}),
	}
}

// SetInput sets the input parameters for the execution
func (e *ScriptExecution) SetInput(params map[string]interface{}) {
	e.InputParams = params
}

// SetResult marks the execution as successful with a result
func (e *ScriptExecution) SetResult(result interface{}, durationMs int) {
	e.OutputResult = result
	e.DurationMs = durationMs
	e.Error = nil
}

// SetError marks the execution as failed with an error
func (e *ScriptExecution) SetError(err error, durationMs int) {
	errMsg := err.Error()
	e.Error = &errMsg
	e.DurationMs = durationMs
	e.OutputResult = nil
}

// IsSuccess returns true if execution was successful
func (e *ScriptExecution) IsSuccess() bool {
	return e.Error == nil
}

// GetErrorMessage returns the error message if execution failed
func (e *ScriptExecution) GetErrorMessage() string {
	if e.Error == nil {
		return ""
	}
	return *e.Error
}

// Validate checks if the execution is valid
func (e *ScriptExecution) Validate() error {
	if e.ScriptID == uuidv7.Nil {
		return fmt.Errorf("script ID cannot be nil")
	}
	if e.ScriptName == "" {
		return fmt.Errorf("script name cannot be empty")
	}
	if e.ExecutedBy == uuidv7.Nil {
		return fmt.Errorf("executed_by cannot be nil")
	}
	if e.DurationMs < 0 {
		return fmt.Errorf("duration cannot be negative")
	}
	return nil
}
