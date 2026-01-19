package usecase

import (
	"context"
	"time"

	stockmovementerrors "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement"
	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/aggregate"
	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IStockMovementUseCase defines the business operations for stock movement management
type IStockMovementUseCase interface {
	// RecordMovement creates a new stock movement record
	RecordMovement(ctx context.Context, inventoryID uuidv7.UUID, movementType aggregate.MovementType, quantity, quantityBefore int, createdBy uuidv7.UUID) (*aggregate.StockMovement, error)

	// RecordReceipt records stock receipt from supplier
	RecordReceipt(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, unitCostCents int64, currencyCode string, createdBy uuidv7.UUID) (*aggregate.StockMovement, error)

	// RecordReservation records stock reservation for order
	RecordReservation(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID uuidv7.UUID, createdBy uuidv7.UUID) (*aggregate.StockMovement, error)

	// RecordCommit records stock commit on order fulfillment
	RecordCommit(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID uuidv7.UUID, createdBy uuidv7.UUID) (*aggregate.StockMovement, error)

	// RecordAdjustment records manual stock adjustment
	RecordAdjustment(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, reason string, createdBy uuidv7.UUID) (*aggregate.StockMovement, error)

	// RecordTransfer records stock transfer between locations
	RecordTransfer(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, fromWarehouse, toWarehouse uuidv7.UUID, fromLocation, toLocation *string, createdBy uuidv7.UUID) (*aggregate.StockMovement, error)

	// GetMovement retrieves movement by ID
	GetMovement(ctx context.Context, id uuidv7.UUID) (*aggregate.StockMovement, error)

	// GetMovementsByInventory retrieves all movements for inventory item
	GetMovementsByInventory(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*aggregate.StockMovement, int, error)

	// GetMovementsByReference retrieves movements by reference (order, PO, etc)
	GetMovementsByReference(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*aggregate.StockMovement, error)

	// GetMovementsByType retrieves movements by type in date range
	GetMovementsByType(ctx context.Context, movementType aggregate.MovementType, startDate, endDate time.Time, page, pageSize int) ([]*aggregate.StockMovement, int, error)

	// GetRecentMovements retrieves N most recent movements
	GetRecentMovements(ctx context.Context, limit int) ([]*aggregate.StockMovement, error)

	// GetInventorySummary calculates total in/out for inventory in date range
	GetInventorySummary(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (totalIn int, totalOut int, error error)
}

// StockMovementUseCase implements IUseCase interface
type StockMovementUseCase struct {
	repo repository.IStockMovementRepository
}

// NewStockMovementUseCase creates a new stock movement use case
func NewStockMovementUseCase(repo repository.IStockMovementRepository) IStockMovementUseCase {
	return &StockMovementUseCase{
		repo: repo,
	}
}

// RecordMovement creates a new stock movement record
func (uc *StockMovementUseCase) RecordMovement(ctx context.Context, inventoryID uuidv7.UUID, movementType aggregate.MovementType, quantity, quantityBefore int, createdBy uuidv7.UUID) (*aggregate.StockMovement, error) {
	// Create movement
	movement, err := aggregate.NewStockMovement(inventoryID, movementType, quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, stockmovementerrors.ErrStockMovementValidationFailed
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, stockmovementerrors.ErrStockMovementSaveFailed
	}

	return movement, nil
}

// RecordReceipt records stock receipt from supplier
func (uc *StockMovementUseCase) RecordReceipt(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, unitCostCents int64, currencyCode string, createdBy uuidv7.UUID) (*aggregate.StockMovement, error) {
	// Create receipt movement
	movement, err := aggregate.NewStockMovement(inventoryID, aggregate.MovementTypeReceipt, quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Set cost information
	if err := movement.SetCost(unitCostCents, currencyCode); err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Set reference
	movement.SetReference("purchase_order", nil) // Can be set later with PO ID

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, stockmovementerrors.ErrStockMovementValidationFailed
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, stockmovementerrors.ErrStockMovementSaveFailed
	}

	return movement, nil
}

// RecordReservation records stock reservation for order
func (uc *StockMovementUseCase) RecordReservation(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID uuidv7.UUID, createdBy uuidv7.UUID) (*aggregate.StockMovement, error) {
	// Create reservation movement (negative quantity)
	movement, err := aggregate.NewStockMovement(inventoryID, aggregate.MovementTypeReservation, -quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Set order reference
	movement.SetReference("order", &orderID)

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, stockmovementerrors.ErrStockMovementValidationFailed
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, stockmovementerrors.ErrStockMovementSaveFailed
	}

	return movement, nil
}

// RecordCommit records stock commit on order fulfillment
func (uc *StockMovementUseCase) RecordCommit(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, orderID uuidv7.UUID, createdBy uuidv7.UUID) (*aggregate.StockMovement, error) {
	// Create commit movement (negative quantity)
	movement, err := aggregate.NewStockMovement(inventoryID, aggregate.MovementTypeCommit, -quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Set order reference
	movement.SetReference("order", &orderID)

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, stockmovementerrors.ErrStockMovementValidationFailed
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, stockmovementerrors.ErrStockMovementSaveFailed
	}

	return movement, nil
}

// RecordAdjustment records manual stock adjustment
func (uc *StockMovementUseCase) RecordAdjustment(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, reason string, createdBy uuidv7.UUID) (*aggregate.StockMovement, error) {
	// Create adjustment movement
	movement, err := aggregate.NewStockMovement(inventoryID, aggregate.MovementTypeAdjustment, quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Set reason (required for adjustments)
	if err := movement.SetReason(reason); err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, stockmovementerrors.ErrStockMovementValidationFailed
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, stockmovementerrors.ErrStockMovementSaveFailed
	}

	return movement, nil
}

// RecordTransfer records stock transfer between locations
func (uc *StockMovementUseCase) RecordTransfer(ctx context.Context, inventoryID uuidv7.UUID, quantity, quantityBefore int, fromWarehouse, toWarehouse uuidv7.UUID, fromLocation, toLocation *string, createdBy uuidv7.UUID) (*aggregate.StockMovement, error) {
	// Create transfer movement
	movement, err := aggregate.NewStockMovement(inventoryID, aggregate.MovementTypeTransfer, quantity, quantityBefore, createdBy)
	if err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Set location information
	if err := movement.SetLocation(&fromWarehouse, &toWarehouse, fromLocation, toLocation); err != nil {
		return nil, stockmovementerrors.ErrStockMovementCreateFailed
	}

	// Validate
	if err := movement.Validate(); err != nil {
		return nil, stockmovementerrors.ErrStockMovementValidationFailed
	}

	// Persist
	if err := uc.repo.Create(ctx, movement); err != nil {
		return nil, stockmovementerrors.ErrStockMovementSaveFailed
	}

	return movement, nil
}

// GetMovement retrieves movement by ID
func (uc *StockMovementUseCase) GetMovement(ctx context.Context, id uuidv7.UUID) (*aggregate.StockMovement, error) {
	movement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return movement, nil
}

// GetMovementsByInventory retrieves all movements for inventory item
func (uc *StockMovementUseCase) GetMovementsByInventory(ctx context.Context, inventoryID uuidv7.UUID, page, pageSize int) ([]*aggregate.StockMovement, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	movements, total, err := uc.repo.GetByInventoryID(ctx, inventoryID, page, pageSize)
	if err != nil {
		return nil, 0, stockmovementerrors.ErrStockMovementListFailed
	}

	return movements, total, nil
}

// GetMovementsByReference retrieves movements by reference (order, PO, etc)
func (uc *StockMovementUseCase) GetMovementsByReference(ctx context.Context, referenceType string, referenceID uuidv7.UUID) ([]*aggregate.StockMovement, error) {
	if referenceType == "" {
		return nil, stockmovementerrors.ErrReferenceTypeRequired
	}
	if referenceID == uuidv7.Nil {
		return nil, stockmovementerrors.ErrReferenceIDRequired
	}

	movements, err := uc.repo.GetByReference(ctx, referenceType, referenceID)
	if err != nil {
		return nil, stockmovementerrors.ErrStockMovementListFailed
	}

	return movements, nil
}

// GetMovementsByType retrieves movements by type in date range
func (uc *StockMovementUseCase) GetMovementsByType(ctx context.Context, movementType aggregate.MovementType, startDate, endDate time.Time, page, pageSize int) ([]*aggregate.StockMovement, int, error) {
	if startDate.After(endDate) {
		return nil, 0, stockmovementerrors.ErrInvalidDateRange
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	movements, total, err := uc.repo.GetByType(ctx, movementType, startDate, endDate, page, pageSize)
	if err != nil {
		return nil, 0, stockmovementerrors.ErrStockMovementListFailed
	}

	return movements, total, nil
}

// GetRecentMovements retrieves N most recent movements
func (uc *StockMovementUseCase) GetRecentMovements(ctx context.Context, limit int) ([]*aggregate.StockMovement, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	movements, err := uc.repo.GetRecentMovements(ctx, limit)
	if err != nil {
		return nil, stockmovementerrors.ErrStockMovementListFailed
	}

	return movements, nil
}

// GetInventorySummary calculates total in/out for inventory in date range
func (uc *StockMovementUseCase) GetInventorySummary(ctx context.Context, inventoryID uuidv7.UUID, startDate, endDate time.Time) (totalIn int, totalOut int, error error) {
	if startDate.After(endDate) {
		return 0, 0, stockmovementerrors.ErrInvalidDateRange
	}

	totalIn, totalOut, err := uc.repo.GetSummaryByInventory(ctx, inventoryID, startDate, endDate)
	if err != nil {
		return 0, 0, stockmovementerrors.ErrStockMovementSummaryFailed
	}

	return totalIn, totalOut, nil
}
