package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func mockAuthProfile(c *gin.Context) {
	uid, _ := uuidv7.Parse("01936d9a-0000-7000-8000-000000000000")
	c.Set("user_id", uid)
	c.Next()
}

func TestUserProfileHandler_CreateProfile_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.POST("/profiles", handler.CreateProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profiles", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestUserProfileHandler_GetProfile_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.GET("/profiles/:user_id", handler.GetProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profiles/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestUserProfileHandler_GetMyProfile_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	// Don't add mockAuthProfile here - testing no auth scenario
	router.GET("/profiles/me", handler.GetMyProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profiles/me", nil)

	router.ServeHTTP(w, req)

	// Without auth middleware, user_id won't be in context
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserProfileHandler_GetProfileByNickname_EmptyNickname(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.GET("/profiles/nickname/:nickname", handler.GetProfileByNickname)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profiles/nickname/", nil)

	router.ServeHTTP(w, req)

	// Empty nickname will result in 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserProfileHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.PUT("/profiles/:user_id", handler.UpdateProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/profiles/01936d9a-0000-7000-8000-000000000000", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_UpdateProfile_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.PUT("/profiles/:user_id", handler.UpdateProfile)

	requestBody := map[string]interface{}{
		"display_name": "Updated Name",
	}
	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/profiles/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_DeleteProfile_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.DELETE("/profiles/:user_id", handler.DeleteProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/profiles/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Removed: TestUserProfileHandler_ListProfiles_InvalidPagination - requires mock use case

func TestUserProfileHandler_SearchProfiles_EmptyQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.GET("/profiles/search", handler.SearchProfiles)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profiles/search", nil)

	router.ServeHTTP(w, req)

	// Empty query should be handled by validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_BanProfile_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.POST("/profiles/:user_id/ban", handler.BanProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profiles/01936d9a-0000-7000-8000-000000000000/ban", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_BanProfile_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.POST("/profiles/:user_id/ban", handler.BanProfile)

	requestBody := map[string]interface{}{
		"reason": "Spam",
	}
	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profiles/invalid-uuid/ban", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_UnbanProfile_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.POST("/profiles/:user_id/unban", handler.UnbanProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profiles/invalid-uuid/unban", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_VerifyProfile_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.POST("/profiles/:user_id/verify", handler.VerifyProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profiles/invalid-uuid/verify", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserProfileHandler_UnverifyProfile_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserProfileHandler{}
	router.Use(mockAuthProfile)
	router.POST("/profiles/:user_id/unverify", handler.UnverifyProfile)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profiles/invalid-uuid/unverify", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
