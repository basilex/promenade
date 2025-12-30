package country

import "errors"

// Domain errors for country aggregate
var (
	// ErrCountryNotFound is returned when a country cannot be found
	ErrCountryNotFound = errors.New("country not found")

	// ErrInvalidCountryCode is returned when country code validation fails
	ErrInvalidCountryCode = errors.New("invalid country code")

	// ErrInvalidPhoneCode is returned when phone code validation fails
	ErrInvalidPhoneCode = errors.New("invalid phone code")
)
