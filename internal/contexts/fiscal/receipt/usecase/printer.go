package usecase

import (
	"context"

	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
)

// PrintResult contains fiscal data returned by printer providers
type PrintResult struct {
	ProviderReceiptID string
	FiscalNumber      string
	FiscalURL         string
	QRCode            string
}

// IPrinter defines an external fiscal printer provider
type IPrinter interface {
	Print(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error)
}

// ICancelPrinter defines optional cancel capability for external printers
type ICancelPrinter interface {
	Cancel(ctx context.Context, rec *aggregate.Receipt, reason string) error
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
func (m *MultiPrinter) Print(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error) {
	if m == nil || m.primary == nil {
		return nil, receipterrors.ErrReceiptPrintFailed
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

// Cancel cancels a receipt via the primary printer when supported.
func (m *MultiPrinter) Cancel(ctx context.Context, rec *aggregate.Receipt, reason string) error {
	if m == nil || m.primary == nil {
		return receipterrors.ErrReceiptCancelFailed
	}

	primary, ok := m.primary.(ICancelPrinter)
	if !ok {
		return receipterrors.ErrReceiptCancelFailed
	}

	if err := primary.Cancel(ctx, rec, reason); err != nil {
		return err
	}

	for _, printer := range m.secondary {
		if printer == nil {
			continue
		}
		if canceler, ok := printer.(ICancelPrinter); ok {
			_ = canceler.Cancel(ctx, rec, reason)
		}
	}

	return nil
}
