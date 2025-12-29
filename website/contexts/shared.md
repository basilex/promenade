# Shared Context

**Shared Context** provides reference data for all contexts (read-only).

## Overview

The Shared Context contains reference data used across all bounded contexts:

- **Country** - ISO country codes (250+ countries)
- **Currency** - Currency codes with symbols and decimals
- **Language** - ISO 639-1 language codes
- **Timezone** - IANA timezone database

## Status

✅ **Production-ready**

## Aggregates

### Country
ISO 3166-1 country codes with localized names.

### Currency
ISO 4217 currency codes (USD, EUR, UAH, etc.)

### Language
ISO 639-1 language codes (en, uk, de, etc.)

### Timezone
IANA timezone identifiers (Europe/Kyiv, America/New_York, etc.)

## API Endpoints

```http
GET /api/v1/shared/countries      # List all countries
GET /api/v1/shared/currencies     # List all currencies
GET /api/v1/shared/languages      # List all languages
GET /api/v1/shared/timezones      # List all timezones
```

## Usage Example

```go
// Get all countries
countries, err := countryUC.GetCountries(ctx)

// Get all active currencies
currencies, err := currencyUC.GetActiveCurrencies(ctx)
```

## Next Steps

- [Identity Context](/contexts/identity) - User authentication and management
- [Customer Management](/contexts/) - CRM functionality
- [Architecture Guide](/guide/architecture) - Bounded Contexts overview
