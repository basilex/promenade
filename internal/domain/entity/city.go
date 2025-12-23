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
	if c.CountryID == uuidv7.Nil {
		return errors.New("country_id is required")
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
