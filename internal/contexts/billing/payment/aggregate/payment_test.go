package aggregate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

func TestNewPayment(t *testing.T) {
	tests := []struct {
		name       string
		customerID uuidv7.UUID
		amount     valueobject.Money
		method     PaymentMethod
		wantErr    bool
	}{
		{
			name:       "valid payment",
			customerID: uuidv7.New(),
			amount:     valueobject.Money{Amount: 10000, Currency: "USD"},
			method:     PaymentMethodCreditCard,
			wantErr:    false,
		},
		{
			name:       "empty customer ID",
			customerID: uuidv7.UUID{},
			amount:     valueobject.Money{Amount: 10000, Currency: "USD"},
			method:     PaymentMethodCreditCard,
			wantErr:    true,
		},
		{
			name:       "zero amount",
			customerID: uuidv7.New(),
			amount:     valueobject.Money{Amount: 0, Currency: "USD"},
			method:     PaymentMethodCreditCard,
			wantErr:    true,
		},
		{
			name:       "negative amount",
			customerID: uuidv7.New(),
			amount:     valueobject.Money{Amount: -10000, Currency: "USD"},
			method:     PaymentMethodCreditCard,
			wantErr:    true,
		},
		{
			name:       "empty method",
			customerID: uuidv7.New(),
			amount:     valueobject.Money{Amount: 10000, Currency: "USD"},
			method:     "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pmt, err := NewPayment(tt.customerID, tt.amount, tt.method)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pmt)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pmt)
				assert.NotEqual(t, uuidv7.UUID{}, pmt.GetID())
				assert.Equal(t, tt.customerID, pmt.CustomerID)
				assert.Equal(t, tt.amount.Amount, pmt.Amount.Amount)
				assert.Equal(t, tt.method, pmt.Method)
				assert.Equal(t, PaymentStatusPending, pmt.Status)
			}
		})
	}
}

func TestPayment_LinkToInvoice(t *testing.T) {
	tests := []struct {
		name      string
		invoiceID *uuidv7.UUID
		wantErr   bool
	}{
		{
			name:      "valid invoice ID",
			invoiceID: func() *uuidv7.UUID { id := uuidv7.New(); return &id }(),
			wantErr:   false,
		},
		{
			name:      "nil invoice ID",
			invoiceID: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerID := uuidv7.New()
			amount := valueobject.Money{Amount: 10000, Currency: "USD"}
			pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
			oldUpdatedAt := pmt.GetUpdatedAt()
			time.Sleep(1 * time.Millisecond)

			var err error
			if tt.invoiceID != nil {
				err = pmt.LinkToInvoice(*tt.invoiceID)
			} else {
				err = pmt.LinkToInvoice(uuidv7.UUID{})
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pmt.InvoiceID)
				assert.Equal(t, *tt.invoiceID, *pmt.InvoiceID)
				assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt), "UpdatedAt should be updated")
			}
		})
	}
}

func TestPayment_Process(t *testing.T) {
	tests := []struct {
		name    string
		status  PaymentStatus
		wantErr bool
	}{
		{
			name:    "pending to processing",
			status:  PaymentStatusPending,
			wantErr: false,
		},
		{
			name:    "already processing",
			status:  PaymentStatusProcessing,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerID := uuidv7.New()
			amount := valueobject.Money{Amount: 10000, Currency: "USD"}
			pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
			pmt.Status = tt.status
			oldUpdatedAt := pmt.GetUpdatedAt()
			time.Sleep(1 * time.Millisecond)

			err := pmt.Process()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, PaymentStatusProcessing, pmt.Status)
				require.NotNil(t, pmt.ProcessedAt)
				assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
			}
		})
	}
}

func TestPayment_Complete(t *testing.T) {
	tests := []struct {
		name          string
		status        PaymentStatus
		transactionID string
		wantErr       bool
	}{
		{
			name:          "processing to completed",
			status:        PaymentStatusProcessing,
			transactionID: "txn_123456",
			wantErr:       false,
		},
		{
			name:          "not processing",
			status:        PaymentStatusPending,
			transactionID: "txn_123456",
			wantErr:       true,
		},
		{
			name:          "empty transaction ID",
			status:        PaymentStatusProcessing,
			transactionID: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerID := uuidv7.New()
			amount := valueobject.Money{Amount: 10000, Currency: "USD"}
			pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
			pmt.Status = tt.status
			oldUpdatedAt := pmt.GetUpdatedAt()
			time.Sleep(1 * time.Millisecond)

			err := pmt.Complete(tt.transactionID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, PaymentStatusCompleted, pmt.Status)
				assert.Equal(t, tt.transactionID, pmt.TransactionID)
				assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
			}
		})
	}
}

func TestPayment_Fail(t *testing.T) {
	tests := []struct {
		name    string
		status  PaymentStatus
		reason  string
		wantErr bool
	}{
		{
			name:    "processing to failed",
			status:  PaymentStatusProcessing,
			reason:  "Insufficient funds",
			wantErr: false,
		},
		{
			name:    "not processing",
			status:  PaymentStatusPending,
			reason:  "Some reason",
			wantErr: true,
		},
		{
			name:    "empty reason",
			status:  PaymentStatusProcessing,
			reason:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerID := uuidv7.New()
			amount := valueobject.Money{Amount: 10000, Currency: "USD"}
			pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
			pmt.Status = tt.status
			oldUpdatedAt := pmt.GetUpdatedAt()
			time.Sleep(1 * time.Millisecond)

			err := pmt.Fail(tt.reason)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, PaymentStatusFailed, pmt.Status)
				assert.Equal(t, tt.reason, pmt.FailureReason)
				assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
			}
		})
	}
}

func TestPayment_Refund(t *testing.T) {
	tests := []struct {
		name         string
		status       PaymentStatus
		refundAmount valueobject.Money
		reason       string
		wantErr      bool
	}{
		{
			name:         "valid refund",
			status:       PaymentStatusCompleted,
			refundAmount: valueobject.Money{Amount: 5000, Currency: "USD"},
			reason:       "Customer request",
			wantErr:      false,
		},
		{
			name:         "not completed",
			status:       PaymentStatusPending,
			refundAmount: valueobject.Money{Amount: 5000, Currency: "USD"},
			reason:       "Customer request",
			wantErr:      true,
		},
		{
			name:         "refund exceeds amount",
			status:       PaymentStatusCompleted,
			refundAmount: valueobject.Money{Amount: 15000, Currency: "USD"},
			reason:       "Customer request",
			wantErr:      true,
		},
		{
			name:         "zero refund amount",
			status:       PaymentStatusCompleted,
			refundAmount: valueobject.Money{Amount: 0, Currency: "USD"},
			reason:       "Customer request",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerID := uuidv7.New()
			amount := valueobject.Money{Amount: 10000, Currency: "USD"}
			pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
			pmt.Status = tt.status
			oldUpdatedAt := pmt.GetUpdatedAt()
			time.Sleep(1 * time.Millisecond)

			err := pmt.Refund(tt.refundAmount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, PaymentStatusRefunded, pmt.Status)
				require.NotNil(t, pmt.RefundedAmount)
				assert.Equal(t, tt.refundAmount.Amount, pmt.RefundedAmount.Amount)
				require.NotNil(t, pmt.RefundedAt)
				assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
			}
		})
	}
}

func TestPayment_Cancel(t *testing.T) {
	tests := []struct {
		name    string
		status  PaymentStatus
		wantErr bool
	}{
		{
			name:    "cancel pending",
			status:  PaymentStatusPending,
			wantErr: false,
		},
		{
			name:    "cannot cancel completed",
			status:  PaymentStatusCompleted,
			wantErr: true,
		},
		{
			name:    "cannot cancel refunded",
			status:  PaymentStatusRefunded,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerID := uuidv7.New()
			amount := valueobject.Money{Amount: 10000, Currency: "USD"}
			pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
			pmt.Status = tt.status
			oldUpdatedAt := pmt.GetUpdatedAt()
			time.Sleep(1 * time.Millisecond)

			err := pmt.Cancel()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, PaymentStatusCancelled, pmt.Status)
				assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
			}
		})
	}
}

func TestPayment_SetCardDetails(t *testing.T) {
	customerID := uuidv7.New()
	amount := valueobject.Money{Amount: 10000, Currency: "USD"}
	pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
	oldUpdatedAt := pmt.GetUpdatedAt()
	time.Sleep(1 * time.Millisecond)

	pmt.SetCardDetails("4242", "Visa")
	assert.Equal(t, "4242", pmt.CardLast4)
	assert.Equal(t, "Visa", pmt.CardBrand)
	assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
}

func TestPayment_SetBankAccount(t *testing.T) {
	customerID := uuidv7.New()
	amount := valueobject.Money{Amount: 10000, Currency: "USD"}
	pmt, _ := NewPayment(customerID, amount, PaymentMethodBankTransfer)
	oldUpdatedAt := pmt.GetUpdatedAt()
	time.Sleep(1 * time.Millisecond)

	pmt.SetBankAccount("IBAN123")
	assert.Equal(t, "IBAN123", pmt.BankAccount)
	assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
}

func TestPayment_SetProvider(t *testing.T) {
	customerID := uuidv7.New()
	amount := valueobject.Money{Amount: 10000, Currency: "USD"}
	pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
	oldUpdatedAt := pmt.GetUpdatedAt()
	time.Sleep(1 * time.Millisecond)

	pmt.SetProvider("Stripe")
	assert.Equal(t, "Stripe", pmt.PaymentProvider)
	assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
}

func TestPayment_AddNote(t *testing.T) {
	customerID := uuidv7.New()
	amount := valueobject.Money{Amount: 10000, Currency: "USD"}
	pmt, _ := NewPayment(customerID, amount, PaymentMethodCreditCard)
	oldUpdatedAt := pmt.GetUpdatedAt()
	time.Sleep(1 * time.Millisecond)

	pmt.AddNote("Payment requires manual review")
	assert.Equal(t, "Payment requires manual review", pmt.Notes)
	assert.True(t, pmt.GetUpdatedAt().After(oldUpdatedAt))
}
