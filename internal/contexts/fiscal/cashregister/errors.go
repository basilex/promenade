package cashregister

import "errors"

// Repository Errors - Data access failures
var (
	// ErrCashRegisterNotFound is returned when cash register doesn't exist
	ErrCashRegisterNotFound = errors.New("cash register not found")

	// ErrCashRegisterFiscalNumberExists is returned when fiscal number already exists
	ErrCashRegisterFiscalNumberExists = errors.New("fiscal number already exists")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrLocationIDRequired is returned when location ID is missing
	ErrLocationIDRequired = errors.New("location ID is required")

	// ErrNameRequired is returned when name is missing
	ErrNameRequired = errors.New("name is required")

	// ErrFiscalNumberRequired is returned when fiscal number is missing
	ErrFiscalNumberRequired = errors.New("fiscal number is required")

	// ErrOrganizationIDRequired is returned when organization ID is missing
	ErrOrganizationIDRequired = errors.New("organization ID is required")

	// ErrModelRequired is returned when model is missing
	ErrModelRequired = errors.New("model is required")

	// ErrCreatedByRequired is returned when created by user ID is missing
	ErrCreatedByRequired = errors.New("created by user ID is required")

	// ErrLicenseKeyRequired is returned when license key is missing
	ErrLicenseKeyRequired = errors.New("license key is required")

	// ErrUnsupportedProvider is returned for unsupported providers
	ErrUnsupportedProvider = errors.New("unsupported provider (MVP: checkbox only)")

	// ErrCredentialsRequired is returned when API credentials are missing
	ErrCredentialsRequired = errors.New("API credentials are required")

	// ErrNotRegistered is returned when cash register is not registered
	ErrNotRegistered = errors.New("cash register not registered with fiscal data")

	// ErrCashRegisterAlreadyActive is returned when activating already active register
	ErrCashRegisterAlreadyActive = errors.New("cash register is already active")

	// ErrCashRegisterAlreadyInactive is returned when deactivating already inactive register
	ErrCashRegisterAlreadyInactive = errors.New("cash register is already inactive")

	// ErrInvalidProvider is returned for invalid provider type
	ErrInvalidProvider = errors.New("invalid provider type")

	// ErrAPICredentialsRequired is returned when API credentials are missing
	ErrAPICredentialsRequired = errors.New("API credentials are required")

	// ErrCashRegisterBlocked is returned when operating on blocked register
	ErrCashRegisterBlocked = errors.New("cash register is blocked")

	// ErrCashRegisterDeregistered is returned when operating on deregistered register
	ErrCashRegisterDeregistered = errors.New("cash register is deregistered")

	// ErrCashRegisterInactive is returned when operating on inactive register
	ErrCashRegisterInactive = errors.New("cash register is inactive")

	// ErrShiftAlreadyOpen is returned when trying to open shift while one is already open
	ErrShiftAlreadyOpen = errors.New("shift already open")

	// ErrNoActiveShift is returned when trying to close shift but none is open
	ErrNoActiveShift = errors.New("no active shift")
)

// Technical Operation Errors - Operation wrappers
var (
	// ErrCashRegisterCreateFailed is returned when creation fails
	ErrCashRegisterCreateFailed = errors.New("failed to create cash register")

	// ErrCashRegisterUpdateFailed is returned when update fails
	ErrCashRegisterUpdateFailed = errors.New("failed to update cash register")

	// ErrCashRegisterDeleteFailed is returned when deletion fails
	ErrCashRegisterDeleteFailed = errors.New("failed to delete cash register")
)
