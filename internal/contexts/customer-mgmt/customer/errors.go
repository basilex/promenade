package customer

import "errors"

// Domain errors for customer context
var (
	// ErrCustomerNotFound is returned when a customer cannot be found
	ErrCustomerNotFound = errors.New("customer not found")

	// ErrCustomerAlreadyExists is returned when a customer with the same email already exists
	ErrCustomerAlreadyExists = errors.New("customer already exists")

	// ErrInvalidStatusTransition is returned when an invalid status transition is attempted
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrInvalidTierTransition is returned when an invalid tier transition is attempted
	ErrInvalidTierTransition = errors.New("invalid tier transition")

	// ErrCustomerNotLinkedToUser is returned when trying to operate on user link when none exists
	ErrCustomerNotLinkedToUser = errors.New("customer not linked to user")
)
