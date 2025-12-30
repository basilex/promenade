package customer

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// ICustomerRepository defines the interface for customer data access
type ICustomerRepository interface {
	// Create inserts a new customer
	Create(ctx context.Context, customer *Customer) error

	// GetByID retrieves a customer by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*Customer, error)

	// GetByEmail retrieves a customer by email address
	GetByEmail(ctx context.Context, email string) (*Customer, error)

	// GetByUserID retrieves a customer by linked user ID
	GetByUserID(ctx context.Context, userID uuidv7.UUID) (*Customer, error)

	// Update updates an existing customer
	Update(ctx context.Context, customer *Customer) error

	// Delete soft-deletes a customer
	Delete(ctx context.Context, id uuidv7.UUID) error

	// ExistsByEmail checks if a customer with given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ListByAssignedTo retrieves all customers assigned to a sales rep
	ListByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*Customer, int, error)

	// ListByCompanyID retrieves all customers for a company (B2B)
	ListByCompanyID(ctx context.Context, companyID uuidv7.UUID) ([]*Customer, error)

	// ListByStatus retrieves customers by status
	ListByStatus(ctx context.Context, status CustomerStatus, limit, offset int) ([]*Customer, int, error)

	// ListByTier retrieves customers by tier
	ListByTier(ctx context.Context, tier CustomerTier, limit, offset int) ([]*Customer, int, error)

	// List retrieves customers with pagination
	List(ctx context.Context, limit, offset int) ([]*Customer, int, error)

	// CountByStatus counts customers by status
	CountByStatus(ctx context.Context, status CustomerStatus) (int, error)

	// CountByTier counts customers by tier
	CountByTier(ctx context.Context, tier CustomerTier) (int, error)

	// CountByAllStatuses returns counts grouped by all statuses (optimized, single query)
	CountByAllStatuses(ctx context.Context) (map[CustomerStatus]int, error)

	// CountByAllTiers returns counts grouped by all tiers (optimized, single query)
	CountByAllTiers(ctx context.Context) (map[CustomerTier]int, error)
}
