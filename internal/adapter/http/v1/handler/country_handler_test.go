package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCountryHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CountryHandler{}
	router.POST("/countries", handler.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/countries", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestCountryHandler_GetByID_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CountryHandler{}
	router.GET("/countries/:id", handler.GetByID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/countries/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "error")
}

func TestCountryHandler_List_InvalidRegion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CountryHandler{}
	router.GET("/countries", handler.List)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/countries?region=invalid_region", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)
	// Response contains "message" field, not "error"
	assert.Contains(t, response, "message")
	assert.False(t, response["success"].(bool))
}

func TestCountryHandler_Delete_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CountryHandler{}
	router.DELETE("/countries/:id", handler.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/countries/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
