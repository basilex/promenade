package smoke

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// SetupRouter creates a test router with Gin in test mode
func SetupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Add recovery middleware to catch panics
	router.Use(gin.Recovery())
	return router
}

// MakeRequest makes an HTTP request to the test router
func MakeRequest(t *testing.T, router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err, "failed to marshal request body")
		reqBody = bytes.NewBuffer(bodyBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// MakeRequestWithHeaders makes an HTTP request with custom headers
func MakeRequestWithHeaders(t *testing.T, router *gin.Engine, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	w := MakeRequest(t, router, method, path, body)
	req := httptest.NewRequest(method, path, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return w
}

// AssertSuccessResponse checks that response has status "success"
func AssertSuccessResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedCode int) {
	t.Helper()

	assert.Equal(t, expectedCode, resp.Code, "unexpected HTTP status code")

	var respBody map[string]any
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	require.NoError(t, err, "failed to parse response JSON")

	status, ok := respBody["status"].(string)
	require.True(t, ok, "response missing 'status' field")
	assert.Equal(t, "success", status, "expected success response")
}

// AssertErrorResponse checks that response has status "error" and specific error code
func AssertErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedHTTPCode int, expectedErrorCode string) {
	t.Helper()

	assert.Equal(t, expectedHTTPCode, resp.Code, "unexpected HTTP status code")

	var respBody struct {
		Status string `json:"status"`
		Error  struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	err := json.Unmarshal(resp.Body.Bytes(), &respBody)
	require.NoError(t, err, "failed to parse response JSON")

	assert.Equal(t, "error", respBody.Status, "expected error response")
	assert.Equal(t, expectedErrorCode, respBody.Error.Code, "unexpected error code")
}

// FakeUUID returns a valid UUID v7 string for testing
func FakeUUID() string {
	return uuidv7.New().String()
}

// FakeInvalidUUID returns an invalid UUID string for testing
func FakeInvalidUUID() string {
	return "invalid-uuid"
}
