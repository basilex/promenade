package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.POST("/auth/register", handler.Register)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.POST("/auth/login", handler.Login)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestAuthHandler_Logout_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.POST("/auth/logout", handler.Logout)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/logout", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_RefreshToken_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.POST("/auth/refresh", handler.RefreshToken)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_GetMe_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.GET("/auth/me", handler.GetMe)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/me", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_GetUserSessions_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.GET("/auth/sessions", handler.GetUserSessions)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth/sessions", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_SuspendUser_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.POST("/auth/users/:id/suspend", handler.SuspendUser)

	reqBody := dto.SuspendUserRequest{
		Reason: "Test reason",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/users/invalid-uuid/suspend", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_BanUser_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.POST("/auth/users/:id/ban", handler.BanUser)

	reqBody := dto.BanUserRequest{
		Reason: "Test reason",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/users/invalid-uuid/ban", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_ReactivateUser_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &AuthHandler{}
	router.POST("/auth/users/:id/reactivate", handler.ReactivateUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/users/invalid-uuid/reactivate", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
