# Ref Package

**Reference (pointer) utilities** for creating pointers to values in Go.

---

## Purpose

In Go, many struct fields are pointers to allow for `nil` (optional/nullable) values. Creating pointers to literals or variables requires taking their address (`&`), which can only be done on named variables, not literals.

This package provides clean utility functions to create references (pointers) from values.

---

## Why "ref" not "ptr"?

The name `ref` (reference) is more Go-idiomatic than `ptr` (pointer):

- **Go context**: Creating references to values for optional fields
- **Not C/C++**: No pointer arithmetic, manual memory management, or dereferencing complexity
- **Clarity**: "Reference" better describes the Go concept of "pointing to a value"

---

## Usage

### Generic Function

```go
import "github.com/basilex/promenade/pkg/ref"

// Works with any type
name := ref.To("John")       // *string
enabled := ref.To(true)      // *bool
count := ref.To(42)          // *int
```

### Type-Specific Functions

For better type control with numeric literals:

```go
// Avoid type inference issues
population := ref.Int64(331002651)  // *int64 (not *int)
area := ref.Int(9833520)            // *int
latitude := ref.Float64(38.8951)    // *float64
```

### Real-World Example

```go
type Country struct {
    Code       string
    Name       string
    Population *int64    // Optional field
    Latitude   *float64  // Optional field
    AreaKm2    *int      // Optional field
}

// Creating country with optional fields
country := &Country{
    Code:       "US",
    Name:       "United States",
    Population: ref.Int64(331002651),  // Type-safe
    Latitude:   ref.Float64(38.8951),
    AreaKm2:    ref.Int(9833520),
}

// Without ref package (verbose):
var pop int64 = 331002651
var lat float64 = 38.8951
var area int = 9833520

country := &Country{
    Code:       "US",
    Name:       "United States",
    Population: &pop,
    Latitude:   &lat,
    AreaKm2:    &area,
}
```

---

## API

### `To[T any](v T) *T`

Generic function that works with any type. Go infers the type from the argument.

**Example**:
```go
ref.To("hello")    // *string
ref.To(42)         // *int
ref.To(3.14)       // *float64
ref.To(true)       // *bool
```

### `Int64(v int64) *int64`

Returns pointer to `int64`. Useful to avoid type ambiguity with large numeric literals.

**Example**:
```go
ref.Int64(331002651)  // *int64
```

### `Int(v int) *int`

Returns pointer to `int`.

**Example**:
```go
ref.Int(42)  // *int
```

### `Float64(v float64) *float64`

Returns pointer to `float64`.

**Example**:
```go
ref.Float64(38.8951)  // *float64
```

---

## Testing

```bash
# Run tests
go test ./pkg/ref -v

# With coverage
go test ./pkg/ref -cover
```

**Test Coverage**: 6 tests, 100% coverage

---

## When to Use

✅ **Use when**:
- Creating struct literals with optional/nullable fields
- Seeding database with test data
- Working with API models that have optional fields
- Any time you need a pointer to a literal value

❌ **Don't use when**:
- Variable is already a pointer
- You can use a named variable
- Working with non-optional fields

---

## Related Packages

- [pkg/valueobject](../valueobject/README.md) - Domain value objects (Email, Phone, Money, Address)
- [pkg/uuidv7](../uuidv7/README.md) - Time-ordered UUIDs

---

**Version**: 0.1.0  
**Status**: Production-ready  
**Tests**: 6 tests, 100% coverage  
**Maintainer**: Promenade Team
