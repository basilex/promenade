package role

import "errors"

// Domain errors for role aggregate
var (
	// ErrRoleNotFound is returned when a role cannot be found
	ErrRoleNotFound = errors.New("role not found")

	// ErrRoleAlreadyExists is returned when a role with the same name already exists
	ErrRoleAlreadyExists = errors.New("role already exists")

	// ErrRoleNameRequired is returned when role name is empty
	ErrRoleNameRequired = errors.New("role name is required")

	// ErrRoleDisplayRequired is returned when role display name is empty
	ErrRoleDisplayRequired = errors.New("role display name is required")

	// ErrRoleNameExists is returned when role name already exists
	ErrRoleNameExists = errors.New("role name already exists")

	// ErrCannotDeleteSystem is returned when trying to delete a system role
	ErrCannotDeleteSystem = errors.New("cannot delete system role")
)
