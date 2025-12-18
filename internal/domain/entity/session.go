package entity

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Session represents an active user session with refresh token
type Session struct {
	ID           uuidv7.UUID `db:"id" json:"id"`
	UserID       uuidv7.UUID `db:"user_id" json:"user_id" validate:"required"`
	RefreshToken string      `db:"refresh_token" json:"-" validate:"required,min=32"` // Hashed, never expose
	UserAgent    *string     `db:"user_agent" json:"user_agent,omitempty" validate:"omitempty,max=500"`
	IPAddress    *string     `db:"ip_address" json:"ip_address,omitempty" validate:"omitempty,ip"`
	ExpiresAt    time.Time   `db:"expires_at" json:"expires_at" validate:"required,gtefield=CreatedAt"`
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
}

// Validate validates the session entity
func (s *Session) Validate() error {
	if s.UserID == (uuidv7.UUID{}) {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if s.RefreshToken == "" {
		return fmt.Errorf("%w: refresh_token is required", ErrInvalidInput)
	}
	if len(s.RefreshToken) < 32 {
		return fmt.Errorf("%w: refresh_token must be at least 32 characters", ErrInvalidInput)
	}
	if s.ExpiresAt.Before(s.CreatedAt) {
		return fmt.Errorf("%w: expires_at must be after created_at", ErrInvalidInput)
	}
	return nil
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
