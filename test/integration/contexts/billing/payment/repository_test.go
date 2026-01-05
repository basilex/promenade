package payment_test

import (
	"fmt"
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/billing/payment"
	"github.com/basilex/promenade/internal/contexts/billing/payment/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/integration"
)

func TestPaymentRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()

		// Create customer first (required for foreign key)
		assignedTo := uuidv7.New() // Sales rep ID (mock)
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create payment
		amount := valueobject.Money{Amount: 10000, Currency: "USD"}
		pmt, err := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, pmt))
		assert.NotEqual(t, uuidv7.UUID{}, pmt.GetID())

		// GetByID
		retrieved, err := repo.GetByID(ctx, pmt.GetID())
		require.NoError(t, err)
		assert.Equal(t, pmt.GetID(), retrieved.GetID())
		assert.Equal(t, customerID, retrieved.CustomerID)
		assert.Equal(t, int64(10000), retrieved.Amount.Amount)
		assert.Equal(t, "USD", retrieved.Amount.Currency)
		assert.Equal(t, payment.PaymentMethodCreditCard, retrieved.Method)
		assert.Equal(t, payment.PaymentStatusPending, retrieved.Status)

		// Update - process and complete
		require.NoError(t, pmt.Process())
		require.NoError(t, pmt.Complete("txn_123456"))
		require.NoError(t, repo.Update(ctx, pmt))
		updated, _ := repo.GetByID(ctx, pmt.GetID())
		assert.Equal(t, payment.PaymentStatusCompleted, updated.Status)
		assert.Equal(t, "txn_123456", updated.TransactionID)

		// GetByTransactionID
		byTxn, err := repo.GetByTransactionID(ctx, "txn_123456")
		require.NoError(t, err)
		assert.Equal(t, pmt.GetID(), byTxn.GetID())

		// Delete
		require.NoError(t, repo.Delete(ctx, pmt.GetID()))
		_, err = repo.GetByID(ctx, pmt.GetID())
		assert.Error(t, err)
	})
}

func TestPaymentRepository_ListByCustomer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create 3 payments with unique payment_no
		amount := valueobject.Money{Amount: 10000, Currency: "USD"}
		for i := 0; i < 3; i++ {
			pmt, _ := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
			pmt.PaymentNo = fmt.Sprintf("PAY-%s-%d", uuidv7.New().String()[:8], i)
			require.NoError(t, repo.Create(ctx, pmt))
		}

		// List payments
		payments, err := repo.ListByCustomerID(ctx, customerID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, payments, 3)
	})
}

func TestPaymentRepository_ListByInvoice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()
		invoiceID := uuidv7.New()
		assignedTo := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create invoice
		_, err = tx.ExecContext(ctx, `
			INSERT INTO billing_invoices (id, invoice_no, customer_id, subtotal_amount, tax_amount, total_amount, currency, issue_date, due_date, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`,
			invoiceID, "INV-001", customerID, 10000, 0, 10000, "USD", "2026-01-01", "2026-01-31", "pending")
		require.NoError(t, err)

		// Create 2 payments linked to invoice (unique payment_no)
		amount := valueobject.Money{Amount: 5000, Currency: "USD"}
		for i := 0; i < 2; i++ {
			pmt, _ := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
			pmt.PaymentNo = fmt.Sprintf("PAY-%s-%d", uuidv7.New().String()[:8], i)
			_ = pmt.LinkToInvoice(invoiceID)
			require.NoError(t, repo.Create(ctx, pmt))
		}

		// List payments by invoice
		payments, err := repo.ListByInvoiceID(ctx, invoiceID)
		require.NoError(t, err)
		assert.Len(t, payments, 2)
		for _, p := range payments {
			require.NotNil(t, p.InvoiceID)
			assert.Equal(t, invoiceID, *p.InvoiceID)
		}
	})
}

func TestPaymentRepository_ListByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create 2 pending payments (unique payment_no, max 20 chars)
		amount := valueobject.Money{Amount: 10000, Currency: "USD"}
		for i := 0; i < 2; i++ {
			pmt, _ := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
			pmt.PaymentNo = fmt.Sprintf("PAY-%s-%d", uuidv7.New().String()[:6], i)
			require.NoError(t, repo.Create(ctx, pmt))
		}

		// Create 1 completed payment (unique payment_no, max 20 chars)
		pmt, _ := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
		pmt.PaymentNo = fmt.Sprintf("PAY-%s-C", uuidv7.New().String()[:6])
		_ = pmt.Process()
		_ = pmt.Complete("txn_123")
		require.NoError(t, repo.Create(ctx, pmt))

		// List pending payments
		pendingPayments, err := repo.ListByStatus(ctx, payment.PaymentStatusPending, 1, 10)
		require.NoError(t, err)
		assert.Len(t, pendingPayments, 2)

		// List completed payments
		completedPayments, err := repo.ListByStatus(ctx, payment.PaymentStatusCompleted, 1, 10)
		require.NoError(t, err)
		assert.Len(t, completedPayments, 1)
	})
}

func TestPaymentRepository_CountByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create 3 completed payments (unique payment_no)
		amount := valueobject.Money{Amount: 10000, Currency: "USD"}
		for i := 0; i < 3; i++ {
			pmt, _ := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
			pmt.PaymentNo = fmt.Sprintf("PAY-%s-%d", uuidv7.New().String()[:8], i)
			_ = pmt.Process()
			_ = pmt.Complete("txn-" + uuidv7.New().String()[:8])
			require.NoError(t, repo.Create(ctx, pmt))
		}

		// Get total for customer (only completed payments)
		count, err := repo.GetTotalByCustomer(ctx, customerID)
		require.NoError(t, err)
		assert.Equal(t, int64(30000), count) // 3 payments × 10000
	})
}

func TestPaymentRepository_Refund(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create and complete payment
		amount := valueobject.Money{Amount: 10000, Currency: "USD"}
		pmt, _ := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
		_ = pmt.Process()
		_ = pmt.Complete("txn_123")
		require.NoError(t, repo.Create(ctx, pmt))

		// Refund partial amount
		refundAmount := valueobject.Money{Amount: 5000, Currency: "USD"}
		require.NoError(t, pmt.Refund(refundAmount))
		require.NoError(t, repo.Update(ctx, pmt))

		// Verify refund
		refunded, err := repo.GetByID(ctx, pmt.GetID())
		require.NoError(t, err)
		assert.Equal(t, payment.PaymentStatusRefunded, refunded.Status)
		require.NotNil(t, refunded.RefundedAmount)
		assert.Equal(t, int64(5000), refunded.RefundedAmount.Amount)
		require.NotNil(t, refunded.RefundedAt)
	})
}

func TestPaymentRepository_GetByPaymentNo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create payment
		amount := valueobject.Money{Amount: 10000, Currency: "USD"}
		pmt, _ := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
		require.NoError(t, repo.Create(ctx, pmt))

		// GetByPaymentNo
		retrieved, err := repo.GetByPaymentNo(ctx, pmt.PaymentNo)
		require.NoError(t, err)
		assert.Equal(t, pmt.GetID(), retrieved.GetID())
		assert.Equal(t, pmt.PaymentNo, retrieved.PaymentNo)
	})
}

func TestPaymentRepository_GetTotalByInvoice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New()
		invoiceID := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create invoice
		_, err = tx.ExecContext(ctx, `
			INSERT INTO billing_invoices (id, invoice_no, customer_id, subtotal_amount, tax_amount, total_amount, currency, issue_date, due_date, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`,
			invoiceID, "INV-001", customerID, 30000, 0, 30000, "USD", "2026-01-01", "2026-01-31", "pending")
		require.NoError(t, err)

		// Create 3 completed payments linked to invoice with different amounts (unique payment_no)
		amounts := []int64{5000, 10000, 15000}
		for i, amt := range amounts {
			pmt, _ := payment.NewPayment(customerID, valueobject.Money{Amount: amt, Currency: "USD"}, payment.PaymentMethodCreditCard)
			pmt.PaymentNo = fmt.Sprintf("PAY-%s-%d", uuidv7.New().String()[:8], i)
			_ = pmt.Process()
			_ = pmt.Complete("txn-" + uuidv7.New().String()[:8])
			_ = pmt.LinkToInvoice(invoiceID)
			require.NoError(t, repo.Create(ctx, pmt))
		}

		// Get total by invoice (only completed payments)
		total, err := repo.GetTotalByInvoice(ctx, invoiceID)
		require.NoError(t, err)
		assert.Equal(t, int64(30000), total) // 5000 + 10000 + 15000
	})
}

func TestPaymentRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewPaymentRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New()

		// Create customer
		_, err := tx.ExecContext(ctx, `INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		// Create 5 payments (unique payment_no)
		amount := valueobject.Money{Amount: 10000, Currency: "USD"}
		for i := 0; i < 5; i++ {
			pmt, _ := payment.NewPayment(customerID, amount, payment.PaymentMethodCreditCard)
			pmt.PaymentNo = fmt.Sprintf("PAY-%s-%d", uuidv7.New().String()[:8], i)
			require.NoError(t, repo.Create(ctx, pmt))
		}

		// List with pagination
		page1, err := repo.List(ctx, 1, 3)
		require.NoError(t, err)
		assert.Len(t, page1, 3)

		page2, err := repo.List(ctx, 2, 3)
		require.NoError(t, err)
		assert.Len(t, page2, 2) // Remaining 2 payments
	})
}
