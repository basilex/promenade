-- Add warehouse_locations table for hierarchical location management
-- Location types: warehouse, zone, aisle, rack, shelf, bin
-- Supports multi-level hierarchy with path-based queries

CREATE TABLE warehouse_locations (
    -- Primary key
    id UUID PRIMARY KEY,
    
    -- Business identifiers
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    
    -- Location type and hierarchy
    type VARCHAR(20) NOT NULL CHECK (type IN ('warehouse', 'zone', 'aisle', 'rack', 'shelf', 'bin')),
    parent_id UUID REFERENCES warehouse_locations(id) ON DELETE RESTRICT,
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

-- Indexes for performance
CREATE INDEX idx_warehouse_locations_code ON warehouse_locations(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_locations_type ON warehouse_locations(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_locations_parent ON warehouse_locations(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_locations_path ON warehouse_locations(path) WHERE deleted_at IS NULL;  -- For hierarchy queries
CREATE INDEX idx_warehouse_locations_status ON warehouse_locations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_warehouse_locations_available ON warehouse_locations(status, is_putawayable, current_occupancy, capacity) 
    WHERE deleted_at IS NULL AND status = 'active';  -- For availability queries

-- Comments
COMMENT ON TABLE warehouse_locations IS 'Hierarchical warehouse location structure (warehouse → zone → aisle → rack → shelf → bin)';
COMMENT ON COLUMN warehouse_locations.code IS 'Unique location code (e.g., WH01, A, 01)';
COMMENT ON COLUMN warehouse_locations.path IS 'Materialized path for fast hierarchy queries (e.g., /WH01/A/01)';
COMMENT ON COLUMN warehouse_locations.level IS 'Hierarchy level: 0=warehouse, 1=zone, 2=aisle, 3=rack, 4=shelf, 5=bin';
COMMENT ON COLUMN warehouse_locations.is_limited IS 'Whether location has capacity limits (true for bins with max units)';
COMMENT ON COLUMN warehouse_locations.is_pickable IS 'Whether items can be picked from this location';
COMMENT ON COLUMN warehouse_locations.is_putawayable IS 'Whether items can be put away to this location';
