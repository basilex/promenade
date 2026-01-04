package profile

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Common errors
var (
	ErrNotFound      = errors.New("profile not found")
	ErrInvalidGender = errors.New("invalid gender value")
)

// Gender defines user's gender
type Gender string

const (
	GenderMale       Gender = "male"
	GenderFemale     Gender = "female"
	GenderOther      Gender = "other"
	GenderNotSpecify Gender = "not_specified"
)

// Profile represents user's profile information
// This is an Aggregate Root in the Identity bounded context
type Profile struct {
	aggregate.BaseAggregate

	// Identity
	UserID uuidv7.UUID

	// Display Information
	DisplayName string // User's display name (e.g., "John Doe")
	Bio         string // Short biography
	AvatarURL   string // URL to profile picture

	// Personal Information
	FirstName   string
	LastName    string
	MiddleName  string
	Gender      Gender
	DateOfBirth *time.Time

	// Localization
	Timezone string // IANA timezone (e.g., "Europe/Kyiv")
	Language string // ISO 639-1 code (e.g., "uk", "en")
	Country  string // ISO 3166-1 alpha-2 code (e.g., "UA")

	// Social Links
	Website   string
	LinkedIn  string
	Twitter   string
	GitHub    string
	Facebook  string
	Instagram string

	// Privacy & Status
	IsPublic bool // Profile visibility
	IsActive bool // Profile status
}

// NewProfile creates a new profile
func NewProfile(userID uuidv7.UUID, displayName string) (*Profile, error) {
	if displayName == "" {
		return nil, fmt.Errorf("display name is required")
	}
	if len(displayName) < 2 {
		return nil, fmt.Errorf("display name must be at least 2 characters")
	}
	if len(displayName) > 100 {
		return nil, fmt.Errorf("display name must not exceed 100 characters")
	}

	displayName = strings.TrimSpace(displayName)

	profile := &Profile{
		BaseAggregate: aggregate.NewBaseAggregate(),
		UserID:        userID,
		DisplayName:   displayName,
		Gender:        GenderNotSpecify,
		IsPublic:      true, // Default to public
		IsActive:      true, // Default to active
	}

	return profile, nil
}

// UpdateDisplayName updates the display name
func (p *Profile) UpdateDisplayName(displayName string) error {
	displayName = strings.TrimSpace(displayName)

	if displayName == "" {
		return fmt.Errorf("display name is required")
	}
	if len(displayName) < 2 {
		return fmt.Errorf("display name must be at least 2 characters")
	}
	if len(displayName) > 100 {
		return fmt.Errorf("display name must not exceed 100 characters")
	}

	p.DisplayName = displayName
	p.Touch()
	return nil
}

// UpdateBio updates the biography
func (p *Profile) UpdateBio(bio string) error {
	bio = strings.TrimSpace(bio)

	if len(bio) > 500 {
		return fmt.Errorf("bio must not exceed 500 characters")
	}

	p.Bio = bio
	p.Touch()
	return nil
}

// UpdateAvatar updates the avatar URL
func (p *Profile) UpdateAvatar(avatarURL string) error {
	avatarURL = strings.TrimSpace(avatarURL)

	if avatarURL != "" && !strings.HasPrefix(avatarURL, "http://") && !strings.HasPrefix(avatarURL, "https://") {
		return fmt.Errorf("avatar URL must be a valid HTTP(S) URL")
	}

	p.AvatarURL = avatarURL
	p.Touch()
	return nil
}

// UpdatePersonalInfo updates personal information
func (p *Profile) UpdatePersonalInfo(firstName, lastName, middleName string) error {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	middleName = strings.TrimSpace(middleName)

	if firstName != "" && len(firstName) > 50 {
		return fmt.Errorf("first name must not exceed 50 characters")
	}
	if lastName != "" && len(lastName) > 50 {
		return fmt.Errorf("last name must not exceed 50 characters")
	}
	if middleName != "" && len(middleName) > 50 {
		return fmt.Errorf("middle name must not exceed 50 characters")
	}

	p.FirstName = firstName
	p.LastName = lastName
	p.MiddleName = middleName
	p.Touch()
	return nil
}

// UpdateGender updates the gender
func (p *Profile) UpdateGender(gender Gender) error {
	if !isValidGender(gender) {
		return ErrInvalidGender
	}

	p.Gender = gender
	p.Touch()
	return nil
}

// UpdateDateOfBirth updates date of birth
func (p *Profile) UpdateDateOfBirth(dateOfBirth *time.Time) error {
	if dateOfBirth != nil {
		// Validate age (must be at least 13 years old)
		minAge := time.Now().AddDate(-13, 0, 0)
		if dateOfBirth.After(minAge) {
			return fmt.Errorf("user must be at least 13 years old")
		}

		// Check not too old (reasonable limit: 120 years)
		maxAge := time.Now().AddDate(-120, 0, 0)
		if dateOfBirth.Before(maxAge) {
			return fmt.Errorf("invalid date of birth")
		}
	}

	p.DateOfBirth = dateOfBirth
	p.Touch()
	return nil
}

// UpdateLocalization updates timezone, language, and country
func (p *Profile) UpdateLocalization(timezone, language, country string) error {
	timezone = strings.TrimSpace(timezone)
	language = strings.TrimSpace(language)
	country = strings.TrimSpace(country)

	// Validate language (ISO 639-1: 2 letters)
	if language != "" && len(language) != 2 {
		return fmt.Errorf("language must be ISO 639-1 code (2 letters)")
	}

	// Validate country (ISO 3166-1: 2 letters)
	if country != "" && len(country) != 2 {
		return fmt.Errorf("country must be ISO 3166-1 alpha-2 code (2 letters)")
	}

	p.Timezone = timezone
	p.Language = strings.ToLower(language)
	p.Country = strings.ToUpper(country)
	p.Touch()
	return nil
}

// UpdateSocialLinks updates social media links
func (p *Profile) UpdateSocialLinks(website, linkedin, twitter, github, facebook, instagram string) error {
	// Validate URLs
	socialLinks := map[string]string{
		"website":   website,
		"linkedin":  linkedin,
		"twitter":   twitter,
		"github":    github,
		"facebook":  facebook,
		"instagram": instagram,
	}

	for name, url := range socialLinks {
		url = strings.TrimSpace(url)
		if url != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return fmt.Errorf("%s URL must be a valid HTTP(S) URL", name)
		}
	}

	p.Website = strings.TrimSpace(website)
	p.LinkedIn = strings.TrimSpace(linkedin)
	p.Twitter = strings.TrimSpace(twitter)
	p.GitHub = strings.TrimSpace(github)
	p.Facebook = strings.TrimSpace(facebook)
	p.Instagram = strings.TrimSpace(instagram)
	p.Touch()
	return nil
}

// SetPublic sets profile visibility to public
func (p *Profile) SetPublic() {
	p.IsPublic = true
	p.Touch()
}

// SetPrivate sets profile visibility to private
func (p *Profile) SetPrivate() {
	p.IsPublic = false
	p.Touch()
}

// Activate activates the profile
func (p *Profile) Activate() {
	p.IsActive = true
	p.Touch()
}

// Deactivate deactivates the profile
func (p *Profile) Deactivate() {
	p.IsActive = false
	p.Touch()
}

// GetFullName returns the full name (FirstName MiddleName LastName)
func (p *Profile) GetFullName() string {
	parts := []string{}

	if p.FirstName != "" {
		parts = append(parts, p.FirstName)
	}
	if p.MiddleName != "" {
		parts = append(parts, p.MiddleName)
	}
	if p.LastName != "" {
		parts = append(parts, p.LastName)
	}

	return strings.Join(parts, " ")
}

// Validate performs validation on the profile
func (p *Profile) Validate() error {
	if p.ID == uuidv7.Nil {
		return fmt.Errorf("profile ID is required")
	}
	if p.UserID == uuidv7.Nil {
		return fmt.Errorf("user ID is required")
	}
	if p.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}
	if !isValidGender(p.Gender) {
		return ErrInvalidGender
	}

	return nil
}

// isValidGender checks if gender value is valid
func isValidGender(gender Gender) bool {
	switch gender {
	case GenderMale, GenderFemale, GenderOther, GenderNotSpecify:
		return true
	default:
		return false
	}
}
