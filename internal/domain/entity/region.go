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
	if len(r.Name) < 2 || len(r.Name) > 100 {
		return errors.New("name must be between 2 and 100 characters")
	}
	if r.Code == "" {
		return errors.New("code is required")
	}
	if len(r.Code) < 2 || len(r.Code) > 10 {
		return errors.New("code must be between 2 and 10 characters")
	}
	if r.RegionType == "" {
		return errors.New("region_type is required")
	}
	// Valid region types: state, province, oblast, region, etc.
	validTypes := []string{"state", "province", "oblast", "region", "territory", "district", "county", "land"}
	validType := false
	for _, vt := range validTypes {
		if r.RegionType == vt {
			validType = true
			break
		}
	}
	if !validType {
		return errors.New("invalid region_type")
	}
	if r.CountryID == uuidv7.Nil {
		return errors.New("country_id is required")
	}
	// Validate coordinates if provided
	if r.Latitude != nil && (*r.Latitude < -90 || *r.Latitude > 90) {
		return errors.New("latitude must be between -90 and 90")
	}
	if r.Longitude != nil && (*r.Longitude < -180 || *r.Longitude > 180) {
		return errors.New("longitude must be between -180 and 180")
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
