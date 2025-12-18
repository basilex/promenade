package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInfoHandler_GetAPIInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewInfoHandler("Promenade API", "1.0.0", "test", "localhost", "8081")
	router := gin.New()
	router.GET("/api", handler.GetAPIInfo)
	req, _ := http.NewRequest("GET", "/api", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response APIVersionsInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Promenade API", response.Service)
	assert.Equal(t, []string{"v1", "v2"}, response.Versions)
	assert.NotNil(t, response.V1)
	assert.NotNil(t, response.V2)
	assert.Contains(t, response.V1.BasePath, "/api/v1")
	assert.Contains(t, response.V1.Documentation, "/api/v1/docs/swagger/index.html")
	assert.Contains(t, response.V1.HealthCheck, "/api/v1/health")
}

func TestInfoHandler_GetV1Info(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewInfoHandler("Promenade API", "1.0.0", "test", "localhost", "8081")
	router := gin.New()
	router.GET("/api/v1", handler.GetV1Info)
	req, _ := http.NewRequest("GET", "/api/v1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response APIVersionInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "v1", response.Version)
	assert.Equal(t, "/api/v1", response.BasePath)
	assert.Contains(t, response.Documentation, "/api/v1/docs/swagger/index.html")
	assert.Contains(t, response.HealthCheck, "/api/v1/health")
}

func TestInfoHandler_GetV2Info(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewInfoHandler("Promenade API", "1.0.0", "test", "localhost", "8081")
	router := gin.New()
	router.GET("/api/v2", handler.GetV2Info)
	req, _ := http.NewRequest("GET", "/api/v2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var response APIVersionInfo
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "v2", response.Version)
	assert.Equal(t, "/api/v2", response.BasePath)
	assert.Contains(t, response.Documentation, "/api/v2/docs/swagger/index.html")
	assert.Contains(t, response.HealthCheck, "/api/v2/health")
}
