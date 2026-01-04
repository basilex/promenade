package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/database"
)

// BaseRepository provides common database operations for Identity context repositories
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
func (r *BaseRepository) Get(ctx context.Context, dest any, query string, args ...any) error {
	executor := r.getExecutor(ctx)
	return sqlx.GetContext(ctx, executor, dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *BaseRepository) Select(ctx context.Context, dest any, query string, args ...any) error {
	executor := r.getExecutor(ctx)
	return sqlx.SelectContext(ctx, executor, dest, query, args...)
}

// Exec executes query and returns sql.Result
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	executor := r.getExecutor(ctx)
	return executor.ExecContext(ctx, query, args...)
}

// NamedExec executes named query and returns sql.Result
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg any) (sql.Result, error) {
	executor := r.getExecutor(ctx)
	return sqlx.NamedExecContext(ctx, executor, query, arg)
}

// parseTime parses RFC3339 timestamp string to time.Time
func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
