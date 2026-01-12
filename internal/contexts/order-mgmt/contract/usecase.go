package contract

import (
	"context"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines the interface for contract business logic operations.
type IUseCase interface {
	// CreateContract creates a new contract for an order
	CreateContract(ctx context.Context, orderID, customerID uuidv7.UUID, terms string) (*Contract, error)

	// GetContract retrieves a contract by ID
	GetContract(ctx context.Context, contractID uuidv7.UUID) (*Contract, error)

	// SubmitForSignature submits a contract for customer signature
	SubmitForSignature(ctx context.Context, contractID uuidv7.UUID) error

	// SignContract marks a contract as signed and activates it
	SignContract(ctx context.Context, contractID uuidv7.UUID, signedByName, signedByEmail, signatureID string) error

	// CompleteContract marks a contract as completed
	CompleteContract(ctx context.Context, contractID uuidv7.UUID) error

	// TerminateContract terminates a contract with a reason
	TerminateContract(ctx context.Context, contractID uuidv7.UUID, reason string) error

	// RenewContract renews a contract (increments version)
	RenewContract(ctx context.Context, contractID uuidv7.UUID) error

	// SetExpirationDate sets or updates the expiration date
	SetExpirationDate(ctx context.Context, contractID uuidv7.UUID, expirationDate time.Time) error

	// ListContractsByOrder retrieves all contracts for an order
	ListContractsByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*Contract, error)

	// ListContractsByCustomer retrieves contracts for a customer with pagination
	ListContractsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Contract, int, error)

	// ListContractsByStatus retrieves contracts by status with pagination
	ListContractsByStatus(ctx context.Context, status ContractStatus, page, pageSize int) ([]*Contract, int, error)

	// GetActiveContracts retrieves all active contracts
	GetActiveContracts(ctx context.Context) ([]*Contract, error)

	// ListExpiringSoon retrieves contracts expiring within specified days
	ListExpiringSoon(ctx context.Context, days int) ([]*Contract, error)

	// UpdateContract updates contract terms
	UpdateContract(ctx context.Context, contractID uuidv7.UUID, terms string) error

	// DeleteContract soft-deletes a contract
	DeleteContract(ctx context.Context, contractID uuidv7.UUID) error
}

// contractUseCase implements IUseCase interface
type contractUseCase struct {
	repo IContractRepository
}

// NewUseCase creates a new contract use case instance
func NewUseCase(repo IContractRepository) IUseCase {
	return &contractUseCase{repo: repo}
}

// CreateContract creates a new contract
func (uc *contractUseCase) CreateContract(
	ctx context.Context,
	orderID uuidv7.UUID,
	customerID uuidv7.UUID,
	terms string,
) (*Contract, error) {
	// Create new contract entity
	c := NewContract(orderID, customerID, terms)

	// Persist to database
	if err := uc.repo.Create(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

// GetContract retrieves a contract by ID
func (uc *contractUseCase) GetContract(ctx context.Context, contractID uuidv7.UUID) (*Contract, error) {
	c, err := uc.repo.GetByID(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// SubmitForSignature submits a contract for customer signature
func (uc *contractUseCase) SubmitForSignature(ctx context.Context, contractID uuidv7.UUID) error {
	// Retrieve contract
	c, err := uc.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Apply business logic
	if err := c.SubmitForSignature(); err != nil {
		return err
	}

	// Persist changes
	if err := uc.repo.Update(ctx, c); err != nil {
		return err
	}

	return nil
}

// SignContract marks a contract as signed and activates it
func (uc *contractUseCase) SignContract(
	ctx context.Context,
	contractID uuidv7.UUID,
	signedByName string,
	signedByEmail string,
	signatureID string,
) error {
	// Retrieve contract
	c, err := uc.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Apply business logic
	if err := c.Sign(signedByName, signedByEmail, signatureID); err != nil {
		return err
	}

	// Persist changes
	if err := uc.repo.Update(ctx, c); err != nil {
		return err
	}

	return nil
}

// CompleteContract marks a contract as completed
func (uc *contractUseCase) CompleteContract(ctx context.Context, contractID uuidv7.UUID) error {
	// Retrieve contract
	c, err := uc.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Apply business logic
	if err := c.Complete(); err != nil {
		return err
	}

	// Persist changes
	if err := uc.repo.Update(ctx, c); err != nil {
		return err
	}

	return nil
}

// TerminateContract terminates a contract with a reason
func (uc *contractUseCase) TerminateContract(
	ctx context.Context,
	contractID uuidv7.UUID,
	reason string,
) error {
	// Retrieve contract
	c, err := uc.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Apply business logic
	if err := c.Terminate(reason); err != nil {
		return err
	}

	// Persist changes
	if err := uc.repo.Update(ctx, c); err != nil {
		return err
	}

	return nil
}

// RenewContract renews a contract (increments version)
func (uc *contractUseCase) RenewContract(
	ctx context.Context,
	contractID uuidv7.UUID,
) error {
	// Retrieve contract
	c, err := uc.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Apply business logic
	if err := c.Renew(); err != nil {
		return err
	}

	// Persist changes
	if err := uc.repo.Update(ctx, c); err != nil {
		return err
	}

	return nil
}

// SetExpirationDate sets or updates the expiration date
func (uc *contractUseCase) SetExpirationDate(
	ctx context.Context,
	contractID uuidv7.UUID,
	expirationDate time.Time,
) error {
	// Retrieve contract
	c, err := uc.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Apply business logic
	if err := c.SetExpirationDate(expirationDate); err != nil {
		return err
	}

	// Persist changes
	if err := uc.repo.Update(ctx, c); err != nil {
		return err
	}

	return nil
}

// ListContractsByOrder retrieves all contracts for a specific order
func (uc *contractUseCase) ListContractsByOrder(
	ctx context.Context,
	orderID uuidv7.UUID,
) ([]*Contract, error) {
	contracts, err := uc.repo.GetByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

// ListContractsByCustomer retrieves contracts for a specific customer with pagination
func (uc *contractUseCase) ListContractsByCustomer(
	ctx context.Context,
	customerID uuidv7.UUID,
	page int,
	pageSize int,
) ([]*Contract, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	// Retrieve contracts
	contracts, total, err := uc.repo.ListByCustomer(ctx, customerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return contracts, total, nil
}

// ListContractsByStatus retrieves contracts with a specific status with pagination
func (uc *contractUseCase) ListContractsByStatus(
	ctx context.Context,
	status ContractStatus,
	page int,
	pageSize int,
) ([]*Contract, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	// Retrieve contracts
	contracts, total, err := uc.repo.ListByStatus(ctx, status, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return contracts, total, nil
}

// GetActiveContracts retrieves all active contracts
func (uc *contractUseCase) GetActiveContracts(ctx context.Context) ([]*Contract, error) {
	contracts, err := uc.repo.GetActiveContracts(ctx)
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

// ListExpiringSoon retrieves contracts expiring within the specified number of days
func (uc *contractUseCase) ListExpiringSoon(ctx context.Context, days int) ([]*Contract, error) {
	// Validate input
	if days <= 0 {
		return nil, ErrInvalidDaysValue
	}

	contracts, err := uc.repo.ListExpiringSoon(ctx, days)
	if err != nil {
		return nil, err
	}

	return contracts, nil
}

// UpdateContract updates contract terms
func (uc *contractUseCase) UpdateContract(
	ctx context.Context,
	contractID uuidv7.UUID,
	terms string,
) error {
	// Retrieve contract
	c, err := uc.repo.GetByID(ctx, contractID)
	if err != nil {
		return err
	}

	// Update fields
	c.Terms = terms
	c.Touch()

	// Persist changes
	if err := uc.repo.Update(ctx, c); err != nil {
		return err
	}

	return nil
}

// DeleteContract deletes a contract
func (uc *contractUseCase) DeleteContract(ctx context.Context, contractID uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, contractID); err != nil {
		return err
	}
	return nil
}

