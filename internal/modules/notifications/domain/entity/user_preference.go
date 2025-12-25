package entity

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// TimeOfDay represents time in HH:MM format for database storage
type TimeOfDay struct {
	time.Time
}

// NewTimeOfDay creates TimeOfDay from time.Time
func NewTimeOfDay(t time.Time) *TimeOfDay {
	return &TimeOfDay{Time: t}
}

// ParseTimeOfDay parses "HH:MM" format string
func ParseTimeOfDay(s string) (*TimeOfDay, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("15:04", s)
	if err != nil {
		return nil, fmt.Errorf("invalid time format, expected HH:MM: %w", err)
	}
	return &TimeOfDay{Time: t}, nil
}

// Value implements driver.Valuer for database storage (converts to "HH:MM")
func (t TimeOfDay) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Format("15:04"), nil
}

// Scan implements sql.Scanner for database retrieval (converts from "HH:MM")
func (t *TimeOfDay) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		parsed, err := time.Parse("15:04", v)
		if err != nil {
			return fmt.Errorf("failed to parse time of day: %w", err)
		}
		t.Time = parsed
		return nil
	case []byte:
		parsed, err := time.Parse("15:04", string(v))
		if err != nil {
			return fmt.Errorf("failed to parse time of day: %w", err)
		}
		t.Time = parsed
		return nil
	default:
		return fmt.Errorf("unsupported type for TimeOfDay: %T", value)
	}
}

// String returns "HH:MM" format
func (t TimeOfDay) String() string {
	if t.Time.IsZero() {
		return ""
	}
	return t.Format("15:04")
}

// MarshalJSON implements json.Marshaler
func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Format("15:04") + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TimeOfDay) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		return nil
	}
	// Remove quotes
	s = s[1 : len(s)-1]
	parsed, err := time.Parse("15:04", s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal time of day: %w", err)
	}
	t.Time = parsed
	return nil
}

// UserPreference represents user notification preferences
type UserPreference struct {
	ID                uuidv7.UUID    `db:"id" json:"id"`
	UserID            uuidv7.UUID    `db:"user_id" json:"user_id"`
	EmailEnabled      bool           `db:"email_enabled" json:"email_enabled"`
	SMSEnabled        bool           `db:"sms_enabled" json:"sms_enabled"`
	PushEnabled       bool           `db:"push_enabled" json:"push_enabled"`
	InAppEnabled      bool           `db:"in_app_enabled" json:"in_app_enabled"`
	SystemEnabled     bool           `db:"system_enabled" json:"system_enabled"`
	SecurityEnabled   bool           `db:"security_enabled" json:"security_enabled"`
	MarketingEnabled  bool           `db:"marketing_enabled" json:"marketing_enabled"`
	ProductEnabled    bool           `db:"product_enabled" json:"product_enabled"`
	SocialEnabled     bool           `db:"social_enabled" json:"social_enabled"`
	QuietHoursStart   *TimeOfDay     `db:"quiet_hours_start" json:"quiet_hours_start,omitempty"`
	QuietHoursEnd     *TimeOfDay     `db:"quiet_hours_end" json:"quiet_hours_end,omitempty"`
	Timezone          string         `db:"timezone" json:"timezone"`
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
}

// NewUserPreference creates default preferences for a user
func NewUserPreference(userID uuidv7.UUID) *UserPreference {
	now := time.Now()
	return &UserPreference{
		ID:               uuidv7.New(),
		UserID:           userID,
		EmailEnabled:     true, // Email enabled by default
		SMSEnabled:       false,
		PushEnabled:      true,
		InAppEnabled:     true,
		SystemEnabled:    true,  // System notifications always on
		SecurityEnabled:  true,  // Security notifications always on
		MarketingEnabled: false, // Marketing disabled by default
		ProductEnabled:   true,
		SocialEnabled:    true,
		Timezone:         "UTC", // Default timezone
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// Validate validates user preference fields
func (p *UserPreference) Validate() error {
	if p.UserID == uuidv7.Nil {
		return errors.New("user_id is required")
	}
	if p.Timezone == "" {
		return errors.New("timezone is required")
	}
	if p.QuietHoursStart != nil && p.QuietHoursEnd == nil {
		return errors.New("quiet_hours_end is required when quiet_hours_start is set")
	}
	if p.QuietHoursEnd != nil && p.QuietHoursStart == nil {
		return errors.New("quiet_hours_start is required when quiet_hours_end is set")
	}
	return nil
}

// IsChannelEnabled checks if a specific channel is enabled
func (p *UserPreference) IsChannelEnabled(channel NotificationChannel) bool {
	switch channel {
	case NotificationChannelEmail:
		return p.EmailEnabled
	case NotificationChannelSMS:
		return p.SMSEnabled
	case NotificationChannelPush:
		return p.PushEnabled
	case NotificationChannelInApp:
		return p.InAppEnabled
	default:
		return false
	}
}

// IsTypeEnabled checks if a specific notification type is enabled
func (p *UserPreference) IsTypeEnabled(notifType NotificationType) bool {
	switch notifType {
	case NotificationTypeSystem:
		return p.SystemEnabled
	case NotificationTypeSecurity:
		return p.SecurityEnabled
	case NotificationTypeMarketing:
		return p.MarketingEnabled
	case NotificationTypeProduct:
		return p.ProductEnabled
	case NotificationTypeSocial:
		return p.SocialEnabled
	default:
		return false
	}
}

// IsInQuietHours checks if current time is within quiet hours
func (p *UserPreference) IsInQuietHours(currentTime time.Time) bool {
	if p.QuietHoursStart == nil || p.QuietHoursEnd == nil {
		return false
	}

	start := p.QuietHoursStart.Hour()*60 + p.QuietHoursStart.Minute()
	end := p.QuietHoursEnd.Hour()*60 + p.QuietHoursEnd.Minute()
	current := currentTime.Hour()*60 + currentTime.Minute()

	// Handle overnight quiet hours (e.g., 22:00 to 08:00)
	if start > end {
		return current >= start || current < end
	}
	
	return current >= start && current < end
}

// SetQuietHours sets quiet hours for notifications
func (p *UserPreference) SetQuietHours(start, end time.Time) {
	p.QuietHoursStart = NewTimeOfDay(start)
	p.QuietHoursEnd = NewTimeOfDay(end)
	p.UpdatedAt = time.Now()
}

// ClearQuietHours removes quiet hours setting
func (p *UserPreference) ClearQuietHours() {
	p.QuietHoursStart = nil
	p.QuietHoursEnd = nil
	p.UpdatedAt = time.Now()
}
