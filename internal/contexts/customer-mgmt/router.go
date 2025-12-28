package customermgmt

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	customerHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/http"
	customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
)

// Router manages routes for Customer Management context
type Router struct {
	customerHandler *customerHTTP.CustomerHandler
}

// NewRouter creates a new Customer Management router with all dependencies
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Customer aggregate
	customerRepository := customerRepo.NewCustomerRepository(db)
	customerUseCase := customer.NewUseCase(customerRepository)
	customerHandler := customerHTTP.NewCustomerHandler(customerUseCase)

	return &Router{
		customerHandler: customerHandler,
	}
}

func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	customerMgmt := api.Group("/customer-mgmt")
	{
		customers := customerMgmt.Group("/customers")
		{
			customers.POST("", r.customerHandler.Create)
			customers.POST("/b2b", r.customerHandler.CreateB2B)
			customers.GET("", r.customerHandler.List)
			customers.GET("/:id", r.customerHandler.GetByID)
			customers.GET("/by-email", r.customerHandler.GetByEmail)
			customers.PUT("/:id", r.customerHandler.Update)
			customers.DELETE("/:id", r.customerHandler.Delete)
			customers.POST("/:id/qualify", r.customerHandler.QualifyAsProspect)
			customers.POST("/:id/convert", r.customerHandler.ConvertToCustomer)
			customers.POST("/:id/churn", r.customerHandler.Churn)
			customers.POST("/:id/reactivate", r.customerHandler.Reactivate)
			customers.POST("/:id/upgrade/:tier", r.customerHandler.UpgradeTier)
			customers.POST("/:id/downgrade/:tier", r.customerHandler.DowngradeTier)
			customers.POST("/:id/link-user", r.customerHandler.LinkToUser)
			customers.POST("/:id/assign", r.customerHandler.AssignTo)
			customers.POST("/:id/tags", r.customerHandler.AddTag)
			customers.DELETE("/:id/tags/:tag", r.customerHandler.RemoveTag)
			customers.GET("/status/:status", r.customerHandler.ListByStatus)
			customers.GET("/tier/:tier", r.customerHandler.ListByTier)
			customers.GET("/stats", r.customerHandler.GetStats)
		}
	}
}
