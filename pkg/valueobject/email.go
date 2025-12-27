// Package valueobject provides immutable value objects for domain modeling.
// Value objects are defined by their attributes rather than identity.
package valueobject

import (
	"fmt"
	"regexp"
	"strings"
)

// Email represents a validated email address.
// It is immutable and guarantees validity.
type Email struct {
	value string
}

var (
	// emailRegex is a simple email validation pattern
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// NewEmail creates a new Email value object after validation.
func NewEmail(address string) (Email, error) {
	address = strings.TrimSpace(strings.ToLower(address))

	if address == "" {
		return Email{}, fmt.Errorf("email address is required")
	}

	if len(address) > 254 {
		return Email{}, fmt.Errorf("email address too long (max 254 characters)")
	}

	if !emailRegex.MatchString(address) {
		return Email{}, fmt.Errorf("invalid email format: %s", address)
	}

	return Email{value: address}, nil
}

// String returns the email address as a string.
func (e Email) String() string {
	return e.value
}

// Value returns the email address (alias for String).
func (e Email) Value() string {
	return e.value
}

// Domain returns the domain part of the email (after @).
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// LocalPart returns the local part of the email (before @).
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

// Equals checks if two emails are equal.
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// IsEmpty checks if the email is empty (zero value).
func (e Email) IsEmpty() bool {
	return e.value == ""
}
