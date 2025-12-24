package handler

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateAuditEventRequest represents request to create audit event
type CreateAuditEventRequest struct {
	UserID     uuidv7.UUID `json:"user_id" binding:"required"`
	Action     string      `json:"action" binding:"required,max=100"`
	EntityType string      `json:"entity_type" binding:"required,max=100"`
	EntityID   uuidv7.UUID `json:"entity_id" binding:"required"`
	OldData    *string     `json:"old_data,omitempty"`
	NewData    *string     `json:"new_data,omitempty"`
	IPAddress  string      `json:"ip_address,omitempty"`
	UserAgent  string      `json:"user_agent,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	Metadata   *string     `json:"metadata,omitempty"`
}

// AuditEventResponse represents audit event response
type AuditEventResponse struct {
	ID         uuidv7.UUID `json:"id"`
	UserID     uuidv7.UUID `json:"user_id"`
	Action     string      `json:"action"`
	EntityType string      `json:"entity_type"`
	EntityID   uuidv7.UUID `json:"entity_id"`
	OldData    *string     `json:"old_data,omitempty"`
	NewData    *string     `json:"new_data,omitempty"`
	IPAddress  string      `json:"ip_address"`
	UserAgent  string      `json:"user_agent"`
	RequestID  string      `json:"request_id"`
	Metadata   *string     `json:"metadata,omitempty"`
	Signature  string      `json:"signature"`
	CreatedAt  time.Time   `json:"created_at"`
}

// ListAuditEventsRequest represents request to list audit events
type ListAuditEventsRequest struct {
	UserID     *uuidv7.UUID `form:"user_id"`
	Action     *string      `form:"action"`
	EntityType *string      `form:"entity_type"`
	EntityID   *uuidv7.UUID `form:"entity_id"`
	DateFrom   *time.Time   `form:"date_from" time_format:"2006-01-02"`
	DateTo     *time.Time   `form:"date_to" time_format:"2006-01-02"`
	Page       int          `form:"page" binding:"omitempty,min=1"`
	PageSize   int          `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// VerifyAuditEventResponse represents signature verification result
type VerifyAuditEventResponse struct {
	EventID  uuidv7.UUID `json:"event_id"`
	Valid    bool        `json:"valid"`
	Verified time.Time   `json:"verified_at"`
}
