package reference

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Language represents an ISO 639-1 language.
// This is a Value Object shared across all contexts.
type Language struct {
	ID         uuidv7.UUID // Unique identifier
	Code       string      // ISO 639-1 alpha-2 (en, uk, de, etc.)
	Code3      string      // ISO 639-2/T alpha-3 (eng, ukr, deu, etc.)
	Name       string      // English name
	NativeName string      // Native name (English, Українська, Deutsch, etc.)
	IsActive   bool        // Whether language is active in system
}

// NewLanguage creates a new Language value object with validation.
func NewLanguage(code, name, nativeName string) (Language, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	nativeName = strings.TrimSpace(nativeName)

	if len(code) != 2 {
		return Language{}, fmt.Errorf("language code must be 2 characters (ISO 639-1)")
	}

	if name == "" {
		return Language{}, fmt.Errorf("language name is required")
	}

	if nativeName == "" {
		nativeName = name
	}

	return Language{
		ID:         uuidv7.New(),
		Code:       code,
		Name:       name,
		NativeName: nativeName,
		IsActive:   true,
	}, nil
}

// String returns the language name.
func (l Language) String() string {
	return l.Name
}

// Equals checks if two languages are the same (by ISO code).
func (l Language) Equals(other Language) bool {
	return l.Code == other.Code
}
