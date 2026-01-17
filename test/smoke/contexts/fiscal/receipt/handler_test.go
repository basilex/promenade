package receipt_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	receiptHandler "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockReceiptUseCase implements receipt.IUseCase for smoke testing
type MockReceiptUseCase struct {
	CreateReceiptFunc func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType receipt.PaymentType, receiptType receipt.ReceiptType, currency string, lines []receipt.ReceiptLine, createdBy uuidv7.UUID) (*receipt.Receipt, error)
	GetReceiptFunc    func(ctx context.Context, id uuidv7.UUID) (*receipt.Receipt, error)
	GetByOrderIDFunc  func(ctx context.Context, orderID uuidv7.UUID) (*receipt.Receipt, error)
	ListReceiptsFunc  func(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error)
	MarkPrintedFunc   func(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*receipt.Receipt, error)
	PrintReceiptFunc  func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error)
	CancelReceiptFunc func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*receipt.Receipt, error)
	DeleteReceiptFunc func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockReceiptUseCase) CreateReceipt(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType receipt.PaymentType, receiptType receipt.ReceiptType, currency string, lines []receipt.ReceiptLine, createdBy uuidv7.UUID) (*receipt.Receipt, error) {
	if m.CreateReceiptFunc != nil {
		return m.CreateReceiptFunc(ctx, cashRegisterID, orderID, paymentType, receiptType, currency, lines, createdBy)
	}
	return nil, fmt.Errorf("CreateReceiptFunc not implemented")
}

func (m *MockReceiptUseCase) GetReceipt(ctx context.Context, id uuidv7.UUID) (*receipt.Receipt, error) {
	if m.GetReceiptFunc != nil {
		return m.GetReceiptFunc(ctx, id)
	}
	return nil, fmt.Errorf("GetReceiptFunc not implemented")
}

func (m *MockReceiptUseCase) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*receipt.Receipt, error) {
	if m.GetByOrderIDFunc != nil {
		return m.GetByOrderIDFunc(ctx, orderID)
	}
	return nil, fmt.Errorf("GetByOrderIDFunc not implemented")
}

func (m *MockReceiptUseCase) ListReceipts(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error) {
	if m.ListReceiptsFunc != nil {
		return m.ListReceiptsFunc(ctx, filters)
	}
	return nil, fmt.Errorf("ListReceiptsFunc not implemented")
}

func (m *MockReceiptUseCase) MarkPrinted(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
	if m.MarkPrintedFunc != nil {
		return m.MarkPrintedFunc(ctx, id, fiscalNumber, fiscalURL, qrCode, printedBy)
	}
	return nil, fmt.Errorf("MarkPrintedFunc not implemented")
}

func (m *MockReceiptUseCase) PrintReceipt(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
	if m.PrintReceiptFunc != nil {
		return m.PrintReceiptFunc(ctx, id, printedBy)
	}
	return nil, fmt.Errorf("PrintReceiptFunc not implemented")
}

func (m *MockReceiptUseCase) CancelReceipt(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*receipt.Receipt, error) {
	if m.CancelReceiptFunc != nil {
		return m.CancelReceiptFunc(ctx, id, reason, cancelledBy)
	}
	return nil, fmt.Errorf("CancelReceiptFunc not implemented")
}

func (m *MockReceiptUseCase) DeleteReceipt(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteReceiptFunc != nil {
		return m.DeleteReceiptFunc(ctx, id)
	}
	return fmt.Errorf("DeleteReceiptFunc not implemented")
}

func fakeReceipt(t *testing.T) *receipt.Receipt {
	rec, err := receipt.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		receipt.PaymentTypeCash,
		receipt.ReceiptTypeSale,
		"UAH",
		[]receipt.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	if err != nil {
		t.Fatalf("failed to create receipt: %v", err)
	}
	return rec
}

func TestReceiptHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		CreateReceiptFunc: func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType receipt.PaymentType, receiptType receipt.ReceiptType, currency string, lines []receipt.ReceiptLine, createdBy uuidv7.UUID) (*receipt.Receipt, error) {
			return fakeReceipt(t), nil
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"cash_register_id": smoke.FakeUUID(),
		"order_id":         smoke.FakeUUID(),
		"payment_type":     "cash",
		"receipt_type":     "sale",
		"currency":         "UAH",
		"created_by":       smoke.FakeUUID(),
		"lines": []map[string]any{
			{
				"name":     "Item",
				"quantity": 1,
				"price":    1000,
				"tax_rate": 20,
			},
		},
	}

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
	smoke.AssertSuccessResponse(t, w, http.StatusCreated)
}

func TestReceiptHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{}
	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"order_id": smoke.FakeUUID(),
	}
	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)

	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		GetReceiptFunc: func(ctx context.Context, id uuidv7.UUID) (*receipt.Receipt, error) {
			return fakeReceipt(t), nil
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.GET("/fiscal/receipts/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestReceiptHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		GetReceiptFunc: func(ctx context.Context, id uuidv7.UUID) (*receipt.Receipt, error) {
			return nil, receipt.ErrReceiptNotFound
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.GET("/fiscal/receipts/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestReceiptHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		ListReceiptsFunc: func(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error) {
			return []*receipt.Receipt{fakeReceipt(t), fakeReceipt(t)}, nil
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.GET("/fiscal/receipts", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts", nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestReceiptHandler_Print_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
			return fakeReceipt(t), nil
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{
		"printed_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestReceiptHandler_Print_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
			return nil, receipt.ErrReceiptNotFound
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{
		"printed_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestReceiptHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		DeleteReceiptFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.DELETE("/fiscal/receipts/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestReceiptHandler_Print_AlreadyCancelled(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		PrintReceiptFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*receipt.Receipt, error) {
			return nil, receipt.ErrReceiptAlreadyCancelled
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{
		"printed_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Cancel_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		CancelReceiptFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*receipt.Receipt, error) {
			return fakeReceipt(t), nil
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{
		"reason":        "customer request",
		"cancelled_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestReceiptHandler_Cancel_ReasonRequired(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		CancelReceiptFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*receipt.Receipt, error) {
			return nil, receipt.ErrReceiptCancelReasonRequired
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{
		"reason":        "",
		"cancelled_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Cancel_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		CancelReceiptFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*receipt.Receipt, error) {
			return nil, receipt.ErrReceiptNotFound
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{
		"reason":        "not found",
		"cancelled_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestReceiptHandler_Cancel_AlreadyCancelled(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		CancelReceiptFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*receipt.Receipt, error) {
			return nil, receipt.ErrReceiptAlreadyCancelled
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{
		"reason":        "duplicate",
		"cancelled_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockReceiptUseCase{
		DeleteReceiptFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return receipt.ErrReceiptNotFound
		},
	}

	handler := receiptHandler.NewReceiptHandler(mockUC)
	router.DELETE("/fiscal/receipts/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}
