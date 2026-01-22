package usecase

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IReconciliationUseCase defines the interface for bank reconciliation business logic
type IReconciliationUseCase interface {
	CreateReconciliation(ctx context.Context, organizationID, bankAccountID, accountID uuidv7.UUID, reconciliationDate, statementDate time.Time, bankStatementBalanceCents, bookBalanceCents int64, currencyCode string, createdBy uuidv7.UUID) (*aggregate.Reconciliation, error)
	AddReconciliationItem(ctx context.Context, reconciliationID uuidv7.UUID, transactionType reconciliation.TransactionType, transactionID *uuidv7.UUID, transactionDate time.Time, description string, amountCents int64, notes string, updatedBy uuidv7.UUID) (*aggregate.Reconciliation, error)
	RemoveReconciliationItem(ctx context.Context, reconciliationID, itemID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.Reconciliation, error)
	MarkItemMatched(ctx context.Context, reconciliationID, itemID uuidv7.UUID, updatedBy uuidv7.UUID) (*aggregate.Reconciliation, error)
	CompleteReconciliation(ctx context.Context, reconciliationID, reconciledBy uuidv7.UUID) (*aggregate.Reconciliation, error)
	ReopenReconciliation(ctx context.Context, reconciliationID, updatedBy uuidv7.UUID) (*aggregate.Reconciliation, error)
	GetReconciliationByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error)
	DeleteReconciliation(ctx context.Context, id uuidv7.UUID) error
	ListReconciliationsByBankAccount(ctx context.Context, bankAccountID uuidv7.UUID) ([]*aggregate.Reconciliation, error)
	ListReconciliationsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.Reconciliation, error)
	ListReconciliationsByStatus(ctx context.Context, organizationID uuidv7.UUID, status reconciliation.Status) ([]*aggregate.Reconciliation, error)
}

type reconciliationUseCase struct {
	reconciliationRepo repository.IReconciliationRepository
	auditLogger        *audit.AuditLogger
}

// NewReconciliationUseCase creates a new bank reconciliation use case
func NewReconciliationUseCase(
	reconciliationRepo repository.IReconciliationRepository,
	auditLogger *audit.AuditLogger,
) IReconciliationUseCase {
	return &reconciliationUseCase{
		reconciliationRepo: reconciliationRepo,
		auditLogger:        auditLogger,
	}
}

func (uc *reconciliationUseCase) CreateReconciliation(
	ctx context.Context,
	organizationID, bankAccountID, accountID uuidv7.UUID,
	reconciliationDate, statementDate time.Time,
	bankStatementBalanceCents, bookBalanceCents int64,
	currencyCode string,
	createdBy uuidv7.UUID,
) (*aggregate.Reconciliation, error) {
	// Create new bank reconciliation
	recon, err := aggregate.NewReconciliation(organizationID, bankAccountID, accountID, reconciliationDate, statementDate, bankStatementBalanceCents, bookBalanceCents, currencyCode, createdBy)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := uc.reconciliationRepo.Create(ctx, recon); err != nil {
		return nil, err
	}

	return recon, nil
}

func (uc *reconciliationUseCase) AddReconciliationItem(
	ctx context.Context,
	reconciliationID uuidv7.UUID,
	transactionType reconciliation.TransactionType,
	transactionID *uuidv7.UUID,
	transactionDate time.Time,
	description string,
	amountCents int64,
	notes string,
	updatedBy uuidv7.UUID,
) (*aggregate.Reconciliation, error) {
	// Get reconciliation
	recon, err := uc.reconciliationRepo.GetByID(ctx, reconciliationID)
	if err != nil {
		return nil, err
	}

	// Verify reconciliation is not completed
	if recon.Status == reconciliation.StatusCompleted {
		return nil, reconciliation.ErrCannotModifyCompleted
	}

	// Add reconciliation item
	if err := recon.AddItem(transactionType, transactionID, transactionDate, description, amountCents, notes); err != nil {
		return nil, err
	}

	recon.LastUpdatedBy = updatedBy
	recon.Touch()

	// Save
	if err := uc.reconciliationRepo.Update(ctx, recon); err != nil {
		return nil, err
	}

	return recon, nil
}

func (uc *reconciliationUseCase) RemoveReconciliationItem(
	ctx context.Context,
	reconciliationID, itemID uuidv7.UUID,
	updatedBy uuidv7.UUID,
) (*aggregate.Reconciliation, error) {
	// Get reconciliation
	recon, err := uc.reconciliationRepo.GetByID(ctx, reconciliationID)
	if err != nil {
		return nil, err
	}

	// Verify reconciliation is not completed
	if recon.Status == reconciliation.StatusCompleted {
		return nil, reconciliation.ErrCannotModifyCompleted
	}

	// Remove reconciliation item
	if err := recon.RemoveItem(itemID); err != nil {
		return nil, err
	}

	recon.LastUpdatedBy = updatedBy
	recon.Touch()

	// Save
	if err := uc.reconciliationRepo.Update(ctx, recon); err != nil {
		return nil, err
	}

	return recon, nil
}

func (uc *reconciliationUseCase) MarkItemMatched(
	ctx context.Context,
	reconciliationID, itemID uuidv7.UUID,
	updatedBy uuidv7.UUID,
) (*aggregate.Reconciliation, error) {
	// Get reconciliation
	recon, err := uc.reconciliationRepo.GetByID(ctx, reconciliationID)
	if err != nil {
		return nil, err
	}

	// Mark item as matched
	if err := recon.MarkItemMatched(itemID); err != nil {
		return nil, err
	}

	recon.LastUpdatedBy = updatedBy
	recon.Touch()

	// Save
	if err := uc.reconciliationRepo.Update(ctx, recon); err != nil {
		return nil, err
	}

	return recon, nil
}

func (uc *reconciliationUseCase) CompleteReconciliation(
	ctx context.Context,
	reconciliationID, reconciledBy uuidv7.UUID,
) (*aggregate.Reconciliation, error) {
	// Get reconciliation
	recon, err := uc.reconciliationRepo.GetByID(ctx, reconciliationID)
	if err != nil {
		return nil, err
	}

	// Complete reconciliation
	if err := recon.Complete(reconciledBy); err != nil {
		return nil, err
	}

	// Save
	if err := uc.reconciliationRepo.Update(ctx, recon); err != nil {
		return nil, err
	}

	return recon, nil
}

func (uc *reconciliationUseCase) ReopenReconciliation(
	ctx context.Context,
	reconciliationID, reopenedBy uuidv7.UUID,
) (*aggregate.Reconciliation, error) {
	// Get bank reconciliation
	recon, err := uc.reconciliationRepo.GetByID(ctx, reconciliationID)
	if err != nil {
		return nil, err
	}

	// Reopen reconciliation
	if err := recon.Reopen(); err != nil {
		return nil, err
	}

	recon.LastUpdatedBy = reopenedBy
	recon.Touch()

	// Save
	if err := uc.reconciliationRepo.Update(ctx, recon); err != nil {
		return nil, err
	}

	return recon, nil
}

func (uc *reconciliationUseCase) GetReconciliationByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Reconciliation, error) {
	return uc.reconciliationRepo.GetByID(ctx, id)
}

func (uc *reconciliationUseCase) DeleteReconciliation(ctx context.Context, id uuidv7.UUID) error {
	// Get reconciliation to verify it can be deleted
	recon, err := uc.reconciliationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Can only delete in-progress reconciliations
	if recon.Status != reconciliation.StatusInProgress {
		return reconciliation.ErrCannotModifyCompleted
	}

	return uc.reconciliationRepo.Delete(ctx, id)
}

func (uc *reconciliationUseCase) ListReconciliationsByBankAccount(ctx context.Context, bankAccountID uuidv7.UUID) ([]*aggregate.Reconciliation, error) {
	return uc.reconciliationRepo.ListByBankAccount(ctx, bankAccountID)
}

func (uc *reconciliationUseCase) ListReconciliationsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.Reconciliation, error) {
	return uc.reconciliationRepo.ListByOrganization(ctx, organizationID, limit, offset)
}

func (uc *reconciliationUseCase) ListReconciliationsByStatus(ctx context.Context, organizationID uuidv7.UUID, status reconciliation.Status) ([]*aggregate.Reconciliation, error) {
	return uc.reconciliationRepo.ListByStatus(ctx, organizationID, status)
}
