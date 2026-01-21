-- ============================================================================
-- Shared Kernel: Reference Data (Countries, Currencies, Languages, Timezones)
-- ============================================================================
-- These tables are part of the Shared Kernel - accessible by all Bounded Contexts.
-- Reference data is stable, standardized, and rarely changes.
--
-- Design decisions:
-- - Read-only from application perspective (managed via migrations)
-- - No soft delete (data doesn't "disappear")
-- - is_active flag for deactivation without breaking foreign keys
-- - UUID v7 primary keys stored as NVARCHAR for cross-database compatibility
-- ============================================================================

-- Countries (ISO 3166-1)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'shared_countries' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    CREATE TABLE shared_countries (
        id NVARCHAR(36) PRIMARY KEY,
        code VARCHAR(2) NOT NULL UNIQUE,           -- ISO 3166-1 alpha-2 (US, UA, DE)
        code3 VARCHAR(3) NOT NULL UNIQUE,          -- ISO 3166-1 alpha-3 (USA, UKR, DEU)
        numeric_code VARCHAR(3) NOT NULL UNIQUE,   -- ISO 3166-1 numeric (840, 804, 276)
        name NVARCHAR(100) NOT NULL,               -- English name
        name_local NVARCHAR(100),                  -- Local name (optional)
        phone_code VARCHAR(10) NOT NULL,           -- Phone country code (+1, +380, +49)
        capital NVARCHAR(100),                     -- Capital city
        region NVARCHAR(50),                       -- Geographic region (Europe, Asia, Americas, etc.)
        subregion NVARCHAR(50),                    -- Geographic subregion (Western Europe, Eastern Europe, etc.)
        flag_emoji NVARCHAR(10),                   -- Country flag emoji (🇺🇸, 🇺🇦, etc.)
        latitude DECIMAL(10, 8),                   -- Country center latitude
        longitude DECIMAL(11, 8),                  -- Country center longitude
        area_km2 INT,                              -- Country area in square kilometers
        population BIGINT,                         -- Country population (approximate)
        translations NVARCHAR(MAX) DEFAULT '{}',   -- Name translations (JSON)
        is_active BIT NOT NULL DEFAULT 1,
        created_at DATETIME2 NOT NULL,
        updated_at DATETIME2 NOT NULL
    );
END;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_countries_code')
    CREATE INDEX idx_shared_countries_code ON shared_countries(code) WHERE is_active = 1;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_countries_name')
    CREATE INDEX idx_shared_countries_name ON shared_countries(name) WHERE is_active = 1;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_countries_region')
    CREATE INDEX idx_shared_countries_region ON shared_countries(region) WHERE is_active = 1;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_countries_subregion')
    CREATE INDEX idx_shared_countries_subregion ON shared_countries(subregion) WHERE is_active = 1;
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'ISO 3166-1 countries (Shared Kernel)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_countries';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'ISO 3166-1 alpha-2 code (US, UA, DE)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_countries',
    @level2type = N'COLUMN', @level2name = 'code';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'International dialing code (+1, +380, +49)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_countries',
    @level2type = N'COLUMN', @level2name = 'phone_code';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Capital city name',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_countries',
    @level2type = N'COLUMN', @level2name = 'capital';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Geographic region (e.g., Europe, Asia)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_countries',
    @level2type = N'COLUMN', @level2name = 'region';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Geographic subregion (e.g., Western Europe)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_countries',
    @level2type = N'COLUMN', @level2name = 'subregion';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Country flag emoji (Unicode)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_countries',
    @level2type = N'COLUMN', @level2name = 'flag_emoji';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Country name translations (JSON map: language_code -> name)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_countries',
    @level2type = N'COLUMN', @level2name = 'translations';
GO

-- Currencies (ISO 4217 + Cryptocurrencies)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'shared_currencies' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    CREATE TABLE shared_currencies (
        id NVARCHAR(36) PRIMARY KEY,
        code VARCHAR(10) NOT NULL UNIQUE,          -- ISO 4217 alpha code (USD, EUR, UAH) or crypto symbol (BTC, USDT)
        numeric_code VARCHAR(3) NOT NULL UNIQUE,   -- ISO 4217 numeric (840, 978, 980) or 900-999 for crypto
        name NVARCHAR(100) NOT NULL,               -- English name
        symbol NVARCHAR(10) NOT NULL,              -- Currency symbol ($, €, ₴)
        decimal_places SMALLINT NOT NULL DEFAULT 2, -- Decimal places (2 for USD, 0 for JPY)
        rounding INT DEFAULT 0,                    -- Rounding mode (0=normal, 1=up, 2=down)
        is_crypto BIT DEFAULT 0,                   -- Is cryptocurrency
        is_active BIT NOT NULL DEFAULT 1,
        created_at DATETIME2 NOT NULL,
        updated_at DATETIME2 NOT NULL
    );
END;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_currencies_code')
    CREATE INDEX idx_shared_currencies_code ON shared_currencies(code) WHERE is_active = 1;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_currencies_is_crypto')
    CREATE INDEX idx_shared_currencies_is_crypto ON shared_currencies(is_crypto) WHERE is_active = 1;
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'ISO 4217 currencies (Shared Kernel)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_currencies';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Number of decimal places (2 for USD, 0 for JPY, etc.)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_currencies',
    @level2type = N'COLUMN', @level2name = 'decimal_places';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Rounding mode (0=normal, 1=up, 2=down)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_currencies',
    @level2type = N'COLUMN', @level2name = 'rounding';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Is cryptocurrency (true/false)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_currencies',
    @level2type = N'COLUMN', @level2name = 'is_crypto';
GO

-- Languages (ISO 639-1)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'shared_languages' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    CREATE TABLE shared_languages (
        id NVARCHAR(36) PRIMARY KEY,
        code VARCHAR(2) NOT NULL UNIQUE,           -- ISO 639-1 alpha-2 (en, uk, de)
        code3 VARCHAR(3) NOT NULL UNIQUE,          -- ISO 639-2/T alpha-3 (eng, ukr, deu)
        name NVARCHAR(100) NOT NULL,               -- English name
        native_name NVARCHAR(100) NOT NULL,        -- Native name (English, Ukrainian, Deutsch)
        direction VARCHAR(3) DEFAULT 'ltr',        -- Text direction (ltr, rtl)
        native_speakers BIGINT,                    -- Number of native speakers (approximate)
        is_active BIT NOT NULL DEFAULT 1,
        created_at DATETIME2 NOT NULL,
        updated_at DATETIME2 NOT NULL,
        CONSTRAINT chk_language_direction CHECK (direction IN ('ltr', 'rtl'))
    );
END;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_languages_code')
    CREATE INDEX idx_shared_languages_code ON shared_languages(code) WHERE is_active = 1;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_languages_direction')
    CREATE INDEX idx_shared_languages_direction ON shared_languages(direction) WHERE is_active = 1;
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'ISO 639-1 languages (Shared Kernel)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_languages';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Text direction (ltr=left-to-right, rtl=right-to-left)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_languages',
    @level2type = N'COLUMN', @level2name = 'direction';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Number of native speakers (approximate)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_languages',
    @level2type = N'COLUMN', @level2name = 'native_speakers';
GO

-- Timezones (IANA)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'shared_timezones' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    CREATE TABLE shared_timezones (
        id NVARCHAR(36) PRIMARY KEY,
        name NVARCHAR(100) NOT NULL UNIQUE,        -- IANA timezone (America/New_York, Europe/Kyiv)
        abbreviation VARCHAR(10) NOT NULL,         -- Common abbreviation (EST, EET, CET)
        utc_offset INT NOT NULL,                   -- UTC offset in seconds
        country_code VARCHAR(2),                   -- Primary country code (ISO 3166-1 alpha-2)
        dst_offset INT,                            -- DST offset in seconds (if applicable)
        display_name NVARCHAR(100),                -- Human-readable timezone name
        is_active BIT NOT NULL DEFAULT 1,
        created_at DATETIME2 NOT NULL,
        updated_at DATETIME2 NOT NULL
    );
END;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_timezones_name')
    CREATE INDEX idx_shared_timezones_name ON shared_timezones(name) WHERE is_active = 1;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_shared_timezones_country_code')
    CREATE INDEX idx_shared_timezones_country_code ON shared_timezones(country_code) WHERE is_active = 1;
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'IANA timezones (Shared Kernel)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_timezones';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'UTC offset in seconds',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_timezones',
    @level2type = N'COLUMN', @level2name = 'utc_offset';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Primary country code (ISO 3166-1 alpha-2)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_timezones',
    @level2type = N'COLUMN', @level2name = 'country_code';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'DST offset in seconds (if applicable)',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_timezones',
    @level2type = N'COLUMN', @level2name = 'dst_offset';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Human-readable timezone name',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_timezones',
    @level2type = N'COLUMN', @level2name = 'display_name';
GO

-- ============================================================================
-- Junction Tables (M2M relationships)
-- ============================================================================

-- Country-Currency M2M (one country can use multiple currencies, one currency used in multiple countries)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'shared_country_currencies' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    CREATE TABLE shared_country_currencies (
        id NVARCHAR(36) PRIMARY KEY,
        country_id NVARCHAR(36) NOT NULL,
        currency_id NVARCHAR(36) NOT NULL,
        is_primary BIT NOT NULL DEFAULT 0,         -- Is this the primary/official currency
        created_at DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
        CONSTRAINT fk_country_currencies_country FOREIGN KEY (country_id) REFERENCES shared_countries(id) ON DELETE CASCADE,
        CONSTRAINT fk_country_currencies_currency FOREIGN KEY (currency_id) REFERENCES shared_currencies(id) ON DELETE CASCADE,
        CONSTRAINT uq_country_currency UNIQUE(country_id, currency_id)
    );
END;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_currencies_country')
    CREATE INDEX idx_country_currencies_country ON shared_country_currencies(country_id);
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_currencies_currency')
    CREATE INDEX idx_country_currencies_currency ON shared_country_currencies(currency_id);
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_currencies_primary')
    CREATE INDEX idx_country_currencies_primary ON shared_country_currencies(country_id, is_primary) WHERE is_primary = 1;
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'M2M: Countries and their currencies',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_country_currencies';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Is this the primary/official currency for the country',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_country_currencies',
    @level2type = N'COLUMN', @level2name = 'is_primary';
GO

-- Country-Language M2M (one country can have multiple official languages, one language used in multiple countries)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'shared_country_languages' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    CREATE TABLE shared_country_languages (
        id NVARCHAR(36) PRIMARY KEY,
        country_id NVARCHAR(36) NOT NULL,
        language_id NVARCHAR(36) NOT NULL,
        is_official BIT NOT NULL DEFAULT 0,        -- Is this an official language
        is_primary BIT NOT NULL DEFAULT 0,         -- Is this the primary language
        created_at DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
        CONSTRAINT fk_country_languages_country FOREIGN KEY (country_id) REFERENCES shared_countries(id) ON DELETE CASCADE,
        CONSTRAINT fk_country_languages_language FOREIGN KEY (language_id) REFERENCES shared_languages(id) ON DELETE CASCADE,
        CONSTRAINT uq_country_language UNIQUE(country_id, language_id)
    );
END;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_languages_country')
    CREATE INDEX idx_country_languages_country ON shared_country_languages(country_id);
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_languages_language')
    CREATE INDEX idx_country_languages_language ON shared_country_languages(language_id);
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_languages_official')
    CREATE INDEX idx_country_languages_official ON shared_country_languages(country_id, is_official) WHERE is_official = 1;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_languages_primary')
    CREATE INDEX idx_country_languages_primary ON shared_country_languages(country_id, is_primary) WHERE is_primary = 1;
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'M2M: Countries and their official languages',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_country_languages';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Is this an official language of the country',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_country_languages',
    @level2type = N'COLUMN', @level2name = 'is_official';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Is this the primary language of the country',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_country_languages',
    @level2type = N'COLUMN', @level2name = 'is_primary';
GO

-- Country-Timezone M2M (one country can have multiple timezones, one timezone can span multiple countries)
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'shared_country_timezones' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    CREATE TABLE shared_country_timezones (
        id NVARCHAR(36) PRIMARY KEY,
        country_id NVARCHAR(36) NOT NULL,
        timezone_id NVARCHAR(36) NOT NULL,
        is_primary BIT NOT NULL DEFAULT 0,         -- Is this the primary/capital timezone
        created_at DATETIME2 NOT NULL DEFAULT SYSDATETIME(),
        CONSTRAINT fk_country_timezones_country FOREIGN KEY (country_id) REFERENCES shared_countries(id) ON DELETE CASCADE,
        CONSTRAINT fk_country_timezones_timezone FOREIGN KEY (timezone_id) REFERENCES shared_timezones(id) ON DELETE CASCADE,
        CONSTRAINT uq_country_timezone UNIQUE(country_id, timezone_id)
    );
END;
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_timezones_country')
    CREATE INDEX idx_country_timezones_country ON shared_country_timezones(country_id);
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_timezones_timezone')
    CREATE INDEX idx_country_timezones_timezone ON shared_country_timezones(timezone_id);
GO

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_country_timezones_primary')
    CREATE INDEX idx_country_timezones_primary ON shared_country_timezones(country_id, is_primary) WHERE is_primary = 1;
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'M2M: Countries and their IANA timezones',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_country_timezones';
GO

EXEC sp_addextendedproperty 
    @name = N'MS_Description', @value = N'Is this the primary/capital timezone of the country',
    @level0type = N'SCHEMA', @level0name = 'dbo',
    @level1type = N'TABLE', @level1name = 'shared_country_timezones',
    @level2type = N'COLUMN', @level2name = 'is_primary';
GO

-- ============================================================================
-- Seed data: Most commonly used reference data
-- ============================================================================
-- Note: Seed data removed for DB-agnostic approach. 
-- Use application layer to populate reference data with proper UUID v7 generation.
