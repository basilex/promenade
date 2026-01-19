package order

import "errors"

// Repository Errors - Data access failures
var (
	// ErrOrderNotFound is returned when order doesn't exist
	ErrOrderNotFound = errors.New("order not found")

	// ErrLineNotFound is returned when order line is not found
	ErrLineNotFound = errors.New("order line not found")

	// ErrOrderLineNotFound is an alias for ErrLineNotFound for consistency
	ErrOrderLineNotFound = ErrLineNotFound
)

// Business Logic Errors - Domain rule violations
var (
	// ErrOrderAlreadyConfirmed is returned when trying to modify confirmed order
	ErrOrderAlreadyConfirmed = errors.New("order already confirmed")

	// ErrOrderAlreadyCancelled is returned when trying to modify cancelled order
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")

	// ErrOrderAlreadyFulfilled is returned when trying to modify fulfilled order
	ErrOrderAlreadyFulfilled = errors.New("order is already fulfilled")

	// ErrOrderEmpty is returned when order has no line items
	ErrOrderEmpty = errors.New("order must have at least one line item")

	// ErrInvalidOrderStatus is returned when invalid status transition attempted
	ErrInvalidOrderStatus = errors.New("invalid order status transition")

	// ErrOrderNotConfirmed is returned when processing unconfirmed order
	ErrOrderNotConfirmed = errors.New("order must be confirmed before processing")

	// ErrOrderNotProcessing is returned when fulfilling order not in processing
	ErrOrderNotProcessing = errors.New("order must be in processing status to be fulfilled")

	// ErrCannotCancelFulfilled is returned when trying to cancel fulfilled order
	ErrCannotCancelFulfilled = errors.New("cannot cancel fulfilled order")

	// ErrInvalidQuantity is returned when quantity is less than or equal to zero
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")

	// ErrInvalidPrice is returned when price is negative
	ErrInvalidPrice = errors.New("price cannot be negative")

	// ErrCurrencyMismatch is returned when line currency doesn't match order currency
	ErrCurrencyMismatch = errors.New("line currency does not match order currency")

	// ErrCustomerIDRequired is returned when customer_id is missing
	ErrCustomerIDRequired = errors.New("customer ID is required")

	// ErrCurrencyRequired is returned when currency is missing
	ErrCurrencyRequired = errors.New("currency is required")

	// ErrProductIDRequired is returned when product_id is missing
	ErrProductIDRequired = errors.New("product ID is required")

	// ErrStatusRequired is returned when status is missing
	ErrStatusRequired = errors.New("status is required")
)

// Technical Operation Errors - Operation wrappers
var (
	// ErrOrderCreateFailed is returned when order creation fails
	ErrOrderCreateFailed = errors.New("failed to create order")

	// ErrOrderPersistFailed is returned when order persistence fails
	ErrOrderPersistFailed = errors.New("failed to persist order")

	// ErrOrderGetFailed is returned when order retrieval fails
	ErrOrderGetFailed = errors.New("failed to get order")

	// ErrOrderUpdateFailed is returned when order update fails
	ErrOrderUpdateFailed = errors.New("failed to update order")

	// ErrOrderListFailed is returned when listing orders fails
	ErrOrderListFailed = errors.New("failed to list orders")

	// ErrOrderLinesFetchFailed is returned when fetching order lines fails
	ErrOrderLinesFetchFailed = errors.New("failed to get order lines")

	// ErrOrderLineCreateFailed is returned when creating order line fails
	ErrOrderLineCreateFailed = errors.New("failed to create order line")

	// ErrOrderLineUpdateFailed is returned when updating order line fails
	ErrOrderLineUpdateFailed = errors.New("failed to update order line")

	// ErrOrderLineDeleteFailed is returned when deleting order line fails
	ErrOrderLineDeleteFailed = errors.New("failed to delete order line")

	// ErrOrderConfirmFailed is returned when order confirmation fails
	ErrOrderConfirmFailed = errors.New("failed to confirm order")

	// ErrOrderProcessingFailed is returned when starting order processing fails
	ErrOrderProcessingFailed = errors.New("failed to start processing order")

	// ErrOrderFulfillFailed is returned when marking order as fulfilled fails
	ErrOrderFulfillFailed = errors.New("failed to mark order as fulfilled")

	// ErrOrderCancelFailed is returned when order cancellation fails
	ErrOrderCancelFailed = errors.New("failed to cancel order")
)
