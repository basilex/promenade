package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// stockMovementRepository implements stockmovement.IRepository using PostgreSQL.
type stockMovementRepository struct {
	*BaseRepository
}

// NewStockMovementRepository creates a new PostgreSQL stock movement repository.
func NewStockMovementRepository(db *sqlx.DB) stockmovement.IRepository {
	return &stockMovementRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// stockMovementRow represents database row structure for stock_movements table.
type stockMovementRow struct {
	ID                 string         `db:"id"`
	InventoryID        string         `db:"inventory_id"`
	Type               string         `db:"type"`
	Quantity           int            `db:"quantity"`
	QuantityBeforeMove int            `db:"quantity_before_move"`
	QuantityAfterMove  int            `db:"quantity_after_move"`
	FromWarehouseID    sql.NullString `db:"from_warehouse_id"`
	FromLocationCode   sql.NullString `db:"from_location_code"`
	ToWarehouseID      sql.NullString `db:"to_warehouse_id"`
	ToLocationCode     sql.NullString `db:"to_location_code"`
	ReferenceType      sql.NullString `db:"reference_type"`
	ReferenceID        sql.NullString `db:"reference_id"`
	UnitCostCents      sql.NullInt64  `db:"unit_cost_cents"`
	TotalCostCents     sql.NullInt64  `db:"total_cost_cents"`
	CurrencyCode       sql.NullString `db:"currency_code"`
	Reason             sql.NullString `db:"reason"`
	Notes              sql.NullString `db:"notes"`
	CreatedBy          string         `db:"created_by"`
	MovementDate       time.Time      `db:"movement_date"`
	CreatedAt          time.Time      `db:"created_at"`
}

// toEntity converts database row to domain entity.
func (r *stockMovementRow) toEntity() (*stockmovement.StockMovement, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid movement id: %w", err)
	}

	inventoryID, err := uuidv7.Parse(r.InventoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid inventory id: %w", err)
	}

	createdBy, err := uuidv7.Parse(r.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid created_by: %w", err)
	}

	// Parse optional reference ID
	var referenceID *uuidv7.UUID
	if r.ReferenceID.Valid && r.ReferenceID.String != "" {
		refID, err := uuidv7.Parse(r.ReferenceID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid reference_id: %w", err)
		}
		referenceID = &refID
	}

	// Create entity (hydrate directly to bypass business logic)
	sm := &stockmovement.StockMovement{}
	sm.ID = id
	sm.InventoryID = inventoryID
	sm.Type = stockmovement.MovementType(r.Type)
	sm.Quantity = r.Quantity
	sm.QuantityBeforeMove = r.QuantityBeforeMove
	sm.QuantityAfterMove = r.QuantityAfterMove

	if r.FromWarehouseID.Valid {
		whID, err := uuidv7.Parse(r.FromWarehouseID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid from_warehouse_id: %w", err)
		}
		sm.FromWarehouseID = &whID
	}

	if r.FromLocationCode.Valid {
		loc := r.FromLocationCode.String
		sm.FromLocationCode = &loc
	}

	if r.ToWarehouseID.Valid {
		whID, err := uuidv7.Parse(r.ToWarehouseID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid to_warehouse_id: %w", err)
		}
		sm.ToWarehouseID = &whID
	}

	if r.ToLocationCode.Valid {
		loc := r.ToLocationCode.String
		sm.ToLocationCode = &loc
	}

	if r.ReferenceType.Valid {
		sm.ReferenceType = r.ReferenceType.String
	}

	sm.ReferenceID = referenceID

	if r.UnitCostCents.Valid {
		cost := r.UnitCostCents.Int64
		sm.UnitCostCents = &cost
	}

	if r.TotalCostCents.Valid {
		cost := r.TotalCostCents.Int64
		sm.TotalCostCents = &cost
	}

	if r.CurrencyCode.Valid {
		currency := r.CurrencyCode.String
		sm.CurrencyCode = &currency
	}

	if r.Reason.Valid {
		sm.Reason = r.Reason.String
	}

	if r.Notes.Valid {
		sm.Notes = r.Notes.String
	}

	sm.CreatedBy = createdBy
	sm.MovementDate = r.MovementDate
	sm.CreatedAt = r.CreatedAt

	return sm, nil
}

// fromEntity converts domain entity to database row.
func fromEntity(sm *stockmovement.StockMovement) *stockMovementRow {
	row := &stockMovementRow{
		ID:                 sm.ID.String(),
		InventoryID:        sm.InventoryID.String(),
		Type:               string(sm.Type),
		Quantity:           sm.Quantity,
		QuantityBeforeMove: sm.QuantityBeforeMove,
		QuantityAfterMove:  sm.QuantityAfterMove,
		CreatedBy:          sm.CreatedBy.String(),
		MovementDate:       sm.MovementDate,
		CreatedAt:          sm.CreatedAt,
	}

	if sm.FromWarehouseID != nil {
		row.FromWarehouseID = sql.NullString{String: sm.FromWarehouseID.String(), Valid: true}
	}

	if sm.FromLocationCode != nil {
		row.FromLocationCode = sql.NullString{String: *sm.FromLocationCode, Valid: true}
	}

	if sm.ToWarehouseID != nil {
		row.ToWarehouseID = sql.NullString{String: sm.ToWarehouseID.String(), Valid: true}
	}

	if sm.ToLocationCode != nil {
		row.ToLocationCode = sql.NullString{String: *sm.ToLocationCode, Valid: true}
	}

	if sm.ReferenceType != "" {
		row.ReferenceType = sql.NullString{String: sm.ReferenceType, Valid: true}
	}

	if sm.ReferenceID != nil {
		row.ReferenceID = sql.NullString{String: sm.ReferenceID.String(), Valid: true}
	}

	if sm.Reason != "" {
		row.Reason = sql.NullString{String: sm.Reason, Valid: true}
	}

	if sm.Notes != "" {
		row.Notes = sql.NullString{String: sm.Notes, Valid: true}
	}

	if sm.UnitCostCents != nil {
		row.UnitCostCents = sql.NullInt64{Int64: *sm.UnitCostCents, Valid: true}
	}

	if sm.TotalCostCents != nil {
		row.TotalCostCents = sql.NullInt64{Int64: *sm.TotalCostCents, Valid: true}
	}

	if sm.CurrencyCode != nil {
		row.CurrencyCode = sql.NullString{String: *sm.CurrencyCode, Valid: true}
	}

	return row
}

// Create inserts a new stock movement into database.
func (r *stockMovementRepository) Create(ctx context.Context, movement *stockmovement.StockMovement) error {
	if movement == nil {
		return fmt.Errorf("movement cannot be nil")
	}

	query := `
		INSERT INTO warehouse_stock_movements (
			id, inventory_id, type, quantity, quantity_before_move, quantity_after_move,
			from_warehouse_id, from_location_code, to_warehouse_id, to_location_code,
			reference_type, reference_id,
			unit_cost_cents, total_cost_cents, currency_code, reason, notes,
			created_by, movement_date, created_at
		) VALUES (
			:id, :inventory_id, :type, :quantity, :quantity_before_move, :quantity_after_move,
			:from_warehouse_id, :from_location_code, :to_warehouse_id, :to_location_code,
			:reference_type, :reference_id,
			:unit_cost_cents, :total_cost_cents, :currency_code, :reason, :notes,
			:created_by, :movement_date, :created_at
		)
	`

	row := fromEntity(movement)
	_, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to create stock movement: %w", err)
	}

	return nil
}

// GetByID retrieves a stock movement by ID.
func (r *stockMovementRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*stockmovement.StockMovement, error) {
	var row stockMovementRow
	query := `
		SELECT id, inventory_id, type, quantity, quantity_before_move, quantity_after_move,
		       from_warehouse_id, from_location_code, to_warehouse_id, to_location_code,
		       reference_type, reference_id,
		       unit_cost_cents, total_cost_cents, currency_code, reason, notes,
		       created_by, movement_date, created_at
		FROM warehouse_stock_movements
		WHERE id = $1
	`

	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, stockmovement.ErrStockMovementNotFound
		}
		return nil, fmt.Errorf("failed to get stock movement: %w", err)
	}

	return row.toEntity()
}

// GetByInventoryID retrieves all movements for an inventory with pagination.
func (r *stockMovementRepository) GetByInventoryID(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*stockmovement.StockMovement, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get total count first
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM warehouse_stock_movements
		WHERE inventory_id = $1
	`
	if err := r.Get(ctx, &totalCount, countQuery, inventoryID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to count movements by inventory: %w", err)
	}

	// Get paginated results
	var rows []stockMovementRow
	query := `
		SELECT id, inventory_id, type, quantity, quantity_before_move, quantity_after_move,
		       from_warehouse_id, from_location_code, to_warehouse_id, to_location_code,
		       reference_type, reference_id,
		       unit_cost_cents, total_cost_cents, currency_code, reason, notes,
		       created_by, movement_date, created_at
		FROM warehouse_stock_movements
		WHERE inventory_id = $1
		ORDER BY movement_date DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	if err := r.Select(ctx, &rows, query, inventoryID.String(), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to get movements by inventory: %w", err)
	}

	movements := make([]*stockmovement.StockMovement, 0, len(rows))
	for _, row := range rows {
		movement, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		movements = append(movements, movement)
	}

	return movements, totalCount, nil
}

// GetByReference retrieves movements by reference (no pagination).
func (r *stockMovementRepository) GetByReference(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*stockmovement.StockMovement, error) {
	var rows []stockMovementRow
	query := `
		SELECT id, inventory_id, type, quantity, quantity_before_move, quantity_after_move,
		       from_warehouse_id, from_location_code, to_warehouse_id, to_location_code,
		       reference_type, reference_id,
		       unit_cost_cents, total_cost_cents, currency_code, reason, notes,
		       created_by, movement_date, created_at
		FROM warehouse_stock_movements
		WHERE reference_type = $1 AND reference_id = $2
		ORDER BY movement_date DESC, created_at DESC
	`

	if err := r.Select(ctx, &rows, query, referenceType, referenceID.String()); err != nil {
		return nil, fmt.Errorf("failed to get movements by reference: %w", err)
	}

	movements := make([]*stockmovement.StockMovement, 0, len(rows))
	for _, row := range rows {
		movement, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

// GetByType retrieves movements by type within date range with pagination.
func (r *stockMovementRepository) GetByType(ctx context.Context, movementType stockmovement.MovementType, startDate, endDate time.Time, page, pageSize int) ([]*stockmovement.StockMovement, int, error) {
	if startDate.After(endDate) {
		return nil, 0, stockmovement.ErrInvalidDateRange
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get total count first
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM warehouse_stock_movements
		WHERE type = $1 AND movement_date >= $2 AND movement_date <= $3
	`
	if err := r.Get(ctx, &totalCount, countQuery, string(movementType), startDate, endDate); err != nil {
		return nil, 0, fmt.Errorf("failed to count movements by type: %w", err)
	}

	// Get paginated results
	var rows []stockMovementRow
	query := `
		SELECT id, inventory_id, type, quantity, quantity_before_move, quantity_after_move,
		       from_warehouse_id, from_location_code, to_warehouse_id, to_location_code,
		       reference_type, reference_id,
		       unit_cost_cents, total_cost_cents, currency_code, reason, notes,
		       created_by, movement_date, created_at
		FROM warehouse_stock_movements
		WHERE type = $1 AND movement_date >= $2 AND movement_date <= $3
		ORDER BY movement_date DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`

	if err := r.Select(ctx, &rows, query, string(movementType), startDate, endDate, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to get movements by type: %w", err)
	}

	movements := make([]*stockmovement.StockMovement, 0, len(rows))
	for _, row := range rows {
		movement, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		movements = append(movements, movement)
	}

	return movements, totalCount, nil
}

// GetByDateRange retrieves movements within date range with pagination and total count.
func (r *stockMovementRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*stockmovement.StockMovement, int, error) {
	if startDate.After(endDate) {
		return nil, 0, stockmovement.ErrInvalidDateRange
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get total count first
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM warehouse_stock_movements
		WHERE movement_date >= $1 AND movement_date <= $2
	`
	if err := r.Get(ctx, &totalCount, countQuery, startDate, endDate); err != nil {
		return nil, 0, fmt.Errorf("failed to count movements by date range: %w", err)
	}

	// Get paginated results
	var rows []stockMovementRow
	query := `
		SELECT id, inventory_id, type, quantity, quantity_before_move, quantity_after_move,
		       from_warehouse_id, from_location_code, to_warehouse_id, to_location_code,
		       reference_type, reference_id,
		       unit_cost_cents, total_cost_cents, currency_code, reason, notes,
		       created_by, movement_date, created_at
		FROM warehouse_stock_movements
		WHERE movement_date >= $1 AND movement_date <= $2
		ORDER BY movement_date DESC, created_at DESC
		LIMIT $3 OFFSET $4
	`

	if err := r.Select(ctx, &rows, query, startDate, endDate, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to get movements by date range: %w", err)
	}

	movements := make([]*stockmovement.StockMovement, 0, len(rows))
	for _, row := range rows {
		movement, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		movements = append(movements, movement)
	}

	return movements, totalCount, nil
}

// GetSummaryByInventory returns summary of movements for an inventory within date range.
func (r *stockMovementRepository) GetSummaryByInventory(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (totalIn int, totalOut int, err error) {
	if startDate.After(endDate) {
		return 0, 0, stockmovement.ErrInvalidDateRange
	}

	query := `
		SELECT 
			SUM(CASE WHEN quantity > 0 THEN quantity ELSE 0 END) as total_in,
			SUM(CASE WHEN quantity < 0 THEN ABS(quantity) ELSE 0 END) as total_out
		FROM warehouse_stock_movements
		WHERE inventory_id = $1 AND movement_date >= $2 AND movement_date <= $3
	`

	var result struct {
		TotalIn  sql.NullInt64 `db:"total_in"`
		TotalOut sql.NullInt64 `db:"total_out"`
	}

	if err := r.Get(ctx, &result, query, inventoryID.String(), startDate, endDate); err != nil {
		return 0, 0, fmt.Errorf("failed to get movement summary: %w", err)
	}

	totalIn = int(result.TotalIn.Int64)
	totalOut = int(result.TotalOut.Int64)

	return totalIn, totalOut, nil
}

// GetRecentMovements retrieves most recent movements across all inventory.
func (r *stockMovementRepository) GetRecentMovements(ctx context.Context, limit int) ([]*stockmovement.StockMovement, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var rows []stockMovementRow
	query := `
		SELECT id, inventory_id, type, quantity, quantity_before_move, quantity_after_move,
		       from_warehouse_id, from_location_code, to_warehouse_id, to_location_code,
		       reference_type, reference_id,
		       unit_cost_cents, total_cost_cents, currency_code, reason, notes,
		       created_by, movement_date, created_at
		FROM warehouse_stock_movements
		ORDER BY movement_date DESC, created_at DESC
		LIMIT $1
	`

	if err := r.Select(ctx, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("failed to get recent movements: %w", err)
	}

	movements := make([]*stockmovement.StockMovement, 0, len(rows))
	for _, row := range rows {
		movement, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

// CountByType returns count of movements by type within date range.
func (r *stockMovementRepository) CountByType(ctx context.Context, startDate, endDate time.Time) (map[stockmovement.MovementType]int, error) {
	if startDate.After(endDate) {
		return nil, stockmovement.ErrInvalidDateRange
	}

	query := `
		SELECT type, COUNT(*) as count
		FROM warehouse_stock_movements
		WHERE movement_date >= $1 AND movement_date <= $2
		GROUP BY type
	`

	var results []struct {
		Type  string `db:"type"`
		Count int    `db:"count"`
	}

	if err := r.Select(ctx, &results, query, startDate, endDate); err != nil {
		return nil, fmt.Errorf("failed to count movements by type: %w", err)
	}

	counts := make(map[stockmovement.MovementType]int)
	for _, result := range results {
		counts[stockmovement.MovementType(result.Type)] = result.Count
	}

	return counts, nil
}
