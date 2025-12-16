package entity

import (
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
	Email           string      `db:"email" json:"email"`
	Name            string      `db:"name" json:"name"`
	Password        string      `db:"password" json:"-"` // Never expose in JSON
	Status          UserStatus  `db:"status" json:"status"`
	EmailVerifiedAt *time.Time  `db:"email_verified_at" json:"email_verified_at"`
	SuspendedReason *string     `db:"suspended_reason" json:"suspended_reason,omitempty"`
	SuspendedUntil  *time.Time  `db:"suspended_until" json:"suspended_until,omitempty"`
	LastLoginAt     *time.Time  `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
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

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}
