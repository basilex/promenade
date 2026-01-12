package customer

import "errors"

// Validation Errors - Input validation failures
var (
	// ErrCustomerNameEmpty is returned when customer name is empty
	ErrCustomerNameEmpty = errors.New("customer name cannot be empty")

	// ErrCustomerEmailInvalid is returned when customer email is invalid
	ErrCustomerEmailInvalid = errors.New("invalid customer email")

	// ErrCustomerPhoneInvalid is returned when customer phone is invalid
	ErrCustomerPhoneInvalid = errors.New("invalid customer phone")

	// ErrCustomerSourceEmpty is returned when customer source is empty
	ErrCustomerSourceEmpty = errors.New("customer source is required")

	// ErrCustomerSalesRepEmpty is returned when assigned sales rep is empty
	ErrCustomerSalesRepEmpty = errors.New("assigned sales rep is required")

	// ErrCustomerCompanyIDEmpty is returned when company ID is required but not provided
	ErrCustomerCompanyIDEmpty = errors.New("company ID is required for B2B customer")

	// ErrCustomerUserIDEmpty is returned when user ID is empty
	ErrCustomerUserIDEmpty = errors.New("user ID cannot be empty")

	// ErrCustomerTagEmpty is returned when tag is empty
	ErrCustomerTagEmpty = errors.New("tag cannot be empty")

	// ErrCustomerChurnReasonEmpty is returned when churn reason is empty
	ErrCustomerChurnReasonEmpty = errors.New("churn reason is required")
)

// Not Found Errors - Entity lookup failures
var (
	// ErrCustomerNotFound is returned when a customer cannot be found
	ErrCustomerNotFound = errors.New("customer not found")
)

// Already Exists Errors - Uniqueness constraint violations
var (
	// ErrCustomerAlreadyExists is returned when a customer with the same email already exists
	ErrCustomerAlreadyExists = errors.New("customer already exists")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrInvalidStatusTransition is returned when an invalid status transition is attempted
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrInvalidTierTransition is returned when an invalid tier transition is attempted
	ErrInvalidTierTransition = errors.New("invalid tier transition")

	// ErrCustomerUserAlreadyLinked is returned when customer is already linked to a user
	ErrCustomerUserAlreadyLinked = errors.New("customer already linked to user")

	// ErrCustomerAssignedToSameRep is returned when customer is already assigned to the specified rep
	ErrCustomerAssignedToSameRep = errors.New("customer already assigned to this rep")

	// ErrCustomerTagAlreadyExists is returned when customer already has the specified tag
	ErrCustomerTagAlreadyExists = errors.New("customer already has this tag")

	// ErrCustomerTagNotFound is returned when customer doesn't have the specified tag
	ErrCustomerTagNotFound = errors.New("customer does not have this tag")

	// ErrCustomerNotLinkedToUser is returned when trying to operate on user link when none exists
	ErrCustomerNotLinkedToUser = errors.New("customer not linked to user")
)

// Technical Operation Errors - Infrastructure failures
var (
	// ErrCustomerCreateFailed is returned when customer creation fails
	ErrCustomerCreateFailed = errors.New("failed to create customer")

	// ErrCustomerGetFailed is returned when customer retrieval fails
	ErrCustomerGetFailed = errors.New("failed to retrieve customer")

	// ErrCustomerUpdateFailed is returned when customer update fails
	ErrCustomerUpdateFailed = errors.New("failed to update customer")

	// ErrCustomerDeleteFailed is returned when customer deletion fails
	ErrCustomerDeleteFailed = errors.New("failed to delete customer")

	// ErrCustomerListFailed is returned when customer listing fails
	ErrCustomerListFailed = errors.New("failed to list customers")

	// ErrCustomerEmailCheckFailed is returned when email existence check fails
	ErrCustomerEmailCheckFailed = errors.New("failed to check customer email existence")

	// ErrCustomerStatsFailed is returned when retrieving customer statistics fails
	ErrCustomerStatsFailed = errors.New("failed to retrieve customer statistics")
)
