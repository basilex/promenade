package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
)

type roleRepository struct {
	*BaseRepository
}

func NewRoleRepository(db *sqlx.DB) repository.RoleRepository {
	return &roleRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *roleRepository) Create(ctx context.Context, role *entity.Role) error {
	query := `
        INSERT INTO roles (id, name, active, created_at, updated_at)
        VALUES (:id, :name, :active, NOW(), NOW())
    `
	return r.NamedExec(ctx, query, role)
}

func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
	var role entity.Role
	query := `
        SELECT id, name, active, created_at, updated_at
        FROM roles
        WHERE id = $1
    `

	err := r.Get(ctx, &role, query, id)
	if err == sql.ErrNoRows {
		return nil, entity.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *entity.Role) error {
	query := `
        UPDATE roles
        SET name = :name, active = :active, updated_at = NOW()
        WHERE id = :id
    `
	return r.NamedExec(ctx, query, role)
}

func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *roleRepository) List(ctx context.Context, limit, offset int) ([]*entity.Role, error) {
	var roles []*entity.Role
	query := `
        SELECT id, name, active, created_at, updated_at
        FROM roles
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	err := r.Select(ctx, &roles, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *roleRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM roles`
	err := r.Get(ctx, &count, query)
	return count, err
}
