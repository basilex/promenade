package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/basilex/promenade/internal/contexts/billing"
	"github.com/basilex/promenade/internal/contexts/banking"
	customermgmt "github.com/basilex/promenade/internal/contexts/customer-mgmt"
	"github.com/basilex/promenade/internal/contexts/fiscal"
	"github.com/basilex/promenade/internal/contexts/identity"
	ordermgmt "github.com/basilex/promenade/internal/contexts/order-mgmt"
	"github.com/basilex/promenade/internal/contexts/scripting"
	"github.com/basilex/promenade/internal/contexts/shared"
	"github.com/basilex/promenade/internal/contexts/ui"
	"github.com/basilex/promenade/internal/contexts/warehouse"
	"github.com/basilex/promenade/internal/infrastructure/health"
	"github.com/basilex/promenade/pkg/fiscal/checkbox"

	_ "github.com/basilex/promenade/docs/swagger" // Import generated docs
)

// Server holds HTTP server configuration
type Server struct {
	app    *App
	router *gin.Engine
}

// NewServer creates a new HTTP server
func NewServer(app *App) *Server {
	// Set Gin mode
	if app.Config.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	return &Server{
		app:    app,
		router: router,
	}
}

// SetupRoutes registers all application routes
func (s *Server) SetupRoutes() {
	// Health checks
	healthHandler := health.NewHandler(s.app.HealthChecker)
	healthHandler.RegisterRoutes(s.router)

	// Swagger UI documentation
	s.router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize context routers
	sharedRouter := shared.NewRouter(s.app.DB, s.app.CacheClient)
	identityRouter := identity.NewRouter(s.app.DB, s.app.JWTManager, s.app.TokenRevoker)
	customerMgmtRouter := customermgmt.NewRouter(s.app.DB)
	orderMgmtRouter := ordermgmt.NewRouter(s.app.DB, s.app.EventBus)
	billingRouter := billing.NewRouter(s.app.DB)
	warehouseRouter := warehouse.NewRouter(s.app.DB)
	var checkboxClient *checkbox.Client
	checkboxCfg := s.app.Config.Fiscal.Checkbox
	if checkboxCfg.APIKey != "" {
		checkboxClient = checkbox.NewClient(&checkbox.Config{
			APIKey:  checkboxCfg.APIKey,
			Sandbox: checkboxCfg.Sandbox,
			Timeout: checkboxCfg.Timeout,
		})
	}
	fiscalRouter := fiscal.NewRouter(s.app.DB, checkboxClient, s.app.Config.Fiscal.PDFOutputDir)
	scriptingRouter := scripting.NewRouter(s.app.DB)
	uiRouter := ui.NewRouter(s.app.DB)

	// API routes
	api := s.router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Promenade CRM Platform API v1",
					"version": s.app.Config.App.Version,
				})
			})

			// Register context routes
			sharedRouter.RegisterRoutes(v1)       // Countries, Currencies, Languages, Timezones
			identityRouter.RegisterRoutes(v1)     // Users, Contacts, Profiles, Roles, Permissions
			customerMgmtRouter.RegisterRoutes(v1) // Customers
			orderMgmtRouter.RegisterRoutes(v1)    // Orders
			billingRouter.RegisterRoutes(v1)      // Invoices, Payments, Subscriptions
			fiscalRouter.RegisterRoutes(v1)       // Cash Registers (ПРРО)
			warehouseRouter.RegisterRoutes(v1)    // Inventory
			scriptingRouter.RegisterRoutes(v1)    // LUA Scripts
			uiRouter.RegisterRoutes(v1)           // UI Metadata Forms

			// Banking routes
			banking.RegisterRoutes(v1, s.app.BankAccountUseCase, s.app.BankTransactionUseCase)

			// Analytics routes
			if s.app.SalesReportHandler != nil {
				s.app.SalesReportHandler.RegisterRoutes(v1)
			}
		}
	}
}

// Start starts the HTTP server
func (s *Server) Start() *http.Server {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.app.Config.Server.Port),
		Handler:      s.router,
		ReadTimeout:  s.app.Config.Server.ReadTimeout,
		WriteTimeout: s.app.Config.Server.WriteTimeout,
	}

	return srv
}
