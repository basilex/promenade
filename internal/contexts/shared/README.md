# Shared Context (Bounded Context)

##  Overview

**Shared Context** (also known as **Shared Kernel**) manages **reference data** used across all bounded contexts in Promenade CRM. This context provides **read-only, globally consistent data** such as countries, currencies, languages, timezones, and payment methods.

**Status**:  Production-ready

---

##  Bounded Context Responsibilities

### Core Domain

**Reference Data Management**:

-  **Countries**: ISO 3166-1 alpha-2 codes, names, regions
-  **Currencies**: ISO 4217 codes, symbols, decimal digits
-  **Languages**: ISO 639-1 codes, native names, text direction (LTR/RTL)
-  **Timezones**: IANA database identifiers, UTC offsets, DST support
-  **Regions**: Geographic regions (Europe, Asia, Americas, etc.)
-  **Cities**: Major world cities (optional, for address autocomplete)
-  **Payment Methods**: Credit cards, bank transfers, e-wallets

### Characteristics

-  **Read-only for applications**: Data changes via migrations only
-  **Globally consistent**: All contexts use same reference data
-  **Rarely updated**: Changes 1-2 times per year (ISO standard updates)
-  **High performance**: Heavy caching (Redis/in-memory)
-  **Multi-language**: Countries, languages, regions support i18n

---

##  Architecture

### Directory Structure

```
shared/
 README.md                  # This file
 router.go                  # HTTP routes registration
 country/                   # Country Aggregate
    aggregate/
       country.go
    repository/
       country_repository.go
    usecase/
       country_usecase.go
    handler/
        country_handler.go
 currency/                  # Currency Aggregate
    aggregate/
    repository/
    usecase/
    handler/
 language/                  # Language Aggregate
    aggregate/
    repository/
    usecase/
    handler/
 timezone/                  # Timezone Aggregate
    aggregate/
    repository/
    usecase/
    handler/
 region/                    # Region Aggregate (planned)
 payment_method/            # Payment Method Aggregate (planned)
```

---

##  Aggregates

### 1. Country Aggregate

**Aggregate Root**: `Country`

**Attributes**:

- `Code` (CHAR(2)) - ISO 3166-1 alpha-2 code (e.g., "US", "GB", "UA")
- `Name` (VARCHAR) - Country name (e.g., "United States", "Ukraine")
- `Region` (VARCHAR) - Geographic region (e.g., "Europe", "Asia")
- `IsActive` (BOOL) - Active status (for filtering sanctioned countries)
- `CurrencyCode` (CHAR(3)) - Default currency (ISO 4217)

**Business Rules**:

- Code must be uppercase 2-letter ISO standard
- Name must be unique
- Must have valid currency relationship
- Cannot deactivate country with active users/customers

**API Endpoints**:

```
GET /api/v1/shared/countries           # List all countries
GET /api/v1/shared/countries/:code     # Get country by code
GET /api/v1/shared/countries/region/:region  # Filter by region
```

**Query Parameters**:

- `?is_active=true` - Filter active countries only
- `?region=Europe` - Filter by geographic region
- `?currency=USD` - Filter countries using specific currency

**Example Response**:

```json
{
  "status": "success",
  "data": [
    {
      "code": "US",
      "name": "United States",
      "region": "Americas",
      "currency_code": "USD",
      "is_active": true
    },
    {
      "code": "UA",
      "name": "Ukraine",
      "region": "Europe",
      "currency_code": "UAH",
      "is_active": true
    }
  ]
}
```

---

### 2. Currency Aggregate

**Aggregate Root**: `Currency`

**Attributes**:

- `Code` (CHAR(3)) - ISO 4217 code (e.g., "USD", "EUR", "UAH")
- `Name` (VARCHAR) - Currency name (e.g., "US Dollar", "Euro")
- `Symbol` (VARCHAR) - Currency symbol (e.g., "$", "€", "₴")
- `DecimalDigits` (INT) - Number of decimal places (usually 2)
- `IsActive` (BOOL) - Active status

**Business Rules**:

- Code must be uppercase 3-letter ISO standard
- Symbol can be Unicode (₴, £, ¥, etc.)
- DecimalDigits typically 2 (except JPY=0, BHD=3)
- Cannot deactivate currency with active financial records

**API Endpoints**:

```
GET /api/v1/shared/currencies          # List all currencies
GET /api/v1/shared/currencies/:code    # Get currency by code
```

**Example Response**:

```json
{
  "status": "success",
  "data": [
    {
      "code": "USD",
      "name": "US Dollar",
      "symbol": "$",
      "decimal_digits": 2,
      "is_active": true
    },
    {
      "code": "EUR",
      "name": "Euro",
      "symbol": "€",
      "decimal_digits": 2,
      "is_active": true
    }
  ]
}
```

---

### 3. Language Aggregate

**Aggregate Root**: `Language`

**Attributes**:

- `Code` (CHAR(2)) - ISO 639-1 code (e.g., "en", "uk", "de")
- `Name` (VARCHAR) - Language name in English (e.g., "English", "Ukrainian")
- `NativeName` (VARCHAR) - Language name in native script (e.g., "Українська")
- `Direction` (ENUM) - Text direction: LTR (left-to-right) or RTL (right-to-left)
- `IsActive` (BOOL) - Active status

**Business Rules**:

- Code must be lowercase 2-letter ISO standard
- NativeName uses native Unicode script
- Direction defaults to LTR (RTL for Arabic, Hebrew, Persian, Urdu)
- Cannot deactivate language if used in user profiles

**API Endpoints**:

```
GET /api/v1/shared/languages           # List all languages
GET /api/v1/shared/languages/:code     # Get language by code
GET /api/v1/shared/languages/rtl       # List RTL languages
```

**Example Response**:

```json
{
  "status": "success",
  "data": [
    {
      "code": "en",
      "name": "English",
      "native_name": "English",
      "direction": "LTR",
      "is_active": true
    },
    {
      "code": "uk",
      "name": "Ukrainian",
      "native_name": "Українська",
      "direction": "LTR",
      "is_active": true
    },
    {
      "code": "ar",
      "name": "Arabic",
      "native_name": "العربية",
      "direction": "RTL",
      "is_active": true
    }
  ]
}
```

---

### 4. Timezone Aggregate

**Aggregate Root**: `Timezone`

**Attributes**:

- `ID` (UUID) - Primary key (UUIDv7)
- `Name` (VARCHAR) - IANA timezone identifier (e.g., "America/New_York")
- `Abbreviation` (VARCHAR) - Common abbreviation (e.g., "EST", "UTC")
- `UtcOffset` (VARCHAR) - Current UTC offset (e.g., "-05:00", "+02:00")
- `SupportsDST` (BOOL) - Daylight Saving Time support
- `IsActive` (BOOL) - Active status

**Business Rules**:

- Name must be valid IANA timezone database identifier
- UtcOffset updates twice per year (DST transitions)
- Cannot deactivate timezone if used in user profiles
- Abbreviations are human-readable hints (not unique)

**API Endpoints**:

```
GET /api/v1/shared/timezones           # List all timezones
GET /api/v1/shared/timezones/:id       # Get timezone by ID
GET /api/v1/shared/timezones/search?q=New+York  # Search timezones
GET /api/v1/shared/timezones/dst       # List timezones with DST
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
      "utc_offset": "-05:00",
      "supports_dst": true,
      "is_active": true
    },
    {
      "id": "01JGXYZ...",
      "name": "Europe/Kiev",
      "abbreviation": "EET",
      "utc_offset": "+02:00",
      "supports_dst": true,
      "is_active": true
    }
  ]
}
```

---

##  Database Schema

### Tables (namespace: `shared_`)

**shared_countries**

```sql
CREATE TABLE shared_countries (
    code CHAR(2) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    region VARCHAR(50) NOT NULL,
    currency_code CHAR(3) NOT NULL REFERENCES shared_currencies(code),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_countries_region ON shared_countries(region);
CREATE INDEX idx_countries_active ON shared_countries(is_active);
```

**shared_currencies**

```sql
CREATE TABLE shared_currencies (
    code CHAR(3) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    decimal_digits INT NOT NULL DEFAULT 2,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_currencies_active ON shared_currencies(is_active);
```

**shared_languages**

```sql
CREATE TABLE shared_languages (
    code CHAR(2) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    native_name VARCHAR(100) NOT NULL,
    direction VARCHAR(3) NOT NULL DEFAULT 'LTR' CHECK (direction IN ('LTR', 'RTL')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_languages_active ON shared_languages(is_active);
CREATE INDEX idx_languages_direction ON shared_languages(direction);
```

**shared_timezones**

```sql
CREATE TABLE shared_timezones (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(100) UNIQUE NOT NULL,
    abbreviation VARCHAR(10) NOT NULL,
    utc_offset VARCHAR(10) NOT NULL,
    supports_dst BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_timezones_name ON shared_timezones(name);
CREATE INDEX idx_timezones_dst ON shared_timezones(supports_dst);
CREATE INDEX idx_timezones_active ON shared_timezones(is_active);
```

---

##  Data Population

### Initial Data (via migrations)

**Countries**: 195 recognized countries (UN members + observers)  
**Currencies**: 170+ active currencies (ISO 4217)  
**Languages**: 50+ major languages (ISO 639-1)  
**Timezones**: 400+ IANA timezones

### Data Sources

- **Countries**: [ISO 3166-1](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)
- **Currencies**: [ISO 4217](https://en.wikipedia.org/wiki/ISO_4217)
- **Languages**: [ISO 639-1](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes)
- **Timezones**: [IANA Time Zone Database](https://www.iana.org/time-zones)

### Migration Files

```
migrations/shared/
 000001_shared_seed_currencies.up.sql      # 170+ currencies
 000002_shared_seed_countries.up.sql       # 195 countries
 000003_shared_seed_languages.up.sql       # 50+ languages
 000004_shared_seed_timezones.up.sql       # 400+ timezones
```

---

##  Updates

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

- **Countries**: 1-2 times per year (new countries rare)
- **Currencies**: 2-3 times per year (new currencies, symbol changes)
- **Languages**: 1 time per year (rarely changes)
- **Timezones**: 2 times per year (DST rule changes)

---

##  Caching Strategy

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

##  Testing

### Test Coverage

**Unit Tests**:

- Country aggregate validation
- Currency aggregate validation
- Language aggregate validation
- Timezone aggregate validation

**Integration Tests**:

- Country repository CRUD
- Currency repository CRUD
- Language repository CRUD
- Timezone repository CRUD

**Run Tests**:

```bash
# All shared context tests
go test ./internal/contexts/shared/... -v

# Specific aggregate
go test ./internal/contexts/shared/country/... -v
```

---

##  Use Cases

### 1. Country Dropdown (Address Form)

**Frontend**:

```javascript
// Fetch countries for dropdown
fetch("/api/v1/shared/countries?is_active=true")
  .then((res) => res.json())
  .then((data) => {
    const countries = data.data;
    // Populate <select> element
  });
```

**Response**:

```json
{
  "status": "success",
  "data": [
    { "code": "US", "name": "United States" },
    { "code": "UA", "name": "Ukraine" },
    { "code": "GB", "name": "United Kingdom" }
  ]
}
```

### 2. Currency Conversion (Billing)

```go
// Get currency details
currency, err := currencyRepo.GetByCode(ctx, "USD")
if err != nil {
    return err
}

// Format money
formatted := fmt.Sprintf("%s %.2f", currency.Symbol, amount)
// Result: "$ 123.45"
```

### 3. Timezone Selection (User Profile)

```go
// List all timezones
timezones, err := timezoneRepo.ListTimezones(ctx)
if err != nil {
    return err
}

// User selects "America/New_York"
profile.TimezoneID = timezone.ID
```

---

##  Dependencies

### Internal

- `pkg/uuidv7` - Time-ordered UUIDs (for timezones)
- `pkg/logger` - Structured logging
- `pkg/response` - Standard HTTP responses
- `pkg/bus` - Domain events (for cache invalidation)

### External

- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/jmoiron/sqlx` - SQL extensions

---

##  Future Enhancements

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

-  User-specific data → Identity Context
-  Customer data → Customer Management Context
-  Financial transactions → Billing Context

---

##  Related Documentation

- [Architecture Overview](../../../docs/ARCHITECTURE_OVERVIEW.md)
- [Database Migrations](../../../docs/MIGRATION_ARCHITECTURE.md)
- [Event Bus Documentation](../../../pkg/bus/README.md)
- [Testing Guide](../../../test/TESTING_STRUCTURE.md)

---

**Last Updated**: 2025-12-27  
**Status**:  Production-ready  
**Data Sources**: ISO standards, IANA database  
**Maintainer**: Promenade Team
