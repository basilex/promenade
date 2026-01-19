package dto

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement/aggregate"
)

// ============================================================================
// Generic Request
// ============================================================================

// RecordMovementRequest represents stock movement recording request
type RecordMovementRequest struct {
	InventoryID    string `json:"inventory_id" binding:"required"`
	Type           string `json:"type" binding:"required"`
	Quantity       int    `json:"quantity" binding:"required"`
	QuantityBefore int    `json:"quantity_before" binding:"required"` // Current quantity before movement
	CreatedBy      string `json:"created_by" binding:"required"`
}

// ============================================================================
// Specialized Requests
// ============================================================================

// RecordReceiptRequest is the request body for recording a receipt movement.
type RecordReceiptRequest struct {
	InventoryID    string  `json:"inventory_id" binding:"required,uuid"`
	Quantity       int     `json:"quantity" binding:"required,min=1"`
	QuantityBefore int     `json:"quantity_before" binding:"required,min=0"`
	UnitCostCents  *int64  `json:"unit_cost_cents,omitempty"`
	CurrencyCode   *string `json:"currency_code,omitempty"`
	Notes          string  `json:"notes,omitempty"`
}

// RecordReservationRequest is the request body for recording a reservation movement.
type RecordReservationRequest struct {
	InventoryID    string `json:"inventory_id" binding:"required,uuid"`
	Quantity       int    `json:"quantity" binding:"required,min=1"`
	QuantityBefore int    `json:"quantity_before" binding:"required,min=0"`
	OrderID        string `json:"order_id" binding:"required,uuid"`
	Notes          string `json:"notes,omitempty"`
}

// ReleaseReservationRequest is the request body for releasing a reservation.
type ReleaseReservationRequest struct {
	InventoryID    string `json:"inventory_id" binding:"required,uuid"`
	Quantity       int    `json:"quantity" binding:"required,min=1"`
	QuantityBefore int    `json:"quantity_before" binding:"required,min=0"`
	OrderID        string `json:"order_id" binding:"required,uuid"`
	Notes          string `json:"notes,omitempty"`
}

// RecordCommitRequest is the request body for committing stock.
type RecordCommitRequest struct {
	InventoryID    string `json:"inventory_id" binding:"required,uuid"`
	Quantity       int    `json:"quantity" binding:"required,min=1"`
	QuantityBefore int    `json:"quantity_before" binding:"required,min=0"`
	OrderID        string `json:"order_id" binding:"required,uuid"`
	Notes          string `json:"notes,omitempty"`
}

// RecordAdjustmentRequest is the request body for recording an adjustment.
type RecordAdjustmentRequest struct {
	InventoryID    string `json:"inventory_id" binding:"required,uuid"`
	Quantity       int    `json:"quantity" binding:"required"`
	QuantityBefore int    `json:"quantity_before" binding:"required,min=0"`
	Reason         string `json:"reason" binding:"required,min=3"`
	Notes          string `json:"notes,omitempty"`
}

// RecordTransferRequest is the request body for recording a transfer.
type RecordTransferRequest struct {
	InventoryID      string  `json:"inventory_id" binding:"required,uuid"`
	Quantity         int     `json:"quantity" binding:"required,min=1"`
	QuantityBefore   int     `json:"quantity_before" binding:"required,min=0"`
	FromWarehouseID  string  `json:"from_warehouse_id" binding:"required,uuid"`
	ToWarehouseID    string  `json:"to_warehouse_id" binding:"required,uuid"`
	FromLocationCode *string `json:"from_location_code,omitempty"`
	ToLocationCode   *string `json:"to_location_code,omitempty"`
	Notes            string  `json:"notes,omitempty"`
}

// RecordDamageRequest is the request body for recording damage.
type RecordDamageRequest struct {
	InventoryID    string `json:"inventory_id" binding:"required,uuid"`
	Quantity       int    `json:"quantity" binding:"required,min=1"`
	QuantityBefore int    `json:"quantity_before" binding:"required,min=0"`
	Reason         string `json:"reason" binding:"required,min=3"`
	Notes          string `json:"notes,omitempty"`
}

// RecordReturnRequest is the request body for recording a return.
type RecordReturnRequest struct {
	InventoryID    string `json:"inventory_id" binding:"required,uuid"`
	Quantity       int    `json:"quantity" binding:"required,min=1"`
	QuantityBefore int    `json:"quantity_before" binding:"required,min=0"`
	OrderID        string `json:"order_id" binding:"required,uuid"`
	Notes          string `json:"notes,omitempty"`
}

// ============================================================================
// Response DTOs
// ============================================================================

// MovementResponse represents stock movement response
type MovementResponse struct {
	ID                 string    `json:"id"`
	InventoryID        string    `json:"inventory_id"`
	Type               string    `json:"type"`
	Quantity           int       `json:"quantity"`
	QuantityBeforeMove int       `json:"quantity_before_move"`
	QuantityAfterMove  int       `json:"quantity_after_move"`
	FromWarehouseID    *string   `json:"from_warehouse_id,omitempty"`
	FromLocationCode   *string   `json:"from_location_code,omitempty"`
	ToWarehouseID      *string   `json:"to_warehouse_id,omitempty"`
	ToLocationCode     *string   `json:"to_location_code,omitempty"`
	ReferenceType      string    `json:"reference_type,omitempty"`
	ReferenceID        *string   `json:"reference_id,omitempty"`
	UnitCostCents      *int64    `json:"unit_cost_cents,omitempty"`
	TotalCostCents     *int64    `json:"total_cost_cents,omitempty"`
	CurrencyCode       *string   `json:"currency_code,omitempty"`
	Reason             string    `json:"reason,omitempty"`
	Notes              string    `json:"notes,omitempty"`
	CreatedBy          string    `json:"created_by"`
	MovementDate       time.Time `json:"movement_date"`
	CreatedAt          time.Time `json:"created_at"`
}

// MovementSummaryResponse is the response body for inventory movement summary.
type MovementSummaryResponse struct {
	InventoryID string `json:"inventory_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	TotalIn     int    `json:"total_in"`
	TotalOut    int    `json:"total_out"`
	NetChange   int    `json:"net_change"`
}

// ToMovementResponse converts stock movement entity to response DTO
func ToMovementResponse(sm *aggregate.StockMovement) *MovementResponse {
	response := &MovementResponse{
		ID:                 sm.ID.String(),
		InventoryID:        sm.InventoryID.String(),
		Type:               string(sm.Type),
		Quantity:           sm.Quantity,
		QuantityBeforeMove: sm.QuantityBeforeMove,
		QuantityAfterMove:  sm.QuantityAfterMove,
		ReferenceType:      sm.ReferenceType,
		Reason:             sm.Reason,
		Notes:              sm.Notes,
		CreatedBy:          sm.CreatedBy.String(),
		MovementDate:       sm.MovementDate,
		CreatedAt:          sm.CreatedAt,
	}

	if sm.FromWarehouseID != nil {
		fromWarehouse := sm.FromWarehouseID.String()
		response.FromWarehouseID = &fromWarehouse
	}
	if sm.ToWarehouseID != nil {
		toWarehouse := sm.ToWarehouseID.String()
		response.ToWarehouseID = &toWarehouse
	}
	if sm.FromLocationCode != nil {
		response.FromLocationCode = sm.FromLocationCode
	}
	if sm.ToLocationCode != nil {
		response.ToLocationCode = sm.ToLocationCode
	}
	if sm.ReferenceID != nil {
		refID := sm.ReferenceID.String()
		response.ReferenceID = &refID
	}
	if sm.UnitCostCents != nil {
		response.UnitCostCents = sm.UnitCostCents
	}
	if sm.TotalCostCents != nil {
		response.TotalCostCents = sm.TotalCostCents
	}
	if sm.CurrencyCode != nil {
		response.CurrencyCode = sm.CurrencyCode
	}

	return response
}

// ToMovementResponseList converts slice of stock movement entities to response DTOs
func ToMovementResponseList(movements []*aggregate.StockMovement) []*MovementResponse {
	responses := make([]*MovementResponse, len(movements))
	for i, movement := range movements {
		responses[i] = ToMovementResponse(movement)
	}
	return responses
}
