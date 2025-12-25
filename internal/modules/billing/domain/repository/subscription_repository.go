package repository

import (
	"context"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ISubscriptionRepository defines methods for subscription persistence
type ISubscriptionRepository interface {
	// Create creates a new subscription
	Create(ctx context.Context, subscription *entity.Subscription) error

	// GetByID retrieves a subscription by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Subscription, error)

	// GetByUserID retrieves all subscriptions for a user
	GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*entity.Subscription, error)

	// GetActiveByUserID retrieves active subscription for a user
	GetActiveByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.Subscription, error)

	// GetExpiring retrieves subscriptions expiring within days
	GetExpiring(ctx context.Context, days, limit int) ([]*entity.Subscription, error)

	// List retrieves subscriptions with optional status filter
	List(ctx context.Context, status *entity.SubscriptionStatus, limit, offset int) ([]*entity.Subscription, error)

	// Count counts subscriptions with optional status filter
	Count(ctx context.Context, status *entity.SubscriptionStatus) (int, error)

	// Update updates an existing subscription
	Update(ctx context.Context, subscription *entity.Subscription) error

	// Delete soft-deletes a subscription
	Delete(ctx context.Context, id uuidv7.UUID) error
}
