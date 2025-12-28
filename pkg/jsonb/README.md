# JSONB Package

**Purpose:** PostgreSQL JSONB type wrappers with Scanner/Valuer  
**Status:** Production-ready  
**Tests:** 8 tests, 100% coverage

---

## Overview

The `jsonb` package provides **type-safe wrappers** for PostgreSQL JSONB columns. It implements `sql.Scanner` and `driver.Valuer` interfaces, allowing seamless marshaling/unmarshaling between Go types and JSONB.

## Why This Package?

### Problem Without It

```go
// Tedious manual JSON marshaling
metadataJSON, _ := json.Marshal(metadata)
query := `INSERT INTO users (metadata) VALUES ($1)`
db.Exec(query, metadataJSON)

// Manual unmarshaling on read
var metadataJSON []byte
db.QueryRow(`SELECT metadata FROM users WHERE id = $1`, id).Scan(&metadataJSON)
var metadata map[string]interface{}
json.Unmarshal(metadataJSON, &metadata)
```

### Solution With This Package

```go
// Automatic marshaling/unmarshaling
metadata := jsonb.Map{"key": "value"}
db.Exec(`INSERT INTO users (metadata) VALUES ($1)`, metadata)

// Direct scan into JSONB type
var metadata jsonb.Map
db.QueryRow(`SELECT metadata FROM users WHERE id = $1`, id).Scan(&metadata)
```

---

## Installation

```go
import "github.com/basilex/promenade/pkg/jsonb"
```

---

## Available Types

### Map

Wrapper for `map[string]interface{}`

```go
type Map map[string]interface{}
```

### Array

Wrapper for `[]interface{}`

```go
type Array []interface{}
```

### JSON[T]

Generic wrapper for any type

```go
type JSON[T any] struct {
    Value T
}
```

---

## Usage Examples

### 1. JSONB Map (Unstructured)

```go
package main

import (
    "context"
    "github.com/basilex/promenade/pkg/jsonb"
    "github.com/jmoiron/sqlx"
)

type User struct {
    ID       string
    Metadata jsonb.Map // Unstructured metadata
}

func main() {
    db, _ := sqlx.Connect("postgres", "...")
    
    // Create user with metadata
    user := User{
        ID: "user123",
        Metadata: jsonb.Map{
            "last_login":    "2025-12-28",
            "login_count":   42,
            "is_verified":   true,
            "preferences":   map[string]interface{}{
                "theme": "dark",
                "lang":  "en",
            },
        },
    }
    
    // Insert (automatic JSON marshaling)
    _, err := db.Exec(
        `INSERT INTO users (id, metadata) VALUES ($1, $2)`,
        user.ID,
        user.Metadata, // Automatically marshaled to JSONB
    )
    
    // Query (automatic JSON unmarshaling)
    var loadedUser User
    err = db.Get(&loadedUser, `SELECT * FROM users WHERE id = $1`, "user123")
    
    // Access metadata
    lastLogin := loadedUser.Metadata["last_login"].(string)
    loginCount := loadedUser.Metadata["login_count"].(float64) // JSON numbers are float64
}
```

### 2. JSONB Array

```go
type Product struct {
    ID   string
    Tags jsonb.Array // Array of tags
}

func main() {
    product := Product{
        ID: "prod123",
        Tags: jsonb.Array{"electronics", "smartphone", "5g"},
    }
    
    // Insert
    db.Exec(
        `INSERT INTO products (id, tags) VALUES ($1, $2)`,
        product.ID,
        product.Tags,
    )
    
    // Query with JSONB array operator
    var products []Product
    db.Select(&products, `
        SELECT * FROM products 
        WHERE tags @> $1
    `, jsonb.Array{"smartphone"}) // Find products with "smartphone" tag
}
```

### 3. Generic JSON[T] (Type-Safe)

```go
// Structured preferences
type UserPreferences struct {
    Theme    string `json:"theme"`
    Language string `json:"language"`
    Timezone string `json:"timezone"`
}

type User struct {
    ID          string
    Preferences jsonb.JSON[UserPreferences] // Type-safe
}

func main() {
    user := User{
        ID: "user123",
        Preferences: jsonb.JSON[UserPreferences]{
            Value: UserPreferences{
                Theme:    "dark",
                Language: "en",
                Timezone: "UTC",
            },
        },
    }
    
    // Insert
    db.Exec(
        `INSERT INTO users (id, preferences) VALUES ($1, $2)`,
        user.ID,
        user.Preferences,
    )
    
    // Query
    var loadedUser User
    db.Get(&loadedUser, `SELECT * FROM users WHERE id = $1`, "user123")
    
    // Type-safe access (no casting needed)
    theme := loadedUser.Preferences.Value.Theme // "dark"
}
```

### 4. Nested JSONB Structures

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type Contact struct {
    ID        string
    Addresses jsonb.JSON[[]Address] // Array of addresses
}

func main() {
    contact := Contact{
        ID: "contact123",
        Addresses: jsonb.JSON[[]Address]{
            Value: []Address{
                {Street: "123 Main St", City: "Kyiv", Country: "UA"},
                {Street: "456 Oak Ave", City: "Lviv", Country: "UA"},
            },
        },
    }
    
    db.Exec(
        `INSERT INTO contacts (id, addresses) VALUES ($1, $2)`,
        contact.ID,
        contact.Addresses,
    )
}
```

### 5. Querying JSONB Fields

```go
// PostgreSQL JSONB operators
func findUsersByTheme(db *sqlx.DB, theme string) ([]User, error) {
    var users []User
    
    // -> operator extracts JSON field
    err := db.Select(&users, `
        SELECT * FROM users 
        WHERE preferences->>'theme' = $1
    `, theme)
    
    return users, err
}

func findProductsByTag(db *sqlx.DB, tag string) ([]Product, error) {
    var products []Product
    
    // @> operator checks JSONB containment
    err := db.Select(&products, `
        SELECT * FROM products 
        WHERE tags @> $1::jsonb
    `, fmt.Sprintf(`["%s"]`, tag))
    
    return products, err
}
```

### 6. Updating JSONB Fields

```go
// Update entire JSONB column
func updateMetadata(db *sqlx.DB, userID string, metadata jsonb.Map) error {
    _, err := db.Exec(`
        UPDATE users 
        SET metadata = $1, updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
    `, metadata, userID)
    return err
}

// Update single JSONB field (PostgreSQL 9.5+)
func updateMetadataField(db *sqlx.DB, userID, key, value string) error {
    _, err := db.Exec(`
        UPDATE users 
        SET metadata = jsonb_set(metadata, $1, $2::jsonb)
        WHERE id = $3
    `, fmt.Sprintf(`{%s}`, key), fmt.Sprintf(`"%s"`, value), userID)
    return err
}
```

---

## PostgreSQL JSONB Operations

### Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `->` | Get JSON field as JSON | `metadata->'theme'` |
| `->>` | Get JSON field as text | `metadata->>'theme'` |
| `@>` | Contains JSON | `tags @> '["smartphone"]'` |
| `?` | Has key | `metadata ? 'theme'` |
| `?\|` | Has any key | `metadata ?\| array['theme', 'lang']` |
| `?&` | Has all keys | `metadata ?& array['theme', 'lang']` |

### Functions

```sql
-- Extract value
SELECT metadata->>'theme' FROM users;

-- Check existence
SELECT * FROM users WHERE metadata ? 'last_login';

-- Update field
UPDATE users SET metadata = jsonb_set(metadata, '{theme}', '"light"');

-- Remove field
UPDATE users SET metadata = metadata - 'temp_data';

-- Merge JSONB
UPDATE users SET metadata = metadata || '{"new_key": "value"}'::jsonb;
```

---

## API Reference

### Map Type

```go
type Map map[string]interface{}

// Scan implements sql.Scanner
func (m *Map) Scan(value interface{}) error

// Value implements driver.Valuer
func (m Map) Value() (driver.Value, error)
```

### Array Type

```go
type Array []interface{}

// Scan implements sql.Scanner
func (a *Array) Scan(value interface{}) error

// Value implements driver.Valuer
func (a Array) Value() (driver.Value, error)
```

### JSON[T] Type

```go
type JSON[T any] struct {
    Value T
}

// Scan implements sql.Scanner
func (j *JSON[T]) Scan(value interface{}) error

// Value implements driver.Valuer
func (j JSON[T]) Value() (driver.Value, error)
```

---

## Best Practices

### DO

- **Use Map for unstructured data** - Flexible metadata
  ```go
  type User struct {
      Metadata jsonb.Map // Flexible key-value pairs
  }
  ```

- **Use JSON[T] for structured data** - Type safety
  ```go
  type User struct {
      Preferences jsonb.JSON[UserPreferences] // Type-safe struct
  }
  ```

- **Index JSONB columns** - Improve query performance
  ```sql
  -- GIN index for existence checks
  CREATE INDEX idx_users_metadata ON users USING GIN (metadata);
  
  -- Partial index for specific field
  CREATE INDEX idx_users_theme ON users ((metadata->>'theme'));
  ```

- **Validate JSONB data** - Use PostgreSQL constraints
  ```sql
  ALTER TABLE users 
  ADD CONSTRAINT check_metadata_has_version 
  CHECK (metadata ? 'version');
  ```

### DON'T

- **Don't store large arrays** - Prefer separate tables
  ```go
  // Bad: Large array in JSONB
  type User struct {
      Orders jsonb.Array // 1000s of orders
  }
  
  // Good: Separate table
  type Order struct {
      UserID string // Foreign key
  }
  ```

- **Don't use JSONB for relational data** - Use proper columns
  ```sql
  -- Bad: Relational data in JSONB
  CREATE TABLE users (
      data JSONB -- {"email": "...", "name": "..."}
  );
  
  -- Good: Proper columns
  CREATE TABLE users (
      email VARCHAR(255),
      name  VARCHAR(255),
      metadata JSONB -- Only flexible data here
  );
  ```

- **Don't forget nil checks** - JSONB can be NULL
  ```go
  var user User
  db.Get(&user, `SELECT * FROM users WHERE id = $1`, id)
  
  // Check for nil
  if user.Metadata != nil {
      theme := user.Metadata["theme"]
  }
  ```

---

## Database Schema

### Table Creation

```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY,
    email       VARCHAR(255) NOT NULL,
    metadata    JSONB NULL,
    preferences JSONB NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for JSONB queries
CREATE INDEX idx_users_metadata_gin ON users USING GIN (metadata);

-- Index for specific field
CREATE INDEX idx_users_theme ON users ((metadata->>'theme'));

-- Partial index
CREATE INDEX idx_users_verified ON users ((metadata->>'is_verified'))
WHERE (metadata->>'is_verified')::boolean = true;
```

### Constraints

```sql
-- Ensure JSONB is valid JSON
ALTER TABLE users 
ADD CONSTRAINT check_metadata_valid 
CHECK (metadata IS NULL OR jsonb_typeof(metadata) = 'object');

-- Require specific keys
ALTER TABLE users
ADD CONSTRAINT check_preferences_has_theme
CHECK (preferences IS NULL OR preferences ? 'theme');
```

---

## Testing

```go
package jsonb_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/basilex/promenade/pkg/jsonb"
)

func TestMap_Scan(t *testing.T) {
    var m jsonb.Map
    
    // Scan from JSON bytes
    data := []byte(`{"key":"value"}`)
    err := m.Scan(data)
    
    assert.NoError(t, err)
    assert.Equal(t, "value", m["key"])
}

func TestMap_Value(t *testing.T) {
    m := jsonb.Map{"key": "value"}
    
    // Convert to driver.Value
    value, err := m.Value()
    
    assert.NoError(t, err)
    assert.Equal(t, []byte(`{"key":"value"}`), value)
}

func TestJSON_TypeSafe(t *testing.T) {
    type Config struct {
        Theme string `json:"theme"`
    }
    
    j := jsonb.JSON[Config]{
        Value: Config{Theme: "dark"},
    }
    
    value, err := j.Value()
    assert.NoError(t, err)
    assert.Contains(t, string(value.([]byte)), `"theme":"dark"`)
}
```

---

## Related Packages

- `internal/infrastructure/database` - Database connection management
- `github.com/jmoiron/sqlx` - SQL extension library
- `encoding/json` - Standard JSON marshaling

---

**Last Updated:** 2025-12-28  
**Status:** Production-ready  
**Maintainer:** Promenade Team