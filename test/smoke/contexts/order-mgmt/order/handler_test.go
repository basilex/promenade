package order_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/order"
	orderHTTP "github.com/basilex/promenade/internal/contexts/order-mgmt/order/adapter/http"
	orderAggregate "github.com/basilex/promenade/internal/contexts/order-mgmt/order/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/smoke"
)

// MockOrderUseCase is a mock implementation of order.IUseCase for testing
type MockOrderUseCase struct {
	CreateOrderFunc             func(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, currency string) (*orderAggregate.Order, error)
	GetOrderFunc                func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error)
	GetOrderByNumberFunc        func(ctx context.Context, orderNumber string) (*orderAggregate.Order, error)
	AddOrderLineFunc            func(ctx context.Context, orderID, productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) (*orderAggregate.Order, error)
	RemoveOrderLineFunc         func(ctx context.Context, orderID, lineID uuidv7.UUID) (*orderAggregate.Order, error)
	UpdateOrderLineQuantityFunc func(ctx context.Context, orderID, lineID uuidv7.UUID, quantity int) (*orderAggregate.Order, error)
	ConfirmOrderFunc            func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error)
	StartProcessingFunc         func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error)
	MarkFulfilledFunc           func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error)
	CancelOrderFunc             func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error)
	ListOrdersFunc              func(ctx context.Context, page, pageSize int) ([]*orderAggregate.Order, int64, error)
	ListOrdersByCustomerFunc    func(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*orderAggregate.Order, int64, error)
	ListOrdersByStatusFunc      func(ctx context.Context, status orderAggregate.OrderStatus, page, pageSize int) ([]*orderAggregate.Order, int64, error)
}

func (m *MockOrderUseCase) CreateOrder(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, currency string) (*orderAggregate.Order, error) {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, customerID, companyID, currency)
	}
	return nil, nil
}

func (m *MockOrderUseCase) GetOrder(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
	if m.GetOrderFunc != nil {
		return m.GetOrderFunc(ctx, orderID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) GetOrderByNumber(ctx context.Context, orderNumber string) (*orderAggregate.Order, error) {
	if m.GetOrderByNumberFunc != nil {
		return m.GetOrderByNumberFunc(ctx, orderNumber)
	}
	return nil, nil
}

func (m *MockOrderUseCase) AddOrderLine(ctx context.Context, orderID, productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) (*orderAggregate.Order, error) {
	if m.AddOrderLineFunc != nil {
		return m.AddOrderLineFunc(ctx, orderID, productID, quantity, unitPrice)
	}
	return nil, nil
}

func (m *MockOrderUseCase) RemoveOrderLine(ctx context.Context, orderID, lineID uuidv7.UUID) (*orderAggregate.Order, error) {
	if m.RemoveOrderLineFunc != nil {
		return m.RemoveOrderLineFunc(ctx, orderID, lineID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) UpdateOrderLineQuantity(ctx context.Context, orderID, lineID uuidv7.UUID, quantity int) (*orderAggregate.Order, error) {
	if m.UpdateOrderLineQuantityFunc != nil {
		return m.UpdateOrderLineQuantityFunc(ctx, orderID, lineID, quantity)
	}
	return nil, nil
}

func (m *MockOrderUseCase) ConfirmOrder(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
	if m.ConfirmOrderFunc != nil {
		return m.ConfirmOrderFunc(ctx, orderID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) StartProcessing(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
	if m.StartProcessingFunc != nil {
		return m.StartProcessingFunc(ctx, orderID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) MarkFulfilled(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
	if m.MarkFulfilledFunc != nil {
		return m.MarkFulfilledFunc(ctx, orderID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) CancelOrder(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
	if m.CancelOrderFunc != nil {
		return m.CancelOrderFunc(ctx, orderID)
	}
	return nil, nil
}

func (m *MockOrderUseCase) ListOrders(ctx context.Context, page, pageSize int) ([]*orderAggregate.Order, int64, error) {
	if m.ListOrdersFunc != nil {
		return m.ListOrdersFunc(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockOrderUseCase) ListOrdersByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*orderAggregate.Order, int64, error) {
	if m.ListOrdersByCustomerFunc != nil {
		return m.ListOrdersByCustomerFunc(ctx, customerID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockOrderUseCase) ListOrdersByStatus(ctx context.Context, status orderAggregate.OrderStatus, page, pageSize int) ([]*orderAggregate.Order, int64, error) {
	if m.ListOrdersByStatusFunc != nil {
		return m.ListOrdersByStatusFunc(ctx, status, page, pageSize)
	}
	return nil, 0, nil
}

// fakeOrder creates a fake order for testing
func fakeOrder() *orderAggregate.Order {
	customerID := uuidv7.New()
	money, _ := valueobject.NewMoney(10000, "USD")
	o, _ := orderAggregate.NewOrder(customerID, "USD")
	productID := uuidv7.New()
	_ = o.AddLine(productID, 2, money)
	return o
}

func TestOrderHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		CreateOrderFunc: func(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, currency string) (*orderAggregate.Order, error) {
			return fakeOrder(), nil
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.POST("/orders", handler.Create)

	body := map[string]any{
		"customer_id": smoke.FakeUUID(),
		"currency":    "USD",
	}

	w := smoke.MakeRequest(t, router, "POST", "/orders", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestOrderHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{}
	handler := orderHTTP.NewOrderHandler(mockUC)
	router.POST("/orders", handler.Create)

	body := map[string]any{
		"currency": "USD",
		// Missing customer_id
	}

	w := smoke.MakeRequest(t, router, "POST", "/orders", body)
	smoke.AssertErrorResponse(t, w, 400, "BAD_REQUEST")
}

func TestOrderHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		GetOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
			return fakeOrder(), nil
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.GET("/orders/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/orders/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestOrderHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		GetOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
			return nil, order.ErrOrderNotFound
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.GET("/orders/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/orders/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}

func TestOrderHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		ListOrdersFunc: func(ctx context.Context, page, pageSize int) ([]*orderAggregate.Order, int64, error) {
			return []*orderAggregate.Order{fakeOrder()}, 1, nil
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.GET("/orders", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/orders?page=1&page_size=10", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestOrderHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		ListOrdersFunc: func(ctx context.Context, page, pageSize int) ([]*orderAggregate.Order, int64, error) {
			return []*orderAggregate.Order{}, 0, nil
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.GET("/orders", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/orders?page=1&page_size=10", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestOrderHandler_Confirm_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		ConfirmOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
			return fakeOrder(), nil
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.POST("/orders/:id/confirm", handler.Confirm)

	w := smoke.MakeRequest(t, router, "POST", "/orders/"+smoke.FakeUUID()+"/confirm", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestOrderHandler_Confirm_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		ConfirmOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
			return nil, order.ErrOrderNotFound
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.POST("/orders/:id/confirm", handler.Confirm)

	w := smoke.MakeRequest(t, router, "POST", "/orders/"+smoke.FakeUUID()+"/confirm", nil)
	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}

func TestOrderHandler_Cancel_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		CancelOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
			return fakeOrder(), nil
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.POST("/orders/:id/cancel", handler.Cancel)

	w := smoke.MakeRequest(t, router, "POST", "/orders/"+smoke.FakeUUID()+"/cancel", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestOrderHandler_Cancel_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockOrderUseCase{
		CancelOrderFunc: func(ctx context.Context, orderID uuidv7.UUID) (*orderAggregate.Order, error) {
			return nil, order.ErrOrderNotFound
		},
	}

	handler := orderHTTP.NewOrderHandler(mockUC)
	router.POST("/orders/:id/cancel", handler.Cancel)

	w := smoke.MakeRequest(t, router, "POST", "/orders/"+smoke.FakeUUID()+"/cancel", nil)
	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}
