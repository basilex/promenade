package customer

import (
	"context"
	"errors"
	"testing"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of ICustomerRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, customer *Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Customer), args.Error(1)
}

func (m *MockRepository) GetByEmail(ctx context.Context, email string) (*Customer, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Customer), args.Error(1)
}

func (m *MockRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*Customer, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Customer), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, customer *Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) ListByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*Customer, int, error) {
	args := m.Called(ctx, repID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Customer), args.Int(1), args.Error(2)
}

func (m *MockRepository) ListByCompanyID(ctx context.Context, companyID uuidv7.UUID) ([]*Customer, error) {
	args := m.Called(ctx, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Customer), args.Error(1)
}

func (m *MockRepository) ListByStatus(ctx context.Context, status CustomerStatus, limit, offset int) ([]*Customer, int, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Customer), args.Int(1), args.Error(2)
}

func (m *MockRepository) ListByTier(ctx context.Context, tier CustomerTier, limit, offset int) ([]*Customer, int, error) {
	args := m.Called(ctx, tier, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Customer), args.Int(1), args.Error(2)
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*Customer, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Customer), args.Int(1), args.Error(2)
}

func (m *MockRepository) CountByStatus(ctx context.Context, status CustomerStatus) (int, error) {
	args := m.Called(ctx, status)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) CountByTier(ctx context.Context, tier CustomerTier) (int, error) {
	args := m.Called(ctx, tier)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) CountByAllStatuses(ctx context.Context) (map[CustomerStatus]int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[CustomerStatus]int), args.Error(1)
}

func (m *MockRepository) CountByAllTiers(ctx context.Context) (map[CustomerTier]int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[CustomerTier]int), args.Error(1)
}

// ============================================================================
// CreateCustomer Tests
// ============================================================================

func TestUseCase_CreateCustomer(t *testing.T) {
	ctx := context.Background()
	assignedTo := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("ExistsByEmail", ctx, "john@example.com").Return(false, nil)
		repo.On("Create", ctx, mock.Anything).Return(nil)

		customer, err := uc.CreateCustomer(ctx, "John Doe", "john@example.com", "website", assignedTo)

		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, "John Doe", customer.Name)
		assert.Equal(t, "john@example.com", customer.Email.Value())
		assert.Equal(t, CustomerStatusLead, customer.Status)
		repo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("ExistsByEmail", ctx, "john@example.com").Return(true, nil)

		customer, err := uc.CreateCustomer(ctx, "John Doe", "john@example.com", "website", assignedTo)

		require.Error(t, err)
		assert.Nil(t, customer)
		assert.Contains(t, err.Error(), "already exists")
		repo.AssertExpectations(t)
	})

	t.Run("invalid email format", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("ExistsByEmail", ctx, "invalid-email").Return(false, nil)

		customer, err := uc.CreateCustomer(ctx, "John Doe", "invalid-email", "website", assignedTo)

		require.Error(t, err)
		assert.Nil(t, customer)
		assert.Contains(t, err.Error(), "invalid email")
		repo.AssertNotCalled(t, "Create")
	})
}

// ============================================================================
// CreateB2BCustomer Tests
// ============================================================================

func TestUseCase_CreateB2BCustomer(t *testing.T) {
	ctx := context.Background()
	assignedTo := uuidv7.New()
	companyID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("ExistsByEmail", ctx, "john@company.com").Return(false, nil)
		repo.On("Create", ctx, mock.Anything).Return(nil)

		customer, err := uc.CreateB2BCustomer(ctx, "John Doe", "john@company.com", "referral", companyID, assignedTo)

		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, "John Doe", customer.Name)
		assert.True(t, customer.IsB2B())
		assert.NotNil(t, customer.CompanyID)
		repo.AssertExpectations(t)
	})

	t.Run("email already exists", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("ExistsByEmail", ctx, "john@company.com").Return(true, nil)

		customer, err := uc.CreateB2BCustomer(ctx, "John Doe", "john@company.com", "referral", companyID, assignedTo)

		require.Error(t, err)
		assert.Nil(t, customer)
		repo.AssertExpectations(t)
	})
}

// ============================================================================
// GetCustomer Tests
// ============================================================================

func TestUseCase_GetCustomer(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		expected, _ := NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		expected.ID = customerID

		repo.On("GetByID", ctx, customerID).Return(expected, nil)

		customer, err := uc.GetCustomer(ctx, customerID)

		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, customerID, customer.ID)
		repo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("GetByID", ctx, customerID).Return(nil, errors.New("customer not found"))

		customer, err := uc.GetCustomer(ctx, customerID)

		require.Error(t, err)
		assert.Nil(t, customer)
		repo.AssertExpectations(t)
	})
}

// ============================================================================
// GetCustomerByEmail Tests
// ============================================================================

func TestUseCase_GetCustomerByEmail(t *testing.T) {
	ctx := context.Background()
	assignedTo := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		expected, _ := NewCustomer("John Doe", "john@example.com", "website", assignedTo)

		repo.On("GetByEmail", ctx, "john@example.com").Return(expected, nil)

		customer, err := uc.GetCustomerByEmail(ctx, "john@example.com")

		require.NoError(t, err)
		assert.NotNil(t, customer)
		assert.Equal(t, "john@example.com", customer.Email.Value())
		repo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		repo.On("GetByEmail", ctx, "notfound@example.com").Return(nil, errors.New("customer not found"))

		customer, err := uc.GetCustomerByEmail(ctx, "notfound@example.com")

		require.Error(t, err)
		assert.Nil(t, customer)
		repo.AssertExpectations(t)
	})
}

// ============================================================================
// QualifyAsProspect Tests
// ============================================================================

func TestUseCase_QualifyAsProspect(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()

	t.Run("success - lead to prospect", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		customer, _ := NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		customer.ID = customerID

		repo.On("GetByID", ctx, customerID).Return(customer, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.QualifyAsProspect(ctx, customerID)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("invalid transition - already prospect", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		customer, _ := NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		customer.ID = customerID
		_ = customer.QualifyAsProspect()

		repo.On("GetByID", ctx, customerID).Return(customer, nil)

		err := uc.QualifyAsProspect(ctx, customerID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only qualify leads")
		repo.AssertNotCalled(t, "Update")
	})
}

// ============================================================================
// ConvertToCustomer Tests
// ============================================================================

func TestUseCase_ConvertToCustomer(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()

	t.Run("success - prospect to customer", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		customer, _ := NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		customer.ID = customerID
		_ = customer.QualifyAsProspect()

		repo.On("GetByID", ctx, customerID).Return(customer, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.ConvertToCustomer(ctx, customerID)

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("invalid transition - from lead", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		customer, _ := NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		customer.ID = customerID

		repo.On("GetByID", ctx, customerID).Return(customer, nil)

		err := uc.ConvertToCustomer(ctx, customerID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only convert prospects")
		repo.AssertNotCalled(t, "Update")
	})
}

// ============================================================================
// Churn Tests
// ============================================================================

func TestUseCase_ChurnCustomer(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		customer, _ := NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		customer.ID = customerID
		_ = customer.QualifyAsProspect()
		_ = customer.ConvertToCustomer()

		repo.On("GetByID", ctx, customerID).Return(customer, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.ChurnCustomer(ctx, customerID, "poor service")

		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("invalid transition - from lead", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		customer, _ := NewCustomer("John Doe", "john@example.com", "website", assignedTo)
		customer.ID = customerID

		repo.On("GetByID", ctx, customerID).Return(customer, nil)

		err := uc.ChurnCustomer(ctx, customerID, "poor service")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only churn active customers")
		repo.AssertNotCalled(t, "Update")
	})
}

// ============================================================================
// List Tests
// ============================================================================

func TestUseCase_ListCustomers(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		expected := []*Customer{}
		repo.On("List", ctx, 20, 0).Return(expected, 0, nil)

		customers, total, err := uc.ListCustomers(ctx, 20, 0)

		require.NoError(t, err)
		assert.NotNil(t, customers)
		assert.Equal(t, 0, total)
		repo.AssertExpectations(t)
	})
}

func TestUseCase_ListCustomersByStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		expected := []*Customer{}
		repo.On("ListByStatus", ctx, CustomerStatusLead, 20, 0).Return(expected, 0, nil)

		customers, total, err := uc.ListCustomersByStatus(ctx, CustomerStatusLead, 20, 0)

		require.NoError(t, err)
		assert.NotNil(t, customers)
		assert.Equal(t, 0, total)
		repo.AssertExpectations(t)
	})
}

// ============================================================================
// Stats Tests
// ============================================================================

func TestUseCase_GetCustomerStats(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		uc := NewUseCase(repo)

		// Mock optimized bulk queries (GROUP BY)
		statusCounts := map[CustomerStatus]int{
			CustomerStatusLead:     10,
			CustomerStatusProspect: 5,
			CustomerStatusCustomer: 20,
			CustomerStatusChurned:  3,
		}
		repo.On("CountByAllStatuses", ctx).Return(statusCounts, nil)

		tierCounts := map[CustomerTier]int{
			CustomerTierFree:       15,
			CustomerTierBasic:      10,
			CustomerTierPro:        8,
			CustomerTierEnterprise: 5,
		}
		repo.On("CountByAllTiers", ctx).Return(tierCounts, nil)

		stats, err := uc.GetCustomerStats(ctx)

		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, 38, stats.TotalCustomers)
		assert.Equal(t, 20, stats.ActiveCustomers)
		assert.Equal(t, 3, stats.ChurnedCustomers)
		assert.Equal(t, 10, stats.ByStatus[CustomerStatusLead])
		assert.Equal(t, 15, stats.ByTier[CustomerTierFree])
		repo.AssertExpectations(t)
	})
}
