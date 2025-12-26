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

// Exec executes a query that doesn't return rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	executor := r.getExecutor(ctx)
	return executor.ExecContext(ctx, query, args...)
}

// NamedExec executes a named query that doesn't return rows
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	executor := r.getExecutor(ctx)
	
	// Type assert to *sqlx.Tx or *sqlx.DB for NamedExecContext
	if tx, ok := executor.(*sqlx.Tx); ok {
		return tx.NamedExecContext(ctx, query, arg)
	}
	
	return r.db.NamedExecContext(ctx, query, arg)
}

// WithTransaction executes a function within a transaction
func (r *BaseRepository) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	// Store transaction in context
	ctx = context.WithValue(ctx, "tx", tx)

	// Execute function
	if err := fn(ctx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	// Commit transaction
	return tx.Commit()
}
