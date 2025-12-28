package currency_test

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

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	currencyHTTP "github.com/basilex/promenade/internal/contexts/shared/currency/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockCurrencyUseCase - minimal mock for smoke tests
type MockCurrencyUseCase struct {
	mock.Mock
}

func (m *MockCurrencyUseCase) List(ctx context.Context) ([]*currency.Currency, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*currency.Currency), args.Error(1)
}

func (m *MockCurrencyUseCase) GetByCode(ctx context.Context, code string) (*currency.Currency, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*currency.Currency), args.Error(1)
}

func (m *MockCurrencyUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*currency.Currency, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*currency.Currency), args.Error(1)
}

func (m *MockCurrencyUseCase) Create(ctx context.Context, c *currency.Currency) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCurrencyUseCase) Update(ctx context.Context, c *currency.Currency) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockCurrencyUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupCurrencyRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestCurrencyHandler_Smoke - smoke tests for Currency handler
func TestCurrencyHandler_Smoke(t *testing.T) {
	mockUC := new(MockCurrencyUseCase)
	handler := currencyHTTP.NewHandler(mockUC)
	router := setupCurrencyRouter()

	// Register routes
	router.GET("/currencies", handler.ListCurrencies)
	router.GET("/currencies/code/:code", handler.GetCurrencyByCode)
	router.POST("/currencies", handler.CreateCurrency)
	router.PUT("/currencies/:id", handler.UpdateCurrency)
	router.DELETE("/currencies/:id", handler.DeleteCurrency)

	t.Run("List returns 200", func(t *testing.T) {
		currencies := []*currency.Currency{
			{ID: uuidv7.New(), Code: "USD", Name: "US Dollar", Symbol: "$", DecimalPlaces: 2, IsActive: true},
			{ID: uuidv7.New(), Code: "EUR", Name: "Euro", Symbol: "€", DecimalPlaces: 2, IsActive: true},
		}
		mockUC.On("List", mock.Anything).Return(currencies, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/currencies", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "List should return 200")
	})

	t.Run("GetByCode returns 200", func(t *testing.T) {
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

		assert.Equal(t, http.StatusOK, w.Code, "GetByCode should return 200")
	})

	t.Run("Create returns 201", func(t *testing.T) {
		reqBody := currencyHTTP.CreateCurrencyRequest{
			Code:          "UAH",
			Name:          "Ukrainian Hryvnia",
			Symbol:        "₴",
			DecimalPlaces: 2,
		}
		mockUC.On("Create", mock.Anything, mock.AnythingOfType("*currency.Currency")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Create should return 201")
	})

	t.Run("Update returns 200", func(t *testing.T) {
		id := uuidv7.New()
		existing := &currency.Currency{
			ID:       id,
			Code:     "UAH",
			Name:     "Ukrainian Hryvnia",
			IsActive: true,
		}
		reqBody := currencyHTTP.UpdateCurrencyRequest{
			Name:          "Ukrainian Hryvnia Updated",
			Symbol:        "₴",
			DecimalPlaces: 2,
			IsActive:      true,
		}
		mockUC.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existing, nil).Once()
		mockUC.On("Update", mock.Anything, mock.AnythingOfType("*currency.Currency")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/currencies/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Update should return 200")
	})

	t.Run("Delete returns 204", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/currencies/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Delete should return 204")
	})
}

// Helper function
func intPtr(i int) *int {
	return &i
}
