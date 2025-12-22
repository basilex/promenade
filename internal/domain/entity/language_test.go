package entity

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestLanguage_Validate(t *testing.T) {
	tests := []struct {
		name     string
		language *Language
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid language",
			language: &Language{
				ID:         uuidv7.New(),
				Name:       "English",
				NativeName: "English",
				Code:       "en",
				ISO639_2:   "eng",
				IsActive:   true,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			language: &Language{
				ID:         uuidv7.New(),
				NativeName: "English",
				Code:       "en",
				ISO639_2:   "eng",
			},
			wantErr: true,
			errMsg:  "language name is required",
		},
		{
			name: "missing native name",
			language: &Language{
				ID:       uuidv7.New(),
				Name:     "English",
				Code:     "en",
				ISO639_2: "eng",
			},
			wantErr: true,
			errMsg:  "native name is required",
		},
		{
			name: "missing code",
			language: &Language{
				ID:         uuidv7.New(),
				Name:       "English",
				NativeName: "English",
				ISO639_2:   "eng",
			},
			wantErr: true,
			errMsg:  "language code is required",
		},
		{
			name: "invalid code length",
			language: &Language{
				ID:         uuidv7.New(),
				Name:       "English",
				NativeName: "English",
				Code:       "e",
				ISO639_2:   "eng",
			},
			wantErr: true,
			errMsg:  "language code must be 2 characters (ISO 639-1)",
		},
		{
			name: "missing ISO639_2",
			language: &Language{
				ID:         uuidv7.New(),
				Name:       "English",
				NativeName: "English",
				Code:       "en",
			},
			wantErr: true,
			errMsg:  "ISO 639-2 code is required",
		},
		{
			name: "invalid ISO639_2 length",
			language: &Language{
				ID:         uuidv7.New(),
				Name:       "English",
				NativeName: "English",
				Code:       "en",
				ISO639_2:   "en",
			},
			wantErr: true,
			errMsg:  "ISO 639-2 code must be 3 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.language.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewLanguage(t *testing.T) {
	name := "English"
	nativeName := "English"
	code := "en"
	iso639_2 := "eng"
	isRtl := false

	language := NewLanguage(name, nativeName, code, iso639_2, isRtl)

	assert.NotEmpty(t, language.ID)
	assert.Equal(t, name, language.Name)
	assert.Equal(t, nativeName, language.NativeName)
	assert.Equal(t, code, language.Code)
	assert.Equal(t, iso639_2, language.ISO639_2)
	assert.Equal(t, isRtl, language.IsRtl)
	assert.True(t, language.IsActive)
	assert.NotZero(t, language.CreatedAt)
	assert.NotZero(t, language.UpdatedAt)
}

func TestLanguage_RTL(t *testing.T) {
	tests := []struct {
		name     string
		language *Language
		wantRTL  bool
	}{
		{
			name:     "LTR language (English)",
			language: NewLanguage("English", "English", "en", "eng", false),
			wantRTL:  false,
		},
		{
			name:     "RTL language (Arabic)",
			language: NewLanguage("Arabic", "العربية", "ar", "ara", true),
			wantRTL:  true,
		},
		{
			name:     "RTL language (Hebrew)",
			language: NewLanguage("Hebrew", "עברית", "he", "heb", true),
			wantRTL:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantRTL, tt.language.IsRtl)
		})
	}
}
