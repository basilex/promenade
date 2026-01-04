package profile

import (
	"strings"
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfile(t *testing.T) {
	userID := uuidv7.New()

	t.Run("valid profile", func(t *testing.T) {
		profile, err := NewProfile(userID, "John Doe")

		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, profile.ID)
		assert.Equal(t, userID, profile.UserID)
		assert.Equal(t, "John Doe", profile.DisplayName)
		assert.Equal(t, GenderNotSpecify, profile.Gender)
		assert.True(t, profile.IsPublic)
		assert.True(t, profile.IsActive)
		assert.NotZero(t, profile.CreatedAt)
		assert.NotZero(t, profile.UpdatedAt)
	})

	t.Run("empty display name", func(t *testing.T) {
		_, err := NewProfile(userID, "")
		assert.Error(t, err)
	})

	t.Run("display name too short", func(t *testing.T) {
		_, err := NewProfile(userID, "A")
		assert.Error(t, err)
	})

	t.Run("display name too long", func(t *testing.T) {
		longName := strings.Repeat("a", 101)
		_, err := NewProfile(userID, longName)
		assert.Error(t, err)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		profile, err := NewProfile(userID, "  John Doe  ")
		require.NoError(t, err)
		assert.Equal(t, "John Doe", profile.DisplayName)
	})
}

func TestProfile_UpdateDisplayName(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
		DisplayName: "Original Name",
	}

	t.Run("valid update", func(t *testing.T) {
		err := profile.UpdateDisplayName("New Name")
		assert.NoError(t, err)
		assert.Equal(t, "New Name", profile.DisplayName)
	})

	t.Run("empty name", func(t *testing.T) {
		err := profile.UpdateDisplayName("")
		assert.Error(t, err)
	})

	t.Run("name too short", func(t *testing.T) {
		err := profile.UpdateDisplayName("A")
		assert.Error(t, err)
	})
}

func TestProfile_UpdateBio(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
	}

	t.Run("valid bio", func(t *testing.T) {
		err := profile.UpdateBio("This is my bio")
		assert.NoError(t, err)
		assert.Equal(t, "This is my bio", profile.Bio)
	})

	t.Run("empty bio", func(t *testing.T) {
		err := profile.UpdateBio("")
		assert.NoError(t, err)
		assert.Equal(t, "", profile.Bio)
	})

	t.Run("bio too long", func(t *testing.T) {
		longBio := strings.Repeat("a", 501)
		err := profile.UpdateBio(longBio)
		assert.Error(t, err)
	})
}

func TestProfile_UpdateAvatar(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
	}

	t.Run("valid HTTPS URL", func(t *testing.T) {
		err := profile.UpdateAvatar("https://example.com/avatar.jpg")
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/avatar.jpg", profile.AvatarURL)
	})

	t.Run("valid HTTP URL", func(t *testing.T) {
		err := profile.UpdateAvatar("http://example.com/avatar.jpg")
		assert.NoError(t, err)
	})

	t.Run("empty URL", func(t *testing.T) {
		err := profile.UpdateAvatar("")
		assert.NoError(t, err)
		assert.Equal(t, "", profile.AvatarURL)
	})

	t.Run("invalid URL", func(t *testing.T) {
		err := profile.UpdateAvatar("not-a-url")
		assert.Error(t, err)
	})
}

func TestProfile_UpdatePersonalInfo(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
	}

	t.Run("valid personal info", func(t *testing.T) {
		err := profile.UpdatePersonalInfo("John", "Doe", "Michael")
		assert.NoError(t, err)
		assert.Equal(t, "John", profile.FirstName)
		assert.Equal(t, "Doe", profile.LastName)
		assert.Equal(t, "Michael", profile.MiddleName)
	})

	t.Run("empty middle name", func(t *testing.T) {
		err := profile.UpdatePersonalInfo("John", "Doe", "")
		assert.NoError(t, err)
		assert.Equal(t, "", profile.MiddleName)
	})

	t.Run("name too long", func(t *testing.T) {
		longName := strings.Repeat("a", 51)
		err := profile.UpdatePersonalInfo(longName, "Doe", "")
		assert.Error(t, err)
	})
}

func TestProfile_UpdateGender(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
		Gender: GenderNotSpecify,
	}

	t.Run("valid genders", func(t *testing.T) {
		validGenders := []Gender{GenderMale, GenderFemale, GenderOther, GenderNotSpecify}
		for _, gender := range validGenders {
			err := profile.UpdateGender(gender)
			assert.NoError(t, err)
			assert.Equal(t, gender, profile.Gender)
		}
	})

	t.Run("invalid gender", func(t *testing.T) {
		err := profile.UpdateGender(Gender("invalid"))
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidGender, err)
	})
}

func TestProfile_UpdateDateOfBirth(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
	}

	t.Run("valid date of birth", func(t *testing.T) {
		dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
		err := profile.UpdateDateOfBirth(&dob)
		assert.NoError(t, err)
		assert.NotNil(t, profile.DateOfBirth)
	})

	t.Run("too young", func(t *testing.T) {
		dob := time.Now().AddDate(-10, 0, 0) // 10 years old
		err := profile.UpdateDateOfBirth(&dob)
		assert.Error(t, err)
	})

	t.Run("too old", func(t *testing.T) {
		dob := time.Now().AddDate(-150, 0, 0) // 150 years old
		err := profile.UpdateDateOfBirth(&dob)
		assert.Error(t, err)
	})

	t.Run("nil date of birth", func(t *testing.T) {
		err := profile.UpdateDateOfBirth(nil)
		assert.NoError(t, err)
		assert.Nil(t, profile.DateOfBirth)
	})
}

func TestProfile_UpdateLocalization(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
	}

	t.Run("valid localization", func(t *testing.T) {
		err := profile.UpdateLocalization("Europe/Kyiv", "uk", "UA")
		assert.NoError(t, err)
		assert.Equal(t, "Europe/Kyiv", profile.Timezone)
		assert.Equal(t, "uk", profile.Language)
		assert.Equal(t, "UA", profile.Country)
	})

	t.Run("invalid language code", func(t *testing.T) {
		err := profile.UpdateLocalization("Europe/Kyiv", "ukr", "UA")
		assert.Error(t, err)
	})

	t.Run("invalid country code", func(t *testing.T) {
		err := profile.UpdateLocalization("Europe/Kyiv", "uk", "UKR")
		assert.Error(t, err)
	})

	t.Run("case normalization", func(t *testing.T) {
		err := profile.UpdateLocalization("Europe/Kyiv", "UK", "ua")
		assert.NoError(t, err)
		assert.Equal(t, "uk", profile.Language) // lowercase
		assert.Equal(t, "UA", profile.Country)  // uppercase
	})
}

func TestProfile_UpdateSocialLinks(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
	}

	t.Run("valid social links", func(t *testing.T) {
		err := profile.UpdateSocialLinks(
			"https://example.com",
			"https://linkedin.com/in/john",
			"https://twitter.com/john",
			"https://github.com/john",
			"https://facebook.com/john",
			"https://instagram.com/john",
		)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", profile.Website)
		assert.Equal(t, "https://linkedin.com/in/john", profile.LinkedIn)
	})

	t.Run("empty links", func(t *testing.T) {
		err := profile.UpdateSocialLinks("", "", "", "", "", "")
		assert.NoError(t, err)
		assert.Equal(t, "", profile.Website)
	})

	t.Run("invalid URL", func(t *testing.T) {
		err := profile.UpdateSocialLinks("not-a-url", "", "", "", "", "")
		assert.Error(t, err)
	})
}

func TestProfile_Visibility(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
		IsPublic: true,
	}

	t.Run("set private", func(t *testing.T) {
		profile.SetPrivate()
		assert.False(t, profile.IsPublic)
	})

	t.Run("set public", func(t *testing.T) {
		profile.SetPublic()
		assert.True(t, profile.IsPublic)
	})
}

func TestProfile_Status(t *testing.T) {
	profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
		IsActive: true,
	}

	t.Run("deactivate", func(t *testing.T) {
		profile.Deactivate()
		assert.False(t, profile.IsActive)
	})

	t.Run("activate", func(t *testing.T) {
		profile.Activate()
		assert.True(t, profile.IsActive)
	})
}

func TestProfile_GetFullName(t *testing.T) {
	t.Run("full name", func(t *testing.T) {
		profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
			FirstName:  "John",
			MiddleName: "Michael",
			LastName:   "Doe",
		}
		assert.Equal(t, "John Michael Doe", profile.GetFullName())
	})

	t.Run("no middle name", func(t *testing.T) {
		profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
			FirstName: "John",
			LastName:  "Doe",
		}
		assert.Equal(t, "John Doe", profile.GetFullName())
	})

	t.Run("only first name", func(t *testing.T) {
		profile := &Profile{
		BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
			FirstName: "John",
		}
		assert.Equal(t, "John", profile.GetFullName())
	})

	t.Run("empty", func(t *testing.T) {
		profile := &Profile{}
		assert.Equal(t, "", profile.GetFullName())
	})
}

func TestProfile_Validate(t *testing.T) {
	t.Run("valid profile", func(t *testing.T) {
		profile := &Profile{
			UserID:      uuidv7.New(),
			DisplayName: "John Doe",
			Gender:      GenderMale,
		}
		profile.ID = uuidv7.New()
		err := profile.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing ID", func(t *testing.T) {
		profile := &Profile{
			DisplayName: "John Doe",
		}
		err := profile.Validate()
		assert.Error(t, err)
	})

	t.Run("missing user ID", func(t *testing.T) {
		profile := &Profile{
			DisplayName: "John Doe",
		}
		err := profile.Validate()
		assert.Error(t, err)
	})

	t.Run("missing display name", func(t *testing.T) {
		profile := &Profile{}
		err := profile.Validate()
		assert.Error(t, err)
	})

	t.Run("invalid gender", func(t *testing.T) {
		profile := &Profile{
			UserID:      uuidv7.New(),
			DisplayName: "John Doe",
			Gender:      Gender("invalid"),
		}
		profile.ID = uuidv7.New()
		err := profile.Validate()
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidGender, err)
	})
}
