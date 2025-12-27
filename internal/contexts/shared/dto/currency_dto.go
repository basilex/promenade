package dto

import "github.com/basilex/promenade/pkg/reference"

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

// ToCurrencyResponse converts reference.Currency to CurrencyResponse
func ToCurrencyResponse(currency *reference.Currency) CurrencyResponse {
	return CurrencyResponse{
		ID:            currency.ID.String(),
		Code:          currency.Code,
		NumericCode:   currency.NumericCode,
		Name:          currency.Name,
		Symbol:        currency.Symbol,
		DecimalPlaces: currency.DecimalPlaces,
		IsActive:      currency.IsActive,
	}
}

// ToCurrencyResponses converts slice of currencies to response DTOs
func ToCurrencyResponses(currencies []reference.Currency) []CurrencyResponse {
	responses := make([]CurrencyResponse, len(currencies))
	for i, currency := range currencies {
		responses[i] = ToCurrencyResponse(&currency)
	}
	return responses
}
