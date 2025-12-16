package entity

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Country represents a country entity
type Country struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" validate:"required,min=2,max=100"`
	Code      string    `db:"code" json:"code" validate:"required,min=2,max=10,alphanum,uppercase"`
	ISO2      string    `db:"iso2" json:"iso2" validate:"required,len=2,alpha,uppercase"`
	ISO3      string    `db:"iso3" json:"iso3" validate:"required,len=3,alpha,uppercase"`
	Region    string    `db:"region" json:"region" validate:"required,oneof=north_america south_america western_europe eastern_europe asia middle_east africa oceania"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relationships
	Currencies []Currency `db:"-" json:"currencies,omitempty"`
}

// Validate validates country fields
func (c *Country) Validate() error {
	var errs []string

	// Name validation
	if strings.TrimSpace(c.Name) == "" {
		errs = append(errs, "name is required")
	} else if len(c.Name) < 2 || len(c.Name) > 100 {
		errs = append(errs, "name must be between 2 and 100 characters")
	}

	// Code validation
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	if c.Code == "" {
		errs = append(errs, "code is required")
	} else if len(c.Code) < 2 || len(c.Code) > 10 {
		errs = append(errs, "code must be between 2 and 10 characters")
	}

	// ISO2 validation
	c.ISO2 = strings.ToUpper(strings.TrimSpace(c.ISO2))
	if len(c.ISO2) != 2 {
		errs = append(errs, "iso2 must be exactly 2 characters")
	} else if !isAlpha(c.ISO2) {
		errs = append(errs, "iso2 must contain only letters")
	}

	// ISO3 validation
	c.ISO3 = strings.ToUpper(strings.TrimSpace(c.ISO3))
	if len(c.ISO3) != 3 {
		errs = append(errs, "iso3 must be exactly 3 characters")
	} else if !isAlpha(c.ISO3) {
		errs = append(errs, "iso3 must contain only letters")
	}

	// Region validation
	c.Region = strings.ToLower(strings.TrimSpace(c.Region))
	validRegions := []string{"north_america", "south_america", "western_europe", "eastern_europe", "asia", "middle_east", "africa", "oceania"}
	validRegion := false
	for _, vr := range validRegions {
		if c.Region == vr {
			validRegion = true
			break
		}
	}
	if !validRegion {
		errs = append(errs, "region must be one of: north_america, south_america, western_europe, eastern_europe, asia, middle_east, africa, oceania")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// Currency represents a currency entity
type Currency struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name" validate:"required,min=2,max=100"`
	Code      string    `db:"code" json:"code" validate:"required,min=3,max=10,alphanum,uppercase"`
	Symbol    string    `db:"symbol" json:"symbol" validate:"max=10"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Relationships
	Countries []Country `db:"-" json:"countries,omitempty"`
}

// Validate validates currency fields
func (c *Currency) Validate() error {
	var errs []string

	// Name validation
	if strings.TrimSpace(c.Name) == "" {
		errs = append(errs, "name is required")
	} else if len(c.Name) < 2 || len(c.Name) > 100 {
		errs = append(errs, "name must be between 2 and 100 characters")
	}

	// Code validation (ISO 4217 standard - 3 characters)
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))
	if c.Code == "" {
		errs = append(errs, "code is required")
	} else if len(c.Code) < 3 || len(c.Code) > 10 {
		errs = append(errs, "code must be between 3 and 10 characters")
	} else if !isAlphaNum(c.Code) {
		errs = append(errs, "code must contain only letters and numbers")
	}

	// Symbol validation (optional but limited)
	if len(c.Symbol) > 10 {
		errs = append(errs, "symbol must not exceed 10 characters")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

// CountryCurrency represents the many-to-many junction
type CountryCurrency struct {
	CountryID  uuid.UUID `db:"country_id" validate:"required"`
	CurrencyID uuid.UUID `db:"currency_id" validate:"required"`
	IsPrimary  bool      `db:"is_primary"`
	CreatedAt  time.Time `db:"created_at"`
}

// Validate validates country-currency relationship
func (cc *CountryCurrency) Validate() error {
	if cc.CountryID == uuid.Nil {
		return fmt.Errorf("country_id is required")
	}
	if cc.CurrencyID == uuid.Nil {
		return fmt.Errorf("currency_id is required")
	}
	return nil
}

// Helper functions
func isAlpha(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return false
		}
	}
	return true
}

func isAlphaNum(s string) bool {
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}
