package account

import "errors"

// Validation Errors - Input validation failures
var (
	// ErrAccountCodeEmpty is returned when account code is empty
	ErrAccountCodeEmpty = errors.New("account code cannot be empty")

	// ErrAccountNameEmpty is returned when account name is empty
	ErrAccountNameEmpty = errors.New("account name cannot be empty")

	// ErrAccountInvalidType is returned when account type is invalid
	ErrAccountInvalidType = errors.New("invalid account type")

	// ErrAccountInvalidCurrency is returned when currency code is invalid
	ErrAccountInvalidCurrency = errors.New("invalid currency code")

	// ErrAccountInvalidLevel is returned when account level is invalid
	ErrAccountInvalidLevel = errors.New("account level must be >= 1")

	// ErrAccountParentNotFound is returned when parent account doesn't exist
	ErrAccountParentNotFound = errors.New("parent account not found")

	// ErrAccountCircularReference is returned when account references itself as parent
	ErrAccountCircularReference = errors.New("circular parent reference detected")
)

// Not Found Errors - Entity lookup failures
var (
	// ErrAccountNotFound is returned when an account cannot be found
	ErrAccountNotFound = errors.New("account not found")
)

// Already Exists Errors - Uniqueness constraint violations
var (
	// ErrAccountCodeAlreadyExists is returned when account code already exists
	ErrAccountCodeAlreadyExists = errors.New("account code already exists")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrAccountHasChildren is returned when trying to deactivate account with children
	ErrAccountHasChildren = errors.New("cannot deactivate account with child accounts")

	// ErrAccountHasTransactions is returned when trying to delete account with transactions
	ErrAccountHasTransactions = errors.New("cannot delete account with existing transactions")

	// ErrAccountInactive is returned when trying to use inactive account
	ErrAccountInactive = errors.New("account is inactive")
)

// Technical Operation Errors - Infrastructure failures
var (
	// ErrAccountCreateFailed is returned when account creation fails
	ErrAccountCreateFailed = errors.New("failed to create account")

	// ErrAccountGetFailed is returned when account retrieval fails
	ErrAccountGetFailed = errors.New("failed to retrieve account")

	// ErrAccountUpdateFailed is returned when account update fails
	ErrAccountUpdateFailed = errors.New("failed to update account")

	// ErrAccountDeleteFailed is returned when account deletion fails
	ErrAccountDeleteFailed = errors.New("failed to delete account")
)
