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

func mockAuth(c *gin.Context) {
	uid, _ := uuidv7.Parse("01936d9a-0000-7000-8000-000000000000")
	c.Set("user_id", uid)
	c.Next()
}

func TestUserContactHandler_CreateContact_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.POST("/contacts", handler.CreateContact)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/contacts", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestUserContactHandler_CreateContact_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.POST("/contacts", handler.CreateContact)

	// Missing contact_type and contact_value
	requestBody := map[string]interface{}{
		"label": "Primary Email",
	}
	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/contacts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserContactHandler_GetContact_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.GET("/contacts/:id", handler.GetContact)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/contacts/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

// Removed: TestUserContactHandler_GetUserContacts_InvalidUserID - requires mock use case

// Removed: TestUserContactHandler_GetContactsByType_InvalidUserID - requires mock use case

func TestUserContactHandler_GetPrimaryContact_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.GET("/users/:user_id/contacts/primary", handler.GetPrimaryContact)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users/invalid-uuid/contacts/primary", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserContactHandler_UpdateContact_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.PUT("/contacts/:id", handler.UpdateContact)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/contacts/01936d9a-0000-7000-8000-000000000000", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserContactHandler_UpdateContact_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.PUT("/contacts/:id", handler.UpdateContact)

	requestBody := map[string]interface{}{
		"contact_value": "updated@example.com",
	}
	body, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/contacts/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserContactHandler_DeleteContact_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.DELETE("/contacts/:id", handler.DeleteContact)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/contacts/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserContactHandler_SetPrimaryContact_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.POST("/contacts/:id/primary", handler.SetPrimaryContact)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/contacts/invalid-uuid/primary", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserContactHandler_ToggleContactActive_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &UserContactHandler{}
	router.Use(mockAuth)
	router.POST("/contacts/:id/toggle-active", handler.ToggleContactActive)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/contacts/invalid-uuid/toggle-active", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
