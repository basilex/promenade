# JSON Store Package

**Type-safe JSON storage** for Promenade Platform - optimized for PostgreSQL 14+ JSONB.

---

## Overview

The `jsonstore` package provides a **generic `Field[T]` type** that seamlessly handles JSON serialization/deserialization. It implements `sql.Scanner` and `driver.Valuer` for transparent database integration.

---

## Features

- **Type-Safe**: Generic `Field[T]` for any JSON-serializable type
- **PostgreSQL Optimized**: Leverages JSONB for fast queries and indexing
- **Automatic Marshaling**: JSON encoding/decoding handled automatically
- **NULL Support**: Proper handling of NULL database values
- **Deep Copy**: `Clone()` method for safe copies
- **Zero Dependencies**: Uses only standard library `encoding/json`

---

## Supported Types

Any type that can be marshaled to JSON:

- `[]string`, `[]int`, `[]uuid.UUID` - Arrays
- `map[string]any`, `map[string]string` - Objects
- Custom structs with JSON tags
- Nested combinations

---

## Quick Start

### 1. Define Entity

```go
import "github.com/basilex/promenade/pkg/jsonstore"

type Customer struct {
    ID       uuid.UUID               `db:"id"`
    Name     string                  `db:"name"`
    Tags     jsonstore.Field[[]string]       `db:"tags"`     // String array
    Metadata jsonstore.Field[map[string]any] `db:"metadata"` // JSON object
}
```

### 2. Create Entity

```go
func NewCustomer(name string) *Customer {
    return &Customer{
        ID:       uuidv7.New(),
        Name:     name,
        Tags:     jsonstore.NewField([]string{}),           // Empty array
        Metadata: jsonstore.NewField(map[string]any{}),     // Empty object
    }
}
```

### 3. Manipulate JSON Data

```go
// Add tags
customer.Tags.Set([]string{"vip", "premium"})

// Get tags
tags := customer.Tags.Get() // []string{"vip", "premium"}

// Add metadata
meta := customer.Metadata.Get()
meta["score"] = 95
meta["tier"] = "gold"
customer.Metadata.Set(meta)
```

### 4. Database Operations

```go
// INSERT - Field automatically marshals to JSON
query := "INSERT INTO customers (id, name, tags, metadata) VALUES ($1, $2, $3, $4)"
_, err := db.Exec(query, customer.ID, customer.Name, customer.Tags, customer.Metadata)

// SELECT - Field automatically unmarshals from JSON
query := "SELECT id, name, tags, metadata FROM customers WHERE id = $1"
var customer Customer
err := db.Get(&customer, query, id)

// Tags are ready to use
fmt.Println(customer.Tags.Get()) // ["vip", "premium"]
```

---

## Database Compatibility

### PostgreSQL (JSONB)

```sql
CREATE TABLE customers (
    id UUID PRIMARY KEY,
    tags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}'
);

-- Optimized queries with GIN index
CREATE INDEX idx_tags ON customers USING gin(tags);

-- Query with @> operator
SELECT * FROM customers WHERE tags @> '["vip"]'::jsonb;
```

```go
// PostgreSQL driver returns []byte
// Field.Scan() handles it automatically
var customer Customer
db.Get(&customer, "SELECT * FROM customers WHERE id = $1", id)
```

---

## API Reference

### Field[T]

Generic type for JSON-storable fields.

#### Constructor Functions

```go
// NewField creates a field with a value
field := jsonstore.NewField([]string{"vip"})

// NewNullField creates a NULL field
field := jsonstore.NewNullField[[]string]()
```

#### Methods

```go
// Get returns the stored value
tags := field.Get() // []string

// Set updates the stored value
field.Set([]string{"premium", "vip"})

// IsNull checks if field is NULL
if field.IsNull() {
    // Handle NULL case
}

// SetNull marks field as NULL
field.SetNull()

// String returns human-readable representation
fmt.Println(field.String()) // ["premium","vip"]

// Equal compares two fields
if field1.Equal(field2) {
    // Fields are equal
}

// Clone creates a deep copy
clone := field.Clone()
```

#### Interfaces

```go
// sql.Scanner - reads from database
func (f *Field[T]) Scan(value interface{}) error

// driver.Valuer - writes to database
func (f Field[T]) Value() (driver.Value, error)

// json.Marshaler - JSON encoding
func (f Field[T]) MarshalJSON() ([]byte, error)

// json.Unmarshaler - JSON decoding
func (f *Field[T]) UnmarshalJSON(data []byte) error
```

---

## Usage Examples

### String Array (Tags)

```go
type Customer struct {
    Tags jsonstore.Field[[]string] `db:"tags"`
}

customer := &Customer{
    Tags: jsonstore.NewField([]string{"vip", "premium"}),
}

// Add tag
tags := customer.Tags.Get()
tags = append(tags, "loyal")
customer.Tags.Set(tags)

// Check tag
func (c *Customer) HasTag(tag string) bool {
    for _, t := range c.Tags.Get() {
        if t == tag {
            return true
        }
    }
    return false
}
```

### UUID Array (Attendees)

```go
type Interaction struct {
    Attendees jsonstore.Field[[]uuid.UUID] `db:"attendees"`
}

interaction := &Interaction{
    Attendees: jsonstore.NewField([]uuid.UUID{}),
}

// Add attendee
attendees := interaction.Attendees.Get()
attendees = append(attendees, userID)
interaction.Attendees.Set(attendees)

// Count attendees
count := len(interaction.Attendees.Get())
```

### Map (Metadata)

```go
type Customer struct {
    Metadata jsonstore.Field[map[string]interface{}] `db:"metadata"`
}

customer := &Customer{
    Metadata: jsonstore.NewField(map[string]interface{}{
        "score": 95,
        "tier": "gold",
        "joined": "2026-01-03",
    }),
}

// Get value
score := customer.Metadata.Get()["score"].(int)

// Update value
meta := customer.Metadata.Get()
meta["score"] = 100
customer.Metadata.Set(meta)
```

### Custom Struct

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type Customer struct {
    Address jsonstore.Field[Address] `db:"address"`
}

customer := &Customer{
    Address: jsonstore.NewField(Address{
        Street:  "123 Main St",
        City:    "Kyiv",
        Country: "Ukraine",
    }),
}

// Access nested fields
city := customer.Address.Get().City
```

### Nested JSON

```go
type Settings struct {
    Notifications map[string]bool `json:"notifications"`
    Theme         string          `json:"theme"`
}

type User struct {
    Settings jsonstore.Field[Settings] `db:"settings"`
}

user := &User{
    Settings: jsonstore.NewField(Settings{
        Notifications: map[string]bool{
            "email": true,
            "sms":   false,
        },
        Theme: "dark",
    }),
}
```

---

## Migration from JSONB

### Before (PostgreSQL-specific)

```go
import "github.com/basilex/promenade/pkg/jsonb"

type Customer struct {
    Tags jsonb.JSON[[]string] `db:"tags"` // Only works with Postgres
}

// Manual parsing in repository
var tagsJSON []byte
db.Get(&tagsJSON, "SELECT tags FROM customers WHERE id = $1", id)
var tags []string
json.Unmarshal(tagsJSON, &tags)
```

### After (Database-agnostic)

```go
import "github.com/basilex/promenade/pkg/jsonstore"

type Customer struct {
    Tags jsonstore.Field[[]string] `db:"tags"` // Works with any DB
}

// Automatic parsing
var customer Customer
db.Get(&customer, "SELECT * FROM customers WHERE id = $1", id)
tags := customer.Tags.Get() // Ready to use
```

---

## NULL Handling

### Database NULL Values

```go
// NULL in database becomes NULL field
var customer Customer
db.Get(&customer, "SELECT * FROM customers WHERE id = $1", id)

if customer.Tags.IsNull() {
    // Tags column was NULL in database
    customer.Tags.Set([]string{}) // Initialize
}
```

### Setting NULL

```go
// Mark field as NULL (will be stored as NULL in database)
customer.Tags.SetNull()

// Or create NULL field explicitly
customer.Tags = jsonstore.NewNullField[[]string]()
```

### INSERT with NULL

```go
customer := &Customer{
    Name: "Test",
    Tags: jsonstore.NewNullField[[]string](), // NULL
}

db.Exec("INSERT INTO customers (name, tags) VALUES ($1, $2)",
    customer.Name, customer.Tags) // tags = NULL
```

---

## Performance Considerations

### PostgreSQL (Native JSONB)

- **Fast**: GIN indexes, native operators (@>, ->, ->>)
- **Storage**: Compressed binary format
- **Queries**: Index-optimized searches

```sql
-- GIN index for array searches
CREATE INDEX idx_tags ON customers USING gin(tags);

-- Fast query with index
SELECT * FROM customers WHERE tags @> '["vip"]'::jsonb;
```

---

## Testing

Repository-wide testing strategy and baseline budgets are documented in [docs/guides/testing-patterns.md](../../docs/guides/testing-patterns.md).

```bash
# Run tests
go test ./pkg/jsonstore -v
```

---

## Best Practices

### DO

- Use `Field[T]` for all JSON columns in entities
- Initialize with `NewField()` in constructors
- Check `IsNull()` before accessing values from database
- Use PostgreSQL JSONB for production (fast, indexed)
- Leverage GIN indexes for array/object queries

### DON'T

- Don't access `.value` directly (use `Get()` method)
- Don't mutate returned values without calling `Set()`
- Don't store large JSON documents (> 1MB)
- Don't rely on JSON queries for high-frequency operations
- Don't use for frequently-queried nested fields (normalize instead)

---

## Troubleshooting

### Error: "cannot scan type"

**Cause**: Database driver returned unexpected type

**Solution**: Check database column type matches dialect

```go
// PostgreSQL: JSONB (returns []byte)
// SQLite: TEXT (returns string)
// Both work with Field.Scan()
```

### Error: "failed to unmarshal JSON"

**Cause**: Invalid JSON in database

**Solution**: Add CHECK constraint

```sql
-- PostgreSQL
ALTER TABLE customers ADD CHECK (jsonb_typeof(tags) = 'array');
```

### Performance Issue: Slow Queries

**Solution**:

1. PostgreSQL: Create GIN index on JSONB column
2. Consider denormalizing frequently-queried fields
3. Use materialized views for complex JSON queries

---

## Related Documentation

- [Database Package](../database/README.md) - Dialect abstraction
- [Database Strategy](../../docs/guides/database-strategy.md) - PostgreSQL-first approach
- [Migration Guide](../../migrations/README.md) - Database schema management
- [Testing Guide](../../test/README.md) - Integration testing

---

**Status**: Production Ready  
**Version**: 1.0.0  
**Test Coverage**: TBD  
**Maintainer**: Promenade Team  
**Last Updated**: January 21, 2026
