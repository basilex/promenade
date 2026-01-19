package aggregate

import (
	"strings"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"

	timezoneerrors "github.com/basilex/promenade/internal/contexts/shared/timezone"
)

// Timezone represents an IANA timezone (Aggregate Root in Shared Context)
type Timezone struct {
	aggregate.BaseAggregate

	Name         string `db:"name"`
	Abbreviation string `db:"abbreviation"`
	UTCOffset    int    `db:"utc_offset"` // UTC offset in seconds
	CountryCode  string `db:"country_code"`
	DSTOffset    *int   `db:"dst_offset"`
	DisplayName  string `db:"display_name"`
	IsActive     bool   `db:"is_active"`
}

// NewTimezone creates a new Timezone entity with validation
func NewTimezone(name, abbreviation string, utcOffsetSeconds int) (*Timezone, error) {
	name = strings.TrimSpace(name)
	abbreviation = strings.ToUpper(strings.TrimSpace(abbreviation))

	if name == "" {
		return nil, timezoneerrors.ErrTimezoneNameRequired
	}

	if abbreviation == "" {
		return nil, timezoneerrors.ErrTimezoneAbbreviationRequired
	}

	// UTCOffset validation (-12 hours to +14 hours in seconds)
	if utcOffsetSeconds < -43200 || utcOffsetSeconds > 50400 {
		return nil, timezoneerrors.ErrUTCOffsetOutOfRange
	}

	return &Timezone{
		BaseAggregate: aggregate.NewBaseAggregate(),
		Name:          name,
		Abbreviation:  abbreviation,
		UTCOffset:     utcOffsetSeconds,
		IsActive:      true,
	}, nil
}

// Validate validates timezone data
func (t *Timezone) Validate() error {
	name := strings.TrimSpace(t.Name)
	if name == "" {
		return timezoneerrors.ErrTimezoneNameRequired
	}
	abbreviation := strings.TrimSpace(t.Abbreviation)
	if abbreviation == "" {
		return timezoneerrors.ErrTimezoneAbbreviationRequired
	}
	// UTCOffset validation (-12 hours to +14 hours in seconds)
	if t.UTCOffset < -43200 || t.UTCOffset > 50400 {
		return timezoneerrors.ErrUTCOffsetOutOfRange
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
