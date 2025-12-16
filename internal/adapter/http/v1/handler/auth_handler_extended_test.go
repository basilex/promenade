package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthExtendedTest() (*gin.Engine, *mocks.MockAuthUseCase) {
	gin.SetMode(gin.TestMode)
	mockUC := new(mocks.MockAuthUseCase)
	handler := NewAuthHandler(mockUC)

	router := gin.New()
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	router.POST("/logout", handler.Logout)
	router.POST("/refresh", handler.RefreshToken)
	router.GET("/me", handler.GetMe)
	router.GET("/sessions", handler.GetUserSessions)
	router.POST("/auth/users/:id/suspend", handler.SuspendUser)
	router.POST("/auth/users/:id/ban", handler.BanUser)
	router.POST("/auth/users/:id/reactivate", handler.ReactivateUser)

	return router, mockUC
}

// Positive tests for Register
func TestAuthHandler_Register_Success(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	user := &entity.User{
		ID:     uuidv7.New(),
		Email:  req.Email,
		Name:   req.Name,
		Status: entity.UserStatusActive,
	}

	mockUC.On("Register", mock.Anything, req.Email, req.Name, req.Password).
		Return(user, nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.RegisterRequest{
		Email:    "existing@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	mockUC.On("Register", mock.Anything, req.Email, req.Name, req.Password).
		Return(nil, usecase.ErrEmailAlreadyExists)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_Register_InternalError(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	mockUC.On("Register", mock.Anything, req.Email, req.Name, req.Password).
		Return(nil, errors.New("database error"))

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// Login tests
func TestAuthHandler_Login_Success(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	userID := uuidv7.New()
	user := &entity.User{
		ID:     userID,
		Email:  req.Email,
		Name:   "Test User",
		Status: entity.UserStatusActive,
	}

	accessToken := "access_token"
	refreshToken := "refresh_token"

	mockUC.On("Login", mock.Anything, req.Email, req.Password, mock.Anything, mock.Anything).
		Return(accessToken, refreshToken, user, nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockUC.On("Login", mock.Anything, req.Email, req.Password, mock.Anything, mock.Anything).
		Return("", "", nil, usecase.ErrInvalidCredentials)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_Login_UserSuspended(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.LoginRequest{
		Email:    "suspended@example.com",
		Password: "password123",
	}

	mockUC.On("Login", mock.Anything, req.Email, req.Password, mock.Anything, mock.Anything).
		Return("", "", nil, usecase.ErrUserSuspended)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_Login_UserBanned(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.LoginRequest{
		Email:    "banned@example.com",
		Password: "password123",
	}

	mockUC.On("Login", mock.Anything, req.Email, req.Password, mock.Anything, mock.Anything).
		Return("", "", nil, usecase.ErrUserBanned)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockUC.AssertExpectations(t)
}

// RefreshToken tests
func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	accessToken := "new_access_token"
	refreshToken := "new_refresh_token"

	mockUC.On("RefreshToken", mock.Anything, req.RefreshToken).
		Return(accessToken, refreshToken, nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	req := dto.RefreshTokenRequest{
		RefreshToken: "invalid_refresh_token",
	}

	mockUC.On("RefreshToken", mock.Anything, req.RefreshToken).
		Return("", "", usecase.ErrInvalidToken)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockUC.AssertExpectations(t)
}

// SuspendUser tests
func TestAuthHandler_SuspendUser_Success(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	userID := uuidv7.New()
	mockUC.On("SuspendUser", mock.Anything, userID, mock.Anything, mock.Anything).Return(nil)

	body := `{"reason": "test suspension"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/users/"+userID.String()+"/suspend", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_SuspendUser_NotFound(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	userID := uuidv7.New()
	mockUC.On("SuspendUser", mock.Anything, userID, mock.Anything, mock.Anything).Return(entity.ErrNotFound)

	body := `{"reason": "test suspension"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/users/"+userID.String()+"/suspend", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// BanUser tests
func TestAuthHandler_BanUser_Success(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	userID := uuidv7.New()
	mockUC.On("BanUser", mock.Anything, userID, mock.Anything).Return(nil)

	body := `{"reason": "test ban"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/users/"+userID.String()+"/ban", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_BanUser_NotFound(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	userID := uuidv7.New()
	mockUC.On("BanUser", mock.Anything, userID, mock.Anything).Return(entity.ErrNotFound)

	body := `{"reason": "test ban"}`
	r := httptest.NewRequest(http.MethodPost, "/auth/users/"+userID.String()+"/ban", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// ReactivateUser tests
func TestAuthHandler_ReactivateUser_Success(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	userID := uuidv7.New()
	mockUC.On("ReactivateUser", mock.Anything, userID).Return(nil)

	r := httptest.NewRequest(http.MethodPost, "/auth/users/"+userID.String()+"/reactivate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestAuthHandler_ReactivateUser_NotFound(t *testing.T) {
	router, mockUC := setupAuthExtendedTest()

	userID := uuidv7.New()
	mockUC.On("ReactivateUser", mock.Anything, userID).Return(entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodPost, "/auth/users/"+userID.String()+"/reactivate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}
