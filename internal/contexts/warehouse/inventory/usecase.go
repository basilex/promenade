package inventory

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines the business operations for inventory management
type IUseCase interface {
	// CreateInventory creates a new inventory item
	CreateInventory(ctx context.Context, productID uuidv7.UUID, sku, productName, warehouseID string, createdBy uuidv7.UUID) (*Inventory, error)

	// GetInventory retrieves inventory by ID
	GetInventory(ctx context.Context, id uuidv7.UUID) (*Inventory, error)

	// GetBySKU retrieves inventory by unique SKU
	GetBySKU(ctx context.Context, sku string) (*Inventory, error)

	// GetByProductID retrieves all inventory for a product across warehouses
	GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*Inventory, error)

	// GetByWarehouse retrieves all inventory in a specific warehouse
	GetByWarehouse(ctx context.Context, warehouseID string) ([]*Inventory, error)

	// GetByLocation retrieves inventory by warehouse and location
	GetByLocation(ctx context.Context, warehouseID, locationCode string) ([]*Inventory, error)

	// GetLowStock retrieves items below reorder point
	GetLowStock(ctx context.Context) ([]*Inventory, error)

	// ListInventory retrieves paginated inventory items
	ListInventory(ctx context.Context, page, pageSize int) ([]*Inventory, int64, error)

	// UpdateInventory updates inventory details
	UpdateInventory(ctx context.Context, inventory *Inventory) error

	// DeleteInventory soft-deletes inventory
	DeleteInventory(ctx context.Context, id uuidv7.UUID) error

	// ReceiveStock receives stock into inventory (increases quantity)
	ReceiveStock(ctx context.Context, id uuidv7.UUID, quantity, unitCostCents int, receivedBy uuidv7.UUID) (*Inventory, error)

	// CommitStock commits stock from inventory (decreases available, increases committed)
	CommitStock(ctx context.Context, id uuidv7.UUID, quantity int, committedBy uuidv7.UUID) (*Inventory, error)
}

// useCase implements IUseCase interface
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new inventory use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateInventory creates a new inventory item
func (uc *useCase) CreateInventory(ctx context.Context, productID uuidv7.UUID, sku, productName, warehouseID string, createdBy uuidv7.UUID) (*Inventory, error) {
	// Validate inputs
	if sku == "" {
		return nil, ErrInventorySKURequired
	}
	if productName == "" {
		return nil, ErrInventoryProductNameRequired
	}
	if warehouseID == "" {
		return nil, ErrInventoryWarehouseRequired
	}
	if createdBy == uuidv7.Nil {
		return nil, ErrInventoryCreatedByRequired
	}

	// Check if SKU already exists
	existing, err := uc.repo.GetBySKU(ctx, sku)
	if err == nil && existing != nil {
		return nil, ErrInventorySKUExists
	}

	// Create new inventory
	inv, err := NewInventory(productID, sku, productName, warehouseID, createdBy)
	if err != nil {
		return nil, ErrInventoryCreateFailed
	}

	// Persist to repository
	if err := uc.repo.Create(ctx, inv); err != nil {
		return nil, ErrInventorySaveFailed
	}

	return inv, nil
}

// GetInventory retrieves inventory by ID
func (uc *useCase) GetInventory(ctx context.Context, id uuidv7.UUID) (*Inventory, error) {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrInventoryNotFound
	}
	return inv, nil
}

// GetBySKU retrieves inventory by unique SKU
func (uc *useCase) GetBySKU(ctx context.Context, sku string) (*Inventory, error) {
	if sku == "" {
		return nil, ErrInventorySKURequired
	}

	inv, err := uc.repo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrInventoryNotFound
	}
	return inv, nil
}

// GetByProductID retrieves all inventory for a product across warehouses
func (uc *useCase) GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*Inventory, error) {
	return uc.repo.GetByProductID(ctx, productID)
}

// GetByWarehouse retrieves all inventory in a specific warehouse
func (uc *useCase) GetByWarehouse(ctx context.Context, warehouseID string) ([]*Inventory, error) {
	if warehouseID == "" {
		return nil, ErrInventoryWarehouseRequired
	}
	return uc.repo.GetByWarehouse(ctx, warehouseID)
}

// GetByLocation retrieves inventory by warehouse and location
func (uc *useCase) GetByLocation(ctx context.Context, warehouseID, locationCode string) ([]*Inventory, error) {
	if warehouseID == "" {
		return nil, ErrInventoryWarehouseRequired
	}
	if locationCode == "" {
		return nil, ErrInventoryLocationRequired
	}
	return uc.repo.GetByLocation(ctx, warehouseID, locationCode)
}

// GetLowStock retrieves items below reorder point
func (uc *useCase) GetLowStock(ctx context.Context) ([]*Inventory, error) {
	return uc.repo.GetLowStock(ctx)
}

// ListInventory retrieves paginated inventory items
func (uc *useCase) ListInventory(ctx context.Context, page, pageSize int) ([]*Inventory, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return uc.repo.List(ctx, pageSize, offset)
}

// UpdateInventory updates inventory details
func (uc *useCase) UpdateInventory(ctx context.Context, inventory *Inventory) error {
	if inventory == nil {
		return ErrInventoryNil
	}

	// Check if exists
	existing, err := uc.repo.GetByID(ctx, inventory.GetID())
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrInventoryNotFound
	}

	return uc.repo.Update(ctx, inventory)
}

// DeleteInventory soft-deletes inventory
func (uc *useCase) DeleteInventory(ctx context.Context, id uuidv7.UUID) error {
	// Check if exists
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return ErrInventoryNotFound
	}

	return uc.repo.Delete(ctx, id)
}

// ReceiveStock receives stock into inventory (increases quantity)
func (uc *useCase) ReceiveStock(ctx context.Context, id uuidv7.UUID, quantity, unitCostCents int, receivedBy uuidv7.UUID) (*Inventory, error) {
	if quantity <= 0 {
		return nil, ErrInventoryQuantityInvalid
	}
	if unitCostCents < 0 {
		return nil, ErrInventoryUnitCostNegative
	}

	// Get inventory
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrInventoryNotFound
	}

	// Receive stock (entity business logic)
	if err := inv.ReceiveStock(quantity, int64(unitCostCents), receivedBy); err != nil {
		return nil, err
	}

	// Update in repository
	if err := uc.repo.Update(ctx, inv); err != nil {
		return nil, ErrInventoryUpdateFailed
	}

	return inv, nil
}

// CommitStock commits stock from inventory (decreases available, increases committed)
func (uc *useCase) CommitStock(ctx context.Context, id uuidv7.UUID, quantity int, committedBy uuidv7.UUID) (*Inventory, error) {
	if quantity <= 0 {
		return nil, ErrInventoryQuantityInvalid
	}

	// Get inventory
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, ErrInventoryNotFound
	}

	// Commit stock (reduce available, increase committed)
	// Use AdjustStock with negative quantity to reduce QuantityOnHand
	if err := inv.AdjustStock(-quantity, "Stock committed", committedBy); err != nil {
		return nil, err
	}
	
	// Manually increase committed quantity
	inv.QuantityCommitted += quantity
	inv.Touch()

	// Update in repository
	if err := uc.repo.Update(ctx, inv); err != nil {
		return nil, ErrInventoryUpdateFailed
	}

	return inv, nil
}
