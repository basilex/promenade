package timezone_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	timezoneHTTP "github.com/basilex/promenade/internal/contexts/shared/timezone/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockTimezoneUseCase is a mock implementation of timezone.IUseCase for testing
type MockTimezoneUseCase struct {
	GetByIDFunc   func(ctx context.Context, id uuidv7.UUID) (*timezone.Timezone, error)
	GetByNameFunc func(ctx context.Context, name string) (*timezone.Timezone, error)
	ListFunc      func(ctx context.Context) ([]*timezone.Timezone, error)
	CreateFunc    func(ctx context.Context, timezone *timezone.Timezone) error
	UpdateFunc    func(ctx context.Context, timezone *timezone.Timezone) error
	DeleteFunc    func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockTimezoneUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*timezone.Timezone, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockTimezoneUseCase) GetByName(ctx context.Context, name string) (*timezone.Timezone, error) {
	if m.GetByNameFunc != nil {
		return m.GetByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *MockTimezoneUseCase) List(ctx context.Context) ([]*timezone.Timezone, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockTimezoneUseCase) Create(ctx context.Context, tz *timezone.Timezone) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, tz)
	}
	return nil
}

func (m *MockTimezoneUseCase) Update(ctx context.Context, tz *timezone.Timezone) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, tz)
	}
	return nil
}

func (m *MockTimezoneUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// fakeTimezone creates a fake timezone for testing
func fakeTimezone() *timezone.Timezone {
	tz, _ := timezone.NewTimezone("Europe/Kyiv", "EET", 7200)
	return tz
}

func TestTimezoneHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{
		CreateFunc: func(ctx context.Context, tz *timezone.Timezone) error {
			return nil
		},
	}

	handler := timezoneHTTP.NewHandler(mockUC)
	router.POST("/timezones", handler.CreateTimezone)

	body := map[string]any{
		"name":         "Europe/Kyiv",
		"abbreviation": "EET",
		"utc_offset":   "+02:00",
	}

	w := smoke.MakeRequest(t, router, "POST", "/timezones", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestTimezoneHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{}
	handler := timezoneHTTP.NewHandler(mockUC)
	router.POST("/timezones", handler.CreateTimezone)

	body := map[string]any{
		"name": "Europe/Kyiv",
		// Missing required fields
	}

	w := smoke.MakeRequest(t, router, "POST", "/timezones", body)
	smoke.AssertErrorResponse(t, w, 400, "VALIDATION_ERROR")
}

func TestTimezoneHandler_GetByName_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{
		GetByNameFunc: func(ctx context.Context, name string) (*timezone.Timezone, error) {
			return fakeTimezone(), nil
		},
	}

	handler := timezoneHTTP.NewHandler(mockUC)
	router.GET("/timezones/*name", handler.GetTimezoneByName)

	w := smoke.MakeRequest(t, router, "GET", "/timezones/Europe/Kyiv", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestTimezoneHandler_GetByName_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{
		GetByNameFunc: func(ctx context.Context, name string) (*timezone.Timezone, error) {
			return nil, timezone.ErrTimezoneNotFound
		},
	}

	handler := timezoneHTTP.NewHandler(mockUC)
	router.GET("/timezones/*name", handler.GetTimezoneByName)

	w := smoke.MakeRequest(t, router, "GET", "/timezones/Invalid/Timezone", nil)
	smoke.AssertErrorResponse(t, w, 404, "TIMEZONE_NOT_FOUND")
}

func TestTimezoneHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{
		ListFunc: func(ctx context.Context) ([]*timezone.Timezone, error) {
			return []*timezone.Timezone{fakeTimezone()}, nil
		},
	}

	handler := timezoneHTTP.NewHandler(mockUC)
	router.GET("/timezones", handler.ListTimezones)

	w := smoke.MakeRequest(t, router, "GET", "/timezones", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestTimezoneHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{
		ListFunc: func(ctx context.Context) ([]*timezone.Timezone, error) {
			return []*timezone.Timezone{}, nil
		},
	}

	handler := timezoneHTTP.NewHandler(mockUC)
	router.GET("/timezones", handler.ListTimezones)

	w := smoke.MakeRequest(t, router, "GET", "/timezones", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestTimezoneHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*timezone.Timezone, error) {
			return fakeTimezone(), nil
		},
		UpdateFunc: func(ctx context.Context, tz *timezone.Timezone) error {
			return nil
		},
	}

	handler := timezoneHTTP.NewHandler(mockUC)
	router.PUT("/timezones/:id", handler.UpdateTimezone)

	body := map[string]any{
		"abbreviation": "EEST",
		"utc_offset":   "+03:00",
	}

	w := smoke.MakeRequest(t, router, "PUT", "/timezones/"+smoke.FakeUUID(), body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestTimezoneHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := timezoneHTTP.NewHandler(mockUC)
	router.DELETE("/timezones/:id", handler.DeleteTimezone)

	w := smoke.MakeRequest(t, router, "DELETE", "/timezones/"+smoke.FakeUUID(), nil)
	
	// Delete returns 204 No Content
	if w.Code != 204 {
		t.Errorf("expected status code 204, got %d", w.Code)
	}
}

func TestTimezoneHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockTimezoneUseCase{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return timezone.ErrTimezoneNotFound
		},
	}

	handler := timezoneHTTP.NewHandler(mockUC)
	router.DELETE("/timezones/:id", handler.DeleteTimezone)

	w := smoke.MakeRequest(t, router, "DELETE", "/timezones/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 500, "DELETE_ERROR")
}
