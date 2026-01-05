package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// inventoryRepository implements inventory.IRepository using PostgreSQL.
type inventoryRepository struct {
	*BaseRepository
}

// NewInventoryRepository creates a new PostgreSQL inventory repository.
func NewInventoryRepository(db *sqlx.DB) inventory.IRepository {
	return &inventoryRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// inventoryRow represents database row structure for inventory table.
// Maps database columns to Go struct fields.
type inventoryRow struct {
	ID                 string         `db:"id"`
	Version            int            `db:"version"`
	ProductID          string         `db:"product_id"`
	SKU                string         `db:"sku"`
	ProductName        string         `db:"product_name"`
	QuantityOnHand     int            `db:"quantity_on_hand"`
	QuantityReserved   int            `db:"quantity_reserved"`
	QuantityCommitted  int            `db:"quantity_committed"`
	QuantityAvailable  int            `db:"quantity_available"`
	WarehouseID        string         `db:"warehouse_id"`
	LocationCode       sql.NullString `db:"location_code"`
	LocationZone       sql.NullString `db:"location_zone"`
	ReorderPoint       int            `db:"reorder_point"`
	ReorderQuantity    int            `db:"reorder_quantity"`
	LastRestocked      sql.NullTime   `db:"last_restocked"`
	Status             string         `db:"status"`
	IsActive           bool           `db:"is_active"`
	Notes              sql.NullString `db:"notes"`
	UnitCostCents      int64          `db:"unit_cost_cents"`
	CurrencyCode       string         `db:"currency_code"`
	LastUpdatedBy      sql.NullString `db:"last_updated_by"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	DeletedAt          sql.NullTime   `db:"deleted_at"`
}

// toEntity converts database row to domain entity.
func (r *inventoryRow) toEntity() (*inventory.Inventory, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid inventory id: %w", err)
	}

	productID, err := uuidv7.Parse(r.ProductID)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}

	var lastUpdatedBy uuidv7.UUID
	if r.LastUpdatedBy.Valid {
		uid, err := uuidv7.Parse(r.LastUpdatedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid last_updated_by: %w", err)
		}
		lastUpdatedBy = uid
	}

	inv := &inventory.Inventory{}
	// Hydrate entity fields directly (bypass business logic for database loading)
	inv.ID = id
	inv.Version = r.Version
	inv.ProductID = productID
	inv.SKU = r.SKU
	inv.ProductName = r.ProductName
	inv.QuantityOnHand = r.QuantityOnHand
	inv.QuantityReserved = r.QuantityReserved
	inv.QuantityCommitted = r.QuantityCommitted
	inv.QuantityAvailable = r.QuantityAvailable
	inv.WarehouseID = r.WarehouseID
	
	if r.LocationCode.Valid {
		inv.LocationCode = r.LocationCode.String
	}
	if r.LocationZone.Valid {
		inv.LocationZone = r.LocationZone.String
	}
	
	inv.ReorderPoint = r.ReorderPoint
	inv.ReorderQuantity = r.ReorderQuantity
	
	if r.LastRestocked.Valid {
		inv.LastRestocked = r.LastRestocked.Time
	}
	
	inv.Status = inventory.InventoryStatus(r.Status)
	inv.IsActive = r.IsActive
	
	if r.Notes.Valid {
		inv.Notes = r.Notes.String
	}
	
	inv.UnitCostCents = r.UnitCostCents
	inv.CurrencyCode = r.CurrencyCode
	inv.LastUpdatedBy = lastUpdatedBy
	inv.CreatedAt = r.CreatedAt
	inv.UpdatedAt = r.UpdatedAt
	
	if r.DeletedAt.Valid {
		inv.DeletedAt = &r.DeletedAt.Time
	}

	return inv, nil
}

// fromEntity converts domain entity to database row.
func fromEntity(inv *inventory.Inventory) *inventoryRow {
	row := &inventoryRow{
		ID:                inv.GetID().String(),
		Version:           inv.GetVersion(),
		ProductID:         inv.GetProductID().String(),
		SKU:               inv.GetSKU(),
		ProductName:       inv.GetProductName(),
		QuantityOnHand:    inv.GetQuantityOnHand(),
		QuantityReserved:  inv.GetQuantityReserved(),
		QuantityCommitted: inv.GetQuantityCommitted(),
		QuantityAvailable: inv.GetQuantityAvailable(),
		WarehouseID:       inv.GetWarehouseID(),
		ReorderPoint:      inv.GetReorderPoint(),
		ReorderQuantity:   inv.GetReorderQuantity(),
		Status:            string(inv.GetStatus()),
		IsActive:          inv.GetIsActive(),
		UnitCostCents:     inv.GetUnitCostCents(),
		CurrencyCode:      inv.GetCurrencyCode(),
		CreatedAt:         inv.GetCreatedAt(),
		UpdatedAt:         inv.GetUpdatedAt(),
	}

	if loc := inv.GetLocationCode(); loc != "" {
		row.LocationCode = sql.NullString{String: loc, Valid: true}
	}
	if zone := inv.GetLocationZone(); zone != "" {
		row.LocationZone = sql.NullString{String: zone, Valid: true}
	}
	if restocked := inv.GetLastRestocked(); restocked != nil {
		row.LastRestocked = sql.NullTime{Time: *restocked, Valid: true}
	}
	if notes := inv.GetNotes(); notes != "" {
		row.Notes = sql.NullString{String: notes, Valid: true}
	}
	if updatedBy := inv.GetLastUpdatedBy(); updatedBy != nil {
		row.LastUpdatedBy = sql.NullString{String: updatedBy.String(), Valid: true}
	}
	if deletedAt := inv.GetDeletedAt(); deletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *deletedAt, Valid: true}
	}

	return row
}

// ==============================================================================
// CRUD Operations
// ==============================================================================

// Create inserts a new inventory item.
func (r *inventoryRepository) Create(ctx context.Context, inv *inventory.Inventory) error {
	query := `
		INSERT INTO warehouse_inventory (
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18,
			$19, $20, $21,
			$22, $23
		)`

	row := fromEntity(inv)

	_, err := r.Exec(ctx, query,
		row.ID, row.Version, row.ProductID, row.SKU, row.ProductName,
		row.QuantityOnHand, row.QuantityReserved, row.QuantityCommitted, row.QuantityAvailable,
		row.WarehouseID, row.LocationCode, row.LocationZone,
		row.ReorderPoint, row.ReorderQuantity, row.LastRestocked,
		row.Status, row.IsActive, row.Notes,
		row.UnitCostCents, row.CurrencyCode, row.LastUpdatedBy,
		row.CreatedAt, row.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation on SKU
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return inventory.ErrInventoryAlreadyExists
			}
		}
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	return nil
}

// GetByID retrieves an inventory item by UUID.
func (r *inventoryRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*inventory.Inventory, error) {
	query := `
		SELECT 
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at, deleted_at
		FROM warehouse_inventory
		WHERE id = $1 AND deleted_at IS NULL`

	var row inventoryRow
	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inventory.ErrInventoryNotFound
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return row.toEntity()
}

// Update persists changes to an existing inventory item with optimistic locking.
func (r *inventoryRepository) Update(ctx context.Context, inv *inventory.Inventory) error {
	query := `
		UPDATE warehouse_inventory SET
			version = version + 1,
			product_name = $3,
			quantity_on_hand = $4,
			quantity_reserved = $5,
			quantity_committed = $6,
			quantity_available = $7,
			location_code = $8,
			location_zone = $9,
			reorder_point = $10,
			reorder_quantity = $11,
			last_restocked = $12,
			status = $13,
			is_active = $14,
			notes = $15,
			unit_cost_cents = $16,
			currency_code = $17,
			last_updated_by = $18,
			updated_at = $19
		WHERE id = $1 AND version = $2 AND deleted_at IS NULL`

	row := fromEntity(inv)

	result, err := r.Exec(ctx, query,
		row.ID, row.Version,
		row.ProductName,
		row.QuantityOnHand, row.QuantityReserved, row.QuantityCommitted, row.QuantityAvailable,
		row.LocationCode, row.LocationZone,
		row.ReorderPoint, row.ReorderQuantity, row.LastRestocked,
		row.Status, row.IsActive, row.Notes,
		row.UnitCostCents, row.CurrencyCode, row.LastUpdatedBy,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		// Check if item exists but version mismatch
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM warehouse_inventory WHERE id = $1 AND deleted_at IS NULL)`
		if err := r.Get(ctx, &exists, checkQuery, row.ID); err != nil {
			return fmt.Errorf("failed to check inventory existence: %w", err)
		}

		if exists {
			return inventory.ErrVersionConflict
		}
		return inventory.ErrInventoryNotFound
	}

	// Increment version in entity
	inv.SetVersion(inv.GetVersion() + 1)

	return nil
}

// Delete soft-deletes an inventory item.
func (r *inventoryRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE warehouse_inventory 
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, id.String(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete inventory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return inventory.ErrInventoryNotFound
	}

	return nil
}

// ==============================================================================
// Business Queries
// ==============================================================================

// GetBySKU retrieves an inventory item by unique SKU.
func (r *inventoryRepository) GetBySKU(ctx context.Context, sku string) (*inventory.Inventory, error) {
	query := `
		SELECT 
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at, deleted_at
		FROM warehouse_inventory
		WHERE sku = $1 AND deleted_at IS NULL`

	var row inventoryRow
	if err := r.Get(ctx, &row, query, sku); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, inventory.ErrInventoryNotFound
		}
		return nil, fmt.Errorf("failed to get inventory by SKU: %w", err)
	}

	return row.toEntity()
}

// GetByProductID retrieves all inventory items for a product across all warehouses.
func (r *inventoryRepository) GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*inventory.Inventory, error) {
	query := `
		SELECT 
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at, deleted_at
		FROM warehouse_inventory
		WHERE product_id = $1 AND deleted_at IS NULL
		ORDER BY warehouse_id, location_code`

	var rows []inventoryRow
	if err := r.Select(ctx, &rows, query, productID.String()); err != nil {
		return nil, fmt.Errorf("failed to get inventory by product ID: %w", err)
	}

	items := make([]*inventory.Inventory, 0, len(rows))
	for _, row := range rows {
		inv, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		items = append(items, inv)
	}

	return items, nil
}

// GetByWarehouse retrieves all inventory items in a warehouse.
func (r *inventoryRepository) GetByWarehouse(ctx context.Context, warehouseID string) ([]*inventory.Inventory, error) {
	query := `
		SELECT 
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at, deleted_at
		FROM warehouse_inventory
		WHERE warehouse_id = $1 AND deleted_at IS NULL
		ORDER BY location_code, product_name`

	var rows []inventoryRow
	if err := r.Select(ctx, &rows, query, warehouseID); err != nil {
		return nil, fmt.Errorf("failed to get inventory by warehouse: %w", err)
	}

	items := make([]*inventory.Inventory, 0, len(rows))
	for _, row := range rows {
		inv, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		items = append(items, inv)
	}

	return items, nil
}

// GetByLocation retrieves all inventory items at specific warehouse location.
func (r *inventoryRepository) GetByLocation(ctx context.Context, warehouseID, locationCode string) ([]*inventory.Inventory, error) {
	query := `
		SELECT 
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at, deleted_at
		FROM warehouse_inventory
		WHERE warehouse_id = $1 AND location_code = $2 AND deleted_at IS NULL
		ORDER BY created_at ASC`

	var rows []inventoryRow
	if err := r.Select(ctx, &rows, query, warehouseID, locationCode); err != nil {
		return nil, fmt.Errorf("failed to get inventory by location: %w", err)
	}

	items := make([]*inventory.Inventory, 0, len(rows))
	for _, row := range rows {
		item, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// List retrieves paginated inventory items.
func (r *inventoryRepository) List(ctx context.Context, limit, offset int) ([]*inventory.Inventory, int64, error) {
	if limit < 1 || offset < 0 {
		return nil, 0, inventory.ErrInvalidPagination
	}

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM warehouse_inventory WHERE deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory: %w", err)
	}

	// Get paginated items
	query := `
		SELECT 
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at, deleted_at
		FROM warehouse_inventory
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	var rows []inventoryRow
	if err := r.Select(ctx, &rows, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory: %w", err)
	}

	items := make([]*inventory.Inventory, 0, len(rows))
	for _, row := range rows {
		inv, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		items = append(items, inv)
	}

	return items, total, nil
}

// ==============================================================================
// CQRS Read Models
// ==============================================================================

// GetLowStock retrieves all active inventory items below reorder point.
func (r *inventoryRepository) GetLowStock(ctx context.Context) ([]*inventory.Inventory, error) {
	query := `
		SELECT 
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at, deleted_at
		FROM warehouse_inventory
		WHERE quantity_available <= reorder_point 
			AND is_active = true 
			AND deleted_at IS NULL
		ORDER BY quantity_available ASC, warehouse_id, product_name`

	var rows []inventoryRow
	if err := r.Select(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("failed to get low stock inventory: %w", err)
	}

	items := make([]*inventory.Inventory, 0, len(rows))
	for _, row := range rows {
		inv, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		items = append(items, inv)
	}

	return items, nil
}

// GetByStatus retrieves all inventory items with specific status.
func (r *inventoryRepository) GetByStatus(ctx context.Context, status inventory.InventoryStatus) ([]*inventory.Inventory, error) {
	query := `
		SELECT 
			id, version, product_id, sku, product_name,
			quantity_on_hand, quantity_reserved, quantity_committed, quantity_available,
			warehouse_id, location_code, location_zone,
			reorder_point, reorder_quantity, last_restocked,
			status, is_active, notes,
			unit_cost_cents, currency_code, last_updated_by,
			created_at, updated_at, deleted_at
		FROM warehouse_inventory
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY warehouse_id, product_name`

	var rows []inventoryRow
	if err := r.Select(ctx, &rows, query, string(status)); err != nil {
		return nil, fmt.Errorf("failed to get inventory by status: %w", err)
	}

	items := make([]*inventory.Inventory, 0, len(rows))
	for _, row := range rows {
		inv, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		items = append(items, inv)
	}

	return items, nil
}

// ==============================================================================
// Batch Operations
// ==============================================================================

// BulkUpdate updates multiple inventory items in a single transaction.
func (r *inventoryRepository) BulkUpdate(ctx context.Context, inventories []*inventory.Inventory) error {
	if len(inventories) == 0 {
		return nil
	}

	// Update each item (transaction will be managed by caller via context)
	for _, inv := range inventories {
		if err := r.Update(ctx, inv); err != nil {
			return fmt.Errorf("bulk update failed for item %s: %w", inv.GetSKU(), err)
		}
	}

	return nil
}
