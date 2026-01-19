package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Provider represents fiscal service provider
type Provider string

const (
	ProviderCheckbox Provider = "checkbox"
)

// ICashRegisterRepository defines cash register data access operations
type ICashRegisterRepository interface {
	// Create creates a new cash register
	Create(ctx context.Context, cr *aggregate.CashRegister) error

	// GetByID retrieves cash register by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error)

	// GetByFiscalNumber retrieves cash register by fiscal number
	GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error)

	// GetByLocation retrieves all cash registers for a location
	GetByLocation(ctx context.Context, locationID uuidv7.UUID) ([]*aggregate.CashRegister, error)

	// ListActive retrieves all active cash registers
	ListActive(ctx context.Context) ([]*aggregate.CashRegister, error)

	// List retrieves cash registers with filters
	List(ctx context.Context, filters *ListFilters) ([]*aggregate.CashRegister, error)

	// Update updates cash register
	Update(ctx context.Context, cr *aggregate.CashRegister) error

	// Delete soft deletes cash register
	Delete(ctx context.Context, id uuidv7.UUID) error
}

// ListFilters defines filters for listing cash registers
type ListFilters struct {
	OrganizationID *uuidv7.UUID
	Provider       *Provider
	IsActive       *bool
}
