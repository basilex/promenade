package receipt

import "errors"

// Repository Errors - Data access failures
var (
	// ErrReceiptNotFound is returned when receipt doesn't exist
	ErrReceiptNotFound = errors.New("receipt not found")

	// ErrReceiptAlreadyExists is returned when receipt already exists for order
	ErrReceiptAlreadyExists = errors.New("receipt already exists for order")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrCashRegisterIDRequired is returned when cash register ID is missing
	ErrCashRegisterIDRequired = errors.New("cash register ID is required")

	// ErrOrderIDRequired is returned when order ID is missing
	ErrOrderIDRequired = errors.New("order ID is required")

	// ErrPaymentTypeRequired is returned when payment type is missing
	ErrPaymentTypeRequired = errors.New("payment type is required")

	// ErrReceiptTypeRequired is returned when receipt type is missing
	ErrReceiptTypeRequired = errors.New("receipt type is required")

	// ErrCurrencyRequired is returned when currency is missing
	ErrCurrencyRequired = errors.New("currency is required")

	// ErrCreatedByRequired is returned when created by user ID is missing
	ErrCreatedByRequired = errors.New("created by user ID is required")

	// ErrReceiptLineNameRequired is returned when receipt line name is missing
	ErrReceiptLineNameRequired = errors.New("receipt line name is required")

	// ErrReceiptLineQuantityInvalid is returned when receipt line quantity is invalid
	ErrReceiptLineQuantityInvalid = errors.New("receipt line quantity must be greater than zero")

	// ErrReceiptLinePriceInvalid is returned when receipt line price is invalid
	ErrReceiptLinePriceInvalid = errors.New("receipt line price must be zero or greater")

	// ErrReceiptLineTaxRateInvalid is returned when receipt line tax rate is invalid
	ErrReceiptLineTaxRateInvalid = errors.New("receipt line tax rate must be between 0 and 100")

	// ErrReceiptAlreadyPrinted is returned when receipt is already printed
	ErrReceiptAlreadyPrinted = errors.New("receipt already printed")

	// ErrReceiptAlreadyCancelled is returned when receipt is already cancelled
	ErrReceiptAlreadyCancelled = errors.New("receipt already cancelled")

	// ErrReceiptCancelReasonRequired is returned when cancellation reason is missing
	ErrReceiptCancelReasonRequired = errors.New("cancellation reason is required")

	// ErrFiscalNumberRequired is returned when fiscal number is missing
	ErrFiscalNumberRequired = errors.New("fiscal number is required")
)

// Technical Operation Errors - Operation wrappers
var (
	// ErrReceiptCreateFailed is returned when creation fails
	ErrReceiptCreateFailed = errors.New("failed to create receipt")

	// ErrReceiptPrintFailed is returned when printing fails
	ErrReceiptPrintFailed = errors.New("failed to print receipt")

	// ErrReceiptUpdateFailed is returned when update fails
	ErrReceiptUpdateFailed = errors.New("failed to update receipt")

	// ErrReceiptDeleteFailed is returned when deletion fails
	ErrReceiptDeleteFailed = errors.New("failed to delete receipt")

	// ErrReceiptListFailed is returned when list fails
	ErrReceiptListFailed = errors.New("failed to list receipts")
)
