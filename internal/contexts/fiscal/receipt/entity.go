package receipt

import (
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ReceiptStatus represents receipt lifecycle state
type ReceiptStatus string

const (
	ReceiptStatusPending   ReceiptStatus = "pending"
	ReceiptStatusPrinted   ReceiptStatus = "printed"
	ReceiptStatusCancelled ReceiptStatus = "cancelled"
)

// PaymentType represents payment method
type PaymentType string

const (
	PaymentTypeCash     PaymentType = "cash"
	PaymentTypeCard     PaymentType = "card"
	PaymentTypeCashless PaymentType = "cashless"
)

// ReceiptType represents receipt type
type ReceiptType string

const (
	ReceiptTypeSale       ReceiptType = "sale"
	ReceiptTypeReturn     ReceiptType = "return"
	ReceiptTypeServiceIn  ReceiptType = "service_in"
	ReceiptTypeServiceOut ReceiptType = "service_out"
)

// Receipt represents a fiscal receipt aggregate
type Receipt struct {
	aggregate.BaseAggregate

	CashRegisterID uuidv7.UUID
	OrderID        uuidv7.UUID

	PaymentType PaymentType
	ReceiptType ReceiptType
	Currency    string

	TotalAmount int64
	TaxAmount   int64

	FiscalNumber string
	FiscalURL    string
	QRCode       string

	PrintedAt   *time.Time
	CancelledAt *time.Time

	CancellationReason string

	Lines jsonstore.Field[[]ReceiptLine]

	CreatedBy     uuidv7.UUID
	LastUpdatedBy uuidv7.UUID

	Status ReceiptStatus
}

// ReceiptLine represents a receipt line item
type ReceiptLine struct {
	Name           string
	Quantity       int
	PriceCents     int64
	TaxRate        int
	TotalCents     int64
	TaxAmountCents int64
}

// NewReceipt creates a new fiscal receipt
func NewReceipt(cashRegisterID, orderID uuidv7.UUID, paymentType PaymentType, receiptType ReceiptType, currency string, lines []ReceiptLine, createdBy uuidv7.UUID) (*Receipt, error) {
	if cashRegisterID == uuidv7.Nil {
		return nil, ErrCashRegisterIDRequired
	}
	if orderID == uuidv7.Nil {
		return nil, ErrOrderIDRequired
	}
	if paymentType == "" {
		return nil, ErrPaymentTypeRequired
	}
	if receiptType == "" {
		return nil, ErrReceiptTypeRequired
	}
	if currency == "" {
		return nil, ErrCurrencyRequired
	}
	if createdBy == uuidv7.Nil {
		return nil, ErrCreatedByRequired
	}
	if len(lines) == 0 {
		return nil, ErrReceiptLineNameRequired
	}

	if err := validatePaymentType(paymentType); err != nil {
		return nil, err
	}
	if err := validateReceiptType(receiptType); err != nil {
		return nil, err
	}

	calculatedLines, totalAmount, taxAmount, err := calculateTotals(lines)
	if err != nil {
		return nil, err
	}

	linesField := jsonstore.Field[[]ReceiptLine]{}
	linesField.Set(calculatedLines)

	return &Receipt{
		BaseAggregate:  aggregate.NewBaseAggregate(),
		CashRegisterID: cashRegisterID,
		OrderID:        orderID,
		PaymentType:    paymentType,
		ReceiptType:    receiptType,
		Currency:       currency,
		TotalAmount:    totalAmount,
		TaxAmount:      taxAmount,
		Lines:          linesField,
		CreatedBy:      createdBy,
		LastUpdatedBy:  createdBy,
		Status:         ReceiptStatusPending,
	}, nil
}

// MarkPrinted sets fiscal details and marks the receipt as printed
func (r *Receipt) MarkPrinted(fiscalNumber, fiscalURL, qrCode string, printedBy uuidv7.UUID) error {
	if r.Status == ReceiptStatusPrinted {
		return ErrReceiptAlreadyPrinted
	}
	if r.Status == ReceiptStatusCancelled {
		return ErrReceiptAlreadyCancelled
	}
	if fiscalNumber == "" {
		return ErrFiscalNumberRequired
	}

	now := time.Now()
	r.Status = ReceiptStatusPrinted
	r.FiscalNumber = fiscalNumber
	r.FiscalURL = fiscalURL
	r.QRCode = qrCode
	r.PrintedAt = &now
	r.LastUpdatedBy = printedBy
	r.Touch()

	return nil
}

// Cancel cancels the receipt
func (r *Receipt) Cancel(reason string, cancelledBy uuidv7.UUID) error {
	if r.Status == ReceiptStatusCancelled {
		return ErrReceiptAlreadyCancelled
	}
	if reason == "" {
		return ErrReceiptCancelReasonRequired
	}

	now := time.Now()
	r.Status = ReceiptStatusCancelled
	r.CancelledAt = &now
	r.CancellationReason = reason
	r.LastUpdatedBy = cancelledBy
	r.Touch()

	return nil
}

// Validate validates receipt state
func (r *Receipt) Validate() error {
	if r.CashRegisterID == uuidv7.Nil {
		return ErrCashRegisterIDRequired
	}
	if r.OrderID == uuidv7.Nil {
		return ErrOrderIDRequired
	}
	if r.PaymentType == "" {
		return ErrPaymentTypeRequired
	}
	if r.ReceiptType == "" {
		return ErrReceiptTypeRequired
	}
	if r.Currency == "" {
		return ErrCurrencyRequired
	}

	lines := r.Lines.Get()
	if len(lines) == 0 {
		return ErrReceiptLineNameRequired
	}

	if err := validatePaymentType(r.PaymentType); err != nil {
		return err
	}
	if err := validateReceiptType(r.ReceiptType); err != nil {
		return err
	}
	if _, _, _, err := calculateTotals(lines); err != nil {
		return err
	}

	return nil
}

func validatePaymentType(paymentType PaymentType) error {
	switch paymentType {
	case PaymentTypeCash, PaymentTypeCard, PaymentTypeCashless:
		return nil
	default:
		return ErrPaymentTypeRequired
	}
}

func validateReceiptType(receiptType ReceiptType) error {
	switch receiptType {
	case ReceiptTypeSale, ReceiptTypeReturn, ReceiptTypeServiceIn, ReceiptTypeServiceOut:
		return nil
	default:
		return ErrReceiptTypeRequired
	}
}

func calculateTotals(lines []ReceiptLine) ([]ReceiptLine, int64, int64, error) {
	if len(lines) == 0 {
		return nil, 0, 0, ErrReceiptLineNameRequired
	}

	var totalAmount int64
	var taxAmount int64
	calculated := make([]ReceiptLine, len(lines))

	for i, line := range lines {
		if line.Name == "" {
			return nil, 0, 0, ErrReceiptLineNameRequired
		}
		if line.Quantity <= 0 {
			return nil, 0, 0, ErrReceiptLineQuantityInvalid
		}
		if line.PriceCents < 0 {
			return nil, 0, 0, ErrReceiptLinePriceInvalid
		}
		if line.TaxRate < 0 || line.TaxRate > 100 {
			return nil, 0, 0, ErrReceiptLineTaxRateInvalid
		}

		lineTotal := int64(line.Quantity) * line.PriceCents
		lineTax := (lineTotal * int64(line.TaxRate)) / 100

		line.TotalCents = lineTotal
		line.TaxAmountCents = lineTax

		calculated[i] = line
		totalAmount += lineTotal
		taxAmount += lineTax
	}

	return calculated, totalAmount, taxAmount, nil
}
