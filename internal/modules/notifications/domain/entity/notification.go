package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// JSONB represents a JSONB column type for PostgreSQL
type JSONB map[string]any

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	var result map[string]any
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = JSONB(result)
	return nil
}

// NotificationStatus represents notification delivery status
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusBounced   NotificationStatus = "bounced"
	NotificationStatusOpened    NotificationStatus = "opened"
	NotificationStatusClicked   NotificationStatus = "clicked"
)

// NotificationChannel represents delivery channel
type NotificationChannel string

const (
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelPush  NotificationChannel = "push"
	NotificationChannelInApp NotificationChannel = "in_app"
)

// NotificationType represents business notification type
type NotificationType string

const (
	NotificationTypeSystem    NotificationType = "system"
	NotificationTypeSecurity  NotificationType = "security"
	NotificationTypeMarketing NotificationType = "marketing"
	NotificationTypeProduct   NotificationType = "product"
	NotificationTypeSocial    NotificationType = "social"
)

// Notification represents a notification record
type Notification struct {
	ID        uuidv7.UUID         `db:"id" json:"id"`
	UserID    uuidv7.UUID         `db:"user_id" json:"user_id"`
	Type      NotificationType    `db:"type" json:"type"`
	Channel   NotificationChannel `db:"channel" json:"channel"`
	Status    NotificationStatus  `db:"status" json:"status"`
	Template  string              `db:"template" json:"template"`
	Subject   string              `db:"subject" json:"subject"`
	Content   string              `db:"content" json:"content"`
	Data      JSONB               `db:"data" json:"data,omitempty"` // JSONB
	SentAt    *time.Time          `db:"sent_at" json:"sent_at,omitempty"`
	OpenedAt  *time.Time          `db:"opened_at" json:"opened_at,omitempty"`
	ClickedAt *time.Time          `db:"clicked_at" json:"clicked_at,omitempty"`
	FailedAt  *time.Time          `db:"failed_at" json:"failed_at,omitempty"`
	ErrorMsg  *string             `db:"error_msg" json:"error_msg,omitempty"`
	CreatedAt time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt time.Time           `db:"updated_at" json:"updated_at"`
}

// NewNotification creates a new notification
func NewNotification(userID uuidv7.UUID, notifType NotificationType, channel NotificationChannel, template, subject, content string) *Notification {
	now := time.Now()
	return &Notification{
		ID:        uuidv7.New(),
		UserID:    userID,
		Type:      notifType,
		Channel:   channel,
		Status:    NotificationStatusPending,
		Template:  template,
		Subject:   subject,
		Content:   content,
		Data:      make(JSONB),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate validates notification fields
func (n *Notification) Validate() error {
	if n.UserID == uuidv7.Nil {
		return errors.New("user_id is required")
	}
	if n.Type == "" {
		return errors.New("type is required")
	}
	if n.Channel == "" {
		return errors.New("channel is required")
	}
	if n.Subject == "" && n.Channel == NotificationChannelEmail {
		return errors.New("subject is required for email notifications")
	}
	if n.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

// MarkAsSent marks notification as sent
func (n *Notification) MarkAsSent() {
	now := time.Now()
	n.Status = NotificationStatusSent
	n.SentAt = &now
	n.UpdatedAt = now
}

// MarkAsDelivered marks notification as delivered
func (n *Notification) MarkAsDelivered() {
	n.Status = NotificationStatusDelivered
	n.UpdatedAt = time.Now()
}

// MarkAsFailed marks notification as failed
func (n *Notification) MarkAsFailed(errorMsg string) {
	now := time.Now()
	n.Status = NotificationStatusFailed
	n.FailedAt = &now
	n.ErrorMsg = &errorMsg
	n.UpdatedAt = now
}

// MarkAsOpened marks notification as opened
func (n *Notification) MarkAsOpened() {
	now := time.Now()
	n.Status = NotificationStatusOpened
	n.OpenedAt = &now
	n.UpdatedAt = now
}

// MarkAsClicked marks notification as clicked
func (n *Notification) MarkAsClicked() {
	now := time.Now()
	n.Status = NotificationStatusClicked
	n.ClickedAt = &now
	n.UpdatedAt = now
}

// IsDelivered checks if notification was successfully delivered
func (n *Notification) IsDelivered() bool {
	return n.Status == NotificationStatusDelivered ||
		n.Status == NotificationStatusOpened ||
		n.Status == NotificationStatusClicked
}

// IsFailed checks if notification delivery failed
func (n *Notification) IsFailed() bool {
	return n.Status == NotificationStatusFailed ||
		n.Status == NotificationStatusBounced
}
