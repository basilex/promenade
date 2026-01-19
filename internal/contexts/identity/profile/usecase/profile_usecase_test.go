package usecase

import (
	"context"
	"testing"
	"time"

	profileerrors "github.com/basilex/promenade/internal/contexts/identity/profile"
	"github.com/basilex/promenade/internal/contexts/identity/profile/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Repository
type mockRepository struct {
	profiles         map[uuidv7.UUID]*aggregate.Profile
	profilesByUser   map[uuidv7.UUID]*aggregate.Profile
	createErr        error
	getErr           error
	updateErr        error
	deleteErr        error
	existsForUser    bool
	existsForUserErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		profiles:       make(map[uuidv7.UUID]*aggregate.Profile),
		profilesByUser: make(map[uuidv7.UUID]*aggregate.Profile),
	}
}

func (m *mockRepository) Create(ctx context.Context, profile *aggregate.Profile) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.profiles[profile.ID] = profile
	m.profilesByUser[profile.UserID] = profile
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Profile, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	profile, ok := m.profiles[id]
	if !ok {
		return nil, profileerrors.ErrNotFound
	}
	return profile, nil
}

func (m *mockRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*aggregate.Profile, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	profile, ok := m.profilesByUser[userID]
	if !ok {
		return nil, profileerrors.ErrNotFound
	}
	return profile, nil
}

func (m *mockRepository) Update(ctx context.Context, profile *aggregate.Profile) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.profiles[profile.ID] = profile
	m.profilesByUser[profile.UserID] = profile
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if profile, ok := m.profiles[id]; ok {
		delete(m.profiles, id)
		delete(m.profilesByUser, profile.UserID)
	}
	return nil
}

func (m *mockRepository) ListPublicProfiles(ctx context.Context, limit, offset int) ([]*aggregate.Profile, error) {
	profiles := []*aggregate.Profile{}
	for _, profile := range m.profiles {
		if profile.IsPublic {
			profiles = append(profiles, profile)
		}
	}
	return profiles, nil
}

func (m *mockRepository) ExistsForUser(ctx context.Context, userID uuidv7.UUID) (bool, error) {
	if m.existsForUserErr != nil {
		return false, m.existsForUserErr
	}
	_, ok := m.profilesByUser[userID]
	return ok || m.existsForUser, nil
}

func TestUseCase_CreateProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewProfileUseCase(repo)

		profile, err := uc.CreateProfile(ctx, userID, "John Doe")

		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "John Doe", profile.DisplayName)
		assert.Equal(t, userID, profile.UserID)
	})

	t.Run("profile already exists", func(t *testing.T) {
		repo := newMockRepository()
		repo.existsForUser = true
		uc := NewProfileUseCase(repo)

		_, err := uc.CreateProfile(ctx, userID, "John Doe")
		assert.Error(t, err)
	})

	t.Run("invalid display name", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewProfileUseCase(repo)

		_, err := uc.CreateProfile(ctx, userID, "")
		assert.Error(t, err)
	})
}

func TestUseCase_GetProfile(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := newMockRepository()
		profile := &aggregate.Profile{
			UserID:      userID,
			DisplayName: "John Doe",
		}
		profile.ID = profileID
		repo.profiles[profileID] = profile
		uc := NewProfileUseCase(repo)

		result, err := uc.GetProfile(ctx, profileID)

		require.NoError(t, err)
		assert.Equal(t, profile.ID, result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewProfileUseCase(repo)

		_, err := uc.GetProfile(ctx, profileID)
		assert.Error(t, err)
	})
}

func TestUseCase_GetProfileByUserID(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := newMockRepository()
		profile := &aggregate.Profile{
			UserID:      userID,
			DisplayName: "John Doe",
		}
		profile.ID = profileID
		repo.profilesByUser[userID] = profile
		uc := NewProfileUseCase(repo)

		result, err := uc.GetProfileByUserID(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, profile.UserID, result.UserID)
	})

	t.Run("not found", func(t *testing.T) {
		repo := newMockRepository()
		uc := NewProfileUseCase(repo)

		_, err := uc.GetProfileByUserID(ctx, userID)
		assert.Error(t, err)
	})
}

func TestUseCase_UpdateDisplayName(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := newMockRepository()
		profile := &aggregate.Profile{
			UserID:      userID,
			DisplayName: "Old Name",
		}
		profile.ID = profileID
		repo.profiles[profileID] = profile
		uc := NewProfileUseCase(repo)

		err := uc.UpdateDisplayName(ctx, profileID, "New Name")

		require.NoError(t, err)
		assert.Equal(t, "New Name", repo.profiles[profileID].DisplayName)
	})

	t.Run("invalid display name", func(t *testing.T) {
		repo := newMockRepository()
		profile := &aggregate.Profile{
			UserID:      userID,
			DisplayName: "Old Name",
		}
		profile.ID = profileID
		repo.profiles[profileID] = profile
		uc := NewProfileUseCase(repo)

		err := uc.UpdateDisplayName(ctx, profileID, "A")
		assert.Error(t, err)
	})
}

func TestUseCase_UpdateBio(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID: userID,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	err := uc.UpdateBio(ctx, profileID, "This is my bio")

	require.NoError(t, err)
	assert.Equal(t, "This is my bio", repo.profiles[profileID].Bio)
}

func TestUseCase_UpdateAvatar(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("valid URL", func(t *testing.T) {
		repo := newMockRepository()
		profile := &aggregate.Profile{
			UserID: userID,
		}
		profile.ID = profileID
		repo.profiles[profileID] = profile
		uc := NewProfileUseCase(repo)

		err := uc.UpdateAvatar(ctx, profileID, "https://example.com/avatar.jpg")

		require.NoError(t, err)
		assert.Equal(t, "https://example.com/avatar.jpg", repo.profiles[profileID].AvatarURL)
	})

	t.Run("invalid URL", func(t *testing.T) {
		repo := newMockRepository()
		profile := &aggregate.Profile{
			UserID: userID,
		}
		profile.ID = profileID
		repo.profiles[profileID] = profile
		uc := NewProfileUseCase(repo)

		err := uc.UpdateAvatar(ctx, profileID, "not-a-url")
		assert.Error(t, err)
	})
}

func TestUseCase_UpdatePersonalInfo(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID: userID,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	err := uc.UpdatePersonalInfo(ctx, profileID, "John", "Doe", "Michael")

	require.NoError(t, err)
	assert.Equal(t, "John", repo.profiles[profileID].FirstName)
	assert.Equal(t, "Doe", repo.profiles[profileID].LastName)
	assert.Equal(t, "Michael", repo.profiles[profileID].MiddleName)
}

func TestUseCase_UpdateGender(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID: userID,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	err := uc.UpdateGender(ctx, profileID, aggregate.GenderMale)

	require.NoError(t, err)
	assert.Equal(t, aggregate.GenderMale, repo.profiles[profileID].Gender)
}

func TestUseCase_UpdateDateOfBirth(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID: userID,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	err := uc.UpdateDateOfBirth(ctx, profileID, &dob)

	require.NoError(t, err)
	assert.NotNil(t, repo.profiles[profileID].DateOfBirth)
}

func TestUseCase_UpdateLocalization(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID: userID,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	err := uc.UpdateLocalization(ctx, profileID, "Europe/Kyiv", "uk", "UA")

	require.NoError(t, err)
	assert.Equal(t, "Europe/Kyiv", repo.profiles[profileID].Timezone)
	assert.Equal(t, "uk", repo.profiles[profileID].Language)
	assert.Equal(t, "UA", repo.profiles[profileID].Country)
}

func TestUseCase_UpdateSocialLinks(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID: userID,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	err := uc.UpdateSocialLinks(ctx, profileID,
		"https://example.com",
		"https://linkedin.com/in/john",
		"", "", "", "")

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", repo.profiles[profileID].Website)
	assert.Equal(t, "https://linkedin.com/in/john", repo.profiles[profileID].LinkedIn)
}

func TestUseCase_Visibility(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID:   userID,
		IsPublic: true,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	t.Run("set private", func(t *testing.T) {
		err := uc.SetPrivate(ctx, profileID)
		require.NoError(t, err)
		assert.False(t, repo.profiles[profileID].IsPublic)
	})

	t.Run("set public", func(t *testing.T) {
		err := uc.SetPublic(ctx, profileID)
		require.NoError(t, err)
		assert.True(t, repo.profiles[profileID].IsPublic)
	})
}

func TestUseCase_Status(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID:   userID,
		IsActive: true,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	t.Run("deactivate", func(t *testing.T) {
		err := uc.Deactivate(ctx, profileID)
		require.NoError(t, err)
		assert.False(t, repo.profiles[profileID].IsActive)
	})

	t.Run("activate", func(t *testing.T) {
		err := uc.Activate(ctx, profileID)
		require.NoError(t, err)
		assert.True(t, repo.profiles[profileID].IsActive)
	})
}

func TestUseCase_DeleteProfile(t *testing.T) {
	ctx := context.Background()
	profileID := uuidv7.New()
	userID := uuidv7.New()

	repo := newMockRepository()
	profile := &aggregate.Profile{
		UserID: userID,
	}
	profile.ID = profileID
	repo.profiles[profileID] = profile
	uc := NewProfileUseCase(repo)

	err := uc.DeleteProfile(ctx, profileID)

	require.NoError(t, err)
	_, exists := repo.profiles[profileID]
	assert.False(t, exists)
}

func TestUseCase_ListPublicProfiles(t *testing.T) {
	ctx := context.Background()

	repo := newMockRepository()
	// Add public profile
	publicProfile := &aggregate.Profile{
		UserID:   uuidv7.New(),
		IsPublic: true,
	}
	publicProfile.ID = uuidv7.New()
	repo.profiles[publicProfile.ID] = publicProfile

	// Add private profile
	privateProfile := &aggregate.Profile{
		UserID:   uuidv7.New(),
		IsPublic: false,
	}
	privateProfile.ID = uuidv7.New()
	repo.profiles[privateProfile.ID] = privateProfile

	uc := NewProfileUseCase(repo)

	profiles, err := uc.ListPublicProfiles(ctx, 20, 0)

	require.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.True(t, profiles[0].IsPublic)
}
