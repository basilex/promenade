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

	// Label validation errors
	ErrLabelRequired = errors.New("label is required")
	ErrLabelEmpty    = errors.New("label cannot be empty")

	// Type-specific errors
	ErrCannotSetEmailOnNonEmailContact     = errors.New("cannot set email on non-email contact")
	ErrCannotSetPhoneOnNonPhoneContact     = errors.New("cannot set phone on non-phone contact")
	ErrCannotSetAddressOnNonAddressContact = errors.New("cannot set address on non-address contact")

	// Validation errors
	ErrEmailRequiredForEmailContact     = errors.New("email is required for email contact")
	ErrPhoneRequiredForPhoneContact     = errors.New("phone is required for phone contact")
	ErrAddressRequiredForAddressContact = errors.New("address is required for address contact")

	// Type validation errors
	ErrUnknownContactType  = errors.New("unknown contact type")
	ErrInvalidContactType = errors.New("invalid contact type (must be email, phone, or address)")
)
