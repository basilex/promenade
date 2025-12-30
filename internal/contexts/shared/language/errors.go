package language

import "errors"

// Domain errors for language aggregate
var (
	// ErrLanguageNotFound is returned when a language cannot be found
	ErrLanguageNotFound = errors.New("language not found")

	// ErrInvalidLanguageCode is returned when language code validation fails
	ErrInvalidLanguageCode = errors.New("invalid language code")
)
