package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/account"
	"github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/account/cache"
	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockAccountRepository struct {
	CreateFunc             func(ctx context.Context, acc *aggregate.Account) error
	CreateManyFunc         func(ctx context.Context, accs []*aggregate.Account) error
	GetByIDFunc            func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error)
	GetByCodeFunc          func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.Account, error)
	UpdateFunc             func(ctx context.Context, acc *aggregate.Account) error
	UpdateManyFunc         func(ctx context.Context, accs []*aggregate.Account) error
	DeleteFunc             func(ctx context.Context, id uuidv7.UUID) error
	ListByOrganizationFunc func(ctx context.Context, orgID uuidv7.UUID, includeInactive bool) ([]*aggregate.Account, error)
	ListByTypeFunc         func(ctx context.Context, orgID uuidv7.UUID, accountType aggregate.AccountType) ([]*aggregate.Account, error)
	ListChildrenFunc       func(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Account, error)
	GetAllActiveFunc       func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.Account, error)
}

func (m *MockAccountRepository) Create(ctx context.Context, acc *aggregate.Account) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, acc)
	}
	return errors.New("CreateFunc not implemented")
}

func (m *MockAccountRepository) CreateMany(ctx context.Context, accs []*aggregate.Account) error {
	if m.CreateManyFunc != nil {
		return m.CreateManyFunc(ctx, accs)
	}
	return errors.New("CreateManyFunc not implemented")
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockAccountRepository) GetByCode(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.Account, error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, orgID, code)
	}
	return nil, errors.New("GetByCodeFunc not implemented")
}

func (m *MockAccountRepository) Update(ctx context.Context, acc *aggregate.Account) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, acc)
	}
	return errors.New("UpdateFunc not implemented")
}

func (m *MockAccountRepository) UpdateMany(ctx context.Context, accs []*aggregate.Account) error {
	if m.UpdateManyFunc != nil {
		return m.UpdateManyFunc(ctx, accs)
	}
	return errors.New("UpdateManyFunc not implemented")
}

func (m *MockAccountRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return errors.New("DeleteFunc not implemented")
}

func (m *MockAccountRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, includeInactive bool) ([]*aggregate.Account, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, orgID, includeInactive)
	}
	return nil, errors.New("ListByOrganizationFunc not implemented")
}

func (m *MockAccountRepository) ListByType(ctx context.Context, orgID uuidv7.UUID, accountType aggregate.AccountType) ([]*aggregate.Account, error) {
	if m.ListByTypeFunc != nil {
		return m.ListByTypeFunc(ctx, orgID, accountType)
	}
	return nil, errors.New("ListByTypeFunc not implemented")
}

func (m *MockAccountRepository) ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Account, error) {
	if m.ListChildrenFunc != nil {
		return m.ListChildrenFunc(ctx, parentID)
	}
	return nil, errors.New("ListChildrenFunc not implemented")
}

func (m *MockAccountRepository) GetAllActive(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.Account, error) {
	if m.GetAllActiveFunc != nil {
		return m.GetAllActiveFunc(ctx, orgID)
	}
	return nil, errors.New("GetAllActiveFunc not implemented")
}

// ============================================================================
// Tests
// ============================================================================

func TestCreateAccount_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	mockRepo := &MockAccountRepository{
		GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.Account, error) {
			return nil, account.ErrAccountNotFound
		},
		CreateFunc: func(ctx context.Context, acc *aggregate.Account) error {
			return nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	acc, err := uc.CreateAccount(context.Background(), orgID, "1000", "Cash", aggregate.AccountTypeAsset, nil, userID)
	require.NoError(t, err)
	assert.NotEqual(t, uuidv7.Nil, acc.ID)
	assert.Equal(t, "1000", acc.Code)
	assert.Equal(t, "Cash", acc.Name)
	assert.Equal(t, aggregate.AccountTypeAsset, acc.Type)
	assert.True(t, acc.IsActive)
	assert.Equal(t, 1, acc.Level)
}

func TestCreateAccount_DuplicateCode(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	existingAccount, _ := aggregate.NewAccount(orgID, "1000", "Existing", aggregate.AccountTypeAsset, "UAH", userID)

	mockRepo := &MockAccountRepository{
		GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.Account, error) {
			return existingAccount, nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	_, err := uc.CreateAccount(context.Background(), orgID, "1000", "Duplicate", aggregate.AccountTypeAsset, nil, userID)
	assert.Error(t, err)
	assert.Equal(t, account.ErrAccountCodeAlreadyExists, err)
}

func TestCreateAccount_WithParent(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	parentID := uuidv7.New()

	parent, _ := aggregate.NewAccount(orgID, "1000", "Assets", aggregate.AccountTypeAsset, "UAH", userID)
	parent.ID = parentID

	mockRepo := &MockAccountRepository{
		GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.Account, error) {
			return nil, account.ErrAccountNotFound
		},
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			if id == parentID {
				return parent, nil
			}
			return nil, account.ErrAccountNotFound
		},
		CreateFunc: func(ctx context.Context, acc *aggregate.Account) error {
			return nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	child, err := uc.CreateAccount(context.Background(), orgID, "1100", "Current Assets", aggregate.AccountTypeAsset, &parentID, userID)
	require.NoError(t, err)
	assert.Equal(t, parentID, *child.ParentID)
	assert.Equal(t, 2, child.Level)
}

func TestGetAccountByID_Success(t *testing.T) {
	acc, _ := aggregate.NewAccount(uuidv7.New(), "1000", "Cash", aggregate.AccountTypeAsset, "UAH", uuidv7.New())
	acc.ID = uuidv7.New()

	mockRepo := &MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			if id == acc.ID {
				return acc, nil
			}
			return nil, account.ErrAccountNotFound
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	retrieved, err := uc.GetAccountByID(context.Background(), acc.ID)
	require.NoError(t, err)
	assert.Equal(t, acc.ID, retrieved.ID)
	assert.Equal(t, "Cash", retrieved.Name)
}

func TestGetAccountByID_NotFound(t *testing.T) {
	mockRepo := &MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			return nil, account.ErrAccountNotFound
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	_, err := uc.GetAccountByID(context.Background(), uuidv7.New())
	assert.Error(t, err)
	assert.Equal(t, account.ErrAccountNotFound, err)
}

func TestGetAccountByCode_Success(t *testing.T) {
	orgID := uuidv7.New()
	acc, _ := aggregate.NewAccount(orgID, "1000", "Cash", aggregate.AccountTypeAsset, "UAH", uuidv7.New())

	mockRepo := &MockAccountRepository{
		GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.Account, error) {
			if code == "1000" {
				return acc, nil
			}
			return nil, account.ErrAccountNotFound
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	retrieved, err := uc.GetAccountByCode(context.Background(), orgID, "1000")
	require.NoError(t, err)
	assert.Equal(t, "1000", retrieved.Code)
	assert.Equal(t, "Cash", retrieved.Name)
}

func TestUpdateAccount_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	acc, _ := aggregate.NewAccount(orgID, "1000", "Old Name", aggregate.AccountTypeAsset, "UAH", userID)
	acc.ID = uuidv7.New()

	mockRepo := &MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			if id == acc.ID {
				return acc, nil
			}
			return nil, account.ErrAccountNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Account) error {
			return nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	updated, err := uc.UpdateAccount(context.Background(), acc.ID, "New Name", userID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
}

func TestActivateAccount_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	acc, _ := aggregate.NewAccount(orgID, "1000", "Account", aggregate.AccountTypeAsset, "UAH", userID)
	acc.ID = uuidv7.New()
	_ = acc.Deactivate()

	mockRepo := &MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			if id == acc.ID {
				return acc, nil
			}
			return nil, account.ErrAccountNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Account) error {
			return nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	err := uc.ActivateAccount(context.Background(), acc.ID, userID)
	require.NoError(t, err)
	assert.True(t, acc.IsActive)
}

func TestDeactivateAccount_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	acc, _ := aggregate.NewAccount(orgID, "1000", "Account", aggregate.AccountTypeAsset, "UAH", userID)
	acc.ID = uuidv7.New()

	mockRepo := &MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			if id == acc.ID {
				return acc, nil
			}
			return nil, account.ErrAccountNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Account) error {
			return nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	err := uc.DeactivateAccount(context.Background(), acc.ID, userID)
	require.NoError(t, err)
	assert.False(t, acc.IsActive)
}

func TestDeleteAccount_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	acc, _ := aggregate.NewAccount(orgID, "1000", "Account", aggregate.AccountTypeAsset, "UAH", userID)
	acc.ID = uuidv7.New()

	mockRepo := &MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			if id == acc.ID {
				return acc, nil
			}
			return nil, account.ErrAccountNotFound
		},
		ListChildrenFunc: func(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Account, error) {
			return []*aggregate.Account{}, nil
		},
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	err := uc.DeleteAccount(context.Background(), acc.ID)
	require.NoError(t, err)
}

func TestDeleteAccount_WithChildren(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	parent, _ := aggregate.NewAccount(orgID, "1000", "Parent", aggregate.AccountTypeAsset, "UAH", userID)
	parent.ID = uuidv7.New()

	child, _ := aggregate.NewAccount(orgID, "1100", "Child", aggregate.AccountTypeAsset, "UAH", userID)
	child.ParentID = &parent.ID

	mockRepo := &MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			if id == parent.ID {
				return parent, nil
			}
			return nil, account.ErrAccountNotFound
		},
		ListChildrenFunc: func(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Account, error) {
			if parentID == parent.ID {
				return []*aggregate.Account{child}, nil
			}
			return []*aggregate.Account{}, nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	err := uc.DeleteAccount(context.Background(), parent.ID)
	assert.Error(t, err)
	assert.Equal(t, account.ErrAccountHasChildren, err)
}

func TestListAccountsByOrganization_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	acc1, _ := aggregate.NewAccount(orgID, "1000", "Account 1", aggregate.AccountTypeAsset, "UAH", userID)
	acc2, _ := aggregate.NewAccount(orgID, "2000", "Account 2", aggregate.AccountTypeLiability, "UAH", userID)

	mockRepo := &MockAccountRepository{
		ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, includeInactive bool) ([]*aggregate.Account, error) {
			return []*aggregate.Account{acc1, acc2}, nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	accounts, err := uc.ListAccountsByOrganization(context.Background(), orgID, false)
	require.NoError(t, err)
	assert.Len(t, accounts, 2)
}

func TestListAccountsByType_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	acc1, _ := aggregate.NewAccount(orgID, "1000", "Asset 1", aggregate.AccountTypeAsset, "UAH", userID)
	acc2, _ := aggregate.NewAccount(orgID, "1100", "Asset 2", aggregate.AccountTypeAsset, "UAH", userID)

	mockRepo := &MockAccountRepository{
		ListByTypeFunc: func(ctx context.Context, orgID uuidv7.UUID, accountType aggregate.AccountType) ([]*aggregate.Account, error) {
			if accountType == aggregate.AccountTypeAsset {
				return []*aggregate.Account{acc1, acc2}, nil
			}
			return []*aggregate.Account{}, nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	accounts, err := uc.ListAccountsByType(context.Background(), orgID, aggregate.AccountTypeAsset)
	require.NoError(t, err)
	assert.Len(t, accounts, 2)
	for _, acc := range accounts {
		assert.Equal(t, aggregate.AccountTypeAsset, acc.Type)
	}
}

func TestListChildAccounts_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	parentID := uuidv7.New()

	child1, _ := aggregate.NewChildAccount(orgID, "1100", "Child 1", aggregate.AccountTypeAsset, "UAH", parentID, 1, userID)
	child2, _ := aggregate.NewChildAccount(orgID, "1200", "Child 2", aggregate.AccountTypeAsset, "UAH", parentID, 1, userID)

	mockRepo := &MockAccountRepository{
		ListChildrenFunc: func(ctx context.Context, pID uuidv7.UUID) ([]*aggregate.Account, error) {
			if pID == parentID {
				return []*aggregate.Account{child1, child2}, nil
			}
			return []*aggregate.Account{}, nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	children, err := uc.ListChildAccounts(context.Background(), parentID)
	require.NoError(t, err)
	assert.Len(t, children, 2)
	for _, child := range children {
		assert.Equal(t, parentID, *child.ParentID)
	}
}

func TestSetParent_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	parentID := uuidv7.New()
	childID := uuidv7.New()

	parent, _ := aggregate.NewAccount(orgID, "1000", "Parent", aggregate.AccountTypeAsset, "UAH", userID)
	parent.ID = parentID

	child, _ := aggregate.NewAccount(orgID, "1100", "Child", aggregate.AccountTypeAsset, "UAH", userID)
	child.ID = childID

	mockRepo := &MockAccountRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
			if id == childID {
				return child, nil
			}
			if id == parentID {
				return parent, nil
			}
			return nil, account.ErrAccountNotFound
		},
		UpdateFunc: func(ctx context.Context, updated *aggregate.Account) error {
			return nil
		},
	}

	accountCache := cache.NewAccountCache(mockRepo)
	var auditLogger *audit.AuditLogger = nil
	uc := NewAccountUseCase(mockRepo, accountCache, auditLogger)

	err := uc.SetParent(context.Background(), childID, parentID, userID)
	require.NoError(t, err)
	assert.NotNil(t, child.ParentID)
	assert.Equal(t, parentID, *child.ParentID)
	assert.Equal(t, 2, child.Level)
}
