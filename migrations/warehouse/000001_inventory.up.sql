-- ============================================================================
-- Warehouse Context: Inventory
-- ============================================================================
-- Inventory aggregate: Product stock management with warehouse/location tracking
-- Part of Warehouse Bounded Context
-- ============================================================================

-- Inventory table (Inventory aggregate)
CREATE TABLE IF NOT EXISTS warehouse_inventory (
    id UUID PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Product identification
    product_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL UNIQUE,
    product_name VARCHAR(255) NOT NULL,
    
    -- Quantity tracking
    quantity_on_hand INTEGER NOT NULL DEFAULT 0 CHECK (quantity_on_hand >= 0),
    quantity_reserved INTEGER NOT NULL DEFAULT 0 CHECK (quantity_reserved >= 0),
    quantity_committed INTEGER NOT NULL DEFAULT 0 CHECK (quantity_committed >= 0),
    quantity_available INTEGER NOT NULL DEFAULT 0 CHECK (quantity_available >= 0),
    
    -- Location
    warehouse_id VARCHAR(50) NOT NULL,
    location_code VARCHAR(100) NOT NULL,
    location_zone VARCHAR(50),  -- Zone for picking optimization (optional)
    
    -- Reorder settings
    reorder_point INTEGER NOT NULL DEFAULT 0,
    reorder_quantity INTEGER NOT NULL DEFAULT 0,
    last_restocked TIMESTAMP,  -- Last stock receipt date (optional)
    
    -- Cost tracking
    unit_cost_cents INTEGER NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'reserved', 'committed', 'on_order', 'damaged', 'quarantine')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT,  -- Admin notes (max 500 chars validated in application layer)
    
    -- Audit
    last_updated_by UUID NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_warehouse_inventory_product_id ON warehouse_inventory(product_id);
CREATE INDEX idx_warehouse_inventory_sku ON warehouse_inventory(sku) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_inventory_warehouse ON warehouse_inventory(warehouse_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_inventory_location ON warehouse_inventory(warehouse_id, location_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_inventory_status ON warehouse_inventory(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_inventory_active ON warehouse_inventory(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_inventory_low_stock ON warehouse_inventory(quantity_available, reorder_point) WHERE deleted_at IS NULL AND is_active = TRUE;

-- Composite index for common queries
CREATE INDEX idx_warehouse_inventory_product_warehouse ON warehouse_inventory(product_id, warehouse_id, location_code) WHERE deleted_at IS NULL;

-- Comment for documentation
COMMENT ON TABLE warehouse_inventory IS 'Inventory tracking by product, warehouse, and location with optimistic locking';
COMMENT ON COLUMN warehouse_inventory.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN warehouse_inventory.quantity_available IS 'Calculated field: quantity_on_hand - quantity_reserved - quantity_committed';
COMMENT ON COLUMN warehouse_inventory.deleted_at IS 'Soft delete timestamp - NULL means not deleted';
