package user

import "errors"

// Domain errors for user aggregate
var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailAlreadyExists is returned when a user with the same email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidCredentials is returned when authentication fails
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrAccountLocked is returned when trying to authenticate with a locked account
	ErrAccountLocked = errors.New("account is locked")

	// ErrAccountNotActive is returned when trying to authenticate with an inactive account
	ErrAccountNotActive = errors.New("account is not active")

	// ErrEmailNotVerified is returned when email verification is required
	ErrEmailNotVerified = errors.New("email is not verified")

	// ErrInvalidPassword is returned when password validation fails
	ErrInvalidPassword = errors.New("invalid password")
)
