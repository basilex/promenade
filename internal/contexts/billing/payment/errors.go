package payment

import "errors"

var (
	// ErrPaymentNotFound is returned when a payment is not found
	ErrPaymentNotFound = errors.New("payment not found")

	// ErrPaymentAlreadyProcessed is returned when trying to process an already processed payment
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")

	// ErrPaymentAlreadyCompleted is returned when trying to modify a completed payment
	ErrPaymentAlreadyCompleted = errors.New("payment already completed")

	// ErrPaymentAlreadyRefunded is returned when trying to refund an already refunded payment
	ErrPaymentAlreadyRefunded = errors.New("payment already refunded")

	// ErrInvalidPaymentAmount is returned when payment amount is invalid
	ErrInvalidPaymentAmount = errors.New("invalid payment amount")

	// ErrRefundAmountExceedsPayment is returned when refund amount exceeds payment amount
	ErrRefundAmountExceedsPayment = errors.New("refund amount exceeds payment amount")

	// ErrPaymentMethodRequired is returned when payment method is not provided
	ErrPaymentMethodRequired = errors.New("payment method is required")

	// ErrTransactionIDRequired is returned when transaction ID is not provided
	ErrTransactionIDRequired = errors.New("transaction ID is required")

	// ErrPaymentCancelled is returned when trying to process a cancelled payment
	ErrPaymentCancelled = errors.New("payment is cancelled")

	// ErrPaymentFailed is returned when payment processing fails
	ErrPaymentFailed = errors.New("payment processing failed")

	// ErrPaymentInvalidStatusForDeletion is returned when attempting to delete payment with invalid status
	ErrPaymentInvalidStatusForDeletion = errors.New("payment status invalid for deletion")
)
