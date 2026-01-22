# Billing API Examples

Examples for **Billing Context** (`/api/v1/invoices`, `/api/v1/payments`, `/api/v1/subscriptions`).

---

## Invoices

### 1. Create Invoice

**Use Case**: Generate invoice for customer with line items

#### Request

```http
POST /api/v1/invoices HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
  "invoice_date": "2026-01-22",
  "due_date": "2026-02-21",
  "currency_code": "USD",
  "items": [
    {
      "description": "Professional Services - January 2026",
      "quantity": 40,
      "unit_price": 150.00,
      "tax_rate": 0.10
    },
    {
      "description": "Software License (Annual)",
      "quantity": 1,
      "unit_price": 2400.00,
      "tax_rate": 0.10
    }
  ],
  "notes": "Payment terms: Net 30"
}
```

#### curl

```bash
curl -X POST http://localhost:8080/api/v1/invoices \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
    "invoice_date": "2026-01-22",
    "due_date": "2026-02-21",
    "currency_code": "USD",
    "items": [
      {
        "description": "Professional Services - January 2026",
        "quantity": 40,
        "unit_price": 150.00,
        "tax_rate": 0.10
      }
    ],
    "notes": "Payment terms: Net 30"
  }'
```

#### Response (201 Created)

```json
{
  "success": true,
  "data": {
    "id": "01932e9d-1234-7abc-9def-0123456789gh",
    "invoice_number": "INV-2026-0001",
    "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
    "status": "draft",
    "invoice_date": "2026-01-22",
    "due_date": "2026-02-21",
    "currency_code": "USD",
    "subtotal": 8400.0,
    "tax": 840.0,
    "total": 9240.0,
    "items": [
      {
        "id": "01932e9d-2345-7abc-9def-0123456789hi",
        "description": "Professional Services - January 2026",
        "quantity": 40,
        "unit_price": 150.0,
        "tax_rate": 0.1,
        "line_total": 6600.0
      },
      {
        "id": "01932e9d-2346-7abc-9def-0123456789hj",
        "description": "Software License (Annual)",
        "quantity": 1,
        "unit_price": 2400.0,
        "tax_rate": 0.1,
        "line_total": 2640.0
      }
    ],
    "created_at": "2026-01-22T10:30:00Z"
  }
}
```

---

### 2. Finalize Invoice

**Use Case**: Mark invoice as finalized (ready to send to customer)

#### Request

```http
POST /api/v1/invoices/01932e9d-1234-7abc-9def-0123456789gh/finalize HTTP/1.1
Host: localhost:8080
```

#### curl

```bash
curl -X POST http://localhost:8080/api/v1/invoices/01932e9d-1234-7abc-9def-0123456789gh/finalize
```

#### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "01932e9d-1234-7abc-9def-0123456789gh",
    "invoice_number": "INV-2026-0001",
    "status": "open",
    "finalized_at": "2026-01-22T11:00:00Z"
  }
}
```

---

### 3. List Invoices

**Use Case**: Get invoices for a customer with status filtering

#### Request

```http
GET /api/v1/invoices?customer_id=01932e8f-1234-7abc-9def-0123456789ab&status=open&page=1&page_size=20 HTTP/1.1
Host: localhost:8080
```

#### curl

```bash
curl "http://localhost:8080/api/v1/invoices?customer_id=01932e8f-1234-7abc-9def-0123456789ab&status=open"
```

#### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "01932e9d-1234-7abc-9def-0123456789gh",
        "invoice_number": "INV-2026-0001",
        "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
        "status": "open",
        "total": 9240.0,
        "due_date": "2026-02-21",
        "created_at": "2026-01-22T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 1,
      "total_pages": 1
    }
  }
}
```

---

## Payments

### 4. Record Payment

**Use Case**: Record payment for an invoice

#### Request

```http
POST /api/v1/payments HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "invoice_id": "01932e9d-1234-7abc-9def-0123456789gh",
  "amount": 9240.00,
  "payment_method": "bank_transfer",
  "payment_date": "2026-02-15",
  "reference": "TXN-2026-0123",
  "notes": "Wire transfer received"
}
```

#### curl

```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_id": "01932e9d-1234-7abc-9def-0123456789gh",
    "amount": 9240.00,
    "payment_method": "bank_transfer",
    "payment_date": "2026-02-15",
    "reference": "TXN-2026-0123",
    "notes": "Wire transfer received"
  }'
```

#### Response (201 Created)

```json
{
  "success": true,
  "data": {
    "id": "01932e9e-1234-7abc-9def-0123456789ij",
    "invoice_id": "01932e9d-1234-7abc-9def-0123456789gh",
    "amount": 9240.0,
    "payment_method": "bank_transfer",
    "payment_date": "2026-02-15",
    "reference": "TXN-2026-0123",
    "status": "completed",
    "created_at": "2026-02-15T14:30:00Z"
  }
}
```

**Note**: Invoice status automatically updated to `paid` when full amount received.

---

### 5. List Payments for Invoice

**Use Case**: Get payment history for an invoice

#### Request

```http
GET /api/v1/invoices/01932e9d-1234-7abc-9def-0123456789gh/payments HTTP/1.1
Host: localhost:8080
```

#### curl

```bash
curl http://localhost:8080/api/v1/invoices/01932e9d-1234-7abc-9def-0123456789gh/payments
```

#### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "invoice_id": "01932e9d-1234-7abc-9def-0123456789gh",
    "total_paid": 9240.0,
    "remaining": 0.0,
    "payments": [
      {
        "id": "01932e9e-1234-7abc-9def-0123456789ij",
        "amount": 9240.0,
        "payment_method": "bank_transfer",
        "payment_date": "2026-02-15",
        "reference": "TXN-2026-0123",
        "status": "completed"
      }
    ]
  }
}
```

---

## Subscriptions

### 6. Create Subscription

**Use Case**: Create recurring subscription for customer

#### Request

```http
POST /api/v1/subscriptions HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
  "plan": "professional",
  "interval": "monthly",
  "amount": 299.00,
  "currency_code": "USD",
  "start_date": "2026-02-01",
  "auto_renew": true
}
```

#### curl

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
    "plan": "professional",
    "interval": "monthly",
    "amount": 299.00,
    "currency_code": "USD",
    "start_date": "2026-02-01",
    "auto_renew": true
  }'
```

#### Response (201 Created)

```json
{
  "success": true,
  "data": {
    "id": "01932e9f-1234-7abc-9def-0123456789kl",
    "customer_id": "01932e8f-1234-7abc-9def-0123456789ab",
    "plan": "professional",
    "interval": "monthly",
    "amount": 299.0,
    "currency_code": "USD",
    "status": "active",
    "start_date": "2026-02-01",
    "next_billing_date": "2026-03-01",
    "auto_renew": true,
    "created_at": "2026-01-22T10:30:00Z"
  }
}
```

---

### 7. Cancel Subscription

**Use Case**: Cancel recurring subscription

#### Request

```http
POST /api/v1/subscriptions/01932e9f-1234-7abc-9def-0123456789kl/cancel HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "cancel_at_period_end": true,
  "reason": "Customer downgrade"
}
```

#### curl

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions/01932e9f-1234-7abc-9def-0123456789kl/cancel \
  -H "Content-Type: application/json" \
  -d '{
    "cancel_at_period_end": true,
    "reason": "Customer downgrade"
  }'
```

#### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "01932e9f-1234-7abc-9def-0123456789kl",
    "status": "active",
    "cancel_at_period_end": true,
    "cancellation_date": "2026-03-01",
    "cancellation_reason": "Customer downgrade",
    "updated_at": "2026-01-22T11:00:00Z"
  }
}
```

---

## Related Contexts

- [Customer Management](customer-mgmt-examples.md) - Customer creation
- [Order Management](order-mgmt-examples.md) - Orders that generate invoices
- [Accounting](../../internal/contexts/accounting/README.md) - Journal entries from payments

---

## API Documentation

See [Swagger UI](http://localhost:8080/swagger/index.html) for full API reference.
