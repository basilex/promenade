package entity

import (
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestUserProfile_Fields(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
	}

	assert.NotNil(t, profile)
	assert.Equal(t, "UTC", profile.Timezone)
	assert.Equal(t, "en", profile.Locale)
}

func TestGender_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		gender Gender
		want   bool
	}{
		{"male", GenderMale, true},
		{"female", GenderFemale, true},
		{"non-binary", GenderNonBinary, true},
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.gender.IsValid())
		})
	}
}

func TestUserProfile_Privacy(t *testing.T) {
	profile := &UserProfile{
		ID:          uuidv7.New(),
		UserID:      uuidv7.New(),
		Timezone:    "UTC",
		Locale:      "en",
		IsPublic:    true,
		ShowEmail:   false,
		ShowLocation: true,
	}

	assert.True(t, profile.IsPublic)
	assert.False(t, profile.ShowEmail)
	assert.True(t, profile.ShowLocation)
}
