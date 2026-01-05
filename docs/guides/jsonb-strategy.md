# JSONB Storage Strategy

**Database-Agnostic JSON Storage** for Promenade Platform using the **jsonstore package**.

---

## Overview

Promenade uses **jsonstore.Field[T]** to store JSON data in a database-agnostic way:

- **PostgreSQL**: Stores as `JSONB` (binary format, indexable)
- **SQLite**: Stores as `TEXT` (JSON string)
- **MySQL**: Stores as `JSON` (native type)
- **SQL Server**: Stores as `NVARCHAR(MAX)` with JSON functions

**Key Benefit**: Write Go code once, works with all databases.

---

## Implementation

### The jsonstore.Field[T] Type

**Package**: `pkg/jsonstore`

```go
// Field is a generic wrapper for database-agnostic JSON storage
type Field[T any] struct {
    data T
    null bool
}

// Implements:
// - sql.Scanner (database → Go)
// - driver.Valuer (Go → database)
// - json.Marshaler/Unmarshaler
```

**Key Features**:
- Type-safe: `Field[[]string]`, `Field[map[string]int]`, `Field[[]uuidv7.UUID]`
- NULL handling: `field.IsNull()`, `field.Set(value)`, `field.SetNull()`
- Deep copy: `field.Clone()`
- Equality: `field.Equals(other)`

---

## Usage Patterns

### 1. String Arrays (Tags)

**Use Case**: Customer tags, product categories, labels

**Example**:
```go
// Entity
type Customer struct {
    aggregate.BaseAggregate
    Name  string
    Tags  jsonstore.Field[[]string] `db:"tags" json:"tags"`
}

// Factory method
func NewCustomer(name string) (*Customer, error) {
    tags := jsonstore.NewField([]string{})  // Empty array, not NULL
    return &Customer{
        BaseAggregate: aggregate.NewBaseAggregate(),
        Name:          name,
        Tags:          tags,
    }, nil
}

// Add tag
func (c *Customer) AddTag(tag string) error {
    current := c.Tags.Get()
    for _, t := range current {
        if t == tag {
            return fmt.Errorf("tag already exists: %s", tag)
        }
    }
    c.Tags.Set(append(current, tag))
    c.Touch()
    return nil
}

// Remove tag
func (c *Customer) RemoveTag(tag string) error {
    current := c.Tags.Get()
    filtered := make([]string, 0, len(current))
    found := false
    for _, t := range current {
        if t != tag {
            filtered = append(filtered, t)
        } else {
            found = true
        }
    }
    if !found {
        return fmt.Errorf("tag not found: %s", tag)
    }
    c.Tags.Set(filtered)
    c.Touch()
    return nil
}
```

**Migration**:
```sql
-- PostgreSQL
ALTER TABLE customer_customers ADD COLUMN tags TEXT;

-- SQLite (same DDL)
ALTER TABLE customer_customers ADD COLUMN tags TEXT;
```

**Repository Row Struct**:
```go
type customerRow struct {
    ID   uuidv7.UUID              `db:"id"`
    Name string                   `db:"name"`
    Tags jsonstore.Field[[]string] `db:"tags"`
}

func (r *customerRow) toEntity() (*Customer, error) {
    return &Customer{
        BaseAggregate: aggregate.BaseAggregate{ID: r.ID},
        Name:          r.Name,
        Tags:          r.Tags,  // Direct assignment
    }, nil
}

func fromEntity(c *Customer) *customerRow {
    return &customerRow{
        ID:   c.ID,
        Name: c.Name,
        Tags: c.Tags,  // Direct assignment
    }
}
```

### 2. UUID Arrays (Attendees)

**Use Case**: Meeting attendees, team members, participants

**Example**:
```go
// Entity
type Interaction struct {
    aggregate.BaseAggregate
    Type      InteractionType
    Attendees jsonstore.Field[[]uuidv7.UUID] `db:"attendees" json:"attendees"`
}

// Factory method
func NewMeeting(customerID uuidv7.UUID, attendees []uuidv7.UUID) (*Interaction, error) {
    if len(attendees) == 0 {
        return nil, fmt.Errorf("meeting must have at least one attendee")
    }
    
    attendeesField := jsonstore.NewField(attendees)
    return &Interaction{
        BaseAggregate: aggregate.NewBaseAggregate(),
        Type:          InteractionTypeMeeting,
        Attendees:     attendeesField,
    }, nil
}

// Add attendee
func (i *Interaction) AddAttendee(userID uuidv7.UUID) error {
    current := i.Attendees.Get()
    for _, id := range current {
        if id == userID {
            return fmt.Errorf("attendee already added")
        }
    }
    i.Attendees.Set(append(current, userID))
    i.Touch()
    return nil
}
```

**Migration**:
```sql
-- Both PostgreSQL and SQLite
ALTER TABLE customer_interactions ADD COLUMN attendees TEXT;
```

### 3. Nested Objects (Metadata)

**Use Case**: Flexible metadata, settings, configuration

**Example**:
```go
type OrderMetadata struct {
    Source      string            `json:"source"`       // "web", "mobile", "api"
    Campaign    string            `json:"campaign"`     // Marketing campaign ID
    ReferralID  *string           `json:"referral_id"`  // Optional referral
    CustomData  map[string]string `json:"custom_data"`  // Arbitrary key-value
}

type Order struct {
    aggregate.BaseAggregate
    OrderNumber string
    Metadata    jsonstore.Field[OrderMetadata] `db:"metadata" json:"metadata"`
}
```

**Usage**:
```go
// Create with metadata
metadata := OrderMetadata{
    Source:   "web",
    Campaign: "summer-2026",
    CustomData: map[string]string{
        "user_agent": "Mozilla/5.0...",
        "ip_address": "192.168.1.1",
    },
}
metadataField := jsonstore.NewField(metadata)

order := &Order{
    BaseAggregate: aggregate.NewBaseAggregate(),
    OrderNumber:   "ORD-2026-001",
    Metadata:      metadataField,
}

// Update metadata
meta := order.Metadata.Get()
meta.CustomData["processed_by"] = "worker-123"
order.Metadata.Set(meta)
order.Touch()
```

---

## Database Behavior

### PostgreSQL

**Storage**: `JSONB` (binary format)

**Features**:
- Indexable: `CREATE INDEX idx_customers_tags ON customers USING GIN (tags);`
- Queryable: `SELECT * FROM customers WHERE tags @> '["vip"]';`
- Operators: `@>`, `?`, `?|`, `?&`, `||`, `-`, `#-`

**Performance**: Excellent for complex JSON queries

### SQLite

**Storage**: `TEXT` (JSON string)

**Features**:
- JSON functions: `json_extract()`, `json_array_length()`
- Queryable: `SELECT * FROM customers WHERE json_extract(tags, '$[0]') = 'vip';`
- No native indexing (use computed columns)

**Performance**: Good for simple queries, slow for complex

### Comparison

| Operation           | PostgreSQL JSONB | SQLite TEXT |
| ------------------- | ---------------- | ----------- |
| **Array contains**  | `@> '["vip"]'`   | `json_extract()` |
| **Array length**    | `jsonb_array_length()` | `json_array_length()` |
| **Indexing**        | GIN index        | Computed column |
| **Storage**         | Binary (smaller) | String (larger) |
| **Query Speed**     | Fast             | Moderate    |

---

## NULL vs Empty

**Important Distinction**:

```go
// NULL (no value)
tags := jsonstore.Field[[]string]{}  // null = true
tags.IsNull()  // true
tags.Get()     // nil

// Empty array (value exists, but empty)
tags := jsonstore.NewField([]string{})  // null = false
tags.IsNull()  // false
tags.Get()     // []string{} (length 0)

// Non-empty array
tags := jsonstore.NewField([]string{"vip", "enterprise"})
tags.IsNull()  // false
tags.Get()     // []string{"vip", "enterprise"} (length 2)
```

**Database Storage**:
- NULL: `NULL` in database column
- Empty array: `[]` in database column
- Non-empty: `["vip", "enterprise"]` in database column

**Best Practice**: Use **empty arrays** instead of NULL when possible (simpler logic, no nil checks).

---

## Testing

### Unit Tests

```go
func TestCustomer_AddTag(t *testing.T) {
    customer, _ := NewCustomer("Test Customer")
    
    // Initially empty
    assert.False(t, customer.Tags.IsNull())
    assert.Equal(t, 0, len(customer.Tags.Get()))
    
    // Add tag
    err := customer.AddTag("vip")
    assert.NoError(t, err)
    assert.Equal(t, 1, len(customer.Tags.Get()))
    assert.Equal(t, "vip", customer.Tags.Get()[0])
    
    // Duplicate tag fails
    err = customer.AddTag("vip")
    assert.Error(t, err)
}
```

### Integration Tests

```go
func TestCustomerRepository_Tags(t *testing.T) {
    db, cleanup := integration.SetupTestDB(t)
    defer cleanup()
    
    repo := postgres.NewCustomerRepository(db)
    ctx := context.Background()
    
    // Create with tags
    customer, _ := NewCustomer("Test Customer")
    customer.AddTag("vip")
    customer.AddTag("enterprise")
    
    err := repo.Create(ctx, customer)
    assert.NoError(t, err)
    
    // Retrieve and verify
    retrieved, err := repo.GetByID(ctx, customer.ID)
    assert.NoError(t, err)
    assert.Equal(t, 2, len(retrieved.Tags.Get()))
    assert.Contains(t, retrieved.Tags.Get(), "vip")
    assert.Contains(t, retrieved.Tags.Get(), "enterprise")
}
```

---

## Best Practices

### DO's 

1. **Use specific types**: `Field[[]string]` not `Field[interface{}]`
2. **Initialize with empty values**: `jsonstore.NewField([]string{})` not NULL
3. **Validate before Set()**: Check business rules before updating
4. **Clone for modifications**: `tags := field.Clone(); modify(tags); field.Set(tags)`
5. **Check IsNull()**: Handle NULL gracefully in business logic

### DON'Ts 

1. **Don't use NULL for collections**: Use empty arrays instead
2. **Don't modify Get() result directly**: Clone first (Get() returns pointer)
3. **Don't store large objects**: >1MB JSON impacts performance
4. **Don't use for structured data**: Use proper database tables for relational data
5. **Don't forget Touch()**: Update timestamps after Set()

---

## Performance Considerations

### When to Use JSONB

 **Good Use Cases**:
- Arrays of primitives (tags, IDs, categories)
- Flexible metadata (settings, config)
- Sparse data (optional fields)
- Audit logs (history, changes)

 **Bad Use Cases**:
- Large objects (>1MB)
- Frequently queried relational data
- Data with strict schema
- High-write throughput columns

### Optimization Tips

1. **Index frequently queried fields** (PostgreSQL only):
   ```sql
   CREATE INDEX idx_customers_tags ON customers USING GIN (tags);
   ```

2. **Limit JSON size**: Keep under 100KB for best performance

3. **Use computed columns** (SQLite):
   ```sql
   ALTER TABLE customers ADD COLUMN tag_count INTEGER AS (json_array_length(tags));
   CREATE INDEX idx_customers_tag_count ON customers(tag_count);
   ```

4. **Benchmark queries**: Use `EXPLAIN ANALYZE` to verify index usage

---

## Migration Guide

### From JSONB to TEXT (Phase 4 Complete)

**Before** (PostgreSQL-specific):
```sql
CREATE TABLE customers (
    tags JSONB DEFAULT '[]'::JSONB
);
```

**After** (database-agnostic):
```sql
CREATE TABLE customers (
    tags TEXT  -- Works in PostgreSQL, SQLite, MySQL
);
```

**Go Code**: No changes needed! `jsonstore.Field[T]` handles both.

### Adding New JSONB Field

1. **Migration**:
   ```sql
   ALTER TABLE customers ADD COLUMN settings TEXT;
   ```

2. **Entity**:
   ```go
   type CustomerSettings struct {
       EmailOptIn bool   `json:"email_opt_in"`
       Theme      string `json:"theme"`
   }
   
   type Customer struct {
       Settings jsonstore.Field[CustomerSettings] `db:"settings"`
   }
   ```

3. **Factory**:
   ```go
   func NewCustomer(name string) (*Customer, error) {
       defaultSettings := CustomerSettings{
           EmailOptIn: true,
           Theme:      "light",
       }
       return &Customer{
           Settings: jsonstore.NewField(defaultSettings),
       }, nil
   }
   ```

---

## Troubleshooting

### Issue: NULL Pointer Panic

**Symptom**: `panic: runtime error: invalid memory address`

**Cause**: Calling `.Get()` on NULL field without checking

**Fix**:
```go
if !customer.Tags.IsNull() {
    tags := customer.Tags.Get()
    // ... use tags
}

// OR initialize with empty
tags := jsonstore.NewField([]string{})
```

### Issue: JSON Marshal Error

**Symptom**: `json: unsupported type: func()`

**Cause**: Trying to store non-JSON-serializable types

**Fix**: Use only JSON-serializable types (primitives, structs, arrays, maps)

### Issue: Database Type Mismatch

**Symptom**: `ERROR: column "tags" is of type jsonb but expression is of type text`

**Cause**: Migration used JSONB instead of TEXT

**Fix**: Change migration to use TEXT:
```sql
ALTER TABLE customers ALTER COLUMN tags TYPE TEXT;
```

---

## Related Documentation

- [Database Adapters Guide](database-adapters.md)
- [Database Conventions](database-conventions.md)
- [Testing Guide](../../test/README.md)
- [jsonstore Package](../../pkg/jsonstore/README.md)

---

**Last Updated**: January 3, 2026  
**Status**: Production-Ready  
**Test Coverage**: 100% (pkg/jsonstore)  
**In Production**: Customer.tags, Interaction.attendees  
**Maintainer**: Promenade Team
