package dto

import (
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateProfileRequest represents the request to create a user profile
type CreateProfileRequest struct {
	FirstName    *string           `json:"first_name,omitempty"`
	LastName     *string           `json:"last_name,omitempty"`
	MiddleName   *string           `json:"middle_name,omitempty"`
	DisplayName  *string           `json:"display_name" binding:"required"`
	Nickname     *string           `json:"nickname,omitempty" binding:"omitempty,min=3,max=50"`
	Bio          *string           `json:"bio,omitempty" binding:"omitempty,max=1000"`
	DateOfBirth  *time.Time        `json:"date_of_birth,omitempty"`
	Gender       *string           `json:"gender,omitempty"`
	CountryID    *uuidv7.UUID        `json:"country_id,omitempty"`
	City         *string           `json:"city,omitempty"`
	Timezone     string            `json:"timezone,omitempty"`
	Locale       string            `json:"locale,omitempty"`
	AvatarURL    *string           `json:"avatar_url,omitempty"`
	CoverURL     *string           `json:"cover_url,omitempty"`
	SocialLinks  map[string]string `json:"social_links,omitempty"`
	WebsiteURL   *string           `json:"website_url,omitempty"`
	Company      *string           `json:"company,omitempty"`
	JobTitle     *string           `json:"job_title,omitempty"`
	Preferences  map[string]any    `json:"preferences,omitempty"`
	IsPublic     bool              `json:"is_public"`
	ShowEmail    bool              `json:"show_email"`
	ShowLocation bool              `json:"show_location"`
	ShowBirthday bool              `json:"show_birthday"`
}

// UpdateProfileRequest represents the request to update a user profile
type UpdateProfileRequest struct {
	FirstName    *string           `json:"first_name,omitempty"`
	LastName     *string           `json:"last_name,omitempty"`
	MiddleName   *string           `json:"middle_name,omitempty"`
	DisplayName  *string           `json:"display_name,omitempty"`
	Nickname     *string           `json:"nickname,omitempty" binding:"omitempty,min=3,max=50"`
	Bio          *string           `json:"bio,omitempty" binding:"omitempty,max=1000"`
	DateOfBirth  *time.Time        `json:"date_of_birth,omitempty"`
	Gender       *string           `json:"gender,omitempty"`
	CountryID    *uuidv7.UUID        `json:"country_id,omitempty"`
	City         *string           `json:"city,omitempty"`
	Timezone     *string           `json:"timezone,omitempty"`
	Locale       *string           `json:"locale,omitempty"`
	AvatarURL    *string           `json:"avatar_url,omitempty"`
	CoverURL     *string           `json:"cover_url,omitempty"`
	SocialLinks  map[string]string `json:"social_links,omitempty"`
	WebsiteURL   *string           `json:"website_url,omitempty"`
	Company      *string           `json:"company,omitempty"`
	JobTitle     *string           `json:"job_title,omitempty"`
	Preferences  map[string]any    `json:"preferences,omitempty"`
	IsPublic     *bool             `json:"is_public,omitempty"`
	ShowEmail    *bool             `json:"show_email,omitempty"`
	ShowLocation *bool             `json:"show_location,omitempty"`
	ShowBirthday *bool             `json:"show_birthday,omitempty"`
}

// BanProfileRequest represents the request to ban a user profile
type BanProfileRequest struct {
	Reason string `json:"reason" binding:"required,min=10,max=500"`
}

// ProfileResponse represents a user profile in API responses
type ProfileResponse struct {
	ID                uuidv7.UUID         `json:"id"`
	UserID            uuidv7.UUID         `json:"user_id"`
	FirstName         *string           `json:"first_name,omitempty"`
	LastName          *string           `json:"last_name,omitempty"`
	MiddleName        *string           `json:"middle_name,omitempty"`
	DisplayName       *string           `json:"display_name"`
	Nickname          *string           `json:"nickname,omitempty"`
	Bio               *string           `json:"bio,omitempty"`
	DateOfBirth       *time.Time        `json:"date_of_birth,omitempty"`
	Gender            *string           `json:"gender,omitempty"`
	CountryID         *uuidv7.UUID        `json:"country_id,omitempty"`
	City              *string           `json:"city,omitempty"`
	Timezone          string            `json:"timezone"`
	Locale            string            `json:"locale"`
	AvatarURL         *string           `json:"avatar_url,omitempty"`
	CoverURL          *string           `json:"cover_url,omitempty"`
	SocialLinks       map[string]string `json:"social_links,omitempty"`
	WebsiteURL        *string           `json:"website_url,omitempty"`
	Company           *string           `json:"company,omitempty"`
	JobTitle          *string           `json:"job_title,omitempty"`
	Preferences       map[string]any    `json:"preferences,omitempty"`
	IsPublic          bool              `json:"is_public"`
	ShowEmail         bool              `json:"show_email"`
	ShowLocation      bool              `json:"show_location"`
	ShowBirthday      bool              `json:"show_birthday"`
	IsVerified        bool              `json:"is_verified"`
	IsBanned          bool              `json:"is_banned"`
	BanReason         *string           `json:"ban_reason,omitempty"`
	BannedAt          *time.Time        `json:"banned_at,omitempty"`
	BannedBy          *uuidv7.UUID        `json:"banned_by,omitempty"`
	LastSeenAt        *time.Time        `json:"last_seen_at,omitempty"`
	ProfileViewsCount int               `json:"profile_views_count"`
	FollowersCount    int               `json:"followers_count"`
	FollowingCount    int               `json:"following_count"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// ProfileListResponse represents a list of user profiles
type ProfileListResponse struct {
	Profiles []*ProfileResponse `json:"profiles"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}

// ToProfileResponse converts entity.UserProfile to ProfileResponse
func ToProfileResponse(profile *entity.UserProfile) *ProfileResponse {
	return &ProfileResponse{
		ID:                profile.ID,
		UserID:            profile.UserID,
		FirstName:         profile.FirstName,
		LastName:          profile.LastName,
		MiddleName:        profile.MiddleName,
		DisplayName:       profile.DisplayName,
		Nickname:          profile.Nickname,
		Bio:               profile.Bio,
		DateOfBirth:       profile.DateOfBirth,
		Gender:            profile.Gender,
		CountryID:         profile.CountryID,
		City:              profile.City,
		Timezone:          profile.Timezone,
		Locale:            profile.Locale,
		AvatarURL:         profile.AvatarURL,
		CoverURL:          profile.CoverURL,
		SocialLinks:       profile.SocialLinks,
		WebsiteURL:        profile.WebsiteURL,
		Company:           profile.Company,
		JobTitle:          profile.JobTitle,
		Preferences:       profile.Preferences,
		IsPublic:          profile.IsPublic,
		ShowEmail:         profile.ShowEmail,
		ShowLocation:      profile.ShowLocation,
		ShowBirthday:      profile.ShowBirthday,
		IsVerified:        profile.IsVerified,
		IsBanned:          profile.IsBanned,
		BanReason:         profile.BanReason,
		BannedAt:          profile.BannedAt,
		BannedBy:          profile.BannedBy,
		LastSeenAt:        profile.LastSeenAt,
		ProfileViewsCount: profile.ProfileViewsCount,
		FollowersCount:    profile.FollowersCount,
		FollowingCount:    profile.FollowingCount,
		CreatedAt:         profile.CreatedAt,
		UpdatedAt:         profile.UpdatedAt,
	}
}

// ToProfileListResponse converts slice of entity.UserProfile to ProfileListResponse
func ToProfileListResponse(profiles []*entity.UserProfile, limit, offset int) *ProfileListResponse {
	responses := make([]*ProfileResponse, 0, len(profiles))
	for _, profile := range profiles {
		responses = append(responses, ToProfileResponse(profile))
	}

	return &ProfileListResponse{
		Profiles: responses,
		Total:    len(profiles),
		Limit:    limit,
		Offset:   offset,
	}
}

// ToEntity converts CreateProfileRequest to entity.UserProfile
func (r *CreateProfileRequest) ToEntity() *entity.UserProfile {
	return &entity.UserProfile{
		FirstName:    r.FirstName,
		LastName:     r.LastName,
		MiddleName:   r.MiddleName,
		DisplayName:  r.DisplayName,
		Nickname:     r.Nickname,
		Bio:          r.Bio,
		DateOfBirth:  r.DateOfBirth,
		Gender:       r.Gender,
		CountryID:    r.CountryID,
		City:         r.City,
		Timezone:     r.Timezone,
		Locale:       r.Locale,
		AvatarURL:    r.AvatarURL,
		CoverURL:     r.CoverURL,
		SocialLinks:  r.SocialLinks,
		WebsiteURL:   r.WebsiteURL,
		Company:      r.Company,
		JobTitle:     r.JobTitle,
		Preferences:  r.Preferences,
		IsPublic:     r.IsPublic,
		ShowEmail:    r.ShowEmail,
		ShowLocation: r.ShowLocation,
		ShowBirthday: r.ShowBirthday,
	}
}

// ToEntity converts UpdateProfileRequest to entity.UserProfile
func (r *UpdateProfileRequest) ToEntity() *entity.UserProfile {
	profile := &entity.UserProfile{
		FirstName:   r.FirstName,
		LastName:    r.LastName,
		MiddleName:  r.MiddleName,
		DisplayName: r.DisplayName,
		Nickname:    r.Nickname,
		Bio:         r.Bio,
		DateOfBirth: r.DateOfBirth,
		Gender:      r.Gender,
		CountryID:   r.CountryID,
		City:        r.City,
		AvatarURL:   r.AvatarURL,
		CoverURL:    r.CoverURL,
		SocialLinks: r.SocialLinks,
		WebsiteURL:  r.WebsiteURL,
		Company:     r.Company,
		JobTitle:    r.JobTitle,
		Preferences: r.Preferences,
	}

	if r.Timezone != nil {
		profile.Timezone = *r.Timezone
	}
	if r.Locale != nil {
		profile.Locale = *r.Locale
	}
	if r.IsPublic != nil {
		profile.IsPublic = *r.IsPublic
	}
	if r.ShowEmail != nil {
		profile.ShowEmail = *r.ShowEmail
	}
	if r.ShowLocation != nil {
		profile.ShowLocation = *r.ShowLocation
	}
	if r.ShowBirthday != nil {
		profile.ShowBirthday = *r.ShowBirthday
	}

	return profile
}
