package usecase

import (
	"context"
	"time"

	subscriptionerrors "github.com/basilex/promenade/internal/contexts/billing/subscription"
	"github.com/basilex/promenade/internal/contexts/billing/subscription/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ISubscriptionUseCase defines subscription business logic operations
type ISubscriptionUseCase interface {
	// CreateSubscription creates a new subscription
	CreateSubscription(ctx context.Context, customerID uuidv7.UUID, planID string, billingPeriod aggregate.BillingPeriod, currency string, amount int64, trialDays int) (*aggregate.Subscription, error)

	// GetSubscription retrieves subscription by ID
	GetSubscription(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error)

	// UpdateSubscription updates subscription details
	UpdateSubscription(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error

	// DeleteSubscription soft deletes a subscription
	DeleteSubscription(ctx context.Context, id uuidv7.UUID) error

	// ListSubscriptions returns paginated list of subscriptions
	ListSubscriptions(ctx context.Context, page, pageSize int) ([]*aggregate.Subscription, error)

	// ListByCustomer returns all subscriptions for a customer
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*aggregate.Subscription, error)

	// ListByStatus returns subscriptions filtered by status
	ListByStatus(ctx context.Context, status aggregate.SubscriptionStatus) ([]*aggregate.Subscription, error)

	// ActivateSubscription activates a subscription
	ActivateSubscription(ctx context.Context, id uuidv7.UUID) error

	// PauseSubscription pauses a subscription
	PauseSubscription(ctx context.Context, id uuidv7.UUID) error

	// ResumeSubscription resumes a paused subscription
	ResumeSubscription(ctx context.Context, id uuidv7.UUID) error

	// CancelSubscription cancels a subscription
	CancelSubscription(ctx context.Context, id uuidv7.UUID, reason string, effectiveDate time.Time) error

	// RenewSubscription renews a subscription for another billing period
	RenewSubscription(ctx context.Context, id uuidv7.UUID) error

	// ExpireSubscription marks a subscription as expired
	ExpireSubscription(ctx context.Context, id uuidv7.UUID) error

	// CountByStatus counts subscriptions by status
	CountByStatus(ctx context.Context, status aggregate.SubscriptionStatus) (int64, error)

	// GetTotalRevenue calculates total Monthly Recurring Revenue (MRR)
	GetTotalRevenue(ctx context.Context) (int64, error)
}

// ISubscriptionRepository defines subscription data access operations (dependency inversion)
type ISubscriptionRepository interface {
	// Create inserts a new subscription
	Create(ctx context.Context, subscription *aggregate.Subscription) error

	// GetByID retrieves subscription by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error)

	// Update modifies an existing subscription
	Update(ctx context.Context, subscription *aggregate.Subscription) error

	// Delete soft deletes a subscription
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ListSubscriptions returns paginated list of subscriptions
	ListSubscriptions(ctx context.Context, page, pageSize int) ([]*aggregate.Subscription, error)

	// ListByCustomer returns all subscriptions for a customer
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*aggregate.Subscription, error)

	// ListByStatus returns subscriptions filtered by status
	ListByStatus(ctx context.Context, status aggregate.SubscriptionStatus) ([]*aggregate.Subscription, error)

	// CountByStatus counts subscriptions by status
	CountByStatus(ctx context.Context, status aggregate.SubscriptionStatus) (int64, error)

	// GetTotalRevenue calculates total Monthly Recurring Revenue (MRR)
	GetTotalRevenue(ctx context.Context) (int64, error)
}

type SubscriptionUseCase struct {
	repo ISubscriptionRepository
}

// NewSubscriptionUseCase creates a new subscription use case
func NewSubscriptionUseCase(repo ISubscriptionRepository) ISubscriptionUseCase {
	return &SubscriptionUseCase{
		repo: repo,
	}
}

func (uc *SubscriptionUseCase) CreateSubscription(ctx context.Context, customerID uuidv7.UUID, planID string, billingPeriod aggregate.BillingPeriod, currency string, amount int64, trialDays int) (*aggregate.Subscription, error) {
	subscription, err := aggregate.NewSubscription(
		customerID,
		planID,
		billingPeriod,
		currency,
		amount,
		time.Now(),
		trialDays,
	)
	if err != nil {
		return nil, err // Domain error from entity constructor
	}

	if err := uc.repo.Create(ctx, subscription); err != nil {
		return nil, subscriptionerrors.ErrCreateFailed
	}

	return subscription, nil
}

func (uc *SubscriptionUseCase) GetSubscription(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err // Propagate repository error (subscriptionerrors.ErrSubscriptionNotFound or technical)
	}
	return subscription, nil
}

func (uc *SubscriptionUseCase) UpdateSubscription(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	subscription.PlanID = planID
	subscription.Amount.Amount = amount
	subscription.Touch()

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return subscriptionerrors.ErrUpdateFailed
	}

	return nil
}

func (uc *SubscriptionUseCase) DeleteSubscription(ctx context.Context, id uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return subscriptionerrors.ErrDeleteFailed
	}
	return nil
}

func (uc *SubscriptionUseCase) ListSubscriptions(ctx context.Context, page, pageSize int) ([]*aggregate.Subscription, error) {
	subscriptions, err := uc.repo.ListSubscriptions(ctx, page, pageSize)
	if err != nil {
		return nil, subscriptionerrors.ErrQueryFailed
	}
	return subscriptions, nil
}

func (uc *SubscriptionUseCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*aggregate.Subscription, error) {
	subscriptions, err := uc.repo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, subscriptionerrors.ErrQueryFailed
	}
	return subscriptions, nil
}

func (uc *SubscriptionUseCase) ListByStatus(ctx context.Context, status aggregate.SubscriptionStatus) ([]*aggregate.Subscription, error) {
	subscriptions, err := uc.repo.ListByStatus(ctx, status)
	if err != nil {
		return nil, subscriptionerrors.ErrQueryFailed
	}
	return subscriptions, nil
}

func (uc *SubscriptionUseCase) ActivateSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Activate(); err != nil {
		return err // Propagate domain error (subscriptionerrors.ErrCannotActivate)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return subscriptionerrors.ErrUpdateFailed
	}

	return nil
}

func (uc *SubscriptionUseCase) PauseSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Pause(); err != nil {
		return err // Propagate domain error (subscriptionerrors.ErrCanOnlyPauseActive)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return subscriptionerrors.ErrUpdateFailed
	}

	return nil
}

func (uc *SubscriptionUseCase) ResumeSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Resume(); err != nil {
		return err // Propagate domain error (subscriptionerrors.ErrCanOnlyResumePaused)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return subscriptionerrors.ErrUpdateFailed
	}

	return nil
}

func (uc *SubscriptionUseCase) CancelSubscription(ctx context.Context, id uuidv7.UUID, reason string, effectiveDate time.Time) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Cancel(reason, effectiveDate); err != nil {
		return err // Propagate domain error (subscriptionerrors.ErrAlreadyInTerminalStatus)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return subscriptionerrors.ErrUpdateFailed
	}

	return nil
}

func (uc *SubscriptionUseCase) RenewSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Renew(); err != nil {
		return err // Propagate domain error (subscriptionerrors.ErrCanOnlyRenewActive)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return subscriptionerrors.ErrUpdateFailed
	}

	return nil
}

func (uc *SubscriptionUseCase) ExpireSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Expire(); err != nil {
		return err // Propagate domain error (subscriptionerrors.ErrAlreadyExpired)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return subscriptionerrors.ErrUpdateFailed
	}

	return nil
}

func (uc *SubscriptionUseCase) CountByStatus(ctx context.Context, status aggregate.SubscriptionStatus) (int64, error) {
	count, err := uc.repo.CountByStatus(ctx, status)
	if err != nil {
		return 0, subscriptionerrors.ErrQueryFailed
	}
	return count, nil
}

func (uc *SubscriptionUseCase) GetTotalRevenue(ctx context.Context) (int64, error) {
	revenue, err := uc.repo.GetTotalRevenue(ctx)
	if err != nil {
		return 0, subscriptionerrors.ErrQueryFailed
	}
	return revenue, nil
}
