package inventory

import "errors"

// Repository Errors - Data access failures
var (
	// ErrInventoryNotFound is returned when inventory record is not found
	ErrInventoryNotFound = errors.New("inventory not found")

	// ErrInventoryUnauthorized is returned when user lacks access permissions
	ErrInventoryUnauthorized = errors.New("unauthorized access to inventory")

	// ErrInventoryAlreadyExists is returned when SKU already exists (duplicate key)
	ErrInventoryAlreadyExists = errors.New("inventory with this SKU already exists")

	// ErrVersionConflict is returned when optimistic locking fails (concurrent update)
	ErrVersionConflict = errors.New("inventory version conflict - item was modified by another process")

	// ErrInvalidPagination is returned when pagination parameters are invalid
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrInventorySKUExists is returned when SKU already exists in warehouse
	ErrInventorySKUExists = errors.New("SKU already exists")

	// ErrInventorySKURequired is returned when SKU is empty
	ErrInventorySKURequired = errors.New("SKU is required")

	// ErrInventoryProductNameRequired is returned when product name is empty
	ErrInventoryProductNameRequired = errors.New("product name is required")

	// ErrInventoryWarehouseRequired is returned when warehouse ID is empty
	ErrInventoryWarehouseRequired = errors.New("warehouse ID is required")

	// ErrInventoryLocationRequired is returned when location code is empty
	ErrInventoryLocationRequired = errors.New("location code is required")

	// ErrInventoryCreatedByRequired is returned when created by user ID is nil
	ErrInventoryCreatedByRequired = errors.New("created by user ID is required")

	// ErrInventoryNil is returned when inventory pointer is nil
	ErrInventoryNil = errors.New("inventory cannot be nil")

	// ErrInventoryQuantityInvalid is returned when quantity is zero or negative
	ErrInventoryQuantityInvalid = errors.New("quantity must be greater than 0")

	// ErrInventoryUnitCostNegative is returned when unit cost is negative
	ErrInventoryUnitCostNegative = errors.New("unit cost cannot be negative")

	// ErrInventoryInsufficientStock is returned when trying to reserve more than available
	ErrInventoryInsufficientStock = errors.New("insufficient stock available")

	// ErrInventoryAlreadyDeleted is returned when operating on soft-deleted inventory
	ErrInventoryAlreadyDeleted = errors.New("inventory already deleted")

	// ErrInventoryProductIDRequired is returned when product ID is nil (UUID zero value)
	ErrInventoryProductIDRequired = errors.New("product ID is required")

	// ErrInventoryInactive is returned when trying to operate on inactive inventory
	ErrInventoryInactive = errors.New("cannot operate on inactive inventory")

	// ErrInventoryInsufficientReserved is returned when trying to release/commit more than reserved
	ErrInventoryInsufficientReserved = errors.New("insufficient reserved stock")

	// ErrInventoryNegativeStock is returned when adjustment would result in negative stock
	ErrInventoryNegativeStock = errors.New("adjustment would result in negative stock")

	// ErrInventoryAdjustmentReasonRequired is returned when adjustment reason is empty
	ErrInventoryAdjustmentReasonRequired = errors.New("adjustment reason is required")

	// ErrInventoryReorderPointNegative is returned when reorder point is negative
	ErrInventoryReorderPointNegative = errors.New("reorder point cannot be negative")

	// ErrInventoryReorderQuantityInvalid is returned when reorder quantity is zero or negative
	ErrInventoryReorderQuantityInvalid = errors.New("reorder quantity must be positive")

	// ErrInventoryInsufficientAvailable is returned when trying to mark damage on unavailable stock
	ErrInventoryInsufficientAvailable = errors.New("insufficient available stock")

	// ErrInventoryAlreadyActive is returned when trying to activate already active inventory
	ErrInventoryAlreadyActive = errors.New("inventory is already active")

	// ErrInventoryCannotDeactivateWithReservedStock is returned when trying to deactivate with reserved stock
	ErrInventoryCannotDeactivateWithReservedStock = errors.New("cannot deactivate inventory with reserved stock")

	// ErrInventoryQuantityOnHandNegative is returned when quantity on hand is negative (validation)
	ErrInventoryQuantityOnHandNegative = errors.New("quantity on hand cannot be negative")

	// ErrInventoryQuantityReservedNegative is returned when quantity reserved is negative (validation)
	ErrInventoryQuantityReservedNegative = errors.New("quantity reserved cannot be negative")

	// ErrInventoryQuantityCommittedNegative is returned when quantity committed is negative (validation)
	ErrInventoryQuantityCommittedNegative = errors.New("quantity committed cannot be negative")

	// ErrInventoryNotesTooLong is returned when notes exceed 500 characters
	ErrInventoryNotesTooLong = errors.New("notes exceed 500 characters")
)

// Technical Operation Errors - Wrapper failures
var (
	// ErrInventoryCreateFailed is returned when inventory creation fails
	ErrInventoryCreateFailed = errors.New("failed to create inventory")

	// ErrInventoryUpdateFailed is returned when inventory update fails
	ErrInventoryUpdateFailed = errors.New("failed to update inventory")

	// ErrInventoryDeleteFailed is returned when inventory deletion fails
	ErrInventoryDeleteFailed = errors.New("failed to delete inventory")

	// ErrInventoryListFailed is returned when inventory listing fails
	ErrInventoryListFailed = errors.New("failed to list inventory")

	// ErrInventorySaveFailed is returned when saving inventory to database fails
	ErrInventorySaveFailed = errors.New("failed to save inventory")
)
