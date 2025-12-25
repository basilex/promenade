package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type PlanRepository struct {
	*BaseRepository
}

func NewPlanRepository(db *sqlx.DB) repository.IPlanRepository {
	return &PlanRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *PlanRepository) Create(ctx context.Context, plan *entity.Plan) error {
	query := `
		INSERT INTO billing_plans (
			id, slug, name, description, status, interval, amount, currency, trial_days,
			features, max_users, max_projects, max_storage,
			is_unlimited_users, is_unlimited_projects, is_unlimited_storage,
			created_at, updated_at
		) VALUES (
			:id, :slug, :name, :description, :status, :interval, :amount, :currency, :trial_days,
			:features, :max_users, :max_projects, :max_storage,
			:is_unlimited_users, :is_unlimited_projects, :is_unlimited_storage,
			:created_at, :updated_at
		)`
	return r.NamedExec(ctx, query, plan)
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Plan, error) {
	var plan entity.Plan
	query := `SELECT * FROM billing_plans WHERE id = $1 AND deleted_at IS NULL`
	err := r.Get(ctx, &plan, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &plan, err
}

func (r *PlanRepository) GetBySlug(ctx context.Context, slug string) (*entity.Plan, error) {
	var plan entity.Plan
	query := `SELECT * FROM billing_plans WHERE slug = $1 AND deleted_at IS NULL`
	err := r.Get(ctx, &plan, query, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &plan, err
}

func (r *PlanRepository) List(ctx context.Context, status *entity.PlanStatus, limit, offset int) ([]*entity.Plan, error) {
	var plans []*entity.Plan
	query := `SELECT * FROM billing_plans WHERE deleted_at IS NULL`
	args := []interface{}{}
	
	if status != nil {
		query += ` AND status = $1`
		args = append(args, *status)
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}
	
	err := r.Select(ctx, &plans, query, args...)
	return plans, err
}

func (r *PlanRepository) Count(ctx context.Context, status *entity.PlanStatus) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM billing_plans WHERE deleted_at IS NULL`
	args := []interface{}{}
	
	if status != nil {
		query += ` AND status = $1`
		args = append(args, *status)
	}
	
	err := r.Get(ctx, &count, query, args...)
	return count, err
}

func (r *PlanRepository) Update(ctx context.Context, plan *entity.Plan) error {
	query := `
		UPDATE billing_plans SET
			slug = :slug,
			name = :name,
			description = :description,
			status = :status,
			interval = :interval,
			amount = :amount,
			currency = :currency,
			trial_days = :trial_days,
			features = :features,
			max_users = :max_users,
			max_projects = :max_projects,
			max_storage = :max_storage,
			is_unlimited_users = :is_unlimited_users,
			is_unlimited_projects = :is_unlimited_projects,
			is_unlimited_storage = :is_unlimited_storage,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`
	return r.NamedExec(ctx, query, plan)
}

func (r *PlanRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE billing_plans SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	return r.Exec(ctx, query, id)
}
