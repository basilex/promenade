package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/billing/invoice"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateInvoiceRequest represents invoice creation request
type CreateInvoiceRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	OrderID    string `json:"order_id"`
	DueDate    string `json:"due_date" binding:"required"` // RFC3339 format
	Currency   string `json:"currency" binding:"required,len=3"`
}

// AddLineItemRequest represents line item addition request
type AddLineItemRequest struct {
	Description string `json:"description" binding:"required,min=1,max=500"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
	UnitPrice   int64  `json:"unit_price" binding:"required,min=0"` // Amount in cents
}

// UpdateLineItemRequest represents line item update request
type UpdateLineItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// UpdateTaxRequest represents tax amount update request
type UpdateTaxRequest struct {
	TaxAmount int64 `json:"tax_amount" binding:"required,min=0"` // Amount in cents
}

// MarkAsPaidRequest represents payment date request
type MarkAsPaidRequest struct {
	PaidDate string `json:"paid_date" binding:"required"` // RFC3339 format
}

// InvoiceLineResponse represents a line item in response
type InvoiceLineResponse struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"` // Amount in cents
	Amount      int64  `json:"amount"`     // Total in cents
	Currency    string `json:"currency"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// InvoiceResponse represents invoice in response
type InvoiceResponse struct {
	ID             string                 `json:"id"`
	InvoiceNo      string                 `json:"invoice_no"`
	CustomerID     string                 `json:"customer_id"`
	OrderID        *string                `json:"order_id,omitempty"`
	SubtotalAmount int64                  `json:"subtotal_amount"` // Amount in cents
	TaxAmount      int64                  `json:"tax_amount"`      // Amount in cents
	TotalAmount    int64                  `json:"total_amount"`    // Amount in cents
	Currency       string                 `json:"currency"`
	IssueDate      string                 `json:"issue_date"`
	DueDate        string                 `json:"due_date"`
	PaidDate       *string                `json:"paid_date,omitempty"`
	Status         string                 `json:"status"`
	Lines          []InvoiceLineResponse  `json:"lines"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
}

// ListInvoicesRequest represents list query parameters
type ListInvoicesRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	CustomerID string `form:"customer_id"`
	OrderID    string `form:"order_id"`
	Status     string `form:"status"`
}

// ToInvoiceResponse converts entity to response DTO
func ToInvoiceResponse(inv *invoice.Invoice) *InvoiceResponse {
	resp := &InvoiceResponse{
		ID:             inv.ID.String(),
		InvoiceNo:      inv.InvoiceNo,
		CustomerID:     inv.CustomerID.String(),
		SubtotalAmount: inv.SubtotalAmount.Amount,
		TaxAmount:      inv.TaxAmount.Amount,
		TotalAmount:    inv.TotalAmount.Amount,
		Currency:       inv.Currency,
		IssueDate:      inv.IssueDate.Format(time.RFC3339),
		DueDate:        inv.DueDate.Format(time.RFC3339),
		Status:         string(inv.Status),
		Lines:          make([]InvoiceLineResponse, 0, len(inv.Lines)),
		CreatedAt:      inv.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      inv.UpdatedAt.Format(time.RFC3339),
	}

	if inv.OrderID != nil {
		orderID := inv.OrderID.String()
		resp.OrderID = &orderID
	}

	if inv.PaidDate != nil {
		paidDate := inv.PaidDate.Format(time.RFC3339)
		resp.PaidDate = &paidDate
	}

	for _, line := range inv.Lines {
		resp.Lines = append(resp.Lines, InvoiceLineResponse{
			ID:          line.ID.String(),
			Description: line.Description,
			Quantity:    line.Quantity,
			UnitPrice:   line.UnitPrice.Amount,
			Amount:      line.Amount.Amount,
			Currency:    line.UnitPrice.Currency,
			CreatedAt:   line.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   line.UpdatedAt.Format(time.RFC3339),
		})
	}

	return resp
}

// ToInvoiceListResponse converts multiple entities to response DTOs
func ToInvoiceListResponse(invoices []*invoice.Invoice) []*InvoiceResponse {
	responses := make([]*InvoiceResponse, 0, len(invoices))
	for _, inv := range invoices {
		responses = append(responses, ToInvoiceResponse(inv))
	}
	return responses
}

// ParseCustomerID parses customer ID from string
func ParseCustomerID(customerID string) (uuidv7.UUID, error) {
	return uuidv7.Parse(customerID)
}

// ParseOrderID parses optional order ID from string
func ParseOrderID(orderID string) (*uuidv7.UUID, error) {
	if orderID == "" {
		return nil, nil
	}
	id, err := uuidv7.Parse(orderID)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// ParseDueDate parses due date from RFC3339 string
func ParseDueDate(dueDate string) (time.Time, error) {
	return time.Parse(time.RFC3339, dueDate)
}

// ParsePaidDate parses paid date from RFC3339 string
func ParsePaidDate(paidDate string) (time.Time, error) {
	return time.Parse(time.RFC3339, paidDate)
}
