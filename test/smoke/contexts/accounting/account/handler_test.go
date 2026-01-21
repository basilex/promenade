package account_test

import (
    "context"
    "fmt"
    "net/http"
    "testing"

    "github.com/basilex/promenade/internal/contexts/accounting/account"
    accountHTTP "github.com/basilex/promenade/internal/contexts/accounting/account/adapter/http"
    accountAggregate "github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/basilex/promenade/test/smoke"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

// MockAccountUseCase implements usecase.IAccountUseCase for testing
type MockAccountUseCase struct {
    CreateAccountFunc              func(ctx context.Context, organizationID uuidv7.UUID, code, name string, accountType accountAggregate.AccountType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*accountAggregate.Account, error)
    GetAccountByIDFunc             func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.Account, error)
    GetAccountByCodeFunc           func(ctx context.Context, organizationID uuidv7.UUID, code string) (*accountAggregate.Account, error)
    UpdateAccountFunc              func(ctx context.Context, id uuidv7.UUID, name string, updatedBy uuidv7.UUID) (*accountAggregate.Account, error)
    SetParentFunc                  func(ctx context.Context, id, parentID uuidv7.UUID, updatedBy uuidv7.UUID) error
    ActivateAccountFunc            func(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error
    DeactivateAccountFunc          func(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error
    DeleteAccountFunc              func(ctx context.Context, id uuidv7.UUID) error
    ListAccountsByOrganizationFunc func(ctx context.Context, organizationID uuidv7.UUID, includeInactive bool) ([]*accountAggregate.Account, error)
    ListAccountsByTypeFunc         func(ctx context.Context, organizationID uuidv7.UUID, accountType accountAggregate.AccountType) ([]*accountAggregate.Account, error)
    ListChildAccountsFunc          func(ctx context.Context, parentID uuidv7.UUID) ([]*accountAggregate.Account, error)
}

func (m *MockAccountUseCase) CreateAccount(ctx context.Context, organizationID uuidv7.UUID, code, name string, accountType accountAggregate.AccountType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*accountAggregate.Account, error) {
    if m.CreateAccountFunc != nil {
        return m.CreateAccountFunc(ctx, organizationID, code, name, accountType, parentID, createdBy)
    }
    return nil, fmt.Errorf("CreateAccountFunc not implemented")
}

func (m *MockAccountUseCase) GetAccountByID(ctx context.Context, id uuidv7.UUID) (*accountAggregate.Account, error) {
    if m.GetAccountByIDFunc != nil {
        return m.GetAccountByIDFunc(ctx, id)
    }
    return nil, fmt.Errorf("GetAccountByIDFunc not implemented")
}

func (m *MockAccountUseCase) GetAccountByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*accountAggregate.Account, error) {
    if m.GetAccountByCodeFunc != nil {
        return m.GetAccountByCodeFunc(ctx, organizationID, code)
    }
    return nil, fmt.Errorf("GetAccountByCodeFunc not implemented")
}

func (m *MockAccountUseCase) UpdateAccount(ctx context.Context, id uuidv7.UUID, name string, updatedBy uuidv7.UUID) (*accountAggregate.Account, error) {
    if m.UpdateAccountFunc != nil {
        return m.UpdateAccountFunc(ctx, id, name, updatedBy)
    }
    return nil, fmt.Errorf("UpdateAccountFunc not implemented")
}

func (m *MockAccountUseCase) SetParent(ctx context.Context, id, parentID uuidv7.UUID, updatedBy uuidv7.UUID) error {
    if m.SetParentFunc != nil {
        return m.SetParentFunc(ctx, id, parentID, updatedBy)
    }
    return fmt.Errorf("SetParentFunc not implemented")
}

func (m *MockAccountUseCase) ActivateAccount(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error {
    if m.ActivateAccountFunc != nil {
        return m.ActivateAccountFunc(ctx, id, updatedBy)
    }
    return fmt.Errorf("ActivateAccountFunc not implemented")
}

func (m *MockAccountUseCase) DeactivateAccount(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error {
    if m.DeactivateAccountFunc != nil {
        return m.DeactivateAccountFunc(ctx, id, updatedBy)
    }
    return fmt.Errorf("DeactivateAccountFunc not implemented")
}

func (m *MockAccountUseCase) DeleteAccount(ctx context.Context, id uuidv7.UUID) error {
    if m.DeleteAccountFunc != nil {
        return m.DeleteAccountFunc(ctx, id)
    }
    return fmt.Errorf("DeleteAccountFunc not implemented")
}

func (m *MockAccountUseCase) ListAccountsByOrganization(ctx context.Context, organizationID uuidv7.UUID, includeInactive bool) ([]*accountAggregate.Account, error) {
    if m.ListAccountsByOrganizationFunc != nil {
        return m.ListAccountsByOrganizationFunc(ctx, organizationID, includeInactive)
    }
    return nil, fmt.Errorf("ListAccountsByOrganizationFunc not implemented")
}

func (m *MockAccountUseCase) ListAccountsByType(ctx context.Context, organizationID uuidv7.UUID, accountType accountAggregate.AccountType) ([]*accountAggregate.Account, error) {
    if m.ListAccountsByTypeFunc != nil {
        return m.ListAccountsByTypeFunc(ctx, organizationID, accountType)
    }
    return nil, fmt.Errorf("ListAccountsByTypeFunc not implemented")
}

func (m *MockAccountUseCase) ListChildAccounts(ctx context.Context, parentID uuidv7.UUID) ([]*accountAggregate.Account, error) {
    if m.ListChildAccountsFunc != nil {
        return m.ListChildAccountsFunc(ctx, parentID)
    }
    return nil, fmt.Errorf("ListChildAccountsFunc not implemented")
}

func setupRouter(mockUC *MockAccountUseCase) *gin.Engine {
    router := smoke.SetupRouter()
    handler := accountHTTP.NewAccountHandler(mockUC)

    router.Use(func(c *gin.Context) {
        c.Set("organization_id", uuidv7.New().String())
        c.Set("user_id", uuidv7.New().String())
        c.Next()
    })

    handler.RegisterRoutes(router.Group("/api/v1/accounting"))
    return router
}

func TestCreateAccount_Success(t *testing.T) {
	accountID := uuidv7.New()

	mockUC := &MockAccountUseCase{
		CreateAccountFunc: func(ctx context.Context, organizationID uuidv7.UUID, code, name string, accountType accountAggregate.AccountType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*accountAggregate.Account, error) {
			acc, _ := accountAggregate.NewAccount(organizationID, code, name, accountType, "UAH", createdBy)
			acc.ID = accountID
			return acc, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"code":          "1000",
		"name":          "Cash",
		"account_type":  "asset",
		"currency_code": "UAH",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/accounts", body)
	if w.Code != http.StatusCreated {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateAccount_ValidationError(t *testing.T) {
	mockUC := &MockAccountUseCase{
		CreateAccountFunc: func(ctx context.Context, organizationID uuidv7.UUID, code, name string, accountType accountAggregate.AccountType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*accountAggregate.Account, error) {
			return nil, account.ErrAccountCodeEmpty
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"code":         "",
		"name":         "Cash",
		"account_type": "asset",
	}

	w := smoke.MakeRequest(t, router, "POST", "/api/v1/accounting/accounts", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAccount_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	acc, _ := accountAggregate.NewAccount(orgID, "1000", "Cash", accountAggregate.AccountTypeAsset, "UAH", userID)
	acc.ID = accountID

	mockUC := &MockAccountUseCase{
		GetAccountByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.Account, error) {
			return acc, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/accounts/"+accountID.String(), nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAccount_NotFound(t *testing.T) {
    accountID := uuidv7.New()

    mockUC := &MockAccountUseCase{
        GetAccountByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.Account, error) {
            return nil, account.ErrAccountNotFound
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/accounts/"+accountID.String(), nil)
    assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateAccount_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	accountID := uuidv7.New()

	acc, _ := accountAggregate.NewAccount(orgID, "1000", "Cash Updated", accountAggregate.AccountTypeAsset, "UAH", userID)
	acc.ID = accountID

	mockUC := &MockAccountUseCase{
		UpdateAccountFunc: func(ctx context.Context, id uuidv7.UUID, name string, updatedBy uuidv7.UUID) (*accountAggregate.Account, error) {
			return acc, nil
		},
	}

	router := setupRouter(mockUC)

	body := map[string]any{
		"name": "Cash Updated",
	}

	w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/accounts/"+accountID.String(), body)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestActivateAccount_Success(t *testing.T) {
    accountID := uuidv7.New()

    mockUC := &MockAccountUseCase{
        ActivateAccountFunc: func(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error {
            return nil
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/accounts/"+accountID.String()+"/activate", nil)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeactivateAccount_Success(t *testing.T) {
    accountID := uuidv7.New()

    mockUC := &MockAccountUseCase{
        DeactivateAccountFunc: func(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error {
            return nil
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "PUT", "/api/v1/accounting/accounts/"+accountID.String()+"/deactivate", nil)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteAccount_Success(t *testing.T) {
    accountID := uuidv7.New()

    mockUC := &MockAccountUseCase{
        DeleteAccountFunc: func(ctx context.Context, id uuidv7.UUID) error {
            return nil
        },
    }

    router := setupRouter(mockUC)

    w := smoke.MakeRequest(t, router, "DELETE", "/api/v1/accounting/accounts/"+accountID.String(), nil)
    assert.Equal(t, http.StatusOK, w.Code)
}

func TestListAccounts_Success(t *testing.T) {
	orgID := uuidv7.New()
	userID := uuidv7.New()

	accounts := []*accountAggregate.Account{
		func() *accountAggregate.Account {
			acc, _ := accountAggregate.NewAccount(orgID, "1000", "Cash", accountAggregate.AccountTypeAsset, "UAH", userID)
			return acc
		}(),
		func() *accountAggregate.Account {
			acc, _ := accountAggregate.NewAccount(orgID, "2000", "Revenue", accountAggregate.AccountTypeRevenue, "UAH", userID)
			return acc
		}(),
	}

	mockUC := &MockAccountUseCase{
		ListAccountsByOrganizationFunc: func(ctx context.Context, organizationID uuidv7.UUID, includeInactive bool) ([]*accountAggregate.Account, error) {
			return accounts, nil
		},
	}

	router := setupRouter(mockUC)

	w := smoke.MakeRequest(t, router, "GET", "/api/v1/accounting/accounts", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}