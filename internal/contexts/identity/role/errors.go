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

	// ErrRoleDescriptionRequired is returned when role description is empty
	ErrRoleDescriptionRequired = errors.New("role description is required")

	// ErrRoleNameEmpty is returned when role name is empty
	ErrRoleNameEmpty = errors.New("role name cannot be empty")

	// ErrRoleNameTooShort is returned when role name is less than 2 characters
	ErrRoleNameTooShort = errors.New("role name must be at least 2 characters")

	// ErrRoleNameTooLong is returned when role name exceeds 50 characters
	ErrRoleNameTooLong = errors.New("role name must not exceed 50 characters")

	// ErrRoleNameInvalidChars is returned when role name contains invalid characters
	ErrRoleNameInvalidChars = errors.New("role name contains invalid characters")
)
