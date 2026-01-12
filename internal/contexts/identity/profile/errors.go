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

	// Display Name validation errors
	ErrDisplayNameRequired = errors.New("display name is required")
	ErrDisplayNameTooShort = errors.New("display name must be at least 2 characters")
	ErrDisplayNameTooLong  = errors.New("display name must not exceed 100 characters")

	// Bio validation errors
	ErrBioTooLong = errors.New("bio must not exceed 500 characters")

	// Avatar validation errors
	ErrAvatarInvalidURL = errors.New("avatar URL must be a valid HTTP(S) URL")

	// Personal info validation errors
	ErrFirstNameTooLong  = errors.New("first name must not exceed 50 characters")
	ErrLastNameTooLong   = errors.New("last name must not exceed 50 characters")
	ErrMiddleNameTooLong = errors.New("middle name must not exceed 50 characters")

	// Date of birth validation errors
	ErrAgeTooYoung       = errors.New("user must be at least 13 years old")
	ErrDateOfBirthInvalid = errors.New("invalid date of birth")

	// Localization validation errors
	ErrLanguageInvalid = errors.New("language must be ISO 639-1 code (2 letters)")
	ErrCountryInvalid  = errors.New("country must be ISO 3166-1 alpha-2 code (2 letters)")

	// Social links validation errors
	ErrSocialLinkInvalidURL = errors.New("social link URL must be a valid HTTP(S) URL")

	// Validate method errors
	ErrProfileIDRequired = errors.New("profile ID is required")
	ErrUserIDRequired    = errors.New("user ID is required")
)
