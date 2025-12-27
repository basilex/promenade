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

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*currency.Currency, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*currency.Currency), args.Error(1)
}

func (m *MockUseCase) GetByCode(ctx context.Context, code string) (*currency.Currency, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*currency.Currency), args.Error(1)
}

func (m *MockUseCase) List(ctx context.Context) ([]*currency.Currency, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*currency.Currency), args.Error(1)
}

func (m *MockUseCase) Create(ctx context.Context, c *currency.Currency) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockUseCase) Update(ctx context.Context, c *currency.Currency) error {
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

func TestHandler_ListCurrencies(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.GET("/currencies", handler.ListCurrencies)

	t.Run("success", func(t *testing.T) {
		currencies := []*currency.Currency{
			{ID: uuidv7.New(), Code: "USD", Name: "US Dollar", Symbol: "$", DecimalPlaces: 2, IsActive: true},
			{ID: uuidv7.New(), Code: "EUR", Name: "Euro", Symbol: "€", DecimalPlaces: 2, IsActive: true},
		}
		mockUC.On("List", mock.Anything).Return(currencies, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/currencies", nil)
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

		req := httptest.NewRequest(http.MethodGet, "/currencies", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_GetCurrencyByCode(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.GET("/currencies/code/:code", handler.GetCurrencyByCode)

	t.Run("success", func(t *testing.T) {
		c := &currency.Currency{
			ID:            uuidv7.New(),
			Code:          "USD",
			Name:          "US Dollar",
			Symbol:        "$",
			DecimalPlaces: 2,
			IsActive:      true,
		}
		mockUC.On("GetByCode", mock.Anything, "USD").Return(c, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/currencies/code/USD", nil)
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
		mockUC.On("GetByCode", mock.Anything, "XXX").Return(nil, currency.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/currencies/code/XXX", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_CreateCurrency(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.POST("/currencies", handler.CreateCurrency)

	t.Run("success", func(t *testing.T) {
		reqBody := CreateCurrencyRequest{
			Code:          "USD",
			Name:          "US Dollar",
			Symbol:        "$",
			DecimalPlaces: 2,
		}
		mockUC.On("Create", mock.Anything, mock.AnythingOfType("*currency.Currency")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		reqBody := CreateCurrencyRequest{
			Code: "", // Empty code
			Name: "US Dollar",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_UpdateCurrency(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.PUT("/currencies/:id", handler.UpdateCurrency)

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		existing := &currency.Currency{
			ID:     id,
			Code:   "USD",
			Name:   "US Dollar",
			Symbol: "$",
		}

		mockUC.On("GetByID", mock.Anything, id).Return(existing, nil).Once()
		mockUC.On("Update", mock.Anything, mock.AnythingOfType("*currency.Currency")).Return(nil).Once()

		reqBody := UpdateCurrencyRequest{
			Name:   "United States Dollar",
			Symbol: "$",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/currencies/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("GetByID", mock.Anything, id).Return(nil, currency.ErrNotFound).Once()

		reqBody := UpdateCurrencyRequest{
			Name: "Test",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/currencies/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_DeleteCurrency(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.DELETE("/currencies/:id", handler.DeleteCurrency)

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/currencies/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/currencies/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(currency.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/currencies/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}
