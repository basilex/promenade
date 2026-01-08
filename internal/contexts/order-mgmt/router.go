package ordermgmt

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract"
	contractHTTP "github.com/basilex/promenade/internal/contexts/order-mgmt/contract/adapter/http"
	contractRepo "github.com/basilex/promenade/internal/contexts/order-mgmt/contract/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/order-mgmt/order"
	orderHTTP "github.com/basilex/promenade/internal/contexts/order-mgmt/order/adapter/http"
	orderRepo "github.com/basilex/promenade/internal/contexts/order-mgmt/order/adapter/repository/postgres"
)

// Router handles all Order Management context routes
type Router struct {
	orderHandler    *orderHTTP.OrderHandler
	contractHandler *contractHTTP.ContractHandler
}

// NewRouter creates a new Order Management context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Order aggregate
	orderRepository := orderRepo.NewOrderRepository(db)
	orderUseCase := order.NewUseCase(orderRepository)
	orderHandler := orderHTTP.NewOrderHandler(orderUseCase)

	// Initialize Contract aggregate
	contractRepository := contractRepo.NewContractRepository(db)
	contractUseCase := contract.NewUseCase(contractRepository)
	contractHandler := contractHTTP.NewContractHandler(contractUseCase)

	return &Router{
		orderHandler:    orderHandler,
		contractHandler: contractHandler,
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

			// Contract routes (nested under orders)
			orders.GET("/:order_id/contracts", r.contractHandler.ListByOrder) // List contracts for order
		}

		// Contract routes
		contracts := orderMgmt.Group("/contracts")
		{
			// Contract CRUD
			contracts.POST("", r.contractHandler.Create)                // Create contract
			contracts.GET("/:id", r.contractHandler.GetByID)            // Get contract by ID
			contracts.PUT("/:id", r.contractHandler.Update)             // Update contract
			contracts.DELETE("/:id", r.contractHandler.Delete)          // Delete contract
			contracts.GET("", r.contractHandler.List)                   // List contracts

			// Contract lifecycle
			contracts.POST("/:id/submit", r.contractHandler.SubmitForSignature) // Submit for signature
			contracts.POST("/:id/sign", r.contractHandler.Sign)                 // Sign contract
			contracts.POST("/:id/complete", r.contractHandler.Complete)         // Complete contract
			contracts.POST("/:id/terminate", r.contractHandler.Terminate)       // Terminate contract
			contracts.POST("/:id/renew", r.contractHandler.Renew)               // Renew contract

			// Contract management
			contracts.PUT("/:id/expiration", r.contractHandler.SetExpirationDate) // Set expiration date
		}

		// Future aggregates:
		// - Fulfillment: /order-mgmt/fulfillments
	}
}
