package receipt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			wantErr:        ErrCashRegisterIDRequired,
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
			wantErr:        ErrOrderIDRequired,
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
			wantErr:        ErrPaymentTypeRequired,
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
			wantErr:        ErrReceiptTypeRequired,
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
			wantErr:        ErrCurrencyRequired,
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
			wantErr:        ErrCreatedByRequired,
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
	assert.ErrorIs(t, err, ErrReceiptAlreadyPrinted)
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
	assert.ErrorIs(t, err, ErrReceiptAlreadyCancelled)
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
