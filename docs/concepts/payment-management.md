# Payment Management Context

**Payment processing and transaction tracking** for Promenade Platform with complete payment lifecycle management, multiple payment methods, and refund handling.

---

## Overview

Payment Management is part of the **Billing Bounded Context** that handles payment processing, transaction tracking, and reconciliation. Built with **Domain-Driven Design** principles, it provides:

- **Payment Processing**: Create and process payments from customers
- **Multiple Payment Methods**: Credit card, bank transfer, PayPal, cash
- **Status Tracking**: Pending → Processing → Completed → Failed/Refunded
- **Invoice Linking**: Associate payments with invoices
- **Refund Management**: Full and partial refunds with audit trail
- **Query Operations**: Find payments by ID, number, customer, invoice, or status
- **Reporting**: Calculate totals by customer, invoice, or status

**Status**: ✅ Production-ready (January 2026)  
**Aggregate**: Payment  
**Routes**: 14 HTTP endpoints  
**Database**: 1 table with soft delete support

---

## Domain Model

### Payment Aggregate

**Payment** is the aggregate root that manages transaction processing and tracking.

```go
type Payment struct {
    ID            uuid.UUID
    PaymentNumber string      // Auto-generated: PAY-YYYY-NNNNNN
    CustomerID    uuid.UUID   // Required: who is paying
    InvoiceID     *uuid.UUID  // Optional: linked invoice
    
    Status        PaymentStatus // pending, processing, completed, failed, refunded
    PaymentMethod PaymentMethod // credit_card, bank_transfer, paypal, cash
    
    // Financial details
    Currency      string      // ISO 4217 (USD, EUR, UAH)
    Amount        Money       // Total payment amount (cents)
    RefundedAmount Money      // Total refunded (cents)
    
    // Payment details
    TransactionID string      // External payment gateway transaction ID
    PaymentDate   time.Time   // When payment was made
    
    // Metadata
    Notes         string      // Additional information
    
    // Audit fields
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time  // Soft delete
}
```

### Payment Status Enum

```go
type PaymentStatus string

const (
    PaymentStatusPending    PaymentStatus = "pending"    // Awaiting processing
    PaymentStatusProcessing PaymentStatus = "processing" // Being processed
    PaymentStatusCompleted  PaymentStatus = "completed"  // Successfully processed
    PaymentStatusFailed     PaymentStatus = "failed"     // Processing failed
    PaymentStatusRefunded   PaymentStatus = "refunded"   // Fully refunded
)
```

### Payment Method Enum

```go
type PaymentMethod string

const (
    PaymentMethodCreditCard   PaymentMethod = "credit_card"
    PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
    PaymentMethodPayPal       PaymentMethod = "paypal"
    PaymentMethodCash         PaymentMethod = "cash"
)
```

### Value Objects

**Money** - Type-safe handling of currency amounts:

```go
type Money struct {
    Amount   int64  // Cents (e.g., 10000 = $100.00)
    Currency string // ISO 4217 (USD, EUR, UAH)
}

// Factory methods
money := NewMoney(10000, "USD")      // $100.00
money := NewMoneyFromFloat(99.99, "EUR") // €99.99
```

---

## Business Rules

### Payment Creation

1. **Required Fields**: CustomerID, Currency, Amount, PaymentMethod, PaymentDate
2. **Auto-generation**: PaymentNumber = "PAY-YYYY-NNNNNN" (sequential)
3. **Status**: New payments start in "pending" status
4. **Amounts**: All amounts stored as cents (integer) for precision
5. **Invoice Linking**: Optional InvoiceID can be set during creation

### Payment Lifecycle

**State Machine**:

```
pending → processing → completed
   ↓                       ↓
 failed                 refunded
```

**Valid Transitions**:
- Pending → Processing: Payment processing started
- Processing → Completed: Payment successful
- Processing → Failed: Payment processing failed
- Completed → Refunded: Full or partial refund issued
- Pending → Failed: Cancelled before processing

### Refund Handling

1. **Full Refund**: RefundedAmount = Amount, Status = Refunded
2. **Partial Refund**: RefundedAmount < Amount, Status remains Completed
3. **Multiple Refunds**: RefundedAmount accumulates from multiple partial refunds
4. **Validation**: Cannot refund more than original Amount
5. **Status Update**: Auto-update to Refunded when fully refunded

---

## Use Cases

### Core Operations

**Create Payment**:
```go
func (uc *PaymentUseCase) CreatePayment(ctx context.Context,
    customerID uuid.UUID,
    currency string,
    amount int64,
    paymentMethod PaymentMethod,
    paymentDate time.Time) (*Payment, error)
```

**Process Payment**:
```go
func (uc *PaymentUseCase) ProcessPayment(ctx context.Context,
    paymentID uuid.UUID,
    transactionID string) error
```

**Complete Payment**:
```go
func (uc *PaymentUseCase) CompletePayment(ctx context.Context,
    paymentID uuid.UUID) error
```

**Refund Payment**:
```go
func (uc *PaymentUseCase) RefundPayment(ctx context.Context,
    paymentID uuid.UUID,
    refundAmount int64,
    reason string) error
```

**Link to Invoice**:
```go
func (uc *PaymentUseCase) LinkToInvoice(ctx context.Context,
    paymentID uuid.UUID,
    invoiceID uuid.UUID) error
```

### Query Operations

**Get by ID**:
```go
func (uc *PaymentUseCase) GetPayment(ctx context.Context,
    paymentID uuid.UUID) (*Payment, error)
```

**List by Customer**:
```go
func (uc *PaymentUseCase) ListByCustomer(ctx context.Context,
    customerID uuid.UUID,
    page, pageSize int) ([]*Payment, int, error)
```

**List by Invoice**:
```go
func (uc *PaymentUseCase) ListByInvoice(ctx context.Context,
    invoiceID uuid.UUID,
    page, pageSize int) ([]*Payment, int, error)
```

**List by Status**:
```go
func (uc *PaymentUseCase) ListByStatus(ctx context.Context,
    status PaymentStatus,
    page, pageSize int) ([]*Payment, int, error)
```

### Reporting Operations

**Count by Status**:
```go
func (uc *PaymentUseCase) CountByStatus(ctx context.Context,
    status PaymentStatus) (int, error)
```

**Total by Customer**:
```go
func (uc *PaymentUseCase) GetTotalByCustomer(ctx context.Context,
    customerID uuid.UUID) (Money, error)
```

**Total by Invoice**:
```go
func (uc *PaymentUseCase) GetTotalByInvoice(ctx context.Context,
    invoiceID uuid.UUID) (Money, error)
```

---

## HTTP API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/billing/payments` | Create new payment |
| GET | `/api/v1/billing/payments/:id` | Get payment by ID |
| PUT | `/api/v1/billing/payments/:id` | Update payment |
| DELETE | `/api/v1/billing/payments/:id` | Delete payment (soft) |
| GET | `/api/v1/billing/payments` | List all payments (paginated) |
| GET | `/api/v1/billing/payments/customer/:customerID` | List by customer |
| GET | `/api/v1/billing/payments/invoice/:invoiceID` | List by invoice |
| GET | `/api/v1/billing/payments/status/:status` | List by status |
| GET | `/api/v1/billing/payments/number/:paymentNumber` | Get by payment number |
| POST | `/api/v1/billing/payments/:id/process` | Process payment |
| POST | `/api/v1/billing/payments/:id/complete` | Mark as completed |
| POST | `/api/v1/billing/payments/:id/refund` | Issue refund |
| POST | `/api/v1/billing/payments/:id/link-invoice` | Link to invoice |
| GET | `/api/v1/billing/payments/stats/total-by-customer/:customerID` | Total by customer |

### Create Payment

**Request**:
```http
POST /api/v1/billing/payments
Content-Type: application/json

{
  "customer_id": "01JGABC1234567890ABCDEFGHI",
  "currency": "USD",
  "amount": 11000,
  "payment_method": "credit_card",
  "payment_date": "2026-01-05T10:30:00Z",
  "invoice_id": "01JGINV1234567890ABCDEFGHI",
  "notes": "Payment for Invoice INV-2026-000001"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGPAY1234567890ABCDEFGHI",
    "payment_number": "PAY-2026-000001",
    "customer_id": "01JGABC1234567890ABCDEFGHI",
    "invoice_id": "01JGINV1234567890ABCDEFGHI",
    "status": "pending",
    "payment_method": "credit_card",
    "currency": "USD",
    "amount": 11000,
    "refunded_amount": 0,
    "payment_date": "2026-01-05T10:30:00Z",
    "transaction_id": "",
    "notes": "Payment for Invoice INV-2026-000001",
    "created_at": "2026-01-05T10:30:00Z",
    "updated_at": "2026-01-05T10:30:00Z"
  }
}
```

### Process Payment

**Request**:
```http
POST /api/v1/billing/payments/01JGPAY1234567890ABCDEFGHI/process
Content-Type: application/json

{
  "transaction_id": "txn_1234567890abcdef"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGPAY1234567890ABCDEFGHI",
    "status": "processing",
    "transaction_id": "txn_1234567890abcdef",
    "updated_at": "2026-01-05T10:31:00Z"
  }
}
```

### Complete Payment

**Request**:
```http
POST /api/v1/billing/payments/01JGPAY1234567890ABCDEFGHI/complete
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGPAY1234567890ABCDEFGHI",
    "status": "completed",
    "updated_at": "2026-01-05T10:32:00Z"
  }
}
```

### Refund Payment

**Request**:
```http
POST /api/v1/billing/payments/01JGPAY1234567890ABCDEFGHI/refund
Content-Type: application/json

{
  "refund_amount": 5500,
  "reason": "Partial refund - defective item"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGPAY1234567890ABCDEFGHI",
    "status": "completed",
    "refunded_amount": 5500,
    "notes": "Partial refund - defective item",
    "updated_at": "2026-01-05T11:00:00Z"
  }
}
```

### List Payments by Customer

**Request**:
```http
GET /api/v1/billing/payments/customer/01JGABC1234567890ABCDEFGHI?page=1&page_size=20
```

**Response**:
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGPAY1234567890ABCDEFGHI",
      "payment_number": "PAY-2026-000001",
      "customer_id": "01JGABC1234567890ABCDEFGHI",
      "invoice_id": "01JGINV1234567890ABCDEFGHI",
      "status": "completed",
      "payment_method": "credit_card",
      "amount": 11000,
      "refunded_amount": 5500,
      "payment_date": "2026-01-05T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 15,
    "page": 1,
    "page_size": 20,
    "total_pages": 1
  }
}
```

---

## Database Schema

### billing_payments Table

```sql
CREATE TABLE billing_payments (
    id UUID PRIMARY KEY,
    payment_number VARCHAR(20) UNIQUE NOT NULL,
    customer_id UUID NOT NULL,
    invoice_id UUID,
    status VARCHAR(20) NOT NULL,
    payment_method VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    amount BIGINT NOT NULL,
    refunded_amount BIGINT DEFAULT 0,
    transaction_id VARCHAR(255),
    payment_date TIMESTAMP NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_billing_payments_customer_id ON billing_payments(customer_id);
CREATE INDEX idx_billing_payments_invoice_id ON billing_payments(invoice_id);
CREATE INDEX idx_billing_payments_status ON billing_payments(status);
CREATE INDEX idx_billing_payments_payment_date ON billing_payments(payment_date);
CREATE INDEX idx_billing_payments_payment_number ON billing_payments(payment_number);
```

---

## Testing

### Test Coverage

**Payment Aggregate**: 75 tests (100% pass rate)
- Unit tests: 58 (entity validation, status transitions, refund logic)
- Integration tests: 9 (repository CRUD with PostgreSQL)
- Smoke tests: 8 (HTTP handler validation)

### Unit Tests

**Entity Tests** (`entity_test.go`):
- Payment creation with valid data
- Payment number generation
- Status transitions (Pending → Processing → Completed)
- Refund calculations (partial and full refunds)
- Invalid refund amounts (exceeding original amount)
- Payment method validation
- Money calculations and precision

### Integration Tests

**Repository Tests** (`test/integration/contexts/billing/payment/repository_test.go`):
- CRUD operations with PostgreSQL
- List by customer with pagination
- List by invoice with filtering
- List by status with filtering
- Count by status (aggregation)
- Refund amount tracking
- Get by payment number (unique constraint)
- Total by invoice (only completed payments)
- Total by customer (only completed payments)

**Key Integration Test Fixes** (January 2026):
- UUID suffixes for uniqueness in concurrent tests
- Shortened [:6] format for VARCHAR(20) payment_number
- GetTotal methods filter by status='completed' (business rule)

### Smoke Tests

**Handler Tests** (`test/smoke/contexts/billing/payment/handler_test.go`):
- Create payment success (201 Created)
- Create payment validation error (400 Bad Request)
- Get by ID success (200 OK)
- Get by ID not found (404 Not Found)
- Process payment success (200 OK)
- Refund payment success (200 OK)
- Link to invoice success (200 OK)
- List payments success (200 OK)

---

## Example Usage

### Process Customer Payment

```go
// 1. Create payment
payment, err := useCase.CreatePayment(ctx,
    customerID,
    "USD",
    11000, // $110.00
    PaymentMethodCreditCard,
    time.Now(),
)

// 2. Link to invoice
err = useCase.LinkToInvoice(ctx, payment.ID, invoiceID)

// 3. Process payment (submit to payment gateway)
err = useCase.ProcessPayment(ctx, payment.ID, "txn_abc123")

// 4. Complete payment (confirmation from gateway)
err = useCase.CompletePayment(ctx, payment.ID)
```

### Issue Refund

```go
// Full refund
err := useCase.RefundPayment(ctx, payment.ID, payment.Amount.Amount, "Order cancelled")

// Partial refund
halfAmount := payment.Amount.Amount / 2
err := useCase.RefundPayment(ctx, payment.ID, halfAmount, "Partial return")
```

### Query Payment History

```go
// Get all payments for a customer
payments, total, err := useCase.ListByCustomer(ctx, customerID, 1, 20)

// Calculate total paid
totalPaid, err := useCase.GetTotalByCustomer(ctx, customerID)
fmt.Printf("Customer has paid: %s %.2f\n",
    totalPaid.Currency,
    float64(totalPaid.Amount)/100,
)

// Check invoice payment status
invoicePayments, _, err := useCase.ListByInvoice(ctx, invoiceID, 1, 100)
var totalReceived int64
for _, payment := range invoicePayments {
    if payment.Status == PaymentStatusCompleted {
        totalReceived += payment.Amount.Amount - payment.RefundedAmount.Amount
    }
}
```

---

## Integration with Other Contexts

### Invoice Management Context

When payment is completed, update invoice amount paid:

```go
// Payment completed event handler
func (h *PaymentEventHandler) OnPaymentCompleted(ctx context.Context, event Event) error {
    payment, err := h.paymentUseCase.GetPayment(ctx, event.AggregateID)
    
    if payment.InvoiceID != nil {
        // Update invoice amount paid
        invoice, err := h.invoiceUseCase.GetInvoice(ctx, *payment.InvoiceID)
        invoice.AmountPaid += payment.Amount.Amount
        
        // Mark invoice as paid if fully paid
        if invoice.AmountPaid >= invoice.Total.Amount {
            err = h.invoiceUseCase.MarkAsPaid(ctx, invoice.ID)
        }
    }
    
    return nil
}
```

### Customer Management Context

Track customer payment behavior:

```go
// Get customer payment statistics
payments, total, err := paymentUseCase.ListByCustomer(ctx, customerID, 1, 1000)

// Calculate statistics
var totalPaid int64
var refundCount int
var avgPaymentTime time.Duration

for _, payment := range payments {
    if payment.Status == PaymentStatusCompleted {
        totalPaid += payment.Amount.Amount - payment.RefundedAmount.Amount
        avgPaymentTime += payment.UpdatedAt.Sub(payment.CreatedAt)
    }
    if payment.RefundedAmount.Amount > 0 {
        refundCount++
    }
}
```

---

## Best Practices

### DO

- **Use Money Value Object**: Always use `Money` type for amounts (never float)
- **Store Cents**: Store amounts as cents (int64) to avoid floating-point errors
- **Generate Payment Numbers**: Auto-generate sequential payment numbers
- **Track Transaction IDs**: Store external payment gateway transaction IDs
- **Soft Delete**: Use soft delete (deleted_at) to preserve audit trail
- **Validate Refunds**: Ensure refund amount doesn't exceed original amount
- **Filter Completed**: Only sum completed payments in reporting queries

### DON'T

- **Don't Use Float**: Never use float for money calculations
- **Don't Skip Status Checks**: Validate status transitions
- **Don't Hard Delete**: Use soft delete for compliance
- **Don't Allow Over-Refunds**: Validate refund amounts
- **Don't Modify Completed**: Payments are immutable after completion
- **Don't Include Pending**: Exclude pending payments from totals

---

## Future Enhancements

### Planned Features

- [ ] **Payment Gateway Integration**: Stripe, PayPal, Square APIs
- [ ] **Webhook Handling**: Async payment confirmations
- [ ] **Payment Plans**: Split payments into installments
- [ ] **Recurring Payments**: Auto-charge on schedule
- [ ] **Multi-Currency**: Support currency conversion
- [ ] **Failed Payment Retry**: Automatic retry logic
- [ ] **Payment Links**: Generate payment URLs for customers
- [ ] **PCI Compliance**: Tokenization and secure storage
- [ ] **Fraud Detection**: Risk scoring and flagging
- [ ] **Payment Analytics**: Success rates, average processing time

---

## Related Documentation

- [Main README](../../README.md)
- [Invoice Management](invoice-management.md)
- [Billing Context README](../../internal/contexts/billing/README.md)
- [Order Management](order-management.md)
- [Customer Management](customer-management.md)
- [Testing Guide](../../test/README.md)

---

**Last Updated**: January 5, 2026  
**Status**: Production-ready  
**Test Coverage**: 75 tests, 100% pass rate  
**Maintainer**: Promenade Team
