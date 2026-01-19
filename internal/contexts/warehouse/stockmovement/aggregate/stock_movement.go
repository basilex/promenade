package aggregate

import (
	"time"

	stockmovementerrors "github.com/basilex/promenade/internal/contexts/warehouse/stockmovement"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MovementType represents the type of stock movement
type MovementType string

const (
	// MovementTypeReceipt - stock received from supplier
	MovementTypeReceipt MovementType = "receipt"
	// MovementTypeReservation - stock reserved for order
	MovementTypeReservation MovementType = "reservation"
	// MovementTypeReservationRelease - reservation cancelled/released
	MovementTypeReservationRelease MovementType = "reservation_release"
	// MovementTypeCommit - stock committed to fulfilled order
	MovementTypeCommit MovementType = "commit"
	// MovementTypeAdjustment - manual stock adjustment (audit/damage/etc)
	MovementTypeAdjustment MovementType = "adjustment"
	// MovementTypeTransfer - transfer between locations
	MovementTypeTransfer MovementType = "transfer"
	// MovementTypeDamage - stock marked as damaged
	MovementTypeDamage MovementType = "damage"
	// MovementTypeReturn - customer return
	MovementTypeReturn MovementType = "return"
)

// StockMovement is an aggregate root for tracking all inventory movements
//
// # It provides complete audit trail for stock changes
//
// Design principles:
// - Immutable records (append-only log)
// - Event sourcing ready
// - Full audit trail for compliance
// - Performance optimized with indexes
type StockMovement struct {
	aggregate.BaseAggregate

	// Movement identification
	InventoryID uuidv7.UUID  `db:"inventory_id"` // Link to inventory record
	Type        MovementType `db:"type"`         // Type of movement

	// Quantity tracking
	Quantity           int `db:"quantity"`             // Quantity moved (positive or negative)
	QuantityBeforeMove int `db:"quantity_before_move"` // Snapshot before movement
	QuantityAfterMove  int `db:"quantity_after_move"`  // Snapshot after movement

	// Location tracking (for transfers)
	FromWarehouseID  *uuidv7.UUID `db:"from_warehouse_id"`  // Source warehouse (nullable)
	FromLocationCode *string      `db:"from_location_code"` // Source location (nullable)
	ToWarehouseID    *uuidv7.UUID `db:"to_warehouse_id"`    // Destination warehouse (nullable)
	ToLocationCode   *string      `db:"to_location_code"`   // Destination location (nullable)

	// Reference tracking
	ReferenceType string       `db:"reference_type"` // Type of reference (order, po, adjustment, etc)
	ReferenceID   *uuidv7.UUID `db:"reference_id"`   // Order ID, PO ID, etc (nullable)

	// Business context
	Reason    string      `db:"reason"`     // Why this movement happened (required for adjustments)
	Notes     string      `db:"notes"`      // Additional context (max 500 chars)
	CreatedBy uuidv7.UUID `db:"created_by"` // User who initiated movement

	// Cost tracking (for valuation)
	UnitCostCents  *int64  `db:"unit_cost_cents"`  // Cost per unit in cents (nullable)
	TotalCostCents *int64  `db:"total_cost_cents"` // Total cost of movement (nullable)
	CurrencyCode   *string `db:"currency_code"`    // Currency (e.g., "USD") (nullable)

	// Timestamp
	MovementDate time.Time `db:"movement_date"` // When movement occurred
}

// NewStockMovement creates a new stock movement record
//
// Modern approach: Domain-driven validation in entity, not database
func NewStockMovement(
	inventoryID uuidv7.UUID,
	movementType MovementType,
	quantity int,
	quantityBefore int,
	createdBy uuidv7.UUID,
) (*StockMovement, error) {
	// Validation
	if inventoryID == uuidv7.Nil {
		return nil, stockmovementerrors.ErrInventoryIDRequired
	}

	if quantity == 0 {
		return nil, stockmovementerrors.ErrQuantityCannotBeZero
	}

	if createdBy == uuidv7.Nil {
		return nil, stockmovementerrors.ErrCreatedByRequired
	}

	if err := validateMovementType(movementType); err != nil {
		return nil, err
	}

	quantityAfter := quantityBefore + quantity

	return &StockMovement{
		BaseAggregate:      aggregate.NewBaseAggregate(),
		InventoryID:        inventoryID,
		Type:               movementType,
		Quantity:           quantity,
		QuantityBeforeMove: quantityBefore,
		QuantityAfterMove:  quantityAfter,
		CreatedBy:          createdBy,
		MovementDate:       time.Now().UTC(),
		ReferenceType:      "", // Set by business logic methods
	}, nil
}

// SetReference sets the reference information for this movement
func (sm *StockMovement) SetReference(referenceType string, referenceID *uuidv7.UUID) {
	sm.ReferenceType = referenceType
	sm.ReferenceID = referenceID
	sm.Touch()
}

// SetLocation sets the location information for this movement
func (sm *StockMovement) SetLocation(fromWarehouse, toWarehouse *uuidv7.UUID, fromLocation, toLocation *string) error {
	// For transfers, both from and to must be set
	if sm.Type == MovementTypeTransfer {
		if fromWarehouse == nil || toWarehouse == nil {
			return stockmovementerrors.ErrTransferRequiresBothWarehouses
		}
	}

	sm.FromWarehouseID = fromWarehouse
	sm.FromLocationCode = fromLocation
	sm.ToWarehouseID = toWarehouse
	sm.ToLocationCode = toLocation
	sm.Touch()
	return nil
}

// SetCost sets the cost information for this movement
func (sm *StockMovement) SetCost(unitCostCents int64, currencyCode string) error {
	if unitCostCents < 0 {
		return stockmovementerrors.ErrNegativeUnitCost
	}

	if currencyCode == "" {
		return stockmovementerrors.ErrCurrencyCodeRequired
	}

	totalCost := unitCostCents * int64(abs(sm.Quantity))

	sm.UnitCostCents = &unitCostCents
	sm.TotalCostCents = &totalCost
	sm.CurrencyCode = &currencyCode
	sm.Touch()
	return nil
}

// SetReason sets the reason for this movement (required for adjustments)
func (sm *StockMovement) SetReason(reason string) error {
	if reason == "" {
		return stockmovementerrors.ErrReasonCannotBeEmpty
	}

	if len(reason) > 500 {
		return stockmovementerrors.ErrReasonTooLong
	}

	sm.Reason = reason
	sm.Touch()
	return nil
}

// SetNotes sets additional notes for this movement
func (sm *StockMovement) SetNotes(notes string) error {
	if len(notes) > 500 {
		return stockmovementerrors.ErrNotesTooLong
	}

	sm.Notes = notes
	sm.Touch()
	return nil
}

// IsAdjustment checks if this is a manual adjustment
func (sm *StockMovement) IsAdjustment() bool {
	return sm.Type == MovementTypeAdjustment
}

// IsTransfer checks if this is a location transfer
func (sm *StockMovement) IsTransfer() bool {
	return sm.Type == MovementTypeTransfer
}

// IsReceiptOrReturn checks if this increases stock
func (sm *StockMovement) IsReceiptOrReturn() bool {
	return sm.Type == MovementTypeReceipt || sm.Type == MovementTypeReturn
}

// GetImpact returns the impact on available stock (+/-)
func (sm *StockMovement) GetImpact() int {
	return sm.Quantity
}

// Validate performs business rule validation
func (sm *StockMovement) Validate() error {
	if sm.InventoryID == uuidv7.Nil {
		return stockmovementerrors.ErrInventoryIDRequired
	}

	if sm.Quantity == 0 {
		return stockmovementerrors.ErrQuantityCannotBeZero
	}

	if sm.CreatedBy == uuidv7.Nil {
		return stockmovementerrors.ErrCreatedByRequired
	}

	// Validate movement type
	if err := validateMovementType(sm.Type); err != nil {
		return err
	}

	// Adjustments require reason
	if sm.Type == MovementTypeAdjustment && sm.Reason == "" {
		return stockmovementerrors.ErrAdjustmentRequiresReason
	}

	// Transfers require both from and to locations
	if sm.Type == MovementTypeTransfer {
		if sm.FromWarehouseID == nil || sm.ToWarehouseID == nil {
			return stockmovementerrors.ErrTransferRequiresBothWarehouses
		}
	}

	// Notes length validation
	if len(sm.Notes) > 500 {
		return stockmovementerrors.ErrNotesTooLong
	}

	if len(sm.Reason) > 500 {
		return stockmovementerrors.ErrReasonTooLong
	}

	return nil
}

// Helper functions

func validateMovementType(movementType MovementType) error {
	validTypes := map[MovementType]bool{
		MovementTypeReceipt:            true,
		MovementTypeReservation:        true,
		MovementTypeReservationRelease: true,
		MovementTypeCommit:             true,
		MovementTypeAdjustment:         true,
		MovementTypeTransfer:           true,
		MovementTypeDamage:             true,
		MovementTypeReturn:             true,
	}

	if !validTypes[movementType] {
		return stockmovementerrors.ErrInvalidMovementType
	}

	return nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
