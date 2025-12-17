package dto

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// CountryResponse represents country response
type CountryResponse struct {
	ID         uuidv7.UUID        `json:"id"`
	Name       string             `json:"name"`
	Code       string             `json:"code"`
	ISO2       string             `json:"iso2"`
	ISO3       string             `json:"iso3"`
	Region     string             `json:"region"`
	Currencies []CurrencyResponse `json:"currencies,omitempty" swaggerignore:"true"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// CreateCountryRequest represents create country request
type CreateCountryRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100" example:"United States"`
	Code   string `json:"code" binding:"required,min=2,max=10,uppercase" example:"+1"`
	ISO2   string `json:"iso2" binding:"required,len=2,alpha,uppercase" example:"US"`
	ISO3   string `json:"iso3" binding:"required,len=3,alpha,uppercase" example:"USA"`
	Region string `json:"region" binding:"required,oneof=north_america south_america western_europe eastern_europe asia middle_east africa oceania" example:"north_america"`
}

// UpdateCountryRequest represents update country request
type UpdateCountryRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100" example:"United States"`
	Code   string `json:"code" binding:"required,min=2,max=10,uppercase" example:"+1"`
	ISO2   string `json:"iso2" binding:"required,len=2,alpha,uppercase" example:"US"`
	ISO3   string `json:"iso3" binding:"required,len=3,alpha,uppercase" example:"USA"`
	Region string `json:"region" binding:"required,oneof=north_america south_america western_europe eastern_europe asia middle_east africa oceania" example:"north_america"`
}

// CurrencyResponse represents currency response
type CurrencyResponse struct {
	ID        uuidv7.UUID       `json:"id"`
	Name      string            `json:"name"`
	Code      string            `json:"code"`
	Symbol    string            `json:"symbol"`
	Countries []CountryResponse `json:"countries,omitempty" swaggerignore:"true"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// CreateCurrencyRequest represents create currency request
type CreateCurrencyRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100" example:"US Dollar"`
	Code   string `json:"code" binding:"required,min=3,max=10,alphanum,uppercase" example:"USD"`
	Symbol string `json:"symbol" binding:"omitempty,max=10" example:"$"`
}

// UpdateCurrencyRequest represents update currency request
type UpdateCurrencyRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100" example:"US Dollar"`
	Code   string `json:"code" binding:"required,min=3,max=10,alphanum,uppercase" example:"USD"`
	Symbol string `json:"symbol" binding:"omitempty,max=10" example:"$"`
}

// AddCurrencyToCountryRequest represents adding currency to country
type AddCurrencyToCountryRequest struct {
	CurrencyID string `json:"currency_id" binding:"required,uuid"`
	IsPrimary  bool   `json:"is_primary"`
}

// AddCountryToCurrencyRequest represents adding country to currency
type AddCountryToCurrencyRequest struct {
	CountryID string `json:"country_id" binding:"required,uuid"`
	IsPrimary bool   `json:"is_primary"`
}
