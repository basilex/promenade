package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/basilex/promenade/pkg/bus"
	_ "github.com/basilex/promenade/pkg/bus/memory" // Register memory bus adapter
	_ "github.com/basilex/promenade/pkg/bus/redis"  // Register redis bus adapter

	jwtpkg "github.com/basilex/promenade/pkg/jwt"
	validatorpkg "github.com/basilex/promenade/pkg/validator"

	"github.com/basilex/promenade/internal/adapter/http/shared/handler"
	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/router"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/internal/infrastructure/notification"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/migration"
	"github.com/basilex/promenade/pkg/module"
	"github.com/basilex/promenade/pkg/version"

	// Import swagger docs
	_ "github.com/basilex/promenade/docs/v1"
	_ "github.com/basilex/promenade/docs/v2"

	// Import modules for auto-registration
	_ "github.com/basilex/promenade/internal/modules/analytics" // Commercial module (requires license)
	_ "github.com/basilex/promenade/internal/modules/posts"
	_ "github.com/basilex/promenade/internal/modules/profiles"
	// _ "github.com/basilex/promenade/internal/modules/warehouse" // Commercial module (requires license)
)

// @title Promenade API
// @version 1.0
// @description Production-ready REST API built with Clean Architecture, featuring PostgreSQL with UUID v7, comprehensive testing infrastructure, JWT authentication, and API versioning.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/basilex/promenade
// @contact.email alexander.vasilenko@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter JWT token in format: Bearer {token}

func main() {
	// Load config from YAML (with env overrides)
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// Log configuration
	slog.Info("Configuration loaded", 
		slog.String("environment", cfg.App.Environment),
		slog.String("version", cfg.App.Version))

	// Initialize structured logger
	logFormat := "text"
	if cfg.App.Environment == "production" {
		logFormat = "json"
	}

	logLevel := "info"
	if cfg.App.Environment == "development" {
		logLevel = "debug"
	}

	logger.Init(logger.Config{
		Level:      logLevel,
		Format:     logFormat,
		AddSource:  cfg.App.Environment == "development",
		TimeFormat: time.RFC3339,
	})

	logger.Info("Starting "+version.ServiceName,
		slog.String("environment", cfg.App.Environment),
		slog.String("version", version.ServiceVersion),
	)

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", slog.Any("error", err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", slog.Any("error", err))
		}
	}()

	// Run migrations automatically on startup
	logger.Info("Running database migrations...")
	migrationManager := migration.NewManager(db, "migrations")
	migrationsCtx := context.Background()
	
	// Get enabled modules for migration
	enabledMigrationModules := cfg.Modules.Enabled
	if err := migrationManager.MigrateAll(migrationsCtx, enabledMigrationModules); err != nil {
		logger.Fatal("Failed to run migrations", slog.Any("error", err))
	}
	logger.Info("Database migrations completed successfully")

	// Initialize JWT Manager
	jwtManager := jwtpkg.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)

	// Initialize infrastructure
	txManager := database.NewTransactionManager(db)
	_ = txManager // Reserved for future use

	// Initialize Event Bus from configuration (supports memory/redis adapters)
	eventBus, err := bus.NewBus(cfg.Bus)
	if err != nil {
		logger.Fatal("Failed to initialize event bus", slog.Any("error", err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := eventBus.Close(ctx); err != nil {
			logger.Error("Failed to close event bus", slog.Any("error", err))
		}
	}()

	// Initialize Notification Service (асинхронная отправка email через event bus)
	emailSender := notification.NewMockEmailSender() // TODO: replace with real SMTP sender in production
	templatesPath := "templates/email"
	emailService, err := notification.NewEmailService(
		eventBus,
		emailSender,
		templatesPath,
		cfg.Email.FromAddress,
		cfg.Email.FromName,
		cfg.Email.AppURL,
		cfg.Email.AppName,
	)
	if err != nil {
		logger.Fatal("Failed to create email service", slog.Any("error", err))
	}
	if err := emailService.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start email service", slog.Any("error", err))
	}
	logger.Info("Email notification service started (async via event bus)",
		slog.String("templates_path", templatesPath),
		slog.String("from_address", cfg.Email.FromAddress),
		slog.String("app_url", cfg.Email.AppURL))

	// Initialize shared middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize RBAC use case and authorization middleware
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	roleUseCase := usecase.NewRoleUseCase(roleRepo, permissionRepo)
	authzMiddleware := middleware.NewAuthorizationMiddleware(roleUseCase)

	// Initialize core modules (minimal infrastructure only)
	healthRouter := router.InitHealthModule()
	authRouter := router.InitAuthModule(db, jwtManager, authMiddleware, authzMiddleware, eventBus)
	countryRouter := router.InitCountryModule(db)
	currencyRouter := router.InitCurrencyModule(db)
	languageRouter := router.InitLanguageModule(db, authMiddleware, authzMiddleware)
	timezoneRouter := router.InitTimezoneModule(db, authMiddleware, authzMiddleware)

	// RBAC module
	rbacRouter := router.InitRBACModule(db, authMiddleware, authzMiddleware)

	// Initialize purge system (includes scheduler that auto-starts)
	purgeUseCase, purgeScheduler, err := router.InitPurgeModule(cfg.Purge, eventBus)
	if err != nil {
		logger.Fatal("Failed to initialize purge module", slog.Any("error", err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := purgeScheduler.Stop(ctx); err != nil {
			logger.Error("Failed to stop purge scheduler", slog.Any("error", err))
		}
	}()

	// Admin module (includes purge operations)
	adminRouter := router.InitAdminModule(purgeUseCase, purgeScheduler, authMiddleware, authzMiddleware)

	// ============================================================================
	// MODULE SYSTEM - Dynamic module loading
	// ============================================================================

	logger.Info("Initializing module system...")

	// Create core infrastructure for modules
	moduleCore := &module.Core{
		DB:       db,
		EventBus: eventBus,
		JWT:      jwtManager,
		Config:   cfg,
		Logger:   slog.Default(),
		Registry: module.DefaultRegistry,
	}

	// Get module registry
	registry := module.DefaultRegistry

	// List all registered modules
	allModules := registry.ListModules()
	logger.Info("Registered modules",
		slog.Int("count", len(allModules)),
		slog.Any("modules", allModules))

	// Get enabled modules from configuration
	enabledModules := cfg.Modules.Enabled
	if len(enabledModules) == 0 {
		logger.Warn("No modules enabled in configuration")
	} else {
		logger.Info("Enabled modules", slog.Any("modules", enabledModules))
	}

	// Initialize all enabled modules
	ctx := context.Background()
	if err := registry.InitializeAll(ctx, moduleCore, enabledModules); err != nil {
		logger.Fatal("Failed to initialize modules", slog.Any("error", err))
	}

	// Start all enabled modules (background workers, etc.)
	if err := registry.StartAll(ctx, enabledModules); err != nil {
		logger.Fatal("Failed to start modules", slog.Any("error", err))
	}

	// Defer module shutdown
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := registry.StopAll(shutdownCtx, enabledModules); err != nil {
			logger.Error("Failed to stop modules gracefully", slog.Any("error", err))
		}
	}()

	// Setup HTTP server
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := validatorpkg.RegisterCustomValidators(v); err != nil {
			logger.Fatal("Failed to register custom validators", slog.Any("error", err))
		}
	}

	logger.Debug("Custom validators registered successfully")

	// Global middleware
	r.Use(
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.CORS(),
	)

	// Initialize shared handlers
	infoHandler := handler.NewInfoHandler(
		version.ServiceName, version.ServiceVersion, cfg.App.Environment, cfg.Server.Host, fmt.Sprintf("%d", cfg.Server.Port),
	)
	errorHandler := handler.NewErrorHandler()

	// Handle HTTP errors (404, 405)
	r.NoRoute(errorHandler.HandleNotFound)
	r.NoMethod(errorHandler.HandleMethodNotAllowed)

	// API routes
	api := r.Group("/api")
	{
		// GET /api - shows available API versions and links
		api.GET("", infoHandler.GetAPIInfo)
	}

	// API v1
	v1 := api.Group("/v1")
	{
		// GET /api/v1 - shows v1 API information
		v1.GET("", infoHandler.GetV1Info)

		// Swagger UI for v1
		v1.GET("/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.InstanceName("v1"),
			ginSwagger.URL("/api/v1/docs/swagger/doc.json")))

		// V1 API endpoints (core routes only - profiles/contacts moved to module)
		v1Router := router.NewV1Router(healthRouter, authRouter, countryRouter, currencyRouter, languageRouter, timezoneRouter, rbacRouter, adminRouter)
		v1Router.Setup(v1)

		// Register module routes dynamically
		for _, moduleName := range enabledModules {
			mod := registry.Get(moduleName)
			if mod != nil {
				logger.Info("Registering routes for module", slog.String("module", moduleName))
				mod.RegisterRoutes(v1)
			}
		}
	}

	// API v2 (future)
	v2 := api.Group("/v2")
	{
		// GET /api/v2 - shows v2 API information
		v2.GET("", infoHandler.GetV2Info)

		// Swagger UI for v2
		v2.GET("/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.InstanceName("v2"),
			ginSwagger.URL("/api/v2/docs/swagger/doc.json")))

		// V2 API endpoints will be registered here
		// v2Router := router.NewV2Router(...)
		// v2Router.Setup(v2)
	}

	// Start server
	// Setup HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		host := cfg.Server.Host
		if host == "0.0.0.0" || host == "" {
			host = "localhost"
		}

		logger.Info("Server started",
			slog.Int("port", cfg.Server.Port),
			slog.String("host", cfg.Server.Host),
			slog.String("health_check", fmt.Sprintf("http://%s:%d/api/v1/health", host, cfg.Server.Port)),
			slog.String("swagger_v1", fmt.Sprintf("http://%s:%d/api/v1/docs/swagger/index.html", host, cfg.Server.Port)),
			slog.String("environment", cfg.App.Environment),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", slog.Any("error", err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", slog.Any("error", err))
	}

	logger.Info("Server exited gracefully")
}
