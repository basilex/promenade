package usecase

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/order/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of IOrderRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, order *aggregate.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregate.Order), args.Error(1)
}

func (m *MockRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*aggregate.Order, error) {
	args := m.Called(ctx, orderNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregate.Order), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, order *aggregate.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListByCustomerID(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Order, int64, error) {
	args := m.Called(ctx, customerID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*aggregate.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) ListByStatus(ctx context.Context, status aggregate.OrderStatus, page, pageSize int) ([]*aggregate.Order, int64, error) {
	args := m.Called(ctx, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*aggregate.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) List(ctx context.Context, page, pageSize int) ([]*aggregate.Order, int64, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*aggregate.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) GetLines(ctx context.Context, orderID uuidv7.UUID) ([]aggregate.OrderLine, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]aggregate.OrderLine), args.Error(1)
}

func (m *MockRepository) CreateLine(ctx context.Context, line *aggregate.OrderLine) error {
	args := m.Called(ctx, line)
	return args.Error(0)
}

func (m *MockRepository) UpdateLine(ctx context.Context, line *aggregate.OrderLine) error {
	args := m.Called(ctx, line)
	return args.Error(0)
}

func (m *MockRepository) DeleteLine(ctx context.Context, lineID uuidv7.UUID) error {
	args := m.Called(ctx, lineID)
	return args.Error(0)
}

func TestUseCase_CreateOrder(t *testing.T) {
	tests := []struct {
		name        string
		customerID  uuidv7.UUID
		companyID   *uuidv7.UUID
		currency    string
		mockSetup   func(*MockRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful creation",
			customerID: uuidv7.New(),
			companyID:  nil,
			currency:   "USD",
			mockSetup: func(repo *MockRepository) {
				repo.On("Create", mock.Anything, mock.AnythingOfType("*aggregate.Order")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "empty customer ID",
			customerID:  uuidv7.Nil,
			companyID:   nil,
			currency:    "USD",
			mockSetup:   func(repo *MockRepository) {},
			wantErr:     true,
			errContains: "customer ID is required",
		},
		{
			name:        "empty currency",
			customerID:  uuidv7.New(),
			companyID:   nil,
			currency:    "",
			mockSetup:   func(repo *MockRepository) {},
			wantErr:     true,
			errContains: "currency is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)
			uc := NewOrderUseCase(mockRepo, nil)

			order, err := uc.CreateOrder(context.Background(), tt.customerID, tt.companyID, tt.currency)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, order)
			assert.Equal(t, tt.customerID, order.CustomerID)
			assert.Equal(t, tt.currency, order.Currency)
			assert.Equal(t, aggregate.OrderStatusPending, order.Status)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_AddOrderLine(t *testing.T) {
	ctx := context.Background()
	orderID := uuidv7.New()
	productID := uuidv7.New()
	customerID := uuidv7.New()

	tests := []struct {
		name        string
		quantity    int
		unitPrice   valueobject.Money
		mockSetup   func(*MockRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:      "successful add line",
			quantity:  2,
			unitPrice: valueobject.Money{Amount: 5000, Currency: "USD"},
			mockSetup: func(repo *MockRepository) {
				existingOrder, _ := aggregate.NewOrder(customerID, "USD")
				existingOrder.ID = orderID
				repo.On("GetByID", mock.Anything, orderID).Return(existingOrder, nil)
				repo.On("GetLines", mock.Anything, orderID).Return([]aggregate.OrderLine{}, nil)
				repo.On("CreateLine", mock.Anything, mock.AnythingOfType("*aggregate.OrderLine")).Return(nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*aggregate.Order")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "zero quantity",
			quantity:  0,
			unitPrice: valueobject.Money{Amount: 5000, Currency: "USD"},
			mockSetup: func(repo *MockRepository) {
				existingOrder, _ := aggregate.NewOrder(customerID, "USD")
				existingOrder.ID = orderID
				repo.On("GetByID", mock.Anything, orderID).Return(existingOrder, nil)
				repo.On("GetLines", mock.Anything, orderID).Return([]aggregate.OrderLine{}, nil)
			},
			wantErr:     true,
			errContains: "quantity must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)
			uc := NewOrderUseCase(mockRepo, nil)

			order, err := uc.AddOrderLine(ctx, orderID, productID, tt.quantity, tt.unitPrice)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, order)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_ConfirmOrder(t *testing.T) {
	ctx := context.Background()
	orderID := uuidv7.New()
	customerID := uuidv7.New()

	tests := []struct {
		name        string
		mockSetup   func(*MockRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful confirmation",
			mockSetup: func(repo *MockRepository) {
				existingOrder, _ := aggregate.NewOrder(customerID, "USD")
				existingOrder.ID = orderID
				_ = existingOrder.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 5000, Currency: "USD"})
				repo.On("GetByID", mock.Anything, orderID).Return(existingOrder, nil)
				repo.On("GetLines", mock.Anything, orderID).Return(existingOrder.Lines, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*aggregate.Order")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "confirm without lines",
			mockSetup: func(repo *MockRepository) {
				existingOrder, _ := aggregate.NewOrder(customerID, "USD")
				existingOrder.ID = orderID
				repo.On("GetByID", mock.Anything, orderID).Return(existingOrder, nil)
				repo.On("GetLines", mock.Anything, orderID).Return([]aggregate.OrderLine{}, nil)
			},
			wantErr:     true,
			errContains: "order must have at least one line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)
			uc := NewOrderUseCase(mockRepo, nil)

			order, err := uc.ConfirmOrder(ctx, orderID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, order)
			assert.Equal(t, aggregate.OrderStatusConfirmed, order.Status)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_StartProcessing(t *testing.T) {
	ctx := context.Background()
	orderID := uuidv7.New()
	customerID := uuidv7.New()

	tests := []struct {
		name        string
		mockSetup   func(*MockRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful start processing",
			mockSetup: func(repo *MockRepository) {
				existingOrder, _ := aggregate.NewOrder(customerID, "USD")
				existingOrder.ID = orderID
				_ = existingOrder.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 5000, Currency: "USD"})
				_ = existingOrder.Confirm()
				repo.On("GetByID", mock.Anything, orderID).Return(existingOrder, nil)
				repo.On("GetLines", mock.Anything, orderID).Return(existingOrder.Lines, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*aggregate.Order")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			mockSetup: func(repo *MockRepository) {
				existingOrder, _ := aggregate.NewOrder(customerID, "USD")
				existingOrder.ID = orderID
				repo.On("GetByID", mock.Anything, orderID).Return(existingOrder, nil)
				repo.On("GetLines", mock.Anything, orderID).Return([]aggregate.OrderLine{}, nil)
			},
			wantErr:     true,
			errContains: "invalid order status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)
			uc := NewOrderUseCase(mockRepo, nil)

			order, err := uc.StartProcessing(ctx, orderID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, order)
			assert.Equal(t, aggregate.OrderStatusProcessing, order.Status)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_CancelOrder(t *testing.T) {
	ctx := context.Background()
	orderID := uuidv7.New()
	customerID := uuidv7.New()

	tests := []struct {
		name        string
		mockSetup   func(*MockRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful cancellation",
			mockSetup: func(repo *MockRepository) {
				existingOrder, _ := aggregate.NewOrder(customerID, "USD")
				existingOrder.ID = orderID
				_ = existingOrder.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 5000, Currency: "USD"})
				repo.On("GetByID", mock.Anything, orderID).Return(existingOrder, nil)
				repo.On("GetLines", mock.Anything, orderID).Return(existingOrder.Lines, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*aggregate.Order")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "cannot cancel fulfilled",
			mockSetup: func(repo *MockRepository) {
				existingOrder, _ := aggregate.NewOrder(customerID, "USD")
				existingOrder.ID = orderID
				_ = existingOrder.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 5000, Currency: "USD"})
				_ = existingOrder.Confirm()
				_ = existingOrder.StartProcessing()
				_ = existingOrder.MarkFulfilled()
				repo.On("GetByID", mock.Anything, orderID).Return(existingOrder, nil)
				repo.On("GetLines", mock.Anything, orderID).Return(existingOrder.Lines, nil)
			},
			wantErr:     true,
			errContains: "cannot cancel fulfilled order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)
			uc := NewOrderUseCase(mockRepo, nil)

			order, err := uc.CancelOrder(ctx, orderID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, order)
			assert.Equal(t, aggregate.OrderStatusCancelled, order.Status)
			mockRepo.AssertExpectations(t)
		})
	}
}
