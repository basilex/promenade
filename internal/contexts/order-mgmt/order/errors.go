package order

import "errors"

// ErrOrderNotFound is returned when an order is not found
var ErrOrderNotFound = errors.New("order not found")

// ErrOrderAlreadyConfirmed is returned when trying to modify a confirmed order
var ErrOrderAlreadyConfirmed = errors.New("order already confirmed")

// ErrOrderAlreadyCancelled is returned when trying to modify a cancelled order
var ErrOrderAlreadyCancelled = errors.New("order already cancelled")

// ErrOrderEmpty is returned when trying to create an order without line items
var ErrOrderEmpty = errors.New("order must have at least one line item")

// ErrInvalidOrderStatus is returned when an invalid status transition is attempted
var ErrInvalidOrderStatus = errors.New("invalid order status transition")

// ErrLineNotFound is returned when an order line is not found
var ErrLineNotFound = errors.New("order line not found")

// ErrOrderLineNotFound is an alias for ErrLineNotFound for consistency with naming conventions
var ErrOrderLineNotFound = ErrLineNotFound

// ErrInvalidQuantity is returned when quantity is less than or equal to zero
var ErrInvalidQuantity = errors.New("quantity must be greater than zero")

// ErrInvalidPrice is returned when price is negative
var ErrInvalidPrice = errors.New("price cannot be negative")

// ErrOrderNotConfirmed is returned when trying to process an order that is not confirmed
var ErrOrderNotConfirmed = errors.New("order must be confirmed before processing")

// ErrOrderNotProcessing is returned when trying to fulfill an order that is not being processed
var ErrOrderNotProcessing = errors.New("order must be in processing status to be fulfilled")

// ErrOrderAlreadyFulfilled is returned when trying to modify a fulfilled order
var ErrOrderAlreadyFulfilled = errors.New("order is already fulfilled")

