package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	invoiceerrors "github.com/basilex/promenade/internal/contexts/billing/invoice"
	"github.com/basilex/promenade/internal/contexts/billing/invoice/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// Mock repository for testing
type mockRepository struct {
	invoices         map[string]*aggregate.Invoice
	invoicesByNumber map[string]*aggregate.Invoice
	nextInvoiceNo    int
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		invoices:         make(map[string]*aggregate.Invoice),
		invoicesByNumber: make(map[string]*aggregate.Invoice),
		nextInvoiceNo:    1,
	}
}

func (m *mockRepository) Create(ctx context.Context, invoice *aggregate.Invoice) error {
	m.invoices[invoice.ID.String()] = invoice
	m.invoicesByNumber[invoice.InvoiceNo] = invoice
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Invoice, error) {
	inv, ok := m.invoices[id.String()]
	if !ok {
		return nil, invoiceerrors.ErrInvoiceNotFound
	}
	return inv, nil
}

func (m *mockRepository) GetByInvoiceNo(ctx context.Context, invoiceNo string) (*aggregate.Invoice, error) {
	inv, ok := m.invoicesByNumber[invoiceNo]
	if !ok {
		return nil, invoiceerrors.ErrInvoiceNotFound
	}
	return inv, nil
}

func (m *mockRepository) Update(ctx context.Context, invoice *aggregate.Invoice) error {
	if _, ok := m.invoices[invoice.ID.String()]; !ok {
		return invoiceerrors.ErrInvoiceNotFound
	}
	m.invoices[invoice.ID.String()] = invoice
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if _, ok := m.invoices[id.String()]; !ok {
		return invoiceerrors.ErrInvoiceNotFound
	}
	delete(m.invoices, id.String())
	return nil
}

func (m *mockRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Invoice, int, error) {
	return nil, 0, nil
}

func (m *mockRepository) ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*aggregate.Invoice, error) {
	return nil, nil
}

func (m *mockRepository) ListByStatus(ctx context.Context, status aggregate.InvoiceStatus, page, pageSize int) ([]*aggregate.Invoice, int, error) {
	return nil, 0, nil
}

func (m *mockRepository) ListOverdue(ctx context.Context, page, pageSize int) ([]*aggregate.Invoice, int, error) {
	return nil, 0, nil
}

func (m *mockRepository) List(ctx context.Context, page, pageSize int) ([]*aggregate.Invoice, int, error) {
	invoices := make([]*aggregate.Invoice, 0, len(m.invoices))
	for _, inv := range m.invoices {
		invoices = append(invoices, inv)
	}
	return invoices, len(invoices), nil
}

func (m *mockRepository) CountByStatus(ctx context.Context, status aggregate.InvoiceStatus) (int, error) {
	return 0, nil
}

func (m *mockRepository) GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error) {
	return 0, nil
}

func (m *mockRepository) CreateLine(ctx context.Context, line *aggregate.InvoiceLine) error {
	return nil
}

func (m *mockRepository) DeleteLine(ctx context.Context, lineID uuidv7.UUID) error {
	return nil
}

func (m *mockRepository) GetLinesByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]aggregate.InvoiceLine, error) {
	return []aggregate.InvoiceLine{}, nil
}

func (m *mockRepository) ExistsByInvoiceNo(ctx context.Context, invoiceNo string) (bool, error) {
	_, ok := m.invoicesByNumber[invoiceNo]
	return ok, nil
}

func (m *mockRepository) GenerateInvoiceNumber(ctx context.Context) (string, error) {
	year := time.Now().Year()
	invoiceNo := testInvoiceNo(year, m.nextInvoiceNo)
	m.nextInvoiceNo++
	return invoiceNo, nil
}

func testInvoiceNo(year, seq int) string {
	return "INV-" + string(rune(year)) + "-" + string(rune(seq))
}

func TestUseCase_CreateInvoice(t *testing.T) {
	repo := newMockRepository()
	uc := NewInvoiceUseCase(repo)
	ctx := context.Background()

	customerID := uuidv7.New()
	dueDate := time.Now().Add(30 * 24 * time.Hour)

	t.Run("create invoice successfully", func(t *testing.T) {
		inv, err := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

		require.NoError(t, err)
		assert.NotNil(t, inv)
		assert.NotEqual(t, "", inv.InvoiceNo)
		assert.Equal(t, customerID, inv.CustomerID)
		assert.Equal(t, aggregate.InvoiceStatusDraft, inv.Status)
		assert.Equal(t, "USD", inv.Currency)
	})

	t.Run("create invoice with order ID", func(t *testing.T) {
		orderID := uuidv7.New()
		inv, err := uc.CreateInvoice(ctx, customerID, &orderID, dueDate, "EUR")

		require.NoError(t, err)
		assert.NotNil(t, inv.OrderID)
		assert.Equal(t, orderID, *inv.OrderID)
	})
}

func TestUseCase_AddLineItem(t *testing.T) {
	repo := newMockRepository()
	uc := NewInvoiceUseCase(repo)
	ctx := context.Background()

	customerID := uuidv7.New()
	dueDate := time.Now().Add(30 * 24 * time.Hour)
	inv, _ := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

	t.Run("add line item to draft invoice", func(t *testing.T) {
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		updatedInv, err := uc.AddLineItem(ctx, inv.ID, "Product A", 2, unitPrice)

		require.NoError(t, err)
		assert.Equal(t, 1, len(updatedInv.Lines))
		assert.Equal(t, int64(20000), updatedInv.SubtotalAmount.Amount) // 2 * $100
	})

	t.Run("cannot add line to sent invoice", func(t *testing.T) {
		// Mark invoice as sent
		_ = uc.SendInvoice(ctx, inv.ID)

		unitPrice, _ := valueobject.NewMoney(5000, "USD")
		_, err := uc.AddLineItem(ctx, inv.ID, "Product B", 1, unitPrice)

		assert.Error(t, err)
		assert.ErrorIs(t, err, invoiceerrors.ErrInvoiceCannotModifyNonDraft)
	})
}

func TestUseCase_SendInvoice(t *testing.T) {
	repo := newMockRepository()
	uc := NewInvoiceUseCase(repo)
	ctx := context.Background()

	customerID := uuidv7.New()
	dueDate := time.Now().Add(30 * 24 * time.Hour)
	inv, _ := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

	t.Run("cannot send invoice without line items", func(t *testing.T) {
		err := uc.SendInvoice(ctx, inv.ID)

		assert.Error(t, err)
	})

	t.Run("send invoice with line items", func(t *testing.T) {
		// Add line item first
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		_, _ = uc.AddLineItem(ctx, inv.ID, "Product A", 1, unitPrice)

		// Reload invoice (mock doesn't auto-reload)
		inv, _ = repo.GetByID(ctx, inv.ID)

		err := uc.SendInvoice(ctx, inv.ID)

		require.NoError(t, err)

		// Verify status changed
		sentInv, _ := repo.GetByID(ctx, inv.ID)
		assert.Equal(t, aggregate.InvoiceStatusSent, sentInv.Status)
	})
}

func TestUseCase_MarkAsPaid(t *testing.T) {
	repo := newMockRepository()
	uc := NewInvoiceUseCase(repo)
	ctx := context.Background()

	customerID := uuidv7.New()
	dueDate := time.Now().Add(30 * 24 * time.Hour)
	inv, _ := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

	// Add line item and send
	unitPrice, _ := valueobject.NewMoney(10000, "USD")
	_, _ = uc.AddLineItem(ctx, inv.ID, "Product A", 1, unitPrice)
	inv, _ = repo.GetByID(ctx, inv.ID)
	_ = uc.SendInvoice(ctx, inv.ID)

	t.Run("mark sent invoice as paid", func(t *testing.T) {
		paidDate := time.Now()
		err := uc.MarkAsPaid(ctx, inv.ID, paidDate)

		require.NoError(t, err)

		// Verify status and paid date
		paidInv, _ := repo.GetByID(ctx, inv.ID)
		assert.Equal(t, aggregate.InvoiceStatusPaid, paidInv.Status)
		assert.NotNil(t, paidInv.PaidDate)
	})
}

func TestUseCase_CancelInvoice(t *testing.T) {
	repo := newMockRepository()
	uc := NewInvoiceUseCase(repo)
	ctx := context.Background()

	customerID := uuidv7.New()
	dueDate := time.Now().Add(30 * 24 * time.Hour)

	t.Run("cancel draft invoice", func(t *testing.T) {
		inv, _ := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

		err := uc.CancelInvoice(ctx, inv.ID)

		require.NoError(t, err)

		cancelledInv, _ := repo.GetByID(ctx, inv.ID)
		assert.Equal(t, aggregate.InvoiceStatusCancelled, cancelledInv.Status)
	})

	t.Run("cannot cancel paid invoice", func(t *testing.T) {
		inv, _ := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

		// Add line, send, and pay
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		_, _ = uc.AddLineItem(ctx, inv.ID, "Product A", 1, unitPrice)
		inv, _ = repo.GetByID(ctx, inv.ID)
		_ = uc.SendInvoice(ctx, inv.ID)
		_ = uc.MarkAsPaid(ctx, inv.ID, time.Now())

		err := uc.CancelInvoice(ctx, inv.ID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, invoiceerrors.ErrInvoiceCannotCancelTerminalState)
	})
}

func TestUseCase_UpdateTaxAmount(t *testing.T) {
	repo := newMockRepository()
	uc := NewInvoiceUseCase(repo)
	ctx := context.Background()

	customerID := uuidv7.New()
	dueDate := time.Now().Add(30 * 24 * time.Hour)
	inv, _ := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

	// Add line item
	unitPrice, _ := valueobject.NewMoney(10000, "USD")
	_, _ = uc.AddLineItem(ctx, inv.ID, "Product A", 1, unitPrice)

	t.Run("update tax amount on draft invoice", func(t *testing.T) {
		inv, _ = repo.GetByID(ctx, inv.ID)

		taxAmount, _ := valueobject.NewMoney(1000, "USD") // $10 tax
		err := uc.UpdateTaxAmount(ctx, inv.ID, taxAmount)

		require.NoError(t, err)

		// Verify total was recalculated
		updatedInv, _ := repo.GetByID(ctx, inv.ID)
		assert.Equal(t, int64(1000), updatedInv.TaxAmount.Amount)
		assert.Equal(t, int64(11000), updatedInv.TotalAmount.Amount) // $100 + $10
	})
}

func TestUseCase_DeleteInvoice(t *testing.T) {
	repo := newMockRepository()
	uc := NewInvoiceUseCase(repo)
	ctx := context.Background()

	customerID := uuidv7.New()
	dueDate := time.Now().Add(30 * 24 * time.Hour)

	t.Run("delete draft invoice", func(t *testing.T) {
		inv, _ := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

		err := uc.DeleteInvoice(ctx, inv.ID)

		require.NoError(t, err)

		// Verify invoice was deleted
		_, err = repo.GetByID(ctx, inv.ID)
		assert.ErrorIs(t, err, invoiceerrors.ErrInvoiceNotFound)
	})

	t.Run("cannot delete sent invoice", func(t *testing.T) {
		inv, _ := uc.CreateInvoice(ctx, customerID, nil, dueDate, "USD")

		// Add line and send
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		_, _ = uc.AddLineItem(ctx, inv.ID, "Product A", 1, unitPrice)
		inv, _ = repo.GetByID(ctx, inv.ID)
		_ = uc.SendInvoice(ctx, inv.ID)

		err := uc.DeleteInvoice(ctx, inv.ID)

		assert.Error(t, err)
	})
}
