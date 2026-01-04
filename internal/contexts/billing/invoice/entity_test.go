package invoice

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

func TestNewInvoice(t *testing.T) {
	customerID := uuidv7.New()
	dueDate := time.Now().AddDate(0, 0, 30)

	t.Run("valid invoice", func(t *testing.T) {
		invoice, err := NewInvoice(customerID, dueDate, "USD")

		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, invoice.ID)
		assert.Equal(t, customerID, invoice.CustomerID)
		assert.Equal(t, InvoiceStatusDraft, invoice.Status)
	})

	t.Run("invalid customer ID", func(t *testing.T) {
		_, err := NewInvoice(uuidv7.Nil, dueDate, "USD")
		assert.ErrorIs(t, err, ErrInvoiceInvalidCustomerID)
	})
}

func TestInvoice_AddLine(t *testing.T) {
	invoice := createTestInvoice(t)
	unitPrice, _ := valueobject.NewMoney(10000, "USD")

	t.Run("add line to draft invoice", func(t *testing.T) {
		err := invoice.AddLine("Product A", 2, unitPrice)

		require.NoError(t, err)
		assert.Len(t, invoice.Lines, 1)
		assert.Equal(t, int64(20000), invoice.SubtotalAmount.Amount)
	})

	t.Run("cannot add line to sent invoice", func(t *testing.T) {
		invoice := createTestInvoice(t)
		_ = invoice.AddLine("Product A", 1, unitPrice)
		_ = invoice.MarkAsSent()

		err := invoice.AddLine("Product B", 1, unitPrice)
		assert.ErrorIs(t, err, ErrInvoiceCannotModifyNonDraft)
	})
}

func TestInvoice_MarkAsSent(t *testing.T) {
	t.Run("mark draft invoice as sent", func(t *testing.T) {
		invoice := createTestInvoice(t)
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		_ = invoice.AddLine("Product A", 1, unitPrice)

		err := invoice.MarkAsSent()

		require.NoError(t, err)
		assert.Equal(t, InvoiceStatusSent, invoice.Status)
	})

	t.Run("cannot send invoice without lines", func(t *testing.T) {
		invoice := createTestInvoice(t)

		err := invoice.MarkAsSent()
		assert.ErrorIs(t, err, ErrInvoiceNoLines)
	})
}

func TestInvoice_MarkAsPaid(t *testing.T) {
	paidDate := time.Now()

	t.Run("mark sent invoice as paid", func(t *testing.T) {
		invoice := createTestInvoice(t)
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		_ = invoice.AddLine("Product A", 1, unitPrice)
		_ = invoice.MarkAsSent()

		err := invoice.MarkAsPaid(paidDate)

		require.NoError(t, err)
		assert.Equal(t, InvoiceStatusPaid, invoice.Status)
		assert.NotNil(t, invoice.PaidDate)
	})
}

func createTestInvoice(t *testing.T) *Invoice {
	t.Helper()
	customerID := uuidv7.New()
	dueDate := time.Now().AddDate(0, 0, 30)
	invoice, err := NewInvoice(customerID, dueDate, "USD")
	require.NoError(t, err)
	return invoice
}
