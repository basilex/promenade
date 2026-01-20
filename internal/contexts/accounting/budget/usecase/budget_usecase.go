package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/budget"
	"github.com/basilex/promenade/internal/contexts/accounting/budget/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/budget/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IBudgetUseCase defines the interface for budget business logic
type IBudgetUseCase interface {
	CreateBudget(ctx context.Context, organizationID uuidv7.UUID, name string, fiscalYear int, createdBy uuidv7.UUID) (*aggregate.Budget, error)
	AddLine(ctx context.Context, budgetID, accountID uuidv7.UUID, budgetAmount int64, description string, updatedBy uuidv7.UUID) (*aggregate.Budget, error)
	UpdateLine(ctx context.Context, budgetID, lineID uuidv7.UUID, budgetAmount int64, updatedBy uuidv7.UUID) (*aggregate.Budget, error)
	RemoveLine(ctx context.Context, budgetID, lineID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.Budget, error)
	ApproveBudget(ctx context.Context, budgetID, approvedBy uuidv7.UUID) (*aggregate.Budget, error)
	ActivateBudget(ctx context.Context, budgetID, activatedBy uuidv7.UUID) (*aggregate.Budget, error)
	CloseBudget(ctx context.Context, budgetID, closedBy uuidv7.UUID) (*aggregate.Budget, error)
	GetBudgetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error)
	GetBudgetByName(ctx context.Context, organizationID uuidv7.UUID, name string) (*aggregate.Budget, error)
	DeleteBudget(ctx context.Context, id uuidv7.UUID) error
	ListBudgetsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.Budget, error)
	ListBudgetsByStatus(ctx context.Context, organizationID uuidv7.UUID, status aggregate.BudgetStatus) ([]*aggregate.Budget, error)
	ListActiveBudgets(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.Budget, error)
}

type budgetUseCase struct {
	budgetRepo  repository.IBudgetRepository
	auditLogger *audit.AuditLogger
}

// NewBudgetUseCase creates a new budget use case
func NewBudgetUseCase(
	budgetRepo repository.IBudgetRepository,
	auditLogger *audit.AuditLogger,
) IBudgetUseCase {
	return &budgetUseCase{
		budgetRepo:  budgetRepo,
		auditLogger: auditLogger,
	}
}

func (uc *budgetUseCase) CreateBudget(
	ctx context.Context,
	organizationID uuidv7.UUID,
	name string,
	fiscalYear int,
	createdBy uuidv7.UUID,
) (*aggregate.Budget, error) {
	// Check if budget with this name already exists
	existing, err := uc.budgetRepo.GetByName(ctx, organizationID, name)
	if err == nil && existing != nil {
		return nil, budget.ErrBudgetNameEmpty
	}

	// Create new budget
	bdg, err := aggregate.NewBudget(organizationID, name, fiscalYear, createdBy)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := uc.budgetRepo.Create(ctx, bdg); err != nil {
		return nil, err
	}

	return bdg, nil
}

func (uc *budgetUseCase) AddLine(
	ctx context.Context,
	budgetID, accountID uuidv7.UUID,
	budgetAmount int64,
	description string,
	updatedBy uuidv7.UUID,
) (*aggregate.Budget, error) {
	// Get budget
	bdg, err := uc.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Verify budget is in draft status
	if bdg.Status != aggregate.BudgetStatusDraft {
		return nil, budget.ErrCannotModifyApprovedBudget
	}

	// Add line
	if err := bdg.AddLine(accountID, budgetAmount, description); err != nil {
		return nil, err
	}

	bdg.LastUpdatedBy = updatedBy
	bdg.Touch()

	// Save
	if err := uc.budgetRepo.Update(ctx, bdg); err != nil {
		return nil, err
	}

	return bdg, nil
}

func (uc *budgetUseCase) UpdateLine(
	ctx context.Context,
	budgetID, lineID uuidv7.UUID,
	budgetAmount int64,
	updatedBy uuidv7.UUID,
) (*aggregate.Budget, error) {
	// Get budget
	bdg, err := uc.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Verify budget is in draft status
	if bdg.Status != aggregate.BudgetStatusDraft {
		return nil, budget.ErrCannotModifyApprovedBudget
	}

	// Update line
	if err := bdg.UpdateLine(lineID, budgetAmount); err != nil {
		return nil, err
	}

	bdg.LastUpdatedBy = updatedBy
	bdg.Touch()

	// Save
	if err := uc.budgetRepo.Update(ctx, bdg); err != nil {
		return nil, err
	}

	return bdg, nil
}

func (uc *budgetUseCase) RemoveLine(
	ctx context.Context,
	budgetID, lineID uuidv7.UUID,
	updatedBy uuidv7.UUID,
) (*aggregate.Budget, error) {
	// Get budget
	bdg, err := uc.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Verify budget is in draft status
	if bdg.Status != aggregate.BudgetStatusDraft {
		return nil, budget.ErrCannotModifyApprovedBudget
	}

	// Remove line
	if err := bdg.RemoveLine(lineID); err != nil {
		return nil, err
	}

	bdg.LastUpdatedBy = updatedBy
	bdg.Touch()

	// Save
	if err := uc.budgetRepo.Update(ctx, bdg); err != nil {
		return nil, err
	}

	return bdg, nil
}

func (uc *budgetUseCase) ApproveBudget(
	ctx context.Context,
	budgetID, approvedBy uuidv7.UUID,
) (*aggregate.Budget, error) {
	// Get budget
	bdg, err := uc.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Approve budget
	if err := bdg.Approve(approvedBy); err != nil {
		return nil, err
	}

	// Save
	if err := uc.budgetRepo.Update(ctx, bdg); err != nil {
		return nil, err
	}

	return bdg, nil
}

func (uc *budgetUseCase) ActivateBudget(
	ctx context.Context,
	budgetID, activatedBy uuidv7.UUID,
) (*aggregate.Budget, error) {
	// Get budget
	bdg, err := uc.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Activate budget
	if err := bdg.Activate(); err != nil {
		return nil, err
	}
	bdg.LastUpdatedBy = activatedBy
	bdg.Touch()

	// Save
	if err := uc.budgetRepo.Update(ctx, bdg); err != nil {
		return nil, err
	}

	return bdg, nil
}

func (uc *budgetUseCase) CloseBudget(
	ctx context.Context,
	budgetID, closedBy uuidv7.UUID,
) (*aggregate.Budget, error) {
	// Get budget
	bdg, err := uc.budgetRepo.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	// Close budget
	if err := bdg.Close(); err != nil {
		return nil, err
	}
	bdg.LastUpdatedBy = closedBy
	bdg.Touch()

	// Save
	if err := uc.budgetRepo.Update(ctx, bdg); err != nil {
		return nil, err
	}

	return bdg, nil
}

func (uc *budgetUseCase) GetBudgetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
	return uc.budgetRepo.GetByID(ctx, id)
}

func (uc *budgetUseCase) GetBudgetByName(ctx context.Context, organizationID uuidv7.UUID, name string) (*aggregate.Budget, error) {
	return uc.budgetRepo.GetByName(ctx, organizationID, name)
}

func (uc *budgetUseCase) DeleteBudget(ctx context.Context, id uuidv7.UUID) error {
	// Get budget to verify it's in draft status
	bdg, err := uc.budgetRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Can only delete draft budgets
	if bdg.Status != aggregate.BudgetStatusDraft {
		return budget.ErrCannotDeleteApproved
	}

	return uc.budgetRepo.Delete(ctx, id)
}

func (uc *budgetUseCase) ListBudgetsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.Budget, error) {
	return uc.budgetRepo.ListByOrganization(ctx, organizationID, limit, offset)
}

func (uc *budgetUseCase) ListBudgetsByStatus(ctx context.Context, organizationID uuidv7.UUID, status aggregate.BudgetStatus) ([]*aggregate.Budget, error) {
	return uc.budgetRepo.ListByStatus(ctx, organizationID, status)
}

func (uc *budgetUseCase) ListActiveBudgets(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.Budget, error) {
	return uc.budgetRepo.ListActive(ctx, organizationID)
}
