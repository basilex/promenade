package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type SubscriptionRepository struct {
	*BaseRepository
}

func NewSubscriptionRepository(db *sqlx.DB) repository.ISubscriptionRepository {
	return &SubscriptionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription *entity.Subscription) error {
	query := `
		INSERT INTO billing_subscriptions (
			id, user_id, plan_id, status, current_period_start, current_period_end,
			trial_start, trial_end, cancel_at_period_end, canceled_at, metadata,
			created_at, updated_at
		) VALUES (
			:id, :user_id, :plan_id, :status, :current_period_start, :current_period_end,
			:trial_start, :trial_end, :cancel_at_period_end, :canceled_at, :metadata,
			:created_at, :updated_at
		)`
	return r.NamedExec(ctx, query, subscription)
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Subscription, error) {
	var subscription entity.Subscription
	query := `SELECT * FROM billing_subscriptions WHERE id = $1 AND deleted_at IS NULL`
	err := r.Get(ctx, &subscription, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &subscription, err
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*entity.Subscription, error) {
	var subscriptions []*entity.Subscription
	query := `SELECT * FROM billing_subscriptions WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	err := r.Select(ctx, &subscriptions, query, userID)
	return subscriptions, err
}

func (r *SubscriptionRepository) GetActiveByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.Subscription, error) {
	var subscription entity.Subscription
	query := `
		SELECT * FROM billing_subscriptions 
		WHERE user_id = $1 
		AND status IN ('active', 'trialing')
		AND deleted_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT 1`
	err := r.Get(ctx, &subscription, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &subscription, err
}

func (r *SubscriptionRepository) GetExpiring(ctx context.Context, days, limit int) ([]*entity.Subscription, error) {
	var subscriptions []*entity.Subscription
	expiryDate := time.Now().AddDate(0, 0, days)
	query := `
		SELECT * FROM billing_subscriptions 
		WHERE status = 'active'
		AND current_period_end <= $1
		AND deleted_at IS NULL
		ORDER BY current_period_end ASC
		LIMIT $2`
	err := r.Select(ctx, &subscriptions, query, expiryDate, limit)
	return subscriptions, err
}

func (r *SubscriptionRepository) List(ctx context.Context, status *entity.SubscriptionStatus, limit, offset int) ([]*entity.Subscription, error) {
	var subscriptions []*entity.Subscription
	query := `SELECT * FROM billing_subscriptions WHERE deleted_at IS NULL`
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
	
	err := r.Select(ctx, &subscriptions, query, args...)
	return subscriptions, err
}

func (r *SubscriptionRepository) Count(ctx context.Context, status *entity.SubscriptionStatus) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM billing_subscriptions WHERE deleted_at IS NULL`
	args := []interface{}{}
	
	if status != nil {
		query += ` AND status = $1`
		args = append(args, *status)
	}
	
	err := r.Get(ctx, &count, query, args...)
	return count, err
}

func (r *SubscriptionRepository) Update(ctx context.Context, subscription *entity.Subscription) error {
	query := `
		UPDATE billing_subscriptions SET
			plan_id = :plan_id,
			status = :status,
			current_period_start = :current_period_start,
			current_period_end = :current_period_end,
			trial_start = :trial_start,
			trial_end = :trial_end,
			cancel_at_period_end = :cancel_at_period_end,
			canceled_at = :canceled_at,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`
	return r.NamedExec(ctx, query, subscription)
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE billing_subscriptions SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	return r.Exec(ctx, query, id)
}
