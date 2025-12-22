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

func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// getExecutor returns either transaction or regular connection
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

// Exec executes query without returning rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...any) error {
	executor := r.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query, args...)
	return err
}

// NamedExec executes named query
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg any) error {
	executor := r.getExecutor(ctx)
	_, err := sqlx.NamedExecContext(ctx, executor, query, arg)
	return err
}
