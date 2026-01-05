# Billing & Invoicing

**Status**:  Planned Q2 2026

Complete billing system with invoice generation, payment tracking, and subscription management.

---

## Overview

The **Billing** context will provide end-to-end billing capabilities, from invoice generation to payment processing and subscription lifecycle management.

---

## Planned Features

###  Invoice Management

**Invoice Generation**:
- Create invoices from orders automatically
- Manual invoice creation
- Recurring invoice schedules
- Pro-forma invoices for quotes

**Invoice Details**:
- Line items with quantity, price, tax
- Multiple tax rates (VAT, GST, sales tax)
- Discounts and promotions
- Currency support (multi-currency)

**Invoice Delivery**:
- Email with PDF attachment
- Customer portal access
- Print-ready PDF format
- Electronic invoicing (e-invoicing)

###  Payment Tracking

**Payment Methods**:
- Credit/debit cards (Stripe, PayPal integration)
- Bank transfer (ACH, SEPA)
- Wire transfer
- Cryptocurrency (optional)

**Payment Status**:
- Pending → Paid → Overdue workflow
- Partial payments tracking
- Payment reminders (automated)
- Late fees calculation

**Payment History**:
- Transaction log per customer
- Payment receipts generation
- Refund processing
- Chargeback handling

###  Subscription Management

**Subscription Plans**:
- Monthly, quarterly, annual billing
- Trial periods (7/14/30 days)
- Freemium plans
- Custom pricing per customer

**Lifecycle Management**:
- Auto-renewal with payment retry
- Upgrade/downgrade flow
- Proration calculations
- Cancellation and pause

**Subscription Analytics**:
- MRR (Monthly Recurring Revenue)
- Churn rate by plan
- Upgrade/downgrade trends
- LTV (Lifetime Value) calculation

###  Revenue Recognition

**Accounting Compliance**:
- ASC 606 / IFRS 15 standards
- Deferred revenue tracking
- Revenue recognition schedules
- Accrual accounting support

**Financial Reports**:
- Accounts Receivable (AR) aging
- Revenue by product/service
- Tax liability reports
- Revenue forecast

###  Automated Reminders

**Payment Reminders**:
- 7 days before due date
- On due date
- 7 days overdue
- 30 days overdue (final notice)

**Subscription Renewal**:
- 14 days before renewal
- Failed payment notifications
- Card expiration alerts
- Dunning management (retry failed payments)

###  Multi-Currency Support

**Currency Handling**:
- 150+ currencies supported
- Real-time exchange rates
- Customer preferred currency
- Multi-currency reporting

**Tax Compliance**:
- VAT calculation (EU)
- GST handling (India, Australia)
- US sales tax by state
- Tax exemptions management

---

## Technical Architecture

### Billing Engine

**Invoice Generation**:
- Template-based PDF generation
- Customizable invoice templates
- Watermarks and branding
- Multi-language support

**Payment Gateway Integration**:
- Stripe Connect
- PayPal Express Checkout
- Square
- Adyen (enterprise)

**Subscription Engine**:
- Cron-based renewal checks
- Payment retry logic (smart dunning)
- Webhook handlers for payment events
- Idempotent operations

### Data Model

**Core Entities**:
- Invoice (aggregate root)
- InvoiceLineItem (entity)
- Payment (aggregate root)
- Subscription (aggregate root)
- PaymentMethod (value object)

**Events**:
- `InvoiceCreated`
- `InvoicePaid`
- `PaymentFailed`
- `SubscriptionRenewed`
- `SubscriptionCancelled`

---

## Integration Points

### Order Management

**Order-to-Invoice**:
- Automatic invoice creation on order fulfillment
- Invoice status updates order
- Refunds trigger credit notes

### Customer Management

**Customer Billing**:
- Billing address from customer profile
- Payment history per customer
- Credit limit enforcement
- Billing preferences

### Event Bus

**Domain Events**:
- Payment notifications to accounting system
- Subscription events trigger onboarding
- Failed payments alert sales team
- Revenue recognition updates analytics

---

## Use Cases

### SaaS Subscription

**Monthly Billing**:
1. Customer subscribes to Pro plan ($99/month)
2. Trial period: 14 days free
3. Day 14: Charge first payment automatically
4. Every 30 days: Auto-renew and charge
5. Failed payment: Retry 3 times over 7 days
6. Still failed: Downgrade to Free plan

### E-commerce Orders

**Order Billing**:
1. Customer places order for $250
2. Invoice generated automatically
3. Email sent with payment link
4. Customer pays via credit card
5. Payment confirmed, order ships
6. Receipt emailed

### B2B Services

**Net-30 Payment Terms**:
1. Service delivered monthly
2. Invoice sent on 1st of next month
3. Payment due: 30 days (Net-30)
4. Reminder: 7 days before due
5. Overdue: Late fees applied (2% per month)
6. Accounting: AR aging report

---

## Roadmap

### Phase 1: Basic Invoicing (Q2 2026)

- [ ] Invoice generation (manual + auto)
- [ ] PDF export with templates
- [ ] Payment tracking (paid/pending/overdue)
- [ ] Email delivery
- [ ] Stripe integration

### Phase 2: Subscriptions (Q3 2026)

- [ ] Subscription plans management
- [ ] Auto-renewal engine
- [ ] Payment retry logic
- [ ] Upgrade/downgrade flow
- [ ] Trial period handling

### Phase 3: Advanced Features (Q4 2026)

- [ ] Multi-currency support
- [ ] Tax calculation engine
- [ ] Revenue recognition
- [ ] Dunning management
- [ ] PayPal, Square integrations

---

## Why Wait for This Module?

You can start using Promenade **today** and add billing later:

1. **Manual Invoicing**: Use external tools (QuickBooks, Xero)
2. **Payment Links**: Stripe Checkout or PayPal buttons
3. **Spreadsheet Tracking**: Export customer data to Excel
4. **Third-Party Integration**: Use Zapier to connect

When Billing module launches (Q2 2026), you'll get:
- Native invoice generation
- Automated subscription billing
- No third-party dependencies
- Full integration with Orders and Customers

---

## Related Documentation

- [Order Management](order-management.md) - Orders generate invoices
- [Customer Management](customer-management.md) - Customer billing data
- [Event-Driven Architecture](event-driven.md) - Payment events
- [RBAC](../guide/rbac.md) - Access control for billing

---

**Questions?** [Open a discussion](https://github.com/basilex/promenade/discussions)

**Want this sooner?** [Sponsor development](https://github.com/sponsors/basilex) to prioritize Billing module.
