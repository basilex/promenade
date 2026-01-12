package user

import "errors"

// ============================================================================
// Repository Errors - Data access failures
// ============================================================================
var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
)

// ============================================================================
// Business Logic Errors - Domain rule violations
// ============================================================================
var (
	// Authentication & Authorization
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

	// Password Validation
	// ErrInvalidPassword is returned when password validation fails (generic)
	ErrInvalidPassword = errors.New("invalid password")

	// ErrPasswordTooShort is returned when password is less than 8 characters
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")

	// ErrPasswordTooLong is returned when password exceeds 72 characters
	ErrPasswordTooLong = errors.New("password must not exceed 72 characters")

	// ErrPasswordRequiresDigit is returned when password doesn't contain a digit
	ErrPasswordRequiresDigit = errors.New("password must contain at least one digit")

	// ErrPasswordRequiresLetter is returned when password doesn't contain a letter
	ErrPasswordRequiresLetter = errors.New("password must contain at least one letter")

	// Entity Validation
	// ErrUserIDRequired is returned when user ID is empty
	ErrUserIDRequired = errors.New("user ID is required")

	// ErrEmailRequired is returned when email is empty
	ErrEmailRequired = errors.New("email is required")

	// ErrPasswordHashRequired is returned when password hash is empty
	ErrPasswordHashRequired = errors.New("password hash is required")

	// ErrInvalidUserStatus is returned when user status is not valid
	ErrInvalidUserStatus = errors.New("invalid user status")

	// ErrInvalidEmailFormat is returned when email format is invalid
	ErrInvalidEmailFormat = errors.New("invalid email format")
)

// ============================================================================
// Technical Operation Errors - Operation wrappers
// ============================================================================
var (
	// ErrEmailCheckFailed is returned when checking email existence fails
	ErrEmailCheckFailed = errors.New("failed to check email existence")

	// ErrUserCreateFailed is returned when user creation fails
	ErrUserCreateFailed = errors.New("failed to create user")

	// ErrUserSaveFailed is returned when saving user to repository fails
	ErrUserSaveFailed = errors.New("failed to save user")

	// ErrUserGetFailed is returned when retrieving user fails
	ErrUserGetFailed = errors.New("failed to get user")

	// ErrUserUpdateFailed is returned when updating user fails
	ErrUserUpdateFailed = errors.New("failed to update user")

	// ErrPasswordHashFailed is returned when password hashing fails
	ErrPasswordHashFailed = errors.New("failed to hash password")

	// ErrAuthenticateFailed is returned when authentication operation fails
	ErrAuthenticateFailed = errors.New("failed to authenticate user")

	// ErrChangePasswordFailed is returned when password change fails
	ErrChangePasswordFailed = errors.New("failed to change password")

	// ErrVerifyEmailFailed is returned when email verification fails
	ErrVerifyEmailFailed = errors.New("failed to verify email")

	// ErrSuspendUserFailed is returned when user suspension fails
	ErrSuspendUserFailed = errors.New("failed to suspend user")

	// ErrBanUserFailed is returned when user ban fails
	ErrBanUserFailed = errors.New("failed to ban user")

	// ErrActivateUserFailed is returned when user activation fails
	ErrActivateUserFailed = errors.New("failed to activate user")

	// ErrUnlockUserFailed is returned when user unlock fails
	ErrUnlockUserFailed = errors.New("failed to unlock user")

	// ErrListUsersFailed is returned when listing users fails
	ErrListUsersFailed = errors.New("failed to list users")
)
