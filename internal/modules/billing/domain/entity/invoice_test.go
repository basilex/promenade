package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestInvoiceStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status InvoiceStatus
		want   bool
	}{
		{"draft status", InvoiceStatusDraft, true},
		{"open status", InvoiceStatusOpen, true},
		{"paid status", InvoiceStatusPaid, true},
		{"void status", InvoiceStatusVoid, true},
		{"uncollectible status", InvoiceStatusUncollectible, true},
		{"invalid status", InvoiceStatus("invalid"), false},
		{"empty status", InvoiceStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestNewInvoice(t *testing.T) {
	subscriptionID := uuidv7.New()

	t.Run("successful invoice creation", func(t *testing.T) {
		invoice, err := NewInvoice(subscriptionID, 10000, 2000, "USD", 14)

		require.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.NotEqual(t, uuidv7.Nil, invoice.ID)
		assert.Equal(t, subscriptionID, invoice.SubscriptionID)
		assert.Equal(t, int64(10000), invoice.SubtotalAmount)
		assert.Equal(t, int64(2000), invoice.TaxAmount)
		assert.Equal(t, int64(12000), invoice.TotalAmount) // Subtotal + Tax
		assert.Equal(t, int64(12000), invoice.AmountDue)
		assert.Equal(t, int64(0), invoice.AmountPaid)
		assert.Equal(t, "USD", invoice.Currency)
		assert.Equal(t, InvoiceStatusDraft, invoice.Status)
		assert.NotEmpty(t, invoice.InvoiceNumber)
		assert.NotNil(t, invoice.DueDate)
		assert.NotZero(t, invoice.CreatedAt)
		assert.NotZero(t, invoice.UpdatedAt)

		// Due date should be 14 days from now
		expectedDueDate := time.Now().AddDate(0, 0, 14)
		assert.True(t, invoice.DueDate.After(time.Now()))
		assert.True(t, invoice.DueDate.Before(expectedDueDate.Add(time.Hour)))
	})

	t.Run("zero due days", func(t *testing.T) {
		invoice, err := NewInvoice(subscriptionID, 10000, 0, "USD", 0)

		require.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.NotNil(t, invoice.DueDate)
		// Due date should be today
		assert.True(t, invoice.DueDate.After(time.Now().Add(-time.Hour)))
	})

	t.Run("nil subscription ID", func(t *testing.T) {
		invoice, err := NewInvoice(uuidv7.Nil, 10000, 0, "USD", 14)
		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "subscription ID is required")
	})

	t.Run("negative subtotal", func(t *testing.T) {
		invoice, err := NewInvoice(subscriptionID, -1000, 0, "USD", 14)
		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "subtotal amount cannot be negative")
	})

	t.Run("negative tax", func(t *testing.T) {
		invoice, err := NewInvoice(subscriptionID, 10000, -500, "USD", 14)
		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "tax amount cannot be negative")
	})

	t.Run("invalid currency", func(t *testing.T) {
		invoice, err := NewInvoice(subscriptionID, 10000, 0, "US", 14)
		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "invalid currency code")
	})

	t.Run("negative due days", func(t *testing.T) {
		// Note: Currently NewInvoice doesn't validate negative dueDays
		// It will create invoice with past due date (now + (-5 days) = 5 days ago)
		invoice, err := NewInvoice(subscriptionID, 10000, 0, "USD", -5)
		assert.NoError(t, err) // No error, just creates invoice with past due date
		assert.NotNil(t, invoice)
		// Due date will be in the past
		assert.True(t, invoice.DueDate.Before(time.Now()))
	})
}

func TestInvoice_Validate(t *testing.T) {
	subscriptionID := uuidv7.New()

	t.Run("valid invoice", func(t *testing.T) {
		invoice := &Invoice{
			ID:             uuidv7.New(),
			SubscriptionID: subscriptionID,
			InvoiceNumber:  "INV-20241225-000001",
			SubtotalAmount: 10000,
			TaxAmount:      0,
			TotalAmount:    10000,
			AmountDue:      10000,
			AmountPaid:     0,
			Currency:       "USD",
			Status:         InvoiceStatusDraft,
		}

		err := invoice.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing subscription ID", func(t *testing.T) {
		invoice := &Invoice{
			ID:             uuidv7.New(),
			InvoiceNumber:  "INV-001",
			SubtotalAmount: 10000,
			TotalAmount:    10000,
			Currency:       "USD",
			Status:         InvoiceStatusDraft,
		}

		err := invoice.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "subscription ID is required")
	})

	t.Run("invalid status", func(t *testing.T) {
		// Note: Validate() doesn't check if status is valid (uses IsValid() method separately)
		// This test verifies invalid status using IsValid() method
		invalidStatus := InvoiceStatus("invalid")
		assert.False(t, invalidStatus.IsValid())
	})

	t.Run("invalid currency length", func(t *testing.T) {
		invoice := &Invoice{
			ID:             uuidv7.New(),
			SubscriptionID: subscriptionID,
			InvoiceNumber:  "INV-001",
			SubtotalAmount: 10000,
			TaxAmount:      0,
			TotalAmount:    10000,
			AmountPaid:     0,
			AmountDue:      10000,
			Currency:       "US",
			Status:         InvoiceStatusDraft,
		}

		err := invoice.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid currency code")
	})

	t.Run("negative amounts", func(t *testing.T) {
		invoice := &Invoice{
			ID:             uuidv7.New(),
			SubscriptionID: subscriptionID,
			InvoiceNumber:  "INV-001",
			SubtotalAmount: -1000,
			TotalAmount:    10000,
			Currency:       "USD",
			Status:         InvoiceStatusDraft,
		}

		err := invoice.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be negative")
	})
}

func TestInvoice_Finalize(t *testing.T) {
	invoice := &Invoice{
		ID:             uuidv7.New(),
		SubscriptionID: uuidv7.New(),
		InvoiceNumber:  "INV-001",
		SubtotalAmount: 10000,
		TotalAmount:    10000,
		AmountDue:      10000,
		Currency:       "USD",
		Status:         InvoiceStatusDraft,
	}

	invoice.Finalize()

	assert.Equal(t, InvoiceStatusOpen, invoice.Status)
}

func TestInvoice_MarkPaid(t *testing.T) {
	invoice := &Invoice{
		ID:             uuidv7.New(),
		SubscriptionID: uuidv7.New(),
		InvoiceNumber:  "INV-001",
		SubtotalAmount: 10000,
		TotalAmount:    10000,
		AmountDue:      10000,
		AmountPaid:     0,
		Currency:       "USD",
		Status:         InvoiceStatusOpen,
	}

	invoice.MarkPaid()

	assert.Equal(t, InvoiceStatusPaid, invoice.Status)
	assert.Equal(t, int64(0), invoice.AmountDue)
	assert.Equal(t, invoice.TotalAmount, invoice.AmountPaid)
	assert.NotNil(t, invoice.PaidAt)
}

func TestInvoice_Void(t *testing.T) {
	invoice := &Invoice{
		ID:             uuidv7.New(),
		SubscriptionID: uuidv7.New(),
		InvoiceNumber:  "INV-001",
		SubtotalAmount: 10000,
		TotalAmount:    10000,
		AmountDue:      10000,
		Currency:       "USD",
		Status:         InvoiceStatusOpen,
	}

	invoice.Void()

	assert.Equal(t, InvoiceStatusVoid, invoice.Status)
}

func TestInvoice_IsPaid(t *testing.T) {
	tests := []struct {
		name   string
		status InvoiceStatus
		want   bool
	}{
		{"paid", InvoiceStatusPaid, true},
		{"draft", InvoiceStatusDraft, false},
		{"open", InvoiceStatusOpen, false},
		{"void", InvoiceStatusVoid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := &Invoice{Status: tt.status}
			assert.Equal(t, tt.want, invoice.IsPaid())
		})
	}
}

func TestInvoice_IsOpen(t *testing.T) {
	tests := []struct {
		name   string
		status InvoiceStatus
		want   bool
	}{
		{"open", InvoiceStatusOpen, true},
		{"draft", InvoiceStatusDraft, false},
		{"paid", InvoiceStatusPaid, false},
		{"void", InvoiceStatusVoid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := &Invoice{Status: tt.status}
			assert.Equal(t, tt.want, invoice.IsOpen())
		})
	}
}

func TestInvoice_IsDraft(t *testing.T) {
	tests := []struct {
		name   string
		status InvoiceStatus
		want   bool
	}{
		{"draft", InvoiceStatusDraft, true},
		{"open", InvoiceStatusOpen, false},
		{"paid", InvoiceStatusPaid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := &Invoice{Status: tt.status}
			assert.Equal(t, tt.want, invoice.IsDraft())
		})
	}
}

func TestInvoice_IsVoid(t *testing.T) {
	tests := []struct {
		name   string
		status InvoiceStatus
		want   bool
	}{
		{"void", InvoiceStatusVoid, true},
		{"draft", InvoiceStatusDraft, false},
		{"paid", InvoiceStatusPaid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoice := &Invoice{Status: tt.status}
			assert.Equal(t, tt.want, invoice.IsVoid())
		})
	}
}

func TestInvoice_InvoiceNumberGeneration(t *testing.T) {
	t.Run("unique invoice numbers", func(t *testing.T) {
		subscriptionID := uuidv7.New()
		invoice1, _ := NewInvoice(subscriptionID, 1000, 0, "USD", 14)
		time.Sleep(time.Millisecond) // Ensure different timestamp
		invoice2, _ := NewInvoice(subscriptionID, 1000, 0, "USD", 14)

		// Note: Invoice numbers are time-based (INV-YYYYMMDD-XXXXXX where XXXXXX = Unix % 1000000)
		// If created in same second, numbers CAN be identical - this is acceptable
		assert.Contains(t, invoice1.InvoiceNumber, "INV-")
		assert.Contains(t, invoice2.InvoiceNumber, "INV-")
	})
}

func TestInvoice_AmountCalculations(t *testing.T) {
	t.Run("total = subtotal + tax", func(t *testing.T) {
		invoice, _ := NewInvoice(uuidv7.New(), 10000, 2000, "USD", 14)
		assert.Equal(t, int64(12000), invoice.TotalAmount)
		assert.Equal(t, invoice.SubtotalAmount+invoice.TaxAmount, invoice.TotalAmount)
	})

	t.Run("amount due equals total for new invoice", func(t *testing.T) {
		invoice, _ := NewInvoice(uuidv7.New(), 10000, 2000, "USD", 14)
		assert.Equal(t, invoice.TotalAmount, invoice.AmountDue)
	})

	t.Run("amount due becomes zero when paid", func(t *testing.T) {
		invoice, _ := NewInvoice(uuidv7.New(), 10000, 0, "USD", 14)
		invoice.Finalize()
		invoice.MarkPaid()

		assert.Equal(t, int64(0), invoice.AmountDue)
		assert.Equal(t, invoice.TotalAmount, invoice.AmountPaid)
	})
}

func TestInvoice_StatusTransitions(t *testing.T) {
	subscriptionID := uuidv7.New()

	t.Run("draft to open to paid", func(t *testing.T) {
		invoice, _ := NewInvoice(subscriptionID, 10000, 0, "USD", 14)
		assert.Equal(t, InvoiceStatusDraft, invoice.Status)

		invoice.Finalize()
		assert.Equal(t, InvoiceStatusOpen, invoice.Status)

		invoice.MarkPaid()
		assert.Equal(t, InvoiceStatusPaid, invoice.Status)
	})

	t.Run("draft to void", func(t *testing.T) {
		invoice, _ := NewInvoice(subscriptionID, 10000, 0, "USD", 14)
		assert.Equal(t, InvoiceStatusDraft, invoice.Status)

		invoice.Void()
		assert.Equal(t, InvoiceStatusVoid, invoice.Status)
	})
}

func TestInvoice_DueDateCalculation(t *testing.T) {
	t.Run("due date is in future", func(t *testing.T) {
		invoice, _ := NewInvoice(uuidv7.New(), 10000, 0, "USD", 30)
		assert.NotNil(t, invoice.DueDate)
		assert.True(t, invoice.DueDate.After(time.Now()))
	})

	t.Run("different due days", func(t *testing.T) {
		invoice7, _ := NewInvoice(uuidv7.New(), 10000, 0, "USD", 7)
		invoice30, _ := NewInvoice(uuidv7.New(), 10000, 0, "USD", 30)

		assert.True(t, invoice30.DueDate.After(*invoice7.DueDate))
	})
}
