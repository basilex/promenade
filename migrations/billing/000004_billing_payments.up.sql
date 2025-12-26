-- Migration: Create billing_payments table
-- Description: Store payment transactions with refund tracking

CREATE TABLE IF NOT EXISTS billing_payments (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- References
    invoice_id UUID, -- Nullable: payments can exist without invoice (deposits, credits)
    user_id UUID NOT NULL,
    
    -- Transaction tracking
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    
    -- Payment amount
    amount BIGINT NOT NULL CHECK (amount > 0), -- Amount in cents/kopecks
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Payment status
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'refunded')),
    
    -- Payment method
    payment_method VARCHAR(50) NOT NULL CHECK (payment_method IN ('card', 'bank_transfer', 'paypal', 'stripe', 'liqpay')),
    
    -- Payment details (card last4, bank name, etc.)
    payment_details JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    -- Failure tracking
    failure_code VARCHAR(255),
    failure_message TEXT,
    
    -- Refund tracking
    refunded_amount BIGINT NOT NULL DEFAULT 0 CHECK (refunded_amount >= 0 AND refunded_amount <= amount),
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP, -- Soft delete
    
    -- Foreign keys
    CONSTRAINT fk_billing_payments_invoice FOREIGN KEY (invoice_id) REFERENCES billing_invoices(id) ON DELETE SET NULL,
    CONSTRAINT fk_billing_payments_user FOREIGN KEY (user_id) REFERENCES core_users(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_billing_payments_invoice_id ON billing_payments(invoice_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_payments_user_id ON billing_payments(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_payments_transaction_id ON billing_payments(transaction_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_payments_status ON billing_payments(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_payments_payment_method ON billing_payments(payment_method) WHERE deleted_at IS NULL;
CREATE INDEX idx_billing_payments_deleted_at ON billing_payments(deleted_at);

-- Composite index for user payment history
CREATE INDEX idx_billing_payments_user_status ON billing_payments(user_id, status) WHERE deleted_at IS NULL;

-- Comments for documentation
COMMENT ON TABLE billing_payments IS 'Payment transactions with support for multiple payment methods';
COMMENT ON COLUMN billing_payments.transaction_id IS 'Unique transaction ID from payment provider';
COMMENT ON COLUMN billing_payments.payment_details IS 'JSON storage for payment-specific data (card last4, etc.)';
COMMENT ON COLUMN billing_payments.refunded_amount IS 'Total amount refunded (can be partial)';
COMMENT ON COLUMN billing_payments.invoice_id IS 'Optional: payment can exist without invoice (deposits, credits)';
