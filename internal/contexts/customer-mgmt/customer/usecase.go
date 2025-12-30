package customer

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICustomerUseCase defines business operations for customer management
type ICustomerUseCase interface {
	// CreateCustomer creates a new B2C customer (Lead status)
	CreateCustomer(ctx context.Context, name, email, source string, assignedTo uuidv7.UUID) (*Customer, error)

	// CreateB2BCustomer creates a new B2B customer linked to a company
	CreateB2BCustomer(ctx context.Context, name, email, source string, companyID, assignedTo uuidv7.UUID) (*Customer, error)

	// GetCustomer retrieves a customer by ID
	GetCustomer(ctx context.Context, id uuidv7.UUID) (*Customer, error)

	// GetCustomerByEmail retrieves a customer by email
	GetCustomerByEmail(ctx context.Context, email string) (*Customer, error)

	// UpdateCustomer updates customer information
	UpdateCustomer(ctx context.Context, customer *Customer) error

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
	UpgradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier CustomerTier) error

	// DowngradeCustomerTier decreases subscription tier
	DowngradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier CustomerTier) error

	// ReassignCustomer changes assigned sales rep
	ReassignCustomer(ctx context.Context, customerID, newRepID uuidv7.UUID) error

	// LinkCustomerToUser links customer to Identity.User account
	LinkCustomerToUser(ctx context.Context, customerID, userID uuidv7.UUID) error

	// AddTagToCustomer adds a segmentation tag
	AddTagToCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error

	// RemoveTagFromCustomer removes a tag
	RemoveTagFromCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error

	// ListCustomersByAssignedTo retrieves customers for a sales rep
	ListCustomersByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*Customer, int, error)

	// ListCustomersByStatus retrieves customers by status
	ListCustomersByStatus(ctx context.Context, status CustomerStatus, limit, offset int) ([]*Customer, int, error)

	// ListCustomersByTier retrieves customers by tier
	ListCustomersByTier(ctx context.Context, tier CustomerTier, limit, offset int) ([]*Customer, int, error)

	// ListCustomers retrieves all customers with pagination
	ListCustomers(ctx context.Context, limit, offset int) ([]*Customer, int, error)

	// GetCustomerStats retrieves statistics by status and tier
	GetCustomerStats(ctx context.Context) (*CustomerStats, error)

	// DeleteCustomer soft-deletes a customer
	DeleteCustomer(ctx context.Context, id uuidv7.UUID) error
}

// CustomerStats represents customer statistics
type CustomerStats struct {
	TotalCustomers   int
	ByStatus         map[CustomerStatus]int
	ByTier           map[CustomerTier]int
	ActiveCustomers  int // Status = customer
	ChurnedCustomers int // Status = churned
}

// useCase implements ICustomerUseCase
type useCase struct {
	repo ICustomerRepository
}

// NewUseCase creates a new customer use case
func NewUseCase(repo ICustomerRepository) ICustomerUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateCustomer creates a new B2C customer
func (uc *useCase) CreateCustomer(ctx context.Context, name, email, source string, assignedTo uuidv7.UUID) (*Customer, error) {
	// Check if email already exists
	exists, err := uc.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("customer with email %s already exists", email)
	}

	// Create customer entity
	customer, err := NewCustomer(name, email, source, assignedTo)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Persist to repository
	if err := uc.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to save customer: %w", err)
	}

	return customer, nil
}

// CreateB2BCustomer creates a new B2B customer linked to a company
func (uc *useCase) CreateB2BCustomer(ctx context.Context, name, email, source string, companyID, assignedTo uuidv7.UUID) (*Customer, error) {
	// Check if email already exists
	exists, err := uc.repo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("customer with email %s already exists", email)
	}

	// Create B2B customer entity
	customer, err := NewB2BCustomer(name, email, source, companyID, assignedTo)
	if err != nil {
		return nil, fmt.Errorf("failed to create B2B customer: %w", err)
	}

	// Persist to repository
	if err := uc.repo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to save B2B customer: %w", err)
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (uc *useCase) GetCustomer(ctx context.Context, id uuidv7.UUID) (*Customer, error) {
	customer, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetCustomerByEmail retrieves a customer by email
func (uc *useCase) GetCustomerByEmail(ctx context.Context, email string) (*Customer, error) {
	customer, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return customer, nil
}

// UpdateCustomer updates customer information
func (uc *useCase) UpdateCustomer(ctx context.Context, customer *Customer) error {
	// Validate customer
	if err := customer.Validate(); err != nil {
		return fmt.Errorf("invalid customer: %w", err)
	}

	// Update in repository
	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// SetCustomerPhone sets or clears customer phone number
func (uc *useCase) SetCustomerPhone(ctx context.Context, customerID uuidv7.UUID, phone string) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.SetPhone(phone); err != nil {
		return fmt.Errorf("failed to set phone: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// QualifyAsProspect transitions customer from Lead to Prospect
func (uc *useCase) QualifyAsProspect(ctx context.Context, customerID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.QualifyAsProspect(); err != nil {
		return fmt.Errorf("failed to qualify as prospect: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// ConvertToCustomer transitions prospect to paying customer
func (uc *useCase) ConvertToCustomer(ctx context.Context, customerID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.ConvertToCustomer(); err != nil {
		return fmt.Errorf("failed to convert to customer: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// ChurnCustomer marks customer as churned (lost)
func (uc *useCase) ChurnCustomer(ctx context.Context, customerID uuidv7.UUID, reason string) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.Churn(reason); err != nil {
		return fmt.Errorf("failed to churn customer: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// ReactivateCustomer brings churned customer back
func (uc *useCase) ReactivateCustomer(ctx context.Context, customerID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.Reactivate(); err != nil {
		return fmt.Errorf("failed to reactivate customer: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// UpgradeCustomerTier increases subscription tier
func (uc *useCase) UpgradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier CustomerTier) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.UpgradeTier(newTier); err != nil {
		return fmt.Errorf("failed to upgrade tier: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// DowngradeCustomerTier decreases subscription tier
func (uc *useCase) DowngradeCustomerTier(ctx context.Context, customerID uuidv7.UUID, newTier CustomerTier) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.DowngradeTier(newTier); err != nil {
		return fmt.Errorf("failed to downgrade tier: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// ReassignCustomer changes assigned sales rep
func (uc *useCase) ReassignCustomer(ctx context.Context, customerID, newRepID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.Reassign(newRepID); err != nil {
		return fmt.Errorf("failed to reassign customer: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// LinkCustomerToUser links customer to Identity.User account
func (uc *useCase) LinkCustomerToUser(ctx context.Context, customerID, userID uuidv7.UUID) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.LinkToUser(userID); err != nil {
		return fmt.Errorf("failed to link to user: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// AddTagToCustomer adds a segmentation tag
func (uc *useCase) AddTagToCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.AddTag(tag); err != nil {
		return fmt.Errorf("failed to add tag: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// RemoveTagFromCustomer removes a tag
func (uc *useCase) RemoveTagFromCustomer(ctx context.Context, customerID uuidv7.UUID, tag string) error {
	customer, err := uc.repo.GetByID(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	if err := customer.RemoveTag(tag); err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}

	if err := uc.repo.Update(ctx, customer); err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

// ListCustomersByAssignedTo retrieves customers for a sales rep
func (uc *useCase) ListCustomersByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*Customer, int, error) {
	customers, total, err := uc.repo.ListByAssignedTo(ctx, repID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers by assigned to: %w", err)
	}

	return customers, total, nil
}

// ListCustomersByStatus retrieves customers by status
func (uc *useCase) ListCustomersByStatus(ctx context.Context, status CustomerStatus, limit, offset int) ([]*Customer, int, error) {
	customers, total, err := uc.repo.ListByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers by status: %w", err)
	}

	return customers, total, nil
}

// ListCustomersByTier retrieves customers by tier
func (uc *useCase) ListCustomersByTier(ctx context.Context, tier CustomerTier, limit, offset int) ([]*Customer, int, error) {
	customers, total, err := uc.repo.ListByTier(ctx, tier, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers by tier: %w", err)
	}

	return customers, total, nil
}

// ListCustomers retrieves all customers with pagination
func (uc *useCase) ListCustomers(ctx context.Context, limit, offset int) ([]*Customer, int, error) {
	customers, total, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}

	return customers, total, nil
}

// GetCustomerStats retrieves statistics by status and tier (optimized with GROUP BY)
func (uc *useCase) GetCustomerStats(ctx context.Context) (*CustomerStats, error) {
	stats := &CustomerStats{
		ByStatus: make(map[CustomerStatus]int),
		ByTier:   make(map[CustomerTier]int),
	}

	// Get all status counts in ONE query (GROUP BY optimization)
	statusCounts, err := uc.repo.CountByAllStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count by all statuses: %w", err)
	}

	// Populate stats from bulk query results
	for status, count := range statusCounts {
		stats.ByStatus[status] = count
		stats.TotalCustomers += count

		if status == CustomerStatusCustomer {
			stats.ActiveCustomers = count
		} else if status == CustomerStatusChurned {
			stats.ChurnedCustomers = count
		}
	}

	// Get all tier counts in ONE query (GROUP BY optimization)
	tierCounts, err := uc.repo.CountByAllTiers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count by all tiers: %w", err)
	}

	// Populate tier stats
	stats.ByTier = tierCounts

	return stats, nil
}

// DeleteCustomer soft-deletes a customer
func (uc *useCase) DeleteCustomer(ctx context.Context, id uuidv7.UUID) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}
