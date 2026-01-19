package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type stubPrinter struct {
	result *PrintResult
	err    error
	calls  int
}

func (s *stubPrinter) Print(ctx context.Context, rec *aggregate.Receipt) (*PrintResult, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func TestMultiPrinter_PrimaryNil(t *testing.T) {
	printer := NewMultiPrinter(nil)
	_, err := printer.Print(context.Background(), &aggregate.Receipt{})
	require.ErrorIs(t, err, receipterrors.ErrReceiptPrintFailed)
}

func TestMultiPrinter_NilReceiver(t *testing.T) {
	var printer *MultiPrinter
	_, err := printer.Print(context.Background(), &aggregate.Receipt{})
	require.ErrorIs(t, err, receipterrors.ErrReceiptPrintFailed)
}

func TestMultiPrinter_PrimaryError(t *testing.T) {
	primary := &stubPrinter{err: errors.New("primary failed")}
	printer := NewMultiPrinter(primary)

	_, err := printer.Print(context.Background(), &aggregate.Receipt{})
	require.Error(t, err)
	require.Equal(t, 1, primary.calls)
}

func TestMultiPrinter_SecondaryBestEffort(t *testing.T) {
	primary := &stubPrinter{result: &PrintResult{FiscalNumber: "F-1"}}
	secondary := &stubPrinter{err: errors.New("secondary failed")}

	printer := NewMultiPrinter(primary, secondary)

	rec, err := aggregate.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	require.NoError(t, err)

	result, err := printer.Print(context.Background(), rec)
	require.NoError(t, err)
	require.Equal(t, "F-1", result.FiscalNumber)
	require.Equal(t, 1, primary.calls)
	require.Equal(t, 1, secondary.calls)
}
