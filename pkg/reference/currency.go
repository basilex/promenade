package reference

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Currency represents an ISO 4217 currency.
// This is a Value Object shared across all contexts.
type Currency struct {
	ID            uuidv7.UUID `db:"id"`
	Code          string      `db:"code"`
	NumericCode   string      `db:"numeric_code"`
	Name          string      `db:"name"`
	Symbol        string      `db:"symbol"`
	DecimalPlaces int         `db:"decimal_places"`
	IsActive      bool        `db:"is_active"`
}

// NewCurrency creates a new Currency value object with validation.
func NewCurrency(code, name, symbol string, decimalPlaces int) (Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	symbol = strings.TrimSpace(symbol)

	if len(code) != 3 {
		return Currency{}, fmt.Errorf("currency code must be 3 characters (ISO 4217)")
	}

	if name == "" {
		return Currency{}, fmt.Errorf("currency name is required")
	}

	if symbol == "" {
		return Currency{}, fmt.Errorf("currency symbol is required")
	}

	if decimalPlaces < 0 || decimalPlaces > 4 {
		return Currency{}, fmt.Errorf("decimal places must be between 0 and 4")
	}

	return Currency{
		ID:            uuidv7.New(),
		Code:          code,
		Name:          name,
		Symbol:        symbol,
		DecimalPlaces: decimalPlaces,
		IsActive:      true,
	}, nil
}

// String returns the currency code.
func (c Currency) String() string {
	return c.Code
}

// Equals checks if two currencies are the same (by ISO code).
func (c Currency) Equals(other Currency) bool {
	return c.Code == other.Code
}

// FormatAmount formats a monetary amount with proper decimal places.
func (c Currency) FormatAmount(amount int64) string {
	if c.DecimalPlaces == 0 {
		return fmt.Sprintf("%s%d", c.Symbol, amount)
	}

	divisor := int64(1)
	for i := 0; i < c.DecimalPlaces; i++ {
		divisor *= 10
	}

	whole := amount / divisor
	fraction := amount % divisor

	format := fmt.Sprintf("%%s%%d.%%0%dd", c.DecimalPlaces)
	return fmt.Sprintf(format, c.Symbol, whole, fraction)
}
