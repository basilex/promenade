package printer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestPDFPrinter_Print_Success(t *testing.T) {
	outputDir := t.TempDir()
	printer := NewPDFPrinter(outputDir)

	rec, err := receipt.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		receipt.PaymentTypeCard,
		receipt.ReceiptTypeSale,
		"UAH",
		[]receipt.ReceiptLine{
			{Name: "Item A", Quantity: 2, PriceCents: 1000, TaxRate: 20},
			{Name: "Item B", Quantity: 1, PriceCents: 2500, TaxRate: 20},
			{Name: "Item C", Quantity: 3, PriceCents: 500, TaxRate: 20},
		},
		uuidv7.New(),
	)
	require.NoError(t, err)

	result, err := printer.Print(context.Background(), rec)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, strings.HasPrefix(result.FiscalNumber, "PDF-"))
	require.True(t, strings.HasSuffix(result.FiscalURL, ".pdf"))

	info, err := os.Stat(result.FiscalURL)
	require.NoError(t, err)
	require.False(t, info.IsDir())
	require.Equal(t, outputDir, filepath.Dir(result.FiscalURL))
}

func TestPDFPrinter_Print_EmptyOutputDir(t *testing.T) {
	printer := NewPDFPrinter("")

	rec, err := receipt.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		receipt.PaymentTypeCash,
		receipt.ReceiptTypeSale,
		"UAH",
		[]receipt.ReceiptLine{
			{Name: "Item A", Quantity: 1, PriceCents: 1000, TaxRate: 20},
		},
		uuidv7.New(),
	)
	require.NoError(t, err)

	result, err := printer.Print(context.Background(), rec)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestPDFPrinter_Print_NilPrinter(t *testing.T) {
	var printer *PDFPrinter

	rec, err := receipt.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		receipt.PaymentTypeCash,
		receipt.ReceiptTypeSale,
		"UAH",
		[]receipt.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	require.NoError(t, err)

	result, err := printer.Print(context.Background(), rec)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestPDFPrinter_Print_MkdirError(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "output.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0o600))

	printer := NewPDFPrinter(filePath)

	rec, err := receipt.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		receipt.PaymentTypeCash,
		receipt.ReceiptTypeSale,
		"UAH",
		[]receipt.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	require.NoError(t, err)

	result, err := printer.Print(context.Background(), rec)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestWriteReceiptPDF_Error(t *testing.T) {
	rec, err := receipt.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		receipt.PaymentTypeCash,
		receipt.ReceiptTypeSale,
		"UAH",
		[]receipt.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	require.NoError(t, err)

	outputDir := t.TempDir()
	filePath := filepath.Join(outputDir, "missing", "receipt.pdf")

	err = writeReceiptPDF(rec, filePath)
	require.Error(t, err)
}
