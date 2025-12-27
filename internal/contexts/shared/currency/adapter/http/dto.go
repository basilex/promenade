package http

import "github.com/basilex/promenade/internal/contexts/shared/currency"

// CurrencyResponse represents a currency in API responses
type CurrencyResponse struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	NumericCode   string `json:"numeric_code"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	DecimalPlaces int    `json:"decimal_places"`
	IsActive      bool   `json:"is_active"`
}

// ToCurrencyResponse converts currency entity to CurrencyResponse
func ToCurrencyResponse(c *currency.Currency) CurrencyResponse {
	return CurrencyResponse{
		ID:            c.ID.String(),
		Code:          c.Code,
		NumericCode:   c.NumericCode,
		Name:          c.Name,
		Symbol:        c.Symbol,
		DecimalPlaces: c.DecimalPlaces,
		IsActive:      c.IsActive,
	}
}

// ToCurrencyResponses converts slice of currencies to response DTOs
func ToCurrencyResponses(currencies []*currency.Currency) []CurrencyResponse {
	responses := make([]CurrencyResponse, len(currencies))
	for i, c := range currencies {
		responses[i] = ToCurrencyResponse(c)
	}
	return responses
}

// CreateCurrencyRequest represents request to create a currency
type CreateCurrencyRequest struct {
	Code          string `json:"code" binding:"required,len=3"`
	NumericCode   string `json:"numeric_code" binding:"omitempty"`
	Name          string `json:"name" binding:"required,min=2,max=100"`
	Symbol        string `json:"symbol" binding:"required"`
	DecimalPlaces int    `json:"decimal_places" binding:"min=0,max=4"`
}

// UpdateCurrencyRequest represents request to update a currency
type UpdateCurrencyRequest struct {
	NumericCode   string `json:"numeric_code" binding:"omitempty"`
	Name          string `json:"name" binding:"required,min=2,max=100"`
	Symbol        string `json:"symbol" binding:"required"`
	DecimalPlaces int    `json:"decimal_places" binding:"min=0,max=4"`
	IsActive      bool   `json:"is_active"`
}
