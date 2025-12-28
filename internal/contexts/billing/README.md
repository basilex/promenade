# Billing Context

**Domain:** Subscriptions, invoicing, payments  
**Ubiquitous Language:** Subscription, Invoice, Payment, Plan, Billing Cycle  
**Status:**  Migration in Progress (from modules/billing)

---

## Overview

The **Billing Context** handles all financial transactions, subscriptions, and payment processing. This context is currently implemented as `internal/modules/billing/` and will be gradually migrated to this bounded context structure.

### Responsibilities

 **What Billing Context DOES:**

- Subscription lifecycle management
- Invoice generation and management
- Payment processing and reconciliation
- Pricing and plan management
- Revenue recognition
- Payment method management

 **What Billing Context DOES NOT DO:**

- Customer relationships → **Customer Management Context**
- Order fulfillment → **Order Management Context**
- Product catalog → **Warehouse Context**

---

## Current Implementation

**Location:** `internal/modules/billing/`

**Status:**  Production-ready (375 tests, 100% passing)

**Entities:**

- Plan - Subscription plans with pricing
- Subscription - Customer subscriptions
- Invoice - Generated invoices
- Payment - Payment records

**See:** `internal/modules/billing/README.md` for current implementation details

---

## Migration Plan

### Phase 1: Keep Module Working

-  Module fully functional
-  375 comprehensive tests
-  Production-ready

### Phase 2: Context Structure (Future)

```
internal/contexts/billing/
   subscription/
      entity.go           # Subscription aggregate
      repository.go
      usecase.go
      handler.go
   invoice/
      entity.go           # Invoice aggregate
      ...
   payment/
       entity.go           # Payment aggregate
       ...
```

### Phase 3: Gradual Route Migration

- Keep module routes: `/api/v1/billing/*`
- Add context handlers
- Module handlers become facade
- Deprecate module handlers
- Remove module code

---

## Aggregates (Target Structure)

### 1. Subscription Aggregate

**Current:** `internal/modules/billing/domain/entity/subscription.go`  
**Target:** `internal/contexts/billing/subscription/entity.go`

**Business Rules:**

- One active subscription per customer per plan
- Trial periods convert to paid subscriptions
- Paused subscriptions retain data
- Cancelled subscriptions have grace period

---

### 2. Invoice Aggregate

**Current:** `internal/modules/billing/domain/entity/invoice.go`  
**Target:** `internal/contexts/billing/invoice/entity.go`

**Business Rules:**

- Invoices are immutable once finalized
- Credit notes for refunds
- Due dates based on payment terms
- Automatic payment retry logic

---

### 3. Payment Aggregate

**Current:** `internal/modules/billing/domain/entity/payment.go`  
**Target:** `internal/contexts/billing/payment/entity.go`

**Business Rules:**

- Payments reference invoices
- Refunds create negative payments
- Payment methods validated before use
- Transaction IDs from payment gateway

---

## Domain Events

```go
// Subscription events
type SubscriptionCreatedEvent struct {
    SubscriptionID uuidv7.UUID
    CustomerID     uuidv7.UUID
    PlanID         uuidv7.UUID
    Timestamp      time.Time
}

type SubscriptionCancelledEvent struct {
    SubscriptionID uuidv7.UUID
    Reason         string
    Timestamp      time.Time
}

// Invoice events
type InvoiceGeneratedEvent struct {
    InvoiceID      uuidv7.UUID
    CustomerID     uuidv7.UUID
    Amount         valueobject.Money
    DueDate        time.Time
    Timestamp      time.Time
}

type InvoicePaidEvent struct {
    InvoiceID  uuidv7.UUID
    PaymentID  uuidv7.UUID
    Amount     valueobject.Money
    Timestamp  time.Time
}

// Payment events
type PaymentProcessedEvent struct {
    PaymentID     uuidv7.UUID
    InvoiceID     uuidv7.UUID
    Amount        valueobject.Money
    Status        string
    Timestamp     time.Time
}
```

---

## Communication with Other Contexts

### With Customer Management

```go
// Billing subscribes to customer creation
eventBus.Subscribe("customer.created", func(event CustomerCreatedEvent) {
    // Create default free subscription
    billingService.CreateDefaultSubscription(event.CustomerID)
})

// Billing publishes payment failures
eventBus.Publish("payment.failed", PaymentFailedEvent{...})

// CustomerMgmt subscribes to update customer status
eventBus.Subscribe("payment.failed", func(event PaymentFailedEvent) {
    customerService.MarkPaymentIssue(event.CustomerID)
})
```

### With Order Management

```go
// OrderMgmt creates invoice for order
eventBus.Publish("order.confirmed", OrderConfirmedEvent{...})

// Billing subscribes and generates invoice
eventBus.Subscribe("order.confirmed", func(event OrderConfirmedEvent) {
    billingService.GenerateInvoiceForOrder(event.OrderID)
})
```

---

## Current API Endpoints

**Base:** `/api/v1/billing/`

```
# Plans
GET    /api/v1/billing/plans
POST   /api/v1/billing/plans
GET    /api/v1/billing/plans/:id
PUT    /api/v1/billing/plans/:id

# Subscriptions
GET    /api/v1/billing/subscriptions
POST   /api/v1/billing/subscriptions
GET    /api/v1/billing/subscriptions/:id
PATCH  /api/v1/billing/subscriptions/:id/pause
PATCH  /api/v1/billing/subscriptions/:id/resume
PATCH  /api/v1/billing/subscriptions/:id/cancel

# Invoices
GET    /api/v1/billing/invoices
POST   /api/v1/billing/invoices
GET    /api/v1/billing/invoices/:id
PATCH  /api/v1/billing/invoices/:id/finalize
PATCH  /api/v1/billing/invoices/:id/void

# Payments
GET    /api/v1/billing/payments
POST   /api/v1/billing/payments
GET    /api/v1/billing/payments/:id
POST   /api/v1/billing/payments/:id/refund
```

---

## Testing

**Current:** 375 tests (100% passing)

- Entity tests: 56 tests
- UseCase tests: 107 tests
- Repository tests: (integration)
- Handler tests: (HTTP)

**Coverage:** 100% for entities and use cases

---

## Migration Timeline

**Phase 1: Current (Complete)** 

- Module fully functional
- Production-ready
- Comprehensive tests

**Phase 2: Context Planning (Q1 2026)** 

- Design context structure
- Plan migration strategy
- Create migration guide

**Phase 3: Parallel Implementation (Q2 2026)** 

- Implement context structure
- Keep module working
- Route both to same logic

**Phase 4: Gradual Migration (Q3 2026)** 

- Migrate routes to context
- Module as facade
- Deprecation warnings

**Phase 5: Cleanup (Q4 2026)** 

- Remove module code
- Update documentation
- Final testing

---

**Status:**  Module-based (migration planned)  
**Current Location:** `internal/modules/billing/`  
**Future Location:** `internal/contexts/billing/`  
**Tests:** 375 tests (100% passing)  
**Documentation:** See `internal/modules/billing/README.md`
