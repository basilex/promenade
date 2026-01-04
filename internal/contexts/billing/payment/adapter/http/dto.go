package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/billing/payment"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreatePaymentRequest represents payment creation request
type CreatePaymentRequest struct {
	CustomerID  string  `json:"customer_id" binding:"required"`
	InvoiceID   *string `json:"invoice_id"`
	Amount      int64   `json:"amount" binding:"required,gt=0"`
	Currency    string  `json:"currency" binding:"required,len=3"`
	Method      string  `json:"method" binding:"required"`
	ProcessedBy string  `json:"processed_by"`
	Notes       string  `json:"notes"`
}

// LinkInvoiceRequest represents invoice linking request
type LinkInvoiceRequest struct {
	InvoiceID string `json:"invoice_id" binding:"required"`
}

// CompletePaymentRequest represents payment completion request
type CompletePaymentRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

// FailPaymentRequest represents payment failure request
type FailPaymentRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// RefundPaymentRequest represents refund request
type RefundPaymentRequest struct {
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required,len=3"`
}

// SetCardDetailsRequest represents card details request
type SetCardDetailsRequest struct {
	Last4 string `json:"last4" binding:"required,len=4"`
	Brand string `json:"brand" binding:"required"`
}

// SetProviderRequest represents provider setting request
type SetProviderRequest struct {
	Provider string `json:"provider" binding:"required"`
}

// AddNoteRequest represents note addition request
type AddNoteRequest struct {
	Note string `json:"note" binding:"required"`
}

// PaymentResponse represents payment response
type PaymentResponse struct {
	ID              string  `json:"id"`
	PaymentNo       string  `json:"payment_no"`
	TransactionID   string  `json:"transaction_id,omitempty"`
	CustomerID      string  `json:"customer_id"`
	InvoiceID       *string `json:"invoice_id,omitempty"`
	Amount          int64   `json:"amount"`
	Currency        string  `json:"currency"`
	Method          string  `json:"method"`
	CardLast4       string  `json:"card_last4,omitempty"`
	CardBrand       string  `json:"card_brand,omitempty"`
	BankAccount     string  `json:"bank_account,omitempty"`
	PaymentProvider string  `json:"payment_provider,omitempty"`
	Status          string  `json:"status"`
	FailureReason   string  `json:"failure_reason,omitempty"`
	ProcessedAt     string  `json:"processed_at"`
	RefundedAt      *string `json:"refunded_at,omitempty"`
	RefundedAmount  *int64  `json:"refunded_amount,omitempty"`
	ProcessedBy     string  `json:"processed_by,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// TotalResponse represents total amount response
type TotalResponse struct {
	Total    int64  `json:"total"`
	Currency string `json:"currency"`
}

// ToPaymentResponse converts Payment entity to response DTO
func ToPaymentResponse(p *payment.Payment) *PaymentResponse {
	resp := &PaymentResponse{
		ID:              p.ID.String(),
		PaymentNo:       p.PaymentNo,
		TransactionID:   p.TransactionID,
		CustomerID:      p.CustomerID.String(),
		Amount:          p.Amount.Amount,
		Currency:        p.Amount.Currency,
		Method:          string(p.Method),
		CardLast4:       p.CardLast4,
		CardBrand:       p.CardBrand,
		BankAccount:     p.BankAccount,
		PaymentProvider: p.PaymentProvider,
		Status:          string(p.Status),
		FailureReason:   p.FailureReason,
		ProcessedAt:     p.ProcessedAt.Format(time.RFC3339),
		ProcessedBy:     p.ProcessedBy,
		Notes:           p.Notes,
		CreatedAt:       p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       p.UpdatedAt.Format(time.RFC3339),
	}

	if p.InvoiceID != nil {
		invoiceIDStr := p.InvoiceID.String()
		resp.InvoiceID = &invoiceIDStr
	}

	if p.RefundedAt != nil {
		refundedAtStr := p.RefundedAt.Format(time.RFC3339)
		resp.RefundedAt = &refundedAtStr
	}

	if p.RefundedAmount != nil {
		refundedAmountCents := p.RefundedAmount.Amount
		resp.RefundedAmount = &refundedAmountCents
	}

	return resp
}

// ToPaymentResponseList converts list of Payment entities to response DTOs
func ToPaymentResponseList(payments []*payment.Payment) []*PaymentResponse {
	responses := make([]*PaymentResponse, len(payments))
	for i, p := range payments {
		responses[i] = ToPaymentResponse(p)
	}
	return responses
}

// ParseUUID parses a string to UUID v7
func ParseUUID(s string) (uuidv7.UUID, error) {
	return uuidv7.Parse(s)
}
