# Invoice Management Context

**Invoice generation and revenue tracking** for Promenade Platform with complete invoice lifecycle management, customer billing, and financial reporting.

---

## Overview

Invoice Management is part of the **Billing Bounded Context** that handles the complete invoice lifecycle from generation through payment tracking. Built with **Domain-Driven Design** principles, it provides:

- **Invoice Generation**: Create invoices for customers with line items
- **Status Tracking**: Draft → Sent → Paid → Overdue → Cancelled
- **Payment Linking**: Track payments associated with invoices
- **Revenue Reporting**: Calculate total revenue and outstanding amounts
- **Query Operations**: Find invoices by ID, number, customer, or status
- **Pagination**: Efficient listing with standard pagination support

**Status**:  Production-ready (January 2026)  
**Aggregate**: Invoice  
**Routes**: 14 HTTP endpoints  
**Database**: 1 table with soft delete support

---

## Domain Model

### Invoice Aggregate

**Invoice** is the aggregate root that manages billing and payment tracking for customers.

```go
type Invoice struct {
    ID            uuid.UUID
    InvoiceNumber string      // Auto-generated: INV-YYYY-NNNNNN
    CustomerID    uuid.UUID   // Required: reference to customer
    Status        InvoiceStatus // draft, sent, paid, overdue, cancelled
    
    // Financial details
    Currency      string      // ISO 4217 (USD, EUR, UAH)
    Subtotal      Money       // Sum before tax (cents)
    Tax           Money       // Tax amount (cents)
    Total         Money       // Final amount (cents)
    AmountPaid    Money       // Total payments received (cents)
    
    // Dates
    IssueDate     time.Time   // When invoice was created
    DueDate       time.Time   // Payment deadline
    PaidDate      *time.Time  // When fully paid (nullable)
    
    // Relationships
    OrderID       *uuid.UUID  // Optional: linked order
    
    // Audit fields
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time  // Soft delete
}
```

### Invoice Status Enum

```go
type InvoiceStatus string

const (
    InvoiceStatusDraft    InvoiceStatus = "draft"     // Being prepared
    InvoiceStatusSent     InvoiceStatus = "sent"      // Sent to customer
    InvoiceStatusPaid     InvoiceStatus = "paid"      // Fully paid
    InvoiceStatusOverdue  InvoiceStatus = "overdue"   // Past due date
    InvoiceStatusCancelled InvoiceStatus = "cancelled" // Voided
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

### Invoice Creation

1. **Required Fields**: CustomerID, Currency, IssueDate, DueDate
2. **Auto-generation**: InvoiceNumber = "INV-YYYY-NNNNNN" (sequential)
3. **Status**: New invoices start in "draft" status
4. **Amounts**: All amounts stored as cents (integer) for precision
5. **Due Date**: Must be after IssueDate

### Invoice Lifecycle

**State Machine**:

```
draft → sent → paid
  ↓      ↓       
cancelled  overdue → paid
```

**Valid Transitions**:
- Draft → Sent: Ready to bill customer
- Draft → Cancelled: Voided before sending
- Sent → Paid: Payment received
- Sent → Overdue: Past due date without payment
- Sent → Cancelled: Voided after sending
- Overdue → Paid: Late payment received

### Payment Tracking

1. **AmountPaid**: Updated when payments are linked to invoice
2. **Paid Status**: Invoice marked "paid" when AmountPaid >= Total
3. **Partial Payments**: Supported (AmountPaid < Total)
4. **Overdue Check**: System checks DueDate vs current date

---

## Use Cases

### Core Operations

**Create Invoice**:
```go
func (uc *InvoiceUseCase) CreateInvoice(ctx context.Context, 
    customerID uuid.UUID,
    currency string,
    issueDate, dueDate time.Time,
    subtotal, tax int64) (*Invoice, error)
```

**Send Invoice**:
```go
func (uc *InvoiceUseCase) SendInvoice(ctx context.Context, 
    invoiceID uuid.UUID) error
```

**Mark as Paid**:
```go
func (uc *InvoiceUseCase) MarkAsPaid(ctx context.Context, 
    invoiceID uuid.UUID) error
```

**Cancel Invoice**:
```go
func (uc *InvoiceUseCase) CancelInvoice(ctx context.Context, 
    invoiceID uuid.UUID) error
```

### Query Operations

**Get by ID**:
```go
func (uc *InvoiceUseCase) GetInvoice(ctx context.Context, 
    invoiceID uuid.UUID) (*Invoice, error)
```

**List by Customer**:
```go
func (uc *InvoiceUseCase) ListByCustomer(ctx context.Context, 
    customerID uuid.UUID, 
    page, pageSize int) ([]*Invoice, int, error)
```

**List by Status**:
```go
func (uc *InvoiceUseCase) ListByStatus(ctx context.Context, 
    status InvoiceStatus, 
    page, pageSize int) ([]*Invoice, int, error)
```

### Reporting Operations

**Count by Status**:
```go
func (uc *InvoiceUseCase) CountByStatus(ctx context.Context, 
    status InvoiceStatus) (int, error)
```

**Total Revenue**:
```go
func (uc *InvoiceUseCase) GetTotalRevenue(ctx context.Context) (Money, error)
```

---

## HTTP API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/billing/invoices` | Create new invoice |
| GET | `/api/v1/billing/invoices/:id` | Get invoice by ID |
| PUT | `/api/v1/billing/invoices/:id` | Update invoice |
| DELETE | `/api/v1/billing/invoices/:id` | Delete invoice (soft) |
| GET | `/api/v1/billing/invoices` | List all invoices (paginated) |
| GET | `/api/v1/billing/invoices/customer/:customerID` | List by customer |
| GET | `/api/v1/billing/invoices/status/:status` | List by status |
| GET | `/api/v1/billing/invoices/number/:invoiceNumber` | Get by invoice number |
| POST | `/api/v1/billing/invoices/:id/send` | Send invoice to customer |
| POST | `/api/v1/billing/invoices/:id/pay` | Mark invoice as paid |
| POST | `/api/v1/billing/invoices/:id/cancel` | Cancel invoice |
| GET | `/api/v1/billing/invoices/stats/count` | Count by status |
| GET | `/api/v1/billing/invoices/stats/revenue` | Total revenue |
| POST | `/api/v1/billing/invoices/:id/link-payment` | Link payment to invoice |

### Create Invoice

**Request**:
```http
POST /api/v1/billing/invoices
Content-Type: application/json

{
  "customer_id": "01JGABC1234567890ABCDEFGHI",
  "currency": "USD",
  "issue_date": "2026-01-05T00:00:00Z",
  "due_date": "2026-02-05T00:00:00Z",
  "subtotal": 10000,
  "tax": 1000,
  "order_id": "01JGXYZ1234567890ABCDEFGHI"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINV1234567890ABCDEFGHI",
    "invoice_number": "INV-2026-000001",
    "customer_id": "01JGABC1234567890ABCDEFGHI",
    "status": "draft",
    "currency": "USD",
    "subtotal": 10000,
    "tax": 1000,
    "total": 11000,
    "amount_paid": 0,
    "issue_date": "2026-01-05T00:00:00Z",
    "due_date": "2026-02-05T00:00:00Z",
    "paid_date": null,
    "order_id": "01JGXYZ1234567890ABCDEFGHI",
    "created_at": "2026-01-05T12:00:00Z",
    "updated_at": "2026-01-05T12:00:00Z"
  }
}
```

### Send Invoice

**Request**:
```http
POST /api/v1/billing/invoices/01JGINV1234567890ABCDEFGHI/send
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINV1234567890ABCDEFGHI",
    "status": "sent",
    "updated_at": "2026-01-05T12:05:00Z"
  }
}
```

### Mark as Paid

**Request**:
```http
POST /api/v1/billing/invoices/01JGINV1234567890ABCDEFGHI/pay
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINV1234567890ABCDEFGHI",
    "status": "paid",
    "paid_date": "2026-01-15T10:30:00Z",
    "updated_at": "2026-01-15T10:30:00Z"
  }
}
```

### List Invoices

**Request**:
```http
GET /api/v1/billing/invoices?page=1&page_size=20&status=sent
```

**Response**:
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGINV1234567890ABCDEFGHI",
      "invoice_number": "INV-2026-000001",
      "customer_id": "01JGABC1234567890ABCDEFGHI",
      "status": "sent",
      "total": 11000,
      "amount_paid": 0,
      "due_date": "2026-02-05T00:00:00Z"
    }
  ],
  "pagination": {
    "total": 45,
    "page": 1,
    "page_size": 20,
    "total_pages": 3
  }
}
```

---

## Database Schema

### billing_invoices Table

```sql
CREATE TABLE billing_invoices (
    id UUID PRIMARY KEY,
    invoice_number VARCHAR(20) UNIQUE NOT NULL,
    customer_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    subtotal BIGINT NOT NULL,
    tax BIGINT NOT NULL,
    total BIGINT NOT NULL,
    amount_paid BIGINT DEFAULT 0,
    issue_date TIMESTAMP NOT NULL,
    due_date TIMESTAMP NOT NULL,
    paid_date TIMESTAMP,
    order_id UUID,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_billing_invoices_customer_id ON billing_invoices(customer_id);
CREATE INDEX idx_billing_invoices_status ON billing_invoices(status);
CREATE INDEX idx_billing_invoices_due_date ON billing_invoices(due_date);
CREATE INDEX idx_billing_invoices_invoice_number ON billing_invoices(invoice_number);
```

---

## Testing

### Test Coverage

**Invoice Aggregate**: 43 tests (100% pass rate)
- Unit tests: 30 (entity validation, status transitions)
- Integration tests: 5 (repository CRUD with PostgreSQL)
- Smoke tests: 8 (HTTP handler validation)

### Unit Tests

**Entity Tests** (`entity_test.go`):
- Invoice creation with valid data
- Invoice number generation
- Status transitions (Draft → Sent → Paid)
- Money calculations (Subtotal + Tax = Total)
- Invalid state transitions (error handling)
- Date validation (DueDate > IssueDate)

### Integration Tests

**Repository Tests** (`test/integration/contexts/billing/invoice/repository_test.go`):
- CRUD operations with PostgreSQL
- List by customer with pagination
- List by status with filtering
- Count by status (aggregation)
- Total revenue calculation

### Smoke Tests

**Handler Tests** (`test/smoke/contexts/billing/invoice/handler_test.go`):
- Create invoice success (201 Created)
- Create invoice validation error (400 Bad Request)
- Get by ID success (200 OK)
- Get by ID not found (404 Not Found)
- Delete invoice success (204 No Content)
- Delete invoice not found (404 Not Found)
- List invoices success (200 OK)
- List invoices empty result (200 OK with empty array)

---

## Example Usage

### Create and Process Invoice

```go
// 1. Create draft invoice
invoice, err := useCase.CreateInvoice(ctx,
    customerID,
    "USD",
    time.Now(),
    time.Now().AddDate(0, 1, 0), // Due in 30 days
    10000, // $100.00 subtotal
    1000,  // $10.00 tax
)

// 2. Send invoice to customer
err = useCase.SendInvoice(ctx, invoice.ID)

// 3. Customer pays (triggered by payment received)
err = useCase.MarkAsPaid(ctx, invoice.ID)
```

### Query Overdue Invoices

```go
// Get all overdue invoices
overdueInvoices, total, err := useCase.ListByStatus(ctx, 
    InvoiceStatusOverdue, 
    1, 20, // page 1, 20 per page
)

for _, invoice := range overdueInvoices {
    daysOverdue := time.Since(invoice.DueDate).Hours() / 24
    fmt.Printf("Invoice %s is %d days overdue\n", 
        invoice.InvoiceNumber, int(daysOverdue))
}
```

### Revenue Reporting

```go
// Calculate total revenue from paid invoices
totalRevenue, err := useCase.GetTotalRevenue(ctx)
fmt.Printf("Total revenue: %s %.2f\n", 
    totalRevenue.Currency, 
    float64(totalRevenue.Amount)/100,
)

// Count invoices by status
draftCount, _ := useCase.CountByStatus(ctx, InvoiceStatusDraft)
sentCount, _ := useCase.CountByStatus(ctx, InvoiceStatusSent)
paidCount, _ := useCase.CountByStatus(ctx, InvoiceStatusPaid)
overdueCount, _ := useCase.CountByStatus(ctx, InvoiceStatusOverdue)
```

---

## Integration with Other Contexts

### Order Management Context

When an order is fulfilled, create an invoice:

```go
// Order fulfilled event handler
func (h *OrderEventHandler) OnOrderFulfilled(ctx context.Context, event Event) error {
    order, err := h.orderUseCase.GetOrder(ctx, event.AggregateID)
    
    // Create invoice for the order
    invoice, err := h.invoiceUseCase.CreateInvoice(ctx,
        order.CustomerID,
        order.Currency,
        time.Now(),
        time.Now().AddDate(0, 0, 30), // 30 days payment term
        order.Subtotal.Amount,
        order.Tax.Amount,
    )
    
    // Link invoice to order
    invoice.OrderID = &order.ID
    
    return h.invoiceUseCase.UpdateInvoice(ctx, invoice)
}
```

### Customer Management Context

Track customer payment history:

```go
// Get all invoices for a customer
invoices, total, err := invoiceUseCase.ListByCustomer(ctx, customerID, 1, 100)

// Calculate customer statistics
var totalOwed int64
var overdueCount int

for _, invoice := range invoices {
    if invoice.Status == InvoiceStatusSent || invoice.Status == InvoiceStatusOverdue {
        totalOwed += invoice.Total.Amount - invoice.AmountPaid.Amount
    }
    if invoice.Status == InvoiceStatusOverdue {
        overdueCount++
    }
}
```

---

## Best Practices

### DO

- **Use Money Value Object**: Always use `Money` type for amounts (never float)
- **Store Cents**: Store amounts as cents (int64) to avoid floating-point errors
- **Generate Invoice Numbers**: Auto-generate sequential invoice numbers
- **Track Payment History**: Link payments to invoices via AmountPaid
- **Soft Delete**: Use soft delete (deleted_at) to preserve audit trail
- **Validate Dates**: Ensure DueDate > IssueDate
- **Handle Partial Payments**: Support multiple payments per invoice

### DON'T

- **Don't Use Float**: Never use float for money calculations
- **Don't Skip Status Checks**: Validate status transitions
- **Don't Hard Delete**: Use soft delete for compliance
- **Don't Allow Backdating**: Validate IssueDate is reasonable
- **Don't Modify Paid Invoices**: Immutable after payment received

---

## Future Enhancements

### Planned Features

- [ ] **PDF Generation**: Generate PDF invoices for download
- [ ] **Email Integration**: Send invoices via email
- [ ] **Recurring Invoices**: Auto-generate invoices on schedule
- [ ] **Multi-Currency**: Support currency conversion rates
- [ ] **Payment Plans**: Split payments into installments
- [ ] **Credit Notes**: Issue refunds and adjustments
- [ ] **Tax Automation**: Calculate tax based on location
- [ ] **Dunning Process**: Automated overdue reminders

---

## Related Documentation

- [Main README](../../README.md)
- [Payment Management](payment-management.md)
- [Billing Context README](../../internal/contexts/billing/README.md)
- [Order Management](order-management.md)
- [Customer Management](customer-management.md)
- [Testing Guide](../../test/README.md)

---

**Last Updated**: January 5, 2026  
**Status**: Production-ready  
**Test Coverage**: 43 tests, 100% pass rate  
**Maintainer**: Promenade Team
