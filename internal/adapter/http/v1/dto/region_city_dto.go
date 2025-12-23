package dto

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// RegionResponse represents region response
type RegionResponse struct {
	ID         uuidv7.UUID `json:"id"`
	CountryID  uuidv7.UUID `json:"country_id"`
	Name       string      `json:"name"`
	Code       string      `json:"code"`
	RegionType string      `json:"region_type"`
	Latitude   *float64    `json:"latitude,omitempty"`
	Longitude  *float64    `json:"longitude,omitempty"`
	IsActive   bool        `json:"is_active"`
	SortOrder  int         `json:"sort_order"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// CreateRegionRequest represents create region request
type CreateRegionRequest struct {
	CountryID  string   `json:"country_id" binding:"required,uuid" example:"019b4b75-3435-77d6-b88d-bb36dc75fda3"`
	Name       string   `json:"name" binding:"required,min=2,max=100" example:"California"`
	Code       string   `json:"code" binding:"required,min=1,max=20" example:"CA"`
	RegionType string   `json:"region_type" binding:"required,min=2,max=30" example:"state"`
	Latitude   *float64 `json:"latitude" binding:"omitempty,latitude" example:"36.7783"`
	Longitude  *float64 `json:"longitude" binding:"omitempty,longitude" example:"-119.4179"`
	IsActive   *bool    `json:"is_active" binding:"omitempty" example:"true"`
	SortOrder  *int     `json:"sort_order" binding:"omitempty,gte=0" example:"1"`
}

// UpdateRegionRequest represents update region request
type UpdateRegionRequest struct {
	Name       string   `json:"name" binding:"required,min=2,max=100" example:"California"`
	Code       string   `json:"code" binding:"required,min=1,max=20" example:"CA"`
	RegionType string   `json:"region_type" binding:"required,min=2,max=30" example:"state"`
	Latitude   *float64 `json:"latitude" binding:"omitempty,latitude" example:"36.7783"`
	Longitude  *float64 `json:"longitude" binding:"omitempty,longitude" example:"-119.4179"`
	IsActive   *bool    `json:"is_active" binding:"omitempty" example:"true"`
	SortOrder  *int     `json:"sort_order" binding:"omitempty,gte=0" example:"1"`
}

// CityResponse represents city response
type CityResponse struct {
	ID                uuidv7.UUID  `json:"id"`
	RegionID          *uuidv7.UUID `json:"region_id,omitempty"`
	CountryID         uuidv7.UUID  `json:"country_id"`
	Name              string       `json:"name"`
	NameLocal         *string      `json:"name_local,omitempty"`
	Latitude          *float64     `json:"latitude,omitempty"`
	Longitude         *float64     `json:"longitude,omitempty"`
	Population        *int         `json:"population,omitempty"`
	IsCapital         bool         `json:"is_capital"`
	IsRegionalCapital bool         `json:"is_regional_capital"`
	IsActive          bool         `json:"is_active"`
	SortOrder         int          `json:"sort_order"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

// CreateCityRequest represents create city request
type CreateCityRequest struct {
	CountryID         string   `json:"country_id" binding:"required,uuid" example:"019b4b75-3435-77d6-b88d-bb36dc75fda3"`
	RegionID          *string  `json:"region_id" binding:"omitempty,uuid" example:"019b4b75-3435-77d6-b88d-bb36dc75fda3"`
	Name              string   `json:"name" binding:"required,min=2,max=100" example:"Los Angeles"`
	NameLocal         *string  `json:"name_local" binding:"omitempty,min=2,max=100" example:"Los Angeles"`
	Latitude          *float64 `json:"latitude" binding:"omitempty,latitude" example:"34.0522"`
	Longitude         *float64 `json:"longitude" binding:"omitempty,longitude" example:"-118.2437"`
	Population        *int     `json:"population" binding:"omitempty,gte=0" example:"3900000"`
	IsCapital         *bool    `json:"is_capital" binding:"omitempty" example:"false"`
	IsRegionalCapital *bool    `json:"is_regional_capital" binding:"omitempty" example:"false"`
	IsActive          *bool    `json:"is_active" binding:"omitempty" example:"true"`
	SortOrder         *int     `json:"sort_order" binding:"omitempty,gte=0" example:"1"`
}

// UpdateCityRequest represents update city request
type UpdateCityRequest struct {
	RegionID          *string  `json:"region_id" binding:"omitempty,uuid" example:"019b4b75-3435-77d6-b88d-bb36dc75fda3"`
	Name              string   `json:"name" binding:"required,min=2,max=100" example:"Los Angeles"`
	NameLocal         *string  `json:"name_local" binding:"omitempty,min=2,max=100" example:"Los Angeles"`
	Latitude          *float64 `json:"latitude" binding:"omitempty,latitude" example:"34.0522"`
	Longitude         *float64 `json:"longitude" binding:"omitempty,longitude" example:"-118.2437"`
	Population        *int     `json:"population" binding:"omitempty,gte=0" example:"3900000"`
	IsCapital         *bool    `json:"is_capital" binding:"omitempty" example:"false"`
	IsRegionalCapital *bool    `json:"is_regional_capital" binding:"omitempty" example:"false"`
	IsActive          *bool    `json:"is_active" binding:"omitempty" example:"true"`
	SortOrder         *int     `json:"sort_order" binding:"omitempty,gte=0" example:"1"`
}
