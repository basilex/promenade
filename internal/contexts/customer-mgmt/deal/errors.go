package deal

import "errors"

// Repository Errors - Data access failures
var (
	// ErrDealNotFound is returned when deal doesn't exist in the database
	ErrDealNotFound = errors.New("deal not found")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrDealNameEmpty is returned when deal name is empty
	ErrDealNameEmpty = errors.New("deal name cannot be empty")

	// ErrDealCustomerRequired is returned when customer ID is not provided
	ErrDealCustomerRequired = errors.New("customer ID is required")

	// ErrDealAssignedToRequired is returned when assigned_to is not provided
	ErrDealAssignedToRequired = errors.New("assigned_to is required")

	// ErrDealValueNegative is returned when deal value is negative
	ErrDealValueNegative = errors.New("deal value must be non-negative")

	// ErrDealDateInPast is returned when expected close date is in the past
	ErrDealDateInPast = errors.New("expected close date cannot be in the past")

	// ErrDealTerminalStage is returned when trying to move deal from terminal stage
	ErrDealTerminalStage = errors.New("cannot move deal from terminal stage")

	// ErrDealInvalidStageTransition is returned when stage transition is not allowed
	ErrDealInvalidStageTransition = errors.New("invalid stage transition")

	// ErrDealAlreadyWon is returned when trying to mark already won deal as won
	ErrDealAlreadyWon = errors.New("deal is already marked as won")

	// ErrDealAlreadyLost is returned when trying to mark already lost deal as lost
	ErrDealAlreadyLost = errors.New("deal is already marked as lost")

	// ErrDealCannotMarkWonAsLost is returned when trying to mark won deal as lost
	ErrDealCannotMarkWonAsLost = errors.New("cannot mark won deal as lost")

	// ErrDealCannotMarkLostAsWon is returned when trying to mark lost deal as won
	ErrDealCannotMarkLostAsWon = errors.New("cannot mark lost deal as won")

	// ErrDealLossReasonRequired is returned when loss reason is not provided
	ErrDealLossReasonRequired = errors.New("loss reason is required")

	// ErrDealProbabilityRange is returned when probability is not between 0-100
	ErrDealProbabilityRange = errors.New("probability must be between 0 and 100")

	// ErrDealClosedMutation is returned when trying to modify closed deal
	ErrDealClosedMutation = errors.New("cannot modify closed deal")

	// ErrDealSalesRepRequired is returned when sales rep ID is not provided
	ErrDealSalesRepRequired = errors.New("sales rep ID is required")
)

// Technical Operation Errors - Operation wrappers
var (
	// ErrDealCreateFailed is returned when deal creation fails
	ErrDealCreateFailed = errors.New("failed to create deal")

	// ErrDealUpdateFailed is returned when deal update fails
	ErrDealUpdateFailed = errors.New("failed to update deal")

	// ErrDealDeleteFailed is returned when deal deletion fails
	ErrDealDeleteFailed = errors.New("failed to delete deal")

	// ErrDealListFailed is returned when deal listing fails
	ErrDealListFailed = errors.New("failed to list deals")

	// ErrDealStatsFailed is returned when pipeline statistics retrieval fails
	ErrDealStatsFailed = errors.New("failed to retrieve pipeline statistics")
)
