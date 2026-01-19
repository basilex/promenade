package profile_test

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/profile"
	profileHTTP "github.com/basilex/promenade/internal/contexts/identity/profile/adapter/http"
	profileAggregate "github.com/basilex/promenade/internal/contexts/identity/profile/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
	"github.com/stretchr/testify/assert"
)

// MockProfileUseCase mocks profile.IUseCase
type MockProfileUseCase struct {
	CreateProfileFunc      func(ctx context.Context, userID uuidv7.UUID, displayName string) (*profileAggregate.Profile, error)
	GetProfileFunc         func(ctx context.Context, profileID uuidv7.UUID) (*profileAggregate.Profile, error)
	GetProfileByUserIDFunc func(ctx context.Context, userID uuidv7.UUID) (*profileAggregate.Profile, error)
	UpdateDisplayNameFunc  func(ctx context.Context, profileID uuidv7.UUID, displayName string) error
	DeleteProfileFunc      func(ctx context.Context, profileID uuidv7.UUID) error
	ListPublicProfilesFunc func(ctx context.Context, limit, offset int) ([]*profileAggregate.Profile, error)
}

func (m *MockProfileUseCase) CreateProfile(ctx context.Context, userID uuidv7.UUID, displayName string) (*profileAggregate.Profile, error) {
	if m.CreateProfileFunc != nil {
		return m.CreateProfileFunc(ctx, userID, displayName)
	}
	return nil, nil
}

func (m *MockProfileUseCase) GetProfile(ctx context.Context, profileID uuidv7.UUID) (*profileAggregate.Profile, error) {
	if m.GetProfileFunc != nil {
		return m.GetProfileFunc(ctx, profileID)
	}
	return nil, nil
}

func (m *MockProfileUseCase) GetProfileByUserID(ctx context.Context, userID uuidv7.UUID) (*profileAggregate.Profile, error) {
	if m.GetProfileByUserIDFunc != nil {
		return m.GetProfileByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockProfileUseCase) UpdateDisplayName(ctx context.Context, profileID uuidv7.UUID, displayName string) error {
	if m.UpdateDisplayNameFunc != nil {
		return m.UpdateDisplayNameFunc(ctx, profileID, displayName)
	}
	return nil
}

func (m *MockProfileUseCase) DeleteProfile(ctx context.Context, profileID uuidv7.UUID) error {
	if m.DeleteProfileFunc != nil {
		return m.DeleteProfileFunc(ctx, profileID)
	}
	return nil
}

func (m *MockProfileUseCase) ListPublicProfiles(ctx context.Context, limit, offset int) ([]*profileAggregate.Profile, error) {
	if m.ListPublicProfilesFunc != nil {
		return m.ListPublicProfilesFunc(ctx, limit, offset)
	}
	return nil, nil
}

// Stub methods (not tested)
func (m *MockProfileUseCase) UpdateBio(ctx context.Context, profileID uuidv7.UUID, bio string) error {
	return nil
}
func (m *MockProfileUseCase) UpdateAvatar(ctx context.Context, profileID uuidv7.UUID, avatarURL string) error {
	return nil
}
func (m *MockProfileUseCase) UpdatePersonalInfo(ctx context.Context, profileID uuidv7.UUID, firstName, lastName, middleName string) error {
	return nil
}
func (m *MockProfileUseCase) UpdateGender(ctx context.Context, profileID uuidv7.UUID, gender profileAggregate.Gender) error {
	return nil
}
func (m *MockProfileUseCase) UpdateDateOfBirth(ctx context.Context, profileID uuidv7.UUID, dateOfBirth *time.Time) error {
	return nil
}
func (m *MockProfileUseCase) UpdateLocalization(ctx context.Context, profileID uuidv7.UUID, timezone, language, country string) error {
	return nil
}
func (m *MockProfileUseCase) UpdateSocialLinks(ctx context.Context, profileID uuidv7.UUID, website, linkedin, twitter, github, facebook, instagram string) error {
	return nil
}
func (m *MockProfileUseCase) SetPublic(ctx context.Context, profileID uuidv7.UUID) error {
	return nil
}
func (m *MockProfileUseCase) SetPrivate(ctx context.Context, profileID uuidv7.UUID) error {
	return nil
}
func (m *MockProfileUseCase) Activate(ctx context.Context, profileID uuidv7.UUID) error {
	return nil
}
func (m *MockProfileUseCase) Deactivate(ctx context.Context, profileID uuidv7.UUID) error {
	return nil
}

func fakeProfile() *profileAggregate.Profile {
	userID := uuidv7.New()
	p, _ := profileAggregate.NewProfile(userID, "Test User")
	return p
}

func TestProfileHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProfileUseCase{
		CreateProfileFunc: func(ctx context.Context, userID uuidv7.UUID, displayName string) (*profileAggregate.Profile, error) {
			return fakeProfile(), nil
		},
	}

	handler := profileHTTP.NewProfileHandler(mockUC)
	router.POST("/profiles", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/profiles?user_id="+smoke.FakeUUID(), map[string]any{
		"display_name": "Test User",
	})

	smoke.AssertSuccessResponse(t, resp, 201)
}

func TestProfileHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()
	mockUC := &MockProfileUseCase{}
	handler := profileHTTP.NewProfileHandler(mockUC)
	router.POST("/profiles", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/profiles?user_id="+smoke.FakeUUID(), map[string]any{
		// Missing display_name
	})

	smoke.AssertErrorResponse(t, resp, 400, "BAD_REQUEST")
}

func TestProfileHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProfileUseCase{
		GetProfileFunc: func(ctx context.Context, profileID uuidv7.UUID) (*profileAggregate.Profile, error) {
			return fakeProfile(), nil
		},
	}

	handler := profileHTTP.NewProfileHandler(mockUC)
	router.GET("/profiles/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/profiles/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestProfileHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProfileUseCase{
		GetProfileFunc: func(ctx context.Context, profileID uuidv7.UUID) (*profileAggregate.Profile, error) {
			return nil, profile.ErrNotFound
		},
	}

	handler := profileHTTP.NewProfileHandler(mockUC)
	router.GET("/profiles/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/profiles/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 404, "NOT_FOUND")
}

func TestProfileHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProfileUseCase{
		UpdateDisplayNameFunc: func(ctx context.Context, profileID uuidv7.UUID, displayName string) error {
			return nil
		},
	}

	handler := profileHTTP.NewProfileHandler(mockUC)
	router.PUT("/profiles/:id/display-name", handler.UpdateDisplayName)

	resp := smoke.MakeRequest(t, router, "PUT", "/profiles/"+smoke.FakeUUID()+"/display-name", map[string]any{
		"display_name": "Updated Name",
	})

	smoke.AssertSuccessResponse(t, resp, 200)
}

func TestProfileHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProfileUseCase{
		DeleteProfileFunc: func(ctx context.Context, profileID uuidv7.UUID) error {
			return nil
		},
	}

	handler := profileHTTP.NewProfileHandler(mockUC)
	router.DELETE("/profiles/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/profiles/"+smoke.FakeUUID(), nil)

	assert.Equal(t, 204, resp.Code)
}

func TestProfileHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProfileUseCase{
		DeleteProfileFunc: func(ctx context.Context, profileID uuidv7.UUID) error {
			return profile.ErrNotFound
		},
	}

	handler := profileHTTP.NewProfileHandler(mockUC)
	router.DELETE("/profiles/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/profiles/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, resp, 500, "INTERNAL_ERROR")
}

func TestProfileHandler_ListPublic_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProfileUseCase{
		ListPublicProfilesFunc: func(ctx context.Context, limit, offset int) ([]*profileAggregate.Profile, error) {
			return []*profileAggregate.Profile{fakeProfile()}, nil
		},
	}

	handler := profileHTTP.NewProfileHandler(mockUC)
	router.GET("/profiles", handler.ListPublic)

	resp := smoke.MakeRequest(t, router, "GET", "/profiles?limit=10&offset=0", nil)

	assert.Equal(t, 200, resp.Code)
}

func TestProfileHandler_ListPublic_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockProfileUseCase{
		ListPublicProfilesFunc: func(ctx context.Context, limit, offset int) ([]*profileAggregate.Profile, error) {
			return []*profileAggregate.Profile{}, nil
		},
	}

	handler := profileHTTP.NewProfileHandler(mockUC)
	router.GET("/profiles", handler.ListPublic)

	resp := smoke.MakeRequest(t, router, "GET", "/profiles?limit=10&offset=0", nil)

	assert.Equal(t, 200, resp.Code)
}
