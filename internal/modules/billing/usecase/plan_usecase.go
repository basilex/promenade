package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type IPlanUseCase interface {
	CreatePlan(ctx context.Context, name, slug, description, currency string, amount int64, interval entity.PlanInterval, features []string) (*entity.Plan, error)
	GetPlan(ctx context.Context, id uuidv7.UUID) (*entity.Plan, error)
	GetPlanBySlug(ctx context.Context, slug string) (*entity.Plan, error)
	UpdatePlan(ctx context.Context, id uuidv7.UUID, updates map[string]any) (*entity.Plan, error)
	DeletePlan(ctx context.Context, id uuidv7.UUID) error
	ListPlans(ctx context.Context, status *entity.PlanStatus, limit, offset int) ([]*entity.Plan, error)
	CountPlans(ctx context.Context, status *entity.PlanStatus) (int, error)
	ActivatePlan(ctx context.Context, id uuidv7.UUID) error
	DeactivatePlan(ctx context.Context, id uuidv7.UUID) error
}

type planUseCase struct {
	planRepo         repository.IPlanRepository
	subscriptionRepo repository.ISubscriptionRepository
	eventBus         bus.IBus
}

func NewPlanUseCase(planRepo repository.IPlanRepository, subscriptionRepo repository.ISubscriptionRepository, eventBus bus.IBus) IPlanUseCase {
	return &planUseCase{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		eventBus:         eventBus,
	}
}

func (uc *planUseCase) CreatePlan(ctx context.Context, name, slug, description, currency string, amount int64, interval entity.PlanInterval, features []string) (*entity.Plan, error) {
	plan, err := entity.NewPlan(name, slug, description, amount, currency, interval)
	if err != nil {
		return nil, err
	}
	plan.Features = features
	if err := uc.planRepo.Create(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (uc *planUseCase) GetPlan(ctx context.Context, id uuidv7.UUID) (*entity.Plan, error) {
	return uc.planRepo.GetByID(ctx, id)
}

func (uc *planUseCase) GetPlanBySlug(ctx context.Context, slug string) (*entity.Plan, error) {
	return uc.planRepo.GetBySlug(ctx, slug)
}

func (uc *planUseCase) UpdatePlan(ctx context.Context, id uuidv7.UUID, updates map[string]any) (*entity.Plan, error) {
	plan, err := uc.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Apply updates (simplified for now)
	if err := uc.planRepo.Update(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (uc *planUseCase) DeletePlan(ctx context.Context, id uuidv7.UUID) error {
	return uc.planRepo.Delete(ctx, id)
}

func (uc *planUseCase) ListPlans(ctx context.Context, status *entity.PlanStatus, limit, offset int) ([]*entity.Plan, error) {
	return uc.planRepo.List(ctx, status, limit, offset)
}

func (uc *planUseCase) CountPlans(ctx context.Context, status *entity.PlanStatus) (int, error) {
	return uc.planRepo.Count(ctx, status)
}

func (uc *planUseCase) ActivatePlan(ctx context.Context, id uuidv7.UUID) error {
	plan, err := uc.planRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	plan.Activate()
	return uc.planRepo.Update(ctx, plan)
}

func (uc *planUseCase) DeactivatePlan(ctx context.Context, id uuidv7.UUID) error {
	plan, err := uc.planRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	plan.Deactivate()
	return uc.planRepo.Update(ctx, plan)
}
