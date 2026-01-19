package printer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/usecase"
)

// PDFPrinter writes receipt data to a PDF file in the configured directory.
type PDFPrinter struct {
	outputDir string
}

// NewPDFPrinter creates a new PDF printer.
func NewPDFPrinter(outputDir string) *PDFPrinter {
	return &PDFPrinter{outputDir: outputDir}
}

// Print writes a receipt PDF and returns a print result referencing the file.
func (p *PDFPrinter) Print(ctx context.Context, rec *aggregate.Receipt) (*usecase.PrintResult, error) {
	if p == nil || p.outputDir == "" {
		return nil, fmt.Errorf("pdf output directory not configured")
	}

	if err := os.MkdirAll(p.outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	fileName := fmt.Sprintf("receipt_%s_%s.pdf", rec.GetID().String(), time.Now().Format("20060102_150405"))
	filePath := filepath.Join(p.outputDir, fileName)

	if err := writeReceiptPDF(rec, filePath); err != nil {
		return nil, err
	}

	return &usecase.PrintResult{
		FiscalNumber: fmt.Sprintf("PDF-%s", rec.GetID().String()),
		FiscalURL:    filePath,
		QRCode:       "",
	}, nil
}

func writeReceiptPDF(rec *aggregate.Receipt, filePath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Fiscal Receipt", false)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(40, 10, "Fiscal Receipt")
	pdf.Ln(12)

	pdf.SetFont("Helvetica", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Receipt ID: %s", rec.GetID().String()))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Cash Register ID: %s", rec.CashRegisterID.String()))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Order ID: %s", rec.OrderID.String()))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Payment Type: %s", rec.PaymentType))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Receipt Type: %s", rec.ReceiptType))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Currency: %s", rec.Currency))
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 7, "Lines")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 10)
	lines := rec.Lines.Get()
	for _, line := range lines {
		pdf.Cell(0, 5, fmt.Sprintf("%s | Qty: %d | Price: %d | Tax: %d%% | Total: %d", line.Name, line.Quantity, line.PriceCents, line.TaxRate, line.TotalCents))
		pdf.Ln(5)
	}

	pdf.Ln(6)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 7, fmt.Sprintf("Total: %d", rec.TotalAmount))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Tax: %d", rec.TaxAmount))

	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return fmt.Errorf("failed to write pdf: %w", err)
	}

	return nil
}
