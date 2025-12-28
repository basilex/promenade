package profile_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/identity/profile"
	profileHTTP "github.com/basilex/promenade/internal/contexts/identity/profile/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockProfileUseCase is a mock implementation of profile.IUseCase for testing
type MockProfileUseCase struct {
	mock.Mock
}

func (m *MockProfileUseCase) CreateProfile(ctx context.Context, userID uuidv7.UUID, displayName string) (*profile.Profile, error) {
	args := m.Called(ctx, userID, displayName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileUseCase) GetProfile(ctx context.Context, profileID uuidv7.UUID) (*profile.Profile, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileUseCase) GetProfileByUserID(ctx context.Context, userID uuidv7.UUID) (*profile.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*profile.Profile), args.Error(1)
}

func (m *MockProfileUseCase) UpdateDisplayName(ctx context.Context, profileID uuidv7.UUID, displayName string) error {
	args := m.Called(ctx, profileID, displayName)
	return args.Error(0)
}

func (m *MockProfileUseCase) UpdateBio(ctx context.Context, profileID uuidv7.UUID, bio string) error {
	args := m.Called(ctx, profileID, bio)
	return args.Error(0)
}

func (m *MockProfileUseCase) UpdateAvatar(ctx context.Context, profileID uuidv7.UUID, avatarURL string) error {
	args := m.Called(ctx, profileID, avatarURL)
	return args.Error(0)
}

func (m *MockProfileUseCase) UpdatePersonalInfo(ctx context.Context, profileID uuidv7.UUID, firstName, lastName, middleName string) error {
	args := m.Called(ctx, profileID, firstName, lastName, middleName)
	return args.Error(0)
}

func (m *MockProfileUseCase) UpdateGender(ctx context.Context, profileID uuidv7.UUID, gender profile.Gender) error {
	args := m.Called(ctx, profileID, gender)
	return args.Error(0)
}

func (m *MockProfileUseCase) UpdateDateOfBirth(ctx context.Context, profileID uuidv7.UUID, dateOfBirth *time.Time) error {
	args := m.Called(ctx, profileID, dateOfBirth)
	return args.Error(0)
}

func (m *MockProfileUseCase) UpdateLocalization(ctx context.Context, profileID uuidv7.UUID, timezone, language, country string) error {
	args := m.Called(ctx, profileID, timezone, language, country)
	return args.Error(0)
}

func (m *MockProfileUseCase) UpdateSocialLinks(ctx context.Context, profileID uuidv7.UUID, website, linkedin, twitter, github, facebook, instagram string) error {
	args := m.Called(ctx, profileID, website, linkedin, twitter, github, facebook, instagram)
	return args.Error(0)
}

func (m *MockProfileUseCase) SetPublic(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockProfileUseCase) SetPrivate(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockProfileUseCase) Activate(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockProfileUseCase) Deactivate(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockProfileUseCase) DeleteProfile(ctx context.Context, profileID uuidv7.UUID) error {
	args := m.Called(ctx, profileID)
	return args.Error(0)
}

func (m *MockProfileUseCase) ListPublicProfiles(ctx context.Context, limit, offset int) ([]*profile.Profile, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*profile.Profile), args.Error(1)
}

func setupProfileRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestProfileHandler_Create(t *testing.T) {
	mockUC := new(MockProfileUseCase)
	handler := profileHTTP.NewProfileHandler(mockUC)
	router := setupProfileRouter()
	router.POST("/profiles", handler.Create)

	userID := uuidv7.New()
	profileID := uuidv7.New()

	expectedProfile := &profile.Profile{
		ID:          profileID,
		UserID:      userID,
		DisplayName: "John Doe",
		Gender:      profile.GenderNotSpecify,
		IsPublic:    true,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Mock expects: ctx, userID, displayName
	mockUC.On("CreateProfile", mock.Anything, userID, "John Doe").Return(expectedProfile, nil)

	body := map[string]interface{}{
		"display_name": "John Doe",
	}
	jsonBody, _ := json.Marshal(body)

	// user_id is passed as query parameter
	req := httptest.NewRequest(http.MethodPost, "/profiles?user_id="+userID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUC.AssertExpectations(t)
}

func TestProfileHandler_GetByID(t *testing.T) {
	mockUC := new(MockProfileUseCase)
	handler := profileHTTP.NewProfileHandler(mockUC)
	router := setupProfileRouter()
	router.GET("/profiles/:id", handler.GetByID)

	profileID := uuidv7.New()
	userID := uuidv7.New()

	expectedProfile := &profile.Profile{
		ID:          profileID,
		UserID:      userID,
		DisplayName: "John Doe",
		Bio:         "Software Developer",
		Gender:      profile.GenderMale,
		IsPublic:    true,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUC.On("GetProfile", mock.Anything, profileID).Return(expectedProfile, nil)

	req := httptest.NewRequest(http.MethodGet, "/profiles/"+profileID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestProfileHandler_GetByUserID(t *testing.T) {
	mockUC := new(MockProfileUseCase)
	handler := profileHTTP.NewProfileHandler(mockUC)
	router := setupProfileRouter()
	router.GET("/profiles/user/:user_id", handler.GetByUserID)

	profileID := uuidv7.New()
	userID := uuidv7.New()

	expectedProfile := &profile.Profile{
		ID:          profileID,
		UserID:      userID,
		DisplayName: "Jane Smith",
		Gender:      profile.GenderFemale,
		IsPublic:    true,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUC.On("GetProfileByUserID", mock.Anything, userID).Return(expectedProfile, nil)

	req := httptest.NewRequest(http.MethodGet, "/profiles/user/"+userID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestProfileHandler_UpdateDisplayName(t *testing.T) {
	mockUC := new(MockProfileUseCase)
	handler := profileHTTP.NewProfileHandler(mockUC)
	router := setupProfileRouter()
	router.PUT("/profiles/:id/display-name", handler.UpdateDisplayName)

	profileID := uuidv7.New()

	mockUC.On("UpdateDisplayName", mock.Anything, profileID, "New Name").Return(nil)

	body := map[string]interface{}{
		"display_name": "New Name",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/profiles/"+profileID.String()+"/display-name", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestProfileHandler_UpdateBio(t *testing.T) {
	mockUC := new(MockProfileUseCase)
	handler := profileHTTP.NewProfileHandler(mockUC)
	router := setupProfileRouter()
	router.PUT("/profiles/:id/bio", handler.UpdateBio)

	profileID := uuidv7.New()

	mockUC.On("UpdateBio", mock.Anything, profileID, "New bio").Return(nil)

	body := map[string]interface{}{
		"bio": "New bio",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/profiles/"+profileID.String()+"/bio", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestProfileHandler_SetPublic(t *testing.T) {
	mockUC := new(MockProfileUseCase)
	handler := profileHTTP.NewProfileHandler(mockUC)
	router := setupProfileRouter()
	router.PUT("/profiles/:id/public", handler.SetPublic)

	profileID := uuidv7.New()

	mockUC.On("SetPublic", mock.Anything, profileID).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/profiles/"+profileID.String()+"/public", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestProfileHandler_Delete(t *testing.T) {
	mockUC := new(MockProfileUseCase)
	handler := profileHTTP.NewProfileHandler(mockUC)
	router := setupProfileRouter()
	router.DELETE("/profiles/:id", handler.Delete)

	profileID := uuidv7.New()

	mockUC.On("DeleteProfile", mock.Anything, profileID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/profiles/"+profileID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Handler returns 204 No Content for successful delete
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUC.AssertExpectations(t)
}

func TestProfileHandler_ListPublic(t *testing.T) {
	mockUC := new(MockProfileUseCase)
	handler := profileHTTP.NewProfileHandler(mockUC)
	router := setupProfileRouter()
	router.GET("/profiles", handler.ListPublic)

	profiles := []*profile.Profile{
		{
			ID:          uuidv7.New(),
			UserID:      uuidv7.New(),
			DisplayName: "User 1",
			IsPublic:    true,
			IsActive:    true,
		},
		{
			ID:          uuidv7.New(),
			UserID:      uuidv7.New(),
			DisplayName: "User 2",
			IsPublic:    true,
			IsActive:    true,
		},
	}

	mockUC.On("ListPublicProfiles", mock.Anything, 20, 0).Return(profiles, nil)

	req := httptest.NewRequest(http.MethodGet, "/profiles?limit=20&offset=0", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}
