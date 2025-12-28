package user

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

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
	lockoutDuration  = 30 * time.Minute
)

// User is an aggregate root representing a user account in the system
type User struct {
	// Identity
	ID    uuidv7.UUID
	Email valueobject.Email

	// Authentication
	PasswordHash string

	// Status
	Status           UserStatus
	EmailVerified    bool
	EmailVerifiedAt  *time.Time
	LastLoginAt      *time.Time
	FailedLoginCount int
	LockedUntil      *time.Time

	// Lifecycle
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewUser creates a new user with email and password
func NewUser(email, password string) (*User, error) {
	// Validate email
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Validate password
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	return &User{
		ID:               uuidv7.New(),
		Email:            emailVO,
		PasswordHash:     passwordHash,
		Status:           UserStatusActive,
		EmailVerified:    false,
		FailedLoginCount: 0,
		CreatedAt:        now,
		UpdatedAt:        now,
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
		return fmt.Errorf("failed to hash password: %w", err)
	}

	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now()
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
	u.UpdatedAt = now
}

// RecordFailedLogin increments failed login count and locks account if necessary
func (u *User) RecordFailedLogin() {
	u.FailedLoginCount++
	u.UpdatedAt = time.Now()

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
	u.UpdatedAt = time.Now()
}

// Activate activates the user account
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now()
}

// Suspend suspends the user account
func (u *User) Suspend() {
	u.Status = UserStatusSuspended
	u.UpdatedAt = time.Now()
}

// Ban bans the user account
func (u *User) Ban() {
	u.Status = UserStatusBanned
	u.UpdatedAt = time.Now()
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && !u.IsLocked() && u.DeletedAt == nil
}

// Validate validates the user entity
func (u *User) Validate() error {
	if u.ID == uuidv7.Nil {
		return fmt.Errorf("user ID is required")
	}

	if u.Email.Value() == "" {
		return fmt.Errorf("email is required")
	}

	if u.PasswordHash == "" {
		return fmt.Errorf("password hash is required")
	}

	if !isValidStatus(u.Status) {
		return fmt.Errorf("invalid user status: %s", u.Status)
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
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 72 {
		return fmt.Errorf("password must not exceed 72 characters")
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
		return fmt.Errorf("password must contain at least one digit")
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
		return fmt.Errorf("password must contain at least one letter")
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
