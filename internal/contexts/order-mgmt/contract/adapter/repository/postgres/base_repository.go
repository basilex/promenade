package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/database"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// getExecutor returns either a transaction or the database connection
func (r *BaseRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes a query that returns a single row
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	executor := r.getExecutor(ctx)
	return sqlx.GetContext(ctx, executor, dest, query, args...)
}

// Select executes a query that returns multiple rows
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	executor := r.getExecutor(ctx)
	return sqlx.SelectContext(ctx, executor, dest, query, args...)
}

// Exec executes a query without returning rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) error {
	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query, args...)
	return err
}

// NamedExec executes a named query
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) error {
	_, err := sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
	return err
}
