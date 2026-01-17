package aggregate


import ordererrors "github.com/basilex/promenade/internal/contexts/order-mgmt/order"

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// OrderStatus represents the lifecycle status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusFulfilled  OrderStatus = "fulfilled"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

// Order is the aggregate root for order management
type Order struct {
	aggregate.BaseAggregate

	// Identity (ID, CreatedAt, UpdatedAt inherited from BaseAggregate)
	OrderNumber string // human-readable: ORD-2026-001
	CustomerID  uuidv7.UUID
	CompanyID   *uuidv7.UUID // optional (B2B)

	// Order Details
	Lines    []OrderLine
	Total    valueobject.Money
	Currency string

	// Status
	Status OrderStatus

	// Dates
	OrderDate   time.Time
	ConfirmedAt *time.Time
	FulfilledAt *time.Time
	CancelledAt *time.Time

	// References (to other contexts)
	ContractID *uuidv7.UUID // references Contract (same context)
	InvoiceID  *uuidv7.UUID // references Billing.Invoice
}

// OrderLine represents a line item in an order
type OrderLine struct {
	ID        uuidv7.UUID
	OrderID   uuidv7.UUID
	ProductID uuidv7.UUID // references Warehouse.Product
	Quantity  int
	UnitPrice valueobject.Money
	Total     valueobject.Money
}

// NewOrder creates a new order in pending status
func NewOrder(customerID uuidv7.UUID, currency string) (*Order, error) {
	if customerID == uuidv7.Nil {
		return nil, ordererrors.ErrCustomerIDRequired
	}

	if currency == "" {
		return nil, ordererrors.ErrCurrencyRequired
	}

	now := time.Now()
	orderNumber := generateOrderNumber(now)

	return &Order{
		BaseAggregate: aggregate.NewBaseAggregate(),
		OrderNumber:   orderNumber,
		CustomerID:    customerID,
		Lines:         []OrderLine{},
		Total:         valueobject.Money{Amount: 0, Currency: currency},
		Currency:      currency,
		Status:        OrderStatusPending,
		OrderDate:     now,
	}, nil
}

// generateOrderNumber creates a human-readable order number
// Format: ORD-YYYY-NNNNNN (e.g., ORD-2026-000001)
func generateOrderNumber(orderDate time.Time) string {
	year := orderDate.Year()
	// Use UUID v7 for guaranteed uniqueness (no collisions even in rapid creation)
	id := uuidv7.New()
	// Convert last 3 bytes to number for 6-digit suffix
	bytes := [16]byte(id)
	number := (uint32(bytes[13]) << 16) | (uint32(bytes[14]) << 8) | uint32(bytes[15])
	number = number % 1000000 // Keep 6 digits
	return fmt.Sprintf("ORD-%d-%06d", year, number)
}

// AddLine adds a new line item to the order
func (o *Order) AddLine(productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) error {
	if o.Status != OrderStatusPending {
		return ordererrors.ErrOrderAlreadyConfirmed
	}

	if productID == uuidv7.Nil {
		return ordererrors.ErrProductIDRequired
	}

	if quantity <= 0 {
		return ordererrors.ErrInvalidQuantity
	}

	if unitPrice.Amount < 0 {
		return ordererrors.ErrInvalidPrice
	}

	if unitPrice.Currency != o.Currency {
		return ordererrors.ErrCurrencyMismatch
	}

	line := OrderLine{
		ID:        uuidv7.New(),
		OrderID:   o.ID,
		ProductID: productID,
		Quantity:  quantity,
		UnitPrice: unitPrice,
		Total:     valueobject.Money{Amount: unitPrice.Amount * int64(quantity), Currency: unitPrice.Currency},
	}

	o.Lines = append(o.Lines, line)
	o.recalculateTotal()
	o.Touch()

	return nil
}

// RemoveLine removes a line item from the order
func (o *Order) RemoveLine(lineID uuidv7.UUID) error {
	if o.Status != OrderStatusPending {
		return ordererrors.ErrOrderAlreadyConfirmed
	}

	index := -1
	for i, line := range o.Lines {
		if line.ID == lineID {
			index = i
			break
		}
	}

	if index == -1 {
		return ordererrors.ErrLineNotFound
	}

	// Remove line by index
	o.Lines = append(o.Lines[:index], o.Lines[index+1:]...)
	o.recalculateTotal()
	o.Touch()

	return nil
}

// UpdateLineQuantity updates the quantity of a line item
func (o *Order) UpdateLineQuantity(lineID uuidv7.UUID, quantity int) error {
	if o.Status != OrderStatusPending {
		return ordererrors.ErrOrderAlreadyConfirmed
	}

	if quantity <= 0 {
		return ordererrors.ErrInvalidQuantity
	}

	for i, line := range o.Lines {
		if line.ID == lineID {
			o.Lines[i].Quantity = quantity
			o.Lines[i].Total = valueobject.Money{
				Amount:   line.UnitPrice.Amount * int64(quantity),
				Currency: line.UnitPrice.Currency,
			}
			o.recalculateTotal()
			o.Touch()
			return nil
		}
	}

	return ordererrors.ErrLineNotFound
}

// Confirm confirms the order (marks as ready for processing)
func (o *Order) Confirm() error {
	if o.Status != OrderStatusPending {
		return ordererrors.ErrInvalidOrderStatus
	}

	if len(o.Lines) == 0 {
		return ordererrors.ErrOrderEmpty
	}

	now := time.Now()
	o.Status = OrderStatusConfirmed
	o.ConfirmedAt = &now
	o.Touch()

	return nil
}

// StartProcessing moves order to processing status
func (o *Order) StartProcessing() error {
	if o.Status != OrderStatusConfirmed {
		return ordererrors.ErrInvalidOrderStatus
	}

	o.Status = OrderStatusProcessing
	o.Touch()

	return nil
}

// MarkFulfilled marks the order as fulfilled
func (o *Order) MarkFulfilled() error {
	if o.Status != OrderStatusProcessing {
		return ordererrors.ErrInvalidOrderStatus
	}

	now := time.Now()
	o.Status = OrderStatusFulfilled
	o.FulfilledAt = &now
	o.Touch()

	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if o.Status == OrderStatusCancelled {
		return ordererrors.ErrOrderAlreadyCancelled
	}

	if o.Status == OrderStatusFulfilled {
		return ordererrors.ErrCannotCancelFulfilled
	}

	now := time.Now()
	o.Status = OrderStatusCancelled
	o.CancelledAt = &now
	o.Touch()

	return nil
}

// recalculateTotal recalculates the order total from all line items
func (o *Order) recalculateTotal() {
	var total int64
	for _, line := range o.Lines {
		total += line.Total.Amount
	}

	o.Total = valueobject.Money{
		Amount:   total,
		Currency: o.Currency,
	}
}

// Validate performs business rule validation
func (o *Order) Validate() error {
	if o.CustomerID == uuidv7.Nil {
		return ordererrors.ErrCustomerIDRequired
	}

	if o.Currency == "" {
		return ordererrors.ErrCurrencyRequired
	}

	if o.Status == "" {
		return ordererrors.ErrStatusRequired
	}

	return nil
}

// IsEditable returns true if the order can be edited
func (o *Order) IsEditable() bool {
	return o.Status == OrderStatusPending
}

// IsCancellable returns true if the order can be cancelled
func (o *Order) IsCancellable() bool {
	return o.Status != OrderStatusCancelled && o.Status != OrderStatusFulfilled
}
