package http

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// InventoryHandler handles HTTP requests for inventory operations
type InventoryHandler struct {
	usecase inventory.IUseCase
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(usecase inventory.IUseCase) *InventoryHandler {
	return &InventoryHandler{
		usecase: usecase,
	}
}

// Create creates a new inventory item
// @Summary Create inventory
// @Description Create a new inventory item
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body CreateInventoryRequest true "Inventory creation request"
// @Success 201 {object} response.Response{data=InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory [post]
func (h *InventoryHandler) Create(c *gin.Context) {
	var req CreateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Parse UUIDs
	productID, err := uuidv7.Parse(req.ProductID)
	if err != nil {
		response.BadRequest(c, "Invalid product ID format")
		return
	}

	createdBy, err := uuidv7.Parse(req.CreatedBy)
	if err != nil {
		response.BadRequest(c, "Invalid created_by format")
		return
	}

	// Create inventory via use case
	inv, err := h.usecase.CreateInventory(c.Request.Context(), productID, req.SKU, req.ProductName, req.WarehouseID, createdBy)
	if err != nil {
		// Map domain errors to user-friendly messages
		switch {
		case errors.Is(err, inventory.ErrInventorySKURequired):
			response.BadRequest(c, "SKU is required")
		case errors.Is(err, inventory.ErrInventoryProductNameRequired):
			response.BadRequest(c, "Product name is required")
		case errors.Is(err, inventory.ErrInventoryWarehouseRequired):
			response.BadRequest(c, "Warehouse ID is required")
		case errors.Is(err, inventory.ErrInventoryCreatedByRequired):
			response.BadRequest(c, "Created by user ID is required")
		case errors.Is(err, inventory.ErrInventorySKUExists):
			response.BadRequest(c, "SKU already exists")
		default:
			// System errors - generic message (no detail leakage)
			response.InternalError(c, "Failed to create inventory")
		}
		return
	}

	// Set optional fields if provided
	if req.LocationCode != "" {
		inv.LocationCode = req.LocationCode
	}
	if req.LocationZone != "" {
		inv.LocationZone = req.LocationZone
	}

	response.Created(c, ToInventoryResponse(inv))
}

// GetByID retrieves inventory by ID
// @Summary Get inventory by ID
// @Description Get inventory item by ID
// @Tags Inventory
// @Produce json
// @Param id path string true "Inventory ID"
// @Success 200 {object} response.Response{data=InventoryResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/{id} [get]
func (h *InventoryHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	inv, err := h.usecase.GetInventory(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, inventory.ErrInventoryNotFound) {
			response.NotFound(c, "Inventory not found")
			return
		}
		response.InternalError(c, "Failed to retrieve inventory")
		return
	}

	response.Success(c, ToInventoryResponse(inv))
}

// GetBySKU retrieves inventory by SKU
// @Summary Get inventory by SKU
// @Description Get inventory item by unique SKU
// @Tags Inventory
// @Produce json
// @Param sku path string true "SKU"
// @Success 200 {object} response.Response{data=InventoryResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/sku/{sku} [get]
func (h *InventoryHandler) GetBySKU(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		response.BadRequest(c, "SKU is required")
		return
	}

	inv, err := h.usecase.GetBySKU(c.Request.Context(), sku)
	if err != nil {
		if errors.Is(err, inventory.ErrInventoryNotFound) {
			response.NotFound(c, "Inventory not found")
			return
		}
		response.InternalError(c, "Failed to retrieve inventory")
		return
	}

	response.Success(c, ToInventoryResponse(inv))
}

// GetByProductID retrieves all inventory for a product
// @Summary Get inventory by product ID
// @Description Get all inventory items for a specific product
// @Tags Inventory
// @Produce json
// @Param product_id path string true "Product ID"
// @Success 200 {object} response.Response{data=[]InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/product/{product_id} [get]
func (h *InventoryHandler) GetByProductID(c *gin.Context) {
	productID, err := uuidv7.Parse(c.Param("product_id"))
	if err != nil {
		response.BadRequest(c, "Invalid product ID format")
		return
	}

	items, err := h.usecase.GetByProductID(c.Request.Context(), productID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve inventory")
		return
	}

	response.Success(c, ToInventoryResponseList(items))
}

// GetByWarehouse retrieves all inventory in a warehouse
// @Summary Get inventory by warehouse
// @Description Get all inventory items in a specific warehouse
// @Tags Inventory
// @Produce json
// @Param warehouse_id path string true "Warehouse ID"
// @Success 200 {object} response.Response{data=[]InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/warehouse/{warehouse_id} [get]
func (h *InventoryHandler) GetByWarehouse(c *gin.Context) {
	warehouseID := c.Param("warehouse_id")
	if warehouseID == "" {
		response.BadRequest(c, "Warehouse ID is required")
		return
	}

	items, err := h.usecase.GetByWarehouse(c.Request.Context(), warehouseID)
	if err != nil {
		response.InternalError(c, "Failed to retrieve inventory")
		return
	}

	response.Success(c, ToInventoryResponseList(items))
}

// List retrieves paginated inventory items
// @Summary List inventory
// @Description Get paginated list of inventory items
// @Tags Inventory
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory [get]
func (h *InventoryHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.usecase.ListInventory(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list inventory")
		return
	}

	response.SuccessWithPagination(c, ToInventoryResponseList(items), total, page, pageSize)
}

// GetLowStock retrieves items below reorder point
// @Summary Get low stock items
// @Description Get all inventory items below reorder point
// @Tags Inventory
// @Produce json
// @Success 200 {object} response.Response{data=[]InventoryResponse}
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/low-stock [get]
func (h *InventoryHandler) GetLowStock(c *gin.Context) {
	items, err := h.usecase.GetLowStock(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Failed to retrieve low stock items")
		return
	}

	response.Success(c, ToInventoryResponseList(items))
}

// Update updates inventory item
// @Summary Update inventory
// @Description Update inventory item details
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory ID"
// @Param request body UpdateInventoryRequest true "Inventory update request"
// @Success 200 {object} response.Response{data=InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/{id} [put]
func (h *InventoryHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	var req UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get current inventory
	inv, err := h.usecase.GetInventory(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, inventory.ErrInventoryNotFound) {
			response.NotFound(c, "Inventory not found")
			return
		}
		response.InternalError(c, "Failed to retrieve inventory")
		return
	}

	// Apply updates
	if req.LocationCode != nil {
		inv.LocationCode = *req.LocationCode
	}
	if req.LocationZone != nil {
		inv.LocationZone = *req.LocationZone
	}
	if req.ReorderPoint != nil {
		inv.ReorderPoint = *req.ReorderPoint
	}
	if req.ReorderQuantity != nil {
		inv.ReorderQuantity = *req.ReorderQuantity
	}
	if req.UnitCostCents != nil {
		inv.UnitCostCents = int64(*req.UnitCostCents)
	}
	if req.CurrencyCode != nil {
		inv.CurrencyCode = *req.CurrencyCode
	}
	if req.Notes != nil {
		inv.Notes = *req.Notes
	}
	if req.IsActive != nil {
		inv.IsActive = *req.IsActive
	}

	if err := h.usecase.UpdateInventory(c.Request.Context(), inv); err != nil {
		// Map domain errors to user-friendly messages
		switch {
		case errors.Is(err, inventory.ErrInventoryNotFound):
			response.NotFound(c, "Inventory not found")
		case errors.Is(err, inventory.ErrInventoryNil):
			response.BadRequest(c, "Invalid inventory data")
		default:
			// System errors - generic message (no detail leakage)
			response.InternalError(c, "Failed to update inventory")
		}
		return
	}

	response.Success(c, ToInventoryResponse(inv))
}

// Delete soft-deletes inventory item
// @Summary Delete inventory
// @Description Soft-delete inventory item
// @Tags Inventory
// @Produce json
// @Param id path string true "Inventory ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/{id} [delete]
func (h *InventoryHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	if err := h.usecase.DeleteInventory(c.Request.Context(), id); err != nil {
		if errors.Is(err, inventory.ErrInventoryNotFound) {
			response.NotFound(c, "Inventory not found")
			return
		}
		response.InternalError(c, "Failed to delete inventory")
		return
	}

	response.Success(c, gin.H{"message": "Inventory deleted successfully"})
}

// ReceiveStock receives stock into inventory
// @Summary Receive stock
// @Description Receive stock into inventory (increases quantity)
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory ID"
// @Param request body ReceiveStockRequest true "Receive stock request"
// @Success 200 {object} response.Response{data=InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/{id}/receive [post]
func (h *InventoryHandler) ReceiveStock(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	var req ReceiveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	receivedBy, err := uuidv7.Parse(req.ReceivedBy)
	if err != nil {
		response.BadRequest(c, "Invalid received_by user ID format")
		return
	}

	inv, err := h.usecase.ReceiveStock(c.Request.Context(), id, req.Quantity, req.UnitCostCents, receivedBy)
	if err != nil {
		// Map domain errors to user-friendly messages
		switch {
		case errors.Is(err, inventory.ErrInventoryNotFound):
			response.NotFound(c, "Inventory not found")
		case errors.Is(err, inventory.ErrInventoryQuantityInvalid):
			response.BadRequest(c, "Quantity must be greater than 0")
		case errors.Is(err, inventory.ErrInventoryUnitCostNegative):
			response.BadRequest(c, "Unit cost cannot be negative")
		default:
			// System errors - generic message (no detail leakage)
			response.InternalError(c, "Failed to receive stock")
		}
		return
	}

	response.Success(c, ToInventoryResponse(inv))
}

// CommitStock commits stock from inventory
// @Summary Commit stock
// @Description Commit stock (decreases available, increases committed)
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory ID"
// @Param request body CommitStockRequest true "Commit stock request"
// @Success 200 {object} response.Response{data=InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/{id}/commit [post]
func (h *InventoryHandler) CommitStock(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	var req CommitStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	committedBy, err := uuidv7.Parse(req.CommittedBy)
	if err != nil {
		response.BadRequest(c, "Invalid committed_by user ID format")
		return
	}

	inv, err := h.usecase.CommitStock(c.Request.Context(), id, req.Quantity, committedBy)
	if err != nil {
		// Map domain errors to user-friendly messages
		switch {
		case errors.Is(err, inventory.ErrInventoryNotFound):
			response.NotFound(c, "Inventory not found")
		case errors.Is(err, inventory.ErrInventoryQuantityInvalid):
			response.BadRequest(c, "Quantity must be greater than 0")
		case errors.Is(err, inventory.ErrInventoryInsufficientStock):
			response.BadRequest(c, "Insufficient stock available")
		default:
			// System errors - generic message (no detail leakage)
			response.InternalError(c, "Failed to commit stock")
		}
		return
	}

	response.Success(c, ToInventoryResponse(inv))
}

// GetByLocation retrieves inventory by warehouse and location
// @Summary Get inventory by location
// @Description Get inventory items by warehouse ID and location code
// @Tags Inventory
// @Produce json
// @Param warehouse_id path string true "Warehouse ID"
// @Param location_code path string true "Location Code"
// @Success 200 {object} response.Response{data=[]InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/location/{warehouse_id}/{location_code} [get]
func (h *InventoryHandler) GetByLocation(c *gin.Context) {
	warehouseID := c.Param("warehouse_id")
	locationCode := c.Param("location_code")

	if warehouseID == "" {
		response.BadRequest(c, "Warehouse ID is required")
		return
	}
	if locationCode == "" {
		response.BadRequest(c, "Location code is required")
		return
	}

	items, err := h.usecase.GetByLocation(c.Request.Context(), warehouseID, locationCode)
	if err != nil {
		response.InternalError(c, "Failed to retrieve inventory")
		return
	}

	response.Success(c, ToInventoryResponseList(items))
}

// ReserveStock reserves stock for an order
// @Summary Reserve stock
// @Description Reserve stock for an order (decreases available, increases reserved)
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory ID"
// @Param request body ReserveStockRequest true "Reserve stock request"
// @Success 200 {object} response.Response{data=InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/{id}/reserve [post]
func (h *InventoryHandler) ReserveStock(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	var req ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	orderID, err := uuidv7.Parse(req.OrderID)
	if err != nil {
		response.BadRequest(c, "Invalid order ID format")
		return
	}

	reservedBy, err := uuidv7.Parse(req.ReservedBy)
	if err != nil {
		response.BadRequest(c, "Invalid reserved_by user ID format")
		return
	}

	// Get inventory
	inv, err := h.usecase.GetInventory(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, inventory.ErrInventoryNotFound) {
			response.NotFound(c, "Inventory not found")
			return
		}
		response.InternalError(c, "Failed to retrieve inventory")
		return
	}

	// Reserve stock
	if err := inv.ReserveStock(req.Quantity, orderID, reservedBy); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update inventory
	if err := h.usecase.UpdateInventory(c.Request.Context(), inv); err != nil {
		switch {
		case errors.Is(err, inventory.ErrVersionConflict):
			response.Conflict(c, "Inventory was modified by another process, please retry")
		default:
			response.InternalError(c, "Failed to update inventory")
		}
		return
	}

	response.Success(c, ToInventoryResponse(inv))
}

// ReleaseReservation releases reserved stock (Saga compensation)
// @Summary Release reservation
// @Description Release reserved stock when order is cancelled (Saga compensation)
// @Tags Inventory
// @Accept json
// @Produce json
// @Param id path string true "Inventory ID"
// @Param request body ReleaseReservationRequest true "Release reservation request"
// @Success 200 {object} response.Response{data=InventoryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/inventory/{id}/release [post]
func (h *InventoryHandler) ReleaseReservation(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid inventory ID format")
		return
	}

	var req ReleaseReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	orderID, err := uuidv7.Parse(req.OrderID)
	if err != nil {
		response.BadRequest(c, "Invalid order ID format")
		return
	}

	releasedBy, err := uuidv7.Parse(req.ReleasedBy)
	if err != nil {
		response.BadRequest(c, "Invalid released_by user ID format")
		return
	}

	// Get inventory
	inv, err := h.usecase.GetInventory(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, inventory.ErrInventoryNotFound) {
			response.NotFound(c, "Inventory not found")
			return
		}
		response.InternalError(c, "Failed to retrieve inventory")
		return
	}

	// Release reservation
	if err := inv.ReleaseReservation(req.Quantity, orderID, releasedBy); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update inventory
	if err := h.usecase.UpdateInventory(c.Request.Context(), inv); err != nil {
		switch {
		case errors.Is(err, inventory.ErrVersionConflict):
			response.Conflict(c, "Inventory was modified by another process, please retry")
		default:
			response.InternalError(c, "Failed to update inventory")
		}
		return
	}

	response.Success(c, ToInventoryResponse(inv))
}
