package form

import "errors"

// Repository errors
var (
	// ErrFormNotFound is returned when a form definition is not found.
	ErrFormNotFound = errors.New("form definition not found")
)

// Business logic errors
var (
	// ErrFormIDRequired is returned when form_id is missing.
	ErrFormIDRequired = errors.New("form_id is required")
	// ErrFormIDExists is returned when form_id already exists.
	ErrFormIDExists = errors.New("form_id already exists")
	// ErrEntityTypeRequired is returned when entity_type is missing.
	ErrEntityTypeRequired = errors.New("entity_type is required")
	// ErrFormNameRequired is returned when name is missing.
	ErrFormNameRequired = errors.New("name is required")
	// ErrLayoutRequired is returned when layout is missing.
	ErrLayoutRequired = errors.New("layout is required")
	// ErrFieldsRequired is returned when fields are missing.
	ErrFieldsRequired = errors.New("fields are required")
)

// Technical errors
var (
	// ErrFormCreateFailed is returned when creation fails.
	ErrFormCreateFailed = errors.New("failed to create form definition")
	// ErrFormUpdateFailed is returned when update fails.
	ErrFormUpdateFailed = errors.New("failed to update form definition")
	// ErrFormDeleteFailed is returned when delete fails.
	ErrFormDeleteFailed = errors.New("failed to delete form definition")
)
