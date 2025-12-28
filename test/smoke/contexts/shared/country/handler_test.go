package country_test

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

	"github.com/basilex/promenade/internal/contexts/shared/country"
	countryHTTP "github.com/basilex/promenade/internal/contexts/shared/country/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockCountryUseCase - minimal mock for smoke tests
type MockCountryUseCase struct {
	mock.Mock
}

func (m *MockCountryUseCase) List(ctx context.Context) ([]*country.Country, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*country.Country), args.Error(1)
}

func (m *MockCountryUseCase) GetByCode(ctx context.Context, code string) (*country.Country, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*country.Country), args.Error(1)
}

func (m *MockCountryUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*country.Country, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*country.Country), args.Error(1)
}

func (m *MockCountryUseCase) Create(ctx context.Context, c *country.Country) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCountryUseCase) Update(ctx context.Context, c *country.Country) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCountryUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupCountryRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestCountryHandler_Smoke - smoke tests for Country handler
func TestCountryHandler_Smoke(t *testing.T) {
	mockUC := new(MockCountryUseCase)
	handler := countryHTTP.NewHandler(mockUC)
	router := setupCountryRouter()

	// Register routes
	router.GET("/countries", handler.ListCountries)
	router.GET("/countries/code/:code", handler.GetCountryByCode)
	router.POST("/countries", handler.CreateCountry)
	router.PUT("/countries/:id", handler.UpdateCountry)
	router.DELETE("/countries/:id", handler.DeleteCountry)

	t.Run("List returns 200", func(t *testing.T) {
		countries := []*country.Country{
			{ID: uuidv7.New(), Code: "US", Name: "United States", PhoneCode: "+1", IsActive: true},
			{ID: uuidv7.New(), Code: "UA", Name: "Ukraine", PhoneCode: "+380", IsActive: true},
		}
		mockUC.On("List", mock.Anything).Return(countries, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/countries", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "List should return 200")
	})

	t.Run("GetByCode returns 200", func(t *testing.T) {
		c := &country.Country{
			ID:        uuidv7.New(),
			Code:      "UA",
			Name:      "Ukraine",
			PhoneCode: "+380",
			IsActive:  true,
		}
		mockUC.On("GetByCode", mock.Anything, "UA").Return(c, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/countries/code/UA", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GetByCode should return 200")
	})

	t.Run("Create returns 201", func(t *testing.T) {
		reqBody := countryHTTP.CreateCountryRequest{
			Code:      "PL",
			Name:      "Poland",
			PhoneCode: "+48",
		}
		mockUC.On("Create", mock.Anything, mock.AnythingOfType("*country.Country")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/countries", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Create should return 201")
	})

	t.Run("Update returns 200", func(t *testing.T) {
		id := uuidv7.New()
		existing := &country.Country{
			ID:       id,
			Code:     "PL",
			Name:     "Poland",
			IsActive: true,
		}
		reqBody := countryHTTP.UpdateCountryRequest{
			Name:      "Poland Updated",
			PhoneCode: "+48",
			IsActive:  true,
		}
		mockUC.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existing, nil).Once()
		mockUC.On("Update", mock.Anything, mock.AnythingOfType("*country.Country")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/countries/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Update should return 200")
	})

	t.Run("Delete returns 204", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/countries/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Delete should return 204")
	})
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
