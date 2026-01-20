package usecase

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/cache"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ITaxCodeUseCase defines the interface for tax code business logic
type ITaxCodeUseCase interface {
	CreateTaxCode(ctx context.Context, organizationID uuidv7.UUID, code, name string, taxType aggregate.TaxType, rate int, createdBy uuidv7.UUID) (*aggregate.TaxCode, error)
	GetTaxCodeByID(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error)
	GetTaxCodeByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.TaxCode, error)
	UpdateTaxCode(ctx context.Context, id uuidv7.UUID, rate int, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error)
	SetTaxPayableAccount(ctx context.Context, id, accountID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error)
	SetTaxReceivableAccount(ctx context.Context, id, accountID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error)
	ActivateTaxCode(ctx context.Context, id, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error)
	DeactivateTaxCode(ctx context.Context, id, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error)
	DeleteTaxCode(ctx context.Context, id uuidv7.UUID) error
	ListTaxCodesByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.TaxCode, error)
	ListTaxCodesByType(ctx context.Context, organizationID uuidv7.UUID, taxType aggregate.TaxType) ([]*aggregate.TaxCode, error)
	ListActiveTaxCodes(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.TaxCode, error)
}

type taxCodeUseCase struct {
	taxCodeRepo  repository.ITaxCodeRepository
	taxCodeCache *cache.TaxCodeCache
	auditLogger  *audit.AuditLogger
}

// NewTaxCodeUseCase creates a new tax code use case
func NewTaxCodeUseCase(
	taxCodeRepo repository.ITaxCodeRepository,
	taxCodeCache *cache.TaxCodeCache,
	auditLogger *audit.AuditLogger,
) ITaxCodeUseCase {
	return &taxCodeUseCase{
		taxCodeRepo:  taxCodeRepo,
		taxCodeCache: taxCodeCache,
		auditLogger:  auditLogger,
	}
}

func (uc *taxCodeUseCase) CreateTaxCode(
	ctx context.Context,
	organizationID uuidv7.UUID,
	code, name string,
	taxType aggregate.TaxType,
	rate int,
	createdBy uuidv7.UUID,
) (*aggregate.TaxCode, error) {
	// Check if tax code with this code already exists
	existing, err := uc.taxCodeRepo.GetByCode(ctx, organizationID, code)
	if err == nil && existing != nil {
		return nil, taxcode.ErrTaxCodeDuplicate
	}

	// Create new tax code
	tc, err := aggregate.NewTaxCode(organizationID, code, name, taxType, rate, createdBy)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := uc.taxCodeRepo.Create(ctx, tc); err != nil {
		return nil, err
	}

	// Log audit
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogCreate(ctx, audit.AuditRecord{
			EntityType:     "tax_code",
			EntityID:       tc.ID,
			Action:         "create",
			OrganizationID: organizationID,
			UserID:         createdBy,
			Details:        map[string]interface{}{"code": code, "name": name, "type": taxType, "rate": rate},
		})
	}

	// Invalidate cache
	uc.taxCodeCache.Invalidate(organizationID)

	return tc, nil
}

func (uc *taxCodeUseCase) GetTaxCodeByID(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
	// Try cache first
	if tc, err := uc.taxCodeCache.GetByID(ctx, id); err == nil && tc != nil {
		return tc, nil
	}

	// Fallback to repository
	return uc.taxCodeRepo.GetByID(ctx, id)
}

func (uc *taxCodeUseCase) GetTaxCodeByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.TaxCode, error) {
	// Try cache first
	if tc, err := uc.taxCodeCache.GetByCode(ctx, organizationID, code); err == nil && tc != nil {
		return tc, nil
	}

	// Fallback to repository
	return uc.taxCodeRepo.GetByCode(ctx, organizationID, code)
}

func (uc *taxCodeUseCase) UpdateTaxCode(ctx context.Context, id uuidv7.UUID, rate int, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error) {
	// Get tax code
	tc, err := uc.taxCodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Capture before state
	beforeRate := tc.Rate

	// Update rate
	if err := tc.UpdateRate(rate); err != nil {
		return nil, err
	}

	tc.LastUpdatedBy = updatedBy
	tc.Touch()

	// Save
	if err := uc.taxCodeRepo.Update(ctx, tc); err != nil {
		return nil, err
	}

	// Log audit
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogUpdate(ctx, audit.AuditRecord{
			EntityType:     "tax_code",
			EntityID:       tc.ID,
			Action:         "update_rate",
			OrganizationID: tc.OrganizationID,
			UserID:         updatedBy,
			ChangesBefore:  map[string]interface{}{"rate": beforeRate},
			ChangesAfter:   map[string]interface{}{"rate": rate},
		})
	}

	// Invalidate cache
	uc.taxCodeCache.Invalidate(tc.OrganizationID)

	return tc, nil
}

func (uc *taxCodeUseCase) SetTaxPayableAccount(ctx context.Context, id, accountID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error) {
	// Get tax code
	tc, err := uc.taxCodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Set tax payable account
	if err := tc.SetTaxPayableAccount(accountID); err != nil {
		return nil, err
	}

	tc.LastUpdatedBy = updatedBy
	tc.Touch()

	// Save
	if err := uc.taxCodeRepo.Update(ctx, tc); err != nil {
		return nil, err
	}

	// Invalidate cache
	uc.taxCodeCache.Invalidate(tc.OrganizationID)

	return tc, nil
}

func (uc *taxCodeUseCase) SetTaxReceivableAccount(ctx context.Context, id, accountID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error) {
	// Get tax code
	tc, err := uc.taxCodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Set tax receivable account
	if err := tc.SetTaxReceivableAccount(accountID); err != nil {
		return nil, err
	}

	tc.LastUpdatedBy = updatedBy
	tc.Touch()

	// Save
	if err := uc.taxCodeRepo.Update(ctx, tc); err != nil {
		return nil, err
	}

	// Invalidate cache
	uc.taxCodeCache.Invalidate(tc.OrganizationID)

	return tc, nil
}

func (uc *taxCodeUseCase) ActivateTaxCode(ctx context.Context, id, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error) {
	// Get tax code
	tc, err := uc.taxCodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Activate
	tc.Activate()
	tc.LastUpdatedBy = updatedBy
	tc.Touch()

	// Save
	if err := uc.taxCodeRepo.Update(ctx, tc); err != nil {
		return nil, err
	}

	// Invalidate cache
	uc.taxCodeCache.Invalidate(tc.OrganizationID)

	return tc, nil
}

func (uc *taxCodeUseCase) DeactivateTaxCode(ctx context.Context, id, updatedBy uuidv7.UUID) (*aggregate.TaxCode, error) {
	// Get tax code
	tc, err := uc.taxCodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Deactivate
	tc.Deactivate()
	tc.LastUpdatedBy = updatedBy
	tc.Touch()

	// Save
	if err := uc.taxCodeRepo.Update(ctx, tc); err != nil {
		return nil, err
	}

	// Invalidate cache
	uc.taxCodeCache.Invalidate(tc.OrganizationID)

	return tc, nil
}

func (uc *taxCodeUseCase) DeleteTaxCode(ctx context.Context, id uuidv7.UUID) error {
	return uc.taxCodeRepo.Delete(ctx, id)
}

func (uc *taxCodeUseCase) ListTaxCodesByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.TaxCode, error) {
	return uc.taxCodeRepo.ListByOrganization(ctx, organizationID, limit, offset)
}

func (uc *taxCodeUseCase) ListTaxCodesByType(ctx context.Context, organizationID uuidv7.UUID, taxType aggregate.TaxType) ([]*aggregate.TaxCode, error) {
	return uc.taxCodeRepo.ListByType(ctx, organizationID, taxType)
}

func (uc *taxCodeUseCase) ListActiveTaxCodes(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.TaxCode, error) {
	// Try cache first
	taxCodes, err := uc.taxCodeCache.GetAllActive(ctx, organizationID)
	if err == nil && len(taxCodes) > 0 {
		return taxCodes, nil
	}

	// Fallback to repository
	return uc.taxCodeRepo.ListActive(ctx, organizationID)
}
