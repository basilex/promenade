package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type ISubscriptionUseCase interface {
	CreateSubscription(ctx context.Context, userID, planID uuidv7.UUID) (*entity.Subscription, error)
	GetSubscription(ctx context.Context, id uuidv7.UUID) (*entity.Subscription, error)
	GetUserSubscriptions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Subscription, error)
	GetActiveSubscription(ctx context.Context, userID uuidv7.UUID) (*entity.Subscription, error)
	CancelSubscription(ctx context.Context, userID, subscriptionID uuidv7.UUID) error
	UpgradeSubscription(ctx context.Context, userID, subscriptionID, newPlanID uuidv7.UUID) (*entity.Subscription, error)
	ListSubscriptions(ctx context.Context, status *entity.SubscriptionStatus, limit, offset int) ([]*entity.Subscription, error)
	CountSubscriptions(ctx context.Context, status *entity.SubscriptionStatus) (int, error)
}

type subscriptionUseCase struct {
	subscriptionRepo repository.ISubscriptionRepository
	planRepo         repository.IPlanRepository
	invoiceRepo      repository.IInvoiceRepository
	eventBus         bus.IBus
}

func NewSubscriptionUseCase(subscriptionRepo repository.ISubscriptionRepository, planRepo repository.IPlanRepository, invoiceRepo repository.IInvoiceRepository, eventBus bus.IBus) ISubscriptionUseCase {
	return &subscriptionUseCase{
		subscriptionRepo: subscriptionRepo,
		planRepo:         planRepo,
		invoiceRepo:      invoiceRepo,
		eventBus:         eventBus,
	}
}

func (uc *subscriptionUseCase) CreateSubscription(ctx context.Context, userID, planID uuidv7.UUID) (*entity.Subscription, error) {
	plan, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, err
	}
	
	subscription, err := entity.NewSubscription(userID, planID, plan.TrialDays)
	if err != nil {
		return nil, err
	}
	
	if err := uc.subscriptionRepo.Create(ctx, subscription); err != nil {
		return nil, err
	}
	return subscription, nil
}

func (uc *subscriptionUseCase) GetSubscription(ctx context.Context, id uuidv7.UUID) (*entity.Subscription, error) {
	return uc.subscriptionRepo.GetByID(ctx, id)
}

func (uc *subscriptionUseCase) GetUserSubscriptions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Subscription, error) {
	return uc.subscriptionRepo.GetByUserID(ctx, userID)
}

func (uc *subscriptionUseCase) GetActiveSubscription(ctx context.Context, userID uuidv7.UUID) (*entity.Subscription, error) {
	return uc.subscriptionRepo.GetActiveByUserID(ctx, userID)
}

func (uc *subscriptionUseCase) CancelSubscription(ctx context.Context, userID, subscriptionID uuidv7.UUID) error {
	subscription, err := uc.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return err
	}
	subscription.Cancel(false) // Cancel at period end
	return uc.subscriptionRepo.Update(ctx, subscription)
}

func (uc *subscriptionUseCase) UpgradeSubscription(ctx context.Context, userID, subscriptionID, newPlanID uuidv7.UUID) (*entity.Subscription, error) {
	subscription, err := uc.subscriptionRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	subscription.PlanID = newPlanID
	if err := uc.subscriptionRepo.Update(ctx, subscription); err != nil {
		return nil, err
	}
	return subscription, nil
}

func (uc *subscriptionUseCase) ListSubscriptions(ctx context.Context, status *entity.SubscriptionStatus, limit, offset int) ([]*entity.Subscription, error) {
	return uc.subscriptionRepo.List(ctx, status, limit, offset)
}

func (uc *subscriptionUseCase) CountSubscriptions(ctx context.Context, status *entity.SubscriptionStatus) (int, error) {
	return uc.subscriptionRepo.Count(ctx, status)
}
