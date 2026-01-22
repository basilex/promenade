-- ============================================================================
-- Warehouse Context: Core Schema
-- ============================================================================
-- Complete warehouse domain schema: Products, Locations, Inventory, Stock Movements
-- Part of Warehouse Bounded Context
-- ============================================================================

-- ============================================================================
-- 1. PRODUCTS TABLE (Product aggregate)
-- ============================================================================
-- Product catalog and specifications
-- Must be created first as it's referenced by inventory

CREATE TABLE IF NOT EXISTS warehouse_products (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Identity
    sku VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Classification
    category VARCHAR(255),
    brand VARCHAR(255),
    tags TEXT,  -- JSON array of tags (database-agnostic)
    
    -- Physical properties
    weight DECIMAL(10,2) DEFAULT 0,  -- Weight in kg
    length DECIMAL(10,2) DEFAULT 0,  -- Length in cm
    width DECIMAL(10,2) DEFAULT 0,   -- Width in cm
    height DECIMAL(10,2) DEFAULT 0,  -- Height in cm
    
    -- Inventory settings
    track_inventory BOOLEAN NOT NULL DEFAULT TRUE,
    allow_backorder BOOLEAN NOT NULL DEFAULT FALSE,
    reorder_point INTEGER DEFAULT 0,
    reorder_quantity INTEGER DEFAULT 0,
    
    -- Serial/Lot tracking
    track_serial_numbers BOOLEAN NOT NULL DEFAULT FALSE,
    track_lot_numbers BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('active', 'draft', 'out_of_stock', 'discontinued')),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Product indexes
CREATE INDEX IF NOT EXISTS idx_warehouse_products_sku ON warehouse_products(sku) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_name ON warehouse_products(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_status ON warehouse_products(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_is_active ON warehouse_products(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_category ON warehouse_products(category) WHERE deleted_at IS NULL AND category IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_brand ON warehouse_products(brand) WHERE deleted_at IS NULL AND brand IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_created_at ON warehouse_products(created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_updated_at ON warehouse_products(updated_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_deleted_at ON warehouse_products(deleted_at) WHERE deleted_at IS NOT NULL;

-- Product comments
COMMENT ON TABLE warehouse_products IS 'Product catalog and specifications (Product aggregate)';
COMMENT ON COLUMN warehouse_products.version IS 'Optimistic locking version';
COMMENT ON COLUMN warehouse_products.sku IS 'Stock Keeping Unit (unique identifier)';
COMMENT ON COLUMN warehouse_products.track_inventory IS 'If false, product has unlimited stock';
COMMENT ON COLUMN warehouse_products.track_serial_numbers IS 'Track individual units by serial number';
COMMENT ON COLUMN warehouse_products.track_lot_numbers IS 'Track batches by lot number';

-- ============================================================================
-- 2. LOCATIONS TABLE (Location aggregate)
-- ============================================================================
-- Hierarchical warehouse location structure
-- warehouse → zone → aisle → rack → shelf → bin

CREATE TABLE IF NOT EXISTS warehouse_locations (
    -- Primary key
    id TEXT PRIMARY KEY,
    
    -- Business identifiers
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    
    -- Location type and hierarchy
    type VARCHAR(20) NOT NULL CHECK (type IN ('warehouse', 'zone', 'aisle', 'rack', 'shelf', 'bin')),
    parent_id TEXT REFERENCES warehouse_locations(id) ON DELETE RESTRICT,
    path TEXT NOT NULL,  -- Materialized path: /WH01/A/01
    level INT NOT NULL,  -- 0=warehouse, 1=zone, 2=aisle, etc.
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance', 'full', 'decommissioned')),
    
    -- Physical dimensions (meters)
    width DECIMAL(10,2),
    height DECIMAL(10,2),
    depth DECIMAL(10,2),
    
    -- Capacity management
    capacity INT,
    current_occupancy INT NOT NULL DEFAULT 0,
    is_limited BOOLEAN NOT NULL DEFAULT false,
    
    -- Operational flags
    is_pickable BOOLEAN NOT NULL DEFAULT false,
    is_putawayable BOOLEAN NOT NULL DEFAULT false,
    notes TEXT,
    
    -- Metadata
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Location indexes
CREATE INDEX IF NOT EXISTS idx_warehouse_locations_code ON warehouse_locations(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_locations_type ON warehouse_locations(type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_locations_parent ON warehouse_locations(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_locations_path ON warehouse_locations(path) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_locations_status ON warehouse_locations(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_locations_available ON warehouse_locations(status, is_putawayable, current_occupancy, capacity) 
    WHERE deleted_at IS NULL AND status = 'active';

-- Location comments
COMMENT ON TABLE warehouse_locations IS 'Hierarchical warehouse location structure (warehouse → zone → aisle → rack → shelf → bin)';
COMMENT ON COLUMN warehouse_locations.path IS 'Materialized path for fast hierarchy queries (e.g., /WH01/A/01)';
COMMENT ON COLUMN warehouse_locations.level IS 'Hierarchy level: 0=warehouse, 1=zone, 2=aisle, 3=rack, 4=shelf, 5=bin';
COMMENT ON COLUMN warehouse_locations.is_limited IS 'Whether location has capacity limits (true for bins with max units)';

-- ============================================================================
-- 3. INVENTORY TABLE (Inventory aggregate)
-- ============================================================================
-- Product stock management with warehouse/location tracking

CREATE TABLE IF NOT EXISTS warehouse_inventory (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Product identification
    product_id TEXT NOT NULL,
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
    last_updated_by TEXT NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Inventory indexes
CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_product_id ON warehouse_inventory(product_id);
CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_sku ON warehouse_inventory(sku) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_warehouse ON warehouse_inventory(warehouse_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_location ON warehouse_inventory(warehouse_id, location_code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_status ON warehouse_inventory(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_active ON warehouse_inventory(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_low_stock ON warehouse_inventory(quantity_available, reorder_point) WHERE deleted_at IS NULL AND is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_warehouse_inventory_product_warehouse ON warehouse_inventory(product_id, warehouse_id, location_code) WHERE deleted_at IS NULL;

-- Inventory comments
COMMENT ON TABLE warehouse_inventory IS 'Inventory tracking by product, warehouse, and location with optimistic locking';
COMMENT ON COLUMN warehouse_inventory.version IS 'Optimistic locking version - incremented on every update';
COMMENT ON COLUMN warehouse_inventory.quantity_available IS 'Calculated field: quantity_on_hand - quantity_reserved - quantity_committed';
COMMENT ON COLUMN warehouse_inventory.deleted_at IS 'Soft delete timestamp - NULL means not deleted';

-- ============================================================================
-- 4. STOCK MOVEMENTS TABLE (Immutable event log)
-- ============================================================================
-- Audit trail of all inventory movements
-- Must be created last as it references inventory

CREATE TABLE IF NOT EXISTS warehouse_stock_movements (
    -- Primary key
    id TEXT PRIMARY KEY,
    
    -- Foreign keys
    inventory_id TEXT NOT NULL REFERENCES warehouse_inventory(id) ON DELETE RESTRICT,
    
    -- Movement type (receipt, reservation, reservation_release, commit, adjustment, transfer, damage, return)
    type VARCHAR(50) NOT NULL CHECK (type IN ('receipt', 'reservation', 'reservation_release', 'commit', 'adjustment', 'transfer', 'damage', 'return')),
    
    -- Quantity tracking
    quantity INTEGER NOT NULL,
    quantity_before_move INTEGER NOT NULL,
    quantity_after_move INTEGER NOT NULL,
    
    -- Warehouse routing (for transfers)
    from_warehouse_id TEXT,
    from_location_code VARCHAR(50),
    to_warehouse_id TEXT,
    to_location_code VARCHAR(50),
    
    -- Reference tracking (order, purchase_order, adjustment, etc.)
    reference_type VARCHAR(50),
    reference_id TEXT,
    
    -- Cost tracking (for receipts)
    unit_cost_cents BIGINT,
    total_cost_cents BIGINT,
    currency_code VARCHAR(3),
    
    -- Audit fields
    reason TEXT,
    notes TEXT,
    created_by TEXT NOT NULL,
    movement_date TIMESTAMP NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Stock movement indexes
CREATE INDEX IF NOT EXISTS idx_stock_movements_inventory_id ON warehouse_stock_movements(inventory_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_type ON warehouse_stock_movements(type);
CREATE INDEX IF NOT EXISTS idx_stock_movements_reference ON warehouse_stock_movements(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_movement_date ON warehouse_stock_movements(movement_date);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON warehouse_stock_movements(created_at);

-- Stock movement comments
COMMENT ON TABLE warehouse_stock_movements IS 'Immutable audit trail of all inventory movements';

