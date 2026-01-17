package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory/aggregate"
	"github.com/basilex/promenade/internal/contexts/warehouse/inventory/repository"
	inventoryerrors "github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IInventoryUseCase defines the business operations for inventory management
type IInventoryUseCase interface {
	// CreateInventory creates a new inventory item
	CreateInventory(ctx context.Context, productID uuidv7.UUID, sku, productName, warehouseID string, createdBy uuidv7.UUID) (*aggregate.Inventory, error)

	// GetInventory retrieves inventory by ID
	GetInventory(ctx context.Context, id uuidv7.UUID) (*aggregate.Inventory, error)

	// GetBySKU retrieves inventory by unique SKU
	GetBySKU(ctx context.Context, sku string) (*aggregate.Inventory, error)

	// GetByProductID retrieves all inventory for a product across warehouses
	GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*aggregate.Inventory, error)

	// GetByWarehouse retrieves all inventory in a specific warehouse
	GetByWarehouse(ctx context.Context, warehouseID string) ([]*aggregate.Inventory, error)

	// GetByLocation retrieves inventory by warehouse and location
	GetByLocation(ctx context.Context, warehouseID, locationCode string) ([]*aggregate.Inventory, error)

	// GetLowStock retrieves items below reorder point
	GetLowStock(ctx context.Context) ([]*aggregate.Inventory, error)

	// ListInventory retrieves paginated inventory items
	ListInventory(ctx context.Context, page, pageSize int) ([]*aggregate.Inventory, int64, error)

	// UpdateInventory updates inventory details
	UpdateInventory(ctx context.Context, inventory *aggregate.Inventory) error

	// DeleteInventory soft-deletes inventory
	DeleteInventory(ctx context.Context, id uuidv7.UUID) error

	// ReceiveStock receives stock into inventory (increases quantity)
	ReceiveStock(ctx context.Context, id uuidv7.UUID, quantity, unitCostCents int, receivedBy uuidv7.UUID) (*aggregate.Inventory, error)

	// CommitStock commits stock from inventory (decreases available, increases committed)
	CommitStock(ctx context.Context, id uuidv7.UUID, quantity int, committedBy uuidv7.UUID) (*aggregate.Inventory, error)
}

// InventoryUseCase implements IInventoryUseCase interface
type InventoryUseCase struct {
	repo repository.IInventoryRepository
}

// NewInventoryUseCase creates a new inventory use case
func NewInventoryUseCase(repo repository.IInventoryRepository) IInventoryUseCase {
	return &InventoryUseCase{
		repo: repo,
	}
}

// CreateInventory creates a new inventory item
func (uc *InventoryUseCase) CreateInventory(ctx context.Context, productID uuidv7.UUID, sku, productName, warehouseID string, createdBy uuidv7.UUID) (*aggregate.Inventory, error) {
	// Validate inputs
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

	// Check if SKU already exists
	existing, err := uc.repo.GetBySKU(ctx, sku)
	if err == nil && existing != nil {
		return nil, inventoryerrors.ErrInventorySKUExists
	}

	// Create new inventory
	inv, err := aggregate.NewInventory(productID, sku, productName, warehouseID, createdBy)
	if err != nil {
		return nil, inventoryerrors.ErrInventoryCreateFailed
	}

	// Persist to repository
	if err := uc.repo.Create(ctx, inv); err != nil {
		return nil, inventoryerrors.ErrInventorySaveFailed
	}

	return inv, nil
}

// GetInventory retrieves inventory by ID
func (uc *InventoryUseCase) GetInventory(ctx context.Context, id uuidv7.UUID) (*aggregate.Inventory, error) {
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, inventoryerrors.ErrInventoryNotFound
	}
	return inv, nil
}

// GetBySKU retrieves inventory by unique SKU
func (uc *InventoryUseCase) GetBySKU(ctx context.Context, sku string) (*aggregate.Inventory, error) {
	if sku == "" {
		return nil, inventoryerrors.ErrInventorySKURequired
	}

	inv, err := uc.repo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, inventoryerrors.ErrInventoryNotFound
	}
	return inv, nil
}

// GetByProductID retrieves all inventory for a product across warehouses
func (uc *InventoryUseCase) GetByProductID(ctx context.Context, productID uuidv7.UUID) ([]*aggregate.Inventory, error) {
	return uc.repo.GetByProductID(ctx, productID)
}

// GetByWarehouse retrieves all inventory in a specific warehouse
func (uc *InventoryUseCase) GetByWarehouse(ctx context.Context, warehouseID string) ([]*aggregate.Inventory, error) {
	if warehouseID == "" {
		return nil, inventoryerrors.ErrInventoryWarehouseRequired
	}
	return uc.repo.GetByWarehouse(ctx, warehouseID)
}

// GetByLocation retrieves inventory by warehouse and location
func (uc *InventoryUseCase) GetByLocation(ctx context.Context, warehouseID, locationCode string) ([]*aggregate.Inventory, error) {
	if warehouseID == "" {
		return nil, inventoryerrors.ErrInventoryWarehouseRequired
	}
	if locationCode == "" {
		return nil, inventoryerrors.ErrInventoryLocationRequired
	}
	return uc.repo.GetByLocation(ctx, warehouseID, locationCode)
}

// GetLowStock retrieves items below reorder point
func (uc *InventoryUseCase) GetLowStock(ctx context.Context) ([]*aggregate.Inventory, error) {
	return uc.repo.GetLowStock(ctx)
}

// ListInventory retrieves paginated inventory items
func (uc *InventoryUseCase) ListInventory(ctx context.Context, page, pageSize int) ([]*aggregate.Inventory, int64, error) {
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
func (uc *InventoryUseCase) UpdateInventory(ctx context.Context, inventory *aggregate.Inventory) error {
	if inventory == nil {
		return inventoryerrors.ErrInventoryNil
	}

	// Check if exists
	existing, err := uc.repo.GetByID(ctx, inventory.GetID())
	if err != nil {
		return err
	}
	if existing == nil {
		return inventoryerrors.ErrInventoryNotFound
	}

	return uc.repo.Update(ctx, inventory)
}

// DeleteInventory soft-deletes inventory
func (uc *InventoryUseCase) DeleteInventory(ctx context.Context, id uuidv7.UUID) error {
	// Check if exists
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if inv == nil {
		return inventoryerrors.ErrInventoryNotFound
	}

	return uc.repo.Delete(ctx, id)
}

// ReceiveStock receives stock into inventory (increases quantity)
func (uc *InventoryUseCase) ReceiveStock(ctx context.Context, id uuidv7.UUID, quantity, unitCostCents int, receivedBy uuidv7.UUID) (*aggregate.Inventory, error) {
	if quantity <= 0 {
		return nil, inventoryerrors.ErrInventoryQuantityInvalid
	}
	if unitCostCents < 0 {
		return nil, inventoryerrors.ErrInventoryUnitCostNegative
	}

	// Get inventory
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, inventoryerrors.ErrInventoryNotFound
	}

	// Receive stock (entity business logic)
	if err := inv.ReceiveStock(quantity, int64(unitCostCents), receivedBy); err != nil {
		return nil, err
	}

	// Update in repository
	if err := uc.repo.Update(ctx, inv); err != nil {
		return nil, inventoryerrors.ErrInventoryUpdateFailed
	}

	return inv, nil
}

// CommitStock commits stock from inventory (decreases available, increases committed)
func (uc *InventoryUseCase) CommitStock(ctx context.Context, id uuidv7.UUID, quantity int, committedBy uuidv7.UUID) (*aggregate.Inventory, error) {
	if quantity <= 0 {
		return nil, inventoryerrors.ErrInventoryQuantityInvalid
	}

	// Get inventory
	inv, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, inventoryerrors.ErrInventoryNotFound
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
		return nil, inventoryerrors.ErrInventoryUpdateFailed
	}

	return inv, nil
}
