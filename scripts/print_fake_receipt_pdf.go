package main

import (
	"context"
	"fmt"
	"log"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	receiptPrinter "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/printer"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func main() {
	printer := receiptPrinter.NewPDFPrinter("tmp/fiscal/receipts")

	rec, err := receipt.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		receipt.PaymentTypeCard,
		receipt.ReceiptTypeSale,
		"UAH",
		[]receipt.ReceiptLine{
			{Name: "Espresso", Quantity: 2, PriceCents: 8500, TaxRate: 20},
			{Name: "Croissant", Quantity: 1, PriceCents: 6500, TaxRate: 20},
			{Name: "Orange Juice", Quantity: 1, PriceCents: 9000, TaxRate: 20},
			{Name: "Chocolate Bar", Quantity: 3, PriceCents: 2500, TaxRate: 20},
		},
		uuidv7.New(),
	)
	if err != nil {
		log.Fatalf("failed to build receipt: %v", err)
	}

	result, err := printer.Print(context.Background(), rec)
	if err != nil {
		log.Fatalf("failed to print receipt: %v", err)
	}

	fmt.Printf("PDF receipt generated: %s\n", result.FiscalURL)
}