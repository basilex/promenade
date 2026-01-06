-- Migration: add_stock_movements_table
-- Context: warehouse
-- Created: 2026-01-06 09:12:25

-- Create stock movements table for immutable audit trail
CREATE TABLE warehouse_stock_movements (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign keys
    inventory_id UUID NOT NULL REFERENCES warehouse_inventory(id) ON DELETE RESTRICT,
    
    -- Movement type (receipt, reservation, reservation_release, commit, adjustment, transfer, damage, return)
    type VARCHAR(50) NOT NULL CHECK (type IN ('receipt', 'reservation', 'reservation_release', 'commit', 'adjustment', 'transfer', 'damage', 'return')),
    
    -- Quantity tracking
    quantity INTEGER NOT NULL,
    quantity_before_move INTEGER NOT NULL,
    quantity_after_move INTEGER NOT NULL,
    
    -- Warehouse routing (for transfers)
    from_warehouse_id UUID,
    from_location_code VARCHAR(50),
    to_warehouse_id UUID,
    to_location_code VARCHAR(50),
    
    -- Reference tracking (order, purchase_order, adjustment, etc.)
    reference_type VARCHAR(50),
    reference_id UUID,
    
    -- Cost tracking (for receipts)
    unit_cost_cents BIGINT,
    total_cost_cents BIGINT,
    currency_code VARCHAR(3),
    
    -- Audit fields
    reason TEXT,
    notes TEXT,
    created_by UUID NOT NULL,
    movement_date TIMESTAMP NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for common queries
CREATE INDEX idx_stock_movements_inventory_id ON warehouse_stock_movements(inventory_id);
CREATE INDEX idx_stock_movements_type ON warehouse_stock_movements(type);
CREATE INDEX idx_stock_movements_reference ON warehouse_stock_movements(reference_type, reference_id);
CREATE INDEX idx_stock_movements_movement_date ON warehouse_stock_movements(movement_date);
CREATE INDEX idx_stock_movements_created_at ON warehouse_stock_movements(created_at);

-- Add comment
COMMENT ON TABLE warehouse_stock_movements IS 'Immutable audit trail of all inventory movements';
