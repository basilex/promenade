package contact

import "errors"

// Domain errors for contact aggregate
var (
	// ErrContactNotFound is returned when a contact cannot be found
	ErrContactNotFound = errors.New("contact not found")

	// ErrContactAlreadyExists is returned when a contact already exists
	ErrContactAlreadyExists = errors.New("contact already exists")

	// ErrPrimaryContactExists is returned when trying to create a primary contact when one already exists
	ErrPrimaryContactExists = errors.New("primary contact already exists for this type")

	// ErrInvalidEmail is returned when email validation fails
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidPhone is returned when phone validation fails
	ErrInvalidPhone = errors.New("invalid phone format")

	// ErrInvalidAddress is returned when address validation fails
	ErrInvalidAddress = errors.New("invalid address format")
)
