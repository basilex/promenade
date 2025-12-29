package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	_ "github.com/lib/pq"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/migration"
)

func main() {
	// Define flags
	var (
		command   = flag.String("cmd", "status", "Command to run: up, down, status, version")
		namespace = flag.String("namespace", "", "Migration namespace (core, posts, profiles, etc.)")
		all       = flag.Bool("all", false, "Apply to all namespaces (core + enabled modules)")
		steps     = flag.Int("steps", 1, "Number of migrations to rollback (for down command)")
	)

	flag.Parse()

	// Initialize context
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init(logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	// Initialize database connection
	db, err := database.NewPostgresConnection(&cfg.Database.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database connection", slog.Any("error", err))
		}
	}()

	// Create migration manager
	migrationsDir := "migrations"
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		migrationsDir = dir
	}

	mgr := migration.NewManager(db, migrationsDir)

	// Execute command
	switch *command {
	case "up", "migrate":
		if err := runMigrateUp(ctx, mgr, *namespace, *all, cfg); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}

	case "down", "rollback":
		if *namespace == "" {
			log.Fatal("Namespace is required for rollback")
		}
		if err := runRollback(ctx, mgr, *namespace, *steps); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}

	case "status":
		if err := runStatus(ctx, mgr); err != nil {
			log.Fatalf("Status check failed: %v", err)
		}

	case "version":
		if *namespace == "" {
			log.Fatal("Namespace is required for version check")
		}
		if err := runVersion(ctx, mgr, *namespace); err != nil {
			log.Fatalf("Version check failed: %v", err)
		}

	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func runMigrateUp(ctx context.Context, mgr migration.Manager, namespace string, all bool, cfg *config.AppConfig) error {
	if all {
		// Get enabled modules from config
		enabledModules := getEnabledModules(cfg)
		fmt.Printf("Migrating all namespaces (core + %d modules)...\n", len(enabledModules))
		return mgr.MigrateAll(ctx, enabledModules)
	}

	if namespace == "" {
		return fmt.Errorf("namespace is required (use --namespace or --all)")
	}

	fmt.Printf("Migrating namespace: %s\n", namespace)
	return mgr.MigrateNamespace(ctx, namespace)
}

func runRollback(ctx context.Context, mgr migration.Manager, namespace string, steps int) error {
	fmt.Printf("Rolling back %d migration(s) for namespace: %s\n", steps, namespace)
	return mgr.Rollback(ctx, namespace, steps)
}

func runStatus(ctx context.Context, mgr migration.Manager) error {
	status, err := mgr.Status(ctx)
	if err != nil {
		return err
	}

	if len(status) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("==================")
	for namespace, s := range status {
		fmt.Printf("Namespace: %s\n", namespace)
		fmt.Printf("  Current Version: %d\n", s.CurrentVersion)
		fmt.Printf("  Pending: %d\n", s.PendingCount)
		if s.Dirty {
			fmt.Printf("  Status: DIRTY (manual intervention required)\n")
		} else {
			fmt.Printf("  Status: OK\n")
		}
		fmt.Println()
	}

	return nil
}

func runVersion(ctx context.Context, mgr migration.Manager, namespace string) error {
	version, err := mgr.Version(ctx, namespace)
	if err != nil {
		return err
	}

	fmt.Printf("Namespace %s is at version: %d\n", namespace, version)
	return nil
}

func getEnabledModules(cfg *config.AppConfig) []string {
	return cfg.Modules.Enabled
}
