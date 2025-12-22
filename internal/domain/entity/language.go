package entity

import (
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Language represents a language entity
type Language struct {
	ID         uuidv7.UUID `db:"id" json:"id"`
	Name       string      `db:"name" json:"name" validate:"required,min=2,max=100"`
	NativeName string      `db:"native_name" json:"native_name" validate:"required,min=2,max=100"`
	Code       string      `db:"code" json:"code" validate:"required,len=2,alpha,lowercase"`
	ISO639_2   string      `db:"iso639_2" json:"iso639_2" validate:"required,len=3,alpha,lowercase"`
	IsRtl      bool        `db:"is_rtl" json:"is_rtl"`
	IsActive   bool        `db:"is_active" json:"is_active"`
	SortOrder  int         `db:"sort_order" json:"sort_order"`
	CreatedAt  time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at" json:"updated_at"`
}

// Validate validates language fields
func (l *Language) Validate() error {
	if l.Name == "" {
		return errors.New("language name is required")
	}
	if l.NativeName == "" {
		return errors.New("native name is required")
	}
	if l.Code == "" {
		return errors.New("language code is required")
	}
	if len(l.Code) != 2 {
		return errors.New("language code must be 2 characters (ISO 639-1)")
	}
	if l.ISO639_2 == "" {
		return errors.New("ISO 639-2 code is required")
	}
	if len(l.ISO639_2) != 3 {
		return errors.New("ISO 639-2 code must be 3 characters")
	}
	return nil
}

// NewLanguage creates a new language with generated UUID v7
func NewLanguage(name, nativeName, code, iso639_2 string, isRtl bool) *Language {
	now := time.Now()
	return &Language{
		ID:         uuidv7.New(),
		Name:       name,
		NativeName: nativeName,
		Code:       code,
		ISO639_2:   iso639_2,
		IsRtl:      isRtl,
		IsActive:   true,
		SortOrder:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// GetLocaleCode returns locale code (e.g., "en-US", "ru-RU")
func (l *Language) GetLocaleCode(countryCode string) string {
	return l.Code + "-" + countryCode
}
