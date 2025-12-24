package postgres

import (
	"context"
	"database/sql"

	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/jmoiron/sqlx"
)

// BaseRepository provides common database operations for audit repositories
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository instance
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// Get executes a query that returns a single row
func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	tx, ok := database.GetTx(ctx)
	if ok {
		return tx.GetContext(ctx, dest, query, args...)
	}
	return r.db.GetContext(ctx, dest, query, args...)
}

// Select executes a query that returns multiple rows
func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	tx, ok := database.GetTx(ctx)
	if ok {
		return tx.SelectContext(ctx, dest, query, args...)
	}
	return r.db.SelectContext(ctx, dest, query, args...)
}

// Exec executes a query without returning any rows
func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tx, ok := database.GetTx(ctx)
	if ok {
		return tx.ExecContext(ctx, query, args...)
	}
	return r.db.ExecContext(ctx, query, args...)
}

// NamedExec executes a named query without returning any rows
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	tx, ok := database.GetTx(ctx)
	if ok {
		return tx.NamedExecContext(ctx, query, arg)
	}
	return r.db.NamedExecContext(ctx, query, arg)
}
