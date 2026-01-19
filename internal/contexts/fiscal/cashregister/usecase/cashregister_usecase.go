package usecase

import (
	"context"

	cashregistererrors "github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICashRegisterUseCase defines the business operations for cash register management
type ICashRegisterUseCase interface {
	// CreateCashRegister creates a new cash register
	CreateCashRegister(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*aggregate.CashRegister, error)

	// GetCashRegister retrieves cash register by ID
	GetCashRegister(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error)

	// GetByFiscalNumber retrieves cash register by fiscal number
	GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error)

	// ListCashRegisters retrieves all cash registers (optional organization filter)
	ListCashRegisters(ctx context.Context, organizationID *uuidv7.UUID) ([]*aggregate.CashRegister, error)

	// ActivateCashRegister activates a cash register with license key
	ActivateCashRegister(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error)

	// DeactivateCashRegister deactivates a cash register
	DeactivateCashRegister(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*aggregate.CashRegister, error)

	// UpdateCashRegister updates cash register details
	UpdateCashRegister(ctx context.Context, cashRegister *aggregate.CashRegister) error

	// DeleteCashRegister soft-deletes cash register
	DeleteCashRegister(ctx context.Context, id uuidv7.UUID) error

	// SyncCashRegister updates last sync timestamp
	SyncCashRegister(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*aggregate.CashRegister, error)
}

// CashRegisterUseCase implements ICashRegisterUseCase interface
type CashRegisterUseCase struct {
	repo repository.ICashRegisterRepository
}

// NewCashRegisterUseCase creates a new cash register use case
func NewCashRegisterUseCase(repo repository.ICashRegisterRepository) ICashRegisterUseCase {
	return &CashRegisterUseCase{
		repo: repo,
	}
}

// CreateCashRegister creates a new cash register
func (u *CashRegisterUseCase) CreateCashRegister(ctx context.Context, organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*aggregate.CashRegister, error) {
	// Check if fiscal number already exists
	existing, err := u.repo.GetByFiscalNumber(ctx, fiscalNumber)
	if err == nil && existing != nil {
		return nil, cashregistererrors.ErrCashRegisterFiscalNumberExists
	}

	// Create new cash register
	cr, err := aggregate.NewCashRegister(organizationID, fiscalNumber, model, createdBy)
	if err != nil {
		return nil, err
	}

	// Persist to repository
	if err := u.repo.Create(ctx, cr); err != nil {
		return nil, cashregistererrors.ErrCashRegisterCreateFailed
	}

	return cr, nil
}

// GetCashRegister retrieves cash register by ID
func (u *CashRegisterUseCase) GetCashRegister(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
	return u.repo.GetByID(ctx, id)
}

// GetByFiscalNumber retrieves cash register by fiscal number
func (u *CashRegisterUseCase) GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error) {
	return u.repo.GetByFiscalNumber(ctx, fiscalNumber)
}

// ListCashRegisters retrieves all cash registers with optional organization filter
func (u *CashRegisterUseCase) ListCashRegisters(ctx context.Context, organizationID *uuidv7.UUID) ([]*aggregate.CashRegister, error) {
	filters := &repository.ListFilters{
		OrganizationID: organizationID,
	}
	return u.repo.List(ctx, filters)
}

// ActivateCashRegister activates a cash register
func (u *CashRegisterUseCase) ActivateCashRegister(ctx context.Context, id uuidv7.UUID, licenseKey string, activatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
	// Get cash register
	cr, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Activate
	if err := cr.Activate(licenseKey, activatedBy); err != nil {
		return nil, err
	}

	// Update repository
	if err := u.repo.Update(ctx, cr); err != nil {
		return nil, cashregistererrors.ErrCashRegisterUpdateFailed
	}

	return cr, nil
}

// DeactivateCashRegister deactivates a cash register
func (u *CashRegisterUseCase) DeactivateCashRegister(ctx context.Context, id uuidv7.UUID, deactivatedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
	// Get cash register
	cr, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Deactivate
	if err := cr.Deactivate(deactivatedBy); err != nil {
		return nil, err
	}

	// Update repository
	if err := u.repo.Update(ctx, cr); err != nil {
		return nil, cashregistererrors.ErrCashRegisterUpdateFailed
	}

	return cr, nil
}

// UpdateCashRegister updates cash register details
func (u *CashRegisterUseCase) UpdateCashRegister(ctx context.Context, cashRegister *aggregate.CashRegister) error {
	// Validate
	if err := cashRegister.Validate(); err != nil {
		return err
	}

	// Update timestamp
	cashRegister.Touch()

	// Update repository
	if err := u.repo.Update(ctx, cashRegister); err != nil {
		return cashregistererrors.ErrCashRegisterUpdateFailed
	}

	return nil
}

// DeleteCashRegister soft-deletes cash register
func (u *CashRegisterUseCase) DeleteCashRegister(ctx context.Context, id uuidv7.UUID) error {
	// Check if exists
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete
	if err := u.repo.Delete(ctx, id); err != nil {
		return cashregistererrors.ErrCashRegisterDeleteFailed
	}

	return nil
}

// SyncCashRegister updates last sync timestamp
func (u *CashRegisterUseCase) SyncCashRegister(ctx context.Context, id uuidv7.UUID, syncedBy uuidv7.UUID) (*aggregate.CashRegister, error) {
	// Get cash register
	cr, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update sync
	if err := cr.UpdateLastSync(syncedBy); err != nil {
		return nil, err
	}

	// Update repository
	if err := u.repo.Update(ctx, cr); err != nil {
		return nil, cashregistererrors.ErrCashRegisterUpdateFailed
	}

	return cr, nil
}
