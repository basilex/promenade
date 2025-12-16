package entity

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Session represents an active user session with refresh token
type Session struct {
	ID           uuidv7.UUID `db:"id" json:"id"`
	UserID       uuidv7.UUID `db:"user_id" json:"user_id"`
	RefreshToken string      `db:"refresh_token" json:"-"` // Hashed, never expose
	UserAgent    *string     `db:"user_agent" json:"user_agent,omitempty"`
	IPAddress    *string     `db:"ip_address" json:"ip_address,omitempty"`
	ExpiresAt    time.Time   `db:"expires_at" json:"expires_at"`
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
}

// IsExpired checks if session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
