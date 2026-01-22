package subscription

import "errors"

// Repository Errors - Data access failures (technical layer)
var (
	// ErrSubscriptionNotFound is returned when subscription doesn't exist in database
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// Business Logic Errors - Domain rule violations (constructor validation)
var (
	// ErrCustomerIDRequired is returned when customer ID is missing or nil
	ErrCustomerIDRequired = errors.New("customer ID is required")

	// ErrPlanIDRequired is returned when plan ID is empty
	ErrPlanIDRequired = errors.New("plan ID is required")

	// ErrCurrencyRequired is returned when currency code is empty
	ErrCurrencyRequired = errors.New("currency is required")

	// ErrAmountMustBePositive is returned when subscription amount is zero or negative
	ErrAmountMustBePositive = errors.New("amount must be positive")

	// ErrInvalidMoney is returned when money value object creation fails
	ErrInvalidMoney = errors.New("invalid money value")
)

// Business Logic Errors - State machine violations (lifecycle transitions)
var (
	// ErrCannotActivate is returned when subscription cannot be activated from current status
	// Valid transitions: trial → active, paused → active
	ErrCannotActivate = errors.New("cannot activate subscription in current status")

	// ErrCanOnlyPauseActive is returned when attempting to pause non-active subscription
	// Valid transition: active → paused
	ErrCanOnlyPauseActive = errors.New("can only pause active subscriptions")

	// ErrCanOnlyResumePaused is returned when attempting to resume non-paused subscription
	// Valid transition: paused → active
	ErrCanOnlyResumePaused = errors.New("can only resume paused subscriptions")

	// ErrAlreadyInTerminalStatus is returned when operating on cancelled/expired subscription
	// Terminal statuses: cancelled, expired (no transitions allowed)
	ErrAlreadyInTerminalStatus = errors.New("subscription already in terminal status")

	// ErrCanOnlyRenewActive is returned when attempting to renew non-active subscription
	// Valid: active → active (with updated renewal date)
	ErrCanOnlyRenewActive = errors.New("can only renew active subscriptions")

	// ErrAlreadyExpired is returned when marking already expired subscription as expired
	ErrAlreadyExpired = errors.New("subscription already expired")
)

// Business Logic Errors - Validation rules (Validate method)
var (
	// ErrStartDateRequired is returned when start date is zero value
	ErrStartDateRequired = errors.New("start date is required")

	// ErrRenewalDateInvalid is returned when renewal date is before start date
	ErrRenewalDateInvalid = errors.New("renewal date must be after start date")
)
