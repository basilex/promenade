package http

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/basilex/promenade/internal/contexts/warehouse/stockmovement"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// StockMovementHandler handles HTTP requests for stock movement operations
type StockMovementHandler struct {
	usecase stockmovement.IUseCase
}

// NewStockMovementHandler creates a new stock movement handler
func NewStockMovementHandler(usecase stockmovement.IUseCase) *StockMovementHandler {
	return &StockMovementHandler{
		usecase: usecase,
	}
}

// RecordMovement records a generic stock movement
func (h *StockMovementHandler) RecordMovement(c *gin.Context) {
	var req RecordMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	inventoryID, err := uuidv7.Parse(req.InventoryID)
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	createdBy, err := uuidv7.Parse(req.CreatedBy)
	if err != nil {
		response.BadRequest(c, "Invalid created_by format")
		return
	}

	movementType := stockmovement.MovementType(req.Type)

	// Note: quantityBefore should come from Inventory aggregate
	// For now, we pass the value from request (in production, fetch from inventory)
	movement, err := h.usecase.RecordMovement(
		c.Request.Context(),
		inventoryID,
		movementType,
		req.Quantity,
		req.QuantityBefore, // Added missing parameter
		createdBy,
	)
	if err != nil {
		if errors.Is(err, stockmovement.ErrStockMovementNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, ToMovementResponse(movement))
}

// GetByID retrieves stock movement by ID
func (h *StockMovementHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid movement ID format")
		return
	}

	movement, err := h.usecase.GetMovement(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, stockmovement.ErrStockMovementNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToMovementResponse(movement))
}

// GetByInventory retrieves stock movements for an inventory with pagination
func (h *StockMovementHandler) GetByInventory(c *gin.Context) {
	inventoryID, err := uuidv7.Parse(c.Param("inventory_id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	page := 1
	pageSize := 20

	movements, total, err := h.usecase.GetMovementsByInventory(
		c.Request.Context(),
		inventoryID,
		page,
		pageSize,
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	result := map[string]interface{}{
		"movements": ToMovementResponseList(movements),
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}

	response.Success(c, result)
}

// GetByReference retrieves stock movements by reference
func (h *StockMovementHandler) GetByReference(c *gin.Context) {
	referenceType := c.Query("type")
	if referenceType == "" {
		response.BadRequest(c, "Reference type is required")
		return
	}

	referenceIDStr := c.Query("id")
	if referenceIDStr == "" {
		response.BadRequest(c, "Reference ID is required")
		return
	}

	referenceID, err := uuidv7.Parse(referenceIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid reference ID format")
		return
	}

	movements, err := h.usecase.GetMovementsByReference(
		c.Request.Context(),
		referenceType,
		referenceID,
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToMovementResponseList(movements))
}

// GetByType retrieves stock movements by type with pagination and date range
func (h *StockMovementHandler) GetByType(c *gin.Context) {
	movementType := stockmovement.MovementType(c.Param("type"))

	// Parse date range (required for usecase interface)
	// Default to last 30 days if not provided
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	page := 1
	pageSize := 20

	movements, total, err := h.usecase.GetMovementsByType(
		c.Request.Context(),
		movementType,
		startDate,
		endDate,
		page,
		pageSize,
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	result := map[string]interface{}{
		"movements": ToMovementResponseList(movements),
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}

	response.Success(c, result)
}

// GetRecent retrieves recent stock movements
func (h *StockMovementHandler) GetRecent(c *gin.Context) {
	limit := 50

	movements, err := h.usecase.GetRecentMovements(c.Request.Context(), limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToMovementResponseList(movements))
}

// GetInventorySummary retrieves stock movement summary for an inventory
func (h *StockMovementHandler) GetInventorySummary(c *gin.Context) {
	inventoryID, err := uuidv7.Parse(c.Param("inventory_id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	// Default to last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	totalIn, totalOut, err := h.usecase.GetInventorySummary(
		c.Request.Context(),
		inventoryID,
		startDate,
		endDate,
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	summary := map[string]interface{}{
		"total_in":   totalIn,
		"total_out":  totalOut,
		"net_change": totalIn - totalOut,
		"start_date": startDate,
		"end_date":   endDate,
	}

	response.Success(c, summary)
}
