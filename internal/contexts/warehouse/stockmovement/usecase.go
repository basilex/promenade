package stockmovement

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines the business operations for stock movement management
type IUseCase interface {
	// RecordMovement creates a new stock movement record
	RecordMovement(ctx context.Context, inventoryID uuidv7.UUID, movementType MovementType, quantity, quantityBefore int, createdBy uuidv7.UUID) (*StockMovement, error)

	// RecordReceipt records stock receipt from supplier
	RecordReceipt(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, unitCostCents int64, currencyCode string, createdBy uuidv7.UUID) (*StockMovement, error)

	// RecordReservation records stock reservation for order
	RecordReservation(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID uuidv7.UUID, createdBy uuidv7.UUID) (*StockMovement, error)

	// RecordCommit records stock commit on order fulfillment
	RecordCommit(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID uuidv7.UUID, createdBy uuidv7.UUID) (*StockMovement, error)

	// RecordAdjustment records manual stock adjustment
	RecordAdjustment(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, reason string, createdBy uuidv7.UUID) (*StockMovement, error)

	// RecordTransfer records stock transfer between locations
	RecordTransfer(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, fromWarehouse, toWarehouse uuidv7.UUID, fromLocation, toLocation *string, createdBy uuidv7.UUID) (*StockMovement, error)

	// GetMovement retrieves movement by ID
	GetMovement(ctx context.Context, id uuidv7.UUID) (*StockMovement, error)

	// GetMovementsByInventory retrieves all movements for inventory item
	GetMovementsByInventory(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*StockMovement, int, error)

	// GetMovementsByReference retrieves movements by reference (order, PO, etc)
	GetMovementsByReference(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*StockMovement, error)

	// GetMovementsByType retrieves movements by type in date range
	GetMovementsByType(ctx context.Context, movementType MovementType, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error)

	// GetRecentMovements retrieves N most recent movements
	GetRecentMovements(ctx context.Context, limit int) ([]*StockMovement, error)

	// GetInventorySummary calculates total in/out for inventory in date range
	GetInventorySummary(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (totalIn int, totalOut int, error error)
}

// useCase implements IUseCase interface
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new stock movement use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// RecordMovement creates a new stock movement record
func (uc *useCase) RecordMovement(ctx context.Context, inventoryID uuidv7.UUID, movementType MovementType, quantity, quantityBefore int, createdBy uuidv7.UUID) (*StockMovement, error) {
	// Create movement
	movement, err := NewStockMovement(inventoryID, movementType, quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create stock movement: %w", err)
	}

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, fmt.Errorf("stock movement validation failed: %w", err)
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to save stock movement: %w", err)
	}

	return movement, nil
}

// RecordReceipt records stock receipt from supplier
func (uc *useCase) RecordReceipt(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, unitCostCents int64, currencyCode string, createdBy uuidv7.UUID) (*StockMovement, error) {
	// Create receipt movement
	movement, err := NewStockMovement(inventoryID, MovementTypeReceipt, quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create receipt movement: %w", err)
	}

	// Set cost information
	if err := movement.SetCost(unitCostCents, currencyCode); err != nil {
		return nil, fmt.Errorf("failed to set cost: %w", err)
	}

	// Set reference
	movement.SetReference("purchase_order", nil) // Can be set later with PO ID

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, fmt.Errorf("receipt validation failed: %w", err)
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to save receipt: %w", err)
	}

	return movement, nil
}

// RecordReservation records stock reservation for order
func (uc *useCase) RecordReservation(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID uuidv7.UUID, createdBy uuidv7.UUID) (*StockMovement, error) {
	// Create reservation movement (negative quantity)
	movement, err := NewStockMovement(inventoryID, MovementTypeReservation, -quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create reservation movement: %w", err)
	}

	// Set order reference
	movement.SetReference("order", &orderID)

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, fmt.Errorf("reservation validation failed: %w", err)
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to save reservation: %w", err)
	}

	return movement, nil
}

// RecordCommit records stock commit on order fulfillment
func (uc *useCase) RecordCommit(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID uuidv7.UUID, createdBy uuidv7.UUID) (*StockMovement, error) {
	// Create commit movement (negative quantity)
	movement, err := NewStockMovement(inventoryID, MovementTypeCommit, -quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create commit movement: %w", err)
	}

	// Set order reference
	movement.SetReference("order", &orderID)

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, fmt.Errorf("commit validation failed: %w", err)
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to save commit: %w", err)
	}

	return movement, nil
}

// RecordAdjustment records manual stock adjustment
func (uc *useCase) RecordAdjustment(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, reason string, createdBy uuidv7.UUID) (*StockMovement, error) {
	// Create adjustment movement
	movement, err := NewStockMovement(inventoryID, MovementTypeAdjustment, quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create adjustment movement: %w", err)
	}

	// Set reason (required for adjustments)
	if err := movement.SetReason(reason); err != nil {
		return nil, fmt.Errorf("failed to set reason: %w", err)
	}

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, fmt.Errorf("adjustment validation failed: %w", err)
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to save adjustment: %w", err)
	}

	return movement, nil
}

// RecordTransfer records stock transfer between locations
func (uc *useCase) RecordTransfer(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, fromWarehouse, toWarehouse uuidv7.UUID, fromLocation, toLocation *string, createdBy uuidv7.UUID) (*StockMovement, error) {
	// Create transfer movement
	movement, err := NewStockMovement(inventoryID, MovementTypeTransfer, quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer movement: %w", err)
	}

	// Set location information
	if err := movement.SetLocation(&fromWarehouse, &toWarehouse, fromLocation, toLocation); err != nil {
		return nil, fmt.Errorf("failed to set location: %w", err)
	}

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, fmt.Errorf("transfer validation failed: %w", err)
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, fmt.Errorf("failed to save transfer: %w", err)
	}

	return movement, nil
}

// GetMovement retrieves movement by ID
func (uc *useCase) GetMovement(ctx context.Context, id uuidv7.UUID) (*StockMovement, error) {
	movement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return movement, nil
}

// GetMovementsByInventory retrieves all movements for inventory item
func (uc *useCase) GetMovementsByInventory(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*StockMovement, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	movements, total, err := uc.repo.GetByInventoryID(ctx, inventoryID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get movements by inventory: %w", err)
	}

	return movements, total, nil
}

// GetMovementsByReference retrieves movements by reference (order, PO, etc)
func (uc *useCase) GetMovementsByReference(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*StockMovement, error) {
	if referenceType == "" {
		return nil, fmt.Errorf("reference type is required")
	}
	if referenceID == uuidv7.Nil {
		return nil, fmt.Errorf("reference ID is required")
	}

	movements, err := uc.repo.GetByReference(ctx, referenceType, referenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get movements by reference: %w", err)
	}

	return movements, nil
}

// GetMovementsByType retrieves movements by type in date range
func (uc *useCase) GetMovementsByType(ctx context.Context, movementType MovementType, startDate, endDate time.Time, page, pageSize int) ([]*StockMovement, int, error) {
	if startDate.After(endDate) {
		return nil, 0, ErrInvalidDateRange
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	movements, total, err := uc.repo.GetByType(ctx, movementType, startDate, endDate, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get movements by type: %w", err)
	}

	return movements, total, nil
}

// GetRecentMovements retrieves N most recent movements
func (uc *useCase) GetRecentMovements(ctx context.Context, limit int) ([]*StockMovement, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	movements, err := uc.repo.GetRecentMovements(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent movements: %w", err)
	}

	return movements, nil
}

// GetInventorySummary calculates total in/out for inventory in date range
func (uc *useCase) GetInventorySummary(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (totalIn int, totalOut int, error error) {
	if startDate.After(endDate) {
		return 0, 0, ErrInvalidDateRange
	}

	totalIn, totalOut, err := uc.repo.GetSummaryByInventory(ctx, inventoryID, startDate, endDate)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get inventory summary: %w", err)
	}

	return totalIn, totalOut, nil
}
