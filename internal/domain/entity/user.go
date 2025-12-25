package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"golang.org/x/crypto/bcrypt"
)

// UserStatus represents the user account status lifecycle
type UserStatus string

const (
	UserStatusUnverified UserStatus = "unverified" // Registered but email not verified
	UserStatusActive     UserStatus = "active"     // Email verified and account active
	UserStatusSuspended  UserStatus = "suspended"  // Temporarily blocked (can be reactivated)
	UserStatusBanned     UserStatus = "banned"     // Permanently blocked
	UserStatusInactive   UserStatus = "inactive"   // Deactivated by user (can be reactivated)
)

// User represents a user account in the system
type User struct {
	ID              uuidv7.UUID `db:"id" json:"id"`
	Email           string      `db:"email" json:"email" validate:"required,email,max=255"`
	Name            string      `db:"name" json:"name" validate:"required,min=2,max=255"`
	Password        string      `db:"password" json:"-" validate:"required,min=8,max=255"` // Never expose in JSON
	Status          UserStatus  `db:"status" json:"status" validate:"required,oneof=unverified active suspended banned inactive"`
	EmailVerifiedAt *time.Time  `db:"email_verified_at" json:"email_verified_at"`
	SuspendedReason *string     `db:"suspended_reason" json:"suspended_reason,omitempty" validate:"omitempty,max=500"`
	SuspendedUntil  *time.Time  `db:"suspended_until" json:"suspended_until,omitempty" validate:"omitempty,gtefield=CreatedAt"`
	LastLoginAt     *time.Time  `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at" validate:"gtefield=CreatedAt"`
}

// HashPassword hashes the user's password using bcrypt
func (u *User) HashPassword(password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedBytes)
	return nil
}

// CheckPassword verifies if the provided password matches the hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsActive checks if user can login (active or unverified)
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive || u.Status == UserStatusUnverified
}

// IsEmailVerified checks if user's email is verified
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// CanLogin checks if user can login (not suspended, banned, or inactive)
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive || u.Status == UserStatusUnverified
}

// Activate sets user status to active and marks email as verified
func (u *User) Activate() {
	u.Status = UserStatusActive
	now := time.Now()
	u.EmailVerifiedAt = &now
}

// Suspend temporarily blocks the user
func (u *User) Suspend(reason string, until *time.Time) {
	u.Status = UserStatusSuspended
	u.SuspendedReason = &reason
	u.SuspendedUntil = until
}

// Ban permanently blocks the user
func (u *User) Ban(reason string) {
	u.Status = UserStatusBanned
	u.SuspendedReason = &reason
}

// Deactivate sets user status to inactive
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
}

// Reactivate sets user status back to active
func (u *User) Reactivate() {
	u.Status = UserStatusActive
	u.SuspendedReason = nil
	u.SuspendedUntil = nil
}

// Validate validates the user entity
func (u *User) Validate() error {
	if u.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	// Basic email format validation
	if !strings.Contains(u.Email, "@") || !strings.Contains(u.Email, ".") {
		return fmt.Errorf("%w: invalid email format", ErrInvalidInput)
	}
	if len(u.Email) > 255 {
		return fmt.Errorf("%w: email must not exceed 255 characters", ErrInvalidInput)
	}
	if u.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if len(u.Name) < 2 || len(u.Name) > 255 {
		return fmt.Errorf("%w: name must be between 2 and 255 characters", ErrInvalidInput)
	}
	if u.Password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidInput)
	}
	if u.UpdatedAt.Before(u.CreatedAt) {
		return fmt.Errorf("%w: updated_at cannot be before created_at", ErrInvalidInput)
	}
	return nil
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}
