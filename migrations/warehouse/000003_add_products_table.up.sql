-- ============================================================================
-- Warehouse Context: Products
-- ============================================================================
-- Product aggregate: Product catalog and specifications
-- Part of Warehouse Bounded Context
-- ============================================================================

-- Products table (Product aggregate)
CREATE TABLE IF NOT EXISTS warehouse_products (
    id UUID PRIMARY KEY,
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

-- ============================================================================
-- Indexes
-- ============================================================================

-- Primary lookup indexes
CREATE INDEX IF NOT EXISTS idx_warehouse_products_sku ON warehouse_products(sku) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_name ON warehouse_products(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_status ON warehouse_products(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_is_active ON warehouse_products(is_active) WHERE deleted_at IS NULL;

-- Classification indexes
CREATE INDEX IF NOT EXISTS idx_warehouse_products_category ON warehouse_products(category) WHERE deleted_at IS NULL AND category IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_brand ON warehouse_products(brand) WHERE deleted_at IS NULL AND brand IS NOT NULL;

-- Full-text search index for name and description
CREATE INDEX IF NOT EXISTS idx_warehouse_products_search ON warehouse_products USING gin(to_tsvector('english', name || ' ' || COALESCE(description, ''))) WHERE deleted_at IS NULL;

-- Timestamps
CREATE INDEX IF NOT EXISTS idx_warehouse_products_created_at ON warehouse_products(created_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_warehouse_products_updated_at ON warehouse_products(updated_at) WHERE deleted_at IS NULL;

-- Soft delete index
CREATE INDEX IF NOT EXISTS idx_warehouse_products_deleted_at ON warehouse_products(deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE warehouse_products IS 'Product catalog and specifications (Product aggregate)';
COMMENT ON COLUMN warehouse_products.id IS 'Primary key (UUID v7)';
COMMENT ON COLUMN warehouse_products.version IS 'Optimistic locking version';
COMMENT ON COLUMN warehouse_products.sku IS 'Stock Keeping Unit (unique identifier)';
COMMENT ON COLUMN warehouse_products.name IS 'Product name';
COMMENT ON COLUMN warehouse_products.description IS 'Product description';
COMMENT ON COLUMN warehouse_products.category IS 'Product category';
COMMENT ON COLUMN warehouse_products.brand IS 'Product brand';
COMMENT ON COLUMN warehouse_products.tags IS 'Product tags (array)';
COMMENT ON COLUMN warehouse_products.weight IS 'Weight in kg';
COMMENT ON COLUMN warehouse_products.length IS 'Length in cm';
COMMENT ON COLUMN warehouse_products.width IS 'Width in cm';
COMMENT ON COLUMN warehouse_products.height IS 'Height in cm';
COMMENT ON COLUMN warehouse_products.track_inventory IS 'If false, product has unlimited stock';
COMMENT ON COLUMN warehouse_products.allow_backorder IS 'Allow orders when out of stock';
COMMENT ON COLUMN warehouse_products.reorder_point IS 'Low stock threshold (triggers reorder alert)';
COMMENT ON COLUMN warehouse_products.reorder_quantity IS 'Quantity to order when below reorder point';
COMMENT ON COLUMN warehouse_products.track_serial_numbers IS 'Track individual units by serial number';
COMMENT ON COLUMN warehouse_products.track_lot_numbers IS 'Track batches by lot number';
COMMENT ON COLUMN warehouse_products.status IS 'Product status (active, draft, out_of_stock, discontinued)';
COMMENT ON COLUMN warehouse_products.is_active IS 'Active flag';
COMMENT ON COLUMN warehouse_products.is_deleted IS 'Soft delete flag';
COMMENT ON COLUMN warehouse_products.created_at IS 'Creation timestamp';
COMMENT ON COLUMN warehouse_products.updated_at IS 'Last update timestamp';
COMMENT ON COLUMN warehouse_products.deleted_at IS 'Soft delete timestamp';
