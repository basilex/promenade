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

	jwtpkg "github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/logger"
	validatorpkg "github.com/basilex/promenade/pkg/validator"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/router"
	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
)

func main() {
	// Load config first (before logger init)
	cfg, err := config.Load()
	if err != nil {
		// Use basic logger since structured logger not yet initialized
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
	defer db.Close()

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
	authRouter := router.InitAuthModule(db, jwtManager, authMiddleware)
	countryRouter := router.InitCountryModule(db)
	currencyRouter := router.InitCurrencyModule(db)
	// Future modules:
	// rbacRouter := router.InitRBACModule(db, authMiddleware)
	// profileRouter := router.InitProfileModule(db, authMiddleware, cache)
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

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "promenade",
			"time":    time.Now().Unix(),
		})
	})

	// API routes
	api := r.Group("/api")

	// V1 Router (aggregates all module routers)
	v1Router := router.NewV1Router(authRouter, countryRouter, currencyRouter)
	v1Router.Setup(api)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info("Server started",
			slog.String("port", cfg.Server.Port),
			slog.String("health_check", "http://localhost:"+cfg.Server.Port+"/health"),
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
