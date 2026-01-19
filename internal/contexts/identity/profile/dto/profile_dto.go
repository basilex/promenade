package dto

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/profile/aggregate"
)

// CreateProfileRequest represents the request to create a new profile
type CreateProfileRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
}

// UpdateDisplayNameRequest represents the request to update display name
type UpdateDisplayNameRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
}

// UpdateBioRequest represents the request to update bio
type UpdateBioRequest struct {
	Bio string `json:"bio" binding:"max=500"`
}

// UpdateAvatarRequest represents the request to update avatar
type UpdateAvatarRequest struct {
	AvatarURL string `json:"avatar_url" binding:"omitempty,url"`
}

// UpdatePersonalInfoRequest represents the request to update personal information
type UpdatePersonalInfoRequest struct {
	FirstName  string `json:"first_name" binding:"max=50"`
	LastName   string `json:"last_name" binding:"max=50"`
	MiddleName string `json:"middle_name" binding:"max=50"`
}

// UpdateGenderRequest represents the request to update gender
type UpdateGenderRequest struct {
	Gender string `json:"gender" binding:"required,oneof=male female other not_specified"`
}

// UpdateDateOfBirthRequest represents the request to update date of birth
type UpdateDateOfBirthRequest struct {
	DateOfBirth *time.Time `json:"date_of_birth"`
}

// UpdateLocalizationRequest represents the request to update localization
type UpdateLocalizationRequest struct {
	Timezone string `json:"timezone" binding:"omitempty"`
	Language string `json:"language" binding:"omitempty,len=2"`
	Country  string `json:"country" binding:"omitempty,len=2"`
}

// UpdateSocialLinksRequest represents the request to update social links
type UpdateSocialLinksRequest struct {
	Website   string `json:"website" binding:"omitempty,url"`
	LinkedIn  string `json:"linkedin" binding:"omitempty,url"`
	Twitter   string `json:"twitter" binding:"omitempty,url"`
	GitHub    string `json:"github" binding:"omitempty,url"`
	Facebook  string `json:"facebook" binding:"omitempty,url"`
	Instagram string `json:"instagram" binding:"omitempty,url"`
}

// ProfileResponse represents a profile in API responses
type ProfileResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	DisplayName string     `json:"display_name"`
	Bio         string     `json:"bio,omitempty"`
	AvatarURL   string     `json:"avatar_url,omitempty"`
	FirstName   string     `json:"first_name,omitempty"`
	LastName    string     `json:"last_name,omitempty"`
	MiddleName  string     `json:"middle_name,omitempty"`
	FullName    string     `json:"full_name,omitempty"`
	Gender      string     `json:"gender"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Timezone    string     `json:"timezone,omitempty"`
	Language    string     `json:"language,omitempty"`
	Country     string     `json:"country,omitempty"`
	Website     string     `json:"website,omitempty"`
	LinkedIn    string     `json:"linkedin,omitempty"`
	Twitter     string     `json:"twitter,omitempty"`
	GitHub      string     `json:"github,omitempty"`
	Facebook    string     `json:"facebook,omitempty"`
	Instagram   string     `json:"instagram,omitempty"`
	IsPublic    bool       `json:"is_public"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToProfileResponse converts a Profile entity to ProfileResponse DTO
func ToProfileResponse(p *aggregate.Profile) ProfileResponse {
	return ProfileResponse{
		ID:          p.ID.String(),
		UserID:      p.UserID.String(),
		DisplayName: p.DisplayName,
		Bio:         p.Bio,
		AvatarURL:   p.AvatarURL,
		FirstName:   p.FirstName,
		LastName:    p.LastName,
		MiddleName:  p.MiddleName,
		FullName:    p.GetFullName(),
		Gender:      string(p.Gender),
		DateOfBirth: p.DateOfBirth,
		Timezone:    p.Timezone,
		Language:    p.Language,
		Country:     p.Country,
		Website:     p.Website,
		LinkedIn:    p.LinkedIn,
		Twitter:     p.Twitter,
		GitHub:      p.GitHub,
		Facebook:    p.Facebook,
		Instagram:   p.Instagram,
		IsPublic:    p.IsPublic,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ToProfileResponseList converts a list of Profile entities to ProfileResponse DTOs
func ToProfileResponseList(profiles []*aggregate.Profile) []ProfileResponse {
	responses := make([]ProfileResponse, len(profiles))
	for i, p := range profiles {
		responses[i] = ToProfileResponse(p)
	}
	return responses
}
