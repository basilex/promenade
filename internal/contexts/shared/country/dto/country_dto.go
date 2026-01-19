package dto

import (
	"github.com/basilex/promenade/internal/contexts/shared/country/aggregate"
)

// CountryResponse represents a country in API responses
type CountryResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Code3       string `json:"code3"`
	NumericCode string `json:"numeric_code"`
	Name        string `json:"name"`
	NameLocal   string `json:"name_local"`
	PhoneCode   string `json:"phone_code"`
	IsActive    bool   `json:"is_active"`
}

// ToCountryResponse converts country entity to CountryResponse
func ToCountryResponse(c *aggregate.Country) CountryResponse {
	return CountryResponse{
		ID:          c.ID.String(),
		Code:        c.Code,
		Code3:       c.Code3,
		NumericCode: c.NumericCode,
		Name:        c.Name,
		NameLocal:   c.NameLocal,
		PhoneCode:   c.PhoneCode,
		IsActive:    c.IsActive,
	}
}

// ToCountryResponses converts slice of countries to response DTOs
func ToCountryResponses(countries []*aggregate.Country) []CountryResponse {
	responses := make([]CountryResponse, len(countries))
	for i, c := range countries {
		responses[i] = ToCountryResponse(c)
	}
	return responses
}

// CreateCountryRequest represents request to create a country
type CreateCountryRequest struct {
	Code        string `json:"code" binding:"required,len=2"`
	Code3       string `json:"code3" binding:"omitempty,len=3"`
	NumericCode string `json:"numeric_code" binding:"omitempty"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	NameLocal   string `json:"name_local" binding:"omitempty,max=100"`
	PhoneCode   string `json:"phone_code" binding:"required"`
}

// UpdateCountryRequest represents request to update a country
type UpdateCountryRequest struct {
	Code3       string `json:"code3" binding:"omitempty,len=3"`
	NumericCode string `json:"numeric_code" binding:"omitempty"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	NameLocal   string `json:"name_local" binding:"omitempty,max=100"`
	PhoneCode   string `json:"phone_code" binding:"required"`
	IsActive    bool   `json:"is_active"`
}
