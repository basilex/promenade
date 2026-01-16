package receipt

import "context"

// PrintResult contains fiscal data returned by printer providers
type PrintResult struct {
	FiscalNumber string
	FiscalURL    string
	QRCode       string
}

// IPrinter defines an external fiscal printer provider
type IPrinter interface {
	Print(ctx context.Context, rec *Receipt) (*PrintResult, error)
}

// MultiPrinter executes a primary printer and optional side printers.
// The primary printer result is returned; secondary printers run best-effort.
type MultiPrinter struct {
	primary   IPrinter
	secondary []IPrinter
}

// NewMultiPrinter creates a new multi-printer.
func NewMultiPrinter(primary IPrinter, secondary ...IPrinter) *MultiPrinter {
	return &MultiPrinter{primary: primary, secondary: secondary}
}

// Print executes primary printer and then secondary printers best-effort.
func (m *MultiPrinter) Print(ctx context.Context, rec *Receipt) (*PrintResult, error) {
	if m == nil || m.primary == nil {
		return nil, ErrReceiptPrintFailed
	}

	result, err := m.primary.Print(ctx, rec)
	if err != nil {
		return nil, err
	}

	for _, printer := range m.secondary {
		if printer == nil {
			continue
		}
		_, _ = printer.Print(ctx, rec)
	}

	return result, nil
}