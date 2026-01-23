package main

import (
	"log/slog"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/logger"

	_ "github.com/basilex/promenade/pkg/bus/memory" // Register memory adapter
	_ "github.com/basilex/promenade/pkg/bus/redis"  // Register redis adapter
)

// @title Promenade Platform
// @version 0.2.0-dev
// @description Modern backend platform for customer management, orders, and business workflows with clean DDD architecture
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

	// Bootstrap application dependencies
	app, err := Bootstrap(cfg)
	if err != nil {
		logger.Fatal("Failed to bootstrap application", slog.Any("error", err))
	}
	defer app.Close()

	// Setup HTTP server
	server := NewServer(app)
	server.SetupRoutes()
	srv := server.Start()

	// Start server in goroutine
	go func() {
		LogServerInfo(cfg.Server.Port, cfg.Server.Host, cfg.App.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", slog.Any("error", err))
		}
	}()

	// Graceful shutdown
	GracefulShutdown(srv, app)
}
