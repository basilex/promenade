package usecase

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
	"github.com/basilex/promenade/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserProfileUseCase_CreateProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("successful creation", func(t *testing.T) {
		displayName := "John Doe"
		nickname := "johndoe"
		profile := &entity.UserProfile{
			DisplayName: &displayName,
			Nickname:    &nickname,
			Timezone:    "UTC",
			Locale:      "en",
		}

		// Mock: no existing profile for user
		mockRepo.On("GetByUserID", ctx, userID).Return(nil, entity.ErrNotFound).Once()
		// Mock: nickname is available
		mockRepo.On("GetByNickname", ctx, nickname).Return(nil, entity.ErrNotFound).Once()
		// Mock: successful creation
		mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.UserProfile")).Return(nil).Once()

		result, err := uc.CreateProfile(ctx, userID, profile)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.NotEqual(t, uuidv7.Nil, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile already exists", func(t *testing.T) {
		existingProfile := helpers.UserProfileFixture(userID, "existing")
		displayName := "New Profile"
		profile := &entity.UserProfile{
			DisplayName: &displayName,
			Timezone:    "UTC",
			Locale:      "en",
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(existingProfile, nil).Once()

		result, err := uc.CreateProfile(ctx, userID, profile)

		assert.ErrorIs(t, err, ErrProfileAlreadyExists)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nickname already taken", func(t *testing.T) {
		nickname := "taken"
		displayName := "John Doe"
		profile := &entity.UserProfile{
			DisplayName: &displayName,
			Nickname:    &nickname,
			Timezone:    "UTC",
			Locale:      "en",
		}

		existingProfile := helpers.UserProfileFixture(uuidv7.New(), nickname)

		mockRepo.On("GetByUserID", ctx, userID).Return(nil, entity.ErrNotFound).Once()
		mockRepo.On("GetByNickname", ctx, nickname).Return(existingProfile, nil).Once()

		result, err := uc.CreateProfile(ctx, userID, profile)

		assert.ErrorIs(t, err, ErrNicknameAlreadyTaken)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid profile data", func(t *testing.T) {
		// DisplayName is nil, which will fail validation
		profile := &entity.UserProfile{
			Timezone: "UTC",
			Locale:   "en",
		}

		mockRepo.On("GetByUserID", ctx, userID).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.CreateProfile(ctx, userID, profile)

		assert.ErrorIs(t, err, ErrInvalidProfileData)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserProfileUseCase_GetProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()

	profileID := uuidv7.New()
	userID := uuidv7.New()
	viewerID := uuidv7.New()

	t.Run("get public profile", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testuser")
		profile.ID = profileID
		profile.IsPublic = true

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()

		result, err := uc.GetProfile(ctx, profileID, &viewerID)

		assert.NoError(t, err)
		assert.Equal(t, profileID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("get own private profile", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testuser")
		profile.ID = profileID
		profile.IsPublic = false

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()

		result, err := uc.GetProfile(ctx, profileID, &userID) // Owner viewing

		assert.NoError(t, err)
		assert.Equal(t, profileID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access to private profile", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testuser")
		profile.ID = profileID
		profile.IsPublic = false

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()

		result, err := uc.GetProfile(ctx, profileID, &viewerID) // Non-owner viewing

		assert.ErrorIs(t, err, ErrUnauthorizedProfileAccess)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, profileID).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.GetProfile(ctx, profileID, &viewerID)

		assert.ErrorIs(t, err, ErrProfileNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("banned profile", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testuser")
		profile.ID = profileID
		profile.IsBanned = true

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()

		result, err := uc.GetProfile(ctx, profileID, &viewerID)

		assert.ErrorIs(t, err, ErrUnauthorizedProfileAccess)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserProfileUseCase_UpdateProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()

	profileID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("successful update", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "oldnick")
		profile.ID = profileID

		newBio := "Updated bio"
		updates := &entity.UserProfile{
			Bio: &newBio,
		}

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserProfile")).Return(nil).Once()

		result, err := uc.UpdateProfile(ctx, profileID, userID, updates)

		assert.NoError(t, err)
		assert.Equal(t, newBio, *result.Bio)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update nickname to available one", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "oldnick")
		profile.ID = profileID

		newNickname := "newnick"
		updates := &entity.UserProfile{
			Nickname: &newNickname,
		}

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()
		mockRepo.On("GetByNickname", ctx, newNickname).Return(nil, entity.ErrNotFound).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.UserProfile")).Return(nil).Once()

		result, err := uc.UpdateProfile(ctx, profileID, userID, updates)

		assert.NoError(t, err)
		assert.Equal(t, newNickname, *result.Nickname)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update nickname to taken one", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "oldnick")
		profile.ID = profileID

		takenNickname := "takennick"
		updates := &entity.UserProfile{
			Nickname: &takenNickname,
		}

		existingProfile := helpers.UserProfileFixture(otherUserID, takenNickname)

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()
		mockRepo.On("GetByNickname", ctx, takenNickname).Return(existingProfile, nil).Once()

		result, err := uc.UpdateProfile(ctx, profileID, userID, updates)

		assert.ErrorIs(t, err, ErrNicknameAlreadyTaken)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized update", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testnick")
		profile.ID = profileID

		updates := &entity.UserProfile{}

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()

		result, err := uc.UpdateProfile(ctx, profileID, otherUserID, updates) // Different user

		assert.ErrorIs(t, err, ErrUnauthorizedProfileAccess)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		updates := &entity.UserProfile{}

		mockRepo.On("GetByID", ctx, profileID).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.UpdateProfile(ctx, profileID, userID, updates)

		assert.ErrorIs(t, err, ErrProfileNotFound)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserProfileUseCase_DeleteProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()

	profileID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("successful deletion", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testuser")
		profile.ID = profileID

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()
		mockRepo.On("Delete", ctx, profileID).Return(nil).Once()

		err := uc.DeleteProfile(ctx, profileID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized deletion", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testuser")
		profile.ID = profileID

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()

		err := uc.DeleteProfile(ctx, profileID, otherUserID)

		assert.ErrorIs(t, err, ErrUnauthorizedProfileAccess)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		mockRepo.On("GetByID", ctx, profileID).Return(nil, entity.ErrNotFound).Once()

		err := uc.DeleteProfile(ctx, profileID, userID)

		assert.ErrorIs(t, err, ErrProfileNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserProfileUseCase_IncrementViews(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()

	profileID := uuidv7.New()
	userID := uuidv7.New()
	viewerID := uuidv7.New()

	t.Run("increment views for other user", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testuser")
		profile.ID = profileID

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()
		mockRepo.On("IncrementProfileViews", ctx, profileID).Return(nil).Once()

		err := uc.IncrementViews(ctx, profileID, &viewerID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("do not increment views for own profile", func(t *testing.T) {
		profile := helpers.UserProfileFixture(userID, "testuser")
		profile.ID = profileID

		mockRepo.On("GetByID", ctx, profileID).Return(profile, nil).Once()
		// Should NOT call IncrementProfileViews

		err := uc.IncrementViews(ctx, profileID, &userID) // Owner viewing

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserProfileUseCase_BanProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()

	profileID := uuidv7.New()
	adminID := uuidv7.New()
	reason := "Violation of terms"

	t.Run("successful ban", func(t *testing.T) {
		mockRepo.On("Ban", ctx, profileID, reason, adminID).Return(nil).Once()

		err := uc.BanProfile(ctx, profileID, reason, adminID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserProfileUseCase_UnbanProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()

	profileID := uuidv7.New()

	t.Run("successful unban", func(t *testing.T) {
		mockRepo.On("Unban", ctx, profileID).Return(nil).Once()

		err := uc.UnbanProfile(ctx, profileID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserProfileUseCase_VerifyProfile(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()

	profileID := uuidv7.New()

	t.Run("successful verification", func(t *testing.T) {
		mockRepo.On("SetVerified", ctx, profileID, true).Return(nil).Once()

		err := uc.VerifyProfile(ctx, profileID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserProfileUseCase_SearchProfiles(t *testing.T) {
	mockRepo := new(mocks.MockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)
	ctx := context.Background()

	t.Run("successful search", func(t *testing.T) {
		query := "john"
		profiles := []*entity.UserProfile{
			helpers.UserProfileFixture(uuidv7.New(), "johndoe"),
			helpers.UserProfileFixture(uuidv7.New(), "johnny"),
		}

		mockRepo.On("Search", ctx, query, 10, 0).Return(profiles, nil).Once()

		results, err := uc.SearchProfiles(ctx, query, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, results, 2)
		mockRepo.AssertExpectations(t)
	})
}
