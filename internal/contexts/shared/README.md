# Shared Context (Bounded Context)

## Overview

**Shared Context** (also known as **Shared Kernel**) manages **reference data** used across all bounded contexts in Promenade Platform. This context provides **read-only, globally consistent data** such as countries, currencies, languages, timezones, and payment methods.

**Status**: Production-ready

---

## Bounded Context Responsibilities

### Core Domain

**Reference Data Management**:

- **Countries**: ISO 3166-1 alpha-2 codes, names, regions
- **Currencies**: ISO 4217 codes, symbols, decimal digits
- **Languages**: ISO 639-1 codes, native names, text direction (LTR/RTL)
- **Timezones**: IANA database identifiers, UTC offsets, DST support
- **Regions**: Geographic regions (Europe, Asia, Americas, etc.)
- **Cities**: Major world cities (optional, for address autocomplete)
- **Payment Methods**: Credit cards, bank transfers, e-wallets

### Characteristics

- **Read-only for applications**: Data changes via migrations only
- **Globally consistent**: All contexts use same reference data
- **Rarely updated**: Changes 1-2 times per year (ISO standard updates)
- **High performance**: Heavy caching (Redis/in-memory)
- **Multi-language**: Countries, languages, regions support i18n

---

## Architecture

### Directory Structure

```
shared/
 README.md                  # This file
 router.go                  # HTTP routes registration (all 4 aggregates)
 country/                   #  Country Aggregate
    entity.go              # Country entity (BaseAggregate)
    repository.go          # IRepository interface
    usecase.go             # IUseCase interface + implementation
    adapter/
       http/
          handler.go       # HTTP handlers (CRUD)
          dto.go           # Request/Response DTOs
       repository/postgres/
          base_repository.go    # Shared base repository
          country_repository.go # PostgreSQL implementation
 currency/                  #  Currency Aggregate
    entity.go              # Currency entity (BaseAggregate)
    repository.go          # IRepository interface
    usecase.go             # IUseCase interface + implementation
    adapter/
       http/
          handler.go       # HTTP handlers (CRUD)
          dto.go           # Request/Response DTOs
       repository/postgres/
          currency_repository.go
 language/                  #  Language Aggregate
    entity.go              # Language entity (BaseAggregate)
    repository.go          # IRepository interface
    usecase.go             # IUseCase interface + implementation
    adapter/
       http/
          handler.go       # HTTP handlers (CRUD)
          dto.go           # Request/Response DTOs
       repository/postgres/
          language_repository.go
 timezone/                  #  Timezone Aggregate
    entity.go              # Timezone entity (BaseAggregate)
    repository.go          # IRepository interface
    usecase.go             # IUseCase interface + implementation
    adapter/
       http/
          handler.go       # HTTP handlers (CRUD)
          dto.go           # Request/Response DTOs
       repository/postgres/
          timezone_repository.go
 region/                    # Region Aggregate (planned)
 payment_method/            # Payment Method Aggregate (planned)
```

---

## Aggregates

### 1. Country Aggregate

**Aggregate Root**: `Country` (extends `BaseAggregate`)

**Attributes**:

- `ID` (UUID v7) - Primary key (from BaseAggregate)
- `Code` (CHAR(2)) - ISO 3166-1 alpha-2 code (e.g., "US", "GB", "UA")
- `Code3` (CHAR(3)) - ISO 3166-1 alpha-3 code (e.g., "USA", "GBR", "UKR")
- `NumericCode` (INT) - ISO 3166-1 numeric code (e.g., 840, 826, 804)
- `Name` (VARCHAR) - Country name (e.g., "United States", "Ukraine")
- `NameLocal` (VARCHAR) - Country name in local language
- `PhoneCode` (VARCHAR) - International dialing code (e.g., "+1", "+380")
- `IsActive` (BOOL) - Active status (for filtering)
- `CreatedAt` (TIMESTAMP) - Record creation time (from BaseAggregate)
- `UpdatedAt` (TIMESTAMP) - Last update time (from BaseAggregate)

**Factory Method**:

```go
func NewCountry(code, name, phoneCode string) (*Country, error)
```

**Business Rules**:

- Code must be uppercase 2-letter ISO standard
- Code3 must be uppercase 3-letter ISO standard
- NumericCode must be valid 3-digit number
- Name must not be empty
- PhoneCode must start with "+"
- Cannot delete country with references

**API Endpoints** (Full CRUD):

```
GET    /api/v1/countries           # List all countries
POST   /api/v1/countries           # Create new country (admin)
GET    /api/v1/countries/:code     # Get country by ISO code
PUT    /api/v1/countries/:id       # Update country (admin)
DELETE /api/v1/countries/:id       # Delete country (admin)
```

**Query Parameters**:

- `?is_active=true` - Filter active countries only

**Example Response**:

```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGABC...",
      "code": "US",
      "code3": "USA",
      "numeric_code": 840,
      "name": "United States",
      "name_local": "United States",
      "phone_code": "+1",
      "is_active": true,
      "created_at": "2025-12-27T10:00:00Z",
      "updated_at": "2025-12-27T10:00:00Z"
    },
    {
      "id": "01JGXYZ...",
      "code": "UA",
      "code3": "UKR",
      "numeric_code": 804,
      "name": "Ukraine",
      "name_local": "Ukraine",
      "phone_code": "+380",
      "is_active": true,
      "created_at": "2025-12-27T10:00:00Z",
      "updated_at": "2025-12-27T10:00:00Z"
    }
  ]
}
```

---

### 2. Currency Aggregate

**Aggregate Root**: `Currency` (extends `BaseAggregate`)

**Attributes**:

- `ID` (UUID v7) - Primary key (from BaseAggregate)
- `Code` (CHAR(3)) - ISO 4217 code (e.g., "USD", "EUR", "UAH")
- `Name` (VARCHAR) - Currency name (e.g., "US Dollar", "Euro")
- `Symbol` (VARCHAR) - Currency symbol (e.g., "$", "€", "₴")
- `DecimalPlaces` (INT) - Number of decimal places (usually 2)
- `IsActive` (BOOL) - Active status
- `CreatedAt` (TIMESTAMP) - Record creation time (from BaseAggregate)
- `UpdatedAt` (TIMESTAMP) - Last update time (from BaseAggregate)

**Factory Method**:

```go
func NewCurrency(code, name, symbol string, decimalPlaces int) (*Currency, error)
```

**Business Rules**:

- Code must be uppercase 3-letter ISO 4217 standard
- Symbol can be Unicode (₴, £, ¥, €, etc.)
- DecimalPlaces typically 2 (except JPY=0, BHD=3)
- Name must not be empty
- Cannot delete currency with financial records

**API Endpoints** (Full CRUD):

```
GET    /api/v1/currencies          # List all currencies
POST   /api/v1/currencies          # Create new currency (admin)
GET    /api/v1/currencies/:code    # Get currency by ISO code
PUT    /api/v1/currencies/:id      # Update currency (admin)
DELETE /api/v1/currencies/:id      # Delete currency (admin)
```

**Example Response**:

```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGABC...",
      "code": "USD",
      "name": "US Dollar",
      "symbol": "$",
      "decimal_places": 2,
      "is_active": true,
      "created_at": "2025-12-27T10:00:00Z",
      "updated_at": "2025-12-27T10:00:00Z"
    },
    {
      "id": "01JGXYZ...",
      "code": "EUR",
      "name": "Euro",
      "symbol": "€",
      "decimal_places": 2,
      "is_active": true,
      "created_at": "2025-12-27T10:00:00Z",
      "updated_at": "2025-12-27T10:00:00Z"
    }
  ]
}
```

---

### 3. Language Aggregate

**Aggregate Root**: `Language` (extends `BaseAggregate`)

**Attributes**:

- `ID` (UUID v7) - Primary key (from BaseAggregate)
- `Code` (CHAR(2)) - ISO 639-1 code (e.g., "en", "uk", "de")
- `Name` (VARCHAR) - Language name in English (e.g., "English", "Ukrainian")
- `NativeName` (VARCHAR) - Language name in native script (e.g., "Ukrainian")
- `IsActive` (BOOL) - Active status
- `CreatedAt` (TIMESTAMP) - Record creation time (from BaseAggregate)
- `UpdatedAt` (TIMESTAMP) - Last update time (from BaseAggregate)

**Factory Method**:

```go
func NewLanguage(code, name, nativeName string) (*Language, error)
```

**Business Rules**:

- Code must be lowercase 2-letter ISO 639-1 standard
- NativeName uses native Unicode script
- Name must not be empty
- Cannot delete language used in user profiles

**API Endpoints** (Full CRUD):

```
GET    /api/v1/languages           # List all languages
POST   /api/v1/languages           # Create new language (admin)
GET    /api/v1/languages/:code     # Get language by ISO code
PUT    /api/v1/languages/:id       # Update language (admin)
DELETE /api/v1/languages/:id       # Delete language (admin)
```

**Example Response**:

```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGABC...",
      "code": "en",
      "name": "English",
      "native_name": "English",
      "is_active": true,
      "created_at": "2025-12-27T10:00:00Z",
      "updated_at": "2025-12-27T10:00:00Z"
    },
    {
      "id": "01JGXYZ...",
      "code": "uk",
      "name": "Ukrainian",
      "native_name": "Ukrainian",
      "is_active": true,
      "created_at": "2025-12-27T10:00:00Z",
      "updated_at": "2025-12-27T10:00:00Z"
    }
  ]
}
```

---

### 4. Timezone Aggregate

**Aggregate Root**: `Timezone` (extends `BaseAggregate`)

**Attributes**:

- `ID` (UUID v7) - Primary key (from BaseAggregate)
- `Name` (VARCHAR) - IANA timezone identifier (e.g., "America/New_York", "Europe/Kyiv")
- `Abbreviation` (VARCHAR) - Common abbreviation (e.g., "EST", "EET", "UTC")
- `UtcOffsetSeconds` (INT) - Current UTC offset in seconds
- `IsActive` (BOOL) - Active status
- `CreatedAt` (TIMESTAMP) - Record creation time (from BaseAggregate)
- `UpdatedAt` (TIMESTAMP) - Last update time (from BaseAggregate)

**Factory Method**:

```go
func NewTimezone(name, abbreviation string, utcOffsetSeconds int) (*Timezone, error)
```

**Business Rules**:

- Name must be valid IANA timezone database identifier (e.g., "Europe/Kyiv")
- Abbreviation is human-readable hint (not unique, e.g., "EST", "PST")
- UtcOffsetSeconds stored as integer (e.g., -18000 for UTC-5, 7200 for UTC+2)
- Cannot delete timezone used in user profiles

**API Endpoints** (Full CRUD):

```
GET    /api/v1/timezones           # List all timezones
POST   /api/v1/timezones           # Create new timezone (admin)
GET    /api/v1/timezones/*name     # Get timezone by IANA name (supports paths like Europe/Kyiv)
PUT    /api/v1/timezones/:id       # Update timezone (admin)
DELETE /api/v1/timezones/:id       # Delete timezone (admin)
```

**Example Response**:

```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGABC...",
      "name": "America/New_York",
      "abbreviation": "EST",
      "utc_offset_seconds": -18000,
      "is_active": true,
      "created_at": "2025-12-27T10:00:00Z",
      "updated_at": "2025-12-27T10:00:00Z"
    },
    {
      "id": "01JGXYZ...",
      "name": "Europe/Kyiv",
      "abbreviation": "EET",
      "utc_offset_seconds": 7200,
      "is_active": true,
      "created_at": "2025-12-27T10:00:00Z",
      "updated_at": "2025-12-27T10:00:00Z"
    }
  ]
}
```

---

## Database Schema

### Tables

**shared_countries** (reference data with BaseAggregate fields):

```sql
CREATE TABLE shared_countries (
    id UUID PRIMARY KEY,                    -- UUID v7 (from BaseAggregate)
    code CHAR(2) UNIQUE NOT NULL,          -- ISO 3166-1 alpha-2
    code3 CHAR(3) NOT NULL,                -- ISO 3166-1 alpha-3
    numeric_code INT NOT NULL,             -- ISO 3166-1 numeric
    name VARCHAR(100) NOT NULL,            -- Country name
    name_local VARCHAR(100),               -- Local language name
    phone_code VARCHAR(10) NOT NULL,       -- International dialing code
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL,         -- Managed by Go code
    updated_at TIMESTAMP NOT NULL          -- Managed by Go code
);

CREATE INDEX idx_countries_code ON shared_countries(code);
CREATE INDEX idx_countries_active ON shared_countries(is_active);
```

**shared_currencies**:

```sql
CREATE TABLE shared_currencies (
    id UUID PRIMARY KEY,                    -- UUID v7 (from BaseAggregate)
    code CHAR(3) UNIQUE NOT NULL,          -- ISO 4217
    name VARCHAR(100) NOT NULL,            -- Currency name
    symbol VARCHAR(10) NOT NULL,           -- Currency symbol
    decimal_places INT NOT NULL DEFAULT 2, -- Number of decimal places
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL,         -- Managed by Go code
    updated_at TIMESTAMP NOT NULL          -- Managed by Go code
);

CREATE INDEX idx_currencies_code ON shared_currencies(code);
CREATE INDEX idx_currencies_active ON shared_currencies(is_active);
```

**shared_languages**:

```sql
CREATE TABLE shared_languages (
    id UUID PRIMARY KEY,                    -- UUID v7 (from BaseAggregate)
    code CHAR(2) UNIQUE NOT NULL,          -- ISO 639-1
    name VARCHAR(100) NOT NULL,            -- Language name in English
    native_name VARCHAR(100) NOT NULL,     -- Native script name
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL,         -- Managed by Go code
    updated_at TIMESTAMP NOT NULL          -- Managed by Go code
);

CREATE INDEX idx_languages_code ON shared_languages(code);
CREATE INDEX idx_languages_active ON shared_languages(is_active);
```

**shared_timezones**:

```sql
CREATE TABLE shared_timezones (
    id UUID PRIMARY KEY,                    -- UUID v7 (from BaseAggregate)
    name VARCHAR(100) UNIQUE NOT NULL,     -- IANA timezone identifier
    abbreviation VARCHAR(10) NOT NULL,     -- Common abbreviation
    utc_offset_seconds INT NOT NULL,       -- Offset in seconds
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL,         -- Managed by Go code
    updated_at TIMESTAMP NOT NULL          -- Managed by Go code
);

CREATE INDEX idx_timezones_name ON shared_timezones(name);
CREATE INDEX idx_timezones_active ON shared_timezones(is_active);
```

**Key Changes from Previous Version**:

-  All tables now use UUID v7 primary keys (from BaseAggregate)
-  Timestamps managed in Go code (no DEFAULT CURRENT_TIMESTAMP)
-  Countries table includes code3, numeric_code, name_local, phone_code
-  Currencies table renamed decimal_digits → decimal_places
-  Languages table removed direction field (not used)
-  Timezones table uses utc_offset_seconds (INT) instead of VARCHAR

---

## Data Population

### Initial Data (via seed files)

**Data loaded from**: `internal/infrastructure/seed/shared/`

- **Countries**: 195+ recognized countries (UN members + observers)
- **Currencies**: 170+ active currencies (ISO 4217)
- **Languages**: 50+ major languages (ISO 639-1)
- **Timezones**: 600+ IANA timezones

**Seed Files**:

```
internal/infrastructure/seed/shared/
 countries.go       # 195+ countries with factory methods
 currencies.go      # 170+ currencies with factory methods
 languages.go       # 50+ languages with factory methods
 timezones.go       # 600+ timezones with factory methods
```

**Running Seeds**:

```bash
# Seed all contexts
make seed

# Seed shared context only
make seed-shared
```

---

## Updates

### How to update reference data

**1. Create migration**:

```bash
make migrate-create-core NAME=update_currency_symbols
```

**2. Edit migration file**:

```sql
-- migrations/shared/000005_update_currency_symbols.up.sql
UPDATE shared_currencies SET symbol = '₴' WHERE code = 'UAH';
UPDATE shared_currencies SET symbol = '₽' WHERE code = 'RUB';
```

**3. Run migration**:

```bash
make migrate
```

### Update Frequency

- **Countries**: 1-2 times per year (new countries rare, ISO updates)
- **Currencies**: 2-3 times per year (new currencies, symbol changes)
- **Languages**: 1 time per year (rarely changes)
- **Timezones**: As needed (DST rule changes, IANA updates)

---

## Caching Strategy

### Application-Level Caching

**In-Memory Cache** (recommended for reference data):

```go
type CountryCache struct {
    data map[string]*Country
    mu   sync.RWMutex
    ttl  time.Duration
}

// Load all countries on startup
func (c *CountryCache) LoadAll(ctx context.Context) error {
    countries, err := repo.ListCountries(ctx)
    if err != nil {
        return err
    }

    c.mu.Lock()
    defer c.mu.Unlock()

    for _, country := range countries {
        c.data[country.Code] = country
    }

    return nil
}

// Get from cache (no DB query)
func (c *CountryCache) GetByCode(code string) (*Country, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    country, ok := c.data[code]
    if !ok {
        return nil, ErrCountryNotFound
    }

    return country, nil
}
```

**Redis Cache** (for distributed deployments):

```yaml
# Cache TTL for reference data
redis:
  cache:
    countries: 24h # Rarely changes
    currencies: 24h # Rarely changes
    languages: 24h # Rarely changes
    timezones: 1h # DST changes
```

### Cache Invalidation

**Manual invalidation** (after migration):

```bash
# Clear Redis cache
redis-cli FLUSHDB

# Or specific keys
redis-cli DEL "shared:countries:*"
redis-cli DEL "shared:currencies:*"
```

**Automatic invalidation** (via domain events):

```go
// After country update
event := bus.NewBaseEvent("shared.country.updated", country.Code)
bus.Publish(ctx, bus.TopicCountryUpdated, event)

// Listener clears cache
func (h *CacheInvalidationHandler) HandleCountryUpdated(ctx context.Context, e bus.Event) error {
    return h.cache.Clear("countries")
}
```

---

## Testing

Baseline budget and risk tiers follow [docs/guides/testing-patterns.md](../../docs/guides/testing-patterns.md): Unit 20–30, Smoke 6–9, Integration 6–10, with high-risk aggregates +30–50%.

### Test Coverage

**Unit Tests** (in-place):

- `country/entity_test.go` - Country validation (20+ tests)
- `currency/entity_test.go` - Currency validation (20+ tests)
- `language/entity_test.go` - Language validation (20+ tests)
- `timezone/entity_test.go` - Timezone validation (20+ tests)
- `country/usecase_test.go` - Business logic tests
- `currency/usecase_test.go` - Business logic tests
- `language/usecase_test.go` - Business logic tests
- `timezone/usecase_test.go` - Business logic tests
- `*/adapter/http/dto_test.go` - DTO conversion tests (4 files)

**Integration Tests** (mirror path structure):

- `test/integration/contexts/shared/country/repository_test.go` - CRUD + queries (6 tests)
- `test/integration/contexts/shared/currency/repository_test.go` - CRUD + queries (6 tests)
- `test/integration/contexts/shared/language/repository_test.go` - CRUD + queries (6 tests)
- `test/integration/contexts/shared/timezone/repository_test.go` - CRUD + queries (6 tests)

**Total**: 80+ unit tests + 24 integration tests = **104+ tests, all PASS** 

**Run Tests**:

```bash
# All shared context tests
go test ./internal/contexts/shared/... -v

# Integration tests with real DB
go test ./test/integration/contexts/shared/... -v

# Specific aggregate
go test ./internal/contexts/shared/country/... -v
```

---

## Use Cases

### 1. Country Dropdown (Address Form)

**Frontend**:

```javascript
// Fetch countries for dropdown
fetch("/api/v1/countries?is_active=true")
  .then((res) => res.json())
  .then((data) => {
    const countries = data.data;
    // Populate <select> element with country.name
  });
```

**Response**:

```json
{
  "status": "success",
  "data": [
    { "id": "01JG...", "code": "US", "name": "United States" },
    { "id": "01JG...", "code": "UA", "name": "Ukraine" },
    { "id": "01JG...", "code": "GB", "name": "United Kingdom" }
  ]
}
```

### 2. Currency Selection (Billing)

```go
// Get currency by code
currency, err := currencyUC.GetByCode(ctx, "USD")
if err != nil {
    return err
}

// Format money with symbol
formatted := fmt.Sprintf("%s %.2f", currency.Symbol, amount)
// Result: "$ 123.45"
```

### 3. Timezone Selection (User Profile)

```go
// List all timezones
timezones, err := timezoneUC.List(ctx)
if err != nil {
    return err
}

// User selects timezone by ID
profile.TimezoneID = selectedTimezoneID
```

---

## Dependencies

### Internal

- `pkg/aggregate` - BaseAggregate pattern (ID, CreatedAt, UpdatedAt, Touch)
- `pkg/uuidv7` - Time-ordered UUIDs (for all entities)
- `pkg/logger` - Structured logging with context
- `pkg/response` - Standard HTTP responses
- `pkg/cache` - Redis-based caching layer
- `pkg/bus` - Domain events (for cache invalidation)

### External

- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/jmoiron/sqlx` - SQL extensions

---

## Future Enhancements

### Planned Features

- [ ] **Regions aggregate**: Africa, Americas, Asia, Europe, Oceania
- [ ] **Cities aggregate**: Major world cities (for address autocomplete)
- [ ] **Payment Methods**: Credit cards, PayPal, Stripe, bank transfers
- [ ] **Phone Prefixes**: Country phone codes (e.g., +1, +380, +44)
- [ ] **VAT Rates**: Tax rates per country (EU VAT, US sales tax)
- [ ] **Postal Code Formats**: Validation patterns per country
- [ ] **Address Formats**: Country-specific address templates
- [ ] **Currency Exchange Rates**: Daily rates from external API

### Not Planned (belongs elsewhere)

- User-specific data → Identity Context
- Customer data → Customer Management Context
- Financial transactions → Billing Context

---

## Related Documentation

- [Documentation Index](../../../docs/INDEX.md)
- [Clean Architecture Summary](../../../docs/concepts/clean-architecture.md)
- [Migrations README](../../../migrations/README.md)
- [Event Bus Documentation](../../../pkg/bus/README.md)
- [Testing Patterns](../../../docs/guides/testing-patterns.md)
- [Testing Guide](../../../test/README.md)

---

**Last Updated**: 2026-01-03  
**Status**: Production-ready ( BaseAggregate refactoring complete)  
**Architecture**: Clean Architecture with DDD + BaseAggregate pattern  
**Test Coverage**: 104+ tests (80 unit + 24 integration), all PASS   
**Data Sources**: ISO standards, IANA database  
**Maintainer**: Promenade Team
