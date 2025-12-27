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

	"github.com/basilex/promenade/internal/contexts/shared/country"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*country.Country, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*country.Country), args.Error(1)
}

func (m *MockUseCase) GetByCode(ctx context.Context, code string) (*country.Country, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*country.Country), args.Error(1)
}

func (m *MockUseCase) List(ctx context.Context) ([]*country.Country, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*country.Country), args.Error(1)
}

func (m *MockUseCase) Create(ctx context.Context, c *country.Country) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockUseCase) Update(ctx context.Context, c *country.Country) error {
	args := m.Called(ctx, c)
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

func TestHandler_ListCountries(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.GET("/countries", handler.ListCountries)

	t.Run("success", func(t *testing.T) {
		countries := []*country.Country{
			{ID: uuidv7.New(), Code: "UA", Code3: "UKR", Name: "Ukraine", PhoneCode: "+380", IsActive: true},
			{ID: uuidv7.New(), Code: "US", Code3: "USA", Name: "United States", PhoneCode: "+1", IsActive: true},
		}
		mockUC.On("List", mock.Anything).Return(countries, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/countries", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		
		mockUC.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockUC.On("List", mock.Anything).Return(nil, errors.New("db error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/countries", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_GetCountryByCode(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.GET("/countries/code/:code", handler.GetCountryByCode)

	t.Run("success", func(t *testing.T) {
		c := &country.Country{
			ID:        uuidv7.New(),
			Code:      "UA",
			Code3:     "UKR",
			Name:      "Ukraine",
			PhoneCode: "+380",
			IsActive:  true,
		}
		mockUC.On("GetByCode", mock.Anything, "UA").Return(c, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/countries/code/UA", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		
		mockUC.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockUC.On("GetByCode", mock.Anything, "XX").Return(nil, country.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/countries/code/XX", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_CreateCountry(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.POST("/countries", handler.CreateCountry)

	t.Run("success", func(t *testing.T) {
		reqBody := CreateCountryRequest{
			Code:      "UA",
			Code3:     "UKR",
			Name:      "Ukraine",
			PhoneCode: "+380",
		}
		mockUC.On("Create", mock.Anything, mock.AnythingOfType("*country.Country")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/countries", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		reqBody := CreateCountryRequest{
			Code: "", // Empty code
			Name: "Ukraine",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/countries", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_UpdateCountry(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.PUT("/countries/:id", handler.UpdateCountry)

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		existing := &country.Country{
			ID:        id,
			Code:      "UA",
			Name:      "Ukraine",
			PhoneCode: "+380",
		}
		
		mockUC.On("GetByID", mock.Anything, id).Return(existing, nil).Once()
		mockUC.On("Update", mock.Anything, mock.AnythingOfType("*country.Country")).Return(nil).Once()

		reqBody := UpdateCountryRequest{
			Name:      "Ukraine Updated",
			PhoneCode: "+380",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/countries/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("GetByID", mock.Anything, id).Return(nil, country.ErrNotFound).Once()

		reqBody := UpdateCountryRequest{
			Name: "Test",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/countries/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_DeleteCountry(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.DELETE("/countries/:id", handler.DeleteCountry)

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/countries/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/countries/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(country.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/countries/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}
