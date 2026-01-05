# Shared Context

**Shared Context** provides reference data for all contexts (read-only).

## Overview

The Shared Context (Shared Kernel) manages **reference data** used across all bounded contexts:

- **Country** - ISO 3166-1 country codes (250+ countries)
- **Currency** - ISO 4217 currency codes with symbols
- **Language** - ISO 639-1 language codes  
- **Timezone** - IANA timezone database

### Characteristics

- **Read-only** - Data changes via migrations only
- **Globally consistent** - All contexts use same data
- **Rarely updated** - Changes 1-2 times per year (ISO updates)
- **High performance** - Heavy caching (Redis/in-memory)
- **Multi-language** - Supports i18n

## Status

 **Production-ready**

## Aggregates

### Country Aggregate

**Attributes:**
- `Code` (CHAR(2)) - ISO 3166-1 alpha-2 (e.g., "US", "UA", "GB")
- `Name` (VARCHAR) - Country name
- `Region` (VARCHAR) - Geographic region (Europe, Asia, etc.)
- `IsActive` (BOOL) - Active status
- `CurrencyCode` (CHAR(3)) - Default currency

**Business Rules:**
- Code must be uppercase 2-letter ISO standard
- Name must be unique
- Must have valid currency relationship
- Cannot deactivate country with active users

**API:**
```http
GET /api/v1/shared/countries              # List all
GET /api/v1/shared/countries/:code        # Get by code
GET /api/v1/shared/countries/region/:region  # Filter by region
```

### Currency Aggregate

**Attributes:**
- `Code` (CHAR(3)) - ISO 4217 code (e.g., "USD", "EUR", "UAH")
- `Name` (VARCHAR) - Currency name
- `Symbol` (VARCHAR) - Symbol (e.g., "$", "€", "₴")
- `DecimalDigits` (INT) - Precision (e.g., 2 for most, 0 for JPY)
- `IsActive` (BOOL) - Active status

**Business Rules:**
- Code must be uppercase 3-letter ISO standard
- Symbol must not conflict with existing currencies
- DecimalDigits must be 0-4
- Cannot deactivate currency used by active countries

**API:**
```http
GET /api/v1/shared/currencies             # List all
GET /api/v1/shared/currencies/:code       # Get by code
GET /api/v1/shared/currencies/active      # Only active
```

### Language Aggregate

**Attributes:**
- `Code` (CHAR(2)) - ISO 639-1 code (e.g., "en", "uk", "de")
- `Name` (VARCHAR) - English name (e.g., "English", "Ukrainian")
- `NativeName` (VARCHAR) - Native name (e.g., "Українська")
- `Direction` (ENUM) - Text direction (LTR, RTL)
- `IsActive` (BOOL) - Active status

**Business Rules:**
- Code must be lowercase 2-letter ISO standard
- Name and NativeName must be unique
- Direction must be LTR or RTL
- Must have at least one active language

**API:**
```http
GET /api/v1/shared/languages              # List all
GET /api/v1/shared/languages/:code        # Get by code
GET /api/v1/shared/languages/active       # Only active
```

### Timezone Aggregate

**Attributes:**
- `Code` (VARCHAR) - IANA identifier (e.g., "Europe/Kyiv", "America/New_York")
- `Name` (VARCHAR) - Display name
- `UTCOffset` (VARCHAR) - Offset from UTC (e.g., "+02:00", "-05:00")
- `SupportsDST` (BOOL) - Daylight Saving Time support
- `IsActive` (BOOL) - Active status

**Business Rules:**
- Code must be valid IANA timezone identifier
- UTC offset must be valid format
- Name must be unique
- Cannot deactivate timezone used by active users

**API:**
```http
GET /api/v1/shared/timezones              # List all
GET /api/v1/shared/timezones/:code        # Get by code
GET /api/v1/shared/timezones/region/:region  # Filter by region
```

## Database Schema

### Tables

**shared_countries:**
```sql
CREATE TABLE shared_countries (
    code CHAR(2) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    region VARCHAR(50),
    currency_code CHAR(3) REFERENCES shared_currencies(code),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**shared_currencies:**
```sql
CREATE TABLE shared_currencies (
    code CHAR(3) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10),
    decimal_digits INT DEFAULT 2,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**shared_languages:**
```sql
CREATE TABLE shared_languages (
    code CHAR(2) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    native_name VARCHAR(100),
    direction VARCHAR(3) DEFAULT 'LTR',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**shared_timezones:**
```sql
CREATE TABLE shared_timezones (
    code VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    utc_offset VARCHAR(10),
    supports_dst BOOLEAN DEFAULT false,
    region VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Seed Data

### Sample Countries (250+ total)

| Code | Name           | Region  | Currency |
| ---- | -------------- | ------- | -------- |
| US   | United States  | Americas| USD      |
| GB   | United Kingdom | Europe  | GBP      |
| UA   | Ukraine        | Europe  | UAH      |
| DE   | Germany        | Europe  | EUR      |
| FR   | France         | Europe  | EUR      |

### Sample Currencies (180+ total)

| Code | Name            | Symbol | Decimals |
| ---- | --------------- | ------ | -------- |
| USD  | US Dollar       | $      | 2        |
| EUR  | Euro            | €      | 2        |
| GBP  | British Pound   | £      | 2        |
| UAH  | Ukrainian Hryvnia | ₴    | 2        |
| JPY  | Japanese Yen    | ¥      | 0        |

### Sample Languages (180+ total)

| Code | Name       | Native Name | Direction |
| ---- | ---------- | ----------- | --------- |
| en   | English    | English     | LTR       |
| uk   | Ukrainian  | Українська  | LTR       |
| de   | German     | Deutsch     | LTR       |
| ar   | Arabic     | العربية     | RTL       |
| he   | Hebrew     | עברית       | RTL       |

### Sample Timezones (400+ total)

| Code             | Name            | Offset  | DST   |
| ---------------- | --------------- | ------- | ----- |
| Europe/Kyiv      | Kyiv            | +02:00  | Yes   |
| America/New_York | Eastern Time    | -05:00  | Yes   |
| Asia/Tokyo       | Tokyo           | +09:00  | No    |
| Europe/London    | London          | +00:00  | Yes   |
| UTC              | UTC             | +00:00  | No    |

## Usage Examples

### Get All Countries

```go
countries, err := countryUC.GetCountries(ctx)
for _, country := range countries {
    fmt.Printf("%s: %s (%s)\n", country.Code, country.Name, country.Region)
}
```

### Get Active Currencies

```go
currencies, err := currencyUC.GetActiveCurrencies(ctx)
for _, currency := range currencies {
    fmt.Printf("%s: %s %s\n", currency.Code, currency.Symbol, currency.Name)
}
```

### Validate Country Code

```go
country, err := countryUC.GetCountryByCode(ctx, "UA")
if err != nil {
    return fmt.Errorf("invalid country code: %w", err)
}
```

## Caching Strategy

### Redis Cache

**TTL Configuration:**
- Countries: 24 hours (rarely change)
- Currencies: 24 hours (rarely change)
- Languages: 24 hours (rarely change)
- Timezones: 7 days (never change)

**Cache Keys:**
```
shared:countries:all
shared:countries:{code}
shared:currencies:all
shared:currencies:{code}
shared:languages:all
shared:timezones:all
```

### In-Memory Cache

Optional in-memory caching for ultra-fast access:

```go
var countriesCache map[string]*Country
var currenciesCache map[string]*Currency
```

## Testing

**Integration Tests:**
- Country repository: 6 tests
- Currency repository: 6 tests
- Language repository: 6 tests
- Timezone repository: 6 tests

**Smoke Tests:**
- Country handler: 5 tests
- Currency handler: 5 tests
- Language handler: 5 tests
- Timezone handler: 5 tests

## Performance

**Benchmarks:**
- Get all countries: ~2ms (cached)
- Get country by code: ~0.5ms (cached)
- List query: ~10ms (no cache)

## Best Practices

1. **Always cache** - Reference data changes rarely
2. **Validate on input** - Check country/currency codes at API boundary
3. **Use ISO standards** - Stick to ISO 3166-1, ISO 4217, ISO 639-1
4. **Update annually** - Review ISO standards once per year
5. **Test migrations** - Seed data must be consistent

## Next Steps

- [Identity Context](/contexts/identity) - User authentication
- [Customer Management](/contexts/customer) - CRM functionality
- [Architecture Guide](/guide/architecture) - Bounded Contexts overview
- [GitHub Docs](https://github.com/basilex/promenade/blob/dev/internal/contexts/shared/README.md) - Complete 663-line guide
