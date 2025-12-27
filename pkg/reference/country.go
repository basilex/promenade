// Package reference provides shared reference data (countries, currencies, languages, timezones)
// accessible by all Bounded Contexts. This is part of the Shared Kernel.
package reference

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Country represents an ISO 3166-1 country.
// This is a Value Object shared across all contexts.
type Country struct {
	ID          uuidv7.UUID `db:"id"`
	Code        string      `db:"code"`
	Code3       string      `db:"code3"`
	NumericCode string      `db:"numeric_code"`
	Name        string      `db:"name"`
	NameLocal   string      `db:"name_local"`
	PhoneCode   string      `db:"phone_code"`
	IsActive    bool        `db:"is_active"`
}

// NewCountry creates a new Country value object with validation.
func NewCountry(code, name, phoneCode string) (Country, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	phoneCode = strings.TrimSpace(phoneCode)

	if len(code) != 2 {
		return Country{}, fmt.Errorf("country code must be 2 characters (ISO 3166-1 alpha-2)")
	}

	if name == "" {
		return Country{}, fmt.Errorf("country name is required")
	}

	if phoneCode == "" {
		return Country{}, fmt.Errorf("phone code is required")
	}

	// Phone code must start with +
	if !strings.HasPrefix(phoneCode, "+") {
		phoneCode = "+" + phoneCode
	}

	return Country{
		ID:        uuidv7.New(),
		Code:      code,
		Name:      name,
		PhoneCode: phoneCode,
		IsActive:  true,
	}, nil
}

// String returns the country name.
func (c Country) String() string {
	return c.Name
}

// Equals checks if two countries are the same (by ISO code).
func (c Country) Equals(other Country) bool {
	return c.Code == other.Code
}
