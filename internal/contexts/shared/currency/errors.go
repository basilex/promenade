package currency

import "errors"

// Domain errors for currency aggregate
var (
	// ErrCurrencyNotFound is returned when a currency cannot be found
	ErrCurrencyNotFound = errors.New("currency not found")

	// ErrInvalidCurrencyCode is returned when currency code validation fails
	ErrInvalidCurrencyCode = errors.New("invalid currency code")

	// ErrInvalidDecimalPlaces is returned when decimal places validation fails
	ErrInvalidDecimalPlaces = errors.New("invalid decimal places")
)
