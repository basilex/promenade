package stockmovement

import (
	"context"
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Domain errors for repository operations
var (
	// ErrStockMovementNotFound is returned when stock movement record is not found
	ErrStockMovementNotFound = errors.New("stock movement not found")

	// ErrInvalidPagination is returned when pagination parameters are invalid
	ErrInvalidPagination = errors.New("invalid pagination parameters")

	// ErrInvalidDateRange is returned when date range parameters are invalid
	ErrInvalidDateRange = errors.New("invalid date range - start date must be before end date")
)

// IRepository defines persistence operations for StockMovement aggregate.
//
// Design Principles:
//   - Immutable records: No Update() method (append-only log)
//   - Context-first: All methods accept context.Context for timeout/cancellation/tracing
//   - Audit trail: All movements are permanently stored (no soft delete)
//   - CQRS support: GetByInventoryID(), GetByReference() are optimized read models
//   - Performance: Indexes on inventory_id, type, reference_id, movement_date
//
// Example usage:
//
//	repo := postgres.NewStockMovementRepository(db)
//	movement, err := repo.GetByID(ctx, id)
//	if errors.Is(err, stockmovement.ErrStockMovementNotFound) {
//	    // Handle not found
//	}
type IRepository interface {
	// ==============================================================================
	// CRUD Operations (No Update - Immutable Records)
	// ==============================================================================

	// Create inserts a new stock movement record.
	// This is an append-only operation - no updates allowed.
	//
	// Example:
	//   sm, _ := stockmovement.NewStockMovement(inventoryID, MovementTypeReceipt, 100, 50, userID)
	//   err := repo.Create(ctx, sm)
	Create(ctx context.Context, movement *StockMovement) error

	// GetByID retrieves a stock movement record by UUID.
	// Returns ErrStockMovementNotFound if not found.
	//
	// Example:
	//   sm, err := repo.GetByID(ctx, id)
	//   if errors.Is(err, stockmovement.ErrStockMovementNotFound) {
	//       return fmt.Errorf("movement not found: %w", err)
	//   }
	GetByID(ctx context.Context, id uuidv7.UUID) (*StockMovement, error)

	// ==============================================================================
	// Business Queries
	// ==============================================================================

	// GetByInventoryID retrieves all movements for a specific inventory item.
	// Ordered by movement_date DESC (most recent first).
	// Returns empty slice if no movements found.
	//
	// Example:
	//   movements, err := repo.GetByInventoryID(ctx, inventoryID, 1, 50)
	GetByInventoryID(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*StockMovement, int, error)

	// GetByReference retrieves all movements linked to a specific reference (order, PO, etc).
	// Useful for tracing all stock changes related to an order.
	// Returns empty slice if no movements found.
	//
	// Example:
	//   movements, err := repo.GetByReference(ctx, "order", orderID)
	GetByReference(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*StockMovement, error)

	// GetByType retrieves all movements of a specific type in date range.
	// Useful for reporting (e.g., all receipts in a month).
	// Returns empty slice if no movements found.
	//
	// Example:
	//   receipts, err := repo.GetByType(ctx, MovementTypeReceipt, startDate, endDate, 1, 100)
	GetByType(ctx context.Context, movementType MovementType, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error)

	// GetByDateRange retrieves all movements within a date range.
	// Ordered by movement_date DESC.
	// Returns empty slice if no movements found.
	//
	// Example:
	//   movements, err := repo.GetByDateRange(ctx, startDate, endDate, 1, 50)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error)

	// ==============================================================================
	// Analytics & Reporting
	// ==============================================================================

	// GetSummaryByInventory calculates total quantity change for an inventory item.
	// Returns total positive movements (receipts) and total negative movements (commits).
	//
	// Example:
	//   totalIn, totalOut, err := repo.GetSummaryByInventory(ctx, inventoryID, startDate, endDate)
	GetSummaryByInventory(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (totalIn int, totalOut int, err error)

	// GetRecentMovements retrieves the N most recent movements across all inventory.
	// Useful for activity feed/dashboard.
	//
	// Example:
	//   recent, err := repo.GetRecentMovements(ctx, 20)
	GetRecentMovements(ctx context.Context, limit int) ([]*StockMovement, error)

	// CountByType counts movements by type within date range.
	// Returns map of MovementType -> count.
	//
	// Example:
	//   counts, err := repo.CountByType(ctx, startDate, endDate)
	//   // Returns: {receipt: 100, commit: 80, adjustment: 5, ...}
	CountByType(ctx context.Context, startDate, endDate time.Time) (map[MovementType]int, error)
}
