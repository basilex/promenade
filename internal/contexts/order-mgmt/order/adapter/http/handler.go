package http

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/order"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	usecase order.IUseCase
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(usecase order.IUseCase) *OrderHandler {
	return &OrderHandler{
		usecase: usecase,
	}
}

// Create handles POST /orders - Create order
// @Summary Create order
// @Description Create a new order
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order details"
// @Success 201 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create order
	created, err := h.usecase.CreateOrder(c.Request.Context(), req.CustomerID, req.CompanyID, req.Currency)
	if err != nil {
		response.InternalError(c, "Failed to create order")
		return
	}

	response.Created(c, ToOrderResponse(created))
}

// GetByID handles GET /orders/:id - Get order by ID
// @Summary Get order by ID
// @Description Get order details by ID
// @Tags orders
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id} [get]
func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	o, err := h.usecase.GetOrder(c.Request.Context(), id)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to retrieve order")
		return
	}

	response.Success(c, ToOrderResponse(o))
}

// GetByOrderNumber handles GET /orders/number/:order_number - Get order by order number
// @Summary Get order by order number
// @Description Get order details by order number
// @Tags orders
// @Produce json
// @Param order_number path string true "Order Number (e.g., ORD-2025-000001)"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/number/{order_number} [get]
func (h *OrderHandler) GetByOrderNumber(c *gin.Context) {
	orderNumber := c.Param("order_number")

	o, err := h.usecase.GetOrderByNumber(c.Request.Context(), orderNumber)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to retrieve order by number")
		return
	}

	response.Success(c, ToOrderResponse(o))
}

// List handles GET /orders - List orders with pagination
// @Summary List orders
// @Description List all orders with pagination
// @Tags orders
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders [get]
func (h *OrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.usecase.ListOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list orders")
		return
	}

	response.SuccessWithPagination(c, ToOrderListResponse(orders), total, page, pageSize)
}

// ListByCustomer handles GET /orders/customer/:customer_id - List orders by customer
// @Summary List orders by customer
// @Description List all orders for a specific customer
// @Tags orders
// @Produce json
// @Param customer_id path string true "Customer ID (UUID)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/customer/{customer_id} [get]
func (h *OrderHandler) ListByCustomer(c *gin.Context) {
	customerID, err := uuidv7.Parse(c.Param("customer_id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.usecase.ListOrdersByCustomer(c.Request.Context(), customerID, page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list customer orders")
		return
	}

	response.SuccessWithPagination(c, ToOrderListResponse(orders), total, page, pageSize)
}

// ListByStatus handles GET /orders/status/:status - List orders by status
// @Summary List orders by status
// @Description List all orders with a specific status
// @Tags orders
// @Produce json
// @Param status path string true "Status" Enums(pending, confirmed, processing, fulfilled, cancelled)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/status/{status} [get]
func (h *OrderHandler) ListByStatus(c *gin.Context) {
	statusStr := c.Param("status")

	// Validate status
	validStatuses := map[string]bool{
		"pending":    true,
		"confirmed":  true,
		"processing": true,
		"fulfilled":  true,
		"cancelled":  true,
	}
	if !validStatuses[statusStr] {
		response.BadRequest(c, "invalid status: must be one of: pending, confirmed, processing, fulfilled, cancelled")
		return
	}

	status := order.OrderStatus(statusStr)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.usecase.ListOrdersByStatus(c.Request.Context(), status, page, pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list orders by status")
		return
	}

	response.SuccessWithPagination(c, ToOrderListResponse(orders), total, page, pageSize)
}

// AddLine handles POST /orders/:id/lines - Add line to order
// @Summary Add line to order
// @Description Add a new line item to an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param line body AddOrderLineRequest true "Line details"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/lines [post]
func (h *OrderHandler) AddLine(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	var req AddOrderLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create Money value object
	unitPrice := valueobject.Money{
		Amount:   req.UnitPrice,
		Currency: req.Currency,
	}

	updated, err := h.usecase.AddOrderLine(c.Request.Context(), orderID, req.ProductID, req.Quantity, unitPrice)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to add order line")
		return
	}

	response.Success(c, ToOrderResponse(updated))
}

// RemoveLine handles DELETE /orders/:id/lines/:line_id - Remove line from order
// @Summary Remove line from order
// @Description Remove a line item from an order
// @Tags orders
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param line_id path string true "Line ID (UUID)"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/lines/{line_id} [delete]
func (h *OrderHandler) RemoveLine(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	lineID, err := uuidv7.Parse(c.Param("line_id"))
	if err != nil {
		response.BadRequest(c, "invalid line ID")
		return
	}

	updated, err := h.usecase.RemoveOrderLine(c.Request.Context(), orderID, lineID)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if errors.Is(err, order.ErrOrderLineNotFound) {
		response.NotFound(c, "Order line not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to remove order line")
		return
	}

	response.Success(c, ToOrderResponse(updated))
}

// UpdateLineQuantity handles PUT /orders/:id/lines/:line_id - Update line quantity
// @Summary Update line quantity
// @Description Update the quantity of a line item in an order
// @Tags orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param line_id path string true "Line ID (UUID)"
// @Param quantity body UpdateOrderLineRequest true "New quantity"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/lines/{line_id} [put]
func (h *OrderHandler) UpdateLineQuantity(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	lineID, err := uuidv7.Parse(c.Param("line_id"))
	if err != nil {
		response.BadRequest(c, "invalid line ID")
		return
	}

	var req UpdateOrderLineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.usecase.UpdateOrderLineQuantity(c.Request.Context(), orderID, lineID, req.Quantity)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if errors.Is(err, order.ErrOrderLineNotFound) {
		response.NotFound(c, "Order line not found")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to update line quantity")
		return
	}

	response.Success(c, ToOrderResponse(updated))
}

// Confirm handles POST /orders/:id/confirm - Confirm order
// @Summary Confirm order
// @Description Confirm an order (pending → confirmed)
// @Tags orders
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/confirm [post]
func (h *OrderHandler) Confirm(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	updated, err := h.usecase.ConfirmOrder(c.Request.Context(), orderID)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if errors.Is(err, order.ErrOrderAlreadyConfirmed) {
		response.BadRequest(c, "Order already confirmed")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to confirm order")
		return
	}

	response.Success(c, ToOrderResponse(updated))
}

// StartProcessing handles POST /orders/:id/process - Start processing order
// @Summary Start processing order
// @Description Start processing an order (confirmed → processing)
// @Tags orders
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/process [post]
func (h *OrderHandler) StartProcessing(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	updated, err := h.usecase.StartProcessing(c.Request.Context(), orderID)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if errors.Is(err, order.ErrOrderNotConfirmed) {
		response.BadRequest(c, "Order not confirmed")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to start order processing")
		return
	}

	response.Success(c, ToOrderResponse(updated))
}

// MarkFulfilled handles POST /orders/:id/fulfill - Mark order as fulfilled
// @Summary Mark order as fulfilled
// @Description Mark an order as fulfilled (processing → fulfilled)
// @Tags orders
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/fulfill [post]
func (h *OrderHandler) MarkFulfilled(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	updated, err := h.usecase.MarkFulfilled(c.Request.Context(), orderID)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if errors.Is(err, order.ErrOrderNotProcessing) {
		response.BadRequest(c, "Order not processing")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to mark order as fulfilled")
		return
	}

	response.Success(c, ToOrderResponse(updated))
}

// Cancel handles POST /orders/:id/cancel - Cancel order
// @Summary Cancel order
// @Description Cancel an order (any status → cancelled)
// @Tags orders
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} response.Response{data=OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /orders/{id}/cancel [post]
func (h *OrderHandler) Cancel(c *gin.Context) {
	orderID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid order ID")
		return
	}

	updated, err := h.usecase.CancelOrder(c.Request.Context(), orderID)
	if errors.Is(err, order.ErrOrderNotFound) {
		response.NotFound(c, "Order not found")
		return
	}
	if errors.Is(err, order.ErrOrderAlreadyCancelled) {
		response.BadRequest(c, "Order already cancelled")
		return
	}
	if errors.Is(err, order.ErrOrderAlreadyFulfilled) {
		response.BadRequest(c, "Order already fulfilled")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to cancel order")
		return
	}

	response.Success(c, ToOrderResponse(updated))
}
