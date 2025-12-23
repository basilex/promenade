package entity

import (
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Region represents an administrative division (state, province, oblast, etc.)
type Region struct {
	ID         uuidv7.UUID `json:"id" db:"id"`
	CountryID  uuidv7.UUID `json:"country_id" db:"country_id"`
	Name       string      `json:"name" db:"name"`
	Code       string      `json:"code" db:"code"`
	RegionType string      `json:"region_type" db:"region_type"`
	Latitude   *float64    `json:"latitude,omitempty" db:"latitude"`
	Longitude  *float64    `json:"longitude,omitempty" db:"longitude"`
	IsActive   bool        `json:"is_active" db:"is_active"`
	SortOrder  int         `json:"sort_order" db:"sort_order"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at" db:"updated_at"`
}

// Validate validates the region entity
func (r *Region) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Code == "" {
		return errors.New("code is required")
	}
	if r.RegionType == "" {
		return errors.New("region_type is required")
	}
	if r.CountryID == uuidv7.Nil {
		return errors.New("country_id is required")
	}
	return nil
}

// NewRegion creates a new region with default values
func NewRegion(countryID uuidv7.UUID, name, code, regionType string) *Region {
	now := time.Now()
	return &Region{
		ID:         uuidv7.New(),
		CountryID:  countryID,
		Name:       name,
		Code:       code,
		RegionType: regionType,
		IsActive:   true,
		SortOrder:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
