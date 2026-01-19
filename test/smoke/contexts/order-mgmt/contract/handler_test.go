package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract"
	contractHTTP "github.com/basilex/promenade/internal/contexts/order-mgmt/contract/adapter/http"
	contractAggregate "github.com/basilex/promenade/internal/contexts/order-mgmt/contract/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockContractUseCase is a mock implementation of contract.IUseCase for testing
type MockContractUseCase struct {
	CreateContractFunc          func(ctx context.Context, orderID, customerID uuidv7.UUID, terms string) (*contractAggregate.Contract, error)
	GetContractFunc             func(ctx context.Context, contractID uuidv7.UUID) (*contractAggregate.Contract, error)
	UpdateContractFunc          func(ctx context.Context, contractID uuidv7.UUID, terms string) error
	DeleteContractFunc          func(ctx context.Context, contractID uuidv7.UUID) error
	SubmitForSignatureFunc      func(ctx context.Context, contractID uuidv7.UUID) error
	SignContractFunc            func(ctx context.Context, contractID uuidv7.UUID, signerName, signerEmail, signatureID string) error
	CompleteContractFunc        func(ctx context.Context, contractID uuidv7.UUID) error
	TerminateContractFunc       func(ctx context.Context, contractID uuidv7.UUID, reason string) error
	RenewContractFunc           func(ctx context.Context, contractID uuidv7.UUID) error
	SetExpirationDateFunc       func(ctx context.Context, contractID uuidv7.UUID, expiresAt time.Time) error
	ListContractsByOrderFunc    func(ctx context.Context, orderID uuidv7.UUID) ([]*contractAggregate.Contract, error)
	ListContractsByCustomerFunc func(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*contractAggregate.Contract, int, error)
	ListContractsByStatusFunc   func(ctx context.Context, status contractAggregate.ContractStatus, page, pageSize int) ([]*contractAggregate.Contract, int, error)
	ListExpiringSoonFunc        func(ctx context.Context, days int) ([]*contractAggregate.Contract, error)
	GetActiveContractsFunc      func(ctx context.Context) ([]*contractAggregate.Contract, error)
}

func (m *MockContractUseCase) CreateContract(ctx context.Context, orderID, customerID uuidv7.UUID, terms string) (*contractAggregate.Contract, error) {
	if m.CreateContractFunc != nil {
		return m.CreateContractFunc(ctx, orderID, customerID, terms)
	}
	return nil, nil
}

func (m *MockContractUseCase) GetContract(ctx context.Context, contractID uuidv7.UUID) (*contractAggregate.Contract, error) {
	if m.GetContractFunc != nil {
		return m.GetContractFunc(ctx, contractID)
	}
	return nil, nil
}

func (m *MockContractUseCase) UpdateContract(ctx context.Context, contractID uuidv7.UUID, terms string) error {
	if m.UpdateContractFunc != nil {
		return m.UpdateContractFunc(ctx, contractID, terms)
	}
	return nil
}

func (m *MockContractUseCase) DeleteContract(ctx context.Context, contractID uuidv7.UUID) error {
	if m.DeleteContractFunc != nil {
		return m.DeleteContractFunc(ctx, contractID)
	}
	return nil
}

func (m *MockContractUseCase) SubmitForSignature(ctx context.Context, contractID uuidv7.UUID) error {
	if m.SubmitForSignatureFunc != nil {
		return m.SubmitForSignatureFunc(ctx, contractID)
	}
	return nil
}

func (m *MockContractUseCase) SignContract(ctx context.Context, contractID uuidv7.UUID, signerName, signerEmail, signatureID string) error {
	if m.SignContractFunc != nil {
		return m.SignContractFunc(ctx, contractID, signerName, signerEmail, signatureID)
	}
	return nil
}

func (m *MockContractUseCase) CompleteContract(ctx context.Context, contractID uuidv7.UUID) error {
	if m.CompleteContractFunc != nil {
		return m.CompleteContractFunc(ctx, contractID)
	}
	return nil
}

func (m *MockContractUseCase) TerminateContract(ctx context.Context, contractID uuidv7.UUID, reason string) error {
	if m.TerminateContractFunc != nil {
		return m.TerminateContractFunc(ctx, contractID, reason)
	}
	return nil
}

func (m *MockContractUseCase) RenewContract(ctx context.Context, contractID uuidv7.UUID) error {
	if m.RenewContractFunc != nil {
		return m.RenewContractFunc(ctx, contractID)
	}
	return nil
}

func (m *MockContractUseCase) SetExpirationDate(ctx context.Context, contractID uuidv7.UUID, expiresAt time.Time) error {
	if m.SetExpirationDateFunc != nil {
		return m.SetExpirationDateFunc(ctx, contractID, expiresAt)
	}
	return nil
}

func (m *MockContractUseCase) ListContractsByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*contractAggregate.Contract, error) {
	if m.ListContractsByOrderFunc != nil {
		return m.ListContractsByOrderFunc(ctx, orderID)
	}
	return nil, nil
}

func (m *MockContractUseCase) ListContractsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*contractAggregate.Contract, int, error) {
	if m.ListContractsByCustomerFunc != nil {
		return m.ListContractsByCustomerFunc(ctx, customerID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockContractUseCase) ListContractsByStatus(ctx context.Context, status contractAggregate.ContractStatus, page, pageSize int) ([]*contractAggregate.Contract, int, error) {
	if m.ListContractsByStatusFunc != nil {
		return m.ListContractsByStatusFunc(ctx, status, page, pageSize)
	}
	return nil, 0, nil
}

func (m *MockContractUseCase) ListExpiringSoon(ctx context.Context, days int) ([]*contractAggregate.Contract, error) {
	if m.ListExpiringSoonFunc != nil {
		return m.ListExpiringSoonFunc(ctx, days)
	}
	return nil, nil
}

func (m *MockContractUseCase) GetActiveContracts(ctx context.Context) ([]*contractAggregate.Contract, error) {
	if m.GetActiveContractsFunc != nil {
		return m.GetActiveContractsFunc(ctx)
	}
	return nil, nil
}

// Helper to create fake contract for testing
func fakeContract() *contractAggregate.Contract {
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	c := contractAggregate.NewContract(orderID, customerID, "Test contract terms")
	return c
}

// TestContractHandler_Create_Success tests successful contract creation
func TestContractHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		CreateContractFunc: func(ctx context.Context, orderID, customerID uuidv7.UUID, terms string) (*contractAggregate.Contract, error) {
			return fakeContract(), nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.POST("/contracts", handler.Create)

	body := map[string]any{
		"order_id":    smoke.FakeUUID(),
		"customer_id": smoke.FakeUUID(),
		"terms":       "Test terms",
	}

	w := smoke.MakeRequest(t, router, "POST", "/contracts", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

// TestContractHandler_Create_ValidationError tests validation error handling
func TestContractHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{}
	handler := contractHTTP.NewContractHandler(mockUC)
	router.POST("/contracts", handler.Create)

	body := map[string]any{
		"order_id": "invalid-uuid",
	}

	w := smoke.MakeRequest(t, router, "POST", "/contracts", body)
	smoke.AssertErrorResponse(t, w, 400, "BAD_REQUEST")
}

// TestContractHandler_GetByID_Success tests successful contract retrieval
func TestContractHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		GetContractFunc: func(ctx context.Context, contractID uuidv7.UUID) (*contractAggregate.Contract, error) {
			return fakeContract(), nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.GET("/contracts/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/contracts/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

// TestContractHandler_GetByID_NotFound tests not found error handling
func TestContractHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		GetContractFunc: func(ctx context.Context, contractID uuidv7.UUID) (*contractAggregate.Contract, error) {
			return nil, contract.ErrContractNotFound
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.GET("/contracts/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/contracts/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}

// TestContractHandler_Update_Success tests successful contract update
func TestContractHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		UpdateContractFunc: func(ctx context.Context, contractID uuidv7.UUID, terms string) error {
			return nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.PUT("/contracts/:id", handler.Update)

	body := map[string]any{
		"terms": "Updated terms",
	}

	w := smoke.MakeRequest(t, router, "PUT", "/contracts/"+smoke.FakeUUID(), body)
	smoke.AssertSuccessResponse(t, w, 200)
}

// TestContractHandler_Delete_Success tests successful contract deletion
func TestContractHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		DeleteContractFunc: func(ctx context.Context, contractID uuidv7.UUID) error {
			return nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.DELETE("/contracts/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/contracts/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

// TestContractHandler_SubmitForSignature_Success tests successful submission
func TestContractHandler_SubmitForSignature_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		SubmitForSignatureFunc: func(ctx context.Context, contractID uuidv7.UUID) error {
			return nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.POST("/contracts/:id/submit", handler.SubmitForSignature)

	w := smoke.MakeRequest(t, router, "POST", "/contracts/"+smoke.FakeUUID()+"/submit", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

// TestContractHandler_Sign_Success tests successful contract signing
func TestContractHandler_Sign_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		SignContractFunc: func(ctx context.Context, contractID uuidv7.UUID, signerName, signerEmail, signatureID string) error {
			return nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.POST("/contracts/:id/sign", handler.Sign)

	body := map[string]any{
		"signer_name":  "John Doe",
		"signer_email": "john@example.com",
		"signature_id": "sig_123",
	}

	w := smoke.MakeRequest(t, router, "POST", "/contracts/"+smoke.FakeUUID()+"/sign", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

// TestContractHandler_Complete_Success tests successful contract completion
func TestContractHandler_Complete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		CompleteContractFunc: func(ctx context.Context, contractID uuidv7.UUID) error {
			return nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.POST("/contracts/:id/complete", handler.Complete)

	w := smoke.MakeRequest(t, router, "POST", "/contracts/"+smoke.FakeUUID()+"/complete", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

// TestContractHandler_Terminate_Success tests successful contract termination
func TestContractHandler_Terminate_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		TerminateContractFunc: func(ctx context.Context, contractID uuidv7.UUID, reason string) error {
			return nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.POST("/contracts/:id/terminate", handler.Terminate)

	body := map[string]any{
		"reason": "Customer request",
	}

	w := smoke.MakeRequest(t, router, "POST", "/contracts/"+smoke.FakeUUID()+"/terminate", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

// TestContractHandler_List_Success tests successful contract listing
func TestContractHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockContractUseCase{
		GetActiveContractsFunc: func(ctx context.Context) ([]*contractAggregate.Contract, error) {
			return []*contractAggregate.Contract{fakeContract()}, nil
		},
	}

	handler := contractHTTP.NewContractHandler(mockUC)
	router.GET("/contracts", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/contracts", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}
