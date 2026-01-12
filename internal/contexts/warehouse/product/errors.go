package product

import "errors"

// errors.go - Product Domain Errors
// Central error definitions for Product aggregate following three-tier architecture.
// Used across entity, usecase, and handler layers for consistent error handling.

// ============================================================================
// Repository Errors (Data Access Layer)
// ============================================================================
// Errors related to persistence operations and data retrieval.

var (
	// ErrProductNotFound is returned when a product cannot be found by ID or SKU.
	ErrProductNotFound = errors.New("product not found")

	// ErrCheckSKUFailed is returned when SKU uniqueness check fails due to database error.
	ErrCheckSKUFailed = errors.New("failed to check SKU uniqueness")

	// ErrCreateFailed is returned when product creation fails due to database error.
	ErrCreateFailed = errors.New("failed to create product")

	// ErrUpdateFailed is returned when product update fails due to database error.
	ErrUpdateFailed = errors.New("failed to update product")

	// ErrDeleteFailed is returned when product deletion fails due to database error.
	ErrDeleteFailed = errors.New("failed to delete product")

	// ErrListFailed is returned when product listing fails due to database error.
	ErrListFailed = errors.New("failed to list products")

	// ErrListByCategoryFailed is returned when category filtering fails.
	ErrListByCategoryFailed = errors.New("failed to list products by category")

	// ErrListByBrandFailed is returned when brand filtering fails.
	ErrListByBrandFailed = errors.New("failed to list products by brand")

	// ErrListByStatusFailed is returned when status filtering fails.
	ErrListByStatusFailed = errors.New("failed to list products by status")

	// ErrSearchFailed is returned when product search fails due to database error.
	ErrSearchFailed = errors.New("failed to search products")

	// ErrCountFailed is returned when product count query fails.
	ErrCountFailed = errors.New("failed to count products")

	// ErrListLowStockFailed is returned when low stock query fails.
	ErrListLowStockFailed = errors.New("failed to list low stock products")
)

// ============================================================================
// Business Logic Errors (Domain Rules)
// ============================================================================
// Errors representing violations of business rules and domain invariants.

var (
	// ErrProductSKURequired is returned when SKU is missing or empty.
	ErrProductSKURequired = errors.New("product SKU is required")

	// ErrProductNameRequired is returned when product name is missing or empty.
	ErrProductNameRequired = errors.New("product name is required")

	// ErrProductSKUDuplicate is returned when SKU already exists (unique constraint).
	ErrProductSKUDuplicate = errors.New("product SKU already exists")

	// ErrProductInvalidReorder is returned when reorder point >= reorder quantity.
	ErrProductInvalidReorder = errors.New("reorder point must be less than reorder quantity")

	// ErrProductInvalidWeight is returned when weight is non-positive for physical products.
	ErrProductInvalidWeight = errors.New("weight must be positive for physical products")

	// ErrProductInvalidDimension is returned when dimensions are non-positive for physical products.
	ErrProductInvalidDimension = errors.New("dimensions must be positive for physical products")

	// ErrProductDiscontinued is returned when attempting operations on discontinued products.
	ErrProductDiscontinued = errors.New("product is discontinued")

	// ErrProductAlreadyActive is returned when trying to activate an already active product.
	ErrProductAlreadyActive = errors.New("product is already active")

	// ErrProductAlreadyInactive is returned when trying to deactivate an already inactive product.
	ErrProductAlreadyInactive = errors.New("product is already inactive")

	// ErrProductNotActive is returned when trying to mark non-active product as out of stock.
	ErrProductNotActive = errors.New("only active products can be marked out of stock")

	// ErrProductNotOutOfStock is returned when trying to restock product that is not out of stock.
	ErrProductNotOutOfStock = errors.New("only out-of-stock products can be restocked")
)

// ============================================================================
// Technical Errors (System Operations)
// ============================================================================
// Errors related to operational failures in use case layer.

var (
	// ErrActivateFailed is returned when product activation fails.
	ErrActivateFailed = errors.New("failed to activate product")

	// ErrDeactivateFailed is returned when product deactivation fails.
	ErrDeactivateFailed = errors.New("failed to deactivate product")

	// ErrDiscontinueFailed is returned when product discontinuation fails.
	ErrDiscontinueFailed = errors.New("failed to discontinue product")

	// ErrUpdateInventorySettingsFailed is returned when inventory settings update fails.
	ErrUpdateInventorySettingsFailed = errors.New("failed to update inventory settings")

	// ErrSetReorderPointFailed is returned when reorder point update fails.
	ErrSetReorderPointFailed = errors.New("failed to set reorder point")

	// ErrSetPhysicalPropertiesFailed is returned when physical properties update fails.
	ErrSetPhysicalPropertiesFailed = errors.New("failed to set physical properties")
)
