package payment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/basilex/promenade/internal/contexts/billing/payment"
	paymentHTTP "github.com/basilex/promenade/internal/contexts/billing/payment/adapter/http"
	paymentAggregate "github.com/basilex/promenade/internal/contexts/billing/payment/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/smoke"
)

// MockPaymentUseCase implements payment.IUseCase for testing
type MockPaymentUseCase struct {
	CreatePaymentFunc   func(ctx context.Context, customerID uuidv7.UUID, amount valueobject.Money, method paymentAggregate.PaymentMethod) (*paymentAggregate.Payment, error)
	GetPaymentFunc      func(ctx context.Context, paymentID uuidv7.UUID) (*paymentAggregate.Payment, error)
	DeletePaymentFunc   func(ctx context.Context, paymentID uuidv7.UUID) error
	ProcessPaymentFunc  func(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error
	CompletePaymentFunc func(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error
	RefundPaymentFunc   func(ctx context.Context, paymentID uuidv7.UUID, amount valueobject.Money) error
	ListPaymentsFunc    func(ctx context.Context, page, pageSize int) ([]*paymentAggregate.Payment, error)
	LinkToInvoiceFunc   func(ctx context.Context, paymentID, invoiceID uuidv7.UUID) error
	// Required methods to implement IUseCase interface
	GetPaymentByNumberFunc        func(ctx context.Context, paymentNo string) (*paymentAggregate.Payment, error)
	GetPaymentByTransactionIDFunc func(ctx context.Context, transactionID string) (*paymentAggregate.Payment, error)
	FailPaymentFunc               func(ctx context.Context, paymentID uuidv7.UUID, reason string) error
	CancelPaymentFunc             func(ctx context.Context, paymentID uuidv7.UUID, reason string) error
	ListPaymentsByCustomerFunc    func(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*paymentAggregate.Payment, error)
	ListPaymentsByInvoiceFunc     func(ctx context.Context, invoiceID uuidv7.UUID) ([]*paymentAggregate.Payment, error)
	ListPaymentsByStatusFunc      func(ctx context.Context, status paymentAggregate.PaymentStatus, page, pageSize int) ([]*paymentAggregate.Payment, error)
	GetTotalByCustomerFunc        func(ctx context.Context, customerID uuidv7.UUID) (valueobject.Money, error)
	GetTotalByInvoiceFunc         func(ctx context.Context, invoiceID uuidv7.UUID) (valueobject.Money, error)
	SetCardDetailsFunc            func(ctx context.Context, paymentID uuidv7.UUID, last4, brand string) error
	SetProviderFunc               func(ctx context.Context, paymentID uuidv7.UUID, provider string) error
	AddNoteFunc                   func(ctx context.Context, paymentID uuidv7.UUID, note string) error
}

// Implement all IUseCase methods
func (m *MockPaymentUseCase) CreatePayment(ctx context.Context, customerID uuidv7.UUID, amount valueobject.Money, method paymentAggregate.PaymentMethod) (*paymentAggregate.Payment, error) {
	if m.CreatePaymentFunc != nil {
		return m.CreatePaymentFunc(ctx, customerID, amount, method)
	}
	return nil, fmt.Errorf("CreatePaymentFunc not implemented")
}

func (m *MockPaymentUseCase) GetPayment(ctx context.Context, paymentID uuidv7.UUID) (*paymentAggregate.Payment, error) {
	if m.GetPaymentFunc != nil {
		return m.GetPaymentFunc(ctx, paymentID)
	}
	return nil, fmt.Errorf("GetPaymentFunc not implemented")
}

func (m *MockPaymentUseCase) GetPaymentByNumber(ctx context.Context, paymentNo string) (*paymentAggregate.Payment, error) {
	if m.GetPaymentByNumberFunc != nil {
		return m.GetPaymentByNumberFunc(ctx, paymentNo)
	}
	return nil, fmt.Errorf("GetPaymentByNumberFunc not implemented")
}

func (m *MockPaymentUseCase) GetPaymentByTransactionID(ctx context.Context, transactionID string) (*paymentAggregate.Payment, error) {
	if m.GetPaymentByTransactionIDFunc != nil {
		return m.GetPaymentByTransactionIDFunc(ctx, transactionID)
	}
	return nil, fmt.Errorf("GetPaymentByTransactionIDFunc not implemented")
}

func (m *MockPaymentUseCase) DeletePayment(ctx context.Context, paymentID uuidv7.UUID) error {
	if m.DeletePaymentFunc != nil {
		return m.DeletePaymentFunc(ctx, paymentID)
	}
	return fmt.Errorf("DeletePaymentFunc not implemented")
}

func (m *MockPaymentUseCase) ProcessPayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
	if m.ProcessPaymentFunc != nil {
		return m.ProcessPaymentFunc(ctx, paymentID, transactionID)
	}
	return fmt.Errorf("ProcessPaymentFunc not implemented")
}

func (m *MockPaymentUseCase) CompletePayment(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
	if m.CompletePaymentFunc != nil {
		return m.CompletePaymentFunc(ctx, paymentID, transactionID)
	}
	return fmt.Errorf("CompletePaymentFunc not implemented")
}

func (m *MockPaymentUseCase) FailPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error {
	if m.FailPaymentFunc != nil {
		return m.FailPaymentFunc(ctx, paymentID, reason)
	}
	return fmt.Errorf("FailPaymentFunc not implemented")
}

func (m *MockPaymentUseCase) RefundPayment(ctx context.Context, paymentID uuidv7.UUID, amount valueobject.Money) error {
	if m.RefundPaymentFunc != nil {
		return m.RefundPaymentFunc(ctx, paymentID, amount)
	}
	return fmt.Errorf("RefundPaymentFunc not implemented")
}

func (m *MockPaymentUseCase) CancelPayment(ctx context.Context, paymentID uuidv7.UUID, reason string) error {
	if m.CancelPaymentFunc != nil {
		return m.CancelPaymentFunc(ctx, paymentID, reason)
	}
	return fmt.Errorf("CancelPaymentFunc not implemented")
}

func (m *MockPaymentUseCase) LinkToInvoice(ctx context.Context, paymentID, invoiceID uuidv7.UUID) error {
	if m.LinkToInvoiceFunc != nil {
		return m.LinkToInvoiceFunc(ctx, paymentID, invoiceID)
	}
	return fmt.Errorf("LinkToInvoiceFunc not implemented")
}

func (m *MockPaymentUseCase) ListPayments(ctx context.Context, page, pageSize int) ([]*paymentAggregate.Payment, error) {
	if m.ListPaymentsFunc != nil {
		return m.ListPaymentsFunc(ctx, page, pageSize)
	}
	return nil, fmt.Errorf("ListPaymentsFunc not implemented")
}

func (m *MockPaymentUseCase) ListPaymentsByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*paymentAggregate.Payment, error) {
	if m.ListPaymentsByCustomerFunc != nil {
		return m.ListPaymentsByCustomerFunc(ctx, customerID, page, pageSize)
	}
	return nil, fmt.Errorf("ListPaymentsByCustomerFunc not implemented")
}

func (m *MockPaymentUseCase) ListPaymentsByInvoice(ctx context.Context, invoiceID uuidv7.UUID) ([]*paymentAggregate.Payment, error) {
	if m.ListPaymentsByInvoiceFunc != nil {
		return m.ListPaymentsByInvoiceFunc(ctx, invoiceID)
	}
	return nil, fmt.Errorf("ListPaymentsByInvoiceFunc not implemented")
}

func (m *MockPaymentUseCase) ListPaymentsByStatus(ctx context.Context, status paymentAggregate.PaymentStatus, page, pageSize int) ([]*paymentAggregate.Payment, error) {
	if m.ListPaymentsByStatusFunc != nil {
		return m.ListPaymentsByStatusFunc(ctx, status, page, pageSize)
	}
	return nil, fmt.Errorf("ListPaymentsByStatusFunc not implemented")
}

func (m *MockPaymentUseCase) GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (valueobject.Money, error) {
	if m.GetTotalByCustomerFunc != nil {
		return m.GetTotalByCustomerFunc(ctx, customerID)
	}
	return valueobject.Money{}, fmt.Errorf("GetTotalByCustomerFunc not implemented")
}

func (m *MockPaymentUseCase) GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (valueobject.Money, error) {
	if m.GetTotalByInvoiceFunc != nil {
		return m.GetTotalByInvoiceFunc(ctx, invoiceID)
	}
	return valueobject.Money{}, fmt.Errorf("GetTotalByInvoiceFunc not implemented")
}

func (m *MockPaymentUseCase) SetCardDetails(ctx context.Context, paymentID uuidv7.UUID, last4, brand string) error {
	if m.SetCardDetailsFunc != nil {
		return m.SetCardDetailsFunc(ctx, paymentID, last4, brand)
	}
	return fmt.Errorf("SetCardDetailsFunc not implemented")
}

func (m *MockPaymentUseCase) SetProvider(ctx context.Context, paymentID uuidv7.UUID, provider string) error {
	if m.SetProviderFunc != nil {
		return m.SetProviderFunc(ctx, paymentID, provider)
	}
	return fmt.Errorf("SetProviderFunc not implemented")
}

func (m *MockPaymentUseCase) AddNote(ctx context.Context, paymentID uuidv7.UUID, note string) error {
	if m.AddNoteFunc != nil {
		return m.AddNoteFunc(ctx, paymentID, note)
	}
	return fmt.Errorf("AddNoteFunc not implemented")
}

// Helper to create fake payment
func fakePayment() *paymentAggregate.Payment {
	customerID := uuidv7.New()
	amount := valueobject.Money{Amount: 10000, Currency: "USD"}
	pmt, _ := paymentAggregate.NewPayment(customerID, amount, paymentAggregate.PaymentMethodCreditCard)
	return pmt
}

// ============================================================================
// Tests
// ============================================================================

func TestPaymentHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPaymentUseCase{
		CreatePaymentFunc: func(ctx context.Context, customerID uuidv7.UUID, amount valueobject.Money, method paymentAggregate.PaymentMethod) (*paymentAggregate.Payment, error) {
			return fakePayment(), nil
		},
		GetPaymentFunc: func(ctx context.Context, paymentID uuidv7.UUID) (*paymentAggregate.Payment, error) {
			return fakePayment(), nil // Reload after create
		},
	}

	handler := paymentHTTP.NewPaymentHandler(mockUC)
	router.POST("/payments", handler.Create)

	body := map[string]any{
		"customer_id": smoke.FakeUUID(),
		"amount":      10000,
		"currency":    "USD",
		"method":      "credit_card",
	}

	w := smoke.MakeRequest(t, router, "POST", "/payments", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestPaymentHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()
	mockUC := &MockPaymentUseCase{}
	handler := paymentHTTP.NewPaymentHandler(mockUC)
	router.POST("/payments", handler.Create)

	body := map[string]any{
		"customer_id": "invalid-uuid",
	}

	w := smoke.MakeRequest(t, router, "POST", "/payments", body)
	smoke.AssertErrorResponse(t, w, 400, "VALIDATION_ERROR")
}

func TestPaymentHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPaymentUseCase{
		GetPaymentFunc: func(ctx context.Context, paymentID uuidv7.UUID) (*paymentAggregate.Payment, error) {
			return fakePayment(), nil
		},
	}

	handler := paymentHTTP.NewPaymentHandler(mockUC)
	router.GET("/payments/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/payments/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestPaymentHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPaymentUseCase{
		GetPaymentFunc: func(ctx context.Context, paymentID uuidv7.UUID) (*paymentAggregate.Payment, error) {
			return nil, payment.ErrPaymentNotFound
		},
	}

	handler := paymentHTTP.NewPaymentHandler(mockUC)
	router.GET("/payments/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/payments/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}

func TestPaymentHandler_ProcessPayment_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPaymentUseCase{
		ProcessPaymentFunc: func(ctx context.Context, paymentID uuidv7.UUID, transactionID string) error {
			return nil
		},
		GetPaymentFunc: func(ctx context.Context, paymentID uuidv7.UUID) (*paymentAggregate.Payment, error) {
			return fakePayment(), nil // Reload after process
		},
	}

	handler := paymentHTTP.NewPaymentHandler(mockUC)
	router.POST("/payments/:id/process", handler.ProcessPayment)

	body := map[string]any{
		"transaction_id": "txn_123",
	}

	w := smoke.MakeRequest(t, router, "POST", "/payments/"+smoke.FakeUUID()+"/process", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestPaymentHandler_RefundPayment_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPaymentUseCase{
		RefundPaymentFunc: func(ctx context.Context, paymentID uuidv7.UUID, amount valueobject.Money) error {
			return nil
		},
		GetPaymentFunc: func(ctx context.Context, paymentID uuidv7.UUID) (*paymentAggregate.Payment, error) {
			return fakePayment(), nil // Reload after refund
		},
	}

	handler := paymentHTTP.NewPaymentHandler(mockUC)
	router.POST("/payments/:id/refund", handler.RefundPayment)

	body := map[string]any{
		"amount":   5000,
		"currency": "USD",
	}

	w := smoke.MakeRequest(t, router, "POST", "/payments/"+smoke.FakeUUID()+"/refund", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestPaymentHandler_LinkToInvoice_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPaymentUseCase{
		LinkToInvoiceFunc: func(ctx context.Context, paymentID, invoiceID uuidv7.UUID) error {
			return nil
		},
		GetPaymentFunc: func(ctx context.Context, paymentID uuidv7.UUID) (*paymentAggregate.Payment, error) {
			return fakePayment(), nil // Reload after link
		},
	}

	handler := paymentHTTP.NewPaymentHandler(mockUC)
	router.POST("/payments/:id/link-invoice", handler.LinkToInvoice)

	body := map[string]any{
		"invoice_id": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/payments/"+smoke.FakeUUID()+"/link-invoice", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestPaymentHandler_ListPayments_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockPaymentUseCase{
		ListPaymentsFunc: func(ctx context.Context, page, pageSize int) ([]*paymentAggregate.Payment, error) {
			return []*paymentAggregate.Payment{fakePayment(), fakePayment()}, nil
		},
	}

	handler := paymentHTTP.NewPaymentHandler(mockUC)
	router.GET("/payments", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/payments?page=1&page_size=10", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}
