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
)
