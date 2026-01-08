package postgres

import (
	"context"
	"database/sql"

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
	if tx, ok := database.GetTx(ctx); ok && tx != nil {
		return tx
	}
	return r.db
}

// Get selects a single row into dest
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	executor := r.getExecutor(ctx)
	return sqlx.GetContext(ctx, executor, dest, query, args...)
}

// Select selects multiple rows into dest
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	executor := r.getExecutor(ctx)
	return sqlx.SelectContext(ctx, executor, dest, query, args...)
}

// Exec executes a query without returning rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	executor := r.getExecutor(ctx)
	return executor.ExecContext(ctx, query, args...)
}

// NamedExec executes a named query
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	executor := r.getExecutor(ctx)
	return sqlx.NamedExecContext(ctx, executor, query, arg)
}

// Rebind transforms a query from QUESTION mark placeholders to the DB driver's bind type
func (r *BaseRepository) Rebind(query string) string {
	return r.db.Rebind(query)
}
