-- ============================================================================
-- Migration: 000004_interactions
-- Description: Create customer_interactions table for logging customer communications
-- Context: Customer Management
-- Created: 2025-12-30
-- ============================================================================

-- ============================================================================
-- Interaction Type Enum
-- ============================================================================
-- Represents types of customer interactions
CREATE TYPE interaction_type AS ENUM (
    'call',      -- Phone call
    'email',     -- Email communication
    'meeting',   -- In-person or virtual meeting
    'note',      -- General note/memo
    'sms',       -- SMS message
    'chat'       -- Chat/messaging
);

COMMENT ON TYPE interaction_type IS 'Types of customer interaction activities';

-- ============================================================================
-- Interaction Direction Enum
-- ============================================================================
-- Represents direction of communication
CREATE TYPE interaction_direction AS ENUM (
    'inbound',   -- From customer to us
    'outbound'   -- From us to customer
);

COMMENT ON TYPE interaction_direction IS 'Direction of customer communication';

-- ============================================================================
-- Interaction Outcome Enum
-- ============================================================================
-- Represents outcome/result of interaction
CREATE TYPE interaction_outcome AS ENUM (
    'successful',      -- Successful interaction
    'no_answer',       -- No answer (calls)
    'voicemail',       -- Left voicemail
    'busy',            -- Line busy
    'scheduled',       -- Meeting scheduled
    'not_interested'   -- Customer not interested
);

COMMENT ON TYPE interaction_outcome IS 'Outcome or result of customer interaction';

-- ============================================================================
-- Customer Interactions Table
-- ============================================================================
-- Stores all customer interactions (calls, emails, meetings, notes)
CREATE TABLE customer_interactions (
    -- Primary Key
    id TEXT PRIMARY KEY,

    -- Relationships
    customer_id TEXT NOT NULL,                  -- Link to customer (required)
    company_id TEXT,                            -- Link to company (optional, for B2B)

    -- Interaction Details
    type interaction_type NOT NULL,             -- Type of interaction
    direction interaction_direction NOT NULL,   -- Direction of communication
    outcome interaction_outcome,                -- Result/outcome

    -- Content
    subject VARCHAR(255) NOT NULL,              -- Subject/title
    description TEXT NOT NULL,                  -- Detailed notes

    -- Participants
    created_by TEXT NOT NULL,                   -- User who logged interaction
    attendees TEXT,                             -- JSON array of attendee UUIDs stored as TEXT for cross-DB compatibility

    -- Timing
    started_at TIMESTAMPTZ NOT NULL,            -- When interaction started
    ended_at TIMESTAMPTZ,                       -- When interaction ended
    duration_sec INTEGER,                       -- Duration in seconds

    -- Follow-up
    follow_up_required BOOLEAN NOT NULL DEFAULT false, -- Needs follow-up?
    follow_up_date TIMESTAMPTZ,                 -- When to follow up
    follow_up_notes TEXT,                       -- Follow-up instructions

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,                     -- Soft delete

    -- Constraints
    CONSTRAINT interactions_subject_not_empty CHECK (subject <> ''),
    CONSTRAINT interactions_description_not_empty CHECK (description <> ''),
    CONSTRAINT interactions_duration_positive CHECK (duration_sec IS NULL OR duration_sec >= 0),
    CONSTRAINT interactions_ended_after_started CHECK (ended_at IS NULL OR ended_at >= started_at),
    CONSTRAINT interactions_follow_up_date_valid CHECK (
        follow_up_required = false OR follow_up_date IS NOT NULL
    ),

    -- Foreign Keys
    CONSTRAINT fk_interactions_customer
        FOREIGN KEY (customer_id)
        REFERENCES customer_customers(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_interactions_company
        FOREIGN KEY (company_id)
        REFERENCES customer_companies(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_interactions_created_by
        FOREIGN KEY (created_by)
        REFERENCES identity_users(id)
        ON DELETE RESTRICT
);

-- ============================================================================
-- Indexes
-- ============================================================================

-- Customer lookups (most common query)
CREATE INDEX idx_interactions_customer_id 
    ON customer_interactions(customer_id) 
    WHERE deleted_at IS NULL;

-- Company lookups (B2B interactions)
CREATE INDEX idx_interactions_company_id 
    ON customer_interactions(company_id) 
    WHERE deleted_at IS NULL;

-- Type filtering (e.g., show all calls)
CREATE INDEX idx_interactions_type 
    ON customer_interactions(type) 
    WHERE deleted_at IS NULL;

-- User activity tracking
CREATE INDEX idx_interactions_created_by 
    ON customer_interactions(created_by) 
    WHERE deleted_at IS NULL;

-- Follow-up management (critical for CRM workflow)
CREATE INDEX idx_interactions_follow_up 
    ON customer_interactions(follow_up_date) 
    WHERE follow_up_required = true AND deleted_at IS NULL;

-- Timeline queries (sorted by time)
CREATE INDEX idx_interactions_started_at 
    ON customer_interactions(started_at DESC) 
    WHERE deleted_at IS NULL;

-- Composite index for customer timeline
CREATE INDEX idx_interactions_customer_started 
    ON customer_interactions(customer_id, started_at DESC) 
    WHERE deleted_at IS NULL;

-- ============================================================================
-- Comments
-- ============================================================================

COMMENT ON TABLE customer_interactions IS 'Stores all customer interactions (calls, emails, meetings, notes)';

COMMENT ON COLUMN customer_interactions.id IS 'UUID v7 stored as TEXT primary key';
COMMENT ON COLUMN customer_interactions.customer_id IS 'Link to customer (required)';
COMMENT ON COLUMN customer_interactions.company_id IS 'Link to company (optional, for B2B)';
COMMENT ON COLUMN customer_interactions.type IS 'Type of interaction (call, email, meeting, note, sms, chat)';
COMMENT ON COLUMN customer_interactions.direction IS 'Direction of communication (inbound/outbound)';
COMMENT ON COLUMN customer_interactions.outcome IS 'Result of interaction (successful, no_answer, etc.)';
COMMENT ON COLUMN customer_interactions.subject IS 'Subject or title of interaction';
COMMENT ON COLUMN customer_interactions.description IS 'Detailed notes about interaction';
COMMENT ON COLUMN customer_interactions.created_by IS 'User who logged this interaction';
COMMENT ON COLUMN customer_interactions.attendees IS 'JSON array of attendee UUIDs stored as TEXT (for meetings)';
COMMENT ON COLUMN customer_interactions.started_at IS 'When interaction started';
COMMENT ON COLUMN customer_interactions.ended_at IS 'When interaction ended';
COMMENT ON COLUMN customer_interactions.duration_sec IS 'Duration in seconds (calculated)';
COMMENT ON COLUMN customer_interactions.follow_up_required IS 'Whether follow-up is needed';
COMMENT ON COLUMN customer_interactions.follow_up_date IS 'When to follow up';
COMMENT ON COLUMN customer_interactions.follow_up_notes IS 'Follow-up instructions';
COMMENT ON COLUMN customer_interactions.created_at IS 'When record was created';
COMMENT ON COLUMN customer_interactions.updated_at IS 'When record was last updated';
COMMENT ON COLUMN customer_interactions.deleted_at IS 'Soft delete timestamp';

-- ============================================================================
-- Statistics
-- ============================================================================

ANALYZE customer_interactions;
