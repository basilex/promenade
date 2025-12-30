package profile

import "errors"

// Domain errors for profile aggregate
var (
	// ErrProfileNotFound is returned when a profile cannot be found
	ErrProfileNotFound = errors.New("profile not found")

	// ErrProfileAlreadyExists is returned when a profile already exists for a user
	ErrProfileAlreadyExists = errors.New("profile already exists for user")

	// ErrInvalidDisplayName is returned when display name validation fails
	ErrInvalidDisplayName = errors.New("invalid display name")

	// ErrInvalidBio is returned when bio validation fails (e.g., too long)
	ErrInvalidBio = errors.New("invalid bio")

	// ErrInvalidURL is returned when URL validation fails
	ErrInvalidURL = errors.New("invalid URL format")
)
