package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestReceiptRow_ToEntityAndFromEntity(t *testing.T) {
	now := time.Now().UTC()
	printedAt := now.Add(-time.Hour)
	cancelledAt := now.Add(-30 * time.Minute)

	lines := jsonstore.Field[[]receipt.ReceiptLine]{}
	lines.Set([]receipt.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}})

	row := &receiptRow{
		ID:                 uuidv7.New().String(),
		Version:            2,
		CashRegisterID:     uuidv7.New().String(),
		OrderID:            uuidv7.New().String(),
		PaymentType:        string(receipt.PaymentTypeCash),
		ReceiptType:        string(receipt.ReceiptTypeSale),
		Currency:           "UAH",
		TotalAmount:        1000,
		TaxAmount:          200,
		Lines:              lines,
		CreatedBy:          uuidv7.New().String(),
		LastUpdatedBy:      uuidv7.New().String(),
		Status:             string(receipt.ReceiptStatusPrinted),
		CreatedAt:          now,
		UpdatedAt:          now,
		FiscalNumber:       nullString("FN-123"),
		FiscalURL:          nullString("https://example.com"),
		QRCode:             nullString("qr"),
		PrintedAt:          nullTime(printedAt),
		CancelledAt:        nullTime(cancelledAt),
		CancellationReason: nullString("test"),
	}

	entity, err := row.toEntity()
	require.NoError(t, err)
	require.Equal(t, receipt.ReceiptStatusPrinted, entity.Status)
	require.Equal(t, "FN-123", entity.FiscalNumber)
	require.Equal(t, "https://example.com", entity.FiscalURL)
	require.Equal(t, "qr", entity.QRCode)
	require.NotNil(t, entity.PrintedAt)
	require.NotNil(t, entity.CancelledAt)
	require.Equal(t, "test", entity.CancellationReason)

	back := fromEntity(entity)
	require.Equal(t, row.ID, back.ID)
	require.Equal(t, row.OrderID, back.OrderID)
	require.Equal(t, row.PaymentType, back.PaymentType)
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: true}
}

func nullTime(value time.Time) sql.NullTime {
	return sql.NullTime{Time: value, Valid: true}
}
