package language

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Language represents an ISO 639-1 language (Aggregate Root in Shared Context)
type Language struct {
	ID             uuidv7.UUID `db:"id"`
	Code           string      `db:"code"`
	Code3          string      `db:"code3"`
	Name           string      `db:"name"`
	NativeName     string      `db:"native_name"`
	Direction      string      `db:"direction"`
	NativeSpeakers *int64      `db:"native_speakers"`
	IsActive       bool        `db:"is_active"`
}

// NewLanguage creates a new Language entity with validation
func NewLanguage(code, name, nativeName string) (*Language, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	nativeName = strings.TrimSpace(nativeName)

	if len(code) != 2 {
		return nil, fmt.Errorf("language code must be 2 characters (ISO 639-1)")
	}

	if name == "" {
		return nil, fmt.Errorf("language name is required")
	}

	if nativeName == "" {
		return nil, fmt.Errorf("native name is required")
	}

	return &Language{
		ID:         uuidv7.New(),
		Code:       code,
		Name:       name,
		NativeName: nativeName,
		IsActive:   true,
	}, nil
}

// Validate validates language data
func (l *Language) Validate() error {
	if len(l.Code) != 2 {
		return fmt.Errorf("language code must be 2 characters")
	}
	if l.Name == "" {
		return fmt.Errorf("language name is required")
	}
	if l.NativeName == "" {
		return fmt.Errorf("native name is required")
	}
	return nil
}

// String returns the language code
func (l *Language) String() string {
	return l.Code
}

// ParseUUID parses a string UUID and returns error on failure
func ParseUUID(s string) (uuidv7.UUID, error) {
return uuidv7.Parse(s)
}
