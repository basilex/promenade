package stockmovement

import "errors"

// ============================================================================
// Domain Errors for StockMovement Aggregate
// ============================================================================
// Three-tier error architecture following unified error handling standard:
// 1. Repository layer errors (data access)
// 2. Business logic errors (domain rules)
// 3. Technical operation errors (system failures)
// ============================================================================

// ============================================================================
// Repository Layer Errors
// ============================================================================

var (
	// ErrStockMovementNotFound is returned when a stock movement cannot be found by ID
	ErrStockMovementNotFound = errors.New("stock movement not found")

	// ErrInvalidPagination is returned when pagination parameters are invalid
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)

// ============================================================================
// Business Logic Errors
// ============================================================================

var (
	// ErrStockMovementNil is returned when a nil stock movement is provided
	ErrStockMovementNil = errors.New("stock movement cannot be nil")

	// ErrStockMovementValidationFailed is returned when stock movement validation fails
	ErrStockMovementValidationFailed = errors.New("stock movement validation failed")

	// ErrReferenceTypeRequired is returned when reference type is missing
	ErrReferenceTypeRequired = errors.New("reference type is required")

	// ErrReferenceIDRequired is returned when reference ID is missing
	ErrReferenceIDRequired = errors.New("reference ID is required")

	// ErrInvalidDateRange is returned when date range is invalid (start after end)
	ErrInvalidDateRange = errors.New("invalid date range - start date must be before end date")

	// Entity Validation Errors
	// ErrInventoryIDRequired is returned when inventory ID is missing
	ErrInventoryIDRequired = errors.New("inventory ID is required")

	// ErrQuantityCannotBeZero is returned when quantity is zero
	ErrQuantityCannotBeZero = errors.New("quantity cannot be zero")

	// ErrCreatedByRequired is returned when created_by field is missing
	ErrCreatedByRequired = errors.New("created_by is required")

	// ErrTransferRequiresBothWarehouses is returned when transfer lacks source/destination
	ErrTransferRequiresBothWarehouses = errors.New("transfer requires both source and destination warehouses")

	// ErrNegativeUnitCost is returned when unit cost is negative
	ErrNegativeUnitCost = errors.New("unit cost cannot be negative")

	// ErrCurrencyCodeRequired is returned when currency code is missing with cost
	ErrCurrencyCodeRequired = errors.New("currency code is required when setting cost")

	// ErrReasonCannotBeEmpty is returned when reason is empty
	ErrReasonCannotBeEmpty = errors.New("reason cannot be empty")

	// ErrReasonTooLong is returned when reason exceeds 500 characters
	ErrReasonTooLong = errors.New("reason exceeds maximum length of 500 characters")

	// ErrNotesTooLong is returned when notes exceed 500 characters
	ErrNotesTooLong = errors.New("notes exceed maximum length of 500 characters")

	// ErrAdjustmentRequiresReason is returned when adjustment movement has no reason
	ErrAdjustmentRequiresReason = errors.New("adjustment movements require a reason")

	// ErrInvalidMovementType is returned when movement type is not valid
	ErrInvalidMovementType = errors.New("invalid movement type")
)

// ============================================================================
// Technical Operation Errors
// ============================================================================

var (
	// ErrStockMovementCreateFailed is returned when creating a stock movement fails
	ErrStockMovementCreateFailed = errors.New("failed to create stock movement")

	// ErrStockMovementSaveFailed is returned when saving a stock movement fails
	ErrStockMovementSaveFailed = errors.New("failed to save stock movement")

	// ErrStockMovementGetFailed is returned when retrieving a stock movement fails
	ErrStockMovementGetFailed = errors.New("failed to get stock movement")

	// ErrStockMovementListFailed is returned when listing stock movements fails
	ErrStockMovementListFailed = errors.New("failed to list stock movements")

	// ErrStockMovementCountFailed is returned when counting stock movements fails
	ErrStockMovementCountFailed = errors.New("failed to count stock movements")

	// ErrStockMovementSummaryFailed is returned when getting movement summary fails
	ErrStockMovementSummaryFailed = errors.New("failed to get movement summary")
)
