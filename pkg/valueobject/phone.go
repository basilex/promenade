package valueobject

import (
	"fmt"
	"regexp"
	"strings"
)

// Phone represents a validated phone number.
// It is immutable and stores the number in E.164 format.
type Phone struct {
	value string // E.164 format: +380501234567
}

var (
	// phoneRegex matches E.164 format: +[country code][number]
	phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
)

// NewPhone creates a new Phone value object after validation.
// Accepts formats like: +380501234567, +1234567890
func NewPhone(number string) (Phone, error) {
	number = strings.TrimSpace(number)

	if number == "" {
		return Phone{}, fmt.Errorf("phone number is required")
	}

	// Remove common separators
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")
	number = strings.ReplaceAll(number, "(", "")
	number = strings.ReplaceAll(number, ")", "")

	// Ensure + prefix
	if !strings.HasPrefix(number, "+") {
		return Phone{}, fmt.Errorf("phone number must start with + (E.164 format)")
	}

	if !phoneRegex.MatchString(number) {
		return Phone{}, fmt.Errorf("invalid phone number format: %s (use E.164: +[country][number])", number)
	}

	return Phone{value: number}, nil
}

// String returns the phone number as a string.
func (p Phone) String() string {
	return p.value
}

// Value returns the phone number (alias for String).
func (p Phone) Value() string {
	return p.value
}

// CountryCode extracts the country code from the phone number.
// Returns empty string if unable to extract.
func (p Phone) CountryCode() string {
	if len(p.value) < 2 {
		return ""
	}

	// Common country codes (1-3 digits)
	for i := 2; i <= 4 && i < len(p.value); i++ {
		code := p.value[1:i]
		if isValidCountryCode(code) {
			return code
		}
	}

	return ""
}

// Formatted returns the phone number with spaces for readability.
// Example: +380 50 123 4567
func (p Phone) Formatted() string {
	if len(p.value) < 4 {
		return p.value
	}

	// Simple formatting: +CC XX XXX XXXX
	result := p.value[:4]
	remaining := p.value[4:]

	for i := 0; i < len(remaining); i += 3 {
		end := i + 3
		if end > len(remaining) {
			end = len(remaining)
		}
		result += " " + remaining[i:end]
	}

	return result
}

// Equals checks if two phone numbers are equal.
func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}

// IsEmpty checks if the phone is empty (zero value).
func (p Phone) IsEmpty() bool {
	return p.value == ""
}

// isValidCountryCode checks if a string is a valid country code.
// This is a simplified check - in production, use a comprehensive list.
func isValidCountryCode(code string) bool {
	// Common country codes
	valid := map[string]bool{
		"1":   true, // USA, Canada
		"7":   true, // Russia, Kazakhstan
		"20":  true, // Egypt
		"27":  true, // South Africa
		"30":  true, // Greece
		"31":  true, // Netherlands
		"32":  true, // Belgium
		"33":  true, // France
		"34":  true, // Spain
		"36":  true, // Hungary
		"39":  true, // Italy
		"40":  true, // Romania
		"41":  true, // Switzerland
		"43":  true, // Austria
		"44":  true, // UK
		"45":  true, // Denmark
		"46":  true, // Sweden
		"47":  true, // Norway
		"48":  true, // Poland
		"49":  true, // Germany
		"351": true, // Portugal
		"352": true, // Luxembourg
		"353": true, // Ireland
		"354": true, // Iceland
		"355": true, // Albania
		"356": true, // Malta
		"357": true, // Cyprus
		"358": true, // Finland
		"359": true, // Bulgaria
		"370": true, // Lithuania
		"371": true, // Latvia
		"372": true, // Estonia
		"373": true, // Moldova
		"374": true, // Armenia
		"375": true, // Belarus
		"376": true, // Andorra
		"377": true, // Monaco
		"378": true, // San Marino
		"380": true, // Ukraine
		"381": true, // Serbia
		"382": true, // Montenegro
		"385": true, // Croatia
		"386": true, // Slovenia
		"387": true, // Bosnia
		"389": true, // Macedonia
		"420": true, // Czech Republic
		"421": true, // Slovakia
	}

	return valid[code]
}
