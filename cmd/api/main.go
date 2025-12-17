package main

import (
	"context"
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

	jwtpkg "github.com/basilex/promenade/pkg/jwt"
	validatorpkg "github.com/basilex/promenade/pkg/validator"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/router"
	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/logger"

	// Import swagger docs
	_ "github.com/basilex/promenade/docs/v1"
	_ "github.com/basilex/promenade/docs/v2"
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

// @host localhost:8081
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load config first (before logger init)
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// Initialize structured logger
	logFormat := "text"
	if cfg.Server.Environment == "production" {
		logFormat = "json"
	}

	logLevel := "info"
	if cfg.Server.Environment == "development" {
		logLevel = "debug"
	}

	logger.Init(logger.Config{
		Level:      logLevel,
		Format:     logFormat,
		AddSource:  cfg.Server.Environment == "development",
		TimeFormat: time.RFC3339,
	})

	logger.Info("Starting Promenade API",
		slog.String("environment", cfg.Server.Environment),
		slog.String("version", "1.0.0"),
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

	// Initialize JWT Manager
	jwtManager := jwtpkg.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Initialize infrastructure
	txManager := database.NewTransactionManager(db)
	_ = txManager // Reserved for future use

	// Initialize shared middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize modules (each module encapsulates its own dependencies)
	healthRouter := router.InitHealthModule()
	authRouter := router.InitAuthModule(db, jwtManager, authMiddleware)
	countryRouter := router.InitCountryModule(db)
	currencyRouter := router.InitCurrencyModule(db)
	userContactRouter := router.InitUserContactModule(db, authMiddleware)
	userProfileRouter := router.InitUserProfileModule(db, authMiddleware)
	userPostRouter := router.InitUserPostModule(db, authMiddleware)
	// Future modules:
	// rbacRouter := router.InitRBACModule(db, authMiddleware)
	// notificationRouter := router.InitNotificationModule(db, authMiddleware, messageQueue)

	// Setup HTTP server
	if cfg.Server.Environment == "production" {
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
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// API routes
	api := r.Group("/api")

	// API v1
	v1 := api.Group("/v1")
	{
		// Swagger UI for v1
		v1.GET("/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.InstanceName("v1"),
			ginSwagger.URL("/api/v1/docs/swagger/doc.json")))

		// V1 API endpoints
		v1Router := router.NewV1Router(healthRouter, authRouter, countryRouter, currencyRouter, userContactRouter, userProfileRouter, userPostRouter)
		v1Router.Setup(v1)
	}

	// API v2 (future)
	v2 := api.Group("/v2")
	{
		// Swagger UI for v2
		v2.GET("/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.InstanceName("v2"),
			ginSwagger.URL("/api/v2/docs/swagger/doc.json")))

		// V2 API endpoints will be registered here
		// v2Router := router.NewV2Router(...)
		// v2Router.Setup(v2)
	}

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
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
			slog.String("port", cfg.Server.Port),
			slog.String("host", cfg.Server.Host),
			slog.String("health_check", "http://"+host+":"+cfg.Server.Port+"/health"),
			slog.String("swagger_v1", "http://"+host+":"+cfg.Server.Port+"/api/v1/docs/swagger/index.html"),
			slog.String("environment", cfg.Server.Environment),
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
