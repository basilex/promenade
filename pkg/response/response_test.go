package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter() (*gin.Engine, *gin.Context, *httptest.ResponseRecorder) {
	router := gin.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return router, c, w
}

func TestSuccess(t *testing.T) {
	_, c, w := setupRouter()

	data := map[string]string{"message": "test data"}
	Success(c, data)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.NotNil(t, response.Data)
	assert.Nil(t, response.Error)
}

func TestCreated(t *testing.T) {
	_, c, w := setupRouter()

	data := map[string]string{"id": "123"}
	Created(c, data)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.NotNil(t, response.Data)
}

func TestErrorResponse(t *testing.T) {
	_, c, w := setupRouter()

	ErrorResponse(c, http.StatusBadRequest, "TEST_ERROR", "Test error message")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "TEST_ERROR", response.Error.Code)
	assert.Equal(t, "Test error message", response.Error.Message)
}

func TestBadRequest(t *testing.T) {
	_, c, w := setupRouter()

	BadRequest(c, "Invalid input")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "BAD_REQUEST", response.Error.Code)
	assert.Equal(t, "Invalid input", response.Error.Message)
}

func TestNotFound(t *testing.T) {
	_, c, w := setupRouter()

	NotFound(c, "Resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, "Resource not found", response.Error.Message)
}

func TestInternalError(t *testing.T) {
	_, c, w := setupRouter()

	InternalError(c, "Something went wrong")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	assert.Equal(t, "Something went wrong", response.Error.Message)
}

func TestUnauthorized(t *testing.T) {
	_, c, w := setupRouter()

	Unauthorized(c, "Authentication required")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
	assert.Equal(t, "Authentication required", response.Error.Message)
}

func TestForbidden(t *testing.T) {
	_, c, w := setupRouter()

	Forbidden(c, "Access denied")

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "error", response.Status)
	assert.Equal(t, "FORBIDDEN", response.Error.Code)
	assert.Equal(t, "Access denied", response.Error.Message)
}

func TestResponse_MultipleFields(t *testing.T) {
	_, c, w := setupRouter()

	complexData := map[string]interface{}{
		"id":      123,
		"name":    "Test Item",
		"active":  true,
		"count":   42,
		"nested":  map[string]string{"key": "value"},
		"list":    []string{"a", "b", "c"},
	}

	Success(c, complexData)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)

	// Verify data structure
	dataMap := response.Data.(map[string]interface{})
	assert.Equal(t, float64(123), dataMap["id"])
	assert.Equal(t, "Test Item", dataMap["name"])
	assert.Equal(t, true, dataMap["active"])
}

func TestResponse_NilData(t *testing.T) {
	_, c, w := setupRouter()

	Success(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	// Nil data should be omitted from JSON
	assert.Nil(t, response.Data)
}

func TestResponse_EmptyStringData(t *testing.T) {
	_, c, w := setupRouter()

	Success(c, "")

	assert.Equal(t, http.StatusOK, w.Code)

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "", response.Data)
}

func TestError_Structure(t *testing.T) {
	err := &Error{
		Code:    "TEST_CODE",
		Message: "Test message",
	}

	assert.Equal(t, "TEST_CODE", err.Code)
	assert.Equal(t, "Test message", err.Message)
}

func TestResponse_Structure(t *testing.T) {
	resp := &Response{
		Status:  "success",
		Data:    map[string]string{"key": "value"},
		Error:   nil,
		Message: "Optional message",
	}

	assert.Equal(t, "success", resp.Status)
	assert.NotNil(t, resp.Data)
	assert.Nil(t, resp.Error)
	assert.Equal(t, "Optional message", resp.Message)
}
