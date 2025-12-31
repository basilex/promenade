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

	"github.com/basilex/promenade/pkg/logger"
)

// GracefulShutdown handles graceful shutdown of the server
func GracefulShutdown(srv *http.Server, app *App) {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server gracefully...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", slog.Any("error", err))
	}

	// Close application dependencies
	app.Close()

	logger.Info("Server exited gracefully")
}

// LogServerInfo logs server startup information
func LogServerInfo(port int, host, environment string) {
	if host == "0.0.0.0" || host == "" {
		host = "localhost"
	}

	logger.Info("Server started",
		slog.Int("port", port),
		slog.String("host", host),
		slog.String("health_check", fmt.Sprintf("http://%s:%d/health", host, port)),
		slog.String("environment", environment),
	)
}
