package script

import "errors"

// ============================================================================
// Repository Errors - Data access failures
// ============================================================================

var (
	// ErrScriptNotFound is returned when a script doesn't exist
	ErrScriptNotFound = errors.New("script not found")

	// ErrScriptExecutionNotFound is returned when a script execution doesn't exist
	ErrScriptExecutionNotFound = errors.New("script execution not found")
)

// ============================================================================
// Business Logic Errors - Domain rule violations
// ============================================================================

var (
	// ErrScriptNameEmpty is returned when script name is empty
	ErrScriptNameEmpty = errors.New("script name cannot be empty")

	// ErrScriptCodeEmpty is returned when script code is empty
	ErrScriptCodeEmpty = errors.New("script code cannot be empty")

	// ErrScriptVersionInvalid is returned when script version is less than 1
	ErrScriptVersionInvalid = errors.New("script version must be at least 1")

	// ErrScriptStatusInvalid is returned when script status is invalid
	ErrScriptStatusInvalid = errors.New("invalid script status")

	// ErrScriptNotActive is returned when attempting to execute non-active script
	ErrScriptNotActive = errors.New("script is not active")

	// ErrScriptArchived is returned when attempting to modify archived script
	ErrScriptArchived = errors.New("cannot modify archived script")

	// ErrScriptCannotActivateArchived is returned when attempting to activate archived script
	ErrScriptCannotActivateArchived = errors.New("cannot activate archived script")

	// ErrScriptCannotDeactivateArchived is returned when attempting to deactivate archived script
	ErrScriptCannotDeactivateArchived = errors.New("cannot deactivate archived script")

	// ErrScriptSyntaxInvalid is returned when LUA script syntax is invalid
	ErrScriptSyntaxInvalid = errors.New("script syntax validation failed")

	// ErrScriptExecutionScriptIDNil is returned when script ID is nil for execution
	ErrScriptExecutionScriptIDNil = errors.New("script ID cannot be nil")

	// ErrScriptExecutionNameEmpty is returned when script name is empty for execution
	ErrScriptExecutionNameEmpty = errors.New("script name cannot be empty")

	// ErrScriptExecutionExecutedByNil is returned when executed_by is nil
	ErrScriptExecutionExecutedByNil = errors.New("executed_by cannot be nil")

	// ErrScriptExecutionDurationNegative is returned when duration is negative
	ErrScriptExecutionDurationNegative = errors.New("duration cannot be negative")
)

// ============================================================================
// Technical Operation Errors - Operation wrappers
// ============================================================================

var (
	// ErrScriptCreateFailed is returned when script creation fails
	ErrScriptCreateFailed = errors.New("failed to create script")

	// ErrScriptUpdateFailed is returned when script update fails
	ErrScriptUpdateFailed = errors.New("failed to update script")

	// ErrScriptDeleteFailed is returned when script deletion fails
	ErrScriptDeleteFailed = errors.New("failed to delete script")

	// ErrScriptExecutionFailed is returned when script execution fails
	ErrScriptExecutionFailed = errors.New("script execution failed")

	// ErrScriptExecutionSaveFailed is returned when saving execution record fails
	ErrScriptExecutionSaveFailed = errors.New("failed to save script execution")

	// ErrScriptListFailed is returned when listing scripts fails
	ErrScriptListFailed = errors.New("failed to list scripts")
)
