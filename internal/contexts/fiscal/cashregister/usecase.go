package cashregister

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines the business operations for cash register management
type IUseCase interface {
	// CreateCashRegister creates a new cash register
	CreateCashRegister(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*CashRegister, error)

	// GetCashRegister retrieves cash register by ID
	GetCashRegister(ctx context.Context, id uuidv7.UUID) (*CashRegister, error)

	// GetByFiscalNumber retrieves cash register by fiscal number
	GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*CashRegister, error)

	// ListCashRegisters retrieves all cash registers (optional organization filter)
	ListCashRegisters(ctx context.Context, organizationID *uuidv7.UUID) ([]*CashRegister, error)

	// ActivateCashRegister activates a cash register with license key
	ActivateCashRegister(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*CashRegister, error)

	// DeactivateCashRegister deactivates a cash register
	DeactivateCashRegister(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*CashRegister, error)

	// UpdateCashRegister updates cash register details
	UpdateCashRegister(ctx context.Context, cashRegister *CashRegister) error

	// DeleteCashRegister soft-deletes cash register
	DeleteCashRegister(ctx context.Context, id uuidv7.UUID) error

	// SyncCashRegister updates last sync timestamp
	SyncCashRegister(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*CashRegister, error)
}

// useCase implements IUseCase interface
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new cash register use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateCashRegister creates a new cash register
func (uc *useCase) CreateCashRegister(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*CashRegister, error) {
	// Check if fiscal number already exists
	existing, err := uc.repo.GetByFiscalNumber(ctx, fiscalNumber)
	if err == nil && existing != nil {
		return nil, ErrCashRegisterFiscalNumberExists
	}

	// Create new cash register
	cr, err := NewCashRegister(organizationID, fiscalNumber, model, createdBy)
	if err != nil {
		return nil, err
	}

	// Persist to repository
	if err := uc.repo.Create(ctx, cr); err != nil {
		return nil, ErrCashRegisterCreateFailed
	}

	return cr, nil
}

// GetCashRegister retrieves cash register by ID
func (uc *useCase) GetCashRegister(ctx context.Context, id uuidv7.UUID) (*CashRegister, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetByFiscalNumber retrieves cash register by fiscal number
func (uc *useCase) GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*CashRegister, error) {
	return uc.repo.GetByFiscalNumber(ctx, fiscalNumber)
}

// ListCashRegisters retrieves all cash registers with optional organization filter
func (uc *useCase) ListCashRegisters(ctx context.Context, organizationID *uuidv7.UUID) ([]*CashRegister, error) {
	filters := &ListFilters{
		OrganizationID: organizationID,
	}
	return uc.repo.List(ctx, filters)
}

// ActivateCashRegister activates a cash register
func (uc *useCase) ActivateCashRegister(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*CashRegister, error) {
	// Get cash register
	cr, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Activate
	if err := cr.Activate(licenseKey, activatedBy); err != nil {
		return nil, err
	}

	// Update repository
	if err := uc.repo.Update(ctx, cr); err != nil {
		return nil, ErrCashRegisterUpdateFailed
	}

	return cr, nil
}

// DeactivateCashRegister deactivates a cash register
func (uc *useCase) DeactivateCashRegister(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*CashRegister, error) {
	// Get cash register
	cr, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Deactivate
	if err := cr.Deactivate(deactivatedBy); err != nil {
		return nil, err
	}

	// Update repository
	if err := uc.repo.Update(ctx, cr); err != nil {
		return nil, ErrCashRegisterUpdateFailed
	}

	return cr, nil
}

// UpdateCashRegister updates cash register details
func (uc *useCase) UpdateCashRegister(ctx context.Context, cashRegister *CashRegister) error {
	// Validate
	if err := cashRegister.Validate(); err != nil {
		return err
	}

	// Update timestamp
	cashRegister.Touch()

	// Update repository
	if err := uc.repo.Update(ctx, cashRegister); err != nil {
		return ErrCashRegisterUpdateFailed
	}

	return nil
}

// DeleteCashRegister soft-deletes cash register
func (uc *useCase) DeleteCashRegister(ctx context.Context, id uuidv7.UUID) error {
	// Check if exists
	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete
	if err := uc.repo.Delete(ctx, id); err != nil {
		return ErrCashRegisterDeleteFailed
	}

	return nil
}

// SyncCashRegister updates last sync timestamp
func (uc *useCase) SyncCashRegister(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*CashRegister, error) {
	// Get cash register
	cr, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update sync
	if err := cr.UpdateLastSync(syncedBy); err != nil {
		return nil, err
	}

	// Update repository
	if err := uc.repo.Update(ctx, cr); err != nil {
		return nil, ErrCashRegisterUpdateFailed
	}

	return cr, nil
}
