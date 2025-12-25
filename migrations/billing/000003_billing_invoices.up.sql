-- Migration: Create billing_invoices table
-- Description: Store billing invoices with payment tracking

CREATE TABLE IF NOT EXISTS billing_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- References
    subscription_id UUID NOT NULL,
    
    -- Invoice identification
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    
    -- Invoice status
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'open', 'paid', 'void', 'uncollectible')),
    
    -- Amount breakdown (all in cents/kopecks)
    subtotal_amount BIGINT NOT NULL CHECK (subtotal_amount >= 0),
    tax_amount BIGINT NOT NULL DEFAULT 0 CHECK (tax_amount >= 0),
    total_amount BIGINT NOT NULL CHECK (total_amount >= 0),
    amount_paid BIGINT NOT NULL DEFAULT 0 CHECK (amount_paid >= 0),
    amount_due BIGINT NOT NULL DEFAULT 0 CHECK (amount_due >= 0),
    
    -- Currency
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Payment tracking
    due_date TIMESTAMP,
    paid_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP, -- Soft delete
    
    -- Foreign keys
    CONSTRAINT fk_billing_invoices_subscription FOREIGN KEY (subscription_id) REFERENCES billing_subscriptions(id) ON DELETE CASCADE,
    
    -- Business constraints
    CONSTRAINT chk_billing_invoices_total CHECK (total_amount = subtotal_amount + tax_amount),
    CONSTRAINT chk_billing_invoices_due CHECK (amount_due = total_amount - amount_paid),
    CONSTRAINT chk_billing_invoices_paid_amount CHECK (amount_paid <= total_amount)
);

-- Indexes for performance
CREATE INDEX idx_billing_invoices_subscription_id ON billing_invoices(subscription_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_invoices_invoice_number ON billing_invoices(invoice_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_invoices_status ON billing_invoices(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_invoices_due_date ON billing_invoices(due_date) WHERE deleted_at IS NULL AND status IN ('open', 'past_due');
CREATE INDEX idx_billing_invoices_deleted_at ON billing_invoices(deleted_at);

-- Comments for documentation
COMMENT ON TABLE billing_invoices IS 'Billing invoices with automatic amount tracking';
COMMENT ON COLUMN billing_invoices.invoice_number IS 'Unique invoice number (format: INV-YYYYMMDD-XXXXXX)';
COMMENT ON COLUMN billing_invoices.subtotal_amount IS 'Amount before tax in smallest currency unit';
COMMENT ON COLUMN billing_invoices.total_amount IS 'Total amount including tax (subtotal + tax)';
COMMENT ON COLUMN billing_invoices.amount_due IS 'Remaining amount to be paid (total - paid)';
