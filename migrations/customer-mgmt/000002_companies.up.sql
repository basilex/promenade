-- ============================================================================
-- Migration: 000002_companies
-- Description: Create customer_companies table for B2B customer management
-- Context: Customer Management
-- Created: 2025-12-30
-- ============================================================================

-- ============================================================================
-- Company Type Enum
-- ============================================================================
-- Represents legal structure of business organizations
CREATE TYPE company_type AS ENUM (
    'llc',              -- Limited Liability Company
    'corporation',      -- Corporation (C-Corp, S-Corp)
    'sole_proprietor',  -- Sole Proprietorship
    'partnership',      -- Partnership (GP, LP, LLP)
    'non_profit',       -- Non-profit organization
    'other'             -- Other legal structure
);

COMMENT ON TYPE company_type IS 'Legal structure types for business organizations';

-- ============================================================================
-- Company Size Enum
-- ============================================================================
-- Categorizes companies by employee count
CREATE TYPE company_size AS ENUM (
    'micro',      -- 1-10 employees
    'small',      -- 11-50 employees
    'medium',     -- 51-250 employees
    'large',      -- 251-1000 employees
    'enterprise'  -- 1000+ employees
);

COMMENT ON TYPE company_size IS 'Company size categories based on employee count';

-- ============================================================================
-- Customer Companies Table
-- ============================================================================
-- Stores B2B customer information (business organizations)
CREATE TABLE customer_companies (
    -- Primary Key
    id TEXT PRIMARY KEY,

    -- Basic Information
    name VARCHAR(255) NOT NULL,                -- Company name (must be unique)
    legal_name VARCHAR(255),                   -- Legal entity name (optional)
    type company_type NOT NULL,                -- Legal structure
    tax_id VARCHAR(100),                       -- Tax identification number
    registration_number VARCHAR(100),          -- Business registration number

    -- Contact Information
    website VARCHAR(500),                      -- Company website URL
    email VARCHAR(255),                        -- Primary email address
    phone VARCHAR(50),                         -- Phone number
    phone_country_code VARCHAR(10),            -- Country code for phone

    -- Address (structured)
    address_line1 VARCHAR(255),                -- Street address line 1
    address_line2 VARCHAR(255),                -- Street address line 2 (optional)
    city VARCHAR(100),                         -- City
    state_province VARCHAR(100),               -- State/Province
    postal_code VARCHAR(20),                   -- Postal/ZIP code
    country VARCHAR(2),                        -- ISO 3166-1 alpha-2 country code

    -- Business Information
    industry VARCHAR(100),                     -- Industry sector
    size company_size NOT NULL,                -- Company size category
    employee_count INTEGER NOT NULL DEFAULT 0, -- Number of employees
    revenue BIGINT NOT NULL DEFAULT 0,         -- Annual revenue (in cents)
    currency VARCHAR(3) NOT NULL DEFAULT 'USD', -- ISO 4217 currency code

    -- Additional Information
    description TEXT,                          -- Company description/notes

    -- Relationships
    parent_company_id TEXT,                    -- Parent company (for subsidiaries)

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,                    -- Soft delete

    -- Constraints
    CONSTRAINT companies_name_not_empty CHECK (name <> ''),
    CONSTRAINT companies_employee_count_positive CHECK (employee_count >= 0),
    CONSTRAINT companies_revenue_positive CHECK (revenue >= 0),
    CONSTRAINT companies_currency_valid CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT companies_country_valid CHECK (country IS NULL OR country ~ '^[A-Z]{2}$'),
    CONSTRAINT companies_parent_not_self CHECK (parent_company_id IS NULL OR parent_company_id != id),

    -- Foreign Keys
    CONSTRAINT fk_companies_parent
        FOREIGN KEY (parent_company_id)
        REFERENCES customer_companies(id)
        ON DELETE SET NULL
);

-- ============================================================================
-- Table Comments
-- ============================================================================
COMMENT ON TABLE customer_companies IS 'Stores B2B customer information (business organizations)';
COMMENT ON COLUMN customer_companies.id IS 'UUID v7 stored as TEXT (time-ordered)';
COMMENT ON COLUMN customer_companies.name IS 'Company name (unique across active companies)';
COMMENT ON COLUMN customer_companies.legal_name IS 'Legal entity name (may differ from trade name)';
COMMENT ON COLUMN customer_companies.type IS 'Legal structure of organization';
COMMENT ON COLUMN customer_companies.tax_id IS 'Tax identification number (EIN, VAT, etc.)';
COMMENT ON COLUMN customer_companies.registration_number IS 'Business registration number';
COMMENT ON COLUMN customer_companies.employee_count IS 'Number of employees';
COMMENT ON COLUMN customer_companies.revenue IS 'Annual revenue in cents (for precision)';
COMMENT ON COLUMN customer_companies.currency IS 'ISO 4217 currency code for revenue';
COMMENT ON COLUMN customer_companies.parent_company_id IS 'Parent company for subsidiaries/divisions';
COMMENT ON COLUMN customer_companies.deleted_at IS 'Soft delete timestamp (NULL = active)';

-- ============================================================================
-- Indexes
-- ============================================================================

-- Primary search indexes
CREATE UNIQUE INDEX idx_companies_name_unique
    ON customer_companies(name)
    WHERE deleted_at IS NULL;
COMMENT ON INDEX idx_companies_name_unique IS 'Ensure name uniqueness for active companies only';

CREATE INDEX idx_companies_tax_id
    ON customer_companies(tax_id)
    WHERE deleted_at IS NULL AND tax_id IS NOT NULL;
COMMENT ON INDEX idx_companies_tax_id IS 'Fast lookup by tax ID';

-- Business filtering indexes
CREATE INDEX idx_companies_industry
    ON customer_companies(industry)
    WHERE deleted_at IS NULL AND industry IS NOT NULL;
COMMENT ON INDEX idx_companies_industry IS 'Filter companies by industry sector';

CREATE INDEX idx_companies_size
    ON customer_companies(size)
    WHERE deleted_at IS NULL;
COMMENT ON INDEX idx_companies_size IS 'Filter companies by size category';

-- Relationship indexes
CREATE INDEX idx_companies_parent_company_id
    ON customer_companies(parent_company_id)
    WHERE deleted_at IS NULL AND parent_company_id IS NOT NULL;
COMMENT ON INDEX idx_companies_parent_company_id IS 'Find subsidiaries of parent company';

-- Soft delete index
CREATE INDEX idx_companies_deleted_at
    ON customer_companies(deleted_at)
    WHERE deleted_at IS NOT NULL;
COMMENT ON INDEX idx_companies_deleted_at IS 'Query deleted companies efficiently';

-- Timestamp indexes
CREATE INDEX idx_companies_created_at
    ON customer_companies(created_at);
COMMENT ON INDEX idx_companies_created_at IS 'Sort by creation date';

CREATE INDEX idx_companies_updated_at
    ON customer_companies(updated_at);
COMMENT ON INDEX idx_companies_updated_at IS 'Sort by last update date';

-- Composite index for common queries
CREATE INDEX idx_companies_type_size_active
    ON customer_companies(type, size, created_at DESC)
    WHERE deleted_at IS NULL;
COMMENT ON INDEX idx_companies_type_size_active IS 'Filter by type and size with recent first';

-- ============================================================================
-- Triggers
-- ============================================================================

-- Auto-update updated_at timestamp
-- ============================================================================
-- Initial Data (Optional - Reference Data)
-- ============================================================================
-- No initial data required for companies (user-generated content)
