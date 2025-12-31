-- ============================================================================
-- Migration: 000003_deals (Rollback)
-- Description: Drop customer_deals table and related objects
-- Context: Customer Management
-- Created: 2025-12-30
-- ============================================================================

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_deals_updated_at ON customer_deals;
DROP FUNCTION IF EXISTS update_deals_updated_at();

-- Drop table (CASCADE will drop dependent objects)
DROP TABLE IF EXISTS customer_deals CASCADE;

-- Drop enums
DROP TYPE IF EXISTS deal_source;
DROP TYPE IF EXISTS deal_stage;
