package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"time"

	_ "github.com/lib/pq"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/internal/infrastructure/seed/shared"
	"github.com/basilex/promenade/pkg/logger"
)

func main() {
	var (
		contextFlag = flag.String("context", "shared", "Context to seed (shared, identity, all)")
		forceFlag   = flag.Bool("force", false, "Force seed even if data exists (WARNING: may cause duplicates)")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	ctx := context.Background()

	// Get database driver from config (default: postgres)
	driver := cfg.Database.Driver
	if driver == "" {
		driver = "postgres"
	}

	// Connect to database (PostgreSQL only)
	if driver != "postgres" && driver != "postgresql" {
		logger.Fatal("Unsupported database driver",
			slog.String("driver", driver),
			slog.String("supported", "postgres"),
		)
	}

	db, err := database.NewPostgresConnection(&cfg.Database.Postgres)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", slog.Any("error", err))
	}
	logger.Info("Connected to PostgreSQL",
		slog.String("host", cfg.Database.Postgres.Host),
		slog.String("database", cfg.Database.Postgres.Database),
	)

	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database", slog.Any("error", err))
		}
	}()

	// Execute seeding based on context
	switch *contextFlag {
	case "shared":
		if err := shared.SeedAll(ctx, db); err != nil {
			logger.Fatal("Failed to seed shared context", slog.Any("error", err))
		}
	case "identity":
		logger.Info("Identity context seeding not yet implemented")
		// TODO: Add identity seed
	case "all":
		logger.Info("Seeding all contexts...")
		if err := shared.SeedAll(ctx, db); err != nil {
			logger.Fatal("Failed to seed shared context", slog.Any("error", err))
		}
		// TODO: Add other contexts
	default:
		logger.Fatal("Unknown context", slog.String("context", *contextFlag))
	}

	logger.Info("Seeding complete!")

	// Suppress unused variable warning
	_ = forceFlag
}
