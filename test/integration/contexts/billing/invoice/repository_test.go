package invoice_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/billing/invoice"
	"github.com/basilex/promenade/internal/contexts/billing/invoice/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/integration"
)

func TestInvoiceRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInvoiceRepository(testDB.DB)
		customerID := uuidv7.New()

		// Create customer first (required for foreign key)
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, email, status, tier) VALUES ($1, $2, $3, $4)`,
			customerID, "customer_"+customerID.String()+"@test.com", "customer", "free")
		require.NoError(t, err)

		// Create invoice
		inv, err := invoice.NewInvoice(customerID, time.Now().Add(30*24*time.Hour), "USD")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, inv))
		assert.NotEqual(t, uuidv7.UUID{}, inv.ID)

		// GetByID
		retrieved, err := repo.GetByID(ctx, inv.ID)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, retrieved.ID)
		assert.Equal(t, customerID, retrieved.CustomerID)
		assert.Equal(t, "USD", retrieved.Currency)

		// GetByInvoiceNo
		byNumber, err := repo.GetByInvoiceNo(ctx, inv.InvoiceNo)
		require.NoError(t, err)
		assert.Equal(t, inv.ID, byNumber.ID)

		// Update
		money, _ := valueobject.NewMoney(5000, "USD")
		inv.UpdateTaxAmount(money)
		require.NoError(t, repo.Update(ctx, inv))
		updated, _ := repo.GetByID(ctx, inv.ID)
		assert.Equal(t, int64(5000), updated.TaxAmount.Amount)

		// Delete
		require.NoError(t, repo.Delete(ctx, inv.ID))
		_, err = repo.GetByID(ctx, inv.ID)
		assert.Error(t, err)
	})
}

func TestInvoiceRepository_ListByCustomer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInvoiceRepository(testDB.DB)
		customerID := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, email, status, tier) VALUES ($1, $2, $3, $4)`,
			customerID, "customer_"+customerID.String()+"@test.com", "customer", "free")
		require.NoError(t, err)

		// Create 3 invoices
		for i := 0; i < 3; i++ {
			inv, _ := invoice.NewInvoice(customerID, time.Now().Add(30*24*time.Hour), "USD")
			require.NoError(t, repo.Create(ctx, inv))
		}

		// List by customer
		invoices, total, err := repo.ListByCustomer(ctx, customerID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, invoices, 3)
		assert.Equal(t, 3, total)
	})
}

func TestInvoiceRepository_ListByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInvoiceRepository(testDB.DB)
		customerID := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, email, status, tier) VALUES ($1, $2, $3, $4)`,
			customerID, "customer_"+customerID.String()+"@test.com", "customer", "free")
		require.NoError(t, err)

		// Create draft invoice
		inv1, _ := invoice.NewInvoice(customerID, time.Now().Add(30*24*time.Hour), "USD")
		require.NoError(t, repo.Create(ctx, inv1))

		// Create and send invoice
		inv2, _ := invoice.NewInvoice(customerID, time.Now().Add(30*24*time.Hour), "USD")
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		inv2.AddLine("Test Item", 1, unitPrice)
		inv2.MarkAsSent()
		require.NoError(t, repo.Create(ctx, inv2))

		// List by status
		drafts, total, err := repo.ListByStatus(ctx, invoice.InvoiceStatusDraft, 1, 10)
		require.NoError(t, err)
		assert.Len(t, drafts, 1)
		assert.Equal(t, 1, total)

		sent, total, err := repo.ListByStatus(ctx, invoice.InvoiceStatusSent, 1, 10)
		require.NoError(t, err)
		assert.Len(t, sent, 1)
		assert.Equal(t, 1, total)
	})
}

func TestInvoiceRepository_CountByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInvoiceRepository(testDB.DB)
		customerID := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, email, status, tier) VALUES ($1, $2, $3, $4)`,
			customerID, "customer_"+customerID.String()+"@test.com", "customer", "free")
		require.NoError(t, err)

		// Create 2 draft invoices
		for i := 0; i < 2; i++ {
			inv, _ := invoice.NewInvoice(customerID, time.Now().Add(30*24*time.Hour), "USD")
			require.NoError(t, repo.Create(ctx, inv))
		}

		// Count drafts
		count, err := repo.CountByStatus(ctx, invoice.InvoiceStatusDraft)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}

func TestInvoiceRepository_GetTotalRevenue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewInvoiceRepository(testDB.DB)
		customerID := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, email, status, tier) VALUES ($1, $2, $3, $4)`,
			customerID, "customer_"+customerID.String()+"@test.com", "customer", "free")
		require.NoError(t, err)

		// Create and pay invoice
		inv, _ := invoice.NewInvoice(customerID, time.Now().Add(30*24*time.Hour), "USD")
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		inv.AddLine("Test Item", 1, unitPrice)
		inv.MarkAsSent()
		paidDate := time.Now()
		inv.MarkAsPaid(paidDate)
		require.NoError(t, repo.Create(ctx, inv))

		// Get total revenue
		from := time.Now().Add(-24 * time.Hour)
		to := time.Now().Add(24 * time.Hour)
		total, err := repo.GetTotalRevenue(ctx, from, to)
		require.NoError(t, err)
		assert.Equal(t, int64(10000), total)
	})
}
