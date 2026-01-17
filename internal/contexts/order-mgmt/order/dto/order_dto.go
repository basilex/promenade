package dto

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/order/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	CustomerID uuidv7.UUID  `json:"customer_id" binding:"required"`
	CompanyID  *uuidv7.UUID `json:"company_id,omitempty"`
	Currency   string       `json:"currency" binding:"required,len=3"`
}

// AddOrderLineRequest represents a request to add a line to an order
type AddOrderLineRequest struct {
	ProductID uuidv7.UUID `json:"product_id" binding:"required"`
	Quantity  int         `json:"quantity" binding:"required,min=1"`
	UnitPrice int64       `json:"unit_price" binding:"required,min=0"` // Amount in cents
	Currency  string      `json:"currency" binding:"required,len=3"`
}

// UpdateOrderLineRequest represents a request to update a line quantity
type UpdateOrderLineRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// OrderResponse represents an order in responses
type OrderResponse struct {
	ID          string              `json:"id"`
	OrderNumber string              `json:"order_number"`
	CustomerID  string              `json:"customer_id"`
	CompanyID   *string             `json:"company_id,omitempty"`
	Lines       []OrderLineResponse `json:"lines"`
	Total       MoneyResponse       `json:"total"`
	Status      string              `json:"status"`
	OrderDate   time.Time           `json:"order_date"`
	ConfirmedAt *time.Time          `json:"confirmed_at,omitempty"`
	FulfilledAt *time.Time          `json:"fulfilled_at,omitempty"`
	CancelledAt *time.Time          `json:"cancelled_at,omitempty"`
	ContractID  *string             `json:"contract_id,omitempty"`
	InvoiceID   *string             `json:"invoice_id,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// OrderLineResponse represents an order line in responses
type OrderLineResponse struct {
	ID        string        `json:"id"`
	ProductID string        `json:"product_id"`
	Quantity  int           `json:"quantity"`
	UnitPrice MoneyResponse `json:"unit_price"`
	Total     MoneyResponse `json:"total"`
}

// MoneyResponse represents money in responses
type MoneyResponse struct {
	Amount   int64  `json:"amount"`   // Amount in cents
	Currency string `json:"currency"` // ISO 4217 currency code
}

// ToOrderResponse converts an order entity to response DTO
func ToOrderResponse(o *aggregate.Order) OrderResponse {
	resp := OrderResponse{
		ID:          o.ID.String(),
		OrderNumber: o.OrderNumber,
		CustomerID:  o.CustomerID.String(),
		Total: MoneyResponse{
			Amount:   o.Total.Amount,
			Currency: o.Total.Currency,
		},
		Status:      string(o.Status),
		OrderDate:   o.OrderDate,
		ConfirmedAt: o.ConfirmedAt,
		FulfilledAt: o.FulfilledAt,
		CancelledAt: o.CancelledAt,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}

	if o.CompanyID != nil {
		companyID := o.CompanyID.String()
		resp.CompanyID = &companyID
	}

	if o.ContractID != nil {
		contractID := o.ContractID.String()
		resp.ContractID = &contractID
	}

	if o.InvoiceID != nil {
		invoiceID := o.InvoiceID.String()
		resp.InvoiceID = &invoiceID
	}

	// Convert lines
	resp.Lines = make([]OrderLineResponse, 0, len(o.Lines))
	for _, line := range o.Lines {
		resp.Lines = append(resp.Lines, OrderLineResponse{
			ID:        line.ID.String(),
			ProductID: line.ProductID.String(),
			Quantity:  line.Quantity,
			UnitPrice: MoneyResponse{
				Amount:   line.UnitPrice.Amount,
				Currency: line.UnitPrice.Currency,
			},
			Total: MoneyResponse{
				Amount:   line.Total.Amount,
				Currency: line.Total.Currency,
			},
		})
	}

	return resp
}

// ToOrderListResponse converts multiple orders to response DTOs
func ToOrderListResponse(orders []*aggregate.Order) []OrderResponse {
	responses := make([]OrderResponse, 0, len(orders))
	for _, o := range orders {
		responses = append(responses, ToOrderResponse(o))
	}
	return responses
}
