package currency_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	currencyHTTP "github.com/basilex/promenade/internal/contexts/shared/currency/adapter/http"
	currencyAggregate "github.com/basilex/promenade/internal/contexts/shared/currency/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockCurrencyUseCase is a mock implementation of currency.IUseCase for testing
type MockCurrencyUseCase struct {
	GetByIDFunc   func(ctx context.Context, id uuidv7.UUID) (*currencyAggregate.Currency, error)
	GetByCodeFunc func(ctx context.Context, code string) (*currencyAggregate.Currency, error)
	ListFunc      func(ctx context.Context) ([]*currencyAggregate.Currency, error)
	CreateFunc    func(ctx context.Context, currency *currencyAggregate.Currency) error
	UpdateFunc    func(ctx context.Context, currency *currencyAggregate.Currency) error
	DeleteFunc    func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockCurrencyUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*currencyAggregate.Currency, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockCurrencyUseCase) GetByCode(ctx context.Context, code string) (*currencyAggregate.Currency, error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, code)
	}
	return nil, nil
}

func (m *MockCurrencyUseCase) List(ctx context.Context) ([]*currencyAggregate.Currency, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockCurrencyUseCase) Create(ctx context.Context, c *currencyAggregate.Currency) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, c)
	}
	return nil
}

func (m *MockCurrencyUseCase) Update(ctx context.Context, c *currencyAggregate.Currency) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, c)
	}
	return nil
}

func (m *MockCurrencyUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// fakeCurrency creates a fake currency for testing
func fakeCurrency() *currencyAggregate.Currency {
	c, _ := currencyAggregate.NewCurrency("USD", "US Dollar", "$", 2)
	return c
}

func TestCurrencyHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{
		CreateFunc: func(ctx context.Context, c *currencyAggregate.Currency) error {
			return nil
		},
	}

	handler := currencyHTTP.NewHandler(mockUC)
	router.POST("/currencies", handler.CreateCurrency)

	body := map[string]any{
		"code":           "USD",
		"name":           "US Dollar",
		"symbol":         "$",
		"decimal_places": 2,
	}

	w := smoke.MakeRequest(t, router, "POST", "/currencies", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestCurrencyHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{}
	handler := currencyHTTP.NewHandler(mockUC)
	router.POST("/currencies", handler.CreateCurrency)

	body := map[string]any{
		"code": "USD",
		// Missing required fields
	}

	w := smoke.MakeRequest(t, router, "POST", "/currencies", body)
	smoke.AssertErrorResponse(t, w, 400, "VALIDATION_ERROR")
}

func TestCurrencyHandler_GetByCode_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{
		GetByCodeFunc: func(ctx context.Context, code string) (*currencyAggregate.Currency, error) {
			return fakeCurrency(), nil
		},
	}

	handler := currencyHTTP.NewHandler(mockUC)
	router.GET("/currencies/:code", handler.GetCurrencyByCode)

	w := smoke.MakeRequest(t, router, "GET", "/currencies/USD", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCurrencyHandler_GetByCode_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{
		GetByCodeFunc: func(ctx context.Context, code string) (*currencyAggregate.Currency, error) {
			return nil, currency.ErrCurrencyNotFound
		},
	}

	handler := currencyHTTP.NewHandler(mockUC)
	router.GET("/currencies/:code", handler.GetCurrencyByCode)

	w := smoke.MakeRequest(t, router, "GET", "/currencies/XXX", nil)
	smoke.AssertErrorResponse(t, w, 404, "CURRENCY_NOT_FOUND")
}

func TestCurrencyHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{
		ListFunc: func(ctx context.Context) ([]*currencyAggregate.Currency, error) {
			return []*currencyAggregate.Currency{fakeCurrency()}, nil
		},
	}

	handler := currencyHTTP.NewHandler(mockUC)
	router.GET("/currencies", handler.ListCurrencies)

	w := smoke.MakeRequest(t, router, "GET", "/currencies", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCurrencyHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{
		ListFunc: func(ctx context.Context) ([]*currencyAggregate.Currency, error) {
			return []*currencyAggregate.Currency{}, nil
		},
	}

	handler := currencyHTTP.NewHandler(mockUC)
	router.GET("/currencies", handler.ListCurrencies)

	w := smoke.MakeRequest(t, router, "GET", "/currencies", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCurrencyHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*currencyAggregate.Currency, error) {
			return fakeCurrency(), nil
		},
		UpdateFunc: func(ctx context.Context, c *currencyAggregate.Currency) error {
			return nil
		},
	}

	handler := currencyHTTP.NewHandler(mockUC)
	router.PUT("/currencies/:id", handler.UpdateCurrency)

	body := map[string]any{
		"name":   "United States Dollar",
		"symbol": "$",
	}

	w := smoke.MakeRequest(t, router, "PUT", "/currencies/"+smoke.FakeUUID(), body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestCurrencyHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := currencyHTTP.NewHandler(mockUC)
	router.DELETE("/currencies/:id", handler.DeleteCurrency)

	w := smoke.MakeRequest(t, router, "DELETE", "/currencies/"+smoke.FakeUUID(), nil)

	// Delete returns 204 No Content
	if w.Code != 204 {
		t.Errorf("expected status code 204, got %d", w.Code)
	}
}

func TestCurrencyHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockCurrencyUseCase{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return currency.ErrCurrencyNotFound
		},
	}

	handler := currencyHTTP.NewHandler(mockUC)
	router.DELETE("/currencies/:id", handler.DeleteCurrency)

	w := smoke.MakeRequest(t, router, "DELETE", "/currencies/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 500, "DELETE_ERROR")
}
