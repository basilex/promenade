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
-- - UUID v7 primary keys for consistency
-- ============================================================================

-- Countries (ISO 3166-1)
CREATE TABLE IF NOT EXISTS shared_countries (
    id UUID PRIMARY KEY,
    code VARCHAR(2) NOT NULL UNIQUE,           -- ISO 3166-1 alpha-2 (US, UA, DE)
    code3 VARCHAR(3) NOT NULL UNIQUE,          -- ISO 3166-1 alpha-3 (USA, UKR, DEU)
    numeric_code VARCHAR(3) NOT NULL UNIQUE,   -- ISO 3166-1 numeric (840, 804, 276)
    name VARCHAR(100) NOT NULL,                -- English name
    name_local VARCHAR(100),                   -- Local name (optional)
    phone_code VARCHAR(10) NOT NULL,           -- Phone country code (+1, +380, +49)
    capital VARCHAR(100),                      -- Capital city
    region VARCHAR(50),                        -- Geographic region (Europe, Asia, Americas, etc.)
    subregion VARCHAR(50),                     -- Geographic subregion (Western Europe, Eastern Europe, etc.)
    flag_emoji VARCHAR(10),                    -- Country flag emoji (🇺🇸, 🇺🇦, etc.)
    latitude DECIMAL(10, 8),                   -- Country center latitude
    longitude DECIMAL(11, 8),                  -- Country center longitude
    area_km2 INTEGER,                          -- Country area in square kilometers
    population BIGINT,                         -- Country population (approximate)
    translations JSONB DEFAULT '{}'::jsonb,    -- Name translations (map: lang_code -> name)
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shared_countries_code ON shared_countries(code) WHERE is_active = true;
CREATE INDEX idx_shared_countries_name ON shared_countries(name) WHERE is_active = true;
CREATE INDEX idx_shared_countries_region ON shared_countries(region) WHERE is_active = true;
CREATE INDEX idx_shared_countries_subregion ON shared_countries(subregion) WHERE is_active = true;
CREATE INDEX idx_shared_countries_translations ON shared_countries USING GIN(translations);

COMMENT ON TABLE shared_countries IS 'ISO 3166-1 countries (Shared Kernel)';
COMMENT ON COLUMN shared_countries.code IS 'ISO 3166-1 alpha-2 code (US, UA, DE)';
COMMENT ON COLUMN shared_countries.phone_code IS 'International dialing code (+1, +380, +49)';
COMMENT ON COLUMN shared_countries.capital IS 'Capital city name';
COMMENT ON COLUMN shared_countries.region IS 'Geographic region (e.g., Europe, Asia)';
COMMENT ON COLUMN shared_countries.subregion IS 'Geographic subregion (e.g., Western Europe)';
COMMENT ON COLUMN shared_countries.flag_emoji IS 'Country flag emoji (Unicode)';
COMMENT ON COLUMN shared_countries.translations IS 'Country name translations (JSONB map: language_code -> name)';

-- Currencies (ISO 4217 + Cryptocurrencies)
CREATE TABLE IF NOT EXISTS shared_currencies (
    id UUID PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,          -- ISO 4217 alpha code (USD, EUR, UAH) or crypto symbol (BTC, USDT)
    numeric_code VARCHAR(3) NOT NULL UNIQUE,   -- ISO 4217 numeric (840, 978, 980) or 900-999 for crypto
    name VARCHAR(100) NOT NULL,                -- English name
    symbol VARCHAR(10) NOT NULL,               -- Currency symbol ($, €, ₴)
    decimal_places SMALLINT NOT NULL DEFAULT 2, -- Decimal places (2 for USD, 0 for JPY)
    rounding INTEGER DEFAULT 0,                -- Rounding mode (0=normal, 1=up, 2=down)
    is_crypto BOOLEAN DEFAULT false,           -- Is cryptocurrency
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shared_currencies_code ON shared_currencies(code) WHERE is_active = true;
CREATE INDEX idx_shared_currencies_is_crypto ON shared_currencies(is_crypto) WHERE is_active = true;

COMMENT ON TABLE shared_currencies IS 'ISO 4217 currencies (Shared Kernel)';
COMMENT ON COLUMN shared_currencies.decimal_places IS 'Number of decimal places (2 for USD, 0 for JPY, etc.)';
COMMENT ON COLUMN shared_currencies.rounding IS 'Rounding mode (0=normal, 1=up, 2=down)';
COMMENT ON COLUMN shared_currencies.is_crypto IS 'Is cryptocurrency (true/false)';

-- Languages (ISO 639-1)
CREATE TABLE IF NOT EXISTS shared_languages (
    id UUID PRIMARY KEY,
    code VARCHAR(2) NOT NULL UNIQUE,           -- ISO 639-1 alpha-2 (en, uk, de)
    code3 VARCHAR(3) NOT NULL UNIQUE,          -- ISO 639-2/T alpha-3 (eng, ukr, deu)
    name VARCHAR(100) NOT NULL,                -- English name
    native_name VARCHAR(100) NOT NULL,         -- Native name (English, Ukrainian, Deutsch)
    direction VARCHAR(3) DEFAULT 'ltr',        -- Text direction (ltr, rtl)
    native_speakers BIGINT,                    -- Number of native speakers (approximate)
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_language_direction CHECK (direction IN ('ltr', 'rtl'))
);

CREATE INDEX idx_shared_languages_code ON shared_languages(code) WHERE is_active = true;
CREATE INDEX idx_shared_languages_direction ON shared_languages(direction) WHERE is_active = true;

COMMENT ON TABLE shared_languages IS 'ISO 639-1 languages (Shared Kernel)';
COMMENT ON COLUMN shared_languages.direction IS 'Text direction (ltr=left-to-right, rtl=right-to-left)';
COMMENT ON COLUMN shared_languages.native_speakers IS 'Number of native speakers (approximate)';

-- Timezones (IANA)
CREATE TABLE IF NOT EXISTS shared_timezones (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,         -- IANA timezone (America/New_York, Europe/Kyiv)
    abbreviation VARCHAR(10) NOT NULL,         -- Common abbreviation (EST, EET, CET)
    utc_offset INTEGER NOT NULL,               -- UTC offset in seconds
    country_code VARCHAR(2),                   -- Primary country code (ISO 3166-1 alpha-2)
    dst_offset INTEGER,                        -- DST offset in seconds (if applicable)
    display_name VARCHAR(100),                 -- Human-readable timezone name
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shared_timezones_name ON shared_timezones(name) WHERE is_active = true;
CREATE INDEX idx_shared_timezones_country_code ON shared_timezones(country_code) WHERE is_active = true;

COMMENT ON TABLE shared_timezones IS 'IANA timezones (Shared Kernel)';
COMMENT ON COLUMN shared_timezones.utc_offset IS 'UTC offset in seconds';
COMMENT ON COLUMN shared_timezones.country_code IS 'Primary country code (ISO 3166-1 alpha-2)';
COMMENT ON COLUMN shared_timezones.dst_offset IS 'DST offset in seconds (if applicable)';
COMMENT ON COLUMN shared_timezones.display_name IS 'Human-readable timezone name';

-- ============================================================================
-- Junction Tables (M2M relationships)
-- ============================================================================

-- Country-Currency M2M (one country can use multiple currencies, one currency used in multiple countries)
CREATE TABLE IF NOT EXISTS shared_country_currencies (
    id UUID PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES shared_countries(id) ON DELETE CASCADE,
    currency_id UUID NOT NULL REFERENCES shared_currencies(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT false,     -- Is this the primary/official currency
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(country_id, currency_id)
);

CREATE INDEX idx_country_currencies_country ON shared_country_currencies(country_id);
CREATE INDEX idx_country_currencies_currency ON shared_country_currencies(currency_id);
CREATE INDEX idx_country_currencies_primary ON shared_country_currencies(country_id, is_primary) WHERE is_primary = true;

COMMENT ON TABLE shared_country_currencies IS 'M2M: Countries and their currencies';
COMMENT ON COLUMN shared_country_currencies.is_primary IS 'Is this the primary/official currency for the country';

-- Country-Language M2M (one country can have multiple official languages, one language used in multiple countries)
CREATE TABLE IF NOT EXISTS shared_country_languages (
    id UUID PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES shared_countries(id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES shared_languages(id) ON DELETE CASCADE,
    is_official BOOLEAN NOT NULL DEFAULT false,    -- Is this an official language
    is_primary BOOLEAN NOT NULL DEFAULT false,     -- Is this the primary language
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(country_id, language_id)
);

CREATE INDEX idx_country_languages_country ON shared_country_languages(country_id);
CREATE INDEX idx_country_languages_language ON shared_country_languages(language_id);
CREATE INDEX idx_country_languages_official ON shared_country_languages(country_id, is_official) WHERE is_official = true;
CREATE INDEX idx_country_languages_primary ON shared_country_languages(country_id, is_primary) WHERE is_primary = true;

COMMENT ON TABLE shared_country_languages IS 'M2M: Countries and their official languages';
COMMENT ON COLUMN shared_country_languages.is_official IS 'Is this an official language of the country';
COMMENT ON COLUMN shared_country_languages.is_primary IS 'Is this the primary language of the country';

-- Country-Timezone M2M (one country can have multiple timezones, one timezone can span multiple countries)
CREATE TABLE IF NOT EXISTS shared_country_timezones (
    id UUID PRIMARY KEY,
    country_id UUID NOT NULL REFERENCES shared_countries(id) ON DELETE CASCADE,
    timezone_id UUID NOT NULL REFERENCES shared_timezones(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT false,     -- Is this the primary/capital timezone
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(country_id, timezone_id)
);

CREATE INDEX idx_country_timezones_country ON shared_country_timezones(country_id);
CREATE INDEX idx_country_timezones_timezone ON shared_country_timezones(timezone_id);
CREATE INDEX idx_country_timezones_primary ON shared_country_timezones(country_id, is_primary) WHERE is_primary = true;

COMMENT ON TABLE shared_country_timezones IS 'M2M: Countries and their IANA timezones';
COMMENT ON COLUMN shared_country_timezones.is_primary IS 'Is this the primary/capital timezone of the country';

-- ============================================================================
-- Seed data: Most commonly used reference data
-- ============================================================================
-- Note: Seed data removed for DB-agnostic approach. 
-- Use application layer to populate reference data with proper UUID v7 generation.
