package inventory

import (
	"context"
	"errors"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Domain errors for repository operations
var (
	// ErrInventoryNotFound is returned when inventory item is not found
	ErrInventoryNotFound = errors.New("inventory not found")

	// ErrInventoryAlreadyExists is returned when SKU already exists
	ErrInventoryAlreadyExists = errors.New("inventory with this SKU already exists")

	// ErrVersionConflict is returned when optimistic locking fails (concurrent update)
	ErrVersionConflict = errors.New("inventory version conflict - item was modified by another process")

	// ErrInvalidPagination is returned when pagination parameters are invalid
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)

// IRepository defines persistence operations for Inventory aggregate.
//
// Design Principles:
//   - Context-first: All methods accept context.Context for timeout/cancellation/tracing
//   - Optimistic locking: Update() checks Version field to prevent concurrent modifications
//   - Soft deletes: Delete() sets DeletedAt timestamp (preserves audit history)
//   - Domain errors: Returns typed errors (ErrInventoryNotFound, ErrVersionConflict)
//   - CQRS support: GetLowStock() and GetByStatus() are denormalized read models
//
// Example usage:
//
//	repo := postgres.NewInventoryRepository(db)
//	inventory, err := repo.GetByID(ctx, id)
//	if errors.Is(err, inventory.ErrInventoryNotFound) {
//	    // Handle not found
//	}
type IRepository interface {
	// ==============================================================================
	// CRUD Operations
	// ==============================================================================

	// Create inserts a new inventory item.
	// Returns ErrInventoryAlreadyExists if SKU already exists.
	//
	// Example:
	//   inv, _ := inventory.NewInventory(productID, "SKU-001", "Product Name", "WH-01")
	//   err := repo.Create(ctx, inv)
	Create(ctx context.Context, inventory *Inventory) error

	// GetByID retrieves an inventory item by UUID.
	// Returns ErrInventoryNotFound if not found or soft-deleted.
	//
	// Example:
	//   inv, err := repo.GetByID(ctx, id)
	//   if errors.Is(err, inventory.ErrInventoryNotFound) {
	//       return fmt.Errorf("inventory not found: %w", err)
	//   }
	GetByID(ctx context.Context, id uuidv7.UUID) (*Inventory, error)

	// Update persists changes to an existing inventory item.
	// Uses optimistic locking: compares Version field in WHERE clause.
	// Returns ErrVersionConflict if item was modified by another process.
	// Increments Version field on successful update.
	//
	// Example:
	//   inv, _ := repo.GetByID(ctx, id)
	//   inv.ReceiveStock(50, 1000) // Receive 50 units at $10.00
	//   err := repo.Update(ctx, inv)
	//   if errors.Is(err, inventory.ErrVersionConflict) {
	//       // Retry with fresh data
	//   }
	Update(ctx context.Context, inventory *Inventory) error

	// Delete soft-deletes an inventory item (sets DeletedAt timestamp).
	// Preserves audit history and prevents hard deletion.
	// Returns ErrInventoryNotFound if not found or already deleted.
	//
	// Example:
	//   err := repo.Delete(ctx, id)
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ==============================================================================
	// Business Queries
	// ==============================================================================

	// GetBySKU retrieves an inventory item by unique SKU.
	// Returns ErrInventoryNotFound if not found.
	//
	// Example:
	//   inv, err := repo.GetBySKU(ctx, "SKU-12345")
	GetBySKU(ctx context.Context, sku string) (*Inventory, error)

	// GetByProductID retrieves all inventory items for a product across all warehouses.
	// Returns empty slice if product has no inventory.
	//
	// Example:
	//   items, err := repo.GetByProductID(ctx, productID)
	//   totalStock := 0
	//   for _, item := range items {
	//       totalStock += item.QuantityAvailable
	//   }
	GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*Inventory, error)

	// GetByWarehouse retrieves all inventory items in a warehouse.
	// Returns empty slice if warehouse has no inventory.
	//
	// Example:
	//   items, err := repo.GetByWarehouse(ctx, "WH-01")
	GetByWarehouse(ctx context.Context, warehouseID string) ([]*Inventory, error)

	// GetByLocation retrieves all inventory items at specific warehouse location.
	// Returns empty slice if location has no inventory.
	//
	// Example:
	//   items, err := repo.GetByLocation(ctx, "WH-01", "A1-B2")
	GetByLocation(ctx context.Context, warehouseID, locationCode string) ([]*Inventory, error)

	// List retrieves paginated inventory items.
	// Returns items, total count, and error.
	// Returns ErrInvalidPagination if limit < 1 or offset < 0.
	//
	// Example:
	//   items, total, err := repo.List(ctx, 20, 0) // First page, 20 items
	//   items, total, err := repo.List(ctx, 20, 20) // Second page
	List(ctx context.Context, limit, offset int) ([]*Inventory, int64, error)

	// ==============================================================================
	// CQRS Read Models (Denormalized Queries for Performance)
	// ==============================================================================

	// GetLowStock retrieves all active inventory items below reorder point.
	// Uses denormalized query: WHERE quantity_available <= reorder_point AND is_active = true
	// Optimized with index: idx_inventory_low_stock (warehouse_id, quantity_available)
	//
	// Example (alerts system):
	//   lowStockItems, err := repo.GetLowStock(ctx)
	//   for _, item := range lowStockItems {
	//       alertService.SendLowStockAlert(item.ProductName, item.QuantityAvailable)
	//   }
	GetLowStock(ctx context.Context) ([]*Inventory, error)

	// GetByStatus retrieves all inventory items with specific status.
	// Optimized with index: idx_inventory_status (status)
	//
	// Example:
	//   damaged, err := repo.GetByStatus(ctx, inventory.StatusDamaged)
	GetByStatus(ctx context.Context, status InventoryStatus) ([]*Inventory, error)

	// ==============================================================================
	// Batch Operations (Performance Optimization)
	// ==============================================================================

	// BulkUpdate updates multiple inventory items in a single transaction.
	// Uses optimistic locking for each item.
	// Returns ErrVersionConflict if any item has version mismatch.
	// On error, entire batch is rolled back (atomicity).
	//
	// Example (end-of-day stock adjustment):
	//   updates := []*Inventory{inv1, inv2, inv3}
	//   err := repo.BulkUpdate(ctx, updates)
	BulkUpdate(ctx context.Context, inventories []*Inventory) error
}
