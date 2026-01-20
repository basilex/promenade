package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/costcenter"
	"github.com/basilex/promenade/internal/contexts/accounting/costcenter/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/costcenter/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICostCenterUseCase defines the interface for cost center business logic
type ICostCenterUseCase interface {
	CreateCostCenter(ctx context.Context, organizationID uuidv7.UUID, code, name string, centerType aggregate.CenterType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*aggregate.CostCenter, error)
	GetCostCenterByID(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error)
	GetCostCenterByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.CostCenter, error)
	SetParent(ctx context.Context, id, parentID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.CostCenter, error)
	SetManager(ctx context.Context, id, managerID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.CostCenter, error)
	ActivateCostCenter(ctx context.Context, id, updatedBy uuidv7.UUID) (*aggregate.CostCenter, error)
	DeactivateCostCenter(ctx context.Context, id, updatedBy uuidv7.UUID) (*aggregate.CostCenter, error)
	DeleteCostCenter(ctx context.Context, id uuidv7.UUID) error
	ListCostCentersByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.CostCenter, error)
	ListChildCostCenters(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.CostCenter, error)
	ListCostCentersByType(ctx context.Context, organizationID uuidv7.UUID, centerType aggregate.CenterType) ([]*aggregate.CostCenter, error)
	ListActiveCostCenters(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.CostCenter, error)
}

type costCenterUseCase struct {
	costCenterRepo repository.ICostCenterRepository
	auditLogger    *audit.AuditLogger
}

// NewCostCenterUseCase creates a new cost center use case
func NewCostCenterUseCase(
	costCenterRepo repository.ICostCenterRepository,
	auditLogger *audit.AuditLogger,
) ICostCenterUseCase {
	return &costCenterUseCase{
		costCenterRepo: costCenterRepo,
		auditLogger:    auditLogger,
	}
}

func (uc *costCenterUseCase) CreateCostCenter(
	ctx context.Context,
	organizationID uuidv7.UUID,
	code, name string,
	centerType aggregate.CenterType,
	parentID *uuidv7.UUID,
	createdBy uuidv7.UUID,
) (*aggregate.CostCenter, error) {
	// Check if cost center with this code already exists
	existing, err := uc.costCenterRepo.GetByCode(ctx, organizationID, code)
	if err == nil && existing != nil {
		return nil, costcenter.ErrCodeAlreadyExists
	}

	// Create new cost center
	var cc *aggregate.CostCenter
	var createErr error

	if parentID != nil {
		parent, err := uc.costCenterRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		cc, createErr = aggregate.NewChildCostCenter(organizationID, code, name, centerType, *parentID, parent.Level, createdBy)
	} else {
		cc, createErr = aggregate.NewCostCenter(organizationID, code, name, centerType, createdBy)
	}

	if createErr != nil {
		return nil, createErr
	}

	// Save to repository
	if err := uc.costCenterRepo.Create(ctx, cc); err != nil {
		return nil, err
	}

	return cc, nil
}

func (uc *costCenterUseCase) GetCostCenterByID(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
	return uc.costCenterRepo.GetByID(ctx, id)
}

func (uc *costCenterUseCase) GetCostCenterByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.CostCenter, error) {
	return uc.costCenterRepo.GetByCode(ctx, organizationID, code)
}

func (uc *costCenterUseCase) SetParent(ctx context.Context, id, parentID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.CostCenter, error) {
	// Get cost center
	cc, err := uc.costCenterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get parent to determine level
	parent, err := uc.costCenterRepo.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}

	// Set parent
	if err := cc.SetParent(parentID, parent.Level+1); err != nil {
		return nil, err
	}

	cc.LastUpdatedBy = updatedBy
	cc.Touch()

	// Save
	if err := uc.costCenterRepo.Update(ctx, cc); err != nil {
		return nil, err
	}

	return cc, nil
}

func (uc *costCenterUseCase) SetManager(ctx context.Context, id, managerID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.CostCenter, error) {
	// Get cost center
	cc, err := uc.costCenterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Set manager
	cc.SetManager(managerID)
	cc.LastUpdatedBy = updatedBy
	cc.Touch()

	// Save
	if err := uc.costCenterRepo.Update(ctx, cc); err != nil {
		return nil, err
	}

	return cc, nil
}

func (uc *costCenterUseCase) ActivateCostCenter(ctx context.Context, id, updatedBy uuidv7.UUID) (*aggregate.CostCenter, error) {
	// Get cost center
	cc, err := uc.costCenterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Activate
	cc.Activate()
	cc.LastUpdatedBy = updatedBy
	cc.Touch()

	// Save
	if err := uc.costCenterRepo.Update(ctx, cc); err != nil {
		return nil, err
	}

	return cc, nil
}

func (uc *costCenterUseCase) DeactivateCostCenter(ctx context.Context, id, updatedBy uuidv7.UUID) (*aggregate.CostCenter, error) {
	// Get cost center
	cc, err := uc.costCenterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Deactivate
	cc.Deactivate()
	cc.LastUpdatedBy = updatedBy
	cc.Touch()

	// Save
	if err := uc.costCenterRepo.Update(ctx, cc); err != nil {
		return nil, err
	}

	return cc, nil
}

func (uc *costCenterUseCase) DeleteCostCenter(ctx context.Context, id uuidv7.UUID) error {
	// Check if cost center has children
	children, err := uc.costCenterRepo.ListChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return costcenter.ErrHasChildren
	}

	return uc.costCenterRepo.Delete(ctx, id)
}

func (uc *costCenterUseCase) ListCostCentersByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.CostCenter, error) {
	return uc.costCenterRepo.ListByOrganization(ctx, organizationID, limit, offset)
}

func (uc *costCenterUseCase) ListChildCostCenters(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
	return uc.costCenterRepo.ListChildren(ctx, parentID)
}

func (uc *costCenterUseCase) ListCostCentersByType(ctx context.Context, organizationID uuidv7.UUID, centerType aggregate.CenterType) ([]*aggregate.CostCenter, error) {
	return uc.costCenterRepo.ListByType(ctx, organizationID, centerType)
}

func (uc *costCenterUseCase) ListActiveCostCenters(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
	return uc.costCenterRepo.ListActive(ctx, organizationID)
}
