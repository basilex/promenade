package timezone_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	timezoneHTTP "github.com/basilex/promenade/internal/contexts/shared/timezone/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockTimezoneUseCase - minimal mock for smoke tests
type MockTimezoneUseCase struct {
	mock.Mock
}

func (m *MockTimezoneUseCase) List(ctx context.Context) ([]*timezone.Timezone, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*timezone.Timezone), args.Error(1)
}

func (m *MockTimezoneUseCase) GetByName(ctx context.Context, name string) (*timezone.Timezone, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*timezone.Timezone), args.Error(1)
}

func (m *MockTimezoneUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*timezone.Timezone, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*timezone.Timezone), args.Error(1)
}

func (m *MockTimezoneUseCase) Create(ctx context.Context, tz *timezone.Timezone) error {
	args := m.Called(ctx, tz)
	return args.Error(0)
}

func (m *MockTimezoneUseCase) Update(ctx context.Context, tz *timezone.Timezone) error {
	args := m.Called(ctx, tz)
	return args.Error(0)
}

func (m *MockTimezoneUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTimezoneRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestTimezoneHandler_Smoke - smoke tests for Timezone handler
func TestTimezoneHandler_Smoke(t *testing.T) {
	mockUC := new(MockTimezoneUseCase)
	handler := timezoneHTTP.NewHandler(mockUC)
	router := setupTimezoneRouter()

	// Register routes
	router.GET("/timezones", handler.ListTimezones)
	router.GET("/timezones/*name", handler.GetTimezoneByName)
	router.POST("/timezones", handler.CreateTimezone)
	router.PUT("/timezones/:id", handler.UpdateTimezone)
	router.DELETE("/timezones/:id", handler.DeleteTimezone)

	t.Run("List returns 200", func(t *testing.T) {
		timezones := []*timezone.Timezone{
			{ID: uuidv7.New(), Name: "UTC", Abbreviation: "UTC", UTCOffset: 0, IsActive: true},
			{ID: uuidv7.New(), Name: "Europe/Kyiv", Abbreviation: "EET", UTCOffset: 7200, IsActive: true},
		}
		mockUC.On("List", mock.Anything).Return(timezones, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/timezones", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "List should return 200")
	})

	t.Run("GetByName returns 200", func(t *testing.T) {
		tz := &timezone.Timezone{
			ID:           uuidv7.New(),
			Name:         "Europe/Kyiv",
			Abbreviation: "EET",
			UTCOffset:    7200,
			IsActive:     true,
		}
		mockUC.On("GetByName", mock.Anything, "Europe/Kyiv").Return(tz, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/timezones/Europe/Kyiv", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GetByName should return 200")
	})

	t.Run("Create returns 201", func(t *testing.T) {
		reqBody := timezoneHTTP.CreateTimezoneRequest{
			Name:         "America/New_York",
			Abbreviation: "EST",
			UTCOffset:    "-05:00",
		}
		mockUC.On("Create", mock.Anything, mock.AnythingOfType("*timezone.Timezone")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/timezones", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Create should return 201")
	})

	t.Run("Update returns 200", func(t *testing.T) {
		id := uuidv7.New()
		existing := &timezone.Timezone{
			ID:           id,
			Name:         "America/New_York",
			Abbreviation: "EST",
			UTCOffset:    -18000,
			IsActive:     true,
		}
		reqBody := timezoneHTTP.UpdateTimezoneRequest{
			Abbreviation: "EST",
			UTCOffset:    "-05:00",
			IsActive:     true,
		}
		mockUC.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existing, nil).Once()
		mockUC.On("Update", mock.Anything, mock.AnythingOfType("*timezone.Timezone")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/timezones/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Update should return 200")
	})

	t.Run("Delete returns 204", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/timezones/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Delete should return 204")
	})
}
