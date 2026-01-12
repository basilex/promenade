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
// @Summary Record stock movement
// @Description Records a generic stock movement (receipt, reservation, commit, adjustment, transfer, damage, return)
// @Tags Stock Movements
// @Accept json
// @Produce json
// @Param movement body RecordMovementRequest true "Stock movement details"
// @Success 201 {object} response.Response{data=MovementResponse} "Movement recorded successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /warehouse/stock-movements [post]
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
		switch {
		case errors.Is(err, stockmovement.ErrStockMovementNil):
			response.BadRequest(c, "Stock movement cannot be nil")
		case errors.Is(err, stockmovement.ErrStockMovementValidationFailed):
			response.BadRequest(c, "Stock movement validation failed")
		case errors.Is(err, stockmovement.ErrStockMovementCreateFailed):
			response.InternalError(c, "Failed to create stock movement")
		case errors.Is(err, stockmovement.ErrStockMovementSaveFailed):
			response.InternalError(c, "Failed to save stock movement")
		default:
			response.InternalError(c, "Failed to record stock movement")
		}
		return
	}

	response.Created(c, ToMovementResponse(movement))
}

// GetByID retrieves stock movement by ID
// @Summary Get stock movement by ID
// @Description Retrieves stock movement details by its unique identifier
// @Tags Stock Movements
// @Produce json
// @Param id path string true "Movement ID" format(uuid)
// @Success 200 {object} response.Response{data=MovementResponse} "Movement details"
// @Failure 400 {object} response.Response "Invalid movement ID"
// @Failure 404 {object} response.Response "Movement not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /warehouse/stock-movements/{id} [get]
func (h *StockMovementHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid movement ID format")
		return
	}

	movement, err := h.usecase.GetMovement(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, stockmovement.ErrStockMovementNotFound):
			response.NotFound(c, "Stock movement not found")
		case errors.Is(err, stockmovement.ErrStockMovementGetFailed):
			response.InternalError(c, "Failed to retrieve stock movement")
		default:
			response.InternalError(c, "Failed to get stock movement")
		}
		return
	}

	response.Success(c, ToMovementResponse(movement))
}

// GetByInventory retrieves stock movements for an inventory with pagination
// @Summary Get stock movements by inventory
// @Description Retrieves paginated stock movements for a specific inventory item
// @Tags Stock Movements
// @Produce json
// @Param inventory_id path string true "Inventory ID" format(uuid)
// @Success 200 {object} response.Response{data=object{movements=[]MovementResponse,total=int,page=int,page_size=int}} "Paginated movements"
// @Failure 400 {object} response.Response "Invalid inventory ID"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /warehouse/stock-movements/inventory/{inventory_id} [get]
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
		switch {
		case errors.Is(err, stockmovement.ErrInvalidPagination):
			response.BadRequest(c, "Invalid pagination parameters")
		case errors.Is(err, stockmovement.ErrStockMovementListFailed):
			response.InternalError(c, "Failed to list stock movements")
		default:
			response.InternalError(c, "Failed to retrieve stock movements")
		}
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
// @Summary Get stock movements by reference
// @Description Retrieves stock movements linked to a specific reference (order, PO, adjustment)
// @Tags Stock Movements
// @Produce json
// @Param type query string true "Reference type (order, po, adjustment)"
// @Param id query string true "Reference ID" format(uuid)
// @Success 200 {object} response.Response{data=[]MovementResponse} "Movements for reference"
// @Failure 400 {object} response.Response "Invalid parameters"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /warehouse/stock-movements/reference [get]
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
		switch {
		case errors.Is(err, stockmovement.ErrReferenceTypeRequired):
			response.BadRequest(c, "Reference type is required")
		case errors.Is(err, stockmovement.ErrReferenceIDRequired):
			response.BadRequest(c, "Reference ID is required")
		case errors.Is(err, stockmovement.ErrStockMovementListFailed):
			response.InternalError(c, "Failed to list stock movements by reference")
		default:
			response.InternalError(c, "Failed to retrieve stock movements")
		}
		return
	}

	response.Success(c, ToMovementResponseList(movements))
}

// GetByType retrieves stock movements by type with pagination and date range
// @Summary Get stock movements by type
// @Description Retrieves paginated stock movements of a specific type (last 30 days by default)
// @Tags Stock Movements
// @Produce json
// @Param type path string true "Movement type (receipt, reservation, reservation_release, commit, adjustment, transfer, damage, return)"
// @Success 200 {object} response.Response{data=object{movements=[]MovementResponse,total=int,page=int,page_size=int}} "Paginated movements"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /warehouse/stock-movements/type/{type} [get]
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
		switch {
		case errors.Is(err, stockmovement.ErrInvalidPagination):
			response.BadRequest(c, "Invalid pagination parameters")
		case errors.Is(err, stockmovement.ErrStockMovementListFailed):
			response.InternalError(c, "Failed to list stock movements by type")
		default:
			response.InternalError(c, "Failed to retrieve stock movements")
		}
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
// @Summary Get recent stock movements
// @Description Retrieves the 50 most recent stock movements across all inventory items
// @Tags Stock Movements
// @Produce json
// @Success 200 {object} response.Response{data=[]MovementResponse} "Recent movements"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /warehouse/stock-movements/recent [get]
func (h *StockMovementHandler) GetRecent(c *gin.Context) {
	limit := 50

	movements, err := h.usecase.GetRecentMovements(c.Request.Context(), limit)
	if err != nil {
		switch {
		case errors.Is(err, stockmovement.ErrStockMovementListFailed):
			response.InternalError(c, "Failed to list recent movements")
		default:
			response.InternalError(c, "Failed to retrieve recent movements")
		}
		return
	}

	response.Success(c, ToMovementResponseList(movements))
}

// GetInventorySummary retrieves stock movement summary for an inventory
// @Summary Get inventory movement summary
// @Description Retrieves total in, total out, and net change for an inventory (last 30 days by default)
// @Tags Stock Movements
// @Produce json
// @Param inventory_id path string true "Inventory ID" format(uuid)
// @Success 200 {object} response.Response{data=object{total_in=int,total_out=int,net_change=int,start_date=string,end_date=string}} "Movement summary"
// @Failure 400 {object} response.Response "Invalid inventory ID"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /warehouse/stock-movements/summary/{inventory_id} [get]
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
		switch {
		case errors.Is(err, stockmovement.ErrStockMovementSummaryFailed):
			response.InternalError(c, "Failed to calculate inventory summary")
		default:
			response.InternalError(c, "Failed to retrieve inventory summary")
		}
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
