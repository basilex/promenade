package fiscal

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	cashregisterHTTP "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/http"
	cashregisterRepo "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	receiptHTTP "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/http"
	receiptPrinter "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/printer"
	receiptRepo "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/fiscal/checkbox"
)

// Router handles all Fiscal context routes
type Router struct {
	cashRegisterHandler *cashregisterHTTP.CashRegisterHandler
	receiptHandler      *receiptHTTP.ReceiptHandler
}

// NewRouter creates a new Fiscal context router
func NewRouter(db *sqlx.DB, checkboxClient *checkbox.Client, pdfOutputDir string) *Router {
	// Initialize CashRegister aggregate
	cashRegisterRepository := cashregisterRepo.NewCashRegisterRepository(db)
	cashRegisterUseCase := cashregister.NewUseCase(cashRegisterRepository)
	cashRegisterHandler := cashregisterHTTP.NewCashRegisterHandler(cashRegisterUseCase)

	// Initialize Receipt aggregate
	receiptRepository := receiptRepo.NewReceiptRepository(db)
	var printer receipt.IPrinter
	var pdfPrinter receipt.IPrinter
	if pdfOutputDir != "" {
		pdfPrinter = receiptPrinter.NewPDFPrinter(pdfOutputDir)
	}
	if checkboxClient != nil {
		checkboxPrinter := receiptPrinter.NewCheckboxPrinter(checkboxClient)
		if pdfPrinter != nil {
			printer = receipt.NewMultiPrinter(checkboxPrinter, pdfPrinter)
		} else {
			printer = checkboxPrinter
		}
	} else if pdfPrinter != nil {
		printer = pdfPrinter
	}
	receiptUseCase := receipt.NewUseCase(receiptRepository, printer)
	receiptHandler := receiptHTTP.NewReceiptHandler(receiptUseCase)

	return &Router{
		cashRegisterHandler: cashRegisterHandler,
		receiptHandler:      receiptHandler,
	}
}

// RegisterRoutes registers all Fiscal context routes
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	fiscal := api.Group("/fiscal")
	{
		// Cash Register routes
		cashRegisters := fiscal.Group("/cash-registers")
		{
			// CRUD operations
			cashRegisters.POST("", r.cashRegisterHandler.Create)       // Create cash register
			cashRegisters.GET("/:id", r.cashRegisterHandler.GetByID)   // Get by ID
			cashRegisters.GET("", r.cashRegisterHandler.List)          // List cash registers
			cashRegisters.DELETE("/:id", r.cashRegisterHandler.Delete) // Soft delete

			// Operations
			cashRegisters.POST("/:id/activate", r.cashRegisterHandler.Activate)     // Activate
			cashRegisters.POST("/:id/deactivate", r.cashRegisterHandler.Deactivate) // Deactivate
		}

		receipts := fiscal.Group("/receipts")
		{
			receipts.POST("", r.receiptHandler.Create)                // Create receipt
			receipts.GET("/:id", r.receiptHandler.GetByID)            // Get by ID
			receipts.GET("", r.receiptHandler.List)                   // List receipts
			receipts.POST("/:id/print", r.receiptHandler.MarkPrinted) // Mark printed
			receipts.POST("/:id/cancel", r.receiptHandler.Cancel)     // Cancel receipt
			receipts.DELETE("/:id", r.receiptHandler.Delete)          // Soft delete
		}
	}
}
