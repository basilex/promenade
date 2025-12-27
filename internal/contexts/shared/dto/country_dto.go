package dto

import "github.com/basilex/promenade/pkg/reference"

// CountryResponse represents a country in API responses
type CountryResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Code3       string `json:"code3"`
	NumericCode string `json:"numeric_code"`
	Name        string `json:"name"`
	NameLocal   string `json:"name_local,omitempty"`
	PhoneCode   string `json:"phone_code"`
	IsActive    bool   `json:"is_active"`
}

// ToCountryResponse converts reference.Country to CountryResponse
func ToCountryResponse(country *reference.Country) CountryResponse {
	return CountryResponse{
		ID:          country.ID.String(),
		Code:        country.Code,
		Code3:       country.Code3,
		NumericCode: country.NumericCode,
		Name:        country.Name,
		NameLocal:   country.NameLocal,
		PhoneCode:   country.PhoneCode,
		IsActive:    country.IsActive,
	}
}

// ToCountryResponses converts slice of countries to response DTOs
func ToCountryResponses(countries []reference.Country) []CountryResponse {
	responses := make([]CountryResponse, len(countries))
	for i, country := range countries {
		responses[i] = ToCountryResponse(&country)
	}
	return responses
}
