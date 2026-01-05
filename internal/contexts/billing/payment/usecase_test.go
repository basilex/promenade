package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// Mock Repository
type mockPaymentRepository struct {
	CreateFunc               func(ctx context.Context, payment *Payment) error
	GetByIDFunc              func(ctx context.Context, id uuidv7.UUID) (*Payment, error)
	GetByPaymentNoFunc       func(ctx context.Context, paymentNo string) (*Payment, error)
	GetByTransactionIDFunc   func(ctx context.Context, transactionID string) (*Payment, error)
	UpdateFunc               func(ctx context.Context, payment *Payment) error
	DeleteFunc               func(ctx context.Context, id uuidv7.UUID) error
	ListFunc                 func(ctx context.Context, page, pageSize int) ([]*Payment, error)
	ListByCustomerIDFunc     func(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Payment, error)
	ListByInvoiceIDFunc      func(ctx context.Context, invoiceID uuidv7.UUID) ([]*Payment, error)
	ListByStatusFunc         func(ctx context.Context, status PaymentStatus, page, pageSize int) ([]*Payment, error)
	GetTotalByCustomerFunc   func(ctx context.Context, customerID uuidv7.UUID) (int64, error)
	GetTotalByInvoiceFunc    func(ctx context.Context, invoiceID uuidv7.UUID) (int64, error)
}

func (m *mockPaymentRepository) Create(ctx context.Context, payment *Payment) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, payment)
	}
	return nil
}

func (m *mockPaymentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, ErrPaymentNotFound
}

func (m *mockPaymentRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error) {
	if m.GetByPaymentNoFunc != nil {
		return m.GetByPaymentNoFunc(ctx, paymentNo)
	}
	return nil, ErrPaymentNotFound
}

func (m *mockPaymentRepository) GetByTransactionID(ctx context.Context, transactionID string) (*Payment, error) {
	if m.GetByTransactionIDFunc != nil {
		return m.GetByTransactionIDFunc(ctx, transactionID)
	}
	return nil, ErrPaymentNotFound
}

func (m *mockPaymentRepository) Update(ctx context.Context, payment *Payment) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, payment)
	}
	return nil
}

func (m *mockPaymentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockPaymentRepository) List(ctx context.Context, page, pageSize int) ([]*Payment, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, page, pageSize)
	}
	return nil, nil
}

func (m *mockPaymentRepository) ListByCustomerID(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Payment, error) {
	if m.ListByCustomerIDFunc != nil {
		return m.ListByCustomerIDFunc(ctx, customerID, page, pageSize)
	}
	return nil, nil
}

func (m *mockPaymentRepository) ListByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]*Payment, error) {
	if m.ListByInvoiceIDFunc != nil {
		return m.ListByInvoiceIDFunc(ctx, invoiceID)
	}
	return nil, nil
}

func (m *mockPaymentRepository) ListByStatus(ctx context.Context, status PaymentStatus, page, pageSize int) ([]*Payment, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, status, page, pageSize)
	}
	return nil, nil
}

func (m *mockPaymentRepository) GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (int64, error) {
	if m.GetTotalByCustomerFunc != nil {
		return m.GetTotalByCustomerFunc(ctx, customerID)
	}
	return 0, nil
}

func (m *mockPaymentRepository) GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (int64, error) {
	if m.GetTotalByInvoiceFunc != nil {
		return m.GetTotalByInvoiceFunc(ctx, invoiceID)
	}
	return 0, nil
}

func TestPaymentUseCase_CreatePayment(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()
	amount := valueobject.Money{Amount: 10000, Currency: "USD"}

	t.Run("create payment success", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			CreateFunc: func(ctx context.Context, payment *Payment) error {
				return nil
			},
		}
		uc := NewUseCase(mockRepo)

		payment, err := uc.CreatePayment(ctx, customerID, amount, PaymentMethodCreditCard)
		require.NoError(t, err)
		require.NotNil(t, payment)
		assert.Equal(t, customerID, payment.CustomerID)
		assert.Equal(t, amount.Amount, payment.Amount.Amount)
		assert.Equal(t, PaymentMethodCreditCard, payment.Method)
		assert.Equal(t, PaymentStatusPending, payment.Status)
	})

	t.Run("create payment repository error", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			CreateFunc: func(ctx context.Context, payment *Payment) error {
				return errors.New("database error")
			},
		}
		uc := NewUseCase(mockRepo)

		payment, err := uc.CreatePayment(ctx, customerID, amount, PaymentMethodCreditCard)
		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Contains(t, err.Error(), "database error")
	})

	t.Run("create payment invalid data", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{}
		uc := NewUseCase(mockRepo)

		payment, err := uc.CreatePayment(ctx, uuidv7.Nil, amount, PaymentMethodCreditCard)
		assert.Error(t, err)
		assert.Nil(t, payment)
	})
}

func TestPaymentUseCase_GetPayment(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()

	t.Run("get payment success", func(t *testing.T) {
		expectedPayment, _ := NewPayment(uuidv7.New(), valueobject.Money{Amount: 10000, Currency: "USD"}, PaymentMethodCreditCard)
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return expectedPayment, nil
			},
		}
		uc := NewUseCase(mockRepo)

		payment, err := uc.GetPayment(ctx, paymentID)
		require.NoError(t, err)
		assert.Equal(t, expectedPayment, payment)
	})

	t.Run("get payment not found", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return nil, ErrPaymentNotFound
			},
		}
		uc := NewUseCase(mockRepo)

		payment, err := uc.GetPayment(ctx, paymentID)
		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.True(t, errors.Is(err, ErrPaymentNotFound))
	})
}

func TestPaymentUseCase_ProcessPayment(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()

	t.Run("process payment success", func(t *testing.T) {
		payment, _ := NewPayment(uuidv7.New(), valueobject.Money{Amount: 10000, Currency: "USD"}, PaymentMethodCreditCard)
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return payment, nil
			},
			UpdateFunc: func(ctx context.Context, p *Payment) error {
				return nil
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.ProcessPayment(ctx, paymentID, "txn_123")
		require.NoError(t, err)
		assert.Equal(t, PaymentStatusProcessing, payment.Status)
	})

	t.Run("process payment not found", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return nil, ErrPaymentNotFound
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.ProcessPayment(ctx, paymentID, "txn_123")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrPaymentNotFound))
	})
}

func TestPaymentUseCase_CompletePayment(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()
	transactionID := "txn_123456"

	t.Run("complete payment success", func(t *testing.T) {
		payment, _ := NewPayment(uuidv7.New(), valueobject.Money{Amount: 10000, Currency: "USD"}, PaymentMethodCreditCard)
		_ = payment.Process()
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return payment, nil
			},
			UpdateFunc: func(ctx context.Context, p *Payment) error {
				return nil
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.CompletePayment(ctx, paymentID, transactionID)
		require.NoError(t, err)
		assert.Equal(t, PaymentStatusCompleted, payment.Status)
		assert.Equal(t, transactionID, payment.TransactionID)
	})

	t.Run("complete payment not found", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return nil, ErrPaymentNotFound
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.CompletePayment(ctx, paymentID, transactionID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrPaymentNotFound))
	})
}

func TestPaymentUseCase_FailPayment(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()
	reason := "Insufficient funds"

	t.Run("fail payment success", func(t *testing.T) {
		payment, _ := NewPayment(uuidv7.New(), valueobject.Money{Amount: 10000, Currency: "USD"}, PaymentMethodCreditCard)
		_ = payment.Process()
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return payment, nil
			},
			UpdateFunc: func(ctx context.Context, p *Payment) error {
				return nil
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.FailPayment(ctx, paymentID, reason)
		require.NoError(t, err)
		assert.Equal(t, PaymentStatusFailed, payment.Status)
		assert.Equal(t, reason, payment.FailureReason)
	})

	t.Run("fail payment not found", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return nil, ErrPaymentNotFound
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.FailPayment(ctx, paymentID, reason)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrPaymentNotFound))
	})
}

func TestPaymentUseCase_RefundPayment(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()
	refundAmount := valueobject.Money{Amount: 5000, Currency: "USD"}

	t.Run("refund payment success", func(t *testing.T) {
		payment, _ := NewPayment(uuidv7.New(), valueobject.Money{Amount: 10000, Currency: "USD"}, PaymentMethodCreditCard)
		_ = payment.Process()
		_ = payment.Complete("txn_123")
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return payment, nil
			},
			UpdateFunc: func(ctx context.Context, p *Payment) error {
				return nil
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.RefundPayment(ctx, paymentID, refundAmount)
		require.NoError(t, err)
		assert.Equal(t, PaymentStatusRefunded, payment.Status)
		assert.Equal(t, int64(5000), payment.RefundedAmount.Amount)
	})

	t.Run("refund payment not found", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return nil, ErrPaymentNotFound
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.RefundPayment(ctx, paymentID, refundAmount)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrPaymentNotFound))
	})
}

func TestPaymentUseCase_ListPaymentsByCustomer(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()

	t.Run("list payments success", func(t *testing.T) {
		expectedPayments := []*Payment{
			{CustomerID: customerID, Status: PaymentStatusCompleted},
			{CustomerID: customerID, Status: PaymentStatusPending},
		}
		mockRepo := &mockPaymentRepository{
			ListByCustomerIDFunc: func(ctx context.Context, cid uuidv7.UUID, page, pageSize int) ([]*Payment, error) {
				return expectedPayments, nil
			},
		}
		uc := NewUseCase(mockRepo)

		payments, err := uc.ListPaymentsByCustomer(ctx, customerID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, payments, 2)
	})

	t.Run("list payments empty result", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			ListByCustomerIDFunc: func(ctx context.Context, cid uuidv7.UUID, page, pageSize int) ([]*Payment, error) {
				return []*Payment{}, nil
			},
		}
		uc := NewUseCase(mockRepo)

		payments, err := uc.ListPaymentsByCustomer(ctx, customerID, 1, 10)
		require.NoError(t, err)
		assert.Empty(t, payments)
	})
}

func TestPaymentUseCase_LinkToInvoice(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()
	invoiceID := uuidv7.New()

	t.Run("link payment to invoice success", func(t *testing.T) {
		payment, _ := NewPayment(uuidv7.New(), valueobject.Money{Amount: 10000, Currency: "USD"}, PaymentMethodCreditCard)
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return payment, nil
			},
			UpdateFunc: func(ctx context.Context, p *Payment) error {
				return nil
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.LinkToInvoice(ctx, paymentID, invoiceID)
		require.NoError(t, err)
		require.NotNil(t, payment.InvoiceID)
		assert.Equal(t, invoiceID, *payment.InvoiceID)
	})

	t.Run("link payment not found", func(t *testing.T) {
		mockRepo := &mockPaymentRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Payment, error) {
				return nil, ErrPaymentNotFound
			},
		}
		uc := NewUseCase(mockRepo)

		err := uc.LinkToInvoice(ctx, paymentID, invoiceID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrPaymentNotFound))
	})
}
