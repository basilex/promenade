package customermgmt

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	// TODO: Uncomment when HTTP handlers are implemented
	// "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	// customerHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/http"
	// customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
)

type Router struct {
	// TODO: Uncomment when HTTP handlers are implemented
	// customerHandler *customerHTTP.CustomerHandler
}

func NewRouter(db *sqlx.DB) *Router {
	// TODO: Initialize Customer handler when HTTP layer is ready
	// customerRepository := customerRepo.NewCustomerRepository(db)
	// customerUseCase := customer.NewCustomerUseCase(customerRepository)
	// customerHandler := customerHTTP.NewCustomerHandler(customerUseCase)

	return &Router{
		// customerHandler: customerHandler,
	}
}

func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	customerMgmt := api.Group("/customer-mgmt")
	{
		// TODO: Register customer routes when HTTP handlers are ready
		_ = customerMgmt

		// customers := customerMgmt.Group("/customers")
		// {
		// 	customers.POST("", r.customerHandler.Create)
		// 	customers.POST("/b2b", r.customerHandler.CreateB2B)
		// 	customers.GET("", r.customerHandler.List)
		// 	customers.GET("/:id", r.customerHandler.GetByID)
		// 	customers.GET("/by-email", r.customerHandler.GetByEmail)
		// 	customers.PUT("/:id", r.customerHandler.Update)
		// 	customers.DELETE("/:id", r.customerHandler.Delete)
		// 	customers.POST("/:id/qualify", r.customerHandler.QualifyAsProspect)
		// 	customers.POST("/:id/convert", r.customerHandler.ConvertToCustomer)
		// 	customers.POST("/:id/churn", r.customerHandler.Churn)
		// 	customers.POST("/:id/reactivate", r.customerHandler.Reactivate)
		// 	customers.POST("/:id/upgrade/:tier", r.customerHandler.UpgradeTier)
		// 	customers.POST("/:id/downgrade/:tier", r.customerHandler.DowngradeTier)
		// 	customers.POST("/:id/tags", r.customerHandler.AddTag)
		// 	customers.DELETE("/:id/tags/:tag", r.customerHandler.RemoveTag)
		// 	customers.GET("/status/:status", r.customerHandler.ListByStatus)
		// 	customers.GET("/tier/:tier", r.customerHandler.ListByTier)
		// 	customers.GET("/stats", r.customerHandler.GetStats)
		// }
	}
}
