package timezone

import (
	"errors"
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Common errors
var (
	ErrNotFound = errors.New("timezone not found")
)

// Timezone represents an IANA timezone (Aggregate Root in Shared Context)
type Timezone struct {
	ID           uuidv7.UUID `db:"id"`
	Name         string      `db:"name"`
	Abbreviation string      `db:"abbreviation"`
	UTCOffset    int         `db:"utc_offset"` // UTC offset in seconds
	IsActive     bool        `db:"is_active"`
}

// NewTimezone creates a new Timezone entity with validation
func NewTimezone(name, abbreviation string, utcOffsetSeconds int) (*Timezone, error) {
	name = strings.TrimSpace(name)
	abbreviation = strings.ToUpper(strings.TrimSpace(abbreviation))

	if name == "" {
		return nil, fmt.Errorf("timezone name is required")
	}

	if abbreviation == "" {
		return nil, fmt.Errorf("timezone abbreviation is required")
	}

	// UTCOffset validation (-12 hours to +14 hours in seconds)
	if utcOffsetSeconds < -43200 || utcOffsetSeconds > 50400 {
		return nil, fmt.Errorf("UTC offset must be between -12h and +14h (in seconds)")
	}

	return &Timezone{
		ID:           uuidv7.New(),
		Name:         name,
		Abbreviation: abbreviation,
		UTCOffset:    utcOffsetSeconds,
		IsActive:     true,
	}, nil
}

// Validate validates timezone data
func (t *Timezone) Validate() error {
	name := strings.TrimSpace(t.Name)
	if name == "" {
		return fmt.Errorf("timezone name is required")
	}
	abbreviation := strings.TrimSpace(t.Abbreviation)
	if abbreviation == "" {
		return fmt.Errorf("timezone abbreviation is required")
	}
	// UTCOffset validation (-12 hours to +14 hours in seconds)
	if t.UTCOffset < -43200 || t.UTCOffset > 50400 {
		return fmt.Errorf("UTC offset must be between -12h and +14h (in seconds)")
	}
	return nil
}

// String returns the timezone name
func (t *Timezone) String() string {
	return t.Name
}

// ParseUUID parses a string UUID and returns error on failure
func ParseUUID(s string) (uuidv7.UUID, error) {
return uuidv7.Parse(s)
}
