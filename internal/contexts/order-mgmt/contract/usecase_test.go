package contract

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockContractRepository is a mock implementation of IContractRepository for testing
type MockContractRepository struct {
	CreateFunc                func(ctx context.Context, c *Contract) error
	GetByIDFunc               func(ctx context.Context, id uuidv7.UUID) (*Contract, error)
	UpdateFunc                func(ctx context.Context, c *Contract) error
	DeleteFunc                func(ctx context.Context, id uuidv7.UUID) error
	ListFunc                  func(ctx context.Context, offset, limit int) ([]*Contract, int, error)
	GetByOrderFunc            func(ctx context.Context, orderID uuidv7.UUID) ([]*Contract, error)
	GetByCustomerFunc         func(ctx context.Context, customerID uuidv7.UUID) ([]*Contract, error)
	GetActiveContractsFunc    func(ctx context.Context) ([]*Contract, error)
	ListByCustomerFunc        func(ctx context.Context, customerID uuidv7.UUID, limit, offset int) ([]*Contract, int, error)
	ListByStatusFunc          func(ctx context.Context, status ContractStatus, limit, offset int) ([]*Contract, int, error)
	ListExpiringSoonFunc      func(ctx context.Context, days int) ([]*Contract, error)
	CountByCustomerFunc       func(ctx context.Context, customerID uuidv7.UUID) (int, error)
	CountByStatusFunc         func(ctx context.Context, status ContractStatus) (int, error)
	ExistsFunc                func(ctx context.Context, id uuidv7.UUID) (bool, error)
}

// Implement IContractRepository interface with nil checks
func (m *MockContractRepository) Create(ctx context.Context, c *Contract) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, c)
	}
	return errors.New("CreateFunc not implemented")
}

func (m *MockContractRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockContractRepository) Update(ctx context.Context, c *Contract) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, c)
	}
	return errors.New("UpdateFunc not implemented")
}

func (m *MockContractRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

func (m *MockContractRepository) List(ctx context.Context, offset, limit int) ([]*Contract, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, offset, limit)
	}
	return nil, 0, errors.New("ListFunc not implemented")
}

func (m *MockContractRepository) GetByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*Contract, error) {
	if m.GetByOrderFunc != nil {
		return m.GetByOrderFunc(ctx, orderID)
	}
	return nil, errors.New("GetByOrderFunc not implemented")
}

func (m *MockContractRepository) GetByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*Contract, error) {
	if m.GetByCustomerFunc != nil {
		return m.GetByCustomerFunc(ctx, customerID)
	}
	return nil, errors.New("GetByCustomerFunc not implemented")
}

func (m *MockContractRepository) GetActiveContracts(ctx context.Context) ([]*Contract, error) {
	if m.GetActiveContractsFunc != nil {
		return m.GetActiveContractsFunc(ctx)
	}
	return nil, errors.New("GetActiveContractsFunc not implemented")
}

func (m *MockContractRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, limit, offset int) ([]*Contract, int, error) {
	if m.ListByCustomerFunc != nil {
		return m.ListByCustomerFunc(ctx, customerID, limit, offset)
	}
	return nil, 0, errors.New("ListByCustomerFunc not implemented")
}

func (m *MockContractRepository) ListByStatus(ctx context.Context, status ContractStatus, limit, offset int) ([]*Contract, int, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, status, limit, offset)
	}
	return nil, 0, errors.New("ListByStatusFunc not implemented")
}

func (m *MockContractRepository) ListExpiringSoon(ctx context.Context, days int) ([]*Contract, error) {
	if m.ListExpiringSoonFunc != nil {
		return m.ListExpiringSoonFunc(ctx, days)
	}
	return nil, errors.New("ListExpiringSoonFunc not implemented")
}

func (m *MockContractRepository) CountByCustomer(ctx context.Context, customerID uuidv7.UUID) (int, error) {
	if m.CountByCustomerFunc != nil {
		return m.CountByCustomerFunc(ctx, customerID)
	}
	return 0, errors.New("CountByCustomerFunc not implemented")
}

func (m *MockContractRepository) CountByStatus(ctx context.Context, status ContractStatus) (int, error) {
	if m.CountByStatusFunc != nil {
		return m.CountByStatusFunc(ctx, status)
	}
	return 0, errors.New("CountByStatusFunc not implemented")
}

func (m *MockContractRepository) Exists(ctx context.Context, id uuidv7.UUID) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, id)
	}
	return false, errors.New("ExistsFunc not implemented")
}

// Helper function to create a test contract
func fakeContract() *Contract {
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	return NewContract(orderID, customerID, "Test Terms")
}

// Test CreateContract
func TestCreateContract_Success(t *testing.T) {
	mockRepo := &MockContractRepository{
		CreateFunc: func(ctx context.Context, c *Contract) error {
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	terms := "Service Agreement Terms"

	result, err := uc.CreateContract(ctx, orderID, customerID, terms)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, orderID, result.OrderID)
	assert.Equal(t, customerID, result.CustomerID)
	assert.Equal(t, terms, result.Terms)
	assert.Equal(t, ContractStatusDraft, result.Status)
}

func TestCreateContract_RepositoryError(t *testing.T) {
	mockRepo := &MockContractRepository{
		CreateFunc: func(ctx context.Context, c *Contract) error {
			return errors.New("database error")
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	terms := "Service Agreement Terms"

	result, err := uc.CreateContract(ctx, orderID, customerID, terms)

	require.Error(t, err)
	assert.Nil(t, result)
	// Repository error propagated directly (no wrapper)
}

// Test GetContract
func TestGetContract_Success(t *testing.T) {
	expectedContract := fakeContract()

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return expectedContract, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	result, err := uc.GetContract(ctx, expectedContract.ID)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedContract.ID, result.ID)
}

func TestGetContract_NotFound(t *testing.T) {
	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return nil, errors.New("contract not found")
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()
	contractID := uuidv7.New()

	result, err := uc.GetContract(ctx, contractID)

	require.Error(t, err)
	assert.Nil(t, result)
	// Repository error propagated directly (no wrapper)
}

// Test SubmitForSignature
func TestSubmitForSignature_Success(t *testing.T) {
	testContract := fakeContract()

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
		UpdateFunc: func(ctx context.Context, c *Contract) error {
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.SubmitForSignature(ctx, testContract.ID)

	require.NoError(t, err)
	assert.Equal(t, ContractStatusPendingSignature, testContract.Status)
}

func TestSubmitForSignature_InvalidState(t *testing.T) {
	testContract := fakeContract()
	_ = testContract.SubmitForSignature()
	_ = testContract.Sign("John Doe", "john@example.com", "sig123")

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.SubmitForSignature(ctx, testContract.ID)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCannotSubmitNonDraft))
}

// Test SignContract
func TestSignContract_Success(t *testing.T) {
	testContract := fakeContract()
	_ = testContract.SubmitForSignature()

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
		UpdateFunc: func(ctx context.Context, c *Contract) error {
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.SignContract(ctx, testContract.ID, "Jane Smith", "jane@example.com", "sig456")

	require.NoError(t, err)
	assert.Equal(t, ContractStatusActive, testContract.Status)
	assert.Equal(t, "Jane Smith", testContract.SignedByName)
	assert.Equal(t, "jane@example.com", testContract.SignedByEmail)
	assert.Equal(t, "sig456", testContract.SignatureID)
}

func TestSignContract_InvalidState(t *testing.T) {
	testContract := fakeContract() // Draft state

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.SignContract(ctx, testContract.ID, "Jane Smith", "jane@example.com", "sig456")

	require.Error(t, err)
	// Repository error propagated directly (no wrapper)
}

// Test CompleteContract
func TestCompleteContract_Success(t *testing.T) {
	testContract := fakeContract()
	_ = testContract.SubmitForSignature()
	_ = testContract.Sign("John Doe", "john@example.com", "sig123")

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
		UpdateFunc: func(ctx context.Context, c *Contract) error {
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.CompleteContract(ctx, testContract.ID)

	require.NoError(t, err)
	assert.Equal(t, ContractStatusCompleted, testContract.Status)
}

// Test TerminateContract
func TestTerminateContract_Success(t *testing.T) {
	testContract := fakeContract()
	_ = testContract.SubmitForSignature()
	_ = testContract.Sign("John Doe", "john@example.com", "sig123")

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
		UpdateFunc: func(ctx context.Context, c *Contract) error {
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()
	reason := "Customer request"

	err := uc.TerminateContract(ctx, testContract.ID, reason)

	require.NoError(t, err)
	assert.Equal(t, ContractStatusTerminated, testContract.Status)
	assert.Equal(t, reason, testContract.TerminationReason)
}

func TestTerminateContract_InvalidState(t *testing.T) {
	testContract := fakeContract() // Draft state

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.TerminateContract(ctx, testContract.ID, "Customer request")

	require.Error(t, err)
	// Repository error propagated directly (no wrapper)
}

// Test RenewContract
func TestRenewContract_Success(t *testing.T) {
	testContract := fakeContract()
	_ = testContract.SubmitForSignature()
	_ = testContract.Sign("John Doe", "john@example.com", "sig123")
	_ = testContract.Complete()

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
		UpdateFunc: func(ctx context.Context, c *Contract) error {
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.RenewContract(ctx, testContract.ID)

	require.NoError(t, err)
	assert.Equal(t, 2, testContract.Version)
}

// Test SetExpirationDate
func TestSetExpirationDate_Success(t *testing.T) {
	testContract := fakeContract()
	expirationDate := time.Now().Add(365 * 24 * time.Hour)

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
		UpdateFunc: func(ctx context.Context, c *Contract) error {
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.SetExpirationDate(ctx, testContract.ID, expirationDate)

	require.NoError(t, err)
	assert.NotNil(t, testContract.ExpiresAt)
}

func TestSetExpirationDate_InvalidDate(t *testing.T) {
	testContract := fakeContract()
	pastDate := time.Now().Add(-24 * time.Hour)

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.SetExpirationDate(ctx, testContract.ID, pastDate)

	require.Error(t, err)
	// Repository error propagated directly (no wrapper)
}

// Test ListContractsByOrder
func TestListContractsByOrder_Success(t *testing.T) {
	orderID := uuidv7.New()
	expectedContracts := []*Contract{fakeContract(), fakeContract()}

	mockRepo := &MockContractRepository{
		GetByOrderFunc: func(ctx context.Context, oID uuidv7.UUID) ([]*Contract, error) {
			assert.Equal(t, orderID, oID)
			return expectedContracts, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	result, err := uc.ListContractsByOrder(ctx, orderID)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// Test ListContractsByCustomer
func TestListContractsByCustomer_Success(t *testing.T) {
	customerID := uuidv7.New()
	expectedContracts := []*Contract{fakeContract()}
	expectedTotal := 1

	mockRepo := &MockContractRepository{
		ListByCustomerFunc: func(ctx context.Context, cID uuidv7.UUID, limit, offset int) ([]*Contract, int, error) {
			assert.Equal(t, customerID, cID)
			return expectedContracts, expectedTotal, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	result, total, err := uc.ListContractsByCustomer(ctx, customerID, 1, 10)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedTotal, total)
}

// Test ListContractsByStatus
func TestListContractsByStatus_Success(t *testing.T) {
	expectedContracts := []*Contract{fakeContract()}
	expectedTotal := 1

	mockRepo := &MockContractRepository{
		ListByStatusFunc: func(ctx context.Context, status ContractStatus, limit, offset int) ([]*Contract, int, error) {
			assert.Equal(t, ContractStatusActive, status)
			return expectedContracts, expectedTotal, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	result, total, err := uc.ListContractsByStatus(ctx, ContractStatusActive, 1, 10)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedTotal, total)
}

// Test ListExpiringSoon
func TestListExpiringSoon_Success(t *testing.T) {
	days := 30
	expectedContracts := []*Contract{fakeContract()}

	mockRepo := &MockContractRepository{
		ListExpiringSoonFunc: func(ctx context.Context, d int) ([]*Contract, error) {
			assert.Equal(t, days, d)
			return expectedContracts, nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	result, err := uc.ListExpiringSoon(ctx, days)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

// Test UpdateContract
func TestUpdateContract_Success(t *testing.T) {
	testContract := fakeContract()
	newTerms := "Updated Terms"

	mockRepo := &MockContractRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Contract, error) {
			return testContract, nil
		},
		UpdateFunc: func(ctx context.Context, c *Contract) error {
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.UpdateContract(ctx, testContract.ID, newTerms)

	require.NoError(t, err)
	assert.Equal(t, newTerms, testContract.Terms)
}

// Test DeleteContract
func TestDeleteContract_Success(t *testing.T) {
	contractID := uuidv7.New()

	mockRepo := &MockContractRepository{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			assert.Equal(t, contractID, id)
			return nil
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.DeleteContract(ctx, contractID)

	require.NoError(t, err)
}

func TestDeleteContract_Error(t *testing.T) {
	contractID := uuidv7.New()

	mockRepo := &MockContractRepository{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return errors.New("database error")
		},
	}

	uc := NewUseCase(mockRepo)
	ctx := context.Background()

	err := uc.DeleteContract(ctx, contractID)

	require.Error(t, err)
	// Repository error propagated directly (no wrapper)
}
