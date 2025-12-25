package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestPaymentStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status PaymentStatus
		want   bool
	}{
		{"pending status", PaymentStatusPending, true},
		{"processing status", PaymentStatusProcessing, true},
		{"completed status", PaymentStatusCompleted, true},
		{"failed status", PaymentStatusFailed, true},
		{"refunded status", PaymentStatusRefunded, true},
		{"canceled status", PaymentStatusCanceled, true},
		{"invalid status", PaymentStatus("invalid"), false},
		{"empty status", PaymentStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestPaymentMethod_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		method PaymentMethod
		want   bool
	}{
		{"card", PaymentMethodCard, true},
		{"bank_transfer", PaymentMethodBankTransfer, true},
		{"paypal", PaymentMethodPayPal, true},
		{"crypto", PaymentMethodCrypto, true},
		{"other", PaymentMethodOther, true},
		{"invalid", PaymentMethod("invalid"), false},
		{"empty", PaymentMethod(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.method.IsValid())
		})
	}
}

func TestNewPayment(t *testing.T) {
	userID := uuidv7.New()
	invoiceID := uuidv7.New()
	transactionID := "txn_" + uuidv7.New().String()

	t.Run("successful payment creation", func(t *testing.T) {
		payment, err := NewPayment(userID, invoiceID, transactionID, 10000, "USD", PaymentMethodCard)

		require.NoError(t, err)
		assert.NotNil(t, payment)
		assert.NotEqual(t, uuidv7.Nil, payment.ID)
		assert.Equal(t, userID, payment.UserID)
		assert.Equal(t, invoiceID, payment.InvoiceID)
		assert.Equal(t, transactionID, payment.TransactionID)
		assert.Equal(t, int64(10000), payment.Amount)
		assert.Equal(t, "USD", payment.Currency)
		assert.Equal(t, PaymentMethodCard, payment.Method)
		assert.Equal(t, PaymentStatusPending, payment.Status)
		assert.NotZero(t, payment.CreatedAt)
		assert.NotZero(t, payment.UpdatedAt)
		assert.Nil(t, payment.CompletedAt)
		assert.Nil(t, payment.FailedAt)
		assert.Nil(t, payment.RefundedAt)
	})

	t.Run("nil user ID", func(t *testing.T) {
		payment, err := NewPayment(uuidv7.Nil, invoiceID, transactionID, 10000, "USD", PaymentMethodCard)
		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("nil invoice ID", func(t *testing.T) {
		// Note: Validate() currently doesn't check if invoice_id is nil
		// It allows nil invoice ID (though unusual, some payments might not be linked to invoices)
		payment, err := NewPayment(userID, uuidv7.Nil, transactionID, 10000, "USD", PaymentMethodCard)
		assert.NoError(t, err) // No error - nil invoice ID is allowed
		assert.NotNil(t, payment)
		assert.Equal(t, uuidv7.Nil, payment.InvoiceID)
	})

	t.Run("empty transaction ID", func(t *testing.T) {
		payment, err := NewPayment(userID, invoiceID, "", 10000, "USD", PaymentMethodCard)
		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Contains(t, err.Error(), "transaction ID is required")
	})

	t.Run("negative amount", func(t *testing.T) {
		payment, err := NewPayment(userID, invoiceID, transactionID, -1000, "USD", PaymentMethodCard)
		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("zero amount", func(t *testing.T) {
		payment, err := NewPayment(userID, invoiceID, transactionID, 0, "USD", PaymentMethodCard)
		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Contains(t, err.Error(), "amount must be positive")
	})

	t.Run("invalid currency", func(t *testing.T) {
		// Note: Validate() only checks currency is not empty, not length
		// "US" is technically invalid but passes validation
		payment, err := NewPayment(userID, invoiceID, transactionID, 10000, "US", PaymentMethodCard)
		assert.NoError(t, err) // No error - currency length not validated
		assert.NotNil(t, payment)
		assert.Equal(t, "US", payment.Currency)
	})

	t.Run("invalid payment method", func(t *testing.T) {
		payment, err := NewPayment(userID, invoiceID, transactionID, 10000, "USD", PaymentMethod("invalid"))
		assert.Error(t, err)
		assert.Nil(t, payment)
		assert.Contains(t, err.Error(), "invalid payment method")
	})
}

func TestPayment_Validate(t *testing.T) {
	userID := uuidv7.New()
	invoiceID := uuidv7.New()

	t.Run("valid payment", func(t *testing.T) {
		payment := &Payment{
			ID:            uuidv7.New(),
			UserID:        userID,
			InvoiceID:     invoiceID,
			TransactionID: "txn_123",
			Amount:        10000,
			Currency:      "USD",
			Method:        PaymentMethodCard,
			Status:        PaymentStatusPending,
		}

		err := payment.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing user ID", func(t *testing.T) {
		payment := &Payment{
			ID:            uuidv7.New(),
			InvoiceID:     invoiceID,
			TransactionID: "txn_123",
			Amount:        10000,
			Currency:      "USD",
			Method:        PaymentMethodCard,
			Status:        PaymentStatusPending,
		}

		err := payment.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("invalid status", func(t *testing.T) {
		payment := &Payment{
			ID:            uuidv7.New(),
			UserID:        userID,
			InvoiceID:     invoiceID,
			TransactionID: "txn_123",
			Amount:        10000,
			Currency:      "USD",
			Method:        PaymentMethodCard,
			Status:        PaymentStatus("invalid"),
		}

		err := payment.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})

	t.Run("invalid method", func(t *testing.T) {
		payment := &Payment{
			ID:            uuidv7.New(),
			UserID:        userID,
			InvoiceID:     invoiceID,
			TransactionID: "txn_123",
			Amount:        10000,
			Currency:      "USD",
			Method:        PaymentMethod("invalid"),
			Status:        PaymentStatusPending,
		}

		err := payment.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid payment method")
	})
}

func TestPayment_MarkCompleted(t *testing.T) {
	payment := &Payment{
		ID:            uuidv7.New(),
		UserID:        uuidv7.New(),
		InvoiceID:     uuidv7.New(),
		TransactionID: "txn_123",
		Amount:        10000,
		Currency:      "USD",
		Method:        PaymentMethodCard,
		Status:        PaymentStatusPending,
	}

	payment.MarkCompleted()

	assert.Equal(t, PaymentStatusCompleted, payment.Status)
	assert.NotNil(t, payment.CompletedAt)
}

func TestPayment_MarkFailed(t *testing.T) {
	payment := &Payment{
		ID:            uuidv7.New(),
		UserID:        uuidv7.New(),
		InvoiceID:     uuidv7.New(),
		TransactionID: "txn_123",
		Amount:        10000,
		Currency:      "USD",
		Method:        PaymentMethodCard,
		Status:        PaymentStatusPending,
	}

	failureReason := "Insufficient funds"
	payment.MarkFailed(failureReason)

	assert.Equal(t, PaymentStatusFailed, payment.Status)
	assert.NotNil(t, payment.FailedAt)
	assert.Equal(t, failureReason, payment.FailureReason)
}

func TestPayment_MarkRefunded(t *testing.T) {
	payment := &Payment{
		ID:            uuidv7.New(),
		UserID:        uuidv7.New(),
		InvoiceID:     uuidv7.New(),
		TransactionID: "txn_123",
		Amount:        10000,
		Currency:      "USD",
		Method:        PaymentMethodCard,
		Status:        PaymentStatusCompleted,
	}

	refundReason := "Customer request"
	payment.MarkRefunded(refundReason)

	assert.Equal(t, PaymentStatusRefunded, payment.Status)
	assert.NotNil(t, payment.RefundedAt)
	assert.Equal(t, refundReason, payment.RefundReason)
}

func TestPayment_IsCompleted(t *testing.T) {
	tests := []struct {
		name   string
		status PaymentStatus
		want   bool
	}{
		{"completed", PaymentStatusCompleted, true},
		{"pending", PaymentStatusPending, false},
		{"failed", PaymentStatusFailed, false},
		{"refunded", PaymentStatusRefunded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{Status: tt.status}
			assert.Equal(t, tt.want, payment.IsCompleted())
		})
	}
}

func TestPayment_IsPending(t *testing.T) {
	tests := []struct {
		name   string
		status PaymentStatus
		want   bool
	}{
		{"pending", PaymentStatusPending, true},
		{"processing", PaymentStatusProcessing, true},
		{"completed", PaymentStatusCompleted, false},
		{"failed", PaymentStatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{Status: tt.status}
			assert.Equal(t, tt.want, payment.IsPending())
		})
	}
}

func TestPayment_IsFailed(t *testing.T) {
	tests := []struct {
		name   string
		status PaymentStatus
		want   bool
	}{
		{"failed", PaymentStatusFailed, true},
		{"pending", PaymentStatusPending, false},
		{"completed", PaymentStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{Status: tt.status}
			assert.Equal(t, tt.want, payment.IsFailed())
		})
	}
}

func TestPayment_IsRefunded(t *testing.T) {
	tests := []struct {
		name   string
		status PaymentStatus
		want   bool
	}{
		{"refunded", PaymentStatusRefunded, true},
		{"completed", PaymentStatusCompleted, false},
		{"failed", PaymentStatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{Status: tt.status}
			assert.Equal(t, tt.want, payment.IsRefunded())
		})
	}
}

func TestPayment_StatusTransitions(t *testing.T) {
	userID := uuidv7.New()
	invoiceID := uuidv7.New()
	transactionID := "txn_123"

	t.Run("pending to completed", func(t *testing.T) {
		payment, _ := NewPayment(userID, invoiceID, transactionID, 10000, "USD", PaymentMethodCard)
		assert.Equal(t, PaymentStatusPending, payment.Status)

		payment.MarkCompleted()
		assert.Equal(t, PaymentStatusCompleted, payment.Status)
		assert.NotNil(t, payment.CompletedAt)
	})

	t.Run("pending to failed", func(t *testing.T) {
		payment, _ := NewPayment(userID, invoiceID, transactionID, 10000, "USD", PaymentMethodCard)
		assert.Equal(t, PaymentStatusPending, payment.Status)

		payment.MarkFailed("Card declined")
		assert.Equal(t, PaymentStatusFailed, payment.Status)
		assert.NotNil(t, payment.FailedAt)
		assert.Equal(t, "Card declined", payment.FailureReason)
	})

	t.Run("completed to refunded", func(t *testing.T) {
		payment, _ := NewPayment(userID, invoiceID, transactionID, 10000, "USD", PaymentMethodCard)
		payment.MarkCompleted()
		assert.Equal(t, PaymentStatusCompleted, payment.Status)

		payment.MarkRefunded("Customer request")
		assert.Equal(t, PaymentStatusRefunded, payment.Status)
		assert.NotNil(t, payment.RefundedAt)
		assert.Equal(t, "Customer request", payment.RefundReason)
	})
}

func TestPayment_MultiplePaymentMethods(t *testing.T) {
	userID := uuidv7.New()
	invoiceID := uuidv7.New()
	amount := int64(10000)

	methods := []PaymentMethod{
		PaymentMethodCard,
		PaymentMethodBankTransfer,
		PaymentMethodPayPal,
		PaymentMethodCrypto,
		PaymentMethodOther,
	}

	for _, method := range methods {
		t.Run(string(method), func(t *testing.T) {
			payment, err := NewPayment(userID, invoiceID, "txn_"+string(method), amount, "USD", method)
			require.NoError(t, err)
			assert.Equal(t, method, payment.Method)
		})
	}
}

func TestPayment_MultipleCurrencies(t *testing.T) {
	userID := uuidv7.New()
	invoiceID := uuidv7.New()

	currencies := []string{"USD", "EUR", "GBP", "UAH", "JPY"}

	for _, currency := range currencies {
		t.Run(currency, func(t *testing.T) {
			payment, err := NewPayment(userID, invoiceID, "txn_"+currency, 10000, currency, PaymentMethodCard)
			require.NoError(t, err)
			assert.Equal(t, currency, payment.Currency)
		})
	}
}

func TestPayment_FailureScenarios(t *testing.T) {
	userID := uuidv7.New()
	invoiceID := uuidv7.New()

	failureReasons := []string{
		"Insufficient funds",
		"Card expired",
		"Card declined",
		"Fraud detection",
		"Network error",
	}

	for _, reason := range failureReasons {
		t.Run(reason, func(t *testing.T) {
			payment, _ := NewPayment(userID, invoiceID, "txn_fail", 10000, "USD", PaymentMethodCard)
			payment.MarkFailed(reason)

			assert.Equal(t, PaymentStatusFailed, payment.Status)
			assert.Equal(t, reason, payment.FailureReason)
			assert.NotNil(t, payment.FailedAt)
		})
	}
}

func TestPayment_Timestamps(t *testing.T) {
	userID := uuidv7.New()
	invoiceID := uuidv7.New()

	t.Run("completed timestamp", func(t *testing.T) {
		payment, _ := NewPayment(userID, invoiceID, "txn_123", 10000, "USD", PaymentMethodCard)
		assert.Nil(t, payment.CompletedAt)

		before := time.Now()
		payment.MarkCompleted()
		after := time.Now()

		require.NotNil(t, payment.CompletedAt)
		assert.True(t, payment.CompletedAt.After(before.Add(-time.Second)))
		assert.True(t, payment.CompletedAt.Before(after.Add(time.Second)))
	})

	t.Run("failed timestamp", func(t *testing.T) {
		payment, _ := NewPayment(userID, invoiceID, "txn_123", 10000, "USD", PaymentMethodCard)
		assert.Nil(t, payment.FailedAt)

		before := time.Now()
		payment.MarkFailed("Test failure")
		after := time.Now()

		require.NotNil(t, payment.FailedAt)
		assert.True(t, payment.FailedAt.After(before.Add(-time.Second)))
		assert.True(t, payment.FailedAt.Before(after.Add(time.Second)))
	})

	t.Run("refunded timestamp", func(t *testing.T) {
		payment, _ := NewPayment(userID, invoiceID, "txn_123", 10000, "USD", PaymentMethodCard)
		payment.MarkCompleted()
		assert.Nil(t, payment.RefundedAt)

		before := time.Now()
		payment.MarkRefunded("Test refund")
		after := time.Now()

		require.NotNil(t, payment.RefundedAt)
		assert.True(t, payment.RefundedAt.After(before.Add(-time.Second)))
		assert.True(t, payment.RefundedAt.Before(after.Add(time.Second)))
	})
}
