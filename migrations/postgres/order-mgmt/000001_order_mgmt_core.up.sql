-- ============================================================================
-- Order Management Context: Core Schema
-- ============================================================================
-- Complete order domain schema: Orders, Order Lines, Contracts
-- Part of Order Management Bounded Context
-- ============================================================================

-- ============================================================================
-- 1. ORDERS TABLE (Order aggregate root)
-- ============================================================================
-- Order lifecycle: pending → confirmed → processing → fulfilled (or cancelled)

CREATE TABLE IF NOT EXISTS order_orders (
    -- Identity
    id TEXT PRIMARY KEY,
    order_number VARCHAR(50) UNIQUE NOT NULL,  -- ORD-2026-000001
    customer_id TEXT NOT NULL,  -- References customer_customers (cross-context)
    company_id TEXT NULL,       -- References customer_companies (B2B orders)

    -- Order Details
    total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',

    -- Status tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- Allowed values: pending, confirmed, processing, fulfilled, cancelled

    -- Important dates
    order_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP WITH TIME ZONE NULL,
    fulfilled_at TIMESTAMP WITH TIME ZONE NULL,
    cancelled_at TIMESTAMP WITH TIME ZONE NULL,

    -- References (to other contexts)
    contract_id TEXT NULL,   -- References order_contracts
    invoice_id TEXT NULL,    -- References billing_invoices (cross-context)

    -- Lifecycle timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    -- Constraints
    CONSTRAINT chk_order_status CHECK (
        status IN ('pending', 'confirmed', 'processing', 'fulfilled', 'cancelled')
    ),
    CONSTRAINT chk_order_total_positive CHECK (total_amount >= 0),
    CONSTRAINT chk_order_currency_length CHECK (LENGTH(currency) = 3)
);

-- Order indexes
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON order_orders (customer_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_company_id ON order_orders (company_id) WHERE deleted_at IS NULL AND company_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_status ON order_orders (status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_order_date ON order_orders (order_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON order_orders (created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_orders_contract_id ON order_orders (contract_id) WHERE deleted_at IS NULL AND contract_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_invoice_id ON order_orders (invoice_id) WHERE deleted_at IS NULL AND invoice_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON order_orders (deleted_at) WHERE deleted_at IS NOT NULL;

-- Order comments
COMMENT ON TABLE order_orders IS 'Orders aggregate: manages order lifecycle from creation to fulfillment';
COMMENT ON COLUMN order_orders.order_number IS 'Human-readable order number (ORD-YYYY-NNNNNN)';
COMMENT ON COLUMN order_orders.status IS 'Order status: pending → confirmed → processing → fulfilled (or cancelled)';
COMMENT ON COLUMN order_orders.customer_id IS 'References customer_customers (cross-context reference)';
COMMENT ON COLUMN order_orders.company_id IS 'Optional: for B2B orders, references customer_companies';
COMMENT ON COLUMN order_orders.contract_id IS 'Optional: associated contract (same context)';
COMMENT ON COLUMN order_orders.invoice_id IS 'Optional: associated invoice (billing context)';

-- ============================================================================
-- 2. ORDER LINES TABLE (Order line items)
-- ============================================================================
-- Products and quantities in an order

CREATE TABLE IF NOT EXISTS order_lines (
    -- Identity
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES order_orders (id) ON DELETE CASCADE,
    product_id TEXT NOT NULL,  -- References warehouse_products (cross-context)

    -- Line details
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(15, 2) NOT NULL,
    total_amount DECIMAL(15, 2) NOT NULL,  -- quantity * unit_price
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',

    -- Lifecycle timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT chk_line_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_line_unit_price_positive CHECK (unit_price >= 0),
    CONSTRAINT chk_line_total_positive CHECK (total_amount >= 0),
    CONSTRAINT chk_line_currency_length CHECK (LENGTH(currency) = 3)
);

-- Order line indexes
CREATE INDEX IF NOT EXISTS idx_order_lines_order_id ON order_lines (order_id);
CREATE INDEX IF NOT EXISTS idx_order_lines_product_id ON order_lines (product_id);
CREATE INDEX IF NOT EXISTS idx_order_lines_created_at ON order_lines (created_at DESC);

-- Order line comments
COMMENT ON TABLE order_lines IS 'Order line items: products and quantities in an order';
COMMENT ON COLUMN order_lines.product_id IS 'References warehouse_products (cross-context reference)';
COMMENT ON COLUMN order_lines.total_amount IS 'Computed: quantity * unit_price';

-- ============================================================================
-- 3. CONTRACTS TABLE (Contract aggregate)
-- ============================================================================
-- Legal agreements associated with orders
-- Lifecycle: draft → pending_signature → active → completed/terminated

CREATE TABLE IF NOT EXISTS order_contracts (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES order_orders(id) ON DELETE CASCADE,
    customer_id TEXT NOT NULL,
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

-- Contract indexes
CREATE INDEX IF NOT EXISTS idx_order_contracts_order_id ON order_contracts(order_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_order_contracts_customer_id ON order_contracts(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_order_contracts_status ON order_contracts(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_order_contracts_expires_at ON order_contracts(expires_at) WHERE deleted_at IS NULL AND status = 'active';
CREATE INDEX IF NOT EXISTS idx_order_contracts_created_at ON order_contracts(created_at DESC) WHERE deleted_at IS NULL;

-- Contract comments
COMMENT ON TABLE order_contracts IS 'Legal agreements associated with orders (5-state lifecycle: draft → pending_signature → active → completed/terminated)';
