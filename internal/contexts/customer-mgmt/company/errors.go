package company

import "errors"

// Repository Errors - Data access failures

var (
	// ErrCompanyNotFound is returned when company doesn't exist
	ErrCompanyNotFound = errors.New("company not found")

	// ErrCompanyAlreadyExists is returned when company with name already exists
	ErrCompanyAlreadyExists = errors.New("company already exists")
)

// Business Logic Errors - Domain rule violations

var (
	// ErrCompanyNameRequired is returned when name is empty
	ErrCompanyNameRequired = errors.New("company name is required")

	// ErrCompanyTypeInvalid is returned when type is not valid
	ErrCompanyTypeInvalid = errors.New("invalid company type")

	// ErrCompanySizeInvalid is returned when size is not valid
	ErrCompanySizeInvalid = errors.New("invalid company size")

	// ErrEmployeeCountNegative is returned when employee count is negative
	ErrEmployeeCountNegative = errors.New("employee count cannot be negative")

	// ErrRevenueNegative is returned when revenue is negative
	ErrRevenueNegative = errors.New("revenue cannot be negative")

	// ErrCompanyCannotBeOwnParent is returned when company tries to set itself as parent
	ErrCompanyCannotBeOwnParent = errors.New("company cannot be its own parent")

	// ErrParentCompanyNotFound is returned when parent company doesn't exist
	ErrParentCompanyNotFound = errors.New("parent company not found")

	// ErrParentCompanyDeleted is returned when parent company is deleted
	ErrParentCompanyDeleted = errors.New("parent company is deleted")
)

// Technical Operation Errors - Operation wrappers

var (
	// ErrCompanyExistenceCheckFailed is returned when checking company existence fails
	ErrCompanyExistenceCheckFailed = errors.New("failed to check company existence")

	// ErrCompanyParentCheckFailed is returned when checking parent company fails
	ErrCompanyParentCheckFailed = errors.New("failed to check parent company")

	// ErrCompanyCreateFailed is returned when creation fails
	ErrCompanyCreateFailed = errors.New("failed to create company")

	// ErrCompanyUpdateFailed is returned when update fails
	ErrCompanyUpdateFailed = errors.New("failed to update company")

	// ErrCompanyDeleteFailed is returned when deletion fails
	ErrCompanyDeleteFailed = errors.New("failed to delete company")
)
