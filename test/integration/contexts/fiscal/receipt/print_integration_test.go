package receipt_test

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	receiptPrinter "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/printer"
	receiptRepo "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/repository/postgres"
	receiptAggregate "github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	receiptUseCase "github.com/basilex/promenade/internal/contexts/fiscal/receipt/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestReceiptUseCase_PrintReceipt_PDF(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	outputDir := t.TempDir()

	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := receiptRepo.NewReceiptRepository(testDB.DB)
		printer := receiptPrinter.NewPDFPrinter(outputDir)
		uc := receiptUseCase.NewReceiptUseCase(repo, printer)

		cashRegisterID := uuidv7.New()
		orderID := uuidv7.New()
		createdBy := uuidv7.New()

		rec, err := receiptAggregate.NewReceipt(
			cashRegisterID,
			orderID,
			receiptAggregate.PaymentTypeCash,
			receiptAggregate.ReceiptTypeSale,
			"UAH",
			[]receiptAggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
			createdBy,
		)
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, rec))

		printedBy := uuidv7.New()
		printed, err := uc.PrintReceipt(ctx, rec.ID, printedBy)
		require.NoError(t, err)
		require.Equal(t, receiptAggregate.ReceiptStatusPrinted, printed.Status)
		require.NotEmpty(t, printed.FiscalNumber)
		require.NotEmpty(t, printed.FiscalURL)
		require.NotNil(t, printed.PrintedAt)

		info, err := os.Stat(printed.FiscalURL)
		require.NoError(t, err)
		require.False(t, info.IsDir())

		found, err := repo.GetByID(ctx, rec.ID)
		require.NoError(t, err)
		require.Equal(t, receiptAggregate.ReceiptStatusPrinted, found.Status)
		require.Equal(t, printed.FiscalNumber, found.FiscalNumber)
		require.Equal(t, printed.FiscalURL, found.FiscalURL)
	})
}
