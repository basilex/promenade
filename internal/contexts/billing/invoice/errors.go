package invoice

import "errors"

// Invoice errors
var (
	// Entity validation errors
	ErrInvoiceInvalidID         = errors.New("invoice ID cannot be nil")
	ErrInvoiceInvalidCustomerID = errors.New("customer ID cannot be nil")
	ErrInvoiceInvalidDueDate    = errors.New("due date must be after issue date")
	ErrInvoiceInvalidStatus     = errors.New("invoice status is required")
	ErrInvoiceInvalidCurrency   = errors.New("currency is required")
	ErrInvoiceInvalidTaxAmount  = errors.New("tax amount cannot be negative")

	// Business logic errors
	ErrInvoiceCannotModifyNonDraft       = errors.New("cannot modify invoice that is not in draft status")
	ErrInvoiceInvalidStatusTransition    = errors.New("invalid invoice status transition")
	ErrInvoiceNoLines                    = errors.New("invoice must have at least one line item")
	ErrInvoiceNotYetDue                  = errors.New("invoice is not yet due")
	ErrInvoiceCannotCancelTerminalState  = errors.New("cannot cancel invoice in terminal state")
	ErrInvoiceCannotVoidTerminalState    = errors.New("cannot void invoice in terminal state")

	// Line item errors
	ErrInvoiceLineNotFound             = errors.New("invoice line not found")
	ErrInvoiceLineInvalidDescription   = errors.New("line description cannot be empty")
	ErrInvoiceLineInvalidQuantity      = errors.New("line quantity must be positive")
	ErrInvoiceLineInvalidUnitPrice     = errors.New("line unit price cannot be negative")

	// Repository errors
	ErrInvoiceNotFound        = errors.New("invoice not found")
	ErrInvoiceNumberExists    = errors.New("invoice number already exists")
	ErrInvoiceUnauthorized    = errors.New("unauthorized to access this invoice")

	// Technical Operation Errors (for wrapping repository/system errors)
	ErrInvoiceCreateFailed         = errors.New("failed to create invoice")
	ErrInvoiceGetFailed            = errors.New("failed to get invoice")
	ErrInvoiceUpdateFailed         = errors.New("failed to update invoice")
	ErrInvoiceDeleteFailed         = errors.New("failed to delete invoice")
	ErrInvoiceListFailed           = errors.New("failed to list invoices")
	ErrInvoiceValidationFailed     = errors.New("invoice validation failed")
	ErrInvoiceGenerateNumberFailed = errors.New("failed to generate invoice number")
	ErrInvoiceLineCreateFailed     = errors.New("failed to create invoice line")
	ErrInvoiceLineDeleteFailed     = errors.New("failed to delete invoice line")
	ErrInvoiceCalculationFailed    = errors.New("failed to calculate invoice amount")
	ErrInvoiceRevenueCalculationFailed = errors.New("failed to calculate total revenue")
)
