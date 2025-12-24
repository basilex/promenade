package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/modules/profiles/adapter/http/dto"
	"github.com/basilex/promenade/internal/modules/profiles/entity"
	"github.com/basilex/promenade/internal/modules/profiles/usecase"
	ucmocks "github.com/basilex/promenade/internal/modules/profiles/usecase/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

//  IModule-independent test: imports only module and pkg, no core dependencies

func setupProfileTest() (*gin.Engine, *ucmocks.MockUserProfileUseCase) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockUC := new(ucmocks.MockUserProfileUseCase)
	return router, mockUC
}

func TestCreateProfile(t *testing.T) {
	router, mockUC := setupProfileTest()
	handler := NewUserProfileHandler(mockUC)
	router.POST("/profiles", handler.CreateProfile)

	userID := uuidv7.New()
	nickname := "testuser"
	firstName := "Test"

	t.Run("success", func(t *testing.T) {
		reqBody := dto.CreateProfileRequest{
			Nickname:  &nickname,
			FirstName: &firstName,
			Timezone:  "UTC",
			Locale:    "en-US",
		}
		body, _ := json.Marshal(reqBody)

		profile := &entity.UserProfile{
			ID:        uuidv7.New(),
			UserID:    userID,
			Nickname:  &nickname,
			FirstName: &firstName,
			Timezone:  "UTC",
			Locale:    "en-US",
		}

		mockUC.On("CreateProfile", mock.Anything, userID, mock.Anything).Return(profile, nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)

		handler.CreateProfile(ctx)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["success"])
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		reqBody := dto.CreateProfileRequest{Timezone: "UTC", Locale: "en-US"}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		// No user_id set

		handler.CreateProfile(ctx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)

		handler.CreateProfile(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("profile already exists", func(t *testing.T) {
		reqBody := dto.CreateProfileRequest{Timezone: "UTC", Locale: "en-US"}
		body, _ := json.Marshal(reqBody)

		mockUC.On("CreateProfile", mock.Anything, userID, mock.Anything).
			Return(nil, usecase.ErrProfileAlreadyExists).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)

		handler.CreateProfile(ctx)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("nickname already taken", func(t *testing.T) {
		reqBody := dto.CreateProfileRequest{
			Nickname: &nickname,
			Timezone: "UTC",
			Locale:   "en-US",
		}
		body, _ := json.Marshal(reqBody)

		mockUC.On("CreateProfile", mock.Anything, userID, mock.Anything).
			Return(nil, usecase.ErrNicknameAlreadyTaken).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/profiles", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("user_id", userID)

		handler.CreateProfile(ctx)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestGetProfile(t *testing.T) {
	router, mockUC := setupProfileTest()
	handler := NewUserProfileHandler(mockUC)
	router.GET("/profiles/:id", handler.GetProfile)

	profileID := uuidv7.New()
	userID := uuidv7.New()
	nickname := "testuser"

	t.Run("success", func(t *testing.T) {
		profile := &entity.UserProfile{
			ID:       profileID,
			UserID:   userID,
			Nickname: &nickname,
			Timezone: "UTC",
			Locale:   "en-US",
		}

		mockUC.On("GetProfile", mock.Anything, profileID, mock.Anything).Return(profile, nil).Once()
		mockUC.On("IncrementViews", mock.Anything, profileID, mock.Anything).Return(nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/profiles/"+profileID.String(), nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		ctx.Set("user_id", userID)

		handler.GetProfile(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["success"])
		mockUC.AssertExpectations(t)
	})

	t.Run("invalid profile ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/profiles/invalid-uuid", nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}

		handler.GetProfile(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("profile not found", func(t *testing.T) {
		mockUC.On("GetProfile", mock.Anything, profileID, mock.Anything).Return(nil, usecase.ErrProfileNotFound).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/profiles/"+profileID.String(), nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}

		handler.GetProfile(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockUC.On("GetProfile", mock.Anything, profileID, mock.Anything).Return(nil, usecase.ErrUnauthorizedProfileAccess).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/profiles/"+profileID.String(), nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}

		handler.GetProfile(ctx)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestUpdateProfile(t *testing.T) {
	router, mockUC := setupProfileTest()
	handler := NewUserProfileHandler(mockUC)
	router.PUT("/profiles/:id", handler.UpdateProfile)

	profileID := uuidv7.New()
	userID := uuidv7.New()
	newNickname := "newnick"

	t.Run("success", func(t *testing.T) {
		reqBody := dto.UpdateProfileRequest{
			Nickname: &newNickname,
		}
		body, _ := json.Marshal(reqBody)

		profile := &entity.UserProfile{
			ID:       profileID,
			UserID:   userID,
			Nickname: &newNickname,
			Timezone: "UTC",
			Locale:   "en-US",
		}

		mockUC.On("UpdateProfile", mock.Anything, profileID, userID, mock.Anything).Return(profile, nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/profiles/"+profileID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		ctx.Set("user_id", userID)

		handler.UpdateProfile(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]any
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, true, response["success"])
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		reqBody := dto.UpdateProfileRequest{}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/profiles/"+profileID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		// No user_id set

		handler.UpdateProfile(ctx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("profile not found", func(t *testing.T) {
		reqBody := dto.UpdateProfileRequest{Nickname: &newNickname}
		body, _ := json.Marshal(reqBody)

		mockUC.On("UpdateProfile", mock.Anything, profileID, userID, mock.Anything).
			Return(nil, usecase.ErrProfileNotFound).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/profiles/"+profileID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		ctx.Set("user_id", userID)

		handler.UpdateProfile(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized profile access", func(t *testing.T) {
		reqBody := dto.UpdateProfileRequest{Nickname: &newNickname}
		body, _ := json.Marshal(reqBody)

		mockUC.On("UpdateProfile", mock.Anything, profileID, userID, mock.Anything).
			Return(nil, usecase.ErrUnauthorizedProfileAccess).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/profiles/"+profileID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		ctx.Set("user_id", userID)

		handler.UpdateProfile(ctx)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestDeleteProfile(t *testing.T) {
	router, mockUC := setupProfileTest()
	handler := NewUserProfileHandler(mockUC)
	router.DELETE("/profiles/:id", handler.DeleteProfile)

	profileID := uuidv7.New()
	userID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		mockUC.On("DeleteProfile", mock.Anything, profileID, userID).Return(nil).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/profiles/"+profileID.String(), nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		ctx.Set("user_id", userID)

		handler.DeleteProfile(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/profiles/"+profileID.String(), nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		// No user_id set

		handler.DeleteProfile(ctx)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("profile not found", func(t *testing.T) {
		mockUC.On("DeleteProfile", mock.Anything, profileID, userID).Return(usecase.ErrProfileNotFound).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/profiles/"+profileID.String(), nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		ctx.Set("user_id", userID)

		handler.DeleteProfile(ctx)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("unauthorized profile access", func(t *testing.T) {
		mockUC.On("DeleteProfile", mock.Anything, profileID, userID).Return(usecase.ErrUnauthorizedProfileAccess).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/profiles/"+profileID.String(), nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		ctx.Set("user_id", userID)

		handler.DeleteProfile(ctx)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("internal error", func(t *testing.T) {
		mockUC.On("DeleteProfile", mock.Anything, profileID, userID).Return(errors.New("db error")).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/profiles/"+profileID.String(), nil)

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "id", Value: profileID.String()}}
		ctx.Set("user_id", userID)

		handler.DeleteProfile(ctx)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}
