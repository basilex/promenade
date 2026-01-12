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

	// Technical operation errors

	// ErrPaymentCreateFailed is returned when payment creation fails
	ErrPaymentCreateFailed = errors.New("payment creation failed")

	// ErrPaymentGetFailed is returned when payment retrieval fails
	ErrPaymentGetFailed = errors.New("payment retrieval failed")

	// ErrPaymentUpdateFailed is returned when payment update fails
	ErrPaymentUpdateFailed = errors.New("payment update failed")

	// ErrPaymentDeleteFailed is returned when payment deletion fails
	ErrPaymentDeleteFailed = errors.New("payment deletion failed")

	// ErrPaymentNumberGenerationFailed is returned when payment number generation fails
	ErrPaymentNumberGenerationFailed = errors.New("payment number generation failed")

	// ErrPaymentLinkFailed is returned when linking payment to invoice fails
	ErrPaymentLinkFailed = errors.New("payment link operation failed")

	// ErrPaymentProcessingFailed is returned when payment processing operation fails
	ErrPaymentProcessingFailed = errors.New("payment processing operation failed")

	// ErrPaymentCompletionFailed is returned when payment completion operation fails
	ErrPaymentCompletionFailed = errors.New("payment completion operation failed")

	// ErrPaymentFailureFailed is returned when marking payment as failed fails
	ErrPaymentFailureFailed = errors.New("payment failure operation failed")

	// ErrPaymentRefundOperationFailed is returned when payment refund operation fails
	ErrPaymentRefundOperationFailed = errors.New("payment refund operation failed")

	// ErrPaymentCancellationFailed is returned when payment cancellation operation fails
	ErrPaymentCancellationFailed = errors.New("payment cancellation operation failed")

	// ErrPaymentListFailed is returned when listing payments fails
	ErrPaymentListFailed = errors.New("payment list operation failed")

	// ErrPaymentTotalCalculationFailed is returned when calculating payment totals fails
	ErrPaymentTotalCalculationFailed = errors.New("payment total calculation failed")

	// ErrMoneyCreationFailed is returned when creating money value object fails
	ErrMoneyCreationFailed = errors.New("money creation failed")

	// ErrPaymentInvalidStatusForDeletion is returned when attempting to delete payment with invalid status
	ErrPaymentInvalidStatusForDeletion = errors.New("payment status invalid for deletion")
)
