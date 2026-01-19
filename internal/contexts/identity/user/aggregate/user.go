package aggregate

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	usererrors "github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
)

const (
	maxLoginAttempts = 5
	// lockoutDuration  = 30 * time.Minute
)

// User is an aggregate root representing a user account in the system
type User struct {
	aggregate.BaseAggregate

	// Identity
	Email valueobject.Email

	// Authentication
	PasswordHash string

	// Authorization
	Roles []string `json:"roles"` // Role names (e.g., ["admin", "user"])

	// Status
	Status           UserStatus
	EmailVerified    bool
	EmailVerifiedAt  *time.Time
	LastLoginAt      *time.Time
	FailedLoginCount int
	LockedUntil      *time.Time
}

// NewUser creates a new user with email and password
func NewUser(email, password string) (*User, error) {
	// Validate email
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, errors.Join(usererrors.ErrInvalidEmailFormat, err)
	}

	// Validate password
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, errors.Join(usererrors.ErrPasswordHashFailed, err)
	}

	return &User{
		BaseAggregate:    aggregate.NewBaseAggregate(),
		Email:            emailVO,
		PasswordHash:     passwordHash,
		Status:           UserStatusActive,
		EmailVerified:    false,
		FailedLoginCount: 0,
		Roles:            []string{"user"}, // Default role
	}, nil
}

// CheckPassword verifies if the provided password matches the stored hash
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// ChangePassword updates the user's password
func (u *User) ChangePassword(newPassword string) error {
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return errors.Join(usererrors.ErrPasswordHashFailed, err)
	}

	u.PasswordHash = passwordHash
	u.Touch()
	return nil
}

// VerifyEmail marks the user's email as verified
func (u *User) VerifyEmail() {
	now := time.Now()
	u.EmailVerified = true
	u.EmailVerifiedAt = &now
	u.UpdatedAt = now
}

// RecordLogin records a successful login
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.FailedLoginCount = 0
	u.LockedUntil = nil
	u.Touch()
}

// RecordFailedLogin increments failed login count and locks account if necessary
func (u *User) RecordFailedLogin() {
	u.FailedLoginCount++
	u.Touch()

	// Lock account after 5 failed attempts for 30 minutes
	if u.FailedLoginCount >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		u.LockedUntil = &lockUntil
	}
}

// IsLocked checks if the account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// UnlockAccount manually unlocks the account
func (u *User) UnlockAccount() {
	u.LockedUntil = nil
	u.FailedLoginCount = 0
	u.Touch()
}

// Activate activates the user account
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.Touch()
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.Touch()
}

// Suspend suspends the user account
func (u *User) Suspend() {
	u.Status = UserStatusSuspended
	u.Touch()
}

// Ban bans the user account
func (u *User) Ban() {
	u.Status = UserStatusBanned
	u.Touch()
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && !u.IsLocked() && u.DeletedAt == nil
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the user has any of the specified roles
func (u *User) HasAnyRole(roles []string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the user has all of the specified roles
func (u *User) HasAllRoles(roles []string) bool {
	for _, role := range roles {
		if !u.HasRole(role) {
			return false
		}
	}
	return true
}

// Validate validates the user entity
func (u *User) Validate() error {
	if u.ID == uuidv7.Nil {
		return usererrors.ErrUserIDRequired
	}

	if u.Email.Value() == "" {
		return usererrors.ErrEmailRequired
	}

	if u.PasswordHash == "" {
		return usererrors.ErrPasswordHashRequired
	}

	if !isValidStatus(u.Status) {
		return usererrors.ErrInvalidUserStatus
	}

	return nil
}

// Helper functions

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return usererrors.ErrPasswordTooShort
	}

	if len(password) > 72 {
		return usererrors.ErrPasswordTooLong
	}

	// Check for at least one digit
	hasDigit := false
	for _, c := range password {
		if c >= '0' && c <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return usererrors.ErrPasswordRequiresDigit
	}

	// Check for at least one letter
	hasLetter := false
	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return usererrors.ErrPasswordRequiresLetter
	}

	return nil
}

func isValidStatus(status UserStatus) bool {
	switch status {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended, UserStatusBanned:
		return true
	default:
		return false
	}
}
