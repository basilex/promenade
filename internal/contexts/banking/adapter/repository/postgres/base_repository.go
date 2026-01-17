package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/database"
)

// BaseRepository provides common database operations for banking context repositories
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// getExecutor returns either transaction or regular connection from context
func (r *BaseRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes query and scans single row
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes query without returning rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := r.getExecutor(ctx).ExecContext(ctx, query, args...)
	return err
}

// NamedExec executes a named query (with struct bindings)
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) error {
	_, err := sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
	return err
}
