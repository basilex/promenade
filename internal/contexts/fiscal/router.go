package fiscal

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	cashregisterHTTP "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/http"
	cashregisterRepo "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/repository/postgres"
)

// Router handles all Fiscal context routes
type Router struct {
	cashRegisterHandler *cashregisterHTTP.CashRegisterHandler
}

// NewRouter creates a new Fiscal context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize CashRegister aggregate
	cashRegisterRepository := cashregisterRepo.NewCashRegisterRepository(db)
	cashRegisterUseCase := cashregister.NewUseCase(cashRegisterRepository)
	cashRegisterHandler := cashregisterHTTP.NewCashRegisterHandler(cashRegisterUseCase)

	return &Router{
		cashRegisterHandler: cashRegisterHandler,
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
			cashRegisters.POST("", r.cashRegisterHandler.Create)           // Create cash register
			cashRegisters.GET("/:id", r.cashRegisterHandler.GetByID)       // Get by ID
			cashRegisters.GET("", r.cashRegisterHandler.List)              // List cash registers
			cashRegisters.DELETE("/:id", r.cashRegisterHandler.Delete)     // Soft delete

			// Operations
			cashRegisters.POST("/:id/activate", r.cashRegisterHandler.Activate)       // Activate
			cashRegisters.POST("/:id/deactivate", r.cashRegisterHandler.Deactivate)   // Deactivate
		}
	}
}