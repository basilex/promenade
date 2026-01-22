package costcenter

import "errors"

var (
	// Validation errors
	ErrCodeRequired      = errors.New("cost center code is required")
	ErrNameRequired      = errors.New("cost center name is required")
	ErrInvalidCenterType = errors.New("invalid cost center type")
	ErrCannotBeOwnParent = errors.New("cost center cannot be its own parent")

	// Business logic errors
	ErrCostCenterNotFound = errors.New("cost center not found")
	ErrCodeAlreadyExists  = errors.New("cost center code already exists")
	ErrHasChildren        = errors.New("cannot delete cost center with children")
)
