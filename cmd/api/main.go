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
	_ "github.com/lib/pq"

	customermgmt "github.com/basilex/promenade/internal/contexts/customer-mgmt"
	"github.com/basilex/promenade/internal/contexts/identity"
	"github.com/basilex/promenade/internal/contexts/shared"
	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/bus"
	_ "github.com/basilex/promenade/pkg/bus/memory" // Register memory adapter
	_ "github.com/basilex/promenade/pkg/bus/redis"  // Register redis adapter
	"github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/migration"
)

// @title Promenade CRM Platform
// @version 2.0
// @description Clean DDD architecture with Bounded Contexts for CRM platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/basilex/promenade
// @contact.email alexander.vasilenko@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1
// @schemes http https

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		slog.Error("Invalid configuration", slog.Any("error", err))
		os.Exit(1)
	}

	// Initialize logger
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

	logger.Info("Starting Promenade CRM Platform",
		slog.String("environment", cfg.App.Environment),
		slog.String("version", cfg.App.Version),
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

	// Run migrations automatically
	logger.Info("Running database migrations...")
	migrationManager := migration.NewManager(db, "migrations")
	migrationsCtx := context.Background()

	// Run core migrations (extensions: uuid_v7, pgcrypto)
	if err := migrationManager.MigrateNamespace(migrationsCtx, "core"); err != nil {
		logger.Fatal("Failed to run core migrations", slog.Any("error", err))
	}

	// Run shared kernel migrations (reference data: countries, currencies, languages, timezones)
	if err := migrationManager.MigrateNamespace(migrationsCtx, "shared"); err != nil {
		logger.Fatal("Failed to run shared migrations", slog.Any("error", err))
	}

	// Run identity context migrations (users, authentication, authorization, contacts)
	if err := migrationManager.MigrateNamespace(migrationsCtx, "identity"); err != nil {
		logger.Fatal("Failed to run identity migrations", slog.Any("error", err))
	}

	// Run customer management context migrations (customers, companies, deals, interactions)
	if err := migrationManager.MigrateNamespace(migrationsCtx, "customer-mgmt"); err != nil {
		logger.Fatal("Failed to run customer-mgmt migrations", slog.Any("error", err))
	}

	logger.Info("Database migrations completed successfully")

	// Initialize JWT Manager
	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            cfg.JWT.Secret,
		AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
		RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
		Issuer:               cfg.JWT.Issuer,
	})
	logger.Info("JWT Manager initialized",
		slog.String("issuer", cfg.JWT.Issuer),
		slog.Duration("access_token_duration", cfg.JWT.AccessTokenDuration),
		slog.Duration("refresh_token_duration", cfg.JWT.RefreshTokenDuration),
	)

	// Initialize Event Bus
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

	// Setup HTTP server
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(
		gin.Recovery(),
		gin.Logger(),
	)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Initialize context routers
	sharedRouter := shared.NewRouter(db)                 // Shared Context (Reference Data)
	identityRouter := identity.NewRouter(db, jwtManager) // Identity Context (User, Contact) with JWT
	customerMgmtRouter := customermgmt.NewRouter(db)     // Customer Management Context (Customer)

	// API routes
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Promenade CRM Platform API v1",
					"version": cfg.App.Version,
				})
			})

			// Register Shared Context routes (Countries, Currencies, Languages, Timezones)
			sharedRouter.RegisterRoutes(v1)

			// Register Identity Context routes (User, Contact aggregates)
			identityRouter.RegisterRoutes(v1)

			// Register Customer Management Context routes (Customer aggregate)
			customerMgmtRouter.RegisterRoutes(v1)

			// TODO: Register additional context routers here:
			// - Order Management context (Order, OrderItem, Fulfillment)
		}
	}

	// Start server
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
			slog.String("health_check", fmt.Sprintf("http://%s:%d/health", host, cfg.Server.Port)),
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
