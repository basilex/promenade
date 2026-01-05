package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines subscription business logic operations
type IUseCase interface {
	// CreateSubscription creates a new subscription
	CreateSubscription(ctx context.Context, customerID uuidv7.UUID, planID string, billingPeriod BillingPeriod, currency string, amount int64, trialDays int) (*Subscription, error)

	// GetSubscription retrieves subscription by ID
	GetSubscription(ctx context.Context, id uuidv7.UUID) (*Subscription, error)

	// UpdateSubscription updates subscription details
	UpdateSubscription(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error

	// DeleteSubscription soft deletes a subscription
	DeleteSubscription(ctx context.Context, id uuidv7.UUID) error

	// ListSubscriptions returns paginated list of subscriptions
	ListSubscriptions(ctx context.Context, page, pageSize int) ([]*Subscription, error)

	// ListByCustomer returns all subscriptions for a customer
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*Subscription, error)

	// ListByStatus returns subscriptions filtered by status
	ListByStatus(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error)

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
	CountByStatus(ctx context.Context, status SubscriptionStatus) (int64, error)

	// GetTotalRevenue calculates total Monthly Recurring Revenue (MRR)
	GetTotalRevenue(ctx context.Context) (int64, error)
}

type useCase struct {
	repo IRepository
}

// NewUseCase creates a new subscription use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

func (uc *useCase) CreateSubscription(ctx context.Context, customerID uuidv7.UUID, planID string, billingPeriod BillingPeriod, currency string, amount int64, trialDays int) (*Subscription, error) {
	subscription, err := NewSubscription(
		customerID,
		planID,
		billingPeriod,
		currency,
		amount,
		time.Now(),
		trialDays,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	if err := uc.repo.Create(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	return subscription, nil
}

func (uc *useCase) GetSubscription(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return subscription, nil
}

func (uc *useCase) UpdateSubscription(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	subscription.PlanID = planID
	subscription.Amount.Amount = amount
	subscription.Touch()

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

func (uc *useCase) DeleteSubscription(ctx context.Context, id uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}
	return nil
}

func (uc *useCase) ListSubscriptions(ctx context.Context, page, pageSize int) ([]*Subscription, error) {
	subscriptions, err := uc.repo.ListSubscriptions(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	return subscriptions, nil
}

func (uc *useCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*Subscription, error) {
	subscriptions, err := uc.repo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by customer: %w", err)
	}
	return subscriptions, nil
}

func (uc *useCase) ListByStatus(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error) {
	subscriptions, err := uc.repo.ListByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by status: %w", err)
	}
	return subscriptions, nil
}

func (uc *useCase) ActivateSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if err := subscription.Activate(); err != nil {
		return fmt.Errorf("failed to activate subscription: %w", err)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}

func (uc *useCase) PauseSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if err := subscription.Pause(); err != nil {
		return fmt.Errorf("failed to pause subscription: %w", err)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}

func (uc *useCase) ResumeSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if err := subscription.Resume(); err != nil {
		return fmt.Errorf("failed to resume subscription: %w", err)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}

func (uc *useCase) CancelSubscription(ctx context.Context, id uuidv7.UUID, reason string, effectiveDate time.Time) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if err := subscription.Cancel(reason, effectiveDate); err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}

func (uc *useCase) RenewSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if err := subscription.Renew(); err != nil {
		return fmt.Errorf("failed to renew subscription: %w", err)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}

func (uc *useCase) ExpireSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if err := subscription.Expire(); err != nil {
		return fmt.Errorf("failed to expire subscription: %w", err)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}

func (uc *useCase) CountByStatus(ctx context.Context, status SubscriptionStatus) (int64, error) {
	count, err := uc.repo.CountByStatus(ctx, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}
	return count, nil
}

func (uc *useCase) GetTotalRevenue(ctx context.Context) (int64, error) {
	revenue, err := uc.repo.GetTotalRevenue(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total revenue: %w", err)
	}
	return revenue, nil
}
