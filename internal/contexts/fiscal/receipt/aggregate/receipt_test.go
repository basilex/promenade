package aggregate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNewReceipt(t *testing.T) {
	cashRegisterID := uuidv7.New()
	orderID := uuidv7.New()
	createdBy := uuidv7.New()

	lines := []ReceiptLine{
		{Name: "Product A", Quantity: 2, PriceCents: 1000, TaxRate: 20},
	}

	tests := []struct {
		name           string
		cashRegisterID uuidv7.UUID
		orderID        uuidv7.UUID
		paymentType    PaymentType
		receiptType    ReceiptType
		currency       string
		lines          []ReceiptLine
		createdBy      uuidv7.UUID
		wantErr        error
	}{
		{
			name:           "valid receipt",
			cashRegisterID: cashRegisterID,
			orderID:        orderID,
			paymentType:    PaymentTypeCash,
			receiptType:    ReceiptTypeSale,
			currency:       "UAH",
			lines:          lines,
			createdBy:      createdBy,
			wantErr:        nil,
		},
		{
			name:           "missing cash register id",
			cashRegisterID: uuidv7.Nil,
			orderID:        orderID,
			paymentType:    PaymentTypeCash,
			receiptType:    ReceiptTypeSale,
			currency:       "UAH",
			lines:          lines,
			createdBy:      createdBy,
			wantErr:        receipterrors.ErrCashRegisterIDRequired,
		},
		{
			name:           "missing order id",
			cashRegisterID: cashRegisterID,
			orderID:        uuidv7.Nil,
			paymentType:    PaymentTypeCash,
			receiptType:    ReceiptTypeSale,
			currency:       "UAH",
			lines:          lines,
			createdBy:      createdBy,
			wantErr:        receipterrors.ErrOrderIDRequired,
		},
		{
			name:           "missing payment type",
			cashRegisterID: cashRegisterID,
			orderID:        orderID,
			paymentType:    "",
			receiptType:    ReceiptTypeSale,
			currency:       "UAH",
			lines:          lines,
			createdBy:      createdBy,
			wantErr:        receipterrors.ErrPaymentTypeRequired,
		},
		{
			name:           "missing receipt type",
			cashRegisterID: cashRegisterID,
			orderID:        orderID,
			paymentType:    PaymentTypeCash,
			receiptType:    "",
			currency:       "UAH",
			lines:          lines,
			createdBy:      createdBy,
			wantErr:        receipterrors.ErrReceiptTypeRequired,
		},
		{
			name:           "missing currency",
			cashRegisterID: cashRegisterID,
			orderID:        orderID,
			paymentType:    PaymentTypeCash,
			receiptType:    ReceiptTypeSale,
			currency:       "",
			lines:          lines,
			createdBy:      createdBy,
			wantErr:        receipterrors.ErrCurrencyRequired,
		},
		{
			name:           "missing created by",
			cashRegisterID: cashRegisterID,
			orderID:        orderID,
			paymentType:    PaymentTypeCash,
			receiptType:    ReceiptTypeSale,
			currency:       "UAH",
			lines:          lines,
			createdBy:      uuidv7.Nil,
			wantErr:        receipterrors.ErrCreatedByRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := NewReceipt(tt.cashRegisterID, tt.orderID, tt.paymentType, tt.receiptType, tt.currency, tt.lines, tt.createdBy)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, rec)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, rec)
			assert.NotEqual(t, uuidv7.Nil, rec.GetID())
			assert.Equal(t, ReceiptStatusPending, rec.Status)
			assert.Equal(t, int64(2000), rec.TotalAmount)
			assert.Equal(t, int64(400), rec.TaxAmount)
		})
	}
}

func TestReceipt_MarkPrinted(t *testing.T) {
	rec := buildReceipt(t)
	printedBy := uuidv7.New()

	err := rec.MarkPrinted("FN-123", "https://fiscal.example/receipt", "qr", printedBy)
	require.NoError(t, err)
	assert.Equal(t, ReceiptStatusPrinted, rec.Status)
	assert.Equal(t, "FN-123", rec.FiscalNumber)
	assert.NotNil(t, rec.PrintedAt)
	assert.Equal(t, printedBy, rec.LastUpdatedBy)

	err = rec.MarkPrinted("FN-456", "", "", printedBy)
	assert.ErrorIs(t, err, receipterrors.ErrReceiptAlreadyPrinted)
}

func TestReceipt_MarkPrinted_Errors(t *testing.T) {
	rec := buildReceipt(t)
	printedBy := uuidv7.New()

	err := rec.MarkPrinted("", "", "", printedBy)
	assert.ErrorIs(t, err, receipterrors.ErrFiscalNumberRequired)

	cancelledBy := uuidv7.New()
	require.NoError(t, rec.Cancel("Reason", cancelledBy))

	err = rec.MarkPrinted("FN-123", "url", "qr", printedBy)
	assert.ErrorIs(t, err, receipterrors.ErrReceiptAlreadyCancelled)
}

func TestReceipt_Cancel(t *testing.T) {
	rec := buildReceipt(t)
	cancelledBy := uuidv7.New()

	err := rec.Cancel("Customer request", cancelledBy)
	require.NoError(t, err)
	assert.Equal(t, ReceiptStatusCancelled, rec.Status)
	assert.NotNil(t, rec.CancelledAt)
	assert.Equal(t, cancelledBy, rec.LastUpdatedBy)

	err = rec.Cancel("Duplicate", cancelledBy)
	assert.ErrorIs(t, err, receipterrors.ErrReceiptAlreadyCancelled)
}

func TestReceipt_Cancel_ReasonRequired(t *testing.T) {
	rec := buildReceipt(t)

	err := rec.Cancel("", uuidv7.New())
	assert.ErrorIs(t, err, receipterrors.ErrReceiptCancelReasonRequired)
}

func TestReceipt_Validate_InvalidFields(t *testing.T) {
	rec, err := NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		PaymentTypeCash,
		ReceiptTypeSale,
		"UAH",
		[]ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	require.NoError(t, err)

	rec.PaymentType = PaymentType("invalid")
	require.ErrorIs(t, rec.Validate(), receipterrors.ErrPaymentTypeRequired)

	rec.PaymentType = PaymentTypeCash
	rec.ReceiptType = ReceiptType("invalid")
	require.ErrorIs(t, rec.Validate(), receipterrors.ErrReceiptTypeRequired)

	rec.ReceiptType = ReceiptTypeSale
	rec.Lines.Set([]ReceiptLine{{Name: "", Quantity: 1, PriceCents: 1000, TaxRate: 20}})
	require.ErrorIs(t, rec.Validate(), receipterrors.ErrReceiptLineNameRequired)
}

func TestReceipt_Validate_Success(t *testing.T) {
	rec := buildReceipt(t)
	assert.NoError(t, rec.Validate())
}

func TestCalculateTotals_Success(t *testing.T) {
	lines := []ReceiptLine{
		{Name: "Item A", Quantity: 2, PriceCents: 1000, TaxRate: 20},
		{Name: "Item B", Quantity: 1, PriceCents: 500, TaxRate: 0},
	}

	calculated, total, tax, err := calculateTotals(lines)
	require.NoError(t, err)
	require.Len(t, calculated, 2)
	require.Equal(t, int64(2500), total)
	require.Equal(t, int64(400), tax)
	require.Equal(t, int64(2000), calculated[0].TotalCents)
	require.Equal(t, int64(400), calculated[0].TaxAmountCents)
}

func TestCalculateTotals_Errors(t *testing.T) {
	_, _, _, err := calculateTotals(nil)
	require.ErrorIs(t, err, receipterrors.ErrReceiptLineNameRequired)

	_, _, _, err = calculateTotals([]ReceiptLine{{Name: "", Quantity: 1, PriceCents: 1000, TaxRate: 20}})
	require.ErrorIs(t, err, receipterrors.ErrReceiptLineNameRequired)

	_, _, _, err = calculateTotals([]ReceiptLine{{Name: "Item", Quantity: 0, PriceCents: 1000, TaxRate: 20}})
	require.ErrorIs(t, err, receipterrors.ErrReceiptLineQuantityInvalid)

	_, _, _, err = calculateTotals([]ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: -1, TaxRate: 20}})
	require.ErrorIs(t, err, receipterrors.ErrReceiptLinePriceInvalid)

	_, _, _, err = calculateTotals([]ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 200}})
	require.ErrorIs(t, err, receipterrors.ErrReceiptLineTaxRateInvalid)
}

func buildReceipt(t *testing.T) *Receipt {
	rec, err := NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		PaymentTypeCash,
		ReceiptTypeSale,
		"UAH",
		[]ReceiptLine{{Name: "Product A", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	require.NoError(t, err)
	return rec
}
