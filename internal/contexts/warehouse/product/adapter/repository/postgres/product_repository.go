package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	producterrors "github.com/basilex/promenade/internal/contexts/warehouse/product"
	"github.com/basilex/promenade/internal/contexts/warehouse/product/aggregate"
	"github.com/basilex/promenade/internal/contexts/warehouse/product/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// productRepository implements repository.IProductRepository interface.
type productRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new Product repository.
func NewProductRepository(db *sqlx.DB) repository.IProductRepository {
	return &productRepository{
		db: db,
	}
}

// getExecutor returns either transaction or regular connection from context
func (r *productRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes query and scans single row
func (r *productRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *productRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning rows
func (r *productRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query without returning rows
func (r *productRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

// productRow represents database row structure.
type productRow struct {
	ID                 string                    `db:"id"`
	SKU                string                    `db:"sku"`
	Name               string                    `db:"name"`
	Brand              string                    `db:"brand"`
	Category           string                    `db:"category"`
	Description        string                    `db:"description"`
	Weight             float64                   `db:"weight"`
	Length             float64                   `db:"length"`
	Width              float64                   `db:"width"`
	Height             float64                   `db:"height"`
	Tags               jsonstore.Field[[]string] `db:"tags"` // JSON array (cross-database compatible)
	TrackInventory     bool                      `db:"track_inventory"`
	AllowBackorder     bool                      `db:"allow_backorder"`
	ReorderPoint       int                       `db:"reorder_point"`
	ReorderQuantity    int                       `db:"reorder_quantity"`
	TrackSerialNumbers bool                      `db:"track_serial_numbers"`
	TrackLotNumbers    bool                      `db:"track_lot_numbers"`
	Status             string                    `db:"status"`
	Version            int                       `db:"version"`
	IsActive           bool                      `db:"is_active"`
	IsDeleted          bool                      `db:"is_deleted"`
	CreatedAt          string                    `db:"created_at"`
	UpdatedAt          string                    `db:"updated_at"`
}

// toEntity converts database row to Product entity.
func (row *productRow) toEntity() (*aggregate.Product, error) {
	id, err := uuidv7.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid product ID: %w", err)
	}

	p := &aggregate.Product{
		SKU:         row.SKU,
		Name:        row.Name,
		Description: row.Description,
		Category:    row.Category,
		Brand:       row.Brand,
		Tags:        row.Tags,
		Weight:      row.Weight,
		Dimensions: aggregate.Dimensions{
			Length: row.Length,
			Width:  row.Width,
			Height: row.Height,
		},
		TrackInventory:     row.TrackInventory,
		AllowBackorder:     row.AllowBackorder,
		ReorderPoint:       row.ReorderPoint,
		ReorderQuantity:    row.ReorderQuantity,
		TrackSerialNumbers: row.TrackSerialNumbers,
		TrackLotNumbers:    row.TrackLotNumbers,
		Status:             aggregate.ProductStatus(row.Status),
		IsActive:           row.IsActive,
		IsDeleted:          row.IsDeleted,
	}

	// Set BaseAggregate fields directly (hydrate from database)
	p.ID = id
	p.Version = row.Version
	p.CreatedAt = parseTime(row.CreatedAt)
	p.UpdatedAt = parseTime(row.UpdatedAt)

	return p, nil
}

// parseTime converts string timestamp to time.Time.
func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

// fromEntity converts Product entity to database row.
func fromEntity(p *aggregate.Product) *productRow {
	return &productRow{
		ID:                 p.GetID().String(),
		SKU:                p.SKU,
		Name:               p.Name,
		Brand:              p.Brand,
		Category:           p.Category,
		Description:        p.Description,
		Weight:             p.Weight,
		Length:             p.Dimensions.Length,
		Width:              p.Dimensions.Width,
		Height:             p.Dimensions.Height,
		Tags:               p.Tags,
		TrackInventory:     p.TrackInventory,
		AllowBackorder:     p.AllowBackorder,
		ReorderPoint:       p.ReorderPoint,
		ReorderQuantity:    p.ReorderQuantity,
		TrackSerialNumbers: p.TrackSerialNumbers,
		TrackLotNumbers:    p.TrackLotNumbers,
		Status:             string(p.Status),
		IsActive:           p.IsActive,
		IsDeleted:          p.IsDeleted,
		Version:            p.GetVersion(),
	}
}

// Create inserts a new product into the database.
func (r *productRepository) Create(ctx context.Context, p *aggregate.Product) error {
	query := `
		INSERT INTO warehouse_products (
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17,
			$18, $19, $20,
			$21, $22, $23
		)`

	row := fromEntity(p)

	_, err := r.getExecutor(ctx).ExecContext(ctx, query,
		row.ID, row.SKU, row.Name, row.Description, row.Category, row.Brand, row.Tags,
		row.Weight, row.Length, row.Width, row.Height,
		row.TrackInventory, row.AllowBackorder, row.ReorderPoint, row.ReorderQuantity,
		row.TrackSerialNumbers, row.TrackLotNumbers,
		row.Status, row.IsActive, row.IsDeleted,
		p.GetCreatedAt(), p.GetUpdatedAt(), row.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by its UUID.
func (r *productRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Product, error) {
	query := `
		SELECT 
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		FROM warehouse_products
		WHERE id = $1 AND is_deleted = FALSE`

	var row productRow
	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, producterrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product by ID: %w", err)
	}

	return row.toEntity()
}

// GetBySKU retrieves a product by its SKU (unique identifier).
func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*aggregate.Product, error) {
	query := `
		SELECT 
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		FROM warehouse_products
		WHERE sku = $1 AND is_deleted = FALSE`

	var row productRow
	if err := r.Get(ctx, &row, query, sku); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, producterrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	return row.toEntity()
}

// Update modifies an existing product with optimistic locking.
func (r *productRepository) Update(ctx context.Context, p *aggregate.Product) error {
	query := `
		UPDATE warehouse_products
		SET 
			sku = $1, name = $2, description = $3, category = $4, brand = $5, tags = $6,
			weight = $7, length = $8, width = $9, height = $10,
			track_inventory = $11, allow_backorder = $12, reorder_point = $13, reorder_quantity = $14,
			track_serial_numbers = $15, track_lot_numbers = $16,
			status = $17, is_active = $18, is_deleted = $19,
			updated_at = $20, version = version + 1
		WHERE id = $21 AND version = $22 AND is_deleted = FALSE`

	row := fromEntity(p)

	result, err := r.getExecutor(ctx).ExecContext(ctx, query,
		row.SKU, row.Name, row.Description, row.Category, row.Brand, row.Tags,
		row.Weight, row.Length, row.Width, row.Height,
		row.TrackInventory, row.AllowBackorder, row.ReorderPoint, row.ReorderQuantity,
		row.TrackSerialNumbers, row.TrackLotNumbers,
		row.Status, row.IsActive, row.IsDeleted,
		p.GetUpdatedAt(), row.ID, row.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return producterrors.ErrProductNotFound
	}

	// Increment version in memory
	p.IncrementVersion()

	return nil
}

// Delete soft-deletes a product.
func (r *productRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE warehouse_products
		SET is_deleted = TRUE, updated_at = NOW()
		WHERE id = $1 AND is_deleted = FALSE`

	result, err := r.getExecutor(ctx).ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return producterrors.ErrProductNotFound
	}

	return nil
}

// List retrieves products with pagination.
func (r *productRepository) List(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT 
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		FROM warehouse_products
		WHERE is_deleted = FALSE
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	var rows []productRow
	if err := r.Select(ctx, &rows, query, pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	products := make([]*aggregate.Product, 0, len(rows))
	for _, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// ListByCategory retrieves products by category with pagination.
func (r *productRepository) ListByCategory(ctx context.Context, category string, page, pageSize int) ([]*aggregate.Product, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT 
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		FROM warehouse_products
		WHERE category = $1 AND is_deleted = FALSE
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`

	var rows []productRow
	if err := r.Select(ctx, &rows, query, category, pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list products by category: %w", err)
	}

	products := make([]*aggregate.Product, 0, len(rows))
	for _, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// ListByBrand retrieves products by brand with pagination.
func (r *productRepository) ListByBrand(ctx context.Context, brand string, page, pageSize int) ([]*aggregate.Product, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT 
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		FROM warehouse_products
		WHERE brand = $1 AND is_deleted = FALSE
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`

	var rows []productRow
	if err := r.Select(ctx, &rows, query, brand, pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list products by brand: %w", err)
	}

	products := make([]*aggregate.Product, 0, len(rows))
	for _, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// ListByStatus retrieves products by status with pagination.
func (r *productRepository) ListByStatus(ctx context.Context, status aggregate.ProductStatus, page, pageSize int) ([]*aggregate.Product, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT 
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		FROM warehouse_products
		WHERE status = $1 AND is_deleted = FALSE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var rows []productRow
	if err := r.Select(ctx, &rows, query, string(status), pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list products by status: %w", err)
	}

	products := make([]*aggregate.Product, 0, len(rows))
	for _, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// Search performs full-text search on product name, description, SKU.
func (r *productRepository) Search(ctx context.Context, query string, page, pageSize int) ([]*aggregate.Product, error) {
	offset := (page - 1) * pageSize
	searchPattern := "%" + query + "%"

	sql := `
		SELECT 
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		FROM warehouse_products
		WHERE (
			name ILIKE $1 OR 
			description ILIKE $1 OR 
			sku ILIKE $1 OR
			category ILIKE $1 OR
			brand ILIKE $1
		) AND is_deleted = FALSE
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`

	var rows []productRow
	if err := r.Select(ctx, &rows, sql, searchPattern, pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	products := make([]*aggregate.Product, 0, len(rows))
	for _, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// Count returns total number of active (non-deleted) products.
func (r *productRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM warehouse_products WHERE is_deleted = FALSE`

	var count int
	if err := r.Get(ctx, &count, query); err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}

// ExistsBySKU checks if a product with given SKU exists.
func (r *productRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM warehouse_products WHERE sku = $1 AND is_deleted = FALSE)`

	var exists bool
	if err := r.Get(ctx, &exists, query, sku); err != nil {
		return false, fmt.Errorf("failed to check SKU existence: %w", err)
	}

	return exists, nil
}

// GetByIDs retrieves multiple products by their UUIDs in a single query.
func (r *productRepository) GetByIDs(ctx context.Context, ids []uuidv7.UUID) ([]*aggregate.Product, error) {
	if len(ids) == 0 {
		return []*aggregate.Product{}, nil
	}

	// Convert UUIDs to strings for SQL IN clause
	idStrings := make([]interface{}, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	// Build placeholders for IN clause ($1, $2, $3, ...)
	placeholderParts := make([]string, len(ids))
	for i := range ids {
		placeholderParts[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholders := strings.Join(placeholderParts, ", ")

	query := fmt.Sprintf(`
		SELECT 
			id, sku, name, description, category, brand, tags,
			weight, length, width, height,
			track_inventory, allow_backorder, reorder_point, reorder_quantity,
			track_serial_numbers, track_lot_numbers,
			status, is_active, is_deleted,
			created_at, updated_at, version
		FROM warehouse_products
		WHERE id IN (%s) AND is_deleted = FALSE
		ORDER BY name ASC
	`, placeholders)

	var rows []productRow
	if err := r.Select(ctx, &rows, query, idStrings...); err != nil {
		return nil, fmt.Errorf("failed to get products by IDs: %w", err)
	}

	products := make([]*aggregate.Product, 0, len(rows))
	for _, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// ListLowStock retrieves products below reorder point.
// Note: This requires integration with Inventory aggregate to check current stock.
// For now, returns products where reorder_point > 0 (meaning they have reorder tracking).
func (r *productRepository) ListLowStock(ctx context.Context, page, pageSize int) ([]*aggregate.Product, error) {
	offset := (page - 1) * pageSize

	// TODO: Join with inventory table when Inventory aggregate is integrated
	query := `
		SELECT 
			p.id, p.sku, p.name, p.description, p.category, p.brand, p.tags,
			p.weight, p.length, p.width, p.height,
			p.track_inventory, p.allow_backorder, p.reorder_point, p.reorder_quantity,
			p.track_serial_numbers, p.track_lot_numbers,
			p.status, p.is_active, p.is_deleted,
			p.created_at, p.updated_at, p.version
		FROM warehouse_products p
		WHERE p.track_inventory = TRUE 
			AND p.reorder_point > 0
			AND p.is_deleted = FALSE
		ORDER BY p.reorder_point DESC
		LIMIT $1 OFFSET $2`

	var rows []productRow
	if err := r.Select(ctx, &rows, query, pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list low stock products: %w", err)
	}

	products := make([]*aggregate.Product, 0, len(rows))
	for _, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}
