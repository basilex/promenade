package entity

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// AuditEvent represents an immutable audit log entry
type AuditEvent struct {
	ID         uuidv7.UUID `json:"id" db:"id"`
	UserID     uuidv7.UUID `json:"user_id" db:"user_id"`
	Action     string      `json:"action" db:"action"`               // "user.created", "post.updated", "post.deleted"
	EntityType string      `json:"entity_type" db:"entity_type"`     // "user", "post", "comment"
	EntityID   uuidv7.UUID `json:"entity_id" db:"entity_id"`         // ID of affected entity
	OldData    *string     `json:"old_data,omitempty" db:"old_data"` // JSON snapshot before change
	NewData    *string     `json:"new_data,omitempty" db:"new_data"` // JSON snapshot after change
	IPAddress  string      `json:"ip_address" db:"ip_address"`
	UserAgent  string      `json:"user_agent" db:"user_agent"`
	RequestID  string      `json:"request_id" db:"request_id"`
	Signature  string      `json:"signature" db:"signature"`         // HMAC-SHA256 signature
	Metadata   *string     `json:"metadata,omitempty" db:"metadata"` // JSON - additional context
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}

// AuditEventCreate represents data for creating new audit event
type AuditEventCreate struct {
	UserID     uuidv7.UUID
	Action     string
	EntityType string
	EntityID   uuidv7.UUID
	OldData    *string
	NewData    *string
	IPAddress  string
	UserAgent  string
	RequestID  string
	Metadata   *string
}

// Validate validates audit event creation data
func (a *AuditEventCreate) Validate() error {
	if a.UserID == (uuidv7.UUID{}) {
		return ErrUserIDRequired
	}
	if a.Action == "" {
		return ErrActionRequired
	}
	if a.EntityType == "" {
		return ErrEntityTypeRequired
	}
	if a.EntityID == (uuidv7.UUID{}) {
		return ErrEntityIDRequired
	}
	return nil
}

// Domain errors
var (
	ErrUserIDRequired      = &AuditError{Code: "USER_ID_REQUIRED", Message: "User ID is required"}
	ErrActionRequired      = &AuditError{Code: "ACTION_REQUIRED", Message: "Action is required"}
	ErrEntityTypeRequired  = &AuditError{Code: "ENTITY_TYPE_REQUIRED", Message: "Entity type is required"}
	ErrEntityIDRequired    = &AuditError{Code: "ENTITY_ID_REQUIRED", Message: "Entity ID is required"}
	ErrAuditEventNotFound  = &AuditError{Code: "AUDIT_EVENT_NOT_FOUND", Message: "Audit event not found"}
	ErrInvalidSignature    = &AuditError{Code: "INVALID_SIGNATURE", Message: "Invalid audit event signature"}
	ErrAuditEventImmutable = &AuditError{Code: "AUDIT_EVENT_IMMUTABLE", Message: "Audit events are immutable"}
)

// AuditError represents audit domain error
type AuditError struct {
	Code    string
	Message string
}

func (e *AuditError) Error() string {
	return e.Message
}
