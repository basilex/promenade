package billing

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/billing/invoice"
	invoiceHTTP "github.com/basilex/promenade/internal/contexts/billing/invoice/adapter/http"
	invoiceRepo "github.com/basilex/promenade/internal/contexts/billing/invoice/adapter/repository/postgres"

	"github.com/basilex/promenade/internal/contexts/billing/payment"
	paymentHTTP "github.com/basilex/promenade/internal/contexts/billing/payment/adapter/http"
	paymentRepo "github.com/basilex/promenade/internal/contexts/billing/payment/adapter/repository/postgres"
)

// Router handles all Billing context routes
type Router struct {
	invoiceHandler *invoiceHTTP.InvoiceHandler
	paymentHandler *paymentHTTP.PaymentHandler
}

// NewRouter creates a new Billing context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Invoice aggregate
	invoiceRepository := invoiceRepo.NewInvoiceRepository(db)
	invoiceUseCase := invoice.NewUseCase(invoiceRepository)
	invoiceHandler := invoiceHTTP.NewInvoiceHandler(invoiceUseCase)

	// Initialize Payment aggregate
	paymentRepository := paymentRepo.NewPaymentRepository(db)
	paymentUseCase := payment.NewUseCase(paymentRepository)
	paymentHandler := paymentHTTP.NewPaymentHandler(paymentUseCase)

	return &Router{
		invoiceHandler: invoiceHandler,
		paymentHandler: paymentHandler,
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

		// Payment routes
		payments := billing.Group("/payments")
		{
			// Payment CRUD
			payments.POST("", r.paymentHandler.Create)                           // Create payment
			payments.GET("/:id", r.paymentHandler.GetByID)                       // Get payment by ID
			payments.GET("/number/:number", r.paymentHandler.GetByNumber)        // Get by payment number
			payments.GET("/transaction/:txId", r.paymentHandler.GetByTransactionID) // Get by transaction ID
			payments.DELETE("/:id", r.paymentHandler.Delete)                     // Delete payment (pending/cancelled only)
			payments.GET("", r.paymentHandler.List)                              // List payments (with filters)

			// Payment filtering
			payments.GET("/customer/:customerId", r.paymentHandler.ListByCustomer) // Payments by customer
			payments.GET("/invoice/:invoiceId", r.paymentHandler.ListByInvoice)    // Payments by invoice
			payments.GET("/status/:status", r.paymentHandler.ListByStatus)         // Payments by status

			// Payment operations
			payments.POST("/:id/link-invoice", r.paymentHandler.LinkToInvoice)    // Link payment to invoice
			payments.POST("/:id/process", r.paymentHandler.ProcessPayment)         // Start processing
			payments.POST("/:id/complete", r.paymentHandler.CompletePayment)       // Mark as completed
			payments.POST("/:id/fail", r.paymentHandler.FailPayment)               // Mark as failed
			payments.POST("/:id/refund", r.paymentHandler.RefundPayment)           // Process refund
			payments.POST("/:id/cancel", r.paymentHandler.CancelPayment)           // Cancel payment

			// Payment details
			payments.PUT("/:id/card-details", r.paymentHandler.SetCardDetails)     // Set card details
			payments.PUT("/:id/provider", r.paymentHandler.SetProvider)            // Set payment provider
			payments.POST("/:id/notes", r.paymentHandler.AddNote)                  // Add note

			// Payment totals/aggregates
			payments.GET("/totals/customer/:customerId", r.paymentHandler.GetTotalByCustomer) // Total by customer
			payments.GET("/totals/invoice/:invoiceId", r.paymentHandler.GetTotalByInvoice)    // Total by invoice
		}
	}
}
