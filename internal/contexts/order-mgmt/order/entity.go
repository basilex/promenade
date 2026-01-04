package order

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

	// Soft delete
	DeletedAt *time.Time
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
		return nil, fmt.Errorf("customer ID is required")
	}

	if currency == "" {
		return nil, fmt.Errorf("currency is required")
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
	// In production, this would use a database sequence or counter
	// For now, use timestamp-based number with microsecond precision
	number := orderDate.UnixMicro() % 1000000
	return fmt.Sprintf("ORD-%d-%06d", year, number)
}

// AddLine adds a new line item to the order
func (o *Order) AddLine(productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) error {
	if o.Status != OrderStatusPending {
		return ErrOrderAlreadyConfirmed
	}

	if productID == uuidv7.Nil {
		return fmt.Errorf("product ID is required")
	}

	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if unitPrice.Amount < 0 {
		return ErrInvalidPrice
	}

	if unitPrice.Currency != o.Currency {
		return fmt.Errorf("line currency %s does not match order currency %s", unitPrice.Currency, o.Currency)
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
		return ErrOrderAlreadyConfirmed
	}

	index := -1
	for i, line := range o.Lines {
		if line.ID == lineID {
			index = i
			break
		}
	}

	if index == -1 {
		return ErrLineNotFound
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
		return ErrOrderAlreadyConfirmed
	}

	if quantity <= 0 {
		return ErrInvalidQuantity
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

	return ErrLineNotFound
}

// Confirm confirms the order (marks as ready for processing)
func (o *Order) Confirm() error {
	if o.Status != OrderStatusPending {
		return ErrInvalidOrderStatus
	}

	if len(o.Lines) == 0 {
		return ErrOrderEmpty
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
		return ErrInvalidOrderStatus
	}

	o.Status = OrderStatusProcessing
	o.Touch()

	return nil
}

// MarkFulfilled marks the order as fulfilled
func (o *Order) MarkFulfilled() error {
	if o.Status != OrderStatusProcessing {
		return ErrInvalidOrderStatus
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
		return ErrOrderAlreadyCancelled
	}

	if o.Status == OrderStatusFulfilled {
		return fmt.Errorf("cannot cancel fulfilled order")
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
		return fmt.Errorf("customer ID is required")
	}

	if o.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	if o.Status == "" {
		return fmt.Errorf("status is required")
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
