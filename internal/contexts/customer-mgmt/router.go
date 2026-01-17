package customermgmt

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	companyHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/adapter/http"
	companyRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/adapter/repository/postgres"
	companyUseCase "github.com/basilex/promenade/internal/contexts/customer-mgmt/company/usecase"
	customerHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/http"
	customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	customerUseCase "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/usecase"
	dealHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/adapter/http"
	dealRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/adapter/repository/postgres"
	dealUseCase "github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/usecase"
	interactionHTTP "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/adapter/http"
	interactionRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/adapter/repository/postgres"
	interactionUseCase "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/usecase"
)

// Router manages routes for Customer Management context
type Router struct {
	customerHandler    *customerHTTP.CustomerHandler
	companyHandler     *companyHTTP.CompanyHandler
	dealHandler        *dealHTTP.DealHandler
	interactionHandler *interactionHTTP.InteractionHandler
	// Note: Analytics routes registered separately in bootstrap.go
}

// NewRouter creates a new Customer Management router with all dependencies
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Customer aggregate
	customerRepository := customerRepo.NewCustomerRepository(db)
	customerUseCaseImpl := customerUseCase.NewCustomerUseCase(customerRepository)
	customerHandler := customerHTTP.NewCustomerHandler(customerUseCaseImpl)

	// Initialize Company aggregate
	companyRepository := companyRepo.NewCompanyRepository(db)
	companyUseCaseImpl := companyUseCase.NewCompanyUseCase(companyRepository)
	companyHandler := companyHTTP.NewCompanyHandler(companyUseCaseImpl)

	// Initialize Deal aggregate
	dealRepository := dealRepo.NewDealRepository(db)
	dealUseCaseImpl := dealUseCase.NewDealUseCase(dealRepository)
	dealHandler := dealHTTP.NewDealHandler(dealUseCaseImpl)

	// Initialize Interaction aggregate
	interactionRepository := interactionRepo.NewInteractionRepository(db)
	interactionUseCaseImpl := interactionUseCase.NewInteractionUseCase(interactionRepository)
	interactionHandler := interactionHTTP.NewInteractionHandler(interactionUseCaseImpl)

	// Note: Analytics initialized separately in bootstrap.go with event handlers

	return &Router{
		customerHandler:    customerHandler,
		companyHandler:     companyHandler,
		dealHandler:        dealHandler,
		interactionHandler: interactionHandler,
	}
}

func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	customerMgmt := api.Group("/customer-mgmt")
	{
		// Customer routes
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

		// Company routes (B2B customers)
		companies := customerMgmt.Group("/companies")
		{
			companies.POST("", r.companyHandler.Create)
			companies.GET("", r.companyHandler.List)
			companies.GET("/:id", r.companyHandler.GetByID)
			companies.GET("/name/:name", r.companyHandler.GetByName)
			companies.GET("/tax/:taxId", r.companyHandler.GetByTaxID)
			companies.GET("/industry/:industry", r.companyHandler.ListByIndustry)
			companies.GET("/size/:size", r.companyHandler.ListBySize)
			companies.GET("/:id/subsidiaries", r.companyHandler.ListSubsidiaries)
			companies.PUT("/:id/basic-info", r.companyHandler.UpdateBasicInfo)
			companies.PUT("/:id/contact-info", r.companyHandler.UpdateContactInfo)
			companies.PUT("/:id/business-info", r.companyHandler.UpdateBusinessInfo)
			companies.PUT("/:id/parent", r.companyHandler.SetParentCompany)
			companies.PUT("/:id/description", r.companyHandler.UpdateDescription)
			companies.DELETE("/:id", r.companyHandler.Delete)
		}

		// Deal routes (Sales Pipeline)
		deals := customerMgmt.Group("/deals")
		{
			deals.POST("", r.dealHandler.Create)
			deals.GET("", r.dealHandler.List)
			deals.GET("/:id", r.dealHandler.GetByID)
			deals.PUT("/:id/basic-info", r.dealHandler.UpdateBasicInfo)
			deals.PUT("/:id/value", r.dealHandler.UpdateValue)
			deals.PUT("/:id/stage", r.dealHandler.MoveToStage)
			deals.POST("/:id/win", r.dealHandler.MarkAsWon)
			deals.POST("/:id/lose", r.dealHandler.MarkAsLost)
			deals.DELETE("/:id", r.dealHandler.Delete)
			deals.GET("/stage/:stage", r.dealHandler.ListByStage)
			deals.GET("/stats/pipeline", r.dealHandler.GetPipelineStats)
			deals.GET("/stats/won", r.dealHandler.GetWonDeals)
		}

		// Interaction routes (Customer Communication Log)
		interactions := customerMgmt.Group("/interactions")
		{
			interactions.POST("", r.interactionHandler.Create)
			interactions.GET("", r.interactionHandler.List)
			interactions.GET("/:id", r.interactionHandler.GetByID)
			interactions.GET("/customer/:customer_id", r.interactionHandler.ListByCustomer)
			interactions.GET("/company/:company_id", r.interactionHandler.ListByCompany)
			interactions.GET("/type/:type", r.interactionHandler.ListByType)
			interactions.GET("/follow-ups/pending", r.interactionHandler.ListPendingFollowUps)
			interactions.PUT("/:id/content", r.interactionHandler.UpdateContent)
			interactions.PUT("/:id/outcome", r.interactionHandler.SetOutcome)
			interactions.POST("/:id/end", r.interactionHandler.EndInteraction)
			interactions.PUT("/:id/follow-up", r.interactionHandler.SetFollowUp)
			interactions.POST("/:id/attendees", r.interactionHandler.AddAttendee)
			interactions.DELETE("/:id/attendees/:attendee_id", r.interactionHandler.RemoveAttendee)
			interactions.DELETE("/:id", r.interactionHandler.Delete)
		}

		// Note: Analytics routes registered separately in bootstrap.go
	}
}
