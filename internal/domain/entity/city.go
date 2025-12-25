package entity

import (
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// City represents a city or major settlement
type City struct {
	ID                uuidv7.UUID  `json:"id" db:"id"`
	RegionID          *uuidv7.UUID `json:"region_id,omitempty" db:"region_id"`
	CountryID         uuidv7.UUID  `json:"country_id" db:"country_id"`
	Name              string       `json:"name" db:"name"`
	NameLocal         *string      `json:"name_local,omitempty" db:"name_local"`
	Latitude          *float64     `json:"latitude,omitempty" db:"latitude"`
	Longitude         *float64     `json:"longitude,omitempty" db:"longitude"`
	Population        *int         `json:"population,omitempty" db:"population"`
	IsCapital         bool         `json:"is_capital" db:"is_capital"`
	IsRegionalCapital bool         `json:"is_regional_capital" db:"is_regional_capital"`
	IsActive          bool         `json:"is_active" db:"is_active"`
	SortOrder         int          `json:"sort_order" db:"sort_order"`
	CreatedAt         time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at" db:"updated_at"`
}

// Validate validates the city entity
func (c *City) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if len(c.Name) < 2 || len(c.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}
	if c.CountryID == uuidv7.Nil {
		return errors.New("country_id is required")
	}
	// Validate coordinates if provided
	if c.Latitude != nil && (*c.Latitude < -90 || *c.Latitude > 90) {
		return errors.New("latitude must be between -90 and 90")
	}
	if c.Longitude != nil && (*c.Longitude < -180 || *c.Longitude > 180) {
		return errors.New("longitude must be between -180 and 180")
	}
	// Validate population if provided
	if c.Population != nil && *c.Population < 0 {
		return errors.New("population cannot be negative")
	}
	return nil
}

// NewCity creates a new city with default values
func NewCity(countryID uuidv7.UUID, name string) *City {
	now := time.Now()
	return &City{
		ID:                uuidv7.New(),
		CountryID:         countryID,
		Name:              name,
		IsCapital:         false,
		IsRegionalCapital: false,
		IsActive:          true,
		SortOrder:         0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}
