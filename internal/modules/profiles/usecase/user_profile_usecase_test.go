package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

//  IModule-independent test: imports only module and pkg, no core dependencies

func TestCreateProfile(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	userID := uuidv7.New()
	nickname := "testuser"
	firstName := "Test"
	lastName := "User"

	profile := &entity.UserProfile{
		Nickname:  &nickname,
		FirstName: &firstName,
		LastName:  &lastName,
		IsPublic:  true,
		Timezone:  "UTC",
		Locale:    "en-US",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.On("GetByUserID", mock.Anything, userID).Return(nil, entity.ErrNotFound).Once()
		mockRepo.On("GetByNickname", mock.Anything, nickname).Return(nil, entity.ErrNotFound).Once()
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.UserProfile) bool {
			return p.UserID == userID && p.Nickname != nil && *p.Nickname == nickname
		})).Return(nil).Once()

		result, err := uc.CreateProfile(context.Background(), userID, profile)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEqual(t, uuidv7.Nil, result.ID)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, nickname, *result.Nickname)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile already exists", func(t *testing.T) {
		existingProfile := &entity.UserProfile{ID: uuidv7.New(), UserID: userID}
		mockRepo.On("GetByUserID", mock.Anything, userID).Return(existingProfile, nil).Once()

		result, err := uc.CreateProfile(context.Background(), userID, profile)

		assert.Error(t, err)
		assert.Equal(t, ErrProfileAlreadyExists, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nickname already taken", func(t *testing.T) {
		otherProfile := &entity.UserProfile{ID: uuidv7.New(), UserID: uuidv7.New()}
		mockRepo.On("GetByUserID", mock.Anything, userID).Return(nil, entity.ErrNotFound).Once()
		mockRepo.On("GetByNickname", mock.Anything, nickname).Return(otherProfile, nil).Once()

		result, err := uc.CreateProfile(context.Background(), userID, profile)

		assert.Error(t, err)
		assert.Equal(t, ErrNicknameAlreadyTaken, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetProfile(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	profileID := uuidv7.New()
	userID := uuidv7.New()
	viewerID := uuidv7.New()

	t.Run("success - public profile", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:       profileID,
			UserID:   userID,
			IsPublic: true,
		}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()

		result, err := uc.GetProfile(context.Background(), profileID, &viewerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, profileID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - own private profile", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:       profileID,
			UserID:   userID,
			IsPublic: false,
		}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()

		result, err := uc.GetProfile(context.Background(), profileID, &userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, profileID, result.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, profileID).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.GetProfile(context.Background(), profileID, &viewerID)

		assert.Error(t, err)
		assert.Equal(t, ErrProfileNotFound, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access to private profile", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:       profileID,
			UserID:   userID,
			IsPublic: false,
		}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()

		result, err := uc.GetProfile(context.Background(), profileID, &viewerID)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorizedProfileAccess, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetProfileByUserID(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	userID := uuidv7.New()
	viewerID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:       uuidv7.New(),
			UserID:   userID,
			IsPublic: true,
		}
		mockRepo.On("GetByUserID", mock.Anything, userID).Return(profile, nil).Once()

		result, err := uc.GetProfileByUserID(context.Background(), userID, &viewerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByUserID", mock.Anything, userID).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.GetProfileByUserID(context.Background(), userID, &viewerID)

		assert.Error(t, err)
		assert.Equal(t, ErrProfileNotFound, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetProfileByNickname(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	nickname := "testuser"
	viewerID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:       uuidv7.New(),
			UserID:   uuidv7.New(),
			Nickname: &nickname,
			IsPublic: true,
		}
		mockRepo.On("GetByNickname", mock.Anything, nickname).Return(profile, nil).Once()

		result, err := uc.GetProfileByNickname(context.Background(), nickname, &viewerID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, nickname, *result.Nickname)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetByNickname", mock.Anything, nickname).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.GetProfileByNickname(context.Background(), nickname, &viewerID)

		assert.Error(t, err)
		assert.Equal(t, ErrProfileNotFound, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateProfile(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	profileID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()
	nickname := "testnick"
	newNickname := "newnick"

	t.Run("success", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:       profileID,
			UserID:   userID,
			Nickname: &nickname,
			Timezone: "UTC",
			Locale:   "en-US",
		}
		updates := &entity.UserProfile{
			Nickname: &newNickname,
		}

		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()
		mockRepo.On("GetByNickname", mock.Anything, newNickname).Return(nil, entity.ErrNotFound).Once()
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.UserProfile) bool {
			return p.ID == profileID && *p.Nickname == newNickname
		})).Return(nil).Once()

		result, err := uc.UpdateProfile(context.Background(), profileID, userID, updates)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newNickname, *result.Nickname)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		updates := &entity.UserProfile{}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(nil, entity.ErrNotFound).Once()

		result, err := uc.UpdateProfile(context.Background(), profileID, userID, updates)

		assert.Error(t, err)
		assert.Equal(t, ErrProfileNotFound, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:     profileID,
			UserID: userID,
		}
		updates := &entity.UserProfile{}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()

		result, err := uc.UpdateProfile(context.Background(), profileID, otherUserID, updates)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorizedProfileAccess, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("nickname already taken", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:       profileID,
			UserID:   userID,
			Nickname: &nickname,
			Timezone: "UTC",
			Locale:   "en-US",
		}
		updates := &entity.UserProfile{
			Nickname: &newNickname,
		}
		otherProfile := &entity.UserProfile{
			ID: uuidv7.New(),
		}

		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()
		mockRepo.On("GetByNickname", mock.Anything, newNickname).Return(otherProfile, nil).Once()

		result, err := uc.UpdateProfile(context.Background(), profileID, userID, updates)

		assert.Error(t, err)
		assert.Equal(t, ErrNicknameAlreadyTaken, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteProfile(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	profileID := uuidv7.New()
	userID := uuidv7.New()
	otherUserID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:     profileID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()
		mockRepo.On("Delete", mock.Anything, profileID).Return(nil).Once()

		err := uc.DeleteProfile(context.Background(), profileID, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, profileID).Return(nil, entity.ErrNotFound).Once()

		err := uc.DeleteProfile(context.Background(), profileID, userID)

		assert.Error(t, err)
		assert.Equal(t, ErrProfileNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:     profileID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()

		err := uc.DeleteProfile(context.Background(), profileID, otherUserID)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorizedProfileAccess, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestListProfiles(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	t.Run("success - public only", func(t *testing.T) {
		profiles := []*entity.UserProfile{
			{ID: uuidv7.New(), IsPublic: true},
			{ID: uuidv7.New(), IsPublic: true},
		}
		isPublic := true
		mockRepo.On("List", mock.Anything, 10, 0, &isPublic).Return(profiles, nil).Once()

		result, err := uc.ListProfiles(context.Background(), 10, 0, true)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success - all profiles", func(t *testing.T) {
		profiles := []*entity.UserProfile{
			{ID: uuidv7.New(), IsPublic: true},
			{ID: uuidv7.New(), IsPublic: false},
		}
		mockRepo.On("List", mock.Anything, 10, 0, (*bool)(nil)).Return(profiles, nil).Once()

		result, err := uc.ListProfiles(context.Background(), 10, 0, false)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})
}

func TestSearchProfiles(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	query := "test"

	t.Run("success", func(t *testing.T) {
		profiles := []*entity.UserProfile{
			{ID: uuidv7.New()},
		}
		mockRepo.On("Search", mock.Anything, query, 10, 0).Return(profiles, nil).Once()

		result, err := uc.SearchProfiles(context.Background(), query, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		mockRepo.AssertExpectations(t)
	})
}

func TestIncrementViews(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	profileID := uuidv7.New()
	userID := uuidv7.New()
	viewerID := uuidv7.New()

	t.Run("success - increment for other viewer", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:     profileID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()
		mockRepo.On("IncrementProfileViews", mock.Anything, profileID).Return(nil).Once()

		err := uc.IncrementViews(context.Background(), profileID, &viewerID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("no increment for own profile", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:     profileID,
			UserID: userID,
		}
		mockRepo.On("GetByID", mock.Anything, profileID).Return(profile, nil).Once()

		err := uc.IncrementViews(context.Background(), profileID, &userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, profileID).Return(nil, entity.ErrNotFound).Once()

		err := uc.IncrementViews(context.Background(), profileID, &viewerID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestBanProfile(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	profileID := uuidv7.New()
	adminID := uuidv7.New()
	reason := "spam"

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Ban", mock.Anything, profileID, reason, adminID).Return(nil).Once()

		err := uc.BanProfile(context.Background(), profileID, reason, adminID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		dbErr := errors.New("db error")
		mockRepo.On("Ban", mock.Anything, profileID, reason, adminID).Return(dbErr).Once()

		err := uc.BanProfile(context.Background(), profileID, reason, adminID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUnbanProfile(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	profileID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("Unban", mock.Anything, profileID).Return(nil).Once()

		err := uc.UnbanProfile(context.Background(), profileID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestVerifyProfile(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	profileID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("SetVerified", mock.Anything, profileID, true).Return(nil).Once()

		err := uc.VerifyProfile(context.Background(), profileID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUnverifyProfile(t *testing.T) {
	mockRepo := new(mockUserProfileRepository)
	uc := NewUserProfileUseCase(mockRepo)

	profileID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.On("SetVerified", mock.Anything, profileID, false).Return(nil).Once()

		err := uc.UnverifyProfile(context.Background(), profileID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
