-- Order Management Context: Orders and Order Lines tables
-- Migration: 000001_orders.up.sql

-- ============================================================================
-- Orders Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS order_orders (
    -- Identity
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    order_number VARCHAR(50) UNIQUE NOT NULL,  -- ORD-2026-000001
    customer_id UUID NOT NULL,  -- References customer_customers (cross-context)
    company_id UUID NULL,       -- References customer_companies (B2B orders)

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
    contract_id UUID NULL,   -- References order_contracts
    invoice_id UUID NULL,    -- References billing_invoices (cross-context)

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

-- Indexes for performance
CREATE INDEX idx_orders_customer_id ON order_orders (customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_company_id ON order_orders (company_id) WHERE deleted_at IS NULL AND company_id IS NOT NULL;
CREATE INDEX idx_orders_status ON order_orders (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_order_date ON order_orders (order_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_created_at ON order_orders (created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_contract_id ON order_orders (contract_id) WHERE deleted_at IS NULL AND contract_id IS NOT NULL;
CREATE INDEX idx_orders_invoice_id ON order_orders (invoice_id) WHERE deleted_at IS NULL AND invoice_id IS NOT NULL;

-- Partial index for soft deletes
CREATE INDEX idx_orders_deleted_at ON order_orders (deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================================
-- Order Lines Table (line items)
-- ============================================================================

CREATE TABLE IF NOT EXISTS order_lines (
    -- Identity
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    order_id UUID NOT NULL REFERENCES order_orders (id) ON DELETE CASCADE,
    product_id UUID NOT NULL,  -- References warehouse_products (cross-context)

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

-- Indexes for performance
CREATE INDEX idx_order_lines_order_id ON order_lines (order_id);
CREATE INDEX idx_order_lines_product_id ON order_lines (product_id);
CREATE INDEX idx_order_lines_created_at ON order_lines (created_at DESC);

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE order_orders IS 'Orders aggregate: manages order lifecycle from creation to fulfillment';
COMMENT ON COLUMN order_orders.order_number IS 'Human-readable order number (ORD-YYYY-NNNNNN)';
COMMENT ON COLUMN order_orders.status IS 'Order status: pending → confirmed → processing → fulfilled (or cancelled)';
COMMENT ON COLUMN order_orders.customer_id IS 'References customer_customers (cross-context reference)';
COMMENT ON COLUMN order_orders.company_id IS 'Optional: for B2B orders, references customer_companies';
COMMENT ON COLUMN order_orders.contract_id IS 'Optional: associated contract (same context)';
COMMENT ON COLUMN order_orders.invoice_id IS 'Optional: associated invoice (billing context)';

COMMENT ON TABLE order_lines IS 'Order line items: products and quantities in an order';
COMMENT ON COLUMN order_lines.product_id IS 'References warehouse_products (cross-context reference)';
COMMENT ON COLUMN order_lines.total_amount IS 'Computed: quantity * unit_price';
