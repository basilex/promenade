package entity

import (
	"fmt"
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
	ID           uuidv7.UUID `db:"id" validate:"required"`
	UserID       uuidv7.UUID `db:"user_id" validate:"required"`
	ContactType  ContactType `db:"contact_type" validate:"required,oneof=email phone telegram whatsapp viber signal skype discord linkedin other"`
	ContactValue string      `db:"contact_value" validate:"required,max=255"`
	Label        *string     `db:"label" validate:"omitempty,max=100"` // Optional label like "Work", "Personal"

	// Status flags
	IsVerified bool `db:"is_verified" validate:"-"`
	IsPrimary  bool `db:"is_primary" validate:"-"`
	IsActive   bool `db:"is_active" validate:"-"`
	IsPublic   bool `db:"is_public" validate:"-"`

	// Availability schedule (optional)
	AvailableFrom *time.Time     `db:"available_from" validate:"omitempty"`                      // Start time of day
	AvailableTo   *time.Time     `db:"available_to" validate:"omitempty,gtefield=AvailableFrom"` // End time of day
	AvailableDays pq.StringArray `db:"available_days" validate:"omitempty"`                      // Days of week: monday, tuesday, etc.
	Timezone      *string        `db:"timezone" validate:"omitempty,max=50"`

	// Additional metadata
	Notes *string `db:"notes" validate:"omitempty,max=1000"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" validate:"required"`
	UpdatedAt time.Time `db:"updated_at" validate:"required"`
}

// Validate performs validation on the UserContact
func (c *UserContact) Validate() error {
	if c.UserID == uuidv7.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	if !c.ContactType.IsValid() {
		return fmt.Errorf("%w: invalid contact_type", ErrInvalidInput)
	}

	if c.ContactValue == "" {
		return fmt.Errorf("%w: contact_value is required", ErrInvalidInput)
	}

	// Validate contact value length
	if len(c.ContactValue) > 255 {
		return fmt.Errorf("%w: contact_value must not exceed 255 characters", ErrInvalidInput)
	}

	// Validate availability schedule if set
	if c.AvailableFrom != nil && c.AvailableTo != nil {
		if c.AvailableFrom.After(*c.AvailableTo) || c.AvailableFrom.Equal(*c.AvailableTo) {
			return fmt.Errorf("%w: available_from must be before available_to", ErrInvalidInput)
		}
	}

	// Validate optional string fields length
	if c.Label != nil && len(*c.Label) > 100 {
		return fmt.Errorf("%w: label must not exceed 100 characters", ErrInvalidInput)
	}

	if c.Timezone != nil && len(*c.Timezone) > 50 {
		return fmt.Errorf("%w: timezone must not exceed 50 characters", ErrInvalidInput)
	}

	if c.Notes != nil && len(*c.Notes) > 1000 {
		return fmt.Errorf("%w: notes must not exceed 1000 characters", ErrInvalidInput)
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
