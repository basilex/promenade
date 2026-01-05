package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/database"
)

type BaseRepository struct {
	db *sqlx.DB
}

func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

func (r *BaseRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	executor := r.getExecutor(ctx)
	return sqlx.GetContext(ctx, executor, dest, query, args...)
}

func (r *BaseRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	executor := r.getExecutor(ctx)
	return sqlx.SelectContext(ctx, executor, dest, query, args...)
}

func (r *BaseRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	executor := r.getExecutor(ctx)
	return executor.ExecContext(ctx, query, args...)
}

// NamedExec executes a named query
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	executor := r.getExecutor(ctx)
	return sqlx.NamedExecContext(ctx, executor, query, arg)
}

func (r *BaseRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

func (r *BaseRepository) Count(ctx context.Context, query string, args ...interface{}) (int, error) {
	var count int
	err := r.Get(ctx, &count, query, args...)
	return count, err
}
