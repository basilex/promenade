package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/profile/aggregate"
	"github.com/basilex/promenade/internal/contexts/identity/profile/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IProfileUseCase defines the business logic interface for Profile operations
type IProfileUseCase interface {
	// CreateProfile creates a new profile for a user
	CreateProfile(ctx context.Context, userID uuidv7.UUID, displayName string) (*aggregate.Profile, error)

	// GetProfile retrieves a profile by ID
	GetProfile(ctx context.Context, profileID uuidv7.UUID) (*aggregate.Profile, error)

	// GetProfileByUserID retrieves a profile by user ID
	GetProfileByUserID(ctx context.Context, userID uuidv7.UUID) (*aggregate.Profile, error)

	// UpdateDisplayName updates profile display name
	UpdateDisplayName(ctx context.Context, profileID uuidv7.UUID, displayName string) error

	// UpdateBio updates profile bio
	UpdateBio(ctx context.Context, profileID uuidv7.UUID, bio string) error

	// UpdateAvatar updates profile avatar
	UpdateAvatar(ctx context.Context, profileID uuidv7.UUID, avatarURL string) error

	// UpdatePersonalInfo updates personal information
	UpdatePersonalInfo(ctx context.Context, profileID uuidv7.UUID, firstName, lastName, middleName string) error

	// UpdateGender updates gender
	UpdateGender(ctx context.Context, profileID uuidv7.UUID, gender aggregate.Gender) error

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
	ListPublicProfiles(ctx context.Context, limit, offset int) ([]*aggregate.Profile, error)
}

// ProfileUseCase implements IProfileUseCase
type ProfileUseCase struct {
	repo repository.IProfileRepository
}

// NewProfileUseCase creates a new ProfileUseCase
func NewProfileUseCase(repo repository.IProfileRepository) IProfileUseCase {
	return &ProfileUseCase{
		repo: repo,
	}
}

// CreateProfile creates a new profile for a user
func (u *ProfileUseCase) CreateProfile(ctx context.Context, userID uuidv7.UUID, displayName string) (*aggregate.Profile, error) {
	// Check if profile already exists for user
	exists, err := u.repo.ExistsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("profile already exists for user")
	}

	profile, err := aggregate.NewProfile(userID, displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	if err := profile.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if err := u.repo.Create(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

// GetProfile retrieves a profile by ID
func (u *ProfileUseCase) GetProfile(ctx context.Context, profileID uuidv7.UUID) (*aggregate.Profile, error) {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

// GetProfileByUserID retrieves a profile by user ID
func (u *ProfileUseCase) GetProfileByUserID(ctx context.Context, userID uuidv7.UUID) (*aggregate.Profile, error) {
	profile, err := u.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return profile, nil
}

// UpdateDisplayName updates profile display name
func (u *ProfileUseCase) UpdateDisplayName(ctx context.Context, profileID uuidv7.UUID, displayName string) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateDisplayName(displayName); err != nil {
		return fmt.Errorf("failed to update display name: %w", err)
	}

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateBio updates profile bio
func (u *ProfileUseCase) UpdateBio(ctx context.Context, profileID uuidv7.UUID, bio string) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateBio(bio); err != nil {
		return fmt.Errorf("failed to update bio: %w", err)
	}

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateAvatar updates profile avatar
func (u *ProfileUseCase) UpdateAvatar(ctx context.Context, profileID uuidv7.UUID, avatarURL string) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateAvatar(avatarURL); err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdatePersonalInfo updates personal information
func (u *ProfileUseCase) UpdatePersonalInfo(ctx context.Context, profileID uuidv7.UUID, firstName, lastName, middleName string) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdatePersonalInfo(firstName, lastName, middleName); err != nil {
		return fmt.Errorf("failed to update personal info: %w", err)
	}

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateGender updates gender
func (u *ProfileUseCase) UpdateGender(ctx context.Context, profileID uuidv7.UUID, gender aggregate.Gender) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateGender(gender); err != nil {
		return fmt.Errorf("failed to update gender: %w", err)
	}

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateDateOfBirth updates date of birth
func (u *ProfileUseCase) UpdateDateOfBirth(ctx context.Context, profileID uuidv7.UUID, dateOfBirth *time.Time) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateDateOfBirth(dateOfBirth); err != nil {
		return fmt.Errorf("failed to update date of birth: %w", err)
	}

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateLocalization updates timezone, language, and country
func (u *ProfileUseCase) UpdateLocalization(ctx context.Context, profileID uuidv7.UUID, timezone, language, country string) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateLocalization(timezone, language, country); err != nil {
		return fmt.Errorf("failed to update localization: %w", err)
	}

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// UpdateSocialLinks updates social media links
func (u *ProfileUseCase) UpdateSocialLinks(ctx context.Context, profileID uuidv7.UUID, website, linkedin, twitter, github, facebook, instagram string) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	if err := profile.UpdateSocialLinks(website, linkedin, twitter, github, facebook, instagram); err != nil {
		return fmt.Errorf("failed to update social links: %w", err)
	}

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// SetPublic sets profile visibility to public
func (u *ProfileUseCase) SetPublic(ctx context.Context, profileID uuidv7.UUID) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile.SetPublic()

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// SetPrivate sets profile visibility to private
func (u *ProfileUseCase) SetPrivate(ctx context.Context, profileID uuidv7.UUID) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile.SetPrivate()

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// Activate activates the profile
func (u *ProfileUseCase) Activate(ctx context.Context, profileID uuidv7.UUID) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile.Activate()

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// Deactivate deactivates the profile
func (u *ProfileUseCase) Deactivate(ctx context.Context, profileID uuidv7.UUID) error {
	profile, err := u.repo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("failed to get profile: %w", err)
	}

	profile.Deactivate()

	if err := u.repo.Update(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// DeleteProfile deletes a profile
func (u *ProfileUseCase) DeleteProfile(ctx context.Context, profileID uuidv7.UUID) error {
	if err := u.repo.Delete(ctx, profileID); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	return nil
}

// ListPublicProfiles retrieves public profiles
func (u *ProfileUseCase) ListPublicProfiles(ctx context.Context, limit, offset int) ([]*aggregate.Profile, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	profiles, err := u.repo.ListPublicProfiles(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list public profiles: %w", err)
	}

	return profiles, nil
}
