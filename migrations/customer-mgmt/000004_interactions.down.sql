-- ============================================================================
-- Migration: 000004_interactions (DOWN)
-- Description: Drop customer_interactions table and related types
-- Context: Customer Management
-- ============================================================================

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_interactions_updated_at ON customer_interactions;

-- Drop table
DROP TABLE IF EXISTS customer_interactions CASCADE;

-- Drop enums
DROP TYPE IF EXISTS interaction_outcome;
DROP TYPE IF EXISTS interaction_direction;
DROP TYPE IF EXISTS interaction_type;
