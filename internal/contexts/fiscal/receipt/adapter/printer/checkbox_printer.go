package printer

import (
	"context"
	"fmt"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/pkg/fiscal/checkbox"
)

// CheckboxPrinter prints receipts via Checkbox API
type CheckboxPrinter struct {
	client *checkbox.Client
}

// NewCheckboxPrinter creates a Checkbox printer
func NewCheckboxPrinter(client *checkbox.Client) *CheckboxPrinter {
	return &CheckboxPrinter{client: client}
}

// Print prints a receipt via Checkbox and returns fiscal data
func (p *CheckboxPrinter) Print(ctx context.Context, rec *receipt.Receipt) (*receipt.PrintResult, error) {
	if p.client == nil {
		return nil, fmt.Errorf("checkbox client not configured")
	}

	paymentType, err := checkbox.ConvertPaymentType(string(rec.PaymentType))
	if err != nil {
		return nil, err
	}

	lines := rec.Lines.Get()
	goods := make([]checkbox.ReceiptGood, len(lines))
	for i, line := range lines {
		price := int(line.PriceCents)
		goods[i] = checkbox.ReceiptGood{
			Good: checkbox.Good{
				Name:  line.Name,
				Price: price,
				Tax:   []int{line.TaxRate},
			},
			Quantity: line.Quantity,
		}
	}

	req := &checkbox.Receipt{
		Goods:   goods,
		Payment: checkbox.Payment{Type: paymentType, Value: int(rec.TotalAmount)},
	}

	resp, err := p.client.CreateReceipt(ctx, req)
	if err != nil {
		return nil, err
	}

	return &receipt.PrintResult{
		FiscalNumber: resp.FiscalCode,
		FiscalURL:    resp.FiscalURL,
		QRCode:       resp.QRCodeURL,
	}, nil
}