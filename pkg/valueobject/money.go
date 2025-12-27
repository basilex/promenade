package valueobject

import (
	"fmt"
	"math"
)

// Money represents a monetary value with currency.
// Amount is stored in the smallest unit (cents, kopiykas, etc.) to avoid floating-point errors.
type Money struct {
	Amount   int64  // stored in smallest unit (cents, kopiykas, etc.)
	Currency string // ISO 4217 code (USD, UAH, EUR)
}

// NewMoney creates a Money value object.
func NewMoney(amount int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, fmt.Errorf("currency is required")
	}
	if len(currency) != 3 {
		return Money{}, fmt.Errorf("currency must be 3-letter ISO 4217 code (got %d characters)", len(currency))
	}
	
	return Money{Amount: amount, Currency: currency}, nil
}

// FromFloat converts a float amount to Money (handles rounding).
// Example: FromFloat(10.50, "USD") creates $10.50
func FromFloat(amount float64, currency string) (Money, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return Money{}, fmt.Errorf("invalid amount: %f", amount)
	}
	
	cents := int64(math.Round(amount * 100))
	return NewMoney(cents, currency)
}

// Zero creates a zero-value Money with the specified currency.
func Zero(currency string) (Money, error) {
	return NewMoney(0, currency)
}

// ToFloat converts Money to float (for display purposes).
// Warning: Should only be used for display, not calculations.
func (m Money) ToFloat() float64 {
	return float64(m.Amount) / 100.0
}

// Add adds two Money values (same currency required).
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract subtracts one Money from another (same currency required).
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// Multiply multiplies Money by a scalar.
func (m Money) Multiply(factor float64) Money {
	return Money{
		Amount:   int64(math.Round(float64(m.Amount) * factor)),
		Currency: m.Currency,
	}
}

// Divide divides Money by a scalar.
func (m Money) Divide(divisor float64) (Money, error) {
	if divisor == 0 {
		return Money{}, fmt.Errorf("cannot divide by zero")
	}
	return Money{
		Amount:   int64(math.Round(float64(m.Amount) / divisor)),
		Currency: m.Currency,
	}, nil
}

// IsZero checks if amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive checks if amount is positive.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// IsNegative checks if amount is negative.
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// Abs returns absolute value of Money.
func (m Money) Abs() Money {
	if m.Amount < 0 {
		return Money{Amount: -m.Amount, Currency: m.Currency}
	}
	return m
}

// Negate returns negated Money.
func (m Money) Negate() Money {
	return Money{Amount: -m.Amount, Currency: m.Currency}
}

// Compare compares two Money values (same currency required).
// Returns: -1 if m < other, 0 if m == other, 1 if m > other
func (m Money) Compare(other Money) (int, error) {
	if m.Currency != other.Currency {
		return 0, fmt.Errorf("cannot compare different currencies: %s and %s", m.Currency, other.Currency)
	}
	
	if m.Amount < other.Amount {
		return -1, nil
	}
	if m.Amount > other.Amount {
		return 1, nil
	}
	return 0, nil
}

// LessThan checks if m < other.
func (m Money) LessThan(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	return cmp < 0, err
}

// LessThanOrEqual checks if m <= other.
func (m Money) LessThanOrEqual(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	return cmp <= 0, err
}

// GreaterThan checks if m > other.
func (m Money) GreaterThan(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	return cmp > 0, err
}

// GreaterThanOrEqual checks if m >= other.
func (m Money) GreaterThanOrEqual(other Money) (bool, error) {
	cmp, err := m.Compare(other)
	return cmp >= 0, err
}

// Equals checks if two Money values are equal.
func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// String returns formatted string representation.
// Example: "USD 10.50" or "UAH 250.00"
func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.Currency, m.ToFloat())
}

// Format returns formatted string with currency symbol.
// Example: "$10.50" or "€9.99"
func (m Money) Format() string {
	symbol := getCurrencySymbol(m.Currency)
	amount := m.ToFloat()
	
	if m.Amount < 0 {
		return fmt.Sprintf("-%s%.2f", symbol, -amount)
	}
	return fmt.Sprintf("%s%.2f", symbol, amount)
}

// getCurrencySymbol returns the symbol for common currencies.
func getCurrencySymbol(currency string) string {
	symbols := map[string]string{
		"USD": "$",
		"EUR": "€",
		"GBP": "£",
		"UAH": "₴",
		"RUB": "₽",
		"JPY": "¥",
		"CNY": "¥",
		"CHF": "Fr",
		"CAD": "CA$",
		"AUD": "A$",
	}
	
	if symbol, ok := symbols[currency]; ok {
		return symbol
	}
	return currency + " "
}
