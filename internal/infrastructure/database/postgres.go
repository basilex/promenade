package database

import (
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/logger"
)

func NewPostgresConnection(cfg *config.PostgresSection) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	logger.Info("Connecting to database",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.Database),
		slog.String("user", cfg.User),
		slog.String("sslmode", cfg.SSLMode),
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		slog.Int("max_open_conns", cfg.MaxOpenConns),
		slog.Int("max_idle_conns", cfg.MaxIdleConns),
	)

	return db, nil
}
