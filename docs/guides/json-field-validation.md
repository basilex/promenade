# JSON Field Validation Strategy

**Status**: Documentation (Implementation Planned - Phase 2)  
**Priority**: MEDIUM  
**Last Updated**: January 22, 2026

---

## Overview

Strategy for validating JSON fields stored in `jsonstore.Field[T]` to prevent abuse, ensure data quality, and maintain database performance.

---

## Current State (Phase 1)

###  Implemented

**Basic Validation**:

- JSON syntax validation (via `json.Unmarshal`)
- Type safety (via Go generics)
- NULL handling
- Empty value handling

**Example**:

```go
type Customer struct {
    Tags jsonstore.Field[[]string]
}

//  Valid - proper JSON array
customer.Tags.Set([]string{"vip", "premium"})

//  Invalid - runtime panic (type mismatch)
// customer.Tags.Set(123) // Won't compile (type safety)
```

###  Missing

**Size Limits**:

- No maximum JSON size enforcement
- No array length limits
- No object depth limits
- No string length limits

**Business Rules**:

- No semantic validation (e.g., valid tag names)
- No range validation for numbers
- No format validation for strings

**Security**:

- No protection against deeply nested objects (DoS)
- No protection against extremely large arrays
- No protection against malicious JSON bombs

---

## Risks Without Validation

### 1. Database Performance

**Problem**: Unbounded JSON size

```go
// Attacker submits 10MB JSON field
customer.Metadata.Set(map[string]interface{}{
    "data": strings.Repeat("x", 10*1024*1024), // 10MB string
})
```

**Impact**:

- Slow database writes
- Large table size
- Index bloat (JSONB is indexed)
- Memory pressure

### 2. Denial of Service

**Problem**: Deeply nested JSON

```json
{
  "a": {
    "b": {
      "c": {
        "d": {
          "e": {
            "f": {
              // ... 1000 levels deep
            }
          }
        }
      }
    }
  }
}
```

**Impact**:

- Stack overflow in JSON parser
- CPU exhaustion
- Application crash

### 3. Storage Cost

**Problem**: Unbounded arrays

```go
// Attacker creates array with 1 million elements
customer.Tags.Set(make([]string, 1000000))
```

**Impact**:

- Excessive storage costs
- Slow queries
- Memory exhaustion

---

## Proposed Validation Rules

### Size Limits

| Field Type              | Max Size | Max Array Length | Max Object Depth | Rationale               |
| ----------------------- | -------- | ---------------- | ---------------- | ----------------------- |
| `Field[[]string]`       | 64 KB    | 1000             | N/A              | Typical tag/label usage |
| `Field[map[string]any]` | 256 KB   | N/A              | 10 levels        | Complex metadata        |
| `Field[[]T]` (general)  | 128 KB   | 5000             | 5 levels         | Generic arrays          |
| `Field[CustomStruct]`   | 512 KB   | N/A              | 10 levels        | Business objects        |

**Defaults** (if not specified):

- Max JSON size: **1 MB** (prevent abuse)
- Max array length: **10,000** elements
- Max object depth: **20** levels
- Max string length: **64 KB**

### Implementation Options

#### Option 1: Validate in `Set()` Method

```go
// pkg/jsonstore/validation.go
type ValidationConfig struct {
    MaxSize       int // Max JSON bytes
    MaxArrayLen   int // Max array length
    MaxDepth      int // Max nesting depth
    MaxStringLen  int // Max string length
}

var DefaultConfig = ValidationConfig{
    MaxSize:      1 * 1024 * 1024, // 1 MB
    MaxArrayLen:  10000,
    MaxDepth:     20,
    MaxStringLen: 65536, // 64 KB
}

// Field with validation
type Field[T any] struct {
    value  T
    null   bool
    config ValidationConfig // Optional: per-field config
}

func (f *Field[T]) Set(value T) error {
    // Serialize to check size
    data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("jsonstore: invalid JSON: %w", err)
    }

    // Check size limit
    if len(data) > f.config.MaxSize {
        return fmt.Errorf("jsonstore: JSON size %d exceeds limit %d", len(data), f.config.MaxSize)
    }

    // Check depth (if map or slice)
    depth := calculateDepth(data)
    if depth > f.config.MaxDepth {
        return fmt.Errorf("jsonstore: JSON depth %d exceeds limit %d", depth, f.config.MaxDepth)
    }

    // Check array length (if slice)
    if err := f.validateArrayLength(value); err != nil {
        return err
    }

    f.value = value
    f.null = false
    return nil
}

func calculateDepth(data []byte) int {
    var obj interface{}
    json.Unmarshal(data, &obj)
    return countDepth(obj, 0)
}

func countDepth(v interface{}, current int) int {
    switch val := v.(type) {
    case map[string]interface{}:
        max := current
        for _, child := range val {
            depth := countDepth(child, current+1)
            if depth > max {
                max = depth
            }
        }
        return max
    case []interface{}:
        max := current
        for _, child := range val {
            depth := countDepth(child, current+1)
            if depth > max {
                max = depth
            }
        }
        return max
    default:
        return current
    }
}

func (f *Field[T]) validateArrayLength(value T) error {
    // Use reflection to check if T is a slice
    rv := reflect.ValueOf(value)
    if rv.Kind() == reflect.Slice {
        if rv.Len() > f.config.MaxArrayLen {
            return fmt.Errorf("jsonstore: array length %d exceeds limit %d", rv.Len(), f.config.MaxArrayLen)
        }
    }
    return nil
}
```

**Pros**:

- Fails fast (before database write)
- Clear error messages
- Configurable per field

**Cons**:

- Breaking change (Set() now returns error)
- Performance overhead on every Set()
- Requires updating all usage sites

#### Option 2: Validate in `Scan()` Method

```go
func (f *Field[T]) Scan(value interface{}) error {
    if value == nil {
        f.SetNull()
        return nil
    }

    var bytes []byte
    switch v := value.(type) {
    case []byte:
        bytes = v
    case string:
        bytes = []byte(v)
    default:
        return fmt.Errorf("jsonstore: cannot scan type %T", value)
    }

    // Validate size BEFORE unmarshaling
    if len(bytes) > DefaultConfig.MaxSize {
        return fmt.Errorf("jsonstore: JSON size %d exceeds limit", len(bytes))
    }

    // Validate depth
    depth := calculateDepth(bytes)
    if depth > DefaultConfig.MaxDepth {
        return fmt.Errorf("jsonstore: JSON depth %d exceeds limit", depth)
    }

    // Unmarshal
    f.null = false
    if err := json.Unmarshal(bytes, &f.value); err != nil {
        return fmt.Errorf("jsonstore: failed to unmarshal: %w", err)
    }

    return nil
}
```

**Pros**:

- No breaking changes
- Protects against malicious database data
- Validates on read (defense in depth)

**Cons**:

- Doesn't prevent invalid data from being written
- Performance overhead on every read
- Errors occur late (after database query)

#### Option 3: Database-Level Constraints

```sql
-- PostgreSQL check constraint
ALTER TABLE customers
ADD CONSTRAINT check_tags_size
CHECK (pg_column_size(tags) < 65536); -- 64 KB

-- Prevent deep nesting (harder, requires custom function)
CREATE OR REPLACE FUNCTION jsonb_depth(j jsonb) RETURNS int AS $$
    SELECT MAX(depth) FROM (
        WITH RECURSIVE tree AS (
            SELECT 1 AS depth, j AS value
            UNION ALL
            SELECT tree.depth + 1, child.value
            FROM tree,
                 jsonb_each(tree.value) AS child
            WHERE jsonb_typeof(tree.value) = 'object'
        )
        SELECT depth FROM tree
    ) depths;
$$ LANGUAGE sql IMMUTABLE;

ALTER TABLE customers
ADD CONSTRAINT check_tags_depth
CHECK (jsonb_depth(tags::jsonb) <= 10);
```

**Pros**:

- Enforced at database level (strongest guarantee)
- Works regardless of client
- No application code changes

**Cons**:

- Database-specific (PostgreSQL)
- Hard to provide good error messages
- Performance impact on writes
- Migrations required

---

## Recommended Approach (Phase 2)

### Hybrid Strategy

1. **Application-level validation** (Set() method) - Phase 2A
2. **Database constraints** (size only) - Phase 2B
3. **Monitoring/alerting** (track large JSON fields) - Phase 2C

### Migration Plan

#### Phase 2A: Add Validation to Set()

**Breaking Change**: Make `Set()` return error

```go
// Before (current)
customer.Tags.Set([]string{"vip"})

// After (Phase 2A)
err := customer.Tags.Set([]string{"vip"})
if err != nil {
    return fmt.Errorf("invalid tags: %w", err)
}
```

**Migration Steps**:

1. Add `SetUnchecked()` method (backward compatibility)
2. Update all usage sites to handle errors
3. Deprecate `SetUnchecked()` in 6 months
4. Remove `SetUnchecked()` in 12 months

#### Phase 2B: Add Database Constraints

```bash
# Add migration
scripts/create-migration.sh core add_jsonb_size_constraints

# Migration content
ALTER TABLE customers
ADD CONSTRAINT check_tags_size CHECK (pg_column_size(tags) < 65536);

ALTER TABLE customers
ADD CONSTRAINT check_metadata_size CHECK (pg_column_size(metadata) < 262144);
```

#### Phase 2C: Add Monitoring

```go
// pkg/metrics/jsonstore.go
var jsonFieldSizeHistogram = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "promenade_jsonstore_field_size_bytes",
        Help:    "Size of JSON fields in bytes",
        Buckets: []float64{1024, 4096, 16384, 65536, 262144, 1048576}, // 1KB to 1MB
    },
    []string{"table", "field"},
)

func (f *Field[T]) Set(value T) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    // Record metric
    jsonFieldSizeHistogram.WithLabelValues(f.tableName, f.fieldName).Observe(float64(len(data)))

    // Validate and set...
}
```

**Alerts**:

```yaml
# Grafana alert
- alert: LargeJSONFields
  expr: histogram_quantile(0.95, promenade_jsonstore_field_size_bytes) > 524288 # 512 KB
  for: 10m
  annotations:
    summary: "95th percentile JSON field size exceeds 512 KB"
```

---

## Validation Errors

### Error Types

```go
// pkg/jsonstore/errors.go
var (
    ErrSizeLimitExceeded   = errors.New("jsonstore: size limit exceeded")
    ErrDepthLimitExceeded  = errors.New("jsonstore: depth limit exceeded")
    ErrArrayLimitExceeded  = errors.New("jsonstore: array length limit exceeded")
    ErrStringLimitExceeded = errors.New("jsonstore: string length limit exceeded")
)

type ValidationError struct {
    Field    string
    Limit    int
    Actual   int
    ErrType  error
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("jsonstore validation failed for %s: limit=%d, actual=%d: %v",
        e.Field, e.Limit, e.Actual, e.ErrType)
}
```

### HTTP Response

```go
// Handler error handling
err := customer.Tags.Set(tags)
if errors.Is(err, jsonstore.ErrSizeLimitExceeded) {
    response.BadRequest(c, "Tags field too large (max 64 KB)")
    return
}
if errors.Is(err, jsonstore.ErrArrayLimitExceeded) {
    response.BadRequest(c, "Too many tags (max 1000)")
    return
}
```

---

## Testing Strategy

### Unit Tests

```go
func TestField_Set_SizeLimit(t *testing.T) {
    f := jsonstore.NewField([]string{})

    // Create large array
    large := make([]string, 100000)
    for i := range large {
        large[i] = "tag" + strconv.Itoa(i)
    }

    err := f.Set(large)
    assert.Error(t, err)
    assert.True(t, errors.Is(err, jsonstore.ErrSizeLimitExceeded))
}

func TestField_Set_DepthLimit(t *testing.T) {
    // Create deeply nested object
    deep := map[string]interface{}{"level0": nil}
    current := deep
    for i := 1; i < 25; i++ {
        current["level"+strconv.Itoa(i)] = map[string]interface{}{}
        current = current["level"+strconv.Itoa(i)].(map[string]interface{})
    }

    f := jsonstore.NewField(map[string]interface{}{})
    err := f.Set(deep)
    assert.Error(t, err)
    assert.True(t, errors.Is(err, jsonstore.ErrDepthLimitExceeded))
}
```

### Integration Tests

```go
func TestCustomer_TagsSizeLimit(t *testing.T) {
    repo := setupTestRepo(t)

    customer := aggregate.NewCustomer("test@example.com")
    customer.Tags.Set(make([]string, 10000)) // At limit

    err := repo.Create(context.Background(), customer)
    assert.NoError(t, err) // Should succeed

    customer2 := aggregate.NewCustomer("test2@example.com")
    customer2.Tags.Set(make([]string, 10001)) // Over limit

    err = repo.Create(context.Background(), customer2)
    assert.Error(t, err) // Should fail
}
```

---

## Configuration

### Per-Environment Limits

```yaml
# config/app.postgres-dev.yaml
jsonstore:
  max_size: 1048576      # 1 MB (generous for dev)
  max_array_length: 10000
  max_depth: 20

# config/app.postgres-prod.yaml
jsonstore:
  max_size: 262144       # 256 KB (stricter for prod)
  max_array_length: 1000
  max_depth: 10
```

### Per-Field Override

```go
type Customer struct {
    // Small tags array
    Tags jsonstore.Field[[]string] `jsonstore:"max_size:65536,max_array:1000"`

    // Large metadata object
    Metadata jsonstore.Field[map[string]interface{}] `jsonstore:"max_size:524288,max_depth:10"`
}
```

---

## Related Documentation

- [jsonstore Package](../../pkg/jsonstore/README.md) - Current implementation
- [Security Patterns](security-patterns.md) - Input validation
- [Database Strategy](database-strategy.md) - JSONB usage
- [Performance Tuning](database-query-patterns.md) - JSON query optimization

---

**Status**: Documentation (Phase 2 implementation planned)  
**Target Date**: Q2 2026  
**Owner**: Platform Team  
**Maintainer**: Promenade Team
