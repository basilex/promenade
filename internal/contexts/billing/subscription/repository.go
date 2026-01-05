package subscription

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines subscription data access operations
type IRepository interface {
	// Create inserts a new subscription
	Create(ctx context.Context, subscription *Subscription) error

	// GetByID retrieves subscription by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*Subscription, error)

	// Update modifies an existing subscription
	Update(ctx context.Context, subscription *Subscription) error

	// Delete soft deletes a subscription
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ListSubscriptions returns paginated list of subscriptions
	ListSubscriptions(ctx context.Context, page, pageSize int) ([]*Subscription, error)

	// ListByCustomer returns all subscriptions for a customer
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*Subscription, error)

	// ListByStatus returns subscriptions filtered by status
	ListByStatus(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error)

	// CountByStatus counts subscriptions by status
	CountByStatus(ctx context.Context, status SubscriptionStatus) (int64, error)

	// GetTotalRevenue calculates total Monthly Recurring Revenue (MRR)
	GetTotalRevenue(ctx context.Context) (int64, error)
}
