-- Create order_contracts table for legal agreements
CREATE TABLE IF NOT EXISTS order_contracts (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES order_orders(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    terms TEXT NOT NULL,
    terms_url TEXT,
    version INTEGER NOT NULL DEFAULT 1,
    signature_id VARCHAR(255),
    signed_at TIMESTAMP,
    activated_at TIMESTAMP,
    completed_at TIMESTAMP,
    terminated_at TIMESTAMP,
    renewed_at TIMESTAMP,
    expires_at TIMESTAMP,
    termination_reason TEXT,
    signed_by_name VARCHAR(255),
    signed_by_email VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_contract_status CHECK (status IN ('draft', 'pending_signature', 'active', 'completed', 'terminated'))
);

-- Indexes for common queries
CREATE INDEX idx_order_contracts_order_id ON order_contracts(order_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_order_contracts_customer_id ON order_contracts(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_order_contracts_status ON order_contracts(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_order_contracts_expires_at ON order_contracts(expires_at) WHERE deleted_at IS NULL AND status = 'active';
CREATE INDEX idx_order_contracts_created_at ON order_contracts(created_at DESC) WHERE deleted_at IS NULL;

-- Comment on table
COMMENT ON TABLE order_contracts IS 'Legal agreements associated with orders (5-state lifecycle: draft → pending_signature → active → completed/terminated)';
