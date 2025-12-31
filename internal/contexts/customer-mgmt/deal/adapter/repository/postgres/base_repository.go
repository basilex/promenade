package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/database"
)

// BaseRepository provides common database operations for all repositories
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// getExecutor returns the appropriate executor (transaction or database)
func (br *BaseRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok && tx != nil {
		return tx
	}
	return br.db
}

// Get executes a query that returns a single row
func (br *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	executor := br.getExecutor(ctx)
	return sqlx.GetContext(ctx, executor, dest, query, args...)
}

// Select executes a query that returns multiple rows
func (br *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	executor := br.getExecutor(ctx)
	return sqlx.SelectContext(ctx, executor, dest, query, args...)
}

// Exec executes a query that doesn't return rows
func (br *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	executor := br.getExecutor(ctx)
	return executor.ExecContext(ctx, query, args...)
}

// NamedExec executes a named query
func (br *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	executor := br.getExecutor(ctx)
	return sqlx.NamedExecContext(ctx, executor, query, arg)
}
