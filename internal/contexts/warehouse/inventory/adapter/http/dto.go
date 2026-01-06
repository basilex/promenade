package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
)

// CreateInventoryRequest represents inventory creation request
type CreateInventoryRequest struct {
	ProductID    string `json:"product_id" binding:"required"`
	SKU          string `json:"sku" binding:"required,min=1,max=100"`
	ProductName  string `json:"product_name" binding:"required,min=1,max=255"`
	WarehouseID  string `json:"warehouse_id" binding:"required,min=1,max=50"`
	LocationCode string `json:"location_code,omitempty" binding:"max=50"`
	LocationZone string `json:"location_zone,omitempty" binding:"max=50"`
	CreatedBy    string `json:"created_by" binding:"required"`
}

// UpdateInventoryRequest represents inventory update request
type UpdateInventoryRequest struct {
	LocationCode   *string `json:"location_code,omitempty"`
	LocationZone   *string `json:"location_zone,omitempty"`
	ReorderPoint   *int    `json:"reorder_point,omitempty"`
	ReorderQuantity *int   `json:"reorder_quantity,omitempty"`
	UnitCostCents  *int    `json:"unit_cost_cents,omitempty"`
	CurrencyCode   *string `json:"currency_code,omitempty"`
	Notes          *string `json:"notes,omitempty"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// ReceiveStockRequest represents stock receiving request
type ReceiveStockRequest struct {
	Quantity      int    `json:"quantity" binding:"required,gt=0"`
	UnitCostCents int    `json:"unit_cost_cents" binding:"required,gte=0"`
	ReceivedBy    string `json:"received_by" binding:"required"`
}

// CommitStockRequest represents stock commitment request
type CommitStockRequest struct {
	Quantity    int    `json:"quantity" binding:"required,gt=0"`
	CommittedBy string `json:"committed_by" binding:"required"`
}

// ReleaseCommittedRequest represents committed stock release request
type ReleaseCommittedRequest struct {
	Quantity   int    `json:"quantity" binding:"required,gt=0"`
	ReleasedBy string `json:"released_by" binding:"required"`
}

// AdjustQuantityRequest represents quantity adjustment request
type AdjustQuantityRequest struct {
	Quantity   int    `json:"quantity" binding:"required"`
	Reason     string `json:"reason" binding:"required,min=1,max=500"`
	AdjustedBy string `json:"adjusted_by" binding:"required"`
}

// BulkUpdateRequest represents bulk update request
type BulkUpdateRequest struct {
	InventoryIDs []string `json:"inventory_ids" binding:"required,min=1"`
}

// ReserveStockRequest represents stock reservation request
type ReserveStockRequest struct {
	Quantity   int    `json:"quantity" binding:"required,gt=0"`
	OrderID    string `json:"order_id" binding:"required"`
	ReservedBy string `json:"reserved_by" binding:"required"`
}

// ReleaseReservationRequest represents reservation release request
type ReleaseReservationRequest struct {
	Quantity   int    `json:"quantity" binding:"required,gt=0"`
	OrderID    string `json:"order_id" binding:"required"`
	ReleasedBy string `json:"released_by" binding:"required"`
}

// InventoryResponse represents inventory response
type InventoryResponse struct {
	ID                 string    `json:"id"`
	Version            int       `json:"version"`
	ProductID          string    `json:"product_id"`
	SKU                string    `json:"sku"`
	ProductName        string    `json:"product_name"`
	QuantityOnHand     int       `json:"quantity_on_hand"`
	QuantityReserved   int       `json:"quantity_reserved"`
	QuantityCommitted  int       `json:"quantity_committed"`
	QuantityAvailable  int       `json:"quantity_available"`
	WarehouseID        string    `json:"warehouse_id"`
	LocationCode       string    `json:"location_code,omitempty"`
	LocationZone       string    `json:"location_zone,omitempty"`
	ReorderPoint       int       `json:"reorder_point"`
	ReorderQuantity    int       `json:"reorder_quantity"`
	LastRestocked      *time.Time `json:"last_restocked,omitempty"`
	Status             string    `json:"status"`
	IsActive           bool      `json:"is_active"`
	Notes              string    `json:"notes,omitempty"`
	UnitCostCents      int       `json:"unit_cost_cents"`
	CurrencyCode       string    `json:"currency_code"`
	LastUpdatedBy      string    `json:"last_updated_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ToInventoryResponse converts inventory entity to response DTO
func ToInventoryResponse(inv *inventory.Inventory) *InventoryResponse {
	return &InventoryResponse{
		ID:                 inv.ID.String(),
		Version:            inv.Version,
		ProductID:          inv.ProductID.String(),
		SKU:                inv.SKU,
		ProductName:        inv.ProductName,
		QuantityOnHand:     inv.QuantityOnHand,
		QuantityReserved:   inv.QuantityReserved,
		QuantityCommitted:  inv.QuantityCommitted,
		QuantityAvailable:  inv.QuantityAvailable,
		WarehouseID:        inv.WarehouseID,
		LocationCode:       inv.LocationCode,
		LocationZone:       inv.LocationZone,
		ReorderPoint:       inv.ReorderPoint,
		ReorderQuantity:    inv.ReorderQuantity,
		LastRestocked:      &inv.LastRestocked,
		Status:             string(inv.Status),
		IsActive:           inv.IsActive,
		Notes:              inv.Notes,
		UnitCostCents:      int(inv.UnitCostCents),
		CurrencyCode:       inv.CurrencyCode,
		LastUpdatedBy:      inv.LastUpdatedBy.String(),
		CreatedAt:          inv.CreatedAt,
		UpdatedAt:          inv.UpdatedAt,
	}
}

// ToInventoryResponseList converts slice of inventory entities to response DTOs
func ToInventoryResponseList(items []*inventory.Inventory) []*InventoryResponse {
	responses := make([]*InventoryResponse, len(items))
	for i, item := range items {
		responses[i] = ToInventoryResponse(item)
	}
	return responses
}
