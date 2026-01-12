package subscription

import (
	"context"
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
		return nil, err // Domain error from entity constructor
	}

	if err := uc.repo.Create(ctx, subscription); err != nil {
		return nil, ErrCreateFailed
	}

	return subscription, nil
}

func (uc *useCase) GetSubscription(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err // Propagate repository error (ErrSubscriptionNotFound or technical)
	}
	return subscription, nil
}

func (uc *useCase) UpdateSubscription(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	subscription.PlanID = planID
	subscription.Amount.Amount = amount
	subscription.Touch()

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return ErrUpdateFailed
	}

	return nil
}

func (uc *useCase) DeleteSubscription(ctx context.Context, id uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return ErrDeleteFailed
	}
	return nil
}

func (uc *useCase) ListSubscriptions(ctx context.Context, page, pageSize int) ([]*Subscription, error) {
	subscriptions, err := uc.repo.ListSubscriptions(ctx, page, pageSize)
	if err != nil {
		return nil, ErrQueryFailed
	}
	return subscriptions, nil
}

func (uc *useCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*Subscription, error) {
	subscriptions, err := uc.repo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, ErrQueryFailed
	}
	return subscriptions, nil
}

func (uc *useCase) ListByStatus(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error) {
	subscriptions, err := uc.repo.ListByStatus(ctx, status)
	if err != nil {
		return nil, ErrQueryFailed
	}
	return subscriptions, nil
}

func (uc *useCase) ActivateSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Activate(); err != nil {
		return err // Propagate domain error (ErrCannotActivate)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return ErrUpdateFailed
	}

	return nil
}

func (uc *useCase) PauseSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Pause(); err != nil {
		return err // Propagate domain error (ErrCanOnlyPauseActive)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return ErrUpdateFailed
	}

	return nil
}

func (uc *useCase) ResumeSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Resume(); err != nil {
		return err // Propagate domain error (ErrCanOnlyResumePaused)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return ErrUpdateFailed
	}

	return nil
}

func (uc *useCase) CancelSubscription(ctx context.Context, id uuidv7.UUID, reason string, effectiveDate time.Time) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Cancel(reason, effectiveDate); err != nil {
		return err // Propagate domain error (ErrAlreadyInTerminalStatus)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return ErrUpdateFailed
	}

	return nil
}

func (uc *useCase) RenewSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Renew(); err != nil {
		return err // Propagate domain error (ErrCanOnlyRenewActive)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return ErrUpdateFailed
	}

	return nil
}

func (uc *useCase) ExpireSubscription(ctx context.Context, id uuidv7.UUID) error {
	subscription, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err // Propagate repository error
	}

	if err := subscription.Expire(); err != nil {
		return err // Propagate domain error (ErrAlreadyExpired)
	}

	if err := uc.repo.Update(ctx, subscription); err != nil {
		return ErrUpdateFailed
	}

	return nil
}

func (uc *useCase) CountByStatus(ctx context.Context, status SubscriptionStatus) (int64, error) {
	count, err := uc.repo.CountByStatus(ctx, status)
	if err != nil {
		return 0, ErrQueryFailed
	}
	return count, nil
}

func (uc *useCase) GetTotalRevenue(ctx context.Context) (int64, error) {
	revenue, err := uc.repo.GetTotalRevenue(ctx)
	if err != nil {
		return 0, ErrQueryFailed
	}
	return revenue, nil
}
