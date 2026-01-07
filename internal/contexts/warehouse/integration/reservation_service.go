package integration

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IReservationService handles stock reservation for orders
type IReservationService interface {
	// ReserveForOrder reserves stock for all items in an order
	ReserveForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, reservedBy uuidv7.UUID) error

	// ReleaseForOrder releases stock reservation for a cancelled order
	ReleaseForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, releasedBy uuidv7.UUID) error

	// CommitForOrder commits reserved stock for a fulfilled order
	CommitForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, committedBy uuidv7.UUID) error
}

// OrderItem represents an order line item for reservation
type OrderItem struct {
	ProductID uuidv7.UUID
	SKU       string // Optional: can use either ProductID or SKU
	Quantity  int
}

// reservationService implements IReservationService
type reservationService struct {
	inventoryUC inventory.IUseCase
}

// NewReservationService creates a new reservation service
func NewReservationService(inventoryUC inventory.IUseCase) IReservationService {
	return &reservationService{
		inventoryUC: inventoryUC,
	}
}

// ReserveForOrder reserves stock for all items in an order
func (s *reservationService) ReserveForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, reservedBy uuidv7.UUID) error {
	if orderID == uuidv7.Nil {
		return fmt.Errorf("order ID is required")
	}
	if len(items) == 0 {
		return fmt.Errorf("no items to reserve")
	}
	if reservedBy == uuidv7.Nil {
		return fmt.Errorf("reservedBy user ID is required")
	}

	// Track successful reservations for rollback on error
	var reservedInventory []*inventory.Inventory

	// Reserve stock for each item
	for _, item := range items {
		if item.Quantity <= 0 {
			s.rollbackReservations(ctx, orderID, reservedInventory, reservedBy)
			return fmt.Errorf("quantity must be positive for product %s", item.ProductID)
		}

		// Get inventory by SKU (primary) or ProductID (fallback)
		var inv *inventory.Inventory
		var err error
		if item.SKU != "" {
			inv, err = s.inventoryUC.GetBySKU(ctx, item.SKU)
		} else if item.ProductID != uuidv7.Nil {
			inventories, err := s.inventoryUC.GetByProductID(ctx, item.ProductID)
			if err != nil {
				s.rollbackReservations(ctx, orderID, reservedInventory, reservedBy)
				return fmt.Errorf("failed to get inventory for product %s: %w", item.ProductID, err)
			}
			if len(inventories) == 0 {
				s.rollbackReservations(ctx, orderID, reservedInventory, reservedBy)
				return fmt.Errorf("no inventory found for product %s", item.ProductID)
			}
			inv = inventories[0]
		} else {
			s.rollbackReservations(ctx, orderID, reservedInventory, reservedBy)
			return fmt.Errorf("either SKU or ProductID must be provided")
		}

		if err != nil {
			s.rollbackReservations(ctx, orderID, reservedInventory, reservedBy)
			return fmt.Errorf("inventory not found for item %s: %w", item.SKU, err)
		}

		// Reserve stock via entity method
		if err := inv.ReserveStock(item.Quantity, orderID, reservedBy); err != nil {
			s.rollbackReservations(ctx, orderID, reservedInventory, reservedBy)
			return fmt.Errorf("failed to reserve stock for %s: %w", item.SKU, err)
		}

		// Persist reservation
		if err := s.inventoryUC.UpdateInventory(ctx, inv); err != nil {
			s.rollbackReservations(ctx, orderID, reservedInventory, reservedBy)
			return fmt.Errorf("failed to save reservation for %s: %w", item.SKU, err)
		}

		reservedInventory = append(reservedInventory, inv)
	}

	return nil
}

// ReleaseForOrder releases stock reservation for a cancelled order
func (s *reservationService) ReleaseForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, releasedBy uuidv7.UUID) error {
	if orderID == uuidv7.Nil {
		return fmt.Errorf("order ID is required")
	}
	if len(items) == 0 {
		return fmt.Errorf("no items to release")
	}
	if releasedBy == uuidv7.Nil {
		return fmt.Errorf("releasedBy user ID is required")
	}

	// Release stock for each item (best effort, continue on errors)
	for _, item := range items {
		if item.Quantity <= 0 {
			continue
		}

		// Get inventory
		var inv *inventory.Inventory
		var err error
		if item.SKU != "" {
			inv, err = s.inventoryUC.GetBySKU(ctx, item.SKU)
		} else if item.ProductID != uuidv7.Nil {
			inventories, err := s.inventoryUC.GetByProductID(ctx, item.ProductID)
			if err != nil || len(inventories) == 0 {
				continue
			}
			inv = inventories[0]
		} else {
			continue
		}

		if err != nil {
			continue
		}

		// Release stock via entity method
		if err := inv.ReleaseReservation(item.Quantity, orderID, releasedBy); err != nil {
			continue
		}

		// Persist release
		_ = s.inventoryUC.UpdateInventory(ctx, inv)
	}

	return nil
}

// CommitForOrder commits reserved stock for a fulfilled order
func (s *reservationService) CommitForOrder(ctx context.Context, orderID uuidv7.UUID, items []OrderItem, committedBy uuidv7.UUID) error {
	if orderID == uuidv7.Nil {
		return fmt.Errorf("order ID is required")
	}
	if len(items) == 0 {
		return fmt.Errorf("no items to commit")
	}
	if committedBy == uuidv7.Nil {
		return fmt.Errorf("committedBy user ID is required")
	}

	// Commit stock for each item
	for _, item := range items {
		if item.Quantity <= 0 {
			return fmt.Errorf("quantity must be positive for product %s", item.ProductID)
		}

		// Get inventory
		var inv *inventory.Inventory
		var err error
		if item.SKU != "" {
			inv, err = s.inventoryUC.GetBySKU(ctx, item.SKU)
		} else if item.ProductID != uuidv7.Nil {
			inventories, err := s.inventoryUC.GetByProductID(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("failed to get inventory for product %s: %w", item.ProductID, err)
			}
			if len(inventories) == 0 {
				return fmt.Errorf("no inventory found for product %s", item.ProductID)
			}
			inv = inventories[0]
		} else {
			return fmt.Errorf("either SKU or ProductID must be provided")
		}

		if err != nil {
			return fmt.Errorf("inventory not found for item %s: %w", item.SKU, err)
		}

		// Commit stock via entity method
		if err := inv.CommitReservation(item.Quantity, orderID, committedBy); err != nil {
			return fmt.Errorf("failed to commit stock for %s: %w", item.SKU, err)
		}

		// Persist commit
		if err := s.inventoryUC.UpdateInventory(ctx, inv); err != nil {
			return fmt.Errorf("failed to save commit for %s: %w", item.SKU, err)
		}
	}

	return nil
}

// rollbackReservations releases previously reserved stock on error
func (s *reservationService) rollbackReservations(ctx context.Context, orderID uuidv7.UUID, inventories []*inventory.Inventory, releasedBy uuidv7.UUID) {
	for _, inv := range inventories {
		if err := inv.ReleaseReservation(inv.GetQuantityReserved(), orderID, releasedBy); err != nil {
			continue
		}
		_ = s.inventoryUC.UpdateInventory(ctx, inv)
	}
}
