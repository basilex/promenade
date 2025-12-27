package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*timezone.Timezone, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*timezone.Timezone), args.Error(1)
}

func (m *MockUseCase) GetByName(ctx context.Context, name string) (*timezone.Timezone, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*timezone.Timezone), args.Error(1)
}

func (m *MockUseCase) List(ctx context.Context) ([]*timezone.Timezone, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*timezone.Timezone), args.Error(1)
}

func (m *MockUseCase) Create(ctx context.Context, tz *timezone.Timezone) error {
	args := m.Called(ctx, tz)
	return args.Error(0)
}

func (m *MockUseCase) Update(ctx context.Context, tz *timezone.Timezone) error {
	args := m.Called(ctx, tz)
	return args.Error(0)
}

func (m *MockUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHandler_ListTimezones(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.GET("/timezones", handler.ListTimezones)

	t.Run("success", func(t *testing.T) {
		timezones := []*timezone.Timezone{
			{ID: uuidv7.New(), Name: "Europe/Kyiv", Abbreviation: "EET", UTCOffset: 7200},        // +02:00
			{ID: uuidv7.New(), Name: "America/New_York", Abbreviation: "EST", UTCOffset: -18000}, // -05:00
		}
		mockUC.On("List", mock.Anything).Return(timezones, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/timezones", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockUC.On("List", mock.Anything).Return(nil, errors.New("db error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/timezones", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_GetTimezoneByName(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.GET("/timezones/*name", handler.GetTimezoneByName) // * to capture full path

	t.Run("success", func(t *testing.T) {
		tz := &timezone.Timezone{ID: uuidv7.New(), Name: "Europe/Kyiv", Abbreviation: "EET", UTCOffset: 7200} // +02:00
		mockUC.On("GetByName", mock.Anything, "Europe/Kyiv").Return(tz, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/timezones/Europe/Kyiv", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockUC.On("GetByName", mock.Anything, "Europe/Unknown").Return(nil, timezone.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/timezones/Europe/Unknown", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_CreateTimezone(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.POST("/timezones", handler.CreateTimezone)

	t.Run("success", func(t *testing.T) {
		mockUC.On("Create", mock.Anything, mock.AnythingOfType("*timezone.Timezone")).Return(nil).Once()

		reqBody := CreateTimezoneRequest{
			Name:         "Europe/Kyiv",
			Abbreviation: "EET",
			UTCOffset:    "+02:00",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/timezones", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		reqBody := CreateTimezoneRequest{
			Name:      "", // Empty name
			UTCOffset: "+02:00",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/timezones", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_UpdateTimezone(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.PUT("/timezones/:id", handler.UpdateTimezone)

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		existing := &timezone.Timezone{ID: id, Name: "Europe/Kyiv", Abbreviation: "EET", UTCOffset: 7200} // +02:00
		mockUC.On("GetByID", mock.Anything, id).Return(existing, nil).Once()
		mockUC.On("Update", mock.Anything, existing).Return(nil).Once()

		reqBody := UpdateTimezoneRequest{
			Abbreviation: "EEST",
			UTCOffset:    "+03:00",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/timezones/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		reqBody := UpdateTimezoneRequest{Abbreviation: "EET"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/timezones/invalid-uuid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_DeleteTimezone(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.DELETE("/timezones/:id", handler.DeleteTimezone)

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/timezones/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/timezones/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestParseUTCOffset tests the parseUTCOffset helper function for error cases
func TestParseUTCOffset(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{
			name:        "valid positive offset",
			input:       "+02:00",
			expected:    7200,
			expectError: false,
		},
		{
			name:        "valid negative offset",
			input:       "-05:00",
			expected:    -18000,
			expectError: false,
		},
		{
			name:        "valid UTC",
			input:       "+00:00",
			expected:    0,
			expectError: false,
		},
		{
			name:        "valid with minutes",
			input:       "+05:30",
			expected:    19800,
			expectError: false,
		},
		{
			name:        "invalid format - too short",
			input:       "+02",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid format - no colon",
			input:       "+0200",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid format - letters",
			input:       "+ab:cd",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid format - no sign",
			input:       "02:00",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid hours - out of range",
			input:       "+25:00",
			expected:    0,
			expectError: true,
		},
		{
			name:        "invalid minutes - out of range",
			input:       "+02:60",
			expected:    0,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseUTCOffset(tt.input)
			if tt.expectError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tt.input)
				assert.Equal(t, tt.expected, result, "Unexpected result for input: %s", tt.input)
			}
		})
	}
}

// TestCreateTimezone_ErrorCases tests error scenarios for CreateTimezone
func TestCreateTimezone_ErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid JSON", func(t *testing.T) {
		mockUC := new(MockUseCase)
		handler := NewHandler(mockUC)

		router := gin.New()
		router.POST("/timezones", handler.CreateTimezone)

		// Invalid JSON body
		req := httptest.NewRequest(http.MethodPost, "/timezones", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid UTC offset format", func(t *testing.T) {
		mockUC := new(MockUseCase)
		handler := NewHandler(mockUC)

		router := gin.New()
		router.POST("/timezones", handler.CreateTimezone)

		reqBody := CreateTimezoneRequest{
			Name:         "Test Timezone",
			Abbreviation: "TST",
			UTCOffset:    "invalid",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/timezones", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase create error", func(t *testing.T) {
		mockUC := new(MockUseCase)
		handler := NewHandler(mockUC)

		router := gin.New()
		router.POST("/timezones", handler.CreateTimezone)

		mockUC.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))

		reqBody := CreateTimezoneRequest{
			Name:         "Test Timezone",
			Abbreviation: "TST",
			UTCOffset:    "+02:00",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/timezones", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}

// TestUpdateTimezone_ErrorCases tests error scenarios for UpdateTimezone
func TestUpdateTimezone_ErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid JSON", func(t *testing.T) {
		mockUC := new(MockUseCase)
		handler := NewHandler(mockUC)

		router := gin.New()
		router.PUT("/timezones/:id", handler.UpdateTimezone)

		id := uuidv7.New().String()
		req := httptest.NewRequest(http.MethodPut, "/timezones/"+id, bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid UTC offset format", func(t *testing.T) {
		mockUC := new(MockUseCase)
		handler := NewHandler(mockUC)

		router := gin.New()
		router.PUT("/timezones/:id", handler.UpdateTimezone)

		id := uuidv7.New()
		existingTz := &timezone.Timezone{
			ID:           id,
			Name:         "Europe/Kiev",
			Abbreviation: "EET",
			UTCOffset:    7200,
		}

		mockUC.On("GetByID", mock.Anything, id).Return(existingTz, nil)

		reqBody := UpdateTimezoneRequest{
			Abbreviation: "EET",
			UTCOffset:    "invalid",
			IsActive:     true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/timezones/"+id.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("timezone not found", func(t *testing.T) {
		mockUC := new(MockUseCase)
		handler := NewHandler(mockUC)

		router := gin.New()
		router.PUT("/timezones/:id", handler.UpdateTimezone)

		id := uuidv7.New()
		mockUC.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

		reqBody := UpdateTimezoneRequest{
			Abbreviation: "EET",
			UTCOffset:    "+02:00",
			IsActive:     true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/timezones/"+id.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("usecase update error", func(t *testing.T) {
		mockUC := new(MockUseCase)
		handler := NewHandler(mockUC)

		router := gin.New()
		router.PUT("/timezones/:id", handler.UpdateTimezone)

		id := uuidv7.New()
		existingTz := &timezone.Timezone{
			ID:           id,
			Name:         "Europe/Kiev",
			Abbreviation: "EET",
			UTCOffset:    7200,
		}

		mockUC.On("GetByID", mock.Anything, id).Return(existingTz, nil)
		mockUC.On("Update", mock.Anything, mock.Anything).Return(errors.New("database error"))

		reqBody := UpdateTimezoneRequest{
			Abbreviation: "EEST",
			UTCOffset:    "+03:00",
			IsActive:     true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/timezones/"+id.String(), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}

// TestDeleteTimezone_ErrorCases tests error scenarios for DeleteTimezone
func TestDeleteTimezone_ErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("usecase delete error", func(t *testing.T) {
		mockUC := new(MockUseCase)
		handler := NewHandler(mockUC)

		router := gin.New()
		router.DELETE("/timezones/:id", handler.DeleteTimezone)

		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(errors.New("database error"))

		req := httptest.NewRequest(http.MethodDelete, "/timezones/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}
