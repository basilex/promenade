package usecase

import (
	"context"
	"errors"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

var (
	ErrProfileAlreadyExists      = errors.New("profile already exists for this user")
	ErrProfileNotFound           = errors.New("profile not found")
	ErrNicknameAlreadyTaken      = errors.New("nickname is already taken")
	ErrUnauthorizedProfileAccess = errors.New("unauthorized to perform this action")
	ErrInvalidProfileData        = errors.New("invalid profile data")
)

type IUserProfileUseCase interface {
	CreateProfile(ctx context.Context, userID uuidv7.UUID, profile *entity.UserProfile) (*entity.UserProfile, error)
	GetProfile(ctx context.Context, profileID uuidv7.UUID, viewerID *uuidv7.UUID) (*entity.UserProfile, error)
	GetProfileByUserID(ctx context.Context, userID uuidv7.UUID, viewerID *uuidv7.UUID) (*entity.UserProfile, error)
	GetProfileByNickname(ctx context.Context, nickname string, viewerID *uuidv7.UUID) (*entity.UserProfile, error)
	UpdateProfile(ctx context.Context, profileID uuidv7.UUID, userID uuidv7.UUID, updates *entity.UserProfile) (*entity.UserProfile, error)
	DeleteProfile(ctx context.Context, profileID uuidv7.UUID, userID uuidv7.UUID) error
	ListProfiles(ctx context.Context, limit, offset int, publicOnly bool) ([]*entity.UserProfile, error)
	SearchProfiles(ctx context.Context, query string, limit, offset int) ([]*entity.UserProfile, error)
	UpdateLastSeen(ctx context.Context, profileID uuidv7.UUID) error
	IncrementViews(ctx context.Context, profileID uuidv7.UUID, viewerID *uuidv7.UUID) error
	BanProfile(ctx context.Context, profileID uuidv7.UUID, reason string, bannedBy uuidv7.UUID) error
	UnbanProfile(ctx context.Context, profileID uuidv7.UUID) error
	VerifyProfile(ctx context.Context, profileID uuidv7.UUID) error
	UnverifyProfile(ctx context.Context, profileID uuidv7.UUID) error
}

type userProfileUseCase struct {
	profileRepo repository.IUserProfileRepository
}

func NewUserProfileUseCase(profileRepo repository.IUserProfileRepository) IUserProfileUseCase {
	return &userProfileUseCase{
		profileRepo: profileRepo,
	}
}

func (uc *userProfileUseCase) CreateProfile(ctx context.Context, userID uuidv7.UUID, profile *entity.UserProfile) (*entity.UserProfile, error) {
	// Check if profile already exists for this user
	existing, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, ErrProfileAlreadyExists
	}
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		return nil, err
	}

	// Check if nickname is already taken
	if profile.Nickname != nil {
		existingNickname, err := uc.profileRepo.GetByNickname(ctx, *profile.Nickname)
		if err == nil && existingNickname != nil {
			return nil, ErrNicknameAlreadyTaken
		}
		if err != nil && !errors.Is(err, entity.ErrNotFound) {
			return nil, err
		}
	}

	// Set profile ID and user ID
	profile.ID = uuidv7.New()
	profile.UserID = userID

	// Validate profile data
	if err := profile.Validate(); err != nil {
		return nil, ErrInvalidProfileData
	}

	// Create profile
	if err := uc.profileRepo.Create(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (uc *userProfileUseCase) GetProfile(ctx context.Context, profileID uuidv7.UUID, viewerID *uuidv7.UUID) (*entity.UserProfile, error) {
	profile, err := uc.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	// Check if viewer can access this profile
	viewerValue := uuidv7.Nil
	if viewerID != nil {
		viewerValue = *viewerID
	}
	if !profile.CanView(viewerValue) {
		return nil, ErrUnauthorizedProfileAccess
	}

	return profile, nil
}

func (uc *userProfileUseCase) GetProfileByUserID(ctx context.Context, userID uuidv7.UUID, viewerID *uuidv7.UUID) (*entity.UserProfile, error) {
	profile, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	// Check if viewer can access this profile
	viewerValue := uuidv7.Nil
	if viewerID != nil {
		viewerValue = *viewerID
	}
	if !profile.CanView(viewerValue) {
		return nil, ErrUnauthorizedProfileAccess
	}

	return profile, nil
}

func (uc *userProfileUseCase) GetProfileByNickname(ctx context.Context, nickname string, viewerID *uuidv7.UUID) (*entity.UserProfile, error) {
	profile, err := uc.profileRepo.GetByNickname(ctx, nickname)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	// Check if viewer can access this profile
	viewerValue := uuidv7.Nil
	if viewerID != nil {
		viewerValue = *viewerID
	}
	if !profile.CanView(viewerValue) {
		return nil, ErrUnauthorizedProfileAccess
	}

	return profile, nil
}

func (uc *userProfileUseCase) UpdateProfile(ctx context.Context, profileID uuidv7.UUID, userID uuidv7.UUID, updates *entity.UserProfile) (*entity.UserProfile, error) {
	// Get existing profile
	profile, err := uc.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	// Check if user owns this profile
	if !profile.IsOwnProfile(userID) {
		return nil, ErrUnauthorizedProfileAccess
	}

	// Check if nickname is being changed and is available
	if updates.Nickname != nil && (profile.Nickname == nil || *updates.Nickname != *profile.Nickname) {
		existingNickname, err := uc.profileRepo.GetByNickname(ctx, *updates.Nickname)
		if err == nil && existingNickname != nil && existingNickname.ID != profileID {
			return nil, ErrNicknameAlreadyTaken
		}
		if err != nil && !errors.Is(err, entity.ErrNotFound) {
			return nil, err
		}
	}

	// Update fields
	if updates.FirstName != nil {
		profile.FirstName = updates.FirstName
	}
	if updates.LastName != nil {
		profile.LastName = updates.LastName
	}
	if updates.MiddleName != nil {
		profile.MiddleName = updates.MiddleName
	}
	if updates.DisplayName != nil {
		profile.DisplayName = updates.DisplayName
	}
	if updates.Nickname != nil {
		profile.Nickname = updates.Nickname
	}
	if updates.Bio != nil {
		profile.Bio = updates.Bio
	}
	if updates.DateOfBirth != nil {
		profile.DateOfBirth = updates.DateOfBirth
	}
	if updates.Gender != nil {
		profile.Gender = updates.Gender
	}
	if updates.CountryID != nil {
		profile.CountryID = updates.CountryID
	}
	if updates.City != nil {
		profile.City = updates.City
	}
	if updates.Timezone != "" {
		profile.Timezone = updates.Timezone
	}
	if updates.Locale != "" {
		profile.Locale = updates.Locale
	}
	if updates.AvatarURL != nil {
		profile.AvatarURL = updates.AvatarURL
	}
	if updates.CoverURL != nil {
		profile.CoverURL = updates.CoverURL
	}
	if updates.SocialLinks != nil {
		profile.SocialLinks = updates.SocialLinks
	}
	if updates.WebsiteURL != nil {
		profile.WebsiteURL = updates.WebsiteURL
	}
	if updates.Company != nil {
		profile.Company = updates.Company
	}
	if updates.JobTitle != nil {
		profile.JobTitle = updates.JobTitle
	}
	if updates.Preferences != nil {
		profile.Preferences = updates.Preferences
	}

	// Update privacy settings (these are boolean, so we check if they're explicitly set)
	profile.IsPublic = updates.IsPublic
	profile.ShowEmail = updates.ShowEmail
	profile.ShowLocation = updates.ShowLocation
	profile.ShowBirthday = updates.ShowBirthday

	// Validate updated profile
	if err := profile.Validate(); err != nil {
		return nil, ErrInvalidProfileData
	}

	// Save updates
	if err := uc.profileRepo.Update(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (uc *userProfileUseCase) DeleteProfile(ctx context.Context, profileID uuidv7.UUID, userID uuidv7.UUID) error {
	// Get profile
	profile, err := uc.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrProfileNotFound
		}
		return err
	}

	// Check if user owns this profile
	if !profile.IsOwnProfile(userID) {
		return ErrUnauthorizedProfileAccess
	}

	return uc.profileRepo.Delete(ctx, profileID)
}

func (uc *userProfileUseCase) ListProfiles(ctx context.Context, limit, offset int, publicOnly bool) ([]*entity.UserProfile, error) {
	var isPublic *bool
	if publicOnly {
		val := true
		isPublic = &val
	}

	return uc.profileRepo.List(ctx, limit, offset, isPublic)
}

func (uc *userProfileUseCase) SearchProfiles(ctx context.Context, query string, limit, offset int) ([]*entity.UserProfile, error) {
	return uc.profileRepo.Search(ctx, query, limit, offset)
}

func (uc *userProfileUseCase) UpdateLastSeen(ctx context.Context, profileID uuidv7.UUID) error {
	return uc.profileRepo.UpdateLastSeen(ctx, profileID)
}

func (uc *userProfileUseCase) IncrementViews(ctx context.Context, profileID uuidv7.UUID, viewerID *uuidv7.UUID) error {
	// Get profile
	profile, err := uc.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return err
	}

	// Don't increment views for own profile
	if viewerID != nil && profile.IsOwnProfile(*viewerID) {
		return nil
	}

	return uc.profileRepo.IncrementProfileViews(ctx, profileID)
}

func (uc *userProfileUseCase) BanProfile(ctx context.Context, profileID uuidv7.UUID, reason string, bannedBy uuidv7.UUID) error {
	return uc.profileRepo.Ban(ctx, profileID, reason, bannedBy)
}

func (uc *userProfileUseCase) UnbanProfile(ctx context.Context, profileID uuidv7.UUID) error {
	return uc.profileRepo.Unban(ctx, profileID)
}

func (uc *userProfileUseCase) VerifyProfile(ctx context.Context, profileID uuidv7.UUID) error {
	return uc.profileRepo.SetVerified(ctx, profileID, true)
}

func (uc *userProfileUseCase) UnverifyProfile(ctx context.Context, profileID uuidv7.UUID) error {
	return uc.profileRepo.SetVerified(ctx, profileID, false)
}
