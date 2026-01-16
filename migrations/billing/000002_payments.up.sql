-- Payments table
CREATE TABLE billing_payments (
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

-- Indexes for performance
CREATE INDEX idx_payments_customer_id ON billing_payments(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_invoice_id ON billing_payments(invoice_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_status ON billing_payments(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_payment_no ON billing_payments(payment_no) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_transaction_id ON billing_payments(transaction_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_created_at ON billing_payments(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_payments_processed_at ON billing_payments(processed_at DESC) WHERE deleted_at IS NULL;

-- Foreign key constraints
ALTER TABLE billing_payments
    ADD CONSTRAINT fk_payments_customer
    FOREIGN KEY (customer_id) REFERENCES customer_customers(id) ON DELETE CASCADE;

ALTER TABLE billing_payments
    ADD CONSTRAINT fk_payments_invoice
    FOREIGN KEY (invoice_id) REFERENCES billing_invoices(id) ON DELETE SET NULL;

-- Check constraints for enums
ALTER TABLE billing_payments
    ADD CONSTRAINT chk_payments_method
    CHECK (method IN ('credit_card', 'debit_card', 'bank_transfer', 'cash', 'paypal', 'stripe', 'square', 'check', 'wire', 'other'));

ALTER TABLE billing_payments
    ADD CONSTRAINT chk_payments_status
    CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'refunded', 'cancelled'));

-- Ensure refunded amount doesn't exceed payment amount
ALTER TABLE billing_payments
    ADD CONSTRAINT chk_payments_refund_amount
    CHECK (refunded_amount IS NULL OR refunded_amount <= amount);

-- Comment for documentation
COMMENT ON TABLE billing_payments IS 'Payment transactions for invoices and orders';
COMMENT ON COLUMN billing_payments.payment_no IS 'Human-readable payment number (PAY-YYYY-NNNNNN)';
COMMENT ON COLUMN billing_payments.transaction_id IS 'External payment processor transaction ID';
COMMENT ON COLUMN billing_payments.amount IS 'Payment amount in cents';
COMMENT ON COLUMN billing_payments.currency IS 'ISO 4217 currency code';
COMMENT ON COLUMN billing_payments.method IS 'Payment method used';
COMMENT ON COLUMN billing_payments.status IS 'Payment processing status';
COMMENT ON COLUMN billing_payments.refunded_amount IS 'Total refunded amount in cents';
COMMENT ON COLUMN billing_payments.processed_at IS 'When payment was processed';
COMMENT ON COLUMN billing_payments.refunded_at IS 'When payment was refunded';
