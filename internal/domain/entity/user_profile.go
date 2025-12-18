package entity

import (
	"encoding/json"
	"fmt"
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
	ID     uuidv7.UUID `db:"id" validate:"required"`
	UserID uuidv7.UUID `db:"user_id" validate:"required"`

	// Personal Information
	FirstName   *string `db:"first_name" validate:"omitempty,max=100"`
	LastName    *string `db:"last_name" validate:"omitempty,max=100"`
	MiddleName  *string `db:"middle_name" validate:"omitempty,max=100"`
	DisplayName *string `db:"display_name" validate:"omitempty,max=100"`
	Nickname    *string `db:"nickname" validate:"omitempty,min=3,max=50,alphanum"`

	// Biographical
	Bio         *string    `db:"bio" validate:"omitempty,max=1000"`
	DateOfBirth *time.Time `db:"date_of_birth" validate:"omitempty"`
	Gender      *string    `db:"gender" validate:"omitempty,oneof=male female non-binary other prefer-not-to-say"`

	// Location & Localization
	CountryID *uuidv7.UUID `db:"country_id" validate:"omitempty"`
	City      *string      `db:"city" validate:"omitempty,max=100"`
	Timezone  string       `db:"timezone" validate:"required,max=50"`
	Locale    string       `db:"locale" validate:"required,max=10"`

	// Visual Identity
	AvatarURL *string `db:"avatar_url" validate:"omitempty,url,max=500"`
	CoverURL  *string `db:"cover_url" validate:"omitempty,url,max=500"`

	// Social & Web
	SocialLinksJSON []byte      `db:"social_links" validate:"omitempty"`
	SocialLinks     SocialLinks `db:"-" validate:"omitempty"`
	WebsiteURL      *string     `db:"website_url" validate:"omitempty,url,max=500"`
	Company         *string     `db:"company" validate:"omitempty,max=100"`
	JobTitle        *string     `db:"job_title" validate:"omitempty,max=100"`

	// Privacy & Verification
	IsPublic     bool `db:"is_public" validate:"-"`
	IsVerified   bool `db:"is_verified" validate:"-"`
	ShowEmail    bool `db:"show_email" validate:"-"`
	ShowLocation bool `db:"show_location" validate:"-"`
	ShowBirthday bool `db:"show_birthday" validate:"-"`

	// Preferences
	PreferencesJSON []byte      `db:"preferences" validate:"omitempty"`
	Preferences     Preferences `db:"-" validate:"omitempty"`

	// Statistics
	ProfileViewsCount int `db:"profile_views_count" validate:"min=0"`
	FollowersCount    int `db:"followers_count" validate:"min=0"`
	FollowingCount    int `db:"following_count" validate:"min=0"`

	// Moderation
	IsBanned  bool         `db:"is_banned" validate:"-"`
	BanReason *string      `db:"ban_reason" validate:"omitempty"`
	BannedAt  *time.Time   `db:"banned_at" validate:"omitempty"`
	BannedBy  *uuidv7.UUID `db:"banned_by" validate:"omitempty"`

	// Timestamps
	CreatedAt  time.Time  `db:"created_at" validate:"required"`
	UpdatedAt  time.Time  `db:"updated_at" validate:"required"`
	LastSeenAt *time.Time `db:"last_seen_at" validate:"omitempty"`
}

// Validate validates user profile data
func (p *UserProfile) Validate() error {
	// UserID is required
	if p.UserID == (uuidv7.UUID{}) {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	// Nickname validation
	if p.Nickname != nil {
		if len(*p.Nickname) < 3 {
			return fmt.Errorf("%w: nickname must be at least 3 characters", ErrInvalidInput)
		}
		if len(*p.Nickname) > 50 {
			return fmt.Errorf("%w: nickname must not exceed 50 characters", ErrInvalidInput)
		}
	}

	// Bio length validation
	if p.Bio != nil && len(*p.Bio) > 1000 {
		return fmt.Errorf("%w: bio must not exceed 1000 characters", ErrInvalidInput)
	}

	// Gender validation
	if p.Gender != nil {
		gender := Gender(*p.Gender)
		if !gender.IsValid() {
			return fmt.Errorf("%w: invalid gender value", ErrInvalidInput)
		}
	}

	// Timezone is required
	if p.Timezone == "" {
		return fmt.Errorf("%w: timezone is required", ErrInvalidInput)
	}

	// Locale is required
	if p.Locale == "" {
		return fmt.Errorf("%w: locale is required", ErrInvalidInput)
	}

	// Statistics must be non-negative
	if p.ProfileViewsCount < 0 || p.FollowersCount < 0 || p.FollowingCount < 0 {
		return fmt.Errorf("%w: statistics counts cannot be negative", ErrInvalidInput)
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
