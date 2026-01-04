package billing

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/billing/invoice"
	invoiceHTTP "github.com/basilex/promenade/internal/contexts/billing/invoice/adapter/http"
	invoiceRepo "github.com/basilex/promenade/internal/contexts/billing/invoice/adapter/repository/postgres"
)

// Router handles all Billing context routes
type Router struct {
	invoiceHandler *invoiceHTTP.InvoiceHandler
}

// NewRouter creates a new Billing context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Invoice aggregate
	invoiceRepository := invoiceRepo.NewInvoiceRepository(db)
	invoiceUseCase := invoice.NewUseCase(invoiceRepository)
	invoiceHandler := invoiceHTTP.NewInvoiceHandler(invoiceUseCase)

	return &Router{
		invoiceHandler: invoiceHandler,
	}
}

// RegisterRoutes registers all Billing context routes
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	billing := api.Group("/billing")
	{
		// Invoice routes
		invoices := billing.Group("/invoices")
		{
			// Invoice CRUD
			invoices.POST("", r.invoiceHandler.Create)                         // Create invoice
			invoices.GET("/:id", r.invoiceHandler.GetByID)                     // Get invoice by ID
			invoices.GET("/number/:number", r.invoiceHandler.GetByNumber)      // Get by invoice number
			invoices.DELETE("/:id", r.invoiceHandler.Delete)                   // Delete invoice (draft only)
			invoices.GET("", r.invoiceHandler.List)                            // List invoices (with filters)
			invoices.GET("/overdue", r.invoiceHandler.ListOverdue)             // List overdue invoices

			// Line item management
			invoices.POST("/:id/lines", r.invoiceHandler.AddLineItem)              // Add line item
			invoices.DELETE("/:id/lines/:lineId", r.invoiceHandler.RemoveLineItem) // Remove line item
			invoices.PUT("/:id/lines/:lineId", r.invoiceHandler.UpdateLineItem)    // Update line item

			// Invoice lifecycle
			invoices.POST("/:id/send", r.invoiceHandler.Send)       // Send invoice (draft → sent)
			invoices.POST("/:id/pay", r.invoiceHandler.MarkAsPaid)  // Mark as paid
			invoices.POST("/:id/cancel", r.invoiceHandler.Cancel)   // Cancel invoice
			invoices.POST("/:id/void", r.invoiceHandler.Void)       // Void invoice

			// Tax management
			invoices.PUT("/:id/tax", r.invoiceHandler.UpdateTax) // Update tax amount
		}
	}
}
