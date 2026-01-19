package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cashregistererrors "github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Test constants
var testUserID = uuidv7.New()
var testOrgID = uuidv7.New()

// MockRepository implements ICashRegisterRepository for testing
type MockRepository struct {
	CreateFunc            func(ctx context.Context, cr *aggregate.CashRegister) error
	GetByIDFunc           func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error)
	GetByFiscalNumberFunc func(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error)
	GetByLocationFunc     func(ctx context.Context, locationID uuidv7.UUID) ([]*aggregate.CashRegister, error)
	ListFunc              func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.CashRegister, error)
	ListActiveFunc        func(ctx context.Context) ([]*aggregate.CashRegister, error)
	UpdateFunc            func(ctx context.Context, cr *aggregate.CashRegister) error
	DeleteFunc            func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockRepository) Create(ctx context.Context, cr *aggregate.CashRegister) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, cr)
	}
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockRepository) GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error) {
	if m.GetByFiscalNumberFunc != nil {
		return m.GetByFiscalNumberFunc(ctx, fiscalNumber)
	}
	return nil, errors.New("not found")
}

func (m *MockRepository) GetByLocation(ctx context.Context, locationID uuidv7.UUID) ([]*aggregate.CashRegister, error) {
	if m.GetByLocationFunc != nil {
		return m.GetByLocationFunc(ctx, locationID)
	}
	return nil, nil
}

func (m *MockRepository) List(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.CashRegister, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters)
	}
	return nil, nil
}

func (m *MockRepository) ListActive(ctx context.Context) ([]*aggregate.CashRegister, error) {
	if m.ListActiveFunc != nil {
		return m.ListActiveFunc(ctx)
	}
	return nil, nil
}

func (m *MockRepository) Update(ctx context.Context, cr *aggregate.CashRegister) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, cr)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// TestNewUseCase tests the constructor
func TestNewUseCase(t *testing.T) {
	repo := &MockRepository{}
	u := NewCashRegisterUseCase(repo)

	assert.NotNil(t, u)
}

// TestCreateCashRegister_Success tests successful cash register creation
func TestCreateCashRegister_Success(t *testing.T) {
	repo := &MockRepository{
		GetByFiscalNumberFunc: func(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error) {
			return nil, errors.New("not found") // Fiscal number doesn't exist
		},
		CreateFunc: func(ctx context.Context, cr *aggregate.CashRegister) error {
			assert.Equal(t, "FN-123456", cr.FiscalNumber)
			assert.Equal(t, "Model X", cr.Model)
			return nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	cr, err := u.CreateCashRegister(context.Background(), testOrgID, "FN-123456", "Model X", testUserID)

	require.NoError(t, err)
	assert.NotNil(t, cr)
	assert.Equal(t, "FN-123456", cr.FiscalNumber)
	assert.Equal(t, "Model X", cr.Model)
	assert.Equal(t, aggregate.StatusInactive, cr.Status)
}

// TestCreateCashRegister_DuplicateFiscalNumber tests duplicate fiscal number validation
func TestCreateCashRegister_DuplicateFiscalNumber(t *testing.T) {
	existingCR, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)

	repo := &MockRepository{
		GetByFiscalNumberFunc: func(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error) {
			return existingCR, nil // Fiscal number already exists
		},
	}

	u := NewCashRegisterUseCase(repo)
	cr, err := u.CreateCashRegister(context.Background(), testOrgID, "FN-123456", "Model X", testUserID)

	assert.Error(t, err)
	assert.Equal(t, cashregistererrors.ErrCashRegisterFiscalNumberExists, err)
	assert.Nil(t, cr)
}

// TestCreateCashRegister_ValidationError tests validation errors
func TestCreateCashRegister_ValidationError(t *testing.T) {
	repo := &MockRepository{
		GetByFiscalNumberFunc: func(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error) {
			return nil, errors.New("not found")
		},
	}

	u := NewCashRegisterUseCase(repo)

	tests := []struct {
		name         string
		orgID        uuidv7.UUID
		fiscalNumber string
		model        string
		createdBy    uuidv7.UUID
	}{
		{
			name:         "missing organization ID",
			orgID:        uuidv7.Nil,
			fiscalNumber: "FN-123456",
			model:        "Model X",
			createdBy:    testUserID,
		},
		{
			name:         "missing fiscal number",
			orgID:        testOrgID,
			fiscalNumber: "",
			model:        "Model X",
			createdBy:    testUserID,
		},
		{
			name:         "missing model",
			orgID:        testOrgID,
			fiscalNumber: "FN-123456",
			model:        "",
			createdBy:    testUserID,
		},
		{
			name:         "missing created by",
			orgID:        testOrgID,
			fiscalNumber: "FN-123456",
			model:        "Model X",
			createdBy:    uuidv7.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr, err := u.CreateCashRegister(context.Background(), tt.orgID, tt.fiscalNumber, tt.model, tt.createdBy)
			assert.Error(t, err)
			assert.Nil(t, cr)
		})
	}
}

// TestGetCashRegister_Success tests successful retrieval
func TestGetCashRegister_Success(t *testing.T) {
	expectedCR, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	crID := expectedCR.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			assert.Equal(t, crID, id)
			return expectedCR, nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	cr, err := u.GetCashRegister(context.Background(), crID)

	require.NoError(t, err)
	assert.Equal(t, expectedCR, cr)
}

// TestGetCashRegister_NotFound tests not found error
func TestGetCashRegister_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterNotFound
		},
	}

	u := NewCashRegisterUseCase(repo)
	cr, err := u.GetCashRegister(context.Background(), uuidv7.New())

	assert.Error(t, err)
	assert.Equal(t, cashregistererrors.ErrCashRegisterNotFound, err)
	assert.Nil(t, cr)
}

// TestGetByFiscalNumber_Success tests successful retrieval by fiscal number
func TestGetByFiscalNumber_Success(t *testing.T) {
	expectedCR, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)

	repo := &MockRepository{
		GetByFiscalNumberFunc: func(ctx context.Context, fiscalNumber string) (*aggregate.CashRegister, error) {
			assert.Equal(t, "FN-123456", fiscalNumber)
			return expectedCR, nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	cr, err := u.GetByFiscalNumber(context.Background(), "FN-123456")

	require.NoError(t, err)
	assert.Equal(t, expectedCR, cr)
}

// TestListCashRegisters_Success tests successful listing
func TestListCashRegisters_Success(t *testing.T) {
	cr1, _ := aggregate.NewCashRegister(testOrgID, "FN-001", "Model X", testUserID)
	cr2, _ := aggregate.NewCashRegister(testOrgID, "FN-002", "Model Y", testUserID)
	expected := []*aggregate.CashRegister{cr1, cr2}

	repo := &MockRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.CashRegister, error) {
			return expected, nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	result, err := u.ListCashRegisters(context.Background(), &testOrgID)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestListCashRegisters_Error(t *testing.T) {
	repo := &MockRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.CashRegister, error) {
			return nil, errors.New("list error")
		},
	}

	u := NewCashRegisterUseCase(repo)
	list, err := u.ListCashRegisters(context.Background(), &testOrgID)

	require.Error(t, err)
	require.Nil(t, list)
}

// TestActivateCashRegister_Success tests successful activation
func TestActivateCashRegister_Success(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.CashRegister) error {
			assert.Equal(t, aggregate.StatusActive, updated.Status)
			assert.NotEmpty(t, updated.LicenseKey)
			return nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	activated, err := u.ActivateCashRegister(context.Background(), crID, "LICENSE-KEY-123", testUserID)

	require.NoError(t, err)
	assert.Equal(t, aggregate.StatusActive, activated.Status)
}

func TestActivateCashRegister_UpdateFailed(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
		UpdateFunc: func(ctx context.Context, cr *aggregate.CashRegister) error {
			return errors.New("update error")
		},
	}

	u := NewCashRegisterUseCase(repo)
	updated, err := u.ActivateCashRegister(context.Background(), crID, "LICENSE-KEY-123", testUserID)

	require.ErrorIs(t, err, cashregistererrors.ErrCashRegisterUpdateFailed)
	require.Nil(t, updated)
}

// TestActivateCashRegister_NotFound tests activation with non-existent cash register
func TestActivateCashRegister_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterNotFound
		},
	}

	u := NewCashRegisterUseCase(repo)
	activated, err := u.ActivateCashRegister(context.Background(), uuidv7.New(), "LICENSE-KEY", testUserID)

	assert.Error(t, err)
	assert.Equal(t, cashregistererrors.ErrCashRegisterNotFound, err)
	assert.Nil(t, activated)
}

// TestActivateCashRegister_AlreadyActive tests activation when already active
func TestActivateCashRegister_AlreadyActive(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	_ = cr.Activate("OLD-LICENSE", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	activated, err := u.ActivateCashRegister(context.Background(), crID, "NEW-LICENSE", testUserID)

	assert.Error(t, err)
	assert.Equal(t, cashregistererrors.ErrCashRegisterAlreadyActive, err)
	assert.Nil(t, activated)
}

// TestDeactivateCashRegister_Success tests successful deactivation
func TestDeactivateCashRegister_Success(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	_ = cr.Activate("LICENSE-KEY", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.CashRegister) error {
			assert.Equal(t, aggregate.StatusInactive, updated.Status)
			return nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	deactivated, err := u.DeactivateCashRegister(context.Background(), crID, testUserID)

	require.NoError(t, err)
	assert.Equal(t, aggregate.StatusInactive, deactivated.Status)
}

func TestDeactivateCashRegister_UpdateFailed(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	_ = cr.Activate("LICENSE-KEY", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
		UpdateFunc: func(ctx context.Context, cr *aggregate.CashRegister) error {
			return errors.New("update error")
		},
	}

	u := NewCashRegisterUseCase(repo)
	updated, err := u.DeactivateCashRegister(context.Background(), crID, testUserID)

	require.ErrorIs(t, err, cashregistererrors.ErrCashRegisterUpdateFailed)
	require.Nil(t, updated)
}

// TestUpdateCashRegister_Success tests successful update
func TestUpdateCashRegister_Success(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)

	repo := &MockRepository{
		UpdateFunc: func(ctx context.Context, updated *aggregate.CashRegister) error {
			return nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	err := u.UpdateCashRegister(context.Background(), cr)

	assert.NoError(t, err)
}

func TestUpdateCashRegister_UpdateFailed(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)

	repo := &MockRepository{
		UpdateFunc: func(ctx context.Context, cr *aggregate.CashRegister) error {
			return errors.New("update error")
		},
	}

	u := NewCashRegisterUseCase(repo)
	err := u.UpdateCashRegister(context.Background(), cr)

	require.ErrorIs(t, err, cashregistererrors.ErrCashRegisterUpdateFailed)
}

// TestDeleteCashRegister_Success tests successful deletion
func TestDeleteCashRegister_Success(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			assert.Equal(t, crID, id)
			return nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	err := u.DeleteCashRegister(context.Background(), crID)

	assert.NoError(t, err)
}

func TestDeleteCashRegister_DeleteFailed(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-DELETE", "Model X", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return errors.New("delete error")
		},
	}

	u := NewCashRegisterUseCase(repo)
	err := u.DeleteCashRegister(context.Background(), crID)

	require.ErrorIs(t, err, cashregistererrors.ErrCashRegisterDeleteFailed)
}

// TestDeleteCashRegister_NotFound tests deletion with non-existent cash register
func TestDeleteCashRegister_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return nil, cashregistererrors.ErrCashRegisterNotFound
		},
	}

	u := NewCashRegisterUseCase(repo)
	err := u.DeleteCashRegister(context.Background(), uuidv7.New())

	assert.Error(t, err)
	assert.Equal(t, cashregistererrors.ErrCashRegisterNotFound, err)
}

// TestSyncCashRegister_Success tests successful sync update
func TestSyncCashRegister_Success(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.CashRegister) error {
			assert.NotNil(t, updated.LastSyncAt)
			return nil
		},
	}

	u := NewCashRegisterUseCase(repo)
	synced, err := u.SyncCashRegister(context.Background(), crID, testUserID)

	require.NoError(t, err)
	assert.NotNil(t, synced.LastSyncAt)
}

func TestSyncCashRegister_UpdateFailed(t *testing.T) {
	cr, _ := aggregate.NewCashRegister(testOrgID, "FN-123456", "Model X", testUserID)
	crID := cr.GetID()

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.CashRegister, error) {
			return cr, nil
		},
		UpdateFunc: func(ctx context.Context, cr *aggregate.CashRegister) error {
			return errors.New("update error")
		},
	}

	u := NewCashRegisterUseCase(repo)
	updated, err := u.SyncCashRegister(context.Background(), crID, testUserID)

	require.ErrorIs(t, err, cashregistererrors.ErrCashRegisterUpdateFailed)
	require.Nil(t, updated)
}
