package receipt_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	receiptRepo "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func createTestReceipt(cashRegisterID, orderID uuidv7.UUID) *receipt.Receipt {
	createdBy := uuidv7.New()
	lines := []receipt.ReceiptLine{
		{
			Name:       "Test Item",
			Quantity:   2,
			PriceCents: 1000,
			TaxRate:    20,
		},
	}

	rec, _ := receipt.NewReceipt(
		cashRegisterID,
		orderID,
		receipt.PaymentTypeCash,
		receipt.ReceiptTypeSale,
		"UAH",
		lines,
		createdBy,
	)
	return rec
}

func TestReceiptRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()
		rec := createTestReceipt(cashRegisterID, orderID)

		err := repo.Create(ctx, rec)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, rec.ID, found.ID)
		assert.Equal(t, rec.OrderID, found.OrderID)
		assert.Equal(t, rec.CashRegisterID, found.CashRegisterID)
		assert.Equal(t, receipt.ReceiptStatusPending, found.Status)
		assert.Equal(t, rec.TotalAmount, found.TotalAmount)
		assert.Equal(t, rec.TaxAmount, found.TaxAmount)
	})
}

func TestReceiptRepository_Create_DuplicateOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()

		rec1 := createTestReceipt(cashRegisterID, orderID)
		rec2 := createTestReceipt(cashRegisterID, orderID)

		require.NoError(t, repo.Create(ctx, rec1))
		err := repo.Create(ctx, rec2)
		assert.Error(t, err)
		assert.Equal(t, receipt.ErrReceiptCreateFailed, err)
	})
}

func TestReceiptRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()
		rec := createTestReceipt(cashRegisterID, orderID)

		require.NoError(t, repo.Create(ctx, rec))

		found, err := repo.GetByID(ctx, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, rec.ID, found.ID)

		notFound, err := repo.GetByID(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.Equal(t, receipt.ErrReceiptNotFound, err)
		assert.Nil(t, notFound)
	})
}

func TestReceiptRepository_GetByOrderID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()
		rec := createTestReceipt(cashRegisterID, orderID)

		require.NoError(t, repo.Create(ctx, rec))

		found, err := repo.GetByOrderID(ctx, orderID)
		require.NoError(t, err)
		assert.Equal(t, rec.ID, found.ID)
		assert.Equal(t, orderID, found.OrderID)

		notFound, err := repo.GetByOrderID(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.Equal(t, receipt.ErrReceiptNotFound, err)
		assert.Nil(t, notFound)
	})
}

func TestReceiptRepository_Update_MarkPrinted(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()
		rec := createTestReceipt(cashRegisterID, orderID)

		require.NoError(t, repo.Create(ctx, rec))

		printedBy := uuidv7.New()
		require.NoError(t, rec.MarkPrinted("FN-12345", "https://fiscal.test/receipt", "qr-data", printedBy))
		err := repo.Update(ctx, rec)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, receipt.ReceiptStatusPrinted, found.Status)
		assert.Equal(t, "FN-12345", found.FiscalNumber)
		assert.Equal(t, "https://fiscal.test/receipt", found.FiscalURL)
		assert.Equal(t, "qr-data", found.QRCode)
		require.NotNil(t, found.PrintedAt)
	})
}

func TestReceiptRepository_Update_Cancel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()
		rec := createTestReceipt(cashRegisterID, orderID)

		require.NoError(t, repo.Create(ctx, rec))

		cancelledBy := uuidv7.New()
		require.NoError(t, rec.Cancel("Customer request", cancelledBy))
		err := repo.Update(ctx, rec)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, rec.ID)
		require.NoError(t, err)
		assert.Equal(t, receipt.ReceiptStatusCancelled, found.Status)
		assert.Equal(t, "Customer request", found.CancellationReason)
		require.NotNil(t, found.CancelledAt)
	})
}

func TestReceiptRepository_Update_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()
		rec := createTestReceipt(cashRegisterID, orderID)

		err := repo.Update(ctx, rec)
		assert.Error(t, err)
		assert.Equal(t, receipt.ErrReceiptNotFound, err)
	})
}

func TestReceiptRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()
		rec := createTestReceipt(cashRegisterID, orderID)

		require.NoError(t, repo.Create(ctx, rec))

		err := repo.Delete(ctx, rec.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, rec.ID)
		assert.Error(t, err)
		assert.Equal(t, receipt.ErrReceiptNotFound, err)
	})
}

func TestReceiptRepository_Delete_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		err := repo.Delete(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.Equal(t, receipt.ErrReceiptNotFound, err)
	})
}

func TestReceiptRepository_List_WithFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		cashRegisterID := uuidv7.New()

		rec1 := createTestReceipt(cashRegisterID, uuidv7.New())
		rec2 := createTestReceipt(cashRegisterID, uuidv7.New())

		require.NoError(t, repo.Create(ctx, rec1))
		require.NoError(t, repo.Create(ctx, rec2))

		printedBy := uuidv7.New()
		require.NoError(t, rec2.MarkPrinted("FN-PRINTED", "https://fiscal.test/printed", "qr", printedBy))
		require.NoError(t, repo.Update(ctx, rec2))

		filters := &receipt.ListFilters{CashRegisterID: &cashRegisterID}
		all, err := repo.List(ctx, filters)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 2)

		status := receipt.ReceiptStatusPrinted
		printed := &receipt.ListFilters{CashRegisterID: &cashRegisterID, Status: &status}
		printedOnly, err := repo.List(ctx, printed)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(printedOnly), 1)

		orderID := rec1.OrderID
		byOrder := &receipt.ListFilters{OrderID: &orderID}
		orderResults, err := repo.List(ctx, byOrder)
		require.NoError(t, err)
		require.Len(t, orderResults, 1)
		assert.Equal(t, rec1.ID, orderResults[0].ID)
	})
}
