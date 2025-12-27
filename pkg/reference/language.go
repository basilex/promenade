package reference

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Language represents an ISO 639-1 language.
// This is a Value Object shared across all contexts.
type Language struct {
	ID         uuidv7.UUID `db:"id"`
	Code       string      `db:"code"`
	Code3      string      `db:"code3"`
	Name       string      `db:"name"`
	NativeName string      `db:"native_name"`
	IsActive   bool        `db:"is_active"`
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
