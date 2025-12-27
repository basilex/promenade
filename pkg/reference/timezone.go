package reference

import (
	"fmt"
	"strings"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Timezone represents an IANA timezone.
// This is a Value Object shared across all contexts.
type Timezone struct {
	ID           uuidv7.UUID `db:"id"`
	Name         string      `db:"name"`
	Abbreviation string      `db:"abbreviation"`
	UTCOffset    int         `db:"utc_offset"`
	IsActive     bool        `db:"is_active"`
}

// NewTimezone creates a new Timezone value object with validation.
func NewTimezone(name, abbreviation string) (Timezone, error) {
	name = strings.TrimSpace(name)
	abbreviation = strings.ToUpper(strings.TrimSpace(abbreviation))

	if name == "" {
		return Timezone{}, fmt.Errorf("timezone name is required")
	}

	// Validate that timezone name exists in IANA database
	loc, err := time.LoadLocation(name)
	if err != nil {
		return Timezone{}, fmt.Errorf("invalid timezone name: %w", err)
	}

	// Calculate current UTC offset
	_, offset := time.Now().In(loc).Zone()

	if abbreviation == "" {
		abbreviation = name
	}

	return Timezone{
		ID:           uuidv7.New(),
		Name:         name,
		Abbreviation: abbreviation,
		UTCOffset:    offset,
		IsActive:     true,
	}, nil
}

// String returns the timezone name.
func (t Timezone) String() string {
	return t.Name
}

// Equals checks if two timezones are the same (by IANA name).
func (t Timezone) Equals(other Timezone) bool {
	return t.Name == other.Name
}

// FormatOffset returns the UTC offset in human-readable format (+01:00, -05:00, etc.)
func (t Timezone) FormatOffset() string {
	hours := t.UTCOffset / 3600
	minutes := (t.UTCOffset % 3600) / 60

	sign := "+"
	if hours < 0 {
		sign = "-"
		hours = -hours
	}

	return fmt.Sprintf("UTC%s%02d:%02d", sign, hours, minutes)
}
