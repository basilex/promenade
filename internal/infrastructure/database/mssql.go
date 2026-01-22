package database

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/microsoft/go-mssqldb"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/logger"
)

// NewMSSQLConnection creates a new MS SQL Server database connection
func NewMSSQLConnection(cfg *config.MSSQLSection) (*sqlx.DB, error) {
	// Build connection string for MS SQL Server
	// Format: sqlserver://username:password@host:port?database=dbname&param=value
	query := url.Values{}
	query.Add("database", cfg.Database)

	if cfg.Encrypt != "" {
		query.Add("encrypt", cfg.Encrypt)
	}

	if cfg.TrustServerCert {
		query.Add("TrustServerCertificate", "true")
	}

	dsn := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?%s",
		url.QueryEscape(cfg.User),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		query.Encode(),
	)

	logger.Info("Connecting to MS SQL Server",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.Database),
		slog.String("user", cfg.User),
		slog.String("encrypt", cfg.Encrypt),
	)

	db, err := sqlx.Connect("sqlserver", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MS SQL Server: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MS SQL Server: %w", err)
	}

	logger.Info("MS SQL Server connection established",
		slog.Int("max_open_conns", cfg.MaxOpenConns),
		slog.Int("max_idle_conns", cfg.MaxIdleConns),
	)

	return db, nil
}
