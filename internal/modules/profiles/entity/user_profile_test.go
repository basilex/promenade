package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/uuidv7"
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
		ID:           uuidv7.New(),
		UserID:       uuidv7.New(),
		Timezone:     "UTC",
		Locale:       "en",
		IsPublic:     true,
		ShowEmail:    false,
		ShowLocation: true,
	}

	assert.True(t, profile.IsPublic)
	assert.False(t, profile.ShowEmail)
	assert.True(t, profile.ShowLocation)
}

func TestUserProfile_Validate(t *testing.T) {
	userID := uuidv7.New()

	tests := []struct {
		name    string
		profile *UserProfile
		wantErr bool
	}{
		{
			name: "valid profile",
			profile: &UserProfile{
				ID:       uuidv7.New(),
				UserID:   userID,
				Timezone: "UTC",
				Locale:   "en",
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			profile: &UserProfile{
				ID:       uuidv7.New(),
				Timezone: "UTC",
				Locale:   "en",
			},
			wantErr: true,
		},
		{
			name: "missing timezone",
			profile: &UserProfile{
				ID:     uuidv7.New(),
				UserID: userID,
				Locale: "en",
			},
			wantErr: true,
		},
		{
			name: "negative stats",
			profile: &UserProfile{
				ID:                uuidv7.New(),
				UserID:            userID,
				Timezone:          "UTC",
				Locale:            "en",
				ProfileViewsCount: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserProfile_SocialLinks(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
		SocialLinks: SocialLinks{
			"twitter":  "https://twitter.com/user",
			"linkedin": "https://linkedin.com/in/user",
		},
	}

	err := profile.MarshalSocialLinks()
	assert.NoError(t, err)
	assert.NotEmpty(t, profile.SocialLinksJSON)

	profile.SocialLinks = nil
	err = profile.UnmarshalSocialLinks()
	assert.NoError(t, err)
	assert.Equal(t, "https://twitter.com/user", profile.SocialLinks["twitter"])
}

func TestUserProfile_Preferences(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
		Preferences: Preferences{
			"theme":         "dark",
			"notifications": true,
		},
	}

	err := profile.MarshalPreferences()
	assert.NoError(t, err)
	assert.NotEmpty(t, profile.PreferencesJSON)

	profile.Preferences = nil
	err = profile.UnmarshalPreferences()
	assert.NoError(t, err)
	assert.Equal(t, "dark", profile.Preferences["theme"])
}

func TestUserProfile_CanView(t *testing.T) {
	ownerID := uuidv7.New()
	viewerID := uuidv7.New()

	tests := []struct {
		name     string
		profile  *UserProfile
		viewerID uuidv7.UUID
		want     bool
	}{
		{
			name: "public profile",
			profile: &UserProfile{
				ID:       uuidv7.New(),
				UserID:   ownerID,
				Timezone: "UTC",
				Locale:   "en",
				IsPublic: true,
			},
			viewerID: viewerID,
			want:     true,
		},
		{
			name: "banned profile",
			profile: &UserProfile{
				ID:       uuidv7.New(),
				UserID:   ownerID,
				Timezone: "UTC",
				Locale:   "en",
				IsPublic: true,
				IsBanned: true,
			},
			viewerID: viewerID,
			want:     false,
		},
		{
			name: "owner viewing own private profile",
			profile: &UserProfile{
				ID:       uuidv7.New(),
				UserID:   ownerID,
				Timezone: "UTC",
				Locale:   "en",
				IsPublic: false,
			},
			viewerID: ownerID,
			want:     true,
		},
		{
			name: "stranger viewing private profile",
			profile: &UserProfile{
				ID:       uuidv7.New(),
				UserID:   ownerID,
				Timezone: "UTC",
				Locale:   "en",
				IsPublic: false,
			},
			viewerID: viewerID,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.profile.CanView(tt.viewerID))
		})
	}
}

func TestUserProfile_Ban(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
	}

	bannedBy := uuidv7.New()
	profile.Ban("spam", bannedBy)

	assert.True(t, profile.IsBanned)
	assert.Equal(t, "spam", *profile.BanReason)
	assert.Equal(t, bannedBy, *profile.BannedBy)
	assert.NotNil(t, profile.BannedAt)
}

func TestUserProfile_Unban(t *testing.T) {
	bannedBy := uuidv7.New()
	reason := "spam"
	now := time.Now()

	profile := &UserProfile{
		ID:         uuidv7.New(),
		UserID:     uuidv7.New(),
		Timezone:   "UTC",
		Locale:     "en",
		IsBanned:   true,
		BanReason:  &reason,
		BannedBy:   &bannedBy,
		BannedAt:   &now,
	}

	profile.Unban()

	assert.False(t, profile.IsBanned)
	assert.Nil(t, profile.BanReason)
	assert.Nil(t, profile.BannedBy)
	assert.Nil(t, profile.BannedAt)
}

func TestUserProfile_Verify(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
	}

	assert.False(t, profile.IsVerified)
	profile.Verify()
	assert.True(t, profile.IsVerified)

	profile.Unverify()
	assert.False(t, profile.IsVerified)
}

func TestUserProfile_IncrementProfileViews(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
	}

	assert.Equal(t, 0, profile.ProfileViewsCount)
	profile.IncrementProfileViews()
	assert.Equal(t, 1, profile.ProfileViewsCount)
	profile.IncrementProfileViews()
	assert.Equal(t, 2, profile.ProfileViewsCount)
}

func TestUserProfile_UpdateLastSeen(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
	}

	assert.Nil(t, profile.LastSeenAt)
	profile.UpdateLastSeen()
	assert.NotNil(t, profile.LastSeenAt)
}

func TestUserProfile_IsOwnProfile(t *testing.T) {
	userID := uuidv7.New()
	otherID := uuidv7.New()

	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   userID,
		Timezone: "UTC",
		Locale:   "en",
	}

	assert.True(t, profile.IsOwnProfile(userID))
	assert.False(t, profile.IsOwnProfile(otherID))
}

func TestUserProfile_EmptySocialLinks(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
	}

	err := profile.MarshalSocialLinks()
	assert.NoError(t, err)
	assert.Equal(t, []byte("{}"), profile.SocialLinksJSON)
}

func TestUserProfile_EmptyPreferences(t *testing.T) {
	profile := &UserProfile{
		ID:       uuidv7.New(),
		UserID:   uuidv7.New(),
		Timezone: "UTC",
		Locale:   "en",
	}

	err := profile.MarshalPreferences()
	assert.NoError(t, err)
	assert.Equal(t, []byte("{}"), profile.PreferencesJSON)
}

func TestContactType_ValidContactTypes(t *testing.T) {
	types := ValidContactTypes()
	assert.Len(t, types, 10)
	assert.Contains(t, types, ContactTypeEmail)
	assert.Contains(t, types, ContactTypePhone)
}

func TestIsValidContactType(t *testing.T) {
	assert.True(t, IsValidContactType("email"))
	assert.True(t, IsValidContactType("phone"))
	assert.False(t, IsValidContactType("invalid"))
}

