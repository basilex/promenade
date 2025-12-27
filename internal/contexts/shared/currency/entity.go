package currency

import (
	"errors"
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Common errors
var (
	ErrNotFound = errors.New("currency not found")
)

// Currency represents an ISO 4217 currency (Aggregate Root in Shared Context)
type Currency struct {
	ID            uuidv7.UUID `db:"id"`
	Code          string      `db:"code"`
	NumericCode   string      `db:"numeric_code"`
	Name          string      `db:"name"`
	Symbol        string      `db:"symbol"`
	DecimalPlaces int         `db:"decimal_places"`
	IsActive      bool        `db:"is_active"`
}

// NewCurrency creates a new Currency entity with validation
func NewCurrency(code, name, symbol string, decimalPlaces int) (*Currency, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	symbol = strings.TrimSpace(symbol)

	if len(code) != 3 {
		return nil, fmt.Errorf("currency code must be 3 characters (ISO 4217)")
	}

	if name == "" {
		return nil, fmt.Errorf("currency name is required")
	}

	if symbol == "" {
		return nil, fmt.Errorf("currency symbol is required")
	}

	if decimalPlaces < 0 || decimalPlaces > 4 {
		return nil, fmt.Errorf("decimal places must be between 0 and 4")
	}

	return &Currency{
		ID:            uuidv7.New(),
		Code:          code,
		Name:          name,
		Symbol:        symbol,
		DecimalPlaces: decimalPlaces,
		IsActive:      true,
	}, nil
}

// Validate validates currency data
func (c *Currency) Validate() error {
	if len(c.Code) != 3 {
		return fmt.Errorf("currency code must be 3 characters")
	}
	if c.Name == "" {
		return fmt.Errorf("currency name is required")
	}
	if c.Symbol == "" {
		return fmt.Errorf("currency symbol is required")
	}
	if c.DecimalPlaces < 0 || c.DecimalPlaces > 4 {
		return fmt.Errorf("decimal places must be between 0 and 4")
	}
	return nil
}

// String returns the currency code
func (c *Currency) String() string {
	return c.Code
}

// ParseUUID parses a string UUID and returns error on failure
func ParseUUID(s string) (uuidv7.UUID, error) {
return uuidv7.Parse(s)
}
