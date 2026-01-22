package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	customererrors "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/aggregate"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICustomerUseCase defines business operations for customer management
type ICustomerUseCase interface {
	// CreateCustomer creates a new B2C customer (Lead status)
	CreateCustomer(ctx context.Context, name, email, source string, assignedTo uuidv7.UUID) (*aggregate.Customer, error)

	// CreateB2BCustomer creates a new B2B customer linked to a company
	CreateB2BCustomer(ctx context.Context, name, email, source string, companyID, assignedTo uuidv7.UUID) (*aggregate.Customer, error)

	// GetCustomer retrieves a customer by ID
	GetCustomer(ctx context.Context, id uuidv7.UUID) (*aggregate.Customer, error)

	// GetCustomerByEmail retrieves a customer by email
	GetCustomerByEmail(ctx context.Context, email string) (*aggregate.Customer, error)

	// UpdateCustomer updates customer information
	UpdateCustomer(ctx context.Context, customer *aggregate.Customer) error

	// SetCustomerPhone sets or clears customer phone number
	SetCustomerPhone(ctx context.Context, customerID uuidv7.UUID, phone string) error

	// QualifyAsProspect transitions customer from Lead to Prospect
	QualifyAsProspect(ctx context.Context, customerID uuidv7.UUID) error

	// ConvertToCustomer transitions prospect to paying customer
	ConvertToCustomer(ctx context.Context, customerID uuidv7.UUID) error

	// ChurnCustomer marks customer as churned (lost)
	ChurnCustomer(ctx context.Context, customerID uuidv7.UUID, reason string) error

	// ReactivateCustomer brings churned customer back
	ReactivateCustomer(ctx context.Context, customerID uuidv7.UUID) error

	// UpgradeCustomerTier increases subscription tier
	UpgradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier aggregate.CustomerTier) error

	// DowngradeCustomerTier decreases subscription tier
	DowngradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier aggregate.CustomerTier) error

	// ReassignCustomer changes assigned sales rep
	ReassignCustomer(ctx context.Context, customerID, newRepID uuidv7.UUID) error

	// LinkCustomerToUser links customer to Identity.User account
	LinkCustomerToUser(ctx context.Context, customerID, userID uuidv7.UUID) error

	// AddTagToCustomer adds a segmentation tag
	AddTagToCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error

	// RemoveTagFromCustomer removes a tag
	RemoveTagFromCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error

	// ListCustomersByAssignedTo retrieves customers for a sales rep
	ListCustomersByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*aggregate.Customer, int, error)

	// ListCustomersByStatus retrieves customers by status
	ListCustomersByStatus(ctx context.Context, status aggregate.CustomerStatus, limit, offset int) ([]*aggregate.Customer, int, error)

	// ListCustomersByTier retrieves customers by tier
	ListCustomersByTier(ctx context.Context, tier aggregate.CustomerTier, limit, offset int) ([]*aggregate.Customer, int, error)

	// ListCustomers retrieves all customers with pagination
	ListCustomers(ctx context.Context, limit, offset int) ([]*aggregate.Customer, int, error)

	// GetCustomerStats retrieves statistics by status and tier
	GetCustomerStats(ctx context.Context) (*CustomerStats, error)

	// DeleteCustomer soft-deletes a customer
	DeleteCustomer(ctx context.Context, id uuidv7.UUID) error
}

// CustomerStats represents customer statistics
type CustomerStats struct {
	TotalCustomers   int
	ByStatus         map[aggregate.CustomerStatus]int
	ByTier           map[aggregate.CustomerTier]int
	ActiveCustomers  int // Status = customer
	ChurnedCustomers int // Status = churned
}

// CustomerUseCase implements ICustomerUseCase
type CustomerUseCase struct {
	repo repository.ICustomerRepository
}

// NewCustomerUseCase creates a new customer use case
func NewCustomerUseCase(repo repository.ICustomerRepository) ICustomerUseCase {
	return &CustomerUseCase{
		repo: repo,
	}
}

// CreateCustomer creates a new B2C customer
func (uc *CustomerUseCase) CreateCustomer(ctx context.Context, name, email, source string, assignedTo uuidv7.UUID) (*aggregate.Customer, error) {
	// Check if email already exists
	exists, err := uc.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check customer email existence %s: %w", email, err)
	}
	if exists {
		return nil, customererrors.ErrCustomerAlreadyExists
	}

	// Create customer entity
	customer, err := aggregate.NewCustomer(name, email, source, assignedTo)
	if err != nil {
		return nil, err
	}

	// Persist to repository
	if err := uc.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("create customer %s: %w", customer.Email, err)
	}

	return customer, nil
}

// CreateB2BCustomer creates a new B2B customer linked to a company
func (uc *CustomerUseCase) CreateB2BCustomer(ctx context.Context, name, email, source string, companyID, assignedTo uuidv7.UUID) (*aggregate.Customer, error) {
	// Check if email already exists
	exists, err := uc.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check customer email existence %s: %w", email, err)
	}
	if exists {
		return nil, customererrors.ErrCustomerAlreadyExists
	}

	// Create B2B customer entity
	customer, err := aggregate.NewB2BCustomer(name, email, source, companyID, assignedTo)
	if err != nil {
		return nil, err
	}

	// Persist to repository
	if err := uc.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("create customer %s: %w", customer.Email, err)
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (uc *CustomerUseCase) GetCustomer(ctx context.Context, id uuidv7.UUID) (*aggregate.Customer, error) {
	customer, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customererrors.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("get customer %s: %w", id, err)
	}

	return customer, nil
}

// GetCustomerByEmail retrieves a customer by email
func (uc *CustomerUseCase) GetCustomerByEmail(ctx context.Context, email string) (*aggregate.Customer, error) {
	customer, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customererrors.ErrCustomerNotFound
		}
		return nil, fmt.Errorf("get customer %s: %w", email, err)
	}

	return customer, nil
}

// UpdateCustomer updates customer information
func (uc *CustomerUseCase) UpdateCustomer(ctx context.Context, customer *aggregate.Customer) error {
	// Validate customer
	if err := customer.Validate(); err != nil {
		return err
	}

	// Update in repository
	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// SetCustomerPhone sets or clears customer phone number
func (uc *CustomerUseCase) SetCustomerPhone(ctx context.Context, customerID uuidv7.UUID, phone string) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.SetPhone(phone); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// QualifyAsProspect transitions customer from Lead to Prospect
func (uc *CustomerUseCase) QualifyAsProspect(ctx context.Context, customerID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.QualifyAsProspect(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// ConvertToCustomer transitions prospect to paying customer
func (uc *CustomerUseCase) ConvertToCustomer(ctx context.Context, customerID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.ConvertToCustomer(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// ChurnCustomer marks customer as churned (lost)
func (uc *CustomerUseCase) ChurnCustomer(ctx context.Context, customerID uuidv7.UUID, reason string) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.Churn(reason); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// ReactivateCustomer brings churned customer back
func (uc *CustomerUseCase) ReactivateCustomer(ctx context.Context, customerID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.Reactivate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// UpgradeCustomerTier increases subscription tier
func (uc *CustomerUseCase) UpgradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier aggregate.CustomerTier) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.UpgradeTier(newTier); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// DowngradeCustomerTier decreases subscription tier
func (uc *CustomerUseCase) DowngradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier aggregate.CustomerTier) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.DowngradeTier(newTier); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// ReassignCustomer changes assigned sales rep
func (uc *CustomerUseCase) ReassignCustomer(ctx context.Context, customerID, newRepID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.Reassign(newRepID); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// LinkCustomerToUser links customer to Identity.User account
func (uc *CustomerUseCase) LinkCustomerToUser(ctx context.Context, customerID, userID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.LinkToUser(userID); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// AddTagToCustomer adds a segmentation tag
func (uc *CustomerUseCase) AddTagToCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.AddTag(tag); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// RemoveTagFromCustomer removes a tag
func (uc *CustomerUseCase) RemoveTagFromCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return customererrors.ErrCustomerNotFound
		}
		return fmt.Errorf("get customer %s: %w", customerID, err)
	}

	if err := customer.RemoveTag(tag); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("update customer %s: %w", customer.ID, err)
	}

	return nil
}

// ListCustomersByAssignedTo retrieves customers for a sales rep
func (uc *CustomerUseCase) ListCustomersByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*aggregate.Customer, int, error) {
	customers, total, err := uc.repo.ListByAssignedTo(ctx, repID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers by assigned to %s: %w", repID, err)
	}

	return customers, total, nil
}

// ListCustomersByStatus retrieves customers by status
func (uc *CustomerUseCase) ListCustomersByStatus(ctx context.Context, status aggregate.CustomerStatus, limit, offset int) ([]*aggregate.Customer, int, error) {
	customers, total, err := uc.repo.ListByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers by status %s: %w", status, err)
	}

	return customers, total, nil
}

// ListCustomersByTier retrieves customers by tier
func (uc *CustomerUseCase) ListCustomersByTier(ctx context.Context, tier aggregate.CustomerTier, limit, offset int) ([]*aggregate.Customer, int, error) {
	customers, total, err := uc.repo.ListByTier(ctx, tier, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers by tier %s: %w", tier, err)
	}

	return customers, total, nil
}

// ListCustomers retrieves all customers with pagination
func (uc *CustomerUseCase) ListCustomers(ctx context.Context, limit, offset int) ([]*aggregate.Customer, int, error) {
	customers, total, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}

	return customers, total, nil
}

// GetCustomerStats retrieves statistics by status and tier (optimized with GROUP BY)
func (uc *CustomerUseCase) GetCustomerStats(ctx context.Context) (*CustomerStats, error) {
	stats := &CustomerStats{
		ByStatus: make(map[aggregate.CustomerStatus]int),
		ByTier:   make(map[aggregate.CustomerTier]int),
	}

	// Get all status counts in ONE query (GROUP BY optimization)
	statusCounts, err := uc.repo.CountByAllStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("get customer stats by status: %w", err)
	}

	// Populate stats from bulk query results
	for status, count := range statusCounts {
		stats.ByStatus[status] = count
		stats.TotalCustomers += count

		switch status {
		case aggregate.CustomerStatusCustomer:
			stats.ActiveCustomers = count
		case aggregate.CustomerStatusChurned:
			stats.ChurnedCustomers = count
		}
	}

	// Get all tier counts in ONE query (GROUP BY optimization)
	tierCounts, err := uc.repo.CountByAllTiers(ctx)
	if err != nil {
		return nil, fmt.Errorf("get customer stats by tier: %w", err)
	}

	// Populate tier stats
	stats.ByTier = tierCounts

	return stats, nil
}

// DeleteCustomer soft-deletes a customer
func (uc *CustomerUseCase) DeleteCustomer(ctx context.Context, id uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete customer %s: %w", id, err)
	}

	return nil
}
