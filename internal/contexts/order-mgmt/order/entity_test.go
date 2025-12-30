package order

import (
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name        string
		customerID  uuidv7.UUID
		currency    string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid order",
			customerID: uuidv7.New(),
			currency:   "USD",
			wantErr:    false,
		},
		{
			name:       "valid order with EUR",
			customerID: uuidv7.New(),
			currency:   "EUR",
			wantErr:    false,
		},
		{
			name:        "empty customer ID",
			customerID:  uuidv7.Nil,
			currency:    "USD",
			wantErr:     true,
			errContains: "customer ID is required",
		},
		{
			name:        "empty currency",
			customerID:  uuidv7.New(),
			currency:    "",
			wantErr:     true,
			errContains: "currency is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(tt.customerID, tt.currency)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, order)
			assert.NotEqual(t, uuidv7.Nil, order.ID)
			assert.Equal(t, tt.customerID, order.CustomerID)
			assert.Equal(t, tt.currency, order.Currency)
			assert.Equal(t, OrderStatusPending, order.Status)
			assert.NotEmpty(t, order.OrderNumber)
			assert.Contains(t, order.OrderNumber, "ORD-")
			assert.Empty(t, order.Lines)
			assert.Equal(t, int64(0), order.Total.Amount)
		})
	}
}

func TestOrder_AddLine(t *testing.T) {
	tests := []struct {
		name        string
		quantity    int
		unitPrice   int64
		currency    string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid line",
			quantity:  3,
			unitPrice: 10000,
			currency:  "USD",
			wantErr:   false,
		},
		{
			name:        "zero quantity",
			quantity:    0,
			unitPrice:   5000,
			currency:    "USD",
			wantErr:     true,
			errContains: "quantity must be greater than zero",
		},
		{
			name:        "negative quantity",
			quantity:    -1,
			unitPrice:   5000,
			currency:    "USD",
			wantErr:     true,
			errContains: "quantity must be greater than zero",
		},
		{
			name:        "zero unit price",
			quantity:    2,
			unitPrice:   0,
			currency:    "USD",
			wantErr:     false, // Zero price is allowed
		},
		{
			name:        "negative unit price",
			quantity:    2,
			unitPrice:   -1000,
			currency:    "USD",
			wantErr:     true,
			errContains: "price cannot be negative",
		},
		{
			name:        "currency mismatch",
			quantity:    2,
			unitPrice:   5000,
			currency:    "EUR",
			wantErr:     true,
			errContains: "currency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(uuidv7.New(), "USD")
			require.NoError(t, err)

			productID := uuidv7.New()
			unitPrice := valueobject.Money{Amount: tt.unitPrice, Currency: tt.currency}

			err = order.AddLine(productID, tt.quantity, unitPrice)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, order.Lines)
			require.Len(t, order.Lines, 1)

			line := order.Lines[0]
			assert.NotEqual(t, uuidv7.Nil, line.ID)
			assert.Equal(t, order.ID, line.OrderID)
			assert.Equal(t, productID, line.ProductID)
			assert.Equal(t, tt.quantity, line.Quantity)
			assert.Equal(t, tt.unitPrice, line.UnitPrice.Amount)

			expectedSubtotal := int64(tt.quantity) * tt.unitPrice
			assert.Equal(t, expectedSubtotal, line.Total.Amount)
			assert.Equal(t, expectedSubtotal, order.Total.Amount)
			assert.Equal(t, 1, len(order.Lines))
		})
	}
}

func TestOrder_AddMultipleLines(t *testing.T) {
	order, err := NewOrder(uuidv7.New(), "USD")
	require.NoError(t, err)

	err = order.AddLine(uuidv7.New(), 3, valueobject.Money{Amount: 10000, Currency: "USD"})
	require.NoError(t, err)
	assert.Equal(t, int64(30000), order.Total.Amount)

	err = order.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 7500, Currency: "USD"})
	require.NoError(t, err)
	assert.Equal(t, int64(45000), order.Total.Amount)
	assert.Equal(t, 2, len(order.Lines))
}

func TestOrder_RemoveLine(t *testing.T) {
	order, err := NewOrder(uuidv7.New(), "USD")
	require.NoError(t, err)

	err = order.AddLine(uuidv7.New(), 3, valueobject.Money{Amount: 10000, Currency: "USD"})
	require.NoError(t, err)
	line1ID := order.Lines[0].ID

	err = order.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 7500, Currency: "USD"})
	require.NoError(t, err)
	line2ID := order.Lines[1].ID

	assert.Equal(t, int64(45000), order.Total.Amount)
	assert.Equal(t, 2, len(order.Lines))

	err = order.RemoveLine(line1ID)
	require.NoError(t, err)
	assert.Equal(t, int64(15000), order.Total.Amount)
	assert.Equal(t, 1, len(order.Lines))
	assert.Equal(t, line2ID, order.Lines[0].ID)

	err = order.RemoveLine(uuidv7.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "order line not found")
}

func TestOrder_UpdateLineQuantity(t *testing.T) {
	tests := []struct {
		name            string
		initialQuantity int
		newQuantity     int
		wantErr         bool
		errContains     string
	}{
		{
			name:            "increase quantity",
			initialQuantity: 3,
			newQuantity:     5,
			wantErr:         false,
		},
		{
			name:            "decrease quantity",
			initialQuantity: 5,
			newQuantity:     2,
			wantErr:         false,
		},
		{
			name:            "zero quantity",
			initialQuantity: 3,
			newQuantity:     0,
			wantErr:         true,
			errContains:     "quantity must be greater than zero",
		},
		{
			name:            "negative quantity",
			initialQuantity: 3,
			newQuantity:     -1,
			wantErr:         true,
			errContains:     "quantity must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(uuidv7.New(), "USD")
			require.NoError(t, err)

			unitPrice := int64(10000)
			err = order.AddLine(uuidv7.New(), tt.initialQuantity, valueobject.Money{Amount: unitPrice, Currency: "USD"})
			require.NoError(t, err)

			initialTotal := int64(tt.initialQuantity) * unitPrice
			assert.Equal(t, initialTotal, order.Total.Amount)

			err = order.UpdateLineQuantity(order.Lines[0].ID, tt.newQuantity)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Equal(t, initialTotal, order.Total.Amount)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.newQuantity, order.Lines[0].Quantity)

			expectedSubtotal := int64(tt.newQuantity) * unitPrice
			assert.Equal(t, expectedSubtotal, order.Lines[0].Total.Amount)
			assert.Equal(t, expectedSubtotal, order.Total.Amount)
		})
	}
}

func TestOrder_Confirm(t *testing.T) {
	tests := []struct {
		name        string
		addLines    bool
		wantErr     bool
		errContains string
	}{
		{
			name:     "confirm with lines",
			addLines: true,
			wantErr:  false,
		},
		{
			name:        "confirm without lines",
			addLines:    false,
			wantErr:     true,
			errContains: "order must have at least one line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(uuidv7.New(), "USD")
			require.NoError(t, err)

			if tt.addLines {
				err = order.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 5000, Currency: "USD"})
				require.NoError(t, err)
			}

			err = order.Confirm()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Equal(t, OrderStatusPending, order.Status)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, OrderStatusConfirmed, order.Status)

			err = order.Confirm()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid order status")
		})
	}
}

func TestOrder_StartProcessing(t *testing.T) {
	order, err := NewOrder(uuidv7.New(), "USD")
	require.NoError(t, err)

	err = order.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 5000, Currency: "USD"})
	require.NoError(t, err)

	err = order.StartProcessing()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order status")

	err = order.Confirm()
	require.NoError(t, err)

	err = order.StartProcessing()
	require.NoError(t, err)
	assert.Equal(t, OrderStatusProcessing, order.Status)

	err = order.StartProcessing()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order status")
}

func TestOrder_MarkFulfilled(t *testing.T) {
	order, err := NewOrder(uuidv7.New(), "USD")
	require.NoError(t, err)

	err = order.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 5000, Currency: "USD"})
	require.NoError(t, err)

	err = order.MarkFulfilled()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order status")

	err = order.Confirm()
	require.NoError(t, err)

	err = order.StartProcessing()
	require.NoError(t, err)

	err = order.MarkFulfilled()
	require.NoError(t, err)
	assert.Equal(t, OrderStatusFulfilled, order.Status)

	err = order.MarkFulfilled()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order status")
}

func TestOrder_Cancel(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus OrderStatus
		wantErr       bool
		errContains   string
	}{
		{
			name:          "cancel pending",
			initialStatus: OrderStatusPending,
			wantErr:       false,
		},
		{
			name:          "cancel confirmed",
			initialStatus: OrderStatusConfirmed,
			wantErr:       false,
		},
		{
			name:          "cancel processing",
			initialStatus: OrderStatusProcessing,
			wantErr:       false,
		},
		{
			name:          "cannot cancel fulfilled",
			initialStatus: OrderStatusFulfilled,
			wantErr:       true,
			errContains:   "cannot cancel fulfilled order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(uuidv7.New(), "USD")
			require.NoError(t, err)

			err = order.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 5000, Currency: "USD"})
			require.NoError(t, err)

			switch tt.initialStatus {
			case OrderStatusConfirmed:
				err = order.Confirm()
				require.NoError(t, err)
			case OrderStatusProcessing:
				err = order.Confirm()
				require.NoError(t, err)
				err = order.StartProcessing()
				require.NoError(t, err)
			case OrderStatusFulfilled:
				err = order.Confirm()
				require.NoError(t, err)
				err = order.StartProcessing()
				require.NoError(t, err)
				err = order.MarkFulfilled()
				require.NoError(t, err)
			}

			err = order.Cancel()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Equal(t, tt.initialStatus, order.Status)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, OrderStatusCancelled, order.Status)
		})
	}
}

func TestOrder_CalculateTotal(t *testing.T) {
	order, err := NewOrder(uuidv7.New(), "USD")
	require.NoError(t, err)
	assert.Equal(t, int64(0), order.Total.Amount)

	err = order.AddLine(uuidv7.New(), 3, valueobject.Money{Amount: 10000, Currency: "USD"})
	require.NoError(t, err)
	assert.Equal(t, int64(30000), order.Total.Amount)

	err = order.AddLine(uuidv7.New(), 2, valueobject.Money{Amount: 7500, Currency: "USD"})
	require.NoError(t, err)
	assert.Equal(t, int64(45000), order.Total.Amount)

	err = order.AddLine(uuidv7.New(), 1, valueobject.Money{Amount: 2500, Currency: "USD"})
	require.NoError(t, err)
	assert.Equal(t, int64(47500), order.Total.Amount)

	firstLineID := order.Lines[0].ID
	err = order.RemoveLine(firstLineID)
	require.NoError(t, err)
	assert.Equal(t, int64(17500), order.Total.Amount)

	secondLineID := order.Lines[0].ID
	err = order.UpdateLineQuantity(secondLineID, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(40000), order.Total.Amount)
}
