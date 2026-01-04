package country_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/shared/country"
	countryHTTP "github.com/basilex/promenade/internal/contexts/shared/country/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockCountryUseCase is a mock implementation of country.IUseCase for testing
type MockCountryUseCase struct {
	GetByIDFunc   func(ctx context.Context, id uuidv7.UUID) (*country.Country, error)
	GetByCodeFunc func(ctx context.Context, code string) (*country.Country, error)
	ListFunc      func(ctx context.Context) ([]*country.Country, error)
	CreateFunc    func(ctx context.Context, country *country.Country) error
	UpdateFunc    func(ctx context.Context, country *country.Country) error
	DeleteFunc    func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockCountryUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*country.Country, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCountryUseCase) GetByCode(ctx context.Context, code string) (*country.Country, error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, code)
	}
	return nil, nil
}

func (m *MockCountryUseCase) List(ctx context.Context) ([]*country.Country, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockCountryUseCase) Create(ctx context.Context, c *country.Country) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, c)
	}
	return nil
}

func (m *MockCountryUseCase) Update(ctx context.Context, c *country.Country) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, c)
	}
	return nil
}

func (m *MockCountryUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// fakeCountry creates a fake country for testing
func fakeCountry() *country.Country {
	c, _ := country.NewCountry("US", "United States", "+1")
	return c
}

func TestCountryHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{
		CreateFunc: func(ctx context.Context, c *country.Country) error {
			return nil
		},
	}

	handler := countryHTTP.NewHandler(mockUC)
	router.POST("/countries", handler.CreateCountry)

	body := map[string]any{
		"code":       "US",
		"code3":      "USA",
		"name":       "United States",
		"phone_code": "+1",
	}

	w := smoke.MakeRequest(t, router, "POST", "/countries", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestCountryHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{}
	handler := countryHTTP.NewHandler(mockUC)
	router.POST("/countries", handler.CreateCountry)

	body := map[string]any{
		"code": "US",
		// Missing required fields
	}

	w := smoke.MakeRequest(t, router, "POST", "/countries", body)
	smoke.AssertErrorResponse(t, w, 400, "VALIDATION_ERROR")
}

func TestCountryHandler_GetByCode_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{
		GetByCodeFunc: func(ctx context.Context, code string) (*country.Country, error) {
			return fakeCountry(), nil
		},
	}

	handler := countryHTTP.NewHandler(mockUC)
	router.GET("/countries/:code", handler.GetCountryByCode)

	w := smoke.MakeRequest(t, router, "GET", "/countries/US", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCountryHandler_GetByCode_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{
		GetByCodeFunc: func(ctx context.Context, code string) (*country.Country, error) {
			return nil, country.ErrCountryNotFound
		},
	}

	handler := countryHTTP.NewHandler(mockUC)
	router.GET("/countries/:code", handler.GetCountryByCode)

	w := smoke.MakeRequest(t, router, "GET", "/countries/XX", nil)
	smoke.AssertErrorResponse(t, w, 404, "COUNTRY_NOT_FOUND")
}

func TestCountryHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{
		ListFunc: func(ctx context.Context) ([]*country.Country, error) {
			return []*country.Country{fakeCountry()}, nil
		},
	}

	handler := countryHTTP.NewHandler(mockUC)
	router.GET("/countries", handler.ListCountries)

	w := smoke.MakeRequest(t, router, "GET", "/countries", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCountryHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{
		ListFunc: func(ctx context.Context) ([]*country.Country, error) {
			return []*country.Country{}, nil
		},
	}

	handler := countryHTTP.NewHandler(mockUC)
	router.GET("/countries", handler.ListCountries)

	w := smoke.MakeRequest(t, router, "GET", "/countries", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCountryHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*country.Country, error) {
			return fakeCountry(), nil
		},
		UpdateFunc: func(ctx context.Context, c *country.Country) error {
			return nil
		},
	}

	handler := countryHTTP.NewHandler(mockUC)
	router.PUT("/countries/:id", handler.UpdateCountry)

	body := map[string]any{
		"name":       "United States of America",
		"phone_code": "+1",
	}

	w := smoke.MakeRequest(t, router, "PUT", "/countries/"+smoke.FakeUUID(), body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCountryHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := countryHTTP.NewHandler(mockUC)
	router.DELETE("/countries/:id", handler.DeleteCountry)

	w := smoke.MakeRequest(t, router, "DELETE", "/countries/"+smoke.FakeUUID(), nil)
	
	// Delete returns 204 No Content
	if w.Code != 204 {
		t.Errorf("expected status code 204, got %d", w.Code)
	}
}

func TestCountryHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCountryUseCase{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return country.ErrCountryNotFound
		},
	}

	handler := countryHTTP.NewHandler(mockUC)
	router.DELETE("/countries/:id", handler.DeleteCountry)

	w := smoke.MakeRequest(t, router, "DELETE", "/countries/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 500, "DELETE_ERROR")
}
