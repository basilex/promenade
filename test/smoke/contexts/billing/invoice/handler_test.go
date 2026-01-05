package invoice_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/billing/invoice"
	invoiceHTTP "github.com/basilex/promenade/internal/contexts/billing/invoice/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/smoke"
)

// MockInvoiceUseCase implements invoice.IUseCase for testing
type MockInvoiceUseCase struct {
	CreateInvoiceFunc      func(ctx context.Context, customerID uuidv7.UUID, orderID *uuidv7.UUID, dueDate time.Time, currency string) (*invoice.Invoice, error)
	GetInvoiceFunc         func(ctx context.Context, id uuidv7.UUID) (*invoice.Invoice, error)
	GetInvoiceByNumberFunc func(ctx context.Context, invoiceNo string) (*invoice.Invoice, error)
	UpdateInvoiceFunc      func(ctx context.Context, inv *invoice.Invoice) error
	DeleteInvoiceFunc      func(ctx context.Context, id uuidv7.UUID) error
	AddLineItemFunc        func(ctx context.Context, invoiceID uuidv7.UUID, description string, quantity int, unitPrice valueobject.Money) (*invoice.Invoice, error)
	RemoveLineItemFunc     func(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID) (*invoice.Invoice, error)
	UpdateLineItemFunc     func(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID, quantity int) (*invoice.Invoice, error)
	SendInvoiceFunc        func(ctx context.Context, id uuidv7.UUID) error
	MarkAsPaidFunc         func(ctx context.Context, id uuidv7.UUID, paidDate time.Time) error
	MarkAsOverdueFunc      func(ctx context.Context, id uuidv7.UUID) error
	CancelInvoiceFunc      func(ctx context.Context, id uuidv7.UUID) error
	VoidInvoiceFunc        func(ctx context.Context, id uuidv7.UUID) error
	UpdateTaxAmountFunc    func(ctx context.Context, id uuidv7.UUID, taxAmount valueobject.Money) error
	ListInvoicesFunc       func(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error)
	ListByCustomerFunc     func(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*invoice.Invoice, int, error)
	ListByOrderFunc        func(ctx context.Context, orderID uuidv7.UUID) ([]*invoice.Invoice, error)
	ListByStatusFunc       func(ctx context.Context, status invoice.InvoiceStatus, page, pageSize int) ([]*invoice.Invoice, int, error)
	ListOverdueFunc        func(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error)
	CountByStatusFunc      func(ctx context.Context, status invoice.InvoiceStatus) (int, error)
	GetTotalRevenueFunc    func(ctx context.Context, from, to time.Time) (int64, error)
}

// Implement all IUseCase methods
func (m *MockInvoiceUseCase) CreateInvoice(ctx context.Context, customerID uuidv7.UUID, orderID *uuidv7.UUID, dueDate time.Time, currency string) (*invoice.Invoice, error) {
	if m.CreateInvoiceFunc != nil {
		return m.CreateInvoiceFunc(ctx, customerID, orderID, dueDate, currency)
	}
	return nil, fmt.Errorf("CreateInvoiceFunc not implemented")
}

func (m *MockInvoiceUseCase) GetInvoice(ctx context.Context, id uuidv7.UUID) (*invoice.Invoice, error) {
	if m.GetInvoiceFunc != nil {
		return m.GetInvoiceFunc(ctx, id)
	}
	return nil, fmt.Errorf("GetInvoiceFunc not implemented")
}

func (m *MockInvoiceUseCase) GetInvoiceByNumber(ctx context.Context, invoiceNo string) (*invoice.Invoice, error) {
	if m.GetInvoiceByNumberFunc != nil {
		return m.GetInvoiceByNumberFunc(ctx, invoiceNo)
	}
	return nil, fmt.Errorf("GetInvoiceByNumberFunc not implemented")
}

func (m *MockInvoiceUseCase) UpdateInvoice(ctx context.Context, inv *invoice.Invoice) error {
	if m.UpdateInvoiceFunc != nil {
		return m.UpdateInvoiceFunc(ctx, inv)
	}
	return fmt.Errorf("UpdateInvoiceFunc not implemented")
}

func (m *MockInvoiceUseCase) DeleteInvoice(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteInvoiceFunc != nil {
		return m.DeleteInvoiceFunc(ctx, id)
	}
	return fmt.Errorf("DeleteInvoiceFunc not implemented")
}

func (m *MockInvoiceUseCase) AddLineItem(ctx context.Context, invoiceID uuidv7.UUID, description string, quantity int, unitPrice valueobject.Money) (*invoice.Invoice, error) {
	if m.AddLineItemFunc != nil {
		return m.AddLineItemFunc(ctx, invoiceID, description, quantity, unitPrice)
	}
	return nil, fmt.Errorf("AddLineItemFunc not implemented")
}

func (m *MockInvoiceUseCase) RemoveLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID) (*invoice.Invoice, error) {
	if m.RemoveLineItemFunc != nil {
		return m.RemoveLineItemFunc(ctx, invoiceID, lineID)
	}
	return nil, fmt.Errorf("RemoveLineItemFunc not implemented")
}

func (m *MockInvoiceUseCase) UpdateLineItem(ctx context.Context, invoiceID uuidv7.UUID, lineID uuidv7.UUID, quantity int) (*invoice.Invoice, error) {
	if m.UpdateLineItemFunc != nil {
		return m.UpdateLineItemFunc(ctx, invoiceID, lineID, quantity)
	}
	return nil, fmt.Errorf("UpdateLineItemFunc not implemented")
}

func (m *MockInvoiceUseCase) SendInvoice(ctx context.Context, id uuidv7.UUID) error {
	if m.SendInvoiceFunc != nil {
		return m.SendInvoiceFunc(ctx, id)
	}
	return fmt.Errorf("SendInvoiceFunc not implemented")
}

func (m *MockInvoiceUseCase) MarkAsPaid(ctx context.Context, id uuidv7.UUID, paidDate time.Time) error {
	if m.MarkAsPaidFunc != nil {
		return m.MarkAsPaidFunc(ctx, id, paidDate)
	}
	return fmt.Errorf("MarkAsPaidFunc not implemented")
}

func (m *MockInvoiceUseCase) MarkAsOverdue(ctx context.Context, id uuidv7.UUID) error {
	if m.MarkAsOverdueFunc != nil {
		return m.MarkAsOverdueFunc(ctx, id)
	}
	return fmt.Errorf("MarkAsOverdueFunc not implemented")
}

func (m *MockInvoiceUseCase) CancelInvoice(ctx context.Context, id uuidv7.UUID) error {
	if m.CancelInvoiceFunc != nil {
		return m.CancelInvoiceFunc(ctx, id)
	}
	return fmt.Errorf("CancelInvoiceFunc not implemented")
}

func (m *MockInvoiceUseCase) VoidInvoice(ctx context.Context, id uuidv7.UUID) error {
	if m.VoidInvoiceFunc != nil {
		return m.VoidInvoiceFunc(ctx, id)
	}
	return fmt.Errorf("VoidInvoiceFunc not implemented")
}

func (m *MockInvoiceUseCase) UpdateTaxAmount(ctx context.Context, id uuidv7.UUID, taxAmount valueobject.Money) error {
	if m.UpdateTaxAmountFunc != nil {
		return m.UpdateTaxAmountFunc(ctx, id, taxAmount)
	}
	return fmt.Errorf("UpdateTaxAmountFunc not implemented")
}

func (m *MockInvoiceUseCase) ListInvoices(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error) {
	if m.ListInvoicesFunc != nil {
		return m.ListInvoicesFunc(ctx, page, pageSize)
	}
	return nil, 0, fmt.Errorf("ListInvoicesFunc not implemented")
}

func (m *MockInvoiceUseCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*invoice.Invoice, int, error) {
	if m.ListByCustomerFunc != nil {
		return m.ListByCustomerFunc(ctx, customerID, page, pageSize)
	}
	return nil, 0, fmt.Errorf("ListByCustomerFunc not implemented")
}

func (m *MockInvoiceUseCase) ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*invoice.Invoice, error) {
	if m.ListByOrderFunc != nil {
		return m.ListByOrderFunc(ctx, orderID)
	}
	return nil, fmt.Errorf("ListByOrderFunc not implemented")
}

func (m *MockInvoiceUseCase) ListByStatus(ctx context.Context, status invoice.InvoiceStatus, page, pageSize int) ([]*invoice.Invoice, int, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, status, page, pageSize)
	}
	return nil, 0, fmt.Errorf("ListByStatusFunc not implemented")
}

func (m *MockInvoiceUseCase) ListOverdue(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error) {
	if m.ListOverdueFunc != nil {
		return m.ListOverdueFunc(ctx, page, pageSize)
	}
	return nil, 0, fmt.Errorf("ListOverdueFunc not implemented")
}

func (m *MockInvoiceUseCase) CountByStatus(ctx context.Context, status invoice.InvoiceStatus) (int, error) {
	if m.CountByStatusFunc != nil {
		return m.CountByStatusFunc(ctx, status)
	}
	return 0, fmt.Errorf("CountByStatusFunc not implemented")
}

func (m *MockInvoiceUseCase) GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error) {
	if m.GetTotalRevenueFunc != nil {
		return m.GetTotalRevenueFunc(ctx, from, to)
	}
	return 0, fmt.Errorf("GetTotalRevenueFunc not implemented")
}

// Helper functions
func fakeInvoice() *invoice.Invoice {
	inv, _ := invoice.NewInvoice(uuidv7.New(), time.Now().Add(30*24*time.Hour), "USD")
	return inv
}

// Test Create - Success
func TestInvoiceHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInvoiceUseCase{
		CreateInvoiceFunc: func(ctx context.Context, customerID uuidv7.UUID, orderID *uuidv7.UUID, dueDate time.Time, currency string) (*invoice.Invoice, error) {
			return fakeInvoice(), nil
		},
	}

	handler := invoiceHTTP.NewInvoiceHandler(mockUC)
	router.POST("/invoices", handler.Create)

	body := map[string]interface{}{
		"customer_id": smoke.FakeUUID(),
		"due_date":    time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
		"currency":    "USD",
	}

	w := smoke.MakeRequest(t, router, "POST", "/invoices", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

// Test Create - ValidationError
func TestInvoiceHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInvoiceUseCase{}
	handler := invoiceHTTP.NewInvoiceHandler(mockUC)
	router.POST("/invoices", handler.Create)

	body := map[string]interface{}{
		// Missing required fields
	}

	w := smoke.MakeRequest(t, router, "POST", "/invoices", body)
	smoke.AssertErrorResponse(t, w, 400, "VALIDATION_ERROR")
}

// Test GetByID - Success
func TestInvoiceHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInvoiceUseCase{
		GetInvoiceFunc: func(ctx context.Context, id uuidv7.UUID) (*invoice.Invoice, error) {
			return fakeInvoice(), nil
		},
	}

	handler := invoiceHTTP.NewInvoiceHandler(mockUC)
	router.GET("/invoices/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/invoices/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

// Test GetByID - NotFound
func TestInvoiceHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInvoiceUseCase{
		GetInvoiceFunc: func(ctx context.Context, id uuidv7.UUID) (*invoice.Invoice, error) {
			return nil, invoice.ErrInvoiceNotFound
		},
	}

	handler := invoiceHTTP.NewInvoiceHandler(mockUC)
	router.GET("/invoices/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/invoices/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 404, "INVOICE_NOT_FOUND")
}

// Test Delete - Success
func TestInvoiceHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInvoiceUseCase{
		DeleteInvoiceFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := invoiceHTTP.NewInvoiceHandler(mockUC)
	router.DELETE("/invoices/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/invoices/"+smoke.FakeUUID(), nil)
	// Delete returns 204 No Content (no JSON body)
	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

// Test Delete - NotFound
func TestInvoiceHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInvoiceUseCase{
		DeleteInvoiceFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return invoice.ErrInvoiceNotFound
		},
	}

	handler := invoiceHTTP.NewInvoiceHandler(mockUC)
	router.DELETE("/invoices/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/invoices/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 404, "INVOICE_NOT_FOUND")
}

// Test List - Success
func TestInvoiceHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInvoiceUseCase{
		ListInvoicesFunc: func(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error) {
			return []*invoice.Invoice{fakeInvoice()}, 1, nil
		},
	}

	handler := invoiceHTTP.NewInvoiceHandler(mockUC)
	router.GET("/invoices", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/invoices?page=1&page_size=10", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

// Test List - EmptyResult
func TestInvoiceHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockInvoiceUseCase{
		ListInvoicesFunc: func(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error) {
			return []*invoice.Invoice{}, 0, nil
		},
	}

	handler := invoiceHTTP.NewInvoiceHandler(mockUC)
	router.GET("/invoices", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/invoices?page=1&page_size=10", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}
