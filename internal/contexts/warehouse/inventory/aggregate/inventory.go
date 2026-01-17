package aggregate

import (
	inventoryerrors "github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// InventoryStatus represents the operational status of inventory
type InventoryStatus string

const (
	// InventoryStatusAvailable - stock available for reservation
	InventoryStatusAvailable InventoryStatus = "available"
	// InventoryStatusReserved - stock reserved for orders (not yet committed)
	InventoryStatusReserved InventoryStatus = "reserved"
	// InventoryStatusCommitted - stock committed to fulfilled orders
	InventoryStatusCommitted InventoryStatus = "committed"
	// InventoryStatusDamaged - damaged goods (not available)
	InventoryStatusDamaged InventoryStatus = "damaged"
	// InventoryStatusInTransit - stock in transit between locations
	InventoryStatusInTransit InventoryStatus = "in_transit"
	// InventoryStatusQuarantined - stock under inspection or recall
	InventoryStatusQuarantined InventoryStatus = "quarantine"
)

// Inventory is an aggregate root for warehouse inventory management
// It tracks stock levels, locations, and reorder points with real-time accuracy
//
// Modern patterns (2026):
// - Event sourcing ready (domain events for all state changes)
// - CQRS compatible (separate read models for analytics)
// - Optimistic locking via Version field (prevents race conditions)
// - Multi-location support (warehouse, shelf, bin)
type Inventory struct {
	aggregate.BaseAggregate

	// Product identification
	ProductID   uuidv7.UUID `db:"product_id"`   // Link to product catalog
	SKU         string      `db:"sku"`          // Stock Keeping Unit (unique)
	ProductName string      `db:"product_name"` // Denormalized for queries

	// Stock tracking
	QuantityOnHand    int `db:"quantity_on_hand"`    // Physical stock available
	QuantityReserved  int `db:"quantity_reserved"`   // Reserved for pending orders
	QuantityCommitted int `db:"quantity_committed"`  // Committed to fulfilled orders
	QuantityAvailable int `db:"quantity_available"`  // Calculated: OnHand - Reserved - Committed

	// Location tracking (multi-location support)
	WarehouseID  string `db:"warehouse_id"`  // Warehouse identifier
	LocationCode string `db:"location_code"` // Shelf/Bin location (e.g., "A-12-03")
	LocationZone string `db:"location_zone"` // Zone for picking optimization (e.g., "Zone-A")

	// Reorder management
	ReorderPoint    int       `db:"reorder_point"`     // Trigger for low stock alert
	ReorderQuantity int       `db:"reorder_quantity"`  // Suggested order quantity
	LastRestocked   time.Time `db:"last_restocked"`    // Last stock receipt date

	// Operational status
	Status   InventoryStatus `db:"status"`    // Current operational status
	IsActive bool            `db:"is_active"` // Can be used for orders
	Notes    string          `db:"notes"`     // Admin notes (max 500 chars)

	// Cost tracking (for inventory valuation)
	UnitCostCents int64  `db:"unit_cost_cents"` // Cost per unit in cents
	CurrencyCode  string `db:"currency_code"`   // Currency (e.g., "USD")

	// Audit metadata (who made last change)
	LastUpdatedBy uuidv7.UUID `db:"last_updated_by"` // User ID
}

// NewInventory creates a new inventory record for a product
// Modern approach: All business logic in domain layer, not in database
func NewInventory(productID uuidv7.UUID, sku, productName string, warehouseID string, createdBy uuidv7.UUID) (*Inventory, error) {
	// Validation
	if productID == uuidv7.Nil {
		return nil, inventoryerrors.ErrInventoryProductIDRequired
	}
	if sku == "" {
		return nil, inventoryerrors.ErrInventorySKURequired
	}
	if productName == "" {
		return nil, inventoryerrors.ErrInventoryProductNameRequired
	}
	if warehouseID == "" {
		return nil, inventoryerrors.ErrInventoryWarehouseRequired
	}
	if createdBy == uuidv7.Nil {
		return nil, inventoryerrors.ErrInventoryCreatedByRequired
	}

	return &Inventory{
		BaseAggregate:     aggregate.NewBaseAggregate(),
		ProductID:         productID,
		SKU:               sku,
		ProductName:       productName,
		WarehouseID:       warehouseID,
		LocationCode:      "UNKNOWN", // Default location until assigned
		QuantityOnHand:    0,
		QuantityReserved:  0,
		QuantityCommitted: 0,
		QuantityAvailable: 0,
		ReorderPoint:      10, // Default: alert when < 10 units
		ReorderQuantity:   50, // Default: reorder 50 units
		Status:            InventoryStatusAvailable,
		IsActive:          true,
		CurrencyCode:      "USD",
		LastRestocked:     time.Now(),
		LastUpdatedBy:     createdBy, // Audit: who created this record
	}, nil
}

// ReceiveStock records stock receipt (purchase, transfer-in, return)
// Modern pattern: Domain events for event-driven architecture
func (i *Inventory) ReceiveStock(quantity int, unitCostCents int64, receivedBy uuidv7.UUID) error {
	if quantity <= 0 {
		return inventoryerrors.ErrInventoryQuantityInvalid
	}
	if !i.IsActive {
		return inventoryerrors.ErrInventoryInactive
	}

	// Update quantities
	i.QuantityOnHand += quantity
	i.recalculateAvailable()

	// Update cost (weighted average)
	if i.QuantityOnHand > quantity {
		totalCost := (i.UnitCostCents * int64(i.QuantityOnHand-quantity)) + (unitCostCents * int64(quantity))
		i.UnitCostCents = totalCost / int64(i.QuantityOnHand)
	} else {
		i.UnitCostCents = unitCostCents
	}

	// Update metadata
	i.LastRestocked = time.Now()
	i.LastUpdatedBy = receivedBy
	i.Touch()

	// TODO: Emit domain event: inventory.stock_received
	// event := NewStockReceivedEvent(i.ID, quantity, unitCostCents)
	// i.AddEvent(event)

	return nil
}

// ReserveStock reserves stock for an order (not yet committed)
// Modern approach: Optimistic locking via Version field (prevents overselling)
func (i *Inventory) ReserveStock(quantity int, orderID uuidv7.UUID, reservedBy uuidv7.UUID) error {
	if quantity <= 0 {
		return inventoryerrors.ErrInventoryQuantityInvalid
	}
	if i.QuantityAvailable < quantity {
		return inventoryerrors.ErrInventoryInsufficientStock
	}
	if !i.IsActive {
		return inventoryerrors.ErrInventoryInactive
	}

	// Reserve stock
	i.QuantityReserved += quantity
	i.recalculateAvailable()
	i.LastUpdatedBy = reservedBy
	i.Touch()

	// TODO: Emit domain event: inventory.stock_reserved
	// event := NewStockReservedEvent(i.ID, quantity, orderID)
	// i.AddEvent(event)

	return nil
}

// ReleaseReservation releases reserved stock (order cancelled)
// Compensation logic for Saga pattern
func (i *Inventory) ReleaseReservation(quantity int, orderID uuidv7.UUID, releasedBy uuidv7.UUID) error {
	if quantity <= 0 {
		return inventoryerrors.ErrInventoryQuantityInvalid
	}
	if i.QuantityReserved < quantity {
		return inventoryerrors.ErrInventoryInsufficientReserved
	}

	// Release reservation
	i.QuantityReserved -= quantity
	i.recalculateAvailable()
	i.LastUpdatedBy = releasedBy
	i.Touch()

	// TODO: Emit domain event: inventory.reservation_released
	// event := NewReservationReleasedEvent(i.ID, quantity, orderID)
	// i.AddEvent(event)

	return nil
}

// CommitReservation commits reserved stock (order fulfilled)
// Final step in order fulfillment saga
func (i *Inventory) CommitReservation(quantity int, orderID uuidv7.UUID, committedBy uuidv7.UUID) error {
	if quantity <= 0 {
		return inventoryerrors.ErrInventoryQuantityInvalid
	}
	if i.QuantityReserved < quantity {
		return inventoryerrors.ErrInventoryInsufficientReserved
	}

	// Commit stock (move from reserved to committed)
	i.QuantityReserved -= quantity
	i.QuantityCommitted += quantity
	i.QuantityOnHand -= quantity
	i.recalculateAvailable()
	i.LastUpdatedBy = committedBy
	i.Touch()

	// TODO: Emit domain event: inventory.stock_committed
	// event := NewStockCommittedEvent(i.ID, quantity, orderID)
	// i.AddEvent(event)

	return nil
}

// AdjustStock adjusts stock for inventory corrections (audit, damage, theft)
// Positive = add stock, Negative = remove stock
func (i *Inventory) AdjustStock(quantityDelta int, reason string, adjustedBy uuidv7.UUID) error {
	if reason == "" {
		return inventoryerrors.ErrInventoryAdjustmentReasonRequired
	}

	newQuantity := i.QuantityOnHand + quantityDelta
	if newQuantity < 0 {
		return inventoryerrors.ErrInventoryNegativeStock
	}

	// Apply adjustment
	i.QuantityOnHand = newQuantity
	i.recalculateAvailable()
	i.LastUpdatedBy = adjustedBy
	i.Touch()

	// TODO: Emit domain event: inventory.stock_adjusted
	// event := NewStockAdjustedEvent(i.ID, quantityDelta, reason)
	// i.AddEvent(event)

	return nil
}

// SetLocation updates warehouse location for stock
// Used for warehouse organization and picking optimization
func (i *Inventory) SetLocation(locationCode, locationZone string, updatedBy uuidv7.UUID) error {
	if locationCode == "" {
		return inventoryerrors.ErrInventoryLocationRequired
	}

	i.LocationCode = locationCode
	i.LocationZone = locationZone
	i.LastUpdatedBy = updatedBy
	i.Touch()

	return nil
}

// SetReorderPoint updates reorder point and quantity
// Used for automated low stock alerts
func (i *Inventory) SetReorderPoint(reorderPoint, reorderQuantity int, updatedBy uuidv7.UUID) error {
	if reorderPoint < 0 {
		return inventoryerrors.ErrInventoryReorderPointNegative
	}
	if reorderQuantity <= 0 {
		return inventoryerrors.ErrInventoryReorderQuantityInvalid
	}

	i.ReorderPoint = reorderPoint
	i.ReorderQuantity = reorderQuantity
	i.LastUpdatedBy = updatedBy
	i.Touch()

	return nil
}

// MarkAsDamaged marks inventory as damaged (not available for orders)
func (i *Inventory) MarkAsDamaged(quantity int, reason string, updatedBy uuidv7.UUID) error {
	if quantity <= 0 {
		return inventoryerrors.ErrInventoryQuantityInvalid
	}
	if i.QuantityAvailable < quantity {
		return inventoryerrors.ErrInventoryInsufficientAvailable
	}

	// Remove from available stock
	i.QuantityOnHand -= quantity
	i.recalculateAvailable()
	i.Status = InventoryStatusDamaged
	i.Notes = fmt.Sprintf("Damaged: %s", reason)
	i.LastUpdatedBy = updatedBy
	i.Touch()

	// TODO: Emit domain event: inventory.marked_damaged
	// event := NewMarkedDamagedEvent(i.ID, quantity, reason)
	// i.AddEvent(event)

	return nil
}

// Activate activates inventory for orders
func (i *Inventory) Activate(activatedBy uuidv7.UUID) error {
	if i.IsActive {
		return inventoryerrors.ErrInventoryAlreadyActive
	}

	i.IsActive = true
	i.Status = InventoryStatusAvailable
	i.LastUpdatedBy = activatedBy
	i.Touch()

	return nil
}

// Deactivate deactivates inventory (discontinue product)
func (i *Inventory) Deactivate(deactivatedBy uuidv7.UUID) error {
	if !i.IsActive {
		return inventoryerrors.ErrInventoryInactive
	}
	if i.QuantityReserved > 0 {
		return inventoryerrors.ErrInventoryCannotDeactivateWithReservedStock
	}

	i.IsActive = false
	i.LastUpdatedBy = deactivatedBy
	i.Touch()

	return nil
}

// IsLowStock checks if stock is below reorder point
// Used for low stock alert system
func (i *Inventory) IsLowStock() bool {
	return i.QuantityAvailable <= i.ReorderPoint
}

// GetStockValue calculates total inventory value
// Used for financial reporting and inventory valuation
func (i *Inventory) GetStockValue() int64 {
	return i.UnitCostCents * int64(i.QuantityOnHand)
}

// recalculateAvailable recalculates available quantity
// Available = OnHand - Reserved - Committed
// Internal helper to maintain data consistency
func (i *Inventory) recalculateAvailable() {
	// Available = OnHand - Reserved
	// Committed is NOT subtracted because it's already deducted from OnHand
	i.QuantityAvailable = i.QuantityOnHand - i.QuantityReserved
	if i.QuantityAvailable < 0 {
		i.QuantityAvailable = 0
	}
}

// Validate performs business rule validation
// Called before persistence operations
func (i *Inventory) Validate() error {
	if i.ProductID == uuidv7.Nil {
		return inventoryerrors.ErrInventoryProductIDRequired
	}
	if i.SKU == "" {
		return inventoryerrors.ErrInventorySKURequired
	}
	if i.ProductName == "" {
		return inventoryerrors.ErrInventoryProductNameRequired
	}
	if i.WarehouseID == "" {
		return inventoryerrors.ErrInventoryWarehouseRequired
	}
	if i.QuantityOnHand < 0 {
		return inventoryerrors.ErrInventoryQuantityOnHandNegative
	}
	if i.QuantityReserved < 0 {
		return inventoryerrors.ErrInventoryQuantityReservedNegative
	}
	if i.QuantityCommitted < 0 {
		return inventoryerrors.ErrInventoryQuantityCommittedNegative
	}
	if i.ReorderPoint < 0 {
		return inventoryerrors.ErrInventoryReorderPointNegative
	}
	if i.ReorderQuantity <= 0 {
		return inventoryerrors.ErrInventoryReorderQuantityInvalid
	}
	if len(i.Notes) > 500 {
		return inventoryerrors.ErrInventoryNotesTooLong
	}

	return nil
}

// ==============================================================================
// Getters (for repository layer)
// ==============================================================================

func (i *Inventory) GetID() uuidv7.UUID                  { return i.ID }
func (i *Inventory) GetVersion() int                     { return i.Version }
func (i *Inventory) GetProductID() uuidv7.UUID           { return i.ProductID }
func (i *Inventory) GetSKU() string                      { return i.SKU }
func (i *Inventory) GetProductName() string              { return i.ProductName }
func (i *Inventory) GetQuantityOnHand() int              { return i.QuantityOnHand }
func (i *Inventory) GetQuantityReserved() int            { return i.QuantityReserved }
func (i *Inventory) GetQuantityCommitted() int           { return i.QuantityCommitted }
func (i *Inventory) GetQuantityAvailable() int           { return i.QuantityAvailable }
func (i *Inventory) GetWarehouseID() string              { return i.WarehouseID }
func (i *Inventory) GetLocationCode() string             { return i.LocationCode }
func (i *Inventory) GetLocationZone() string             { return i.LocationZone }
func (i *Inventory) GetReorderPoint() int                { return i.ReorderPoint }
func (i *Inventory) GetReorderQuantity() int             { return i.ReorderQuantity }
func (i *Inventory) GetLastRestocked() *time.Time        { 
	if i.LastRestocked.IsZero() { return nil }
	t := i.LastRestocked
	return &t 
}
func (i *Inventory) GetStatus() InventoryStatus          { return i.Status }
func (i *Inventory) GetIsActive() bool                   { return i.IsActive }
func (i *Inventory) GetNotes() string                    { return i.Notes }
func (i *Inventory) GetUnitCostCents() int64             { return i.UnitCostCents }
func (i *Inventory) GetCurrencyCode() string             { return i.CurrencyCode }
func (i *Inventory) GetLastUpdatedBy() *uuidv7.UUID      { 
	if i.LastUpdatedBy == uuidv7.Nil { return nil }
	id := i.LastUpdatedBy
	return &id
}
func (i *Inventory) GetCreatedAt() time.Time             { return i.CreatedAt }
func (i *Inventory) GetUpdatedAt() time.Time             { return i.UpdatedAt }
func (i *Inventory) GetDeletedAt() *time.Time            { return i.DeletedAt }

// ==============================================================================
// Setters (for repository layer - hydration from database)
// ==============================================================================

func (i *Inventory) SetID(id uuidv7.UUID)                        { i.ID = id }
func (i *Inventory) SetVersion(version int)                      { i.Version = version }
func (i *Inventory) SetProductID(id uuidv7.UUID)                 { i.ProductID = id }
func (i *Inventory) SetSKU(sku string)                           { i.SKU = sku }
func (i *Inventory) SetProductName(name string)                  { i.ProductName = name }
func (i *Inventory) SetQuantityOnHand(qty int)                   { i.QuantityOnHand = qty }
func (i *Inventory) SetQuantityReserved(qty int)                 { i.QuantityReserved = qty }
func (i *Inventory) SetQuantityCommitted(qty int)                { i.QuantityCommitted = qty }
func (i *Inventory) SetQuantityAvailable(qty int)                { i.QuantityAvailable = qty }
func (i *Inventory) SetWarehouseID(id string)                    { i.WarehouseID = id }
func (i *Inventory) SetLocationCode(code string)                 { i.LocationCode = code }
func (i *Inventory) SetLocationZone(zone string)                 { i.LocationZone = zone }
func (i *Inventory) SetReorderQuantity(qty int)                  { i.ReorderQuantity = qty }
func (i *Inventory) SetLastRestocked(t *time.Time)               { 
	if t != nil { i.LastRestocked = *t }
}
func (i *Inventory) SetStatus(status InventoryStatus)            { i.Status = status }
func (i *Inventory) SetIsActive(active bool)                     { i.IsActive = active }
func (i *Inventory) SetNotes(notes string)                       { i.Notes = notes }
func (i *Inventory) SetUnitCostCents(cost int64)                 { i.UnitCostCents = cost }
func (i *Inventory) SetCurrencyCode(code string)                 { i.CurrencyCode = code }
func (i *Inventory) SetLastUpdatedBy(id *uuidv7.UUID)            { 
	if id != nil { i.LastUpdatedBy = *id }
}
func (i *Inventory) SetCreatedAt(t time.Time)                    { i.CreatedAt = t }
func (i *Inventory) SetUpdatedAt(t time.Time)                    { i.UpdatedAt = t }
func (i *Inventory) SetDeletedAt(t *time.Time)                   { i.DeletedAt = t }
