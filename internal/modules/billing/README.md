# Billing Module - Commercial

🇬🇧 **English** | [🇺🇦 Українська](README.uk.md) | [🇩🇪 Deutsch](README.de.md)

**Status**: Commercial - Production-ready subscription billing system  
**License**: Requires commercial license key  
**Version**: 1.0.0

---

## Overview

The Billing module provides a complete subscription billing and payment processing system for SaaS platforms. It handles billing plans, subscriptions, invoice generation, and payment tracking with comprehensive business logic and validation.

### Key Features

- **Flexible Billing Plans**: Monthly, quarterly, and annual billing intervals
- **Trial Periods**: Configurable trial periods with automatic conversion
- **Subscription Lifecycle**: Active, paused, cancelled, expired states
- **Automated Invoicing**: Automatic invoice generation on billing cycles
- **Payment Processing**: Payment tracking, reconciliation, and refunds
- **Comprehensive Testing**: 375 tests with 100% entity and usecase coverage

---

## Entities

### 1. Plan

Billing plan with pricing and features:

- **Fields**: name, description, price, currency, interval, trial_days, features
- **Intervals**: monthly (30 days), quarterly (90 days), annual (365 days)
- **Validation**: Price ≥ 0, valid currency code, valid interval
- **Use Case**: Define subscription tiers (Free, Basic, Pro, Enterprise)

### 2. Subscription

User subscription to a plan:

- **Fields**: user_id, plan_id, status, current_period_start, current_period_end, trial_start, trial_end
- **Status**: active, paused, cancelled, expired
- **Lifecycle**: Trial → Active → (Paused/Cancelled) → Expired
- **Auto-renewal**: Automatically renews unless cancelled

### 3. Invoice

Generated billing invoice:

- **Fields**: subscription_id, amount, currency, due_date, status, issued_at, paid_at
- **Status**: draft, pending, paid, overdue, void
- **Auto-generation**: Created at subscription start and renewal
- **Validation**: Amount > 0, valid due date

### 4. Payment

Payment transaction:

- **Fields**: invoice_id, amount, currency, method, status, transaction_id
- **Methods**: credit_card, debit_card, bank_transfer, paypal, stripe, etc.
- **Status**: pending, completed, failed, refunded
- **Reconciliation**: Links to invoices and external payment providers

---

## Use Cases

### Plan Management

```go
// Create billing plan
plan, err := planUC.CreatePlan(ctx, name, description, price, currency, interval, trialDays, features)

// Update plan pricing
err := planUC.UpdatePlan(ctx, planID, updates)

// Archive plan (soft delete)
err := planUC.ArchivePlan(ctx, planID)
```

### Subscription Management

```go
// Subscribe user to plan
subscription, err := subscriptionUC.Subscribe(ctx, userID, planID)

// Renew subscription
err := subscriptionUC.RenewSubscription(ctx, subscriptionID)

// Pause subscription
err := subscriptionUC.PauseSubscription(ctx, subscriptionID)

// Cancel subscription
err := subscriptionUC.CancelSubscription(ctx, subscriptionID)
```

### Invoice Management

```go
// Generate invoice for subscription
invoice, err := invoiceUC.GenerateInvoice(ctx, subscriptionID)

// Mark invoice as paid
err := invoiceUC.MarkInvoiceAsPaid(ctx, invoiceID, paidAt)

// Void invoice
err := invoiceUC.VoidInvoice(ctx, invoiceID)
```

### Payment Processing

```go
// Record payment
payment, err := paymentUC.CreatePayment(ctx, invoiceID, amount, currency, method, transactionID)

// Mark payment as completed
err := paymentUC.CompletePayment(ctx, paymentID)

// Refund payment
err := paymentUC.RefundPayment(ctx, paymentID, reason)
```

---

## Database Schema

### Migrations

The module includes 4 migrations in the `billing` namespace:

1. **000001_billing_plans.sql**: Billing plans table
2. **000002_billing_subscriptions.sql**: User subscriptions table
3. **000003_billing_invoices.sql**: Invoices table
4. **000004_billing_payments.sql**: Payment transactions table

**Run migrations**:

```bash
make migrate-module MODULE=billing
```

### Tables

#### billing_plans

```sql
CREATE TABLE billing_plans (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    billing_interval VARCHAR(20) NOT NULL, -- monthly, quarterly, annual
    trial_days INTEGER DEFAULT 0,
    features JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### billing_subscriptions

```sql
CREATE TABLE billing_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    plan_id UUID NOT NULL REFERENCES billing_plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### billing_invoices

```sql
CREATE TABLE billing_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    subscription_id UUID NOT NULL REFERENCES billing_subscriptions(id),
    amount DECIMAL(10,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    due_date DATE NOT NULL,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

#### billing_payments

```sql
CREATE TABLE billing_payments (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    invoice_id UUID NOT NULL REFERENCES billing_invoices(id),
    amount DECIMAL(10,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    transaction_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);
```

---

## Configuration

### config/config.dev.yaml

```yaml
module:
  name: "billing"
  enabled: true
  version: "1.0.0"

billing:
  default_currency: "USD"
  trial_enabled: true
  auto_invoice: true
  invoice_due_days: 7

purge:
  billing_subscriptions:
    retention_days: 2555 # ~7 years after cancellation
    enabled: true
  billing_invoices:
    retention_days: 2555 # ~7 years for audit compliance
    enabled: true
  billing_payments:
    retention_days: 1095 # ~3 years for financial records
    enabled: true
```

### Environment Variables

```bash
# License key (required for production)
export BILLING_LICENSE_KEY="PROMENADE-BILLING-PRO-20261231-SIGNATURE"
```

---

## Testing

### Test Coverage

- **Entity Tests**: 99 tests
  - Plan: 25 tests
  - Subscription: 31 tests
  - Invoice: 24 tests
  - Payment: 19 tests
- **Use Case Tests**: 276 tests
  - PlanUseCase: 69 tests
  - SubscriptionUseCase: 92 tests
  - InvoiceUseCase: 61 tests
  - PaymentUseCase: 54 tests

**Total**: 375 tests with 100% coverage

### Run Tests

```bash
# All billing tests
make test-module-billing

# Entity tests only
go test ./internal/modules/billing/domain/entity/...

# Use case tests only
go test ./internal/modules/billing/usecase/...
```

---

## HTTP API

### Endpoints

**Plans**:

- `GET /api/v1/billing/plans` - List plans
- `GET /api/v1/billing/plans/:id` - Get plan
- `POST /api/v1/billing/plans` - Create plan (admin)
- `PUT /api/v1/billing/plans/:id` - Update plan (admin)
- `DELETE /api/v1/billing/plans/:id` - Archive plan (admin)

**Subscriptions**:

- `GET /api/v1/billing/subscriptions` - List my subscriptions
- `GET /api/v1/billing/subscriptions/:id` - Get subscription
- `POST /api/v1/billing/subscriptions` - Subscribe to plan
- `POST /api/v1/billing/subscriptions/:id/renew` - Renew subscription
- `POST /api/v1/billing/subscriptions/:id/pause` - Pause subscription
- `POST /api/v1/billing/subscriptions/:id/cancel` - Cancel subscription

**Invoices**:

- `GET /api/v1/billing/invoices` - List my invoices
- `GET /api/v1/billing/invoices/:id` - Get invoice
- `POST /api/v1/billing/invoices` - Generate invoice (system)

**Payments**:

- `GET /api/v1/billing/payments` - List my payments
- `GET /api/v1/billing/payments/:id` - Get payment
- `POST /api/v1/billing/payments` - Create payment
- `POST /api/v1/billing/payments/:id/refund` - Refund payment

---

## License Requirement

This is a **commercial module** and requires a valid license key to use in production.

### Development Mode

For development, you can disable license validation:

```yaml
# config/modules.yaml
modules:
  config:
    billing:
      license_required: false
```

### Production License

Get a commercial license:

```bash
# Generate license (for authorized distributors only)
./scripts/generate-license.sh billing PRO 365

# Set license key
export BILLING_LICENSE_KEY="PROMENADE-BILLING-PRO-20261231-..."
```

**Contact**: alexander.vasilenko@gmail.com for licensing inquiries

---

## Dependencies

- **Core**: Users, Sessions, Permissions
- **External**: None (fully self-contained)
- **Optional**: Payment provider integrations (Stripe, PayPal)

---

## Integration Example

### 1. Enable Module

```yaml
# config/modules.yaml
modules:
  enabled:
    - billing
```

### 2. Run Migrations

```bash
make migrate-module MODULE=billing
```

### 3. Create Billing Plans

```bash
curl -X POST http://localhost:8081/api/v1/billing/plans \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pro Plan",
    "description": "Professional features",
    "price": 29.99,
    "currency": "USD",
    "billing_interval": "monthly",
    "trial_days": 14
  }'
```

### 4. Subscribe User

```bash
curl -X POST http://localhost:8081/api/v1/billing/subscriptions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plan_id": "01234567-89ab-cdef-0123-456789abcdef"
  }'
```

---

## Roadmap

- [ ] Payment provider integrations (Stripe, PayPal)
- [ ] Proration on plan changes
- [ ] Coupons and discounts
- [ ] Usage-based billing
- [ ] Multi-currency pricing
- [ ] Tax calculation (VAT, GST)

---

## Support

- **Documentation**: [docs/MODULE_DEVELOPMENT.md](../../../docs/MODULE_DEVELOPMENT.md)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Email**: alexander.vasilenko@gmail.com

---

**Built with Clean Architecture and extensive testing**  
**Production-ready for SaaS platforms**
