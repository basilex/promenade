-- ============================================================================
-- Customer Management Context: Sales Report Read Model (Analytics)
-- ============================================================================
-- Denormalized read model for sales reporting
-- ============================================================================

CREATE TABLE IF NOT EXISTS analytics_sales_orders (
    order_id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    currency_code CHAR(3) NOT NULL,
    total_cents INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'paid', 'fulfilled', 'cancelled')),
    manager_id TEXT,
    confirmed_at TIMESTAMP NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_analytics_sales_orders_customer ON analytics_sales_orders(customer_id);
CREATE INDEX idx_analytics_sales_orders_status ON analytics_sales_orders(status);
CREATE INDEX idx_analytics_sales_orders_manager ON analytics_sales_orders(manager_id) WHERE manager_id IS NOT NULL;
CREATE INDEX idx_analytics_sales_orders_confirmed_at ON analytics_sales_orders(confirmed_at DESC);

COMMENT ON TABLE analytics_sales_orders IS 'Sales report read model (orders)';

CREATE TABLE IF NOT EXISTS analytics_sales_order_items (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_cents INTEGER NOT NULL,
    total_cents INTEGER NOT NULL,
    category_id TEXT,
    manager_id TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_analytics_sales_items_order ON analytics_sales_order_items(order_id);
CREATE INDEX idx_analytics_sales_items_product ON analytics_sales_order_items(product_id);
CREATE INDEX idx_analytics_sales_items_category ON analytics_sales_order_items(category_id) WHERE category_id IS NOT NULL;
CREATE INDEX idx_analytics_sales_items_manager ON analytics_sales_order_items(manager_id) WHERE manager_id IS NOT NULL;

COMMENT ON TABLE analytics_sales_order_items IS 'Sales report read model (order line items)';
