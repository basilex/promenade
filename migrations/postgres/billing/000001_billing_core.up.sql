-- ============================================================================
-- Billing Context: Core Schema
-- ============================================================================
-- Complete billing domain schema: Invoices, Payments, Subscriptions
-- Part of Billing Bounded Context
-- ============================================================================

-- ============================================================================
-- 1. INVOICES TABLE (Invoice aggregate)
-- ============================================================================
-- Invoice header with subtotal, tax, and total amounts

CREATE TABLE IF NOT EXISTS billing_invoices (
    id TEXT PRIMARY KEY,
    invoice_no VARCHAR(50) NOT NULL UNIQUE,
    customer_id TEXT NOT NULL,
    order_id TEXT,
    subtotal_amount BIGINT NOT NULL DEFAULT 0,
    tax_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    issue_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_date TIMESTAMP NOT NULL,
    paid_date TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Invoice lines (child entity)
CREATE TABLE IF NOT EXISTS billing_invoice_lines (
    id TEXT PRIMARY KEY,
    invoice_id TEXT NOT NULL,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price BIGINT NOT NULL DEFAULT 0,
    amount BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (invoice_id) REFERENCES billing_invoices(id) ON DELETE CASCADE
);

-- Invoice indexes
CREATE INDEX IF NOT EXISTS idx_billing_invoices_customer_id ON billing_invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_billing_invoices_order_id ON billing_invoices(order_id);
CREATE INDEX IF NOT EXISTS idx_billing_invoices_status ON billing_invoices(status);
CREATE INDEX IF NOT EXISTS idx_billing_invoices_due_date ON billing_invoices(due_date);
CREATE INDEX IF NOT EXISTS idx_billing_invoices_paid_date ON billing_invoices(paid_date);
CREATE INDEX IF NOT EXISTS idx_billing_invoices_deleted_at ON billing_invoices(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_billing_invoices_invoice_no ON billing_invoices(invoice_no) WHERE deleted_at IS NULL;

-- Invoice line indexes
CREATE INDEX IF NOT EXISTS idx_billing_invoice_lines_invoice_id ON billing_invoice_lines(invoice_id);

-- ============================================================================
-- 2. PAYMENTS TABLE (Payment aggregate)
-- ============================================================================
-- Payment transactions for invoices and orders

CREATE TABLE IF NOT EXISTS billing_payments (
    id TEXT PRIMARY KEY,
    payment_no VARCHAR(20) UNIQUE NOT NULL,
    transaction_id VARCHAR(100),
    customer_id TEXT NOT NULL,
    invoice_id TEXT,
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    method VARCHAR(50) NOT NULL,
    card_last4 VARCHAR(4),
    card_brand VARCHAR(20),
    bank_account VARCHAR(100),
    payment_provider VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    failure_reason TEXT,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    refunded_at TIMESTAMP,
    refunded_amount BIGINT CHECK (refunded_amount >= 0),
    processed_by VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Payment indexes
CREATE INDEX IF NOT EXISTS idx_payments_customer_id ON billing_payments(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_invoice_id ON billing_payments(invoice_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_status ON billing_payments(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_payment_no ON billing_payments(payment_no) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON billing_payments(transaction_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON billing_payments(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payments_processed_at ON billing_payments(processed_at DESC) WHERE deleted_at IS NULL;

-- Payment foreign keys
ALTER TABLE billing_payments
    ADD CONSTRAINT fk_payments_customer
    FOREIGN KEY (customer_id) REFERENCES customer_customers(id) ON DELETE CASCADE;

ALTER TABLE billing_payments
    ADD CONSTRAINT fk_payments_invoice
    FOREIGN KEY (invoice_id) REFERENCES billing_invoices(id) ON DELETE SET NULL;

-- Payment check constraints
ALTER TABLE billing_payments
    ADD CONSTRAINT chk_payments_method
    CHECK (method IN ('credit_card', 'debit_card', 'bank_transfer', 'cash', 'paypal', 'stripe', 'square', 'check', 'wire', 'other'));

ALTER TABLE billing_payments
    ADD CONSTRAINT chk_payments_status
    CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'refunded', 'cancelled'));

ALTER TABLE billing_payments
    ADD CONSTRAINT chk_payments_refund_amount
    CHECK (refunded_amount IS NULL OR refunded_amount <= amount);

-- Payment comments
COMMENT ON TABLE billing_payments IS 'Payment transactions for invoices and orders';
COMMENT ON COLUMN billing_payments.payment_no IS 'Human-readable payment number (PAY-YYYY-NNNNNN)';
COMMENT ON COLUMN billing_payments.transaction_id IS 'External payment processor transaction ID';
COMMENT ON COLUMN billing_payments.amount IS 'Payment amount in cents';
COMMENT ON COLUMN billing_payments.status IS 'Payment processing status';
COMMENT ON COLUMN billing_payments.refunded_amount IS 'Total refunded amount in cents';

-- ============================================================================
-- 3. SUBSCRIPTIONS TABLE (Subscription aggregate)
-- ============================================================================
-- Recurring billing subscriptions with auto-renewal

CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id TEXT PRIMARY KEY,
    subscription_no VARCHAR(15) NOT NULL UNIQUE, -- SUB-YYYY-NNNNNN format
    customer_id TEXT NOT NULL, -- Reference to customer (no FK for context isolation)
    plan_id VARCHAR(100) NOT NULL, -- Subscription plan identifier
    status VARCHAR(20) NOT NULL CHECK (status IN ('trial', 'active', 'paused', 'cancelled', 'expired')),
    billing_period VARCHAR(20) NOT NULL CHECK (billing_period IN ('monthly', 'quarterly', 'yearly')),
    
    -- Financial details
    currency VARCHAR(3) NOT NULL, -- ISO 4217
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0), -- Subscription price in cents
    
    -- Dates
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NULL, -- Nullable for active subscriptions
    renewal_date TIMESTAMP NOT NULL,
    trial_end_date TIMESTAMP NULL, -- Nullable if no trial
    
    -- Cancellation
    cancelled_at TIMESTAMP NULL,
    cancel_reason TEXT NULL,
    cancellation_effective_date TIMESTAMP NULL,
    
    -- Metadata (TEXT for database-agnostic JSON storage)
    metadata TEXT DEFAULT '{}', -- Additional subscription metadata stored as JSON
    
    -- Standard audit fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Subscription indexes
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_customer_id ON billing_subscriptions(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_status ON billing_subscriptions(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_renewal_date ON billing_subscriptions(renewal_date) WHERE deleted_at IS NULL AND status = 'active';
CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_subscription_no ON billing_subscriptions(subscription_no) WHERE deleted_at IS NULL;

-- Subscription comments
COMMENT ON TABLE billing_subscriptions IS 'Recurring billing subscriptions with auto-renewal and lifecycle management';
COMMENT ON COLUMN billing_subscriptions.subscription_no IS 'Unique subscription identifier in SUB-YYYY-NNNNNN format';
COMMENT ON COLUMN billing_subscriptions.customer_id IS 'Reference to customer (no FK for bounded context isolation)';
COMMENT ON COLUMN billing_subscriptions.status IS 'Subscription lifecycle status: trial, active, paused, cancelled, expired';
COMMENT ON COLUMN billing_subscriptions.billing_period IS 'Renewal frequency: monthly, quarterly, yearly';
COMMENT ON COLUMN billing_subscriptions.amount_cents IS 'Subscription price in cents for precise currency handling';
