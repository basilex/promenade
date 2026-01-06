package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/database"
)

// BaseRepository provides common database operations for all repositories.
//
// All repositories in warehouse context should embed this base to get:
// - Transaction support via context
// - Standard CRUD helpers (Get, Select, Exec, NamedExec)
// - Automatic executor selection (tx or db)
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository.
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// getExecutor returns either transaction or database connection from context.
//
// Usage:
//
//	executor := r.getExecutor(ctx)
//	err := sqlx.GetContext(ctx, executor, &result, query, args...)
func (r *BaseRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get retrieves a single row into dest.
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select retrieves multiple rows into dest.
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning rows.
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := r.getExecutor(ctx).ExecContext(ctx, query, args...)
	return err
}

// NamedExec executes a named query (with struct bindings).
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) error {
	_, err := sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
	return err
}