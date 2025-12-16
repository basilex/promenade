package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	jwtpkg "github.com/basilex/promenade/pkg/jwt"
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/router"
	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
)

func main() {
	log.Println("Starting Promenade API...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

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
	// Future modules:
	// rbacRouter := router.InitRBACModule(db, authMiddleware)
	// profileRouter := router.InitProfileModule(db, authMiddleware, cache)
	// notificationRouter := router.InitNotificationModule(db, authMiddleware, messageQueue)

	// Setup HTTP server
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

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
	v1Router := router.NewV1Router(authRouter)
	v1Router.Setup(api)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Server started on port %s", cfg.Server.Port)
		log.Printf("Server health check: http://localhost:%s/health", cfg.Server.Port)
		log.Println("")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
