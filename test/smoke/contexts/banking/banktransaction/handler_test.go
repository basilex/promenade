package banktransaction_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	txAggregate "github.com/basilex/promenade/internal/contexts/banking/banktransaction/aggregate"
	txHTTP "github.com/basilex/promenade/internal/contexts/banking/banktransaction/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

type MockBankTransactionUseCase struct {
	RecordTransactionFunc              func(ctx context.Context, accountID uuidv7.UUID, direction txAggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*txAggregate.BankTransaction, error)
	RecordTransactionWithExternalIDFunc func(ctx context.Context, accountID uuidv7.UUID, externalID string, direction txAggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*txAggregate.BankTransaction, error)
	GetTransactionFunc                  func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error)
	GetTransactionByExternalIDFunc      func(ctx context.Context, accountID uuidv7.UUID, externalID string) (*txAggregate.BankTransaction, error)
	ListTransactionsByAccountFunc       func(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*txAggregate.BankTransaction, error)
	ListUnmatchedTransactionsFunc       func(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*txAggregate.BankTransaction, error)
	CountTransactionsByAccountFunc      func(ctx context.Context, accountID uuidv7.UUID) (int, error)
	BookTransactionFunc                 func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	CancelTransactionFunc               func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	MatchToInvoiceFunc                  func(ctx context.Context, transactionID, invoiceID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	MatchToOrderFunc                    func(ctx context.Context, transactionID, orderID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	MatchToPaymentFunc                  func(ctx context.Context, transactionID, paymentID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	UnmatchTransactionFunc              func(ctx context.Context, transactionID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error
	SetCounterpartyFunc                 func(ctx context.Context, transactionID uuidv7.UUID, name, iban string, lastUpdatedBy uuidv7.UUID) error
	DeleteTransactionFunc               func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockBankTransactionUseCase) RecordTransaction(ctx context.Context, accountID uuidv7.UUID, direction txAggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*txAggregate.BankTransaction, error) {
	if m.RecordTransactionFunc != nil {
		return m.RecordTransactionFunc(ctx, accountID, direction, amountCents, currencyCode, transactionAt, description, lastUpdatedBy)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) RecordTransactionWithExternalID(ctx context.Context, accountID uuidv7.UUID, externalID string, direction txAggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*txAggregate.BankTransaction, error) {
	if m.RecordTransactionWithExternalIDFunc != nil {
		return m.RecordTransactionWithExternalIDFunc(ctx, accountID, externalID, direction, amountCents, currencyCode, transactionAt, description, lastUpdatedBy)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) GetTransaction(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
	if m.GetTransactionFunc != nil {
		return m.GetTransactionFunc(ctx, id)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) GetTransactionByExternalID(ctx context.Context, accountID uuidv7.UUID, externalID string) (*txAggregate.BankTransaction, error) {
	if m.GetTransactionByExternalIDFunc != nil {
		return m.GetTransactionByExternalIDFunc(ctx, accountID, externalID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) ListTransactionsByAccount(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*txAggregate.BankTransaction, error) {
	if m.ListTransactionsByAccountFunc != nil {
		return m.ListTransactionsByAccountFunc(ctx, accountID, limit, offset)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) ListUnmatchedTransactions(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*txAggregate.BankTransaction, error) {
	if m.ListUnmatchedTransactionsFunc != nil {
		return m.ListUnmatchedTransactionsFunc(ctx, accountID, limit, offset)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) CountTransactionsByAccount(ctx context.Context, accountID uuidv7.UUID) (int, error) {
	if m.CountTransactionsByAccountFunc != nil {
		return m.CountTransactionsByAccountFunc(ctx, accountID)
	}
	return 0, fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) BookTransaction(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.BookTransactionFunc != nil {
		return m.BookTransactionFunc(ctx, id, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) CancelTransaction(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.CancelTransactionFunc != nil {
		return m.CancelTransactionFunc(ctx, id, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) MatchToInvoice(ctx context.Context, transactionID, invoiceID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.MatchToInvoiceFunc != nil {
		return m.MatchToInvoiceFunc(ctx, transactionID, invoiceID, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) MatchToOrder(ctx context.Context, transactionID, orderID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.MatchToOrderFunc != nil {
		return m.MatchToOrderFunc(ctx, transactionID, orderID, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) MatchToPayment(ctx context.Context, transactionID, paymentID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.MatchToPaymentFunc != nil {
		return m.MatchToPaymentFunc(ctx, transactionID, paymentID, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) UnmatchTransaction(ctx context.Context, transactionID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
	if m.UnmatchTransactionFunc != nil {
		return m.UnmatchTransactionFunc(ctx, transactionID, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) SetCounterparty(ctx context.Context, transactionID uuidv7.UUID, name, iban string, lastUpdatedBy uuidv7.UUID) error {
	if m.SetCounterpartyFunc != nil {
		return m.SetCounterpartyFunc(ctx, transactionID, name, iban, lastUpdatedBy)
	}
	return fmt.Errorf("not implemented")
}

func (m *MockBankTransactionUseCase) DeleteTransaction(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteTransactionFunc != nil {
		return m.DeleteTransactionFunc(ctx, id)
	}
	return fmt.Errorf("not implemented")
}

func fakeTransaction() *txAggregate.BankTransaction {
	tx, _ := txAggregate.NewBankTransaction(
		uuidv7.New(),
		txAggregate.DirectionCredit,
		100000,
		"UAH",
		time.Now(),
		"Test payment",
		uuidv7.New(),
	)
	return tx
}

func TestBankTransactionHandler_Record_Debit_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		RecordTransactionFunc: func(ctx context.Context, accountID uuidv7.UUID, direction txAggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
		SetCounterpartyFunc: func(ctx context.Context, transactionID uuidv7.UUID, name, iban string, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.POST("/bank-transactions", handler.Record)

	body := map[string]interface{}{
		"account_id":     smoke.FakeUUID(),
		"direction":      "debit",
		"amount_cents":   125050,
		"currency_code":  "UAH",
		"transaction_at": time.Now().Format(time.RFC3339),
		"description":    "Оплата за товар",
	}

	w := smoke.MakeRequest(t, router, "POST", "/bank-transactions", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestBankTransactionHandler_Record_Credit_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		RecordTransactionFunc: func(ctx context.Context, accountID uuidv7.UUID, direction txAggregate.TransactionDirection, amountCents int64, currencyCode string, transactionAt time.Time, description string, lastUpdatedBy uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
		SetCounterpartyFunc: func(ctx context.Context, transactionID uuidv7.UUID, name, iban string, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.POST("/bank-transactions", handler.Record)

	body := map[string]interface{}{
		"account_id":     smoke.FakeUUID(),
		"direction":      "credit",
		"amount_cents":   500000,
		"currency_code":  "UAH",
		"transaction_at": time.Now().Format(time.RFC3339),
		"description":    "Надходження від клієнта",
	}

	w := smoke.MakeRequest(t, router, "POST", "/bank-transactions", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestBankTransactionHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.GET("/bank-transactions/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/bank-transactions/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_ListByAccount_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		ListTransactionsByAccountFunc: func(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*txAggregate.BankTransaction, error) {
			return []*txAggregate.BankTransaction{fakeTransaction()}, nil
		},
		CountTransactionsByAccountFunc: func(ctx context.Context, accountID uuidv7.UUID) (int, error) {
			return 1, nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.GET("/bank-transactions", handler.ListByAccount)

	w := smoke.MakeRequest(t, router, "GET", "/bank-transactions?account_id="+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_ListUnmatched_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		ListUnmatchedTransactionsFunc: func(ctx context.Context, accountID uuidv7.UUID, limit, offset int) ([]*txAggregate.BankTransaction, error) {
			return []*txAggregate.BankTransaction{fakeTransaction()}, nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.GET("/bank-transactions/unmatched", handler.ListUnmatched)

	w := smoke.MakeRequest(t, router, "GET", "/bank-transactions/unmatched?account_id="+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_Book_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		BookTransactionFunc: func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.PUT("/bank-transactions/:id/book", handler.Book)

	body := map[string]interface{}{
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-transactions/"+smoke.FakeUUID()+"/book", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_Match_ToInvoice_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		MatchToInvoiceFunc: func(ctx context.Context, transactionID, invoiceID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.PUT("/bank-transactions/:id/match", handler.Match)

	body := map[string]interface{}{
		"entity_type":     "invoice",
		"entity_id":       smoke.FakeUUID(),
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-transactions/"+smoke.FakeUUID()+"/match", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_Match_ToOrder_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		MatchToOrderFunc: func(ctx context.Context, transactionID, orderID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.PUT("/bank-transactions/:id/match", handler.Match)

	body := map[string]interface{}{
		"entity_type":     "order",
		"entity_id":       smoke.FakeUUID(),
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-transactions/"+smoke.FakeUUID()+"/match", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_Unmatch_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		UnmatchTransactionFunc: func(ctx context.Context, transactionID uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.PUT("/bank-transactions/:id/unmatch", handler.Unmatch)

	body := map[string]interface{}{
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-transactions/"+smoke.FakeUUID()+"/unmatch", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_SetCounterparty_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		SetCounterpartyFunc: func(ctx context.Context, transactionID uuidv7.UUID, name, iban string, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.PUT("/bank-transactions/:id/counterparty", handler.SetCounterparty)

	body := map[string]interface{}{
		"name":            "ТОВ Постачальник",
		"iban":            "UA1234567890",
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-transactions/"+smoke.FakeUUID()+"/counterparty", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_Cancel_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		CancelTransactionFunc: func(ctx context.Context, id uuidv7.UUID, lastUpdatedBy uuidv7.UUID) error {
			return nil
		},
		GetTransactionFunc: func(ctx context.Context, id uuidv7.UUID) (*txAggregate.BankTransaction, error) {
			return fakeTransaction(), nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.PUT("/bank-transactions/:id/cancel", handler.Cancel)

	body := map[string]interface{}{
		"last_updated_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "PUT", "/bank-transactions/"+smoke.FakeUUID()+"/cancel", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestBankTransactionHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockBankTransactionUseCase{
		DeleteTransactionFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := txHTTP.NewBankTransactionHandler(mockUC)
	router.DELETE("/bank-transactions/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/bank-transactions/"+smoke.FakeUUID(), nil)
	if w.Code != 204 {
		t.Errorf("expected status code 204, got %d", w.Code)
	}
}
