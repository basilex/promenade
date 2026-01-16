package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
)

// CreateReceiptRequest represents receipt creation request
type CreateReceiptRequest struct {
	CashRegisterID string               `json:"cash_register_id" binding:"required"`
	OrderID        string               `json:"order_id" binding:"required"`
	PaymentType    string               `json:"payment_type" binding:"required"`
	ReceiptType    string               `json:"receipt_type" binding:"required"`
	Currency       string               `json:"currency" binding:"required"`
	Lines          []ReceiptLineRequest `json:"lines" binding:"required"`
	CreatedBy      string               `json:"created_by" binding:"required"`
}

// MarkPrintedRequest represents printing request
type MarkPrintedRequest struct {
	FiscalNumber string `json:"fiscal_number" binding:"required"`
	FiscalURL    string `json:"fiscal_url"`
	QRCode       string `json:"qr_code"`
	PrintedBy    string `json:"printed_by" binding:"required"`
}

// CancelReceiptRequest represents cancellation request
type CancelReceiptRequest struct {
	Reason      string `json:"reason" binding:"required"`
	CancelledBy string `json:"cancelled_by" binding:"required"`
}

// ReceiptLineRequest represents receipt line in request
type ReceiptLineRequest struct {
	Name       string `json:"name" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required"`
	PriceCents int64  `json:"price" binding:"required"`
	TaxRate    int    `json:"tax_rate" binding:"required"`
}

// ReceiptResponse represents receipt response
type ReceiptResponse struct {
	ID                 string        `json:"id"`
	CashRegisterID     string        `json:"cash_register_id"`
	OrderID            string        `json:"order_id"`
	PaymentType        string        `json:"payment_type"`
	ReceiptType        string        `json:"receipt_type"`
	Currency           string        `json:"currency"`
	TotalAmount        int64         `json:"total_amount"`
	TaxAmount          int64         `json:"tax_amount"`
	FiscalNumber       string        `json:"fiscal_number,omitempty"`
	FiscalURL          string        `json:"fiscal_url,omitempty"`
	QRCode             string        `json:"qr_code,omitempty"`
	PrintedAt          *time.Time    `json:"printed_at,omitempty"`
	CancelledAt        *time.Time    `json:"cancelled_at,omitempty"`
	CancellationReason string        `json:"cancellation_reason,omitempty"`
	Lines              []ReceiptLine `json:"lines"`
	CreatedBy          string        `json:"created_by"`
	LastUpdatedBy      string        `json:"last_updated_by"`
	Status             string        `json:"status"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// ReceiptLine represents receipt line in response
type ReceiptLine struct {
	Name           string `json:"name"`
	Quantity       int    `json:"quantity"`
	PriceCents     int64  `json:"price"`
	TaxRate        int    `json:"tax_rate"`
	TotalCents     int64  `json:"total"`
	TaxAmountCents int64  `json:"tax_amount"`
}

// ToReceiptResponse converts entity to response DTO
func ToReceiptResponse(rec *receipt.Receipt) *ReceiptResponse {
	lines := rec.Lines.Get()
	respLines := make([]ReceiptLine, len(lines))
	for i, line := range lines {
		respLines[i] = ReceiptLine{
			Name:           line.Name,
			Quantity:       line.Quantity,
			PriceCents:     line.PriceCents,
			TaxRate:        line.TaxRate,
			TotalCents:     line.TotalCents,
			TaxAmountCents: line.TaxAmountCents,
		}
	}

	return &ReceiptResponse{
		ID:                 rec.GetID().String(),
		CashRegisterID:     rec.CashRegisterID.String(),
		OrderID:            rec.OrderID.String(),
		PaymentType:        string(rec.PaymentType),
		ReceiptType:        string(rec.ReceiptType),
		Currency:           rec.Currency,
		TotalAmount:        rec.TotalAmount,
		TaxAmount:          rec.TaxAmount,
		FiscalNumber:       rec.FiscalNumber,
		FiscalURL:          rec.FiscalURL,
		QRCode:             rec.QRCode,
		PrintedAt:          rec.PrintedAt,
		CancelledAt:        rec.CancelledAt,
		CancellationReason: rec.CancellationReason,
		Lines:              respLines,
		CreatedBy:          rec.CreatedBy.String(),
		LastUpdatedBy:      rec.LastUpdatedBy.String(),
		Status:             string(rec.Status),
		CreatedAt:          rec.CreatedAt,
		UpdatedAt:          rec.UpdatedAt,
	}
}

// ToReceiptListResponse converts slice of entities to response DTOs
func ToReceiptListResponse(receipts []*receipt.Receipt) []*ReceiptResponse {
	responses := make([]*ReceiptResponse, len(receipts))
	for i, rec := range receipts {
		responses[i] = ToReceiptResponse(rec)
	}
	return responses
}
