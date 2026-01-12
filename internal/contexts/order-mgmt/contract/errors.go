package contract

import "errors"

// Repository Errors - data access failures
var (
	// ErrContractNotFound is returned when contract doesn't exist
	ErrContractNotFound = errors.New("contract not found")
)

// Business Logic Errors - domain rule violations (State Machine)
var (
	// ErrCannotSubmitNonDraft is returned when trying to submit non-draft contract
	ErrCannotSubmitNonDraft = errors.New("can only submit draft contracts for signature")

	// ErrCannotSignNonPending is returned when trying to sign non-pending contract
	ErrCannotSignNonPending = errors.New("can only sign contracts in pending_signature status")

	// ErrCannotCompleteNonActive is returned when trying to complete non-active contract
	ErrCannotCompleteNonActive = errors.New("can only complete active contracts")

	// ErrCannotTerminateNonActive is returned when trying to terminate non-active contract
	ErrCannotTerminateNonActive = errors.New("can only terminate active contracts")

	// ErrCannotRenewInactiveContract is returned when trying to renew non-active/completed contract
	ErrCannotRenewInactiveContract = errors.New("can only renew active or completed contracts")

	// ErrCannotSetExpirationForActive is returned when trying to set expiration for non-draft/pending contract
	ErrCannotSetExpirationForActive = errors.New("can only set expiration date for draft or pending contracts")
)

// Technical Operation Errors - operation wrappers
var (
	// ErrQueryFailed is returned when database query fails
	ErrQueryFailed = errors.New("failed to query contracts")

	// ErrCreateFailed is returned when contract creation fails
	ErrCreateFailed = errors.New("failed to create contract")

	// ErrUpdateFailed is returned when contract update fails
	ErrUpdateFailed = errors.New("failed to update contract")

	// ErrDeleteFailed is returned when contract deletion fails
	ErrDeleteFailed = errors.New("failed to delete contract")
)

// Validation Errors - input validation failures
var (
	// ErrOrderIDRequired is returned when order_id is missing
	ErrOrderIDRequired = errors.New("order_id is required")

	// ErrCustomerIDRequired is returned when customer_id is missing
	ErrCustomerIDRequired = errors.New("customer_id is required")

	// ErrTermsRequired is returned when terms are missing
	ErrTermsRequired = errors.New("terms are required")

	// ErrStatusRequired is returned when status is missing
	ErrStatusRequired = errors.New("status is required")

	// ErrSignerNameRequired is returned when signer name is missing
	ErrSignerNameRequired = errors.New("signer name is required")

	// ErrSignerEmailRequired is returned when signer email is missing
	ErrSignerEmailRequired = errors.New("signer email is required")

	// ErrTerminationReasonRequired is returned when termination reason is missing
	ErrTerminationReasonRequired = errors.New("termination reason is required")

	// ErrExpirationInPast is returned when expiration date is in the past
	ErrExpirationInPast = errors.New("expiration date must be in the future")

	// ErrInvalidDaysValue is returned when days parameter is invalid
	ErrInvalidDaysValue = errors.New("days must be greater than 0")
)
