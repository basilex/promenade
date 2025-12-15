package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
)

type userRepository struct {
	*BaseRepository
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
        INSERT INTO users (id, email, name, password, active, created_at, updated_at)
        VALUES (:id, :email, :name, :password, :active, NOW(), NOW())
    `
	return r.NamedExec(ctx, query, user)
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	query := `
        SELECT id, email, name, password, active, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	err := r.Get(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, entity.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	query := `
        SELECT id, email, name, password, active, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	err := r.Get(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, nil // Не ошибка - просто пользователь не найден
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
        UPDATE users
        SET email = :email, 
            name = :name, 
            password = :password, 
            active = :active, 
            updated_at = NOW()
        WHERE id = :id
    `
	return r.NamedExec(ctx, query, user)
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	return r.Exec(ctx, query, id)
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*entity.User, error) {
	var users []*entity.User
	query := `
        SELECT id, email, name, password, active, created_at, updated_at
        FROM users
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	err := r.Select(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	err := r.Get(ctx, &count, query)
	return count, err
}
