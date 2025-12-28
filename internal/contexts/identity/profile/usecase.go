package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines the business logic interface for Profile operations
type IUseCase interface {
	// CreateProfile creates a new profile for a user
	CreateProfile(ctx context.Context, userID uuidv7.UUID, displayName string) (*Profile, error)

	// GetProfile retrieves a profile by ID
	GetProfile(ctx context.Context, profileID uuidv7.UUID) (*Profile, error)

	// GetProfileByUserID retrieves a profile by user ID
	GetProfileByUserID(ctx context.Context, userID uuidv7.UUID) (*Profile, error)

	// UpdateDisplayName updates profile display name
	UpdateDisplayName(ctx context.Context, profileID uuidv7.UUID, displayName string) error

	// UpdateBio updates profile bio
	UpdateBio(ctx context.Context, profileID uuidv7.UUID, bio string) error

	// UpdateAvatar updates profile avatar
	UpdateAvatar(ctx context.Context, profileID uuidv7.UUID, avatarURL string) error

	// UpdatePersonalInfo updates personal information
	UpdatePersonalInfo(ctx context.Context, profileID uuidv7.UUID, firstName, lastName, middleName string) error

	// UpdateGender updates gender
	UpdateGender(ctx context.Context, profileID uuidv7.UUID, gender Gender) error

	// UpdateDateOfBirth updates date of birth
	UpdateDateOfBirth(ctx context.Context, profileID uuidv7.UUID, dateOfBirth *time.Time) error

	// UpdateLocalization updates timezone, language, and country
	UpdateLocalization(ctx context.Context, profileID uuidv7.UUID, timezone, language, country string) error

	// UpdateSocialLinks updates social media links
	UpdateSocialLinks(ctx context.Context, profileID uuidv7.UUID, website, linkedin, twitter, github, facebook, instagram string) error

	// SetPublic sets profile visibility to public
	SetPublic(ctx context.Context, profileID uuidv7.UUID) error

	// SetPrivate sets profile visibility to private
	SetPrivate(ctx context.Context, profileID uuidv7.UUID) error

	// Activate activates the profile
	Activate(ctx context.Context, profileID uuidv7.UUID) error

	// Deactivate deactivates the profile
	Deactivate(ctx context.Context, profileID uuidv7.UUID) error

	// DeleteProfile deletes a profile
	DeleteProfile(ctx context.Context, profileID uuidv7.UUID) error

	// ListPublicProfiles retrieves public profiles
	ListPublicProfiles(ctx context.Context, limit, offset int) ([]*Profile, error)
}

// UseCase implements IUseCase
type UseCase struct {
	repo IRepository
}

// NewUseCase creates a new UseCase
func NewUseCase(repo IRepository) IUseCase {
	return &UseCase{
		repo: repo,
	}
}

// CreateProfile creates a new profile for a user
func (uc *UseCase) CreateProfile(ctx context.Context, userID uuidv7.UUID, displayName string) (*Profile, error) {
	// Check if profile already exists for user
	exists, err := uc.repo.ExistsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("profile already exists for user")
	}

	profile, err := NewProfile(userID, displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	if err := profile.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := uc.repo.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

// GetProfile retrieves a profile by ID
func (uc *UseCase) GetProfile(ctx context.Context, profileID uuidv7.UUID) (*Profile, error) {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

// GetProfileByUserID retrieves a profile by user ID
func (uc *UseCase) GetProfileByUserID(ctx context.Context, userID uuidv7.UUID) (*Profile, error) {
	profile, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

// UpdateDisplayName updates profile display name
func (uc *UseCase) UpdateDisplayName(ctx context.Context, profileID uuidv7.UUID, displayName string) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateDisplayName(displayName); err != nil {
		return fmt.Errorf("failed to update display name: %w", err)
	}

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateBio updates profile bio
func (uc *UseCase) UpdateBio(ctx context.Context, profileID uuidv7.UUID, bio string) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateBio(bio); err != nil {
		return fmt.Errorf("failed to update bio: %w", err)
	}

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateAvatar updates profile avatar
func (uc *UseCase) UpdateAvatar(ctx context.Context, profileID uuidv7.UUID, avatarURL string) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateAvatar(avatarURL); err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdatePersonalInfo updates personal information
func (uc *UseCase) UpdatePersonalInfo(ctx context.Context, profileID uuidv7.UUID, firstName, lastName, middleName string) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdatePersonalInfo(firstName, lastName, middleName); err != nil {
		return fmt.Errorf("failed to update personal info: %w", err)
	}

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateGender updates gender
func (uc *UseCase) UpdateGender(ctx context.Context, profileID uuidv7.UUID, gender Gender) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateGender(gender); err != nil {
		return fmt.Errorf("failed to update gender: %w", err)
	}

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateDateOfBirth updates date of birth
func (uc *UseCase) UpdateDateOfBirth(ctx context.Context, profileID uuidv7.UUID, dateOfBirth *time.Time) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateDateOfBirth(dateOfBirth); err != nil {
		return fmt.Errorf("failed to update date of birth: %w", err)
	}

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateLocalization updates timezone, language, and country
func (uc *UseCase) UpdateLocalization(ctx context.Context, profileID uuidv7.UUID, timezone, language, country string) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateLocalization(timezone, language, country); err != nil {
		return fmt.Errorf("failed to update localization: %w", err)
	}

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateSocialLinks updates social media links
func (uc *UseCase) UpdateSocialLinks(ctx context.Context, profileID uuidv7.UUID, website, linkedin, twitter, github, facebook, instagram string) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateSocialLinks(website, linkedin, twitter, github, facebook, instagram); err != nil {
		return fmt.Errorf("failed to update social links: %w", err)
	}

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// SetPublic sets profile visibility to public
func (uc *UseCase) SetPublic(ctx context.Context, profileID uuidv7.UUID) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile.SetPublic()

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// SetPrivate sets profile visibility to private
func (uc *UseCase) SetPrivate(ctx context.Context, profileID uuidv7.UUID) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile.SetPrivate()

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// Activate activates the profile
func (uc *UseCase) Activate(ctx context.Context, profileID uuidv7.UUID) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile.Activate()

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// Deactivate deactivates the profile
func (uc *UseCase) Deactivate(ctx context.Context, profileID uuidv7.UUID) error {
	profile, err := uc.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile.Deactivate()

	if err := uc.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// DeleteProfile deletes a profile
func (uc *UseCase) DeleteProfile(ctx context.Context, profileID uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, profileID); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// ListPublicProfiles retrieves public profiles
func (uc *UseCase) ListPublicProfiles(ctx context.Context, limit, offset int) ([]*Profile, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	profiles, err := uc.repo.ListPublicProfiles(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list public profiles: %w", err)
	}

	return profiles, nil
}
