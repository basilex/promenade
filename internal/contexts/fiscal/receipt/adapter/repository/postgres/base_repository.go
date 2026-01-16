package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/database"
)

// BaseRepository provides common database operations for all repositories in this context.
// Supports transaction management via context propagation.
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository instance.
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// getExecutor returns the appropriate executor (transaction or database) from context.
// If a transaction is active in the context, returns the transaction.
// Otherwise, returns the database connection.
func (r *BaseRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes a query and scans the result into dest.
// Returns sql.ErrNoRows if no rows found.
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes a query and scans multiple rows into dest.
// dest must be a pointer to a slice.
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning any rows.
// Returns sql.Result with affected rows count.
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query (with :name placeholders) without returning rows.
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}
