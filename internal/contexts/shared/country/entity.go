package country

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Country represents an ISO 3166-1 country (Aggregate Root in Shared Context)
type Country struct {
	aggregate.BaseAggregate

	Code         string         `db:"code"`
	Code3        string         `db:"code3"`
	NumericCode  string         `db:"numeric_code"`
	Name         string         `db:"name"`
	NameLocal    string         `db:"name_local"`
	PhoneCode    string         `db:"phone_code"`
	Capital      string         `db:"capital"`
	Region       string         `db:"region"`
	Subregion    string         `db:"subregion"`
	FlagEmoji    string         `db:"flag_emoji"`
	Latitude     *float64       `db:"latitude"`
	Longitude    *float64       `db:"longitude"`
	AreaKm2      *int           `db:"area_km2"`
	Population   *int64         `db:"population"`
	Translations map[string]any `db:"translations"` // JSONB map
	IsActive     bool           `db:"is_active"`
}

// NewCountry creates a new Country entity with validation
func NewCountry(code, name, phoneCode string) (*Country, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	phoneCode = strings.TrimSpace(phoneCode)

	if len(code) != 2 {
		return nil, fmt.Errorf("country code must be 2 characters (ISO 3166-1 alpha-2)")
	}

	if name == "" {
		return nil, fmt.Errorf("country name is required")
	}

	if phoneCode == "" {
		return nil, fmt.Errorf("phone code is required")
	}

	if !strings.HasPrefix(phoneCode, "+") {
		phoneCode = "+" + phoneCode
	}

	return &Country{
		BaseAggregate: aggregate.NewBaseAggregate(),
		Code:          code,
		Name:          name,
		PhoneCode:     phoneCode,
		IsActive:      true,
	}, nil
}

// Validate validates country data
func (c *Country) Validate() error {
	if len(c.Code) != 2 {
		return fmt.Errorf("country code must be 2 characters")
	}
	if c.Name == "" {
		return fmt.Errorf("country name is required")
	}
	if c.PhoneCode == "" {
		return fmt.Errorf("phone code is required")
	}
	return nil
}

// String returns the country code
func (c *Country) String() string {
	return c.Code
}

// ParseUUID parses a string UUID and returns error on failure
func ParseUUID(s string) (uuidv7.UUID, error) {
	return uuidv7.Parse(s)
}
