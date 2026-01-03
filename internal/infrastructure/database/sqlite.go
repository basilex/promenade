package database

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/logger"
)

// NewSQLiteConnection creates a new SQLite database connection
func NewSQLiteConnection(cfg *config.SQLiteSection) (*sqlx.DB, error) {
	// Create data directory if it doesn't exist
	if cfg.Path != ":memory:" {
		dir := filepath.Dir(cfg.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// Build DSN with SQLite-specific parameters
	dsn := cfg.Path
	if cfg.Mode != "" || cfg.Cache != "" {
		dsn += "?"
		if cfg.Mode != "" {
			dsn += "mode=" + cfg.Mode
		}
		if cfg.Cache != "" {
			if cfg.Mode != "" {
				dsn += "&"
			}
			dsn += "cache=" + cfg.Cache
		}
	}

	logger.Info("Connecting to SQLite database",
		slog.String("path", cfg.Path),
		slog.String("mode", cfg.Mode),
		slog.String("cache", cfg.Cache),
	)

	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// SQLite-specific settings
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	} else {
		db.SetMaxOpenConns(1) // SQLite best practice: single writer
	}

	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	} else {
		db.SetMaxIdleConns(1)
	}

	// Enable foreign keys (disabled by default in SQLite)
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		logger.Warn("Failed to enable WAL mode", slog.Any("error", err))
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	logger.Info("SQLite database connection established",
		slog.String("path", cfg.Path),
		slog.Int("max_open_conns", cfg.MaxOpenConns),
	)

	return db, nil
}
