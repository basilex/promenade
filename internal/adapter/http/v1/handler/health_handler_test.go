package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_HealthCheck(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewHealthHandler()

	t.Run("returns ok status", func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		c.Request = req

		// Act
		handler.HealthCheck(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))

		data := response["data"].(map[string]any)
		assert.Equal(t, "ok", data["status"])
		assert.Equal(t, "promenade", data["service"])
		assert.NotZero(t, data["time"])
	})

	t.Run("response contains required fields", func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		c.Request = req

		// Act
		handler.HealthCheck(c)

		// Assert
		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check response structure
		assert.Contains(t, response, "success")
		assert.Contains(t, response, "data")

		data := response["data"].(map[string]any)
		assert.Contains(t, data, "status")
		assert.Contains(t, data, "service")
		assert.Contains(t, data, "time")
	})

	t.Run("response is valid JSON", func(t *testing.T) {
		// Arrange
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		c.Request = req

		// Act
		handler.HealthCheck(c)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

		// Verify JSON is valid
		var response any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
	})
}
