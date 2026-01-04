-- Create billing_invoices table
CREATE TABLE billing_invoices (
    id UUID PRIMARY KEY,
    invoice_no VARCHAR(50) NOT NULL UNIQUE,
    customer_id UUID NOT NULL,
    order_id UUID,
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

-- Create billing_invoice_lines table
CREATE TABLE billing_invoice_lines (
    id UUID PRIMARY KEY,
    invoice_id UUID NOT NULL,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price BIGINT NOT NULL DEFAULT 0,
    amount BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (invoice_id) REFERENCES billing_invoices(id) ON DELETE CASCADE
);

-- Indexes for billing_invoices
CREATE INDEX idx_billing_invoices_customer_id ON billing_invoices(customer_id);
CREATE INDEX idx_billing_invoices_order_id ON billing_invoices(order_id);
CREATE INDEX idx_billing_invoices_status ON billing_invoices(status);
CREATE INDEX idx_billing_invoices_due_date ON billing_invoices(due_date);
CREATE INDEX idx_billing_invoices_paid_date ON billing_invoices(paid_date);
CREATE INDEX idx_billing_invoices_deleted_at ON billing_invoices(deleted_at);
CREATE UNIQUE INDEX idx_billing_invoices_invoice_no ON billing_invoices(invoice_no) WHERE deleted_at IS NULL;

-- Indexes for billing_invoice_lines
CREATE INDEX idx_billing_invoice_lines_invoice_id ON billing_invoice_lines(invoice_id);

-- Add check constraints (SQLite compatible)
-- Note: SQLite doesn't support CHECK constraints in the same way as PostgreSQL
-- These will be enforced at the application level
