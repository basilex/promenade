package entity

import (
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestGender_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		gender  Gender
		isValid bool
	}{
		{"male", GenderMale, true},
		{"female", GenderFemale, true},
		{"non-binary", GenderNonBinary, true},
		{"other", GenderOther, true},
		{"prefer-not-to-say", GenderPreferNotToSay, true},
		{"invalid", Gender("invalid"), false},
		{"empty", Gender(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.gender.IsValid())
		})
	}
}

func TestUserProfile_Validate(t *testing.T) {
	validDisplayName := "John Doe"
	validNickname := "johndoe"
	shortNickname := "ab"
	longBio := string(make([]byte, 1001))
	validGender := string(GenderMale)
	invalidGender := "invalid"

	tests := []struct {
		name    string
		profile *UserProfile
		wantErr bool
	}{
		{
			name: "valid profile",
			profile: &UserProfile{
				DisplayName: &validDisplayName,
				Nickname:    &validNickname,
				Gender:      &validGender,
			},
			wantErr: false,
		},
		{
			name: "nil display name",
			profile: &UserProfile{
				DisplayName: nil,
			},
			wantErr: true,
		},
		{
			name: "short nickname",
			profile: &UserProfile{
				DisplayName: &validDisplayName,
				Nickname:    &shortNickname,
			},
			wantErr: true,
		},
		{
			name: "long bio",
			profile: &UserProfile{
				DisplayName: &validDisplayName,
				Bio:         &longBio,
			},
			wantErr: true,
		},
		{
			name: "invalid gender",
			profile: &UserProfile{
				DisplayName: &validDisplayName,
				Gender:      &invalidGender,
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

func TestUserProfile_IsOwnProfile(t *testing.T) {
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	profile := &UserProfile{
		UserID: userID,
	}

	assert.True(t, profile.IsOwnProfile(userID))
	assert.False(t, profile.IsOwnProfile(otherUserID))
}

func TestUserProfile_CanView(t *testing.T) {
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	tests := []struct {
		name     string
		profile  *UserProfile
		viewerID uuidv7.UUID
		canView  bool
	}{
		{
			name: "banned profile",
			profile: &UserProfile{
				UserID:   userID,
				IsBanned: true,
			},
			viewerID: otherUserID,
			canView:  false,
		},
		{
			name: "private profile - owner viewing",
			profile: &UserProfile{
				UserID:   userID,
				IsPublic: false,
			},
			viewerID: userID,
			canView:  true,
		},
		{
			name: "private profile - other viewing",
			profile: &UserProfile{
				UserID:   userID,
				IsPublic: false,
			},
			viewerID: otherUserID,
			canView:  false,
		},
		{
			name: "public profile - anyone viewing",
			profile: &UserProfile{
				UserID:   userID,
				IsPublic: true,
			},
			viewerID: otherUserID,
			canView:  true,
		},
		{
			name: "public profile - anonymous viewing",
			profile: &UserProfile{
				UserID:   userID,
				IsPublic: true,
			},
			viewerID: uuidv7.Nil,
			canView:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.canView, tt.profile.CanView(tt.viewerID))
		})
	}
}

func TestUserProfile_Ban(t *testing.T) {
	profile := &UserProfile{
		IsBanned: false,
	}

	adminID := uuidv7.New()
	reason := "Violation of terms"

	profile.Ban(reason, adminID)

	assert.True(t, profile.IsBanned)
	assert.NotNil(t, profile.BanReason)
	assert.Equal(t, reason, *profile.BanReason)
	assert.NotNil(t, profile.BannedAt)
	assert.NotNil(t, profile.BannedBy)
	assert.Equal(t, adminID, *profile.BannedBy)
}

func TestUserProfile_Unban(t *testing.T) {
	reason := "Violation of terms"
	adminID := uuidv7.New()

	profile := &UserProfile{
		IsBanned:  true,
		BanReason: &reason,
		BannedBy:  &adminID,
	}

	profile.Unban()

	assert.False(t, profile.IsBanned)
	assert.Nil(t, profile.BanReason)
	assert.Nil(t, profile.BannedAt)
	assert.Nil(t, profile.BannedBy)
}

func TestUserProfile_Verify(t *testing.T) {
	profile := &UserProfile{
		IsVerified: false,
	}

	profile.Verify()
	assert.True(t, profile.IsVerified)
}

func TestUserProfile_Unverify(t *testing.T) {
	profile := &UserProfile{
		IsVerified: true,
	}

	profile.Unverify()
	assert.False(t, profile.IsVerified)
}

func TestUserProfile_UpdateLastSeen(t *testing.T) {
	profile := &UserProfile{
		LastSeenAt: nil,
	}

	profile.UpdateLastSeen()
	assert.NotNil(t, profile.LastSeenAt)
}

func TestUserProfile_IncrementProfileViews(t *testing.T) {
	profile := &UserProfile{
		ProfileViewsCount: 10,
	}

	profile.IncrementProfileViews()
	assert.Equal(t, 11, profile.ProfileViewsCount)

	profile.IncrementProfileViews()
	assert.Equal(t, 12, profile.ProfileViewsCount)
}

func TestSocialLinks_Marshal_Unmarshal(t *testing.T) {
	profile := &UserProfile{
		SocialLinks: SocialLinks{
			"twitter":  "https://twitter.com/johndoe",
			"linkedin": "https://linkedin.com/in/johndoe",
			"github":   "https://github.com/johndoe",
		},
	}

	// Marshal
	err := profile.MarshalSocialLinks()
	assert.NoError(t, err)
	assert.NotNil(t, profile.SocialLinksJSON)

	// Clear original data
	profile.SocialLinks = nil

	// Unmarshal
	err = profile.UnmarshalSocialLinks()
	assert.NoError(t, err)
	assert.NotNil(t, profile.SocialLinks)
	assert.Equal(t, 3, len(profile.SocialLinks))
	assert.Equal(t, "https://twitter.com/johndoe", profile.SocialLinks["twitter"])
}

func TestPreferences_Marshal_Unmarshal(t *testing.T) {
	profile := &UserProfile{
		Preferences: Preferences{
			"theme":            "dark",
			"notifications":    true,
			"privacy_level":    2,
			"language":         "en",
			"timezone_display": "12h",
		},
	}

	// Marshal
	err := profile.MarshalPreferences()
	assert.NoError(t, err)
	assert.NotNil(t, profile.PreferencesJSON)

	// Clear original data
	profile.Preferences = nil

	// Unmarshal
	err = profile.UnmarshalPreferences()
	assert.NoError(t, err)
	assert.NotNil(t, profile.Preferences)
	assert.Equal(t, 5, len(profile.Preferences))
	assert.Equal(t, "dark", profile.Preferences["theme"])
	assert.Equal(t, true, profile.Preferences["notifications"])
	assert.Equal(t, float64(2), profile.Preferences["privacy_level"]) // JSON numbers are float64
}
