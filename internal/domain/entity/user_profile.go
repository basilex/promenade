package entity

import (
	"encoding/json"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Gender represents user's gender
type Gender string

const (
	GenderMale           Gender = "male"
	GenderFemale         Gender = "female"
	GenderNonBinary      Gender = "non-binary"
	GenderOther          Gender = "other"
	GenderPreferNotToSay Gender = "prefer-not-to-say"
)

// IsValid checks if gender value is valid
func (g Gender) IsValid() bool {
	switch g {
	case GenderMale, GenderFemale, GenderNonBinary, GenderOther, GenderPreferNotToSay:
		return true
	default:
		return false
	}
}

// SocialLinks represents user's social media links
type SocialLinks map[string]string

// Preferences represents user's preferences
type Preferences map[string]any

// UserProfile represents a user profile with personal information
type UserProfile struct {
	ID     uuidv7.UUID `db:"id"`
	UserID uuidv7.UUID `db:"user_id"`

	// Personal Information
	FirstName   *string `db:"first_name"`
	LastName    *string `db:"last_name"`
	MiddleName  *string `db:"middle_name"`
	DisplayName *string `db:"display_name"`
	Nickname    *string `db:"nickname"`

	// Biographical
	Bio         *string    `db:"bio"`
	DateOfBirth *time.Time `db:"date_of_birth"`
	Gender      *string    `db:"gender"`

	// Location & Localization
	CountryID *uuidv7.UUID `db:"country_id"`
	City      *string      `db:"city"`
	Timezone  string       `db:"timezone"`
	Locale    string       `db:"locale"`

	// Visual Identity
	AvatarURL *string `db:"avatar_url"`
	CoverURL  *string `db:"cover_url"`

	// Social & Web
	SocialLinksJSON []byte      `db:"social_links"`
	SocialLinks     SocialLinks `db:"-"`
	WebsiteURL      *string     `db:"website_url"`
	Company         *string     `db:"company"`
	JobTitle        *string     `db:"job_title"`

	// Privacy & Verification
	IsPublic     bool `db:"is_public"`
	IsVerified   bool `db:"is_verified"`
	ShowEmail    bool `db:"show_email"`
	ShowLocation bool `db:"show_location"`
	ShowBirthday bool `db:"show_birthday"`

	// Preferences
	PreferencesJSON []byte      `db:"preferences"`
	Preferences     Preferences `db:"-"`

	// Statistics
	ProfileViewsCount int `db:"profile_views_count"`
	FollowersCount    int `db:"followers_count"`
	FollowingCount    int `db:"following_count"`

	// Moderation
	IsBanned  bool         `db:"is_banned"`
	BanReason *string      `db:"ban_reason"`
	BannedAt  *time.Time   `db:"banned_at"`
	BannedBy  *uuidv7.UUID `db:"banned_by"`

	// Timestamps
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	LastSeenAt *time.Time `db:"last_seen_at"`
}

// Validate validates user profile data
func (p *UserProfile) Validate() error {
	// Nickname validation
	if p.Nickname != nil && len(*p.Nickname) < 3 {
		return ErrInvalidInput
	}

	// Bio length validation
	if p.Bio != nil && len(*p.Bio) > 1000 {
		return ErrInvalidInput
	}

	// Gender validation
	if p.Gender != nil {
		gender := Gender(*p.Gender)
		if !gender.IsValid() {
			return ErrInvalidInput
		}
	}

	// Display name validation
	if p.DisplayName == nil || len(*p.DisplayName) == 0 {
		return ErrInvalidInput
	}

	return nil
}

// MarshalSocialLinks converts SocialLinks to JSON
func (p *UserProfile) MarshalSocialLinks() error {
	if p.SocialLinks == nil {
		p.SocialLinksJSON = []byte("{}")
		return nil
	}

	data, err := json.Marshal(p.SocialLinks)
	if err != nil {
		return err
	}
	p.SocialLinksJSON = data
	return nil
}

// UnmarshalSocialLinks converts JSON to SocialLinks
func (p *UserProfile) UnmarshalSocialLinks() error {
	if len(p.SocialLinksJSON) == 0 {
		p.SocialLinks = make(SocialLinks)
		return nil
	}

	return json.Unmarshal(p.SocialLinksJSON, &p.SocialLinks)
}

// MarshalPreferences converts Preferences to JSON
func (p *UserProfile) MarshalPreferences() error {
	if p.Preferences == nil {
		p.PreferencesJSON = []byte("{}")
		return nil
	}

	data, err := json.Marshal(p.Preferences)
	if err != nil {
		return err
	}
	p.PreferencesJSON = data
	return nil
}

// UnmarshalPreferences converts JSON to Preferences
func (p *UserProfile) UnmarshalPreferences() error {
	if len(p.PreferencesJSON) == 0 {
		p.Preferences = make(Preferences)
		return nil
	}

	return json.Unmarshal(p.PreferencesJSON, &p.Preferences)
}

// IsOwnProfile checks if profile belongs to given user
func (p *UserProfile) IsOwnProfile(userID uuidv7.UUID) bool {
	return p.UserID == userID
}

// CanView checks if user can view this profile
func (p *UserProfile) CanView(viewerID uuidv7.UUID) bool {
	// Banned profiles are not viewable
	if p.IsBanned {
		return false
	}

	// Public profiles are viewable by anyone
	if p.IsPublic {
		return true
	}

	// Owner can always view own profile
	if viewerID != uuidv7.Nil && p.UserID == viewerID {
		return true
	}

	// Private profiles are not viewable by others
	return false
}

// UpdateLastSeen updates the last seen timestamp
func (p *UserProfile) UpdateLastSeen() {
	now := time.Now()
	p.LastSeenAt = &now
}

// IncrementProfileViews increments profile view counter
func (p *UserProfile) IncrementProfileViews() {
	p.ProfileViewsCount++
}

// Ban bans the profile
func (p *UserProfile) Ban(reason string, bannedBy uuidv7.UUID) {
	p.IsBanned = true
	p.BanReason = &reason
	now := time.Now()
	p.BannedAt = &now
	p.BannedBy = &bannedBy
}

// Unban removes ban from profile
func (p *UserProfile) Unban() {
	p.IsBanned = false
	p.BanReason = nil
	p.BannedAt = nil
	p.BannedBy = nil
}

// Verify marks profile as verified
func (p *UserProfile) Verify() {
	p.IsVerified = true
}

// Unverify removes verification from profile
func (p *UserProfile) Unverify() {
	p.IsVerified = false
}
