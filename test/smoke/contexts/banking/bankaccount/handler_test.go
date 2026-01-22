package bankaccount_test

import (
	"context"
	"fmt"
	"testing"

	accountHTTP "github.com/basilex/promenade/internal/contexts/banking/bankaccount/adapter/http"
	accountAggregate "github.com/basilex/promenade/internal/contexts/banking/bankaccount/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

type MockBankAccountUseCase struct {
	CreateManualAccountFunc    func(ctx context.Context, organizationID uuidv7.UUID, name, bankName, currencyCode string, lastUpdatedBy uuidv7.UUID) (*accountAggregate.BankAccount, error)
	ConnectProviderAccountFunc func(ctx context.Context, organizationID uuidv7.UUID, name, bankName string, provider accountAggregate.BankProvider, providerAccountID string, lastUpdatedBy uuidv7.UUID) (*accountAggregate.BankAccount, error)
	GetAccountFunc             func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.BankAccount, error)
	GetAccountByProviderFunc   func(ctx context.Context, provider accountAggregate.BankProvider, providerAccountID string) (*accountAggregate.BankAccount, error)
	ListAccountsFunc           func(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*accountAggregate.BankAccount, error)
	CountAccountsFunc          func(ctx context.Context, organizationID uuidv7.UUID) (int, error)
	UpdateAccountDetailsFunc   func(ctx context.Context, id uuidv7.UUID, name, bankName, iban, accountNumber string, lastUpdatedBy uuidv7.UUID) error
	UpdateBalanceFunc          func(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error
	RecordSyncFunc             func(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error
	ActivateAccountFunc        func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	DeactivateAccountFunc      func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	ArchiveAccountFunc         func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	DeleteAccountFunc          func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockBankAccountUseCase) CreateManualAccount(ctx context.Context, organizationID uuidv7.UUID, name, bankName, currencyCode string, lastUpdatedBy uuidv7.UUID) (*accountAggregate.BankAccount, error) {
	if m.CreateManualAccountFunc != nil {
		return m.CreateManualAccountFunc(ctx, organizationID, name, bankName, currencyCode, lastUpdatedBy)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) ConnectProviderAccount(ctx context.Context, organizationID uuidv7.UUID, name, bankName string, provider accountAggregate.BankProvider, providerAccountID string, lastUpdatedBy uuidv7.UUID) (*accountAggregate.BankAccount, error) {
	if m.ConnectProviderAccountFunc != nil {
		return m.ConnectProviderAccountFunc(ctx, organizationID, name, bankName, provider, providerAccountID, lastUpdatedBy)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) GetAccount(ctx context.Context, id uuidv7.UUID) (*accountAggregate.BankAccount, error) {
	if m.GetAccountFunc != nil {
		return m.GetAccountFunc(ctx, id)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) GetAccountByProvider(ctx context.Context, provider accountAggregate.BankProvider, providerAccountID string) (*accountAggregate.BankAccount, error) {
	if m.GetAccountByProviderFunc != nil {
		return m.GetAccountByProviderFunc(ctx, provider, providerAccountID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) ListAccounts(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*accountAggregate.BankAccount, error) {
	if m.ListAccountsFunc != nil {
		return m.ListAccountsFunc(ctx, organizationID, limit, offset)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) CountAccounts(ctx context.Context, organizationID uuidv7.UUID) (int, error) {
	if m.CountAccountsFunc != nil {
		return m.CountAccountsFunc(ctx, organizationID)
	}
	return 0, fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) UpdateAccountDetails(ctx context.Context, id uuidv7.UUID, name, bankName, iban, accountNumber string, lastUpdatedBy uuidv7.UUID) error {
	if m.UpdateAccountDetailsFunc != nil {
		return m.UpdateAccountDetailsFunc(ctx, id, name, bankName, iban, accountNumber, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) UpdateBalance(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error {
	if m.UpdateBalanceFunc != nil {
		return m.UpdateBalanceFunc(ctx, id, balanceCents, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) RecordSync(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error {
	if m.RecordSyncFunc != nil {
		return m.RecordSyncFunc(ctx, id, balanceCents, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) ActivateAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.ActivateAccountFunc != nil {
		return m.ActivateAccountFunc(ctx, id, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) DeactivateAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.DeactivateAccountFunc != nil {
		return m.DeactivateAccountFunc(ctx, id, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) ArchiveAccount(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.ArchiveAccountFunc != nil {
		return m.ArchiveAccountFunc(ctx, id, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankAccountUseCase) DeleteAccount(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteAccountFunc != nil {
		return m.DeleteAccountFunc(ctx, id)
	}
	return fmt.Errorf("not implemented")
}

func fakeAccount() *accountAggregate.BankAccount {
	acc, _ := accountAggregate.NewBankAccount(uuidv7.New(), "Test Account", "Test Bank", "UAH", uuidv7.New())
	return acc
}

func TestBankAccountHandler_CreateManual_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		CreateManualAccountFunc: func(ctx context.Context, organizationID uuidv7.UUID, name, bankName, currencyCode string, lastUpdatedBy uuidv7.UUID) (*accountAggregate.BankAccount, error) {
			return fakeAccount(), nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.POST("/bank-accounts/manual", handler.CreateManual)

	body := map[string]interface{}{
		"organization_id": smoke.FakeUUID(),
		"name":            "Test Account",
		"bank_name":       "Test Bank",
		"currency_code":   "UAH",
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/bank-accounts/manual", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestBankAccountHandler_ConnectProvider_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		ConnectProviderAccountFunc: func(ctx context.Context, organizationID uuidv7.UUID, name, bankName string, provider accountAggregate.BankProvider, providerAccountID string, lastUpdatedBy uuidv7.UUID) (*accountAggregate.BankAccount, error) {
			return fakeAccount(), nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.POST("/bank-accounts/provider", handler.ConnectProvider)

	body := map[string]interface{}{
		"organization_id":     smoke.FakeUUID(),
		"name":                "Monobank Account",
		"bank_name":           "Monobank",
		"provider":            "monobank",
		"provider_account_id": "mono_123",
		"last_updated_by":     smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/bank-accounts/provider", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestBankAccountHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		GetAccountFunc: func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.BankAccount, error) {
			return fakeAccount(), nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.GET("/bank-accounts/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/bank-accounts/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankAccountHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		ListAccountsFunc: func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*accountAggregate.BankAccount, error) {
			return []*accountAggregate.BankAccount{fakeAccount()}, nil
		},
		CountAccountsFunc: func(ctx context.Context, orgID uuidv7.UUID) (int, error) {
			return 1, nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.GET("/bank-accounts", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/bank-accounts?organization_id="+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankAccountHandler_UpdateBalance_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		UpdateBalanceFunc: func(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetAccountFunc: func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.BankAccount, error) {
			return fakeAccount(), nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.PUT("/bank-accounts/:id/balance", handler.UpdateBalance)

	body := map[string]interface{}{
		"balance_cents":   150000,
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-accounts/"+smoke.FakeUUID()+"/balance", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankAccountHandler_RecordSync_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		RecordSyncFunc: func(ctx context.Context, id uuidv7.UUID, balanceCents int64, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetAccountFunc: func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.BankAccount, error) {
			return fakeAccount(), nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.PUT("/bank-accounts/:id/sync", handler.RecordSync)

	body := map[string]interface{}{
		"balance_cents":   250000,
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-accounts/"+smoke.FakeUUID()+"/sync", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankAccountHandler_Activate_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		ActivateAccountFunc: func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetAccountFunc: func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.BankAccount, error) {
			return fakeAccount(), nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.PUT("/bank-accounts/:id/activate", handler.Activate)

	body := map[string]interface{}{
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-accounts/"+smoke.FakeUUID()+"/activate", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankAccountHandler_Deactivate_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		DeactivateAccountFunc: func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetAccountFunc: func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.BankAccount, error) {
			return fakeAccount(), nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.PUT("/bank-accounts/:id/deactivate", handler.Deactivate)

	body := map[string]interface{}{
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-accounts/"+smoke.FakeUUID()+"/deactivate", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankAccountHandler_Archive_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		ArchiveAccountFunc: func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetAccountFunc: func(ctx context.Context, id uuidv7.UUID) (*accountAggregate.BankAccount, error) {
			return fakeAccount(), nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.PUT("/bank-accounts/:id/archive", handler.Archive)

	body := map[string]interface{}{
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-accounts/"+smoke.FakeUUID()+"/archive", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankAccountHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankAccountUseCase{
		DeleteAccountFunc: func(ctx context.Context, accountID uuidv7.UUID) error {
			return nil
		},
	}

	handler := accountHTTP.NewBankAccountHandler(mockUC)
	router.DELETE("/bank-accounts/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/bank-accounts/"+smoke.FakeUUID(), nil)
	if w.Code != 204 {
		t.Errorf("expected status code 204, got %d", w.Code)
	}
}
