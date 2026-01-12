package permission

import "errors"

// Domain errors for permission aggregate
var (
	// ErrPermissionNotFound is returned when a permission cannot be found
	ErrPermissionNotFound = errors.New("permission not found")

	// ErrPermissionAlreadyExists is returned when a permission already exists
	ErrPermissionAlreadyExists = errors.New("permission already exists")

	// ErrPermissionNameExists is returned when permission name already exists
	ErrPermissionNameExists = errors.New("permission name already exists")

	// ErrResourceRequired is returned when resource is empty
	ErrResourceRequired = errors.New("resource is required")

	// ErrActionRequired is returned when action is empty
	ErrActionRequired = errors.New("action is required")

	// ErrInvalidAction is returned when action format is invalid
	ErrInvalidAction = errors.New("invalid action format")

	// ErrPermissionDescriptionRequired is returned when description is empty
	ErrPermissionDescriptionRequired = errors.New("description is required")

	// ErrResourceEmpty is returned when resource is empty
	ErrResourceEmpty = errors.New("resource cannot be empty")

	// ErrResourceTooShort is returned when resource is less than 2 characters
	ErrResourceTooShort = errors.New("resource must be at least 2 characters")

	// ErrResourceTooLong is returned when resource exceeds 50 characters
	ErrResourceTooLong = errors.New("resource must not exceed 50 characters")

	// ErrActionEmpty is returned when action is empty
	ErrActionEmpty = errors.New("action cannot be empty")

	// ErrActionTooLong is returned when action exceeds 50 characters
	ErrActionTooLong = errors.New("action must not exceed 50 characters")

	// ErrActionInvalidChars is returned when action contains invalid characters
	ErrActionInvalidChars = errors.New("action must contain only lowercase letters, numbers, underscores and hyphens")
)
