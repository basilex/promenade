package ordermgmt

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/order"
	orderHTTP "github.com/basilex/promenade/internal/contexts/order-mgmt/order/adapter/http"
	orderRepo "github.com/basilex/promenade/internal/contexts/order-mgmt/order/adapter/repository/postgres"
)

// Router handles all Order Management context routes
type Router struct {
	orderHandler *orderHTTP.OrderHandler
}

// NewRouter creates a new Order Management context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Order aggregate
	orderRepository := orderRepo.NewOrderRepository(db)
	orderUseCase := order.NewUseCase(orderRepository)
	orderHandler := orderHTTP.NewOrderHandler(orderUseCase)

	return &Router{
		orderHandler: orderHandler,
	}
}

// RegisterRoutes registers all Order Management context routes
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	orderMgmt := api.Group("/order-mgmt")
	{
		// Order routes
		orders := orderMgmt.Group("/orders")
		{
			// Order CRUD
			orders.POST("", r.orderHandler.Create)                         // Create order
			orders.GET("/:id", r.orderHandler.GetByID)                     // Get order by ID
			orders.GET("/number/:order_number", r.orderHandler.GetByOrderNumber) // Get by order number
			orders.GET("", r.orderHandler.List)                            // List orders
			
			// Query routes
			orders.GET("/customer/:customer_id", r.orderHandler.ListByCustomer) // List by customer
			orders.GET("/status/:status", r.orderHandler.ListByStatus)          // List by status
			
			// Order lifecycle
			orders.POST("/:id/confirm", r.orderHandler.Confirm)         // Confirm order
			orders.POST("/:id/process", r.orderHandler.StartProcessing) // Start processing
			orders.POST("/:id/fulfill", r.orderHandler.MarkFulfilled)   // Mark fulfilled
			orders.POST("/:id/cancel", r.orderHandler.Cancel)           // Cancel order
			
			// Order lines
			orders.POST("/:id/lines", r.orderHandler.AddLine)                      // Add line
			orders.DELETE("/:id/lines/:line_id", r.orderHandler.RemoveLine)        // Remove line
			orders.PUT("/:id/lines/:line_id", r.orderHandler.UpdateLineQuantity)   // Update line quantity
		}

		// Future aggregates:
		// - Contracts: /order-mgmt/contracts
		// - Fulfillment: /order-mgmt/fulfillments
	}
}
