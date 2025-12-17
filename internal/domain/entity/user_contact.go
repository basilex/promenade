package entity

import (
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// ContactType represents the type of contact
type ContactType string

const (
	ContactTypeEmail    ContactType = "email"
	ContactTypePhone    ContactType = "phone"
	ContactTypeTelegram ContactType = "telegram"
	ContactTypeWhatsApp ContactType = "whatsapp"
	ContactTypeViber    ContactType = "viber"
	ContactTypeSignal   ContactType = "signal"
	ContactTypeSkype    ContactType = "skype"
	ContactTypeDiscord  ContactType = "discord"
	ContactTypeLinkedIn ContactType = "linkedin"
	ContactTypeOther    ContactType = "other"
)

// ValidContactTypes returns all valid contact types
func ValidContactTypes() []ContactType {
	return []ContactType{
		ContactTypeEmail,
		ContactTypePhone,
		ContactTypeTelegram,
		ContactTypeWhatsApp,
		ContactTypeViber,
		ContactTypeSignal,
		ContactTypeSkype,
		ContactTypeDiscord,
		ContactTypeLinkedIn,
		ContactTypeOther,
	}
}

// IsValid checks if the contact type is valid
func (ct ContactType) IsValid() bool {
	for _, valid := range ValidContactTypes() {
		if ct == valid {
			return true
		}
	}
	return false
}

// IsValidContactType checks if the contact type string is valid
func IsValidContactType(ct string) bool {
	for _, valid := range ValidContactTypes() {
		if string(valid) == ct {
			return true
		}
	}
	return false
}

// UserContact represents a user's contact method
type UserContact struct {
	ID           uuidv7.UUID `db:"id"`
	UserID       uuidv7.UUID `db:"user_id"`
	ContactType  ContactType `db:"contact_type"`
	ContactValue string      `db:"contact_value"`
	Label        *string     `db:"label"` // Optional label like "Work", "Personal"

	// Status flags
	IsVerified bool `db:"is_verified"`
	IsPrimary  bool `db:"is_primary"`
	IsActive   bool `db:"is_active"`
	IsPublic   bool `db:"is_public"`

	// Availability schedule (optional)
	AvailableFrom *time.Time     `db:"available_from"` // Start time of day
	AvailableTo   *time.Time     `db:"available_to"`   // End time of day
	AvailableDays pq.StringArray `db:"available_days"` // Days of week: monday, tuesday, etc.
	Timezone      *string        `db:"timezone"`

	// Additional metadata
	Notes *string `db:"notes"`

	// Timestamps
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Validate performs validation on the UserContact
func (c *UserContact) Validate() error {
	if c.UserID == uuidv7.Nil {
		return ErrInvalidInput
	}

	if !c.ContactType.IsValid() {
		return ErrInvalidInput
	}

	if c.ContactValue == "" {
		return ErrInvalidInput
	}

	// Validate contact value length
	if len(c.ContactValue) > 255 {
		return ErrInvalidInput
	}

	// Validate availability schedule if set
	if c.AvailableFrom != nil && c.AvailableTo != nil {
		if c.AvailableFrom.After(*c.AvailableTo) || c.AvailableFrom.Equal(*c.AvailableTo) {
			return ErrInvalidInput
		}
	}

	// Validate optional string fields length
	if c.Label != nil && len(*c.Label) > 100 {
		return ErrInvalidInput
	}

	if c.Timezone != nil && len(*c.Timezone) > 50 {
		return ErrInvalidInput
	}

	if c.Notes != nil && len(*c.Notes) > 1000 {
		return ErrInvalidInput
	}

	return nil
}

// IsAvailable checks if the contact is available at the current time
func (c *UserContact) IsAvailable(now time.Time) bool {
	// Check time of day if specified
	if c.AvailableFrom != nil && c.AvailableTo != nil {
		currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
		availableFrom := time.Date(0, 1, 1, c.AvailableFrom.Hour(), c.AvailableFrom.Minute(), 0, 0, time.UTC)
		availableTo := time.Date(0, 1, 1, c.AvailableTo.Hour(), c.AvailableTo.Minute(), 0, 0, time.UTC)

		if currentTime.Before(availableFrom) || currentTime.After(availableTo) {
			return false
		}
	}

	// Check day of week if specified
	if len(c.AvailableDays) > 0 {
		currentDay := strings.ToLower(now.Weekday().String())
		dayFound := false
		for _, day := range c.AvailableDays {
			if strings.ToLower(day) == currentDay {
				dayFound = true
				break
			}
		}
		if !dayFound {
			return false
		}
	}

	// If we get here, all checks passed
	return true
}
