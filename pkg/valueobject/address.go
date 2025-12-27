package valueobject

import (
	"fmt"
	"strings"
)

// Address represents a postal address.
// It is immutable and validates all fields.
type Address struct {
	Street     string // Street address (line 1)
	Street2    string // Apartment, suite, etc. (optional, line 2)
	City       string // City name
	State      string // State/province/region (optional)
	PostalCode string // ZIP/postal code
	Country    string // ISO 3166-1 alpha-2 country code (e.g., "US", "UA", "DE")
}

// NewAddress creates a new Address value object.
func NewAddress(street, city, postalCode, country string) (Address, error) {
	street = strings.TrimSpace(street)
	city = strings.TrimSpace(city)
	postalCode = strings.TrimSpace(postalCode)
	country = strings.TrimSpace(strings.ToUpper(country))

	if street == "" {
		return Address{}, fmt.Errorf("street is required")
	}
	if city == "" {
		return Address{}, fmt.Errorf("city is required")
	}
	if postalCode == "" {
		return Address{}, fmt.Errorf("postal code is required")
	}
	if country == "" {
		return Address{}, fmt.Errorf("country is required")
	}

	// Validate country code (2 letters)
	if len(country) != 2 {
		return Address{}, fmt.Errorf("country must be 2-letter ISO code (e.g., US, UA, DE)")
	}

	return Address{
		Street:     street,
		City:       city,
		PostalCode: postalCode,
		Country:    country,
	}, nil
}

// NewAddressWithState creates an address with state/province.
func NewAddressWithState(street, city, state, postalCode, country string) (Address, error) {
	addr, err := NewAddress(street, city, postalCode, country)
	if err != nil {
		return Address{}, err
	}
	
	addr.State = strings.TrimSpace(state)
	return addr, nil
}

// WithStreet2 adds a second street line (apartment, suite, etc.).
func (a Address) WithStreet2(street2 string) Address {
	a.Street2 = strings.TrimSpace(street2)
	return a
}

// String returns a formatted address string.
func (a Address) String() string {
	var parts []string
	
	parts = append(parts, a.Street)
	if a.Street2 != "" {
		parts = append(parts, a.Street2)
	}
	
	cityLine := a.City
	if a.State != "" {
		cityLine += ", " + a.State
	}
	cityLine += " " + a.PostalCode
	parts = append(parts, cityLine)
	
	parts = append(parts, a.Country)
	
	return strings.Join(parts, "\n")
}

// SingleLine returns the address as a single line.
func (a Address) SingleLine() string {
	var parts []string
	
	parts = append(parts, a.Street)
	if a.Street2 != "" {
		parts = append(parts, a.Street2)
	}
	parts = append(parts, a.City)
	if a.State != "" {
		parts = append(parts, a.State)
	}
	parts = append(parts, a.PostalCode)
	parts = append(parts, a.Country)
	
	return strings.Join(parts, ", ")
}

// Equals checks if two addresses are equal.
func (a Address) Equals(other Address) bool {
	return a.Street == other.Street &&
		a.Street2 == other.Street2 &&
		a.City == other.City &&
		a.State == other.State &&
		a.PostalCode == other.PostalCode &&
		a.Country == other.Country
}

// IsEmpty checks if the address is empty (zero value).
func (a Address) IsEmpty() bool {
	return a.Street == "" && a.City == "" && a.PostalCode == "" && a.Country == ""
}

// IsDomestic checks if the address is in the specified country.
func (a Address) IsDomestic(country string) bool {
	return strings.EqualFold(a.Country, country)
}

// IsInternational checks if the address is outside the specified country.
func (a Address) IsInternational(country string) bool {
	return !a.IsDomestic(country)
}
