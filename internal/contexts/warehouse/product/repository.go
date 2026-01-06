package product

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the contract for Product persistence operations.
//
// Implementation notes:
// - All methods accept context for transaction support
// - GetBy* methods return ErrProductNotFound if not found
// - List methods return empty slice if no results
// - Soft delete pattern: deleted products remain in DB with IsDeleted=true
type IRepository interface {
	// Create inserts a new product into the database.
	// Returns error if SKU already exists (unique constraint).
	Create(ctx context.Context, product *Product) error

	// GetByID retrieves a product by its UUID.
	// Returns ErrProductNotFound if product doesn't exist or is soft-deleted.
	GetByID(ctx context.Context, id uuidv7.UUID) (*Product, error)

	// GetBySKU retrieves a product by its SKU (unique identifier).
	// Returns ErrProductNotFound if product doesn't exist or is soft-deleted.
	GetBySKU(ctx context.Context, sku string) (*Product, error)

	// Update modifies an existing product.
	// Uses optimistic locking with version field.
	Update(ctx context.Context, product *Product) error

	// Delete soft-deletes a product (sets IsDeleted=true).
	// Returns ErrProductNotFound if product doesn't exist.
	Delete(ctx context.Context, id uuidv7.UUID) error

	// List retrieves products with pagination and optional filters.
	// Returns empty slice if no products match criteria.
	List(ctx context.Context, page, pageSize int) ([]*Product, error)

	// ListByCategory retrieves products by category with pagination.
	ListByCategory(ctx context.Context, category string, page, pageSize int) ([]*Product, error)

	// ListByBrand retrieves products by brand with pagination.
	ListByBrand(ctx context.Context, brand string, page, pageSize int) ([]*Product, error)

	// ListByStatus retrieves products by status with pagination.
	ListByStatus(ctx context.Context, status ProductStatus, page, pageSize int) ([]*Product, error)

	// Search performs full-text search on product name, description, SKU.
	// Returns products matching search query with pagination.
	Search(ctx context.Context, query string, page, pageSize int) ([]*Product, error)

	// Count returns total number of active (non-deleted) products.
	Count(ctx context.Context) (int, error)

	// ExistsBySKU checks if a product with given SKU exists.
	// Used for SKU uniqueness validation before creating/updating.
	ExistsBySKU(ctx context.Context, sku string) (bool, error)

	// GetByIDs retrieves multiple products by their UUIDs in a single query.
	// Returns only found products (no error for missing IDs).
	GetByIDs(ctx context.Context, ids []uuidv7.UUID) ([]*Product, error)

	// ListLowStock retrieves products below reorder point.
	// Requires integration with Inventory aggregate to check current stock.
	// Returns products where TrackInventory=true and stock <= ReorderPoint.
	ListLowStock(ctx context.Context, page, pageSize int) ([]*Product, error)
}
