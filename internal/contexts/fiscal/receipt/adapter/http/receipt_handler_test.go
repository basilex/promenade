package http

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

type mockReceiptUseCase struct {
	createFunc func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error)
	getFunc    func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error)
	printFunc  func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error)
	listFunc   func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error)
	cancelFunc func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error)
	deleteFunc func(ctx context.Context, id uuidv7.UUID) error
}

func (m *mockReceiptUseCase) CreateReceipt(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, cashRegisterID, orderID, paymentType, receiptType, currency, lines, createdBy)
	}
	return nil, fmt.Errorf("create not implemented")
}

func (m *mockReceiptUseCase) GetReceipt(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, fmt.Errorf("get not implemented")
}

func (m *mockReceiptUseCase) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockReceiptUseCase) ListReceipts(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filters)
	}
	return nil, fmt.Errorf("list not implemented")
}

func (m *mockReceiptUseCase) MarkPrinted(ctx context.Context, id uuidv7.UUID, fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockReceiptUseCase) PrintReceipt(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
	if m.printFunc != nil {
		return m.printFunc(ctx, id, printedBy)
	}
	return nil, fmt.Errorf("print not implemented")
}

func (m *mockReceiptUseCase) CancelReceipt(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error) {
	if m.cancelFunc != nil {
		return m.cancelFunc(ctx, id, reason, cancelledBy)
	}
	return nil, fmt.Errorf("cancel not implemented")
}

func (m *mockReceiptUseCase) DeleteReceipt(ctx context.Context, id uuidv7.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return fmt.Errorf("delete not implemented")
}

func TestReceiptHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		createFunc: func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return aggregate.NewReceipt(cashRegisterID, orderID, paymentType, receiptType, currency, lines, createdBy)
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"cash_register_id": smoke.FakeUUID(),
		"order_id":         smoke.FakeUUID(),
		"payment_type":     "cash",
		"receipt_type":     "sale",
		"currency":         "UAH",
		"created_by":       smoke.FakeUUID(),
		"lines": []map[string]any{{
			"name":     "Item",
			"quantity": 1,
			"price":    1000,
			"tax_rate": 20,
		}},
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
	smoke.AssertSuccessResponse(t, resp, http.StatusCreated)
}

func TestReceiptHandler_Create_BindError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts", handler.Create)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", map[string]any{})
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		getFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
	}

	handler := NewReceiptHandler(usecase)
	router.GET("/fiscal/receipts/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestReceiptHandler_Create_InvalidCashRegisterID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"cash_register_id": "invalid",
		"order_id":         smoke.FakeUUID(),
		"payment_type":     "cash",
		"receipt_type":     "sale",
		"currency":         "UAH",
		"created_by":       smoke.FakeUUID(),
		"lines": []map[string]any{{
			"name":     "Item",
			"quantity": 1,
			"price":    1000,
			"tax_rate": 20,
		}},
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Create_DuplicateOrder(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		createFunc: func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptAlreadyExists
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"cash_register_id": smoke.FakeUUID(),
		"order_id":         smoke.FakeUUID(),
		"payment_type":     "cash",
		"receipt_type":     "sale",
		"currency":         "UAH",
		"created_by":       smoke.FakeUUID(),
		"lines": []map[string]any{{
			"name":     "Item",
			"quantity": 1,
			"price":    1000,
			"tax_rate": 20,
		}},
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Create_InvalidOrderID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"cash_register_id": smoke.FakeUUID(),
		"order_id":         "invalid",
		"payment_type":     "cash",
		"receipt_type":     "sale",
		"currency":         "UAH",
		"created_by":       smoke.FakeUUID(),
		"lines": []map[string]any{{
			"name":     "Item",
			"quantity": 1,
			"price":    1000,
			"tax_rate": 20,
		}},
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Create_InvalidCreatedBy(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"cash_register_id": smoke.FakeUUID(),
		"order_id":         smoke.FakeUUID(),
		"payment_type":     "cash",
		"receipt_type":     "sale",
		"currency":         "UAH",
		"created_by":       "invalid",
		"lines": []map[string]any{{
			"name":     "Item",
			"quantity": 1,
			"price":    1000,
			"tax_rate": 20,
		}},
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Create_LineTaxRateInvalid(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		createFunc: func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptLineTaxRateInvalid
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"cash_register_id": smoke.FakeUUID(),
		"order_id":         smoke.FakeUUID(),
		"payment_type":     "cash",
		"receipt_type":     "sale",
		"currency":         "UAH",
		"created_by":       smoke.FakeUUID(),
		"lines": []map[string]any{{
			"name":     "Item",
			"quantity": 1,
			"price":    1000,
			"tax_rate": 120,
		}},
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Create_PaymentTypeRequired(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		createFunc: func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrPaymentTypeRequired
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts", handler.Create)

	body := map[string]any{
		"cash_register_id": smoke.FakeUUID(),
		"order_id":         smoke.FakeUUID(),
		"payment_type":     "",
		"receipt_type":     "sale",
		"currency":         "UAH",
		"created_by":       smoke.FakeUUID(),
		"lines": []map[string]any{{
			"name":     "Item",
			"quantity": 1,
			"price":    1000,
			"tax_rate": 20,
		}},
	}

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Create_ErrorMappings(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{name: "cash register required", err: receipterrors.ErrCashRegisterIDRequired},
		{name: "order required", err: receipterrors.ErrOrderIDRequired},
		{name: "payment type required", err: receipterrors.ErrPaymentTypeRequired},
		{name: "receipt type required", err: receipterrors.ErrReceiptTypeRequired},
		{name: "currency required", err: receipterrors.ErrCurrencyRequired},
		{name: "created by required", err: receipterrors.ErrCreatedByRequired},
		{name: "receipt already exists", err: receipterrors.ErrReceiptAlreadyExists},
		{name: "line name required", err: receipterrors.ErrReceiptLineNameRequired},
		{name: "line quantity invalid", err: receipterrors.ErrReceiptLineQuantityInvalid},
		{name: "line price invalid", err: receipterrors.ErrReceiptLinePriceInvalid},
		{name: "line tax invalid", err: receipterrors.ErrReceiptLineTaxRateInvalid},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := smoke.SetupRouter()

			usecase := &mockReceiptUseCase{
				createFunc: func(ctx context.Context, cashRegisterID, orderID uuidv7.UUID, paymentType aggregate.PaymentType, receiptType aggregate.ReceiptType, currency string, lines []aggregate.ReceiptLine, createdBy uuidv7.UUID) (*aggregate.Receipt, error) {
					return nil, tc.err
				},
			}

			handler := NewReceiptHandler(usecase)
			router.POST("/fiscal/receipts", handler.Create)

			body := map[string]any{
				"cash_register_id": smoke.FakeUUID(),
				"order_id":         smoke.FakeUUID(),
				"payment_type":     "cash",
				"receipt_type":     "sale",
				"currency":         "UAH",
				"created_by":       smoke.FakeUUID(),
				"lines": []map[string]any{{
					"name":     "Item",
					"quantity": 1,
					"price":    1000,
					"tax_rate": 20,
				}},
			}

			resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts", body)
			smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
		})
	}
}

func TestReceiptHandler_List_InvalidOrderID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.GET("/fiscal/receipts", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts?order_id=invalid", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_List_InvalidCashRegisterID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.GET("/fiscal/receipts", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts?cash_register_id=invalid", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		listFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error) {
			rec, _ := aggregate.NewReceipt(uuidv7.New(), uuidv7.New(), aggregate.PaymentTypeCash, aggregate.ReceiptTypeSale, "UAH", []aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}}, uuidv7.New())
			return []*aggregate.Receipt{rec}, nil
		},
	}

	handler := NewReceiptHandler(usecase)
	router.GET("/fiscal/receipts", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts", nil)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}

func TestReceiptHandler_List_Error(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		listFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error) {
			return nil, fmt.Errorf("list error")
		},
	}

	handler := NewReceiptHandler(usecase)
	router.GET("/fiscal/receipts", handler.List)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestReceiptHandler_GetByID_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.GET("/fiscal/receipts/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts/invalid", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_GetByID_InternalError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		getFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, fmt.Errorf("get error")
		},
	}

	handler := NewReceiptHandler(usecase)
	router.GET("/fiscal/receipts/:id", handler.GetByID)

	resp := smoke.MakeRequest(t, router, "GET", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestReceiptHandler_Print_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		printFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{"printed_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestReceiptHandler_Print_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{"printed_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/invalid/print", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Print_InvalidPrintedBy(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{"printed_by": "invalid"}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Print_AlreadyCancelled(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		printFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptAlreadyCancelled
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{"printed_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Print_Failed(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		printFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptPrintFailed
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{"printed_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestReceiptHandler_Print_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		printFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return aggregate.NewReceipt(uuidv7.New(), uuidv7.New(), aggregate.PaymentTypeCash, aggregate.ReceiptTypeSale, "UAH", []aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}}, uuidv7.New())
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{"printed_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}

func TestReceiptHandler_Print_BindError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", map[string]any{})
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Print_AlreadyPrinted(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		printFunc: func(ctx context.Context, id uuidv7.UUID, printedBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptAlreadyPrinted
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/print", handler.MarkPrinted)

	body := map[string]any{"printed_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/print", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Cancel_ReasonRequired(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		cancelFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptCancelReasonRequired
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{"reason": "", "cancelled_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Cancel_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		cancelFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptNotFound
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{"reason": "test", "cancelled_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestReceiptHandler_Cancel_AlreadyCancelled(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		cancelFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, receipterrors.ErrReceiptAlreadyCancelled
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{"reason": "test", "cancelled_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Cancel_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{"reason": "test", "cancelled_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/invalid/cancel", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Cancel_InvalidCancelledBy(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{"reason": "test", "cancelled_by": "invalid"}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Cancel_InternalError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		cancelFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return nil, fmt.Errorf("cancel error")
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{"reason": "test", "cancelled_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestReceiptHandler_Cancel_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		cancelFunc: func(ctx context.Context, id uuidv7.UUID, reason string, cancelledBy uuidv7.UUID) (*aggregate.Receipt, error) {
			return aggregate.NewReceipt(uuidv7.New(), uuidv7.New(), aggregate.PaymentTypeCash, aggregate.ReceiptTypeSale, "UAH", []aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}}, uuidv7.New())
		},
	}

	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	body := map[string]any{"reason": "test", "cancelled_by": smoke.FakeUUID()}
	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}

func TestReceiptHandler_Cancel_BindError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.POST("/fiscal/receipts/:id/cancel", handler.Cancel)

	resp := smoke.MakeRequest(t, router, "POST", "/fiscal/receipts/"+smoke.FakeUUID()+"/cancel", map[string]any{})
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		deleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return receipterrors.ErrReceiptNotFound
		},
	}

	handler := NewReceiptHandler(usecase)
	router.DELETE("/fiscal/receipts/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, http.StatusNotFound, "NOT_FOUND")
}

func TestReceiptHandler_Delete_InternalError(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		deleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return fmt.Errorf("delete failed")
		},
	}

	handler := NewReceiptHandler(usecase)
	router.DELETE("/fiscal/receipts/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, resp, http.StatusInternalServerError, "INTERNAL_ERROR")
}

func TestReceiptHandler_Delete_InvalidID(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{}
	handler := NewReceiptHandler(usecase)
	router.DELETE("/fiscal/receipts/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/fiscal/receipts/invalid", nil)
	smoke.AssertErrorResponse(t, resp, http.StatusBadRequest, "BAD_REQUEST")
}

func TestReceiptHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	usecase := &mockReceiptUseCase{
		deleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := NewReceiptHandler(usecase)
	router.DELETE("/fiscal/receipts/:id", handler.Delete)

	resp := smoke.MakeRequest(t, router, "DELETE", "/fiscal/receipts/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, resp, http.StatusOK)
}
