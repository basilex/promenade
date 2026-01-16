package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/database"
)

// BaseRepository provides common database operations for UI context repositories.
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository instance.
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// Get executes a query and scans the result into dest.
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes a query and scans multiple rows into dest.
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning any rows.
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// getExecutor returns the appropriate executor (transaction or database) from context.
func (r *BaseRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}