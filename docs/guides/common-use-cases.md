# Common Use Cases

**Real-world business scenarios** with complete API workflows for Promenade Platform.

---

## Overview

This guide provides **end-to-end examples** for common business operations:

1. **Customer Lifecycle** - Lead to customer conversion
2. **B2B Customer Onboarding** - Company and contact setup
3. **Sales Pipeline** - Deal management from lead to win
4. **Order Processing** - Complete order fulfillment
5. **Invoice & Payment** - Billing workflow
6. **Subscription Management** - Recurring billing
7. **Customer Interaction Tracking** - Calls, emails, meetings
8. **Multi-Tenant Operations** - B2C vs B2B workflows

---

## Use Case 1: Customer Lifecycle Management

**Scenario**: Convert a website visitor into a paying customer.

### Flow Diagram

```
Lead → Prospect → Customer → Tier Upgrade → Churned
 |        |          |            |              |
 v        v          v            v              v
Tags    Phone    First Order   Pro Tier      Reason
```

### Step 1: Create Lead

**New website signup**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Sarah Johnson",
    "email": "sarah@example.com",
    "customer_type": "b2c",
    "status": "lead",
    "tier": "free",
    "source": "website",
    "tags": ["website-signup", "newsletter-subscriber"]
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCUST123456789ABCDEF012",
    "name": "Sarah Johnson",
    "email": "sarah@example.com",
    "customer_type": "b2c",
    "status": "lead",
    "tier": "free",
    "source": "website",
    "tags": ["website-signup", "newsletter-subscriber"],
    "created_at": "2026-01-05T16:00:00Z"
  }
}
```

**Save customer ID**:
```bash
export CUSTOMER_ID="01JGCUST123456789ABCDEF012"
```

---

### Step 2: Add Phone Number

**Sales rep calls the lead**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/phone \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1-555-0123"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCUST123456789ABCDEF012",
    "phone": "+1-555-0123"
  }
}
```

---

### Step 3: Qualify as Prospect

**Lead shows interest, qualify as prospect**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/qualify \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCUST123456789ABCDEF012",
    "status": "prospect",
    "status_changed_at": "2026-01-05T16:15:00Z"
  }
}
```

**What happened**:
- Status changed: `lead` → `prospect`
- `status_changed_at` timestamp updated
- Customer now appears in "Prospects" pipeline

---

### Step 4: Convert to Customer

**Prospect makes first purchase**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/convert \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCUST123456789ABCDEF012",
    "status": "customer",
    "tier": "free",
    "status_changed_at": "2026-01-05T16:30:00Z"
  }
}
```

**What happened**:
- Status changed: `prospect` → `customer`
- Customer now counted in active customer metrics
- Eligible for customer-only features

---

### Step 5: Upgrade Tier

**Customer upgrades to Pro tier**:

```bash
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/tier \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tier": "pro"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCUST123456789ABCDEF012",
    "tier": "pro",
    "tier_changed_at": "2026-01-05T17:00:00Z"
  }
}
```

**Available tiers**: `free`, `basic`, `pro`, `enterprise`

---

### Step 6: Add Tags

**Track customer segments**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/tags \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tag": "high-value"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCUST123456789ABCDEF012",
    "tags": ["website-signup", "newsletter-subscriber", "high-value"]
  }
}
```

---

### Step 7: Mark as Churned (Optional)

**Customer cancels subscription**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID/churn \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "price-too-high"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCUST123456789ABCDEF012",
    "status": "churned",
    "churn_reason": "price-too-high",
    "churned_at": "2026-01-05T18:00:00Z"
  }
}
```

---

## Use Case 2: B2B Customer Onboarding

**Scenario**: Onboard a company with multiple contacts.

### Step 1: Create Company

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/companies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "legal_name": "Acme Corp. LLC",
    "industry": "technology",
    "employee_count": 150,
    "website": "https://acme-corp.com",
    "tax_id": "12-3456789",
    "registration_number": "REG-2025-001",
    "address": {
      "street": "123 Tech Street",
      "city": "San Francisco",
      "state": "CA",
      "postal_code": "94105",
      "country": "US"
    }
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCOMP123456789ABCDEF012",
    "name": "Acme Corporation",
    "legal_name": "Acme Corp. LLC",
    "industry": "technology",
    "employee_count": 150,
    "website": "https://acme-corp.com",
    "tax_id": "12-3456789",
    "created_at": "2026-01-05T16:00:00Z"
  }
}
```

**Save company ID**:
```bash
export COMPANY_ID="01JGCOMP123456789ABCDEF012"
```

---

### Step 2: Create B2B Customer

**Link customer to company**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/b2b \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Smith",
    "email": "john@acme-corp.com",
    "company_id": "'"$COMPANY_ID"'",
    "customer_type": "b2b",
    "status": "prospect",
    "tier": "enterprise",
    "source": "sales-call"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGCUST987654321ZYXWVU098",
    "name": "John Smith",
    "email": "john@acme-corp.com",
    "company_id": "01JGCOMP123456789ABCDEF012",
    "customer_type": "b2b",
    "status": "prospect",
    "tier": "enterprise",
    "source": "sales-call",
    "created_at": "2026-01-05T16:10:00Z"
  }
}
```

---

### Step 3: Add Multiple Contacts

**Create decision maker**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/b2b \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@acme-corp.com",
    "company_id": "'"$COMPANY_ID"'",
    "customer_type": "b2b",
    "status": "customer",
    "tier": "enterprise",
    "tags": ["decision-maker"]
  }'
```

**Create technical contact**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers/b2b \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob Wilson",
    "email": "bob@acme-corp.com",
    "company_id": "'"$COMPANY_ID"'",
    "customer_type": "b2b",
    "status": "customer",
    "tier": "enterprise",
    "tags": ["technical-contact"]
  }'
```

---

### Step 4: Query Company Customers

**List all contacts for company**:

```bash
curl -X GET "http://localhost:8081/api/v1/customer-mgmt/customers?company_id=$COMPANY_ID" \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGCUST987654321ZYXWVU098",
      "name": "John Smith",
      "email": "john@acme-corp.com",
      "company_id": "01JGCOMP123456789ABCDEF012",
      "tags": []
    },
    {
      "id": "01JGCUST111222333444555666",
      "name": "Jane Doe",
      "email": "jane@acme-corp.com",
      "company_id": "01JGCOMP123456789ABCDEF012",
      "tags": ["decision-maker"]
    },
    {
      "id": "01JGCUST777888999000111222",
      "name": "Bob Wilson",
      "email": "bob@acme-corp.com",
      "company_id": "01JGCOMP123456789ABCDEF012",
      "tags": ["technical-contact"]
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 3
  }
}
```

---

## Use Case 3: Sales Pipeline Management

**Scenario**: Manage a deal from lead to closed-won.

### Step 1: Create Deal

**New sales opportunity**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/deals \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Q1 2026 Enterprise License",
    "customer_id": "'"$CUSTOMER_ID"'",
    "value": 50000,
    "currency": "USD",
    "stage": "lead",
    "sales_rep_id": "'"$USER_ID"'"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEAL123456789ABCDEF012",
    "name": "Q1 2026 Enterprise License",
    "customer_id": "01JGCUST123456789ABCDEF012",
    "value": 5000000,
    "currency": "USD",
    "stage": "lead",
    "probability": 10,
    "sales_rep_id": "01JGUSER123456789ABCDEF012",
    "created_at": "2026-01-05T16:00:00Z"
  }
}
```

**Save deal ID**:
```bash
export DEAL_ID="01JGDEAL123456789ABCDEF012"
```

**Deal stages**: `lead` (10%) → `qualified` (25%) → `proposal` (50%) → `negotiation` (75%) → `closed_won` (100%) / `closed_lost` (0%)

---

### Step 2: Qualify Deal

**After discovery call**:

```bash
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/stage \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stage": "qualified"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEAL123456789ABCDEF012",
    "stage": "qualified",
    "probability": 25,
    "stage_changed_at": "2026-01-05T16:30:00Z"
  }
}
```

**What happened**:
- Stage changed: `lead` → `qualified`
- Probability auto-updated: 10% → 25%
- Deal moved in pipeline view

---

### Step 3: Send Proposal

**Proposal sent to customer**:

```bash
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/stage \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stage": "proposal"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEAL123456789ABCDEF012",
    "stage": "proposal",
    "probability": 50,
    "stage_changed_at": "2026-01-06T10:00:00Z"
  }
}
```

---

### Step 4: Enter Negotiation

**Contract negotiation started**:

```bash
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/stage \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stage": "negotiation"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEAL123456789ABCDEF012",
    "stage": "negotiation",
    "probability": 75,
    "stage_changed_at": "2026-01-08T14:00:00Z"
  }
}
```

---

### Step 5: Close Deal (Won)

**Contract signed**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/win \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "actual_value": 48000,
    "notes": "Customer negotiated 4% discount"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEAL123456789ABCDEF012",
    "stage": "closed_won",
    "probability": 100,
    "actual_value": 4800000,
    "expected_close_date": "2026-01-31",
    "actual_close_date": "2026-01-10",
    "won_at": "2026-01-10T16:00:00Z"
  }
}
```

**What happened**:
- Stage changed to `closed_won`
- Probability set to 100%
- Actual close date recorded
- Revenue counted in sales metrics

---

### Alternative: Close Deal (Lost)

**If deal is lost**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/lose \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "went-with-competitor",
    "notes": "Customer chose cheaper alternative"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGDEAL123456789ABCDEF012",
    "stage": "closed_lost",
    "probability": 0,
    "loss_reason": "went-with-competitor",
    "lost_at": "2026-01-10T16:00:00Z"
  }
}
```

---

### Step 6: View Pipeline Statistics

**Get sales pipeline overview**:

```bash
curl -X GET http://localhost:8081/api/v1/customer-mgmt/analytics/pipeline \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "pipeline_by_stage": [
      {
        "stage": "lead",
        "count": 15,
        "total_value": 750000
      },
      {
        "stage": "qualified",
        "count": 8,
        "total_value": 400000
      },
      {
        "stage": "proposal",
        "count": 5,
        "total_value": 300000
      },
      {
        "stage": "negotiation",
        "count": 3,
        "total_value": 200000
      }
    ],
    "won_deals": {
      "count": 12,
      "total_value": 600000,
      "avg_days_to_close": 45
    },
    "lost_deals": {
      "count": 8,
      "total_value": 200000
    }
  }
}
```

---

## Use Case 4: Order Processing

**Scenario**: Create order, add products, confirm, and fulfill.

### Step 1: Create Order

```bash
curl -X POST http://localhost:8081/api/v1/order-mgmt/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "currency": "USD"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGORDER123456789ABCDEF01",
    "order_number": "ORD-2026-000001",
    "customer_id": "01JGCUST123456789ABCDEF012",
    "status": "pending",
    "currency": "USD",
    "total": 0,
    "line_items": [],
    "created_at": "2026-01-05T16:00:00Z"
  }
}
```

**Save order ID**:
```bash
export ORDER_ID="01JGORDER123456789ABCDEF01"
```

---

### Step 2: Add Line Items

**Add Product 1**:

```bash
curl -X POST http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/lines \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "01JGPROD111222333444555666",
    "quantity": 2,
    "unit_price": 99.99
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGORDER123456789ABCDEF01",
    "order_number": "ORD-2026-000001",
    "total": 19998,
    "line_items": [
      {
        "id": "01JGLINE111222333444555666",
        "product_id": "01JGPROD111222333444555666",
        "quantity": 2,
        "unit_price": 9999,
        "subtotal": 19998
      }
    ]
  }
}
```

**Note**: Amounts in cents (19998 = $199.98)

---

**Add Product 2**:

```bash
curl -X POST http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/lines \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "01JGPROD777888999000111222",
    "quantity": 1,
    "unit_price": 49.99
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGORDER123456789ABCDEF01",
    "total": 24997,
    "line_items": [
      {
        "id": "01JGLINE111222333444555666",
        "quantity": 2,
        "unit_price": 9999,
        "subtotal": 19998
      },
      {
        "id": "01JGLINE333444555666777888",
        "quantity": 1,
        "unit_price": 4999,
        "subtotal": 4999
      }
    ]
  }
}
```

**Total**: $249.97 (2 items)

---

### Step 3: Update Line Item Quantity

**Change quantity**:

```bash
curl -X PUT http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/lines/01JGLINE111222333444555666 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 3
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGORDER123456789ABCDEF01",
    "total": 34996,
    "line_items": [
      {
        "id": "01JGLINE111222333444555666",
        "quantity": 3,
        "unit_price": 9999,
        "subtotal": 29997
      },
      {
        "id": "01JGLINE333444555666777888",
        "quantity": 1,
        "unit_price": 4999,
        "subtotal": 4999
      }
    ]
  }
}
```

**New Total**: $349.96 (3 + 1 items)

---

### Step 4: Confirm Order

**Customer confirms order**:

```bash
curl -X POST http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/confirm \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGORDER123456789ABCDEF01",
    "order_number": "ORD-2026-000001",
    "status": "confirmed",
    "confirmed_at": "2026-01-05T16:30:00Z"
  }
}
```

**What happened**:
- Status changed: `pending` → `confirmed`
- Order locked (no more line item changes)
- Ready for processing

---

### Step 5: Process Order

**Order processing started**:

```bash
curl -X POST http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/process \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGORDER123456789ABCDEF01",
    "status": "processing",
    "processing_started_at": "2026-01-05T17:00:00Z"
  }
}
```

---

### Step 6: Fulfill Order

**Order shipped**:

```bash
curl -X POST http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/fulfill \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGORDER123456789ABCDEF01",
    "status": "fulfilled",
    "fulfilled_at": "2026-01-06T10:00:00Z"
  }
}
```

**Order lifecycle complete**: `pending` → `confirmed` → `processing` → `fulfilled`

---

### Alternative: Cancel Order

**If order needs to be cancelled**:

```bash
curl -X POST http://localhost:8081/api/v1/order-mgmt/orders/$ORDER_ID/cancel \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "customer-requested"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGORDER123456789ABCDEF01",
    "status": "cancelled",
    "cancel_reason": "customer-requested",
    "cancelled_at": "2026-01-05T16:45:00Z"
  }
}
```

---

## Use Case 5: Invoice & Payment Processing

**Scenario**: Generate invoice, record payment, process payment.

### Step 1: Create Invoice

```bash
curl -X POST http://localhost:8081/api/v1/billing/invoices \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "order_id": "'"$ORDER_ID"'",
    "currency": "USD",
    "due_date": "2026-02-05"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINV123456789ABCDEF0123",
    "invoice_number": "INV-2026-000001",
    "customer_id": "01JGCUST123456789ABCDEF012",
    "order_id": "01JGORDER123456789ABCDEF01",
    "status": "draft",
    "total_amount": 34996,
    "currency": "USD",
    "due_date": "2026-02-05",
    "created_at": "2026-01-05T16:00:00Z"
  }
}
```

**Save invoice ID**:
```bash
export INVOICE_ID="01JGINV123456789ABCDEF0123"
```

---

### Step 2: Send Invoice

**Email invoice to customer**:

```bash
curl -X POST http://localhost:8081/api/v1/billing/invoices/$INVOICE_ID/send \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINV123456789ABCDEF0123",
    "status": "sent",
    "sent_at": "2026-01-05T16:10:00Z"
  }
}
```

**What happened**:
- Status changed: `draft` → `sent`
- Email notification sent to customer
- Invoice now visible in customer portal

---

### Step 3: Record Payment

**Customer pays invoice**:

```bash
curl -X POST http://localhost:8081/api/v1/billing/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "invoice_id": "'"$INVOICE_ID"'",
    "amount": 349.96,
    "currency": "USD",
    "payment_method": "credit_card",
    "payment_date": "2026-01-10"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGPAY123456789ABCDEF01234",
    "payment_number": "PAY-2026-000001",
    "invoice_id": "01JGINV123456789ABCDEF0123",
    "amount": 34996,
    "currency": "USD",
    "payment_method": "credit_card",
    "status": "pending",
    "payment_date": "2026-01-10",
    "created_at": "2026-01-10T14:00:00Z"
  }
}
```

**Save payment ID**:
```bash
export PAYMENT_ID="01JGPAY123456789ABCDEF01234"
```

---

### Step 4: Process Payment

**Payment gateway confirms payment**:

```bash
curl -X POST http://localhost:8081/api/v1/billing/payments/$PAYMENT_ID/process \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "TXN-STRIPE-ABC123",
    "processor_response": "approved"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGPAY123456789ABCDEF01234",
    "status": "completed",
    "transaction_id": "TXN-STRIPE-ABC123",
    "processed_at": "2026-01-10T14:05:00Z"
  }
}
```

**What happened**:
- Payment status: `pending` → `completed`
- Transaction ID recorded
- Invoice automatically marked as `paid`

---

### Step 5: Verify Invoice Paid

**Check invoice status**:

```bash
curl -X GET http://localhost:8081/api/v1/billing/invoices/$INVOICE_ID \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINV123456789ABCDEF0123",
    "invoice_number": "INV-2026-000001",
    "status": "paid",
    "total_amount": 34996,
    "paid_amount": 34996,
    "paid_at": "2026-01-10T14:05:00Z"
  }
}
```

---

## Use Case 6: Subscription Management

**Scenario**: Create recurring subscription, process renewals, handle cancellation.

### Step 1: Create Subscription

```bash
curl -X POST http://localhost:8081/api/v1/billing/subscriptions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "plan": "pro-monthly",
    "billing_cycle": "monthly",
    "amount": 29.99,
    "currency": "USD",
    "start_date": "2026-01-01"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGSUB123456789ABCDEF01234",
    "customer_id": "01JGCUST123456789ABCDEF012",
    "plan": "pro-monthly",
    "billing_cycle": "monthly",
    "amount": 2999,
    "currency": "USD",
    "status": "active",
    "start_date": "2026-01-01",
    "next_billing_date": "2026-02-01",
    "created_at": "2026-01-05T16:00:00Z"
  }
}
```

**Save subscription ID**:
```bash
export SUBSCRIPTION_ID="01JGSUB123456789ABCDEF01234"
```

---

### Step 2: Activate Subscription

**Activate after trial or payment**:

```bash
curl -X POST http://localhost:8081/api/v1/billing/subscriptions/$SUBSCRIPTION_ID/activate \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGSUB123456789ABCDEF01234",
    "status": "active",
    "activated_at": "2026-01-05T16:10:00Z"
  }
}
```

---

### Step 3: Process Renewal

**Monthly renewal (automated)**:

```bash
curl -X POST http://localhost:8081/api/v1/billing/subscriptions/$SUBSCRIPTION_ID/renew \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGSUB123456789ABCDEF01234",
    "status": "active",
    "last_billing_date": "2026-02-01",
    "next_billing_date": "2026-03-01",
    "renewed_at": "2026-02-01T00:00:00Z"
  }
}
```

**What happened**:
- Invoice auto-generated
- Payment auto-charged
- Next billing date updated
- Subscription remains active

---

### Step 4: Pause Subscription

**Temporarily pause billing**:

```bash
curl -X POST http://localhost:8081/api/v1/billing/subscriptions/$SUBSCRIPTION_ID/pause \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGSUB123456789ABCDEF01234",
    "status": "paused",
    "paused_at": "2026-03-15T10:00:00Z"
  }
}
```

---

### Step 5: Resume Subscription

**Resume billing**:

```bash
curl -X POST http://localhost:8081/api/v1/billing/subscriptions/$SUBSCRIPTION_ID/resume \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGSUB123456789ABCDEF01234",
    "status": "active",
    "resumed_at": "2026-04-01T00:00:00Z",
    "next_billing_date": "2026-05-01"
  }
}
```

---

### Step 6: Cancel Subscription

**Customer cancels**:

```bash
curl -X POST http://localhost:8081/api/v1/billing/subscriptions/$SUBSCRIPTION_ID/cancel \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cancel_reason": "switching-to-competitor",
    "immediate": false
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGSUB123456789ABCDEF01234",
    "status": "cancelled",
    "cancel_reason": "switching-to-competitor",
    "cancelled_at": "2026-05-15T14:00:00Z",
    "end_date": "2026-06-01"
  }
}
```

**What happened**:
- Subscription marked as `cancelled`
- Service continues until end of billing period (2026-06-01)
- No renewal after end date

---

## Use Case 7: Customer Interaction Tracking

**Scenario**: Track all customer touchpoints (calls, emails, meetings).

### Step 1: Log Phone Call

**Outbound sales call**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "type": "call",
    "direction": "outbound",
    "subject": "Product demo discussion",
    "notes": "Discussed Pro plan features, customer interested in analytics module",
    "started_at": "2026-01-05T14:00:00Z",
    "ended_at": "2026-01-05T14:30:00Z",
    "outcome": "successful",
    "follow_up_required": true,
    "follow_up_date": "2026-01-12"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINT123456789ABCDEF01234",
    "customer_id": "01JGCUST123456789ABCDEF012",
    "type": "call",
    "direction": "outbound",
    "subject": "Product demo discussion",
    "notes": "Discussed Pro plan features, customer interested in analytics module",
    "started_at": "2026-01-05T14:00:00Z",
    "ended_at": "2026-01-05T14:30:00Z",
    "duration_sec": 1800,
    "outcome": "successful",
    "follow_up_required": true,
    "follow_up_date": "2026-01-12",
    "created_at": "2026-01-05T14:31:00Z"
  }
}
```

**Duration auto-calculated**: 30 minutes (1800 seconds)

---

### Step 2: Log Email

**Follow-up email sent**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "type": "email",
    "direction": "outbound",
    "subject": "Pro Plan Pricing and Analytics Features",
    "notes": "Sent detailed pricing breakdown and analytics module documentation",
    "started_at": "2026-01-05T15:00:00Z",
    "outcome": "successful"
  }'
```

---

### Step 3: Schedule Meeting

**In-person meeting scheduled**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "company_id": "'"$COMPANY_ID"'",
    "type": "meeting",
    "direction": "outbound",
    "subject": "Contract Negotiation Meeting",
    "notes": "Meeting with decision makers: Jane Doe (CFO), Bob Wilson (CTO)",
    "attendees": [
      {"name": "Jane Doe", "email": "jane@acme-corp.com", "role": "CFO"},
      {"name": "Bob Wilson", "email": "bob@acme-corp.com", "role": "CTO"}
    ],
    "started_at": "2026-01-12T10:00:00Z",
    "ended_at": "2026-01-12T11:30:00Z"
  }'
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINT456789012345678901234",
    "type": "meeting",
    "subject": "Contract Negotiation Meeting",
    "attendees": [
      {"name": "Jane Doe", "email": "jane@acme-corp.com", "role": "CFO"},
      {"name": "Bob Wilson", "email": "bob@acme-corp.com", "role": "CTO"}
    ],
    "duration_sec": 5400
  }
}
```

---

### Step 4: Add Quick Note

**Internal note**:

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "'"$CUSTOMER_ID"'",
    "type": "note",
    "subject": "Customer feedback",
    "notes": "Customer mentioned they need better reporting features. Flagged for product team."
  }'
```

---

### Step 5: Query Interaction History

**Get all interactions for customer**:

```bash
curl -X GET "http://localhost:8081/api/v1/customer-mgmt/interactions?customer_id=$CUSTOMER_ID" \
  -H "Authorization: Bearer $TOKEN"
```

**Response**:
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGINT123456789ABCDEF01234",
      "type": "call",
      "subject": "Product demo discussion",
      "started_at": "2026-01-05T14:00:00Z",
      "duration_sec": 1800,
      "outcome": "successful"
    },
    {
      "id": "01JGINT234567890123456789012",
      "type": "email",
      "subject": "Pro Plan Pricing and Analytics Features",
      "started_at": "2026-01-05T15:00:00Z"
    },
    {
      "id": "01JGINT456789012345678901234",
      "type": "meeting",
      "subject": "Contract Negotiation Meeting",
      "started_at": "2026-01-12T10:00:00Z",
      "duration_sec": 5400
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 3
  }
}
```

---

## Summary

### Customer Lifecycle
**Lead** → Add phone → **Prospect** → First purchase → **Customer** → Upgrade tier → **Churned** (optional)

### B2B Onboarding
**Company** → **Multiple contacts** → Link to company → Track by company_id

### Sales Pipeline
**Lead** (10%) → **Qualified** (25%) → **Proposal** (50%) → **Negotiation** (75%) → **Closed Won** (100%)

### Order Processing
**Pending** → Add lines → **Confirmed** → **Processing** → **Fulfilled** (or **Cancelled**)

### Billing
**Invoice** (draft → sent) → **Payment** (pending → completed) → Invoice auto-marked **paid**

### Subscription
**Create** → **Active** → **Renew** (monthly) → **Pause** → **Resume** → **Cancel**

### Interactions
Log **calls**, **emails**, **meetings**, **notes** → Query history → Track follow-ups

---

## Related Documentation

- **[Quick Start Guide](quick-start.md)** - Getting started with curl examples
- **[Authentication Flow](authentication-flow.md)** - JWT authentication details
- **[API Reference](../reference/api-reference.md)** - Complete endpoint documentation
- **[Customer Management](../concepts/customer-management.md)** - Customer context architecture
- **[Order Management](../concepts/order-management.md)** - Order context architecture
- **[Deal Management](../concepts/deal-management.md)** - Sales pipeline architecture

---

**Version**: 0.1.0  
**Last Updated**: January 5, 2026  
**Status**: Production-ready
