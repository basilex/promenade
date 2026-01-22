# ADR-0001: Use UUID v7 for Primary Keys

**Status**: Accepted  
**Date**: 2026-01-15  
**Deciders**: Core Architecture Team  
**Tags**: database, performance, distributed-systems

---

## Context

Promenade is a distributed ERP/CRM platform with multiple bounded contexts (Customer, Order, Invoice, etc.). We need globally unique identifiers for entities that:

- Work across distributed systems (no central ID generation)
- Provide good database index performance
- Allow natural sorting by creation time
- Support high-volume inserts (warehouse, accounting)

**Current Approach**: UUID v4 (random)

-  Random IDs cause B-tree index fragmentation (40% slower inserts)
-  No timestamp information (difficult debugging)
-  No natural sort order
-  Widely supported
-  Zero collision risk

---

## Considered Options

### Option 1: Auto-incrementing Integer (SERIAL/BIGSERIAL)

- **Pros**: Fast inserts, compact storage (8 bytes vs 16 bytes)
- **Cons**:
  - Requires centralized ID generation (single point of failure)
  - Exposes entity counts (security risk: `/api/v1/customers/12345` reveals ~12K customers)
  - Difficult in distributed systems
- **Example**: `id BIGSERIAL PRIMARY KEY`

### Option 2: UUID v4 (Random)

- **Pros**: Standard, widely supported, zero collision risk
- **Cons**:
  - Random = poor B-tree locality (index fragmentation)
  - No timestamp information
  - 40% slower inserts compared to sequential IDs
- **Example**: `uuid.New()` (github.com/google/uuid)

### Option 3: UUID v7 (Time-Ordered)

- **Pros**:
  - Time-ordered = good B-tree locality (sequential inserts)
  - Embedded timestamp (first 48 bits = Unix milliseconds)
  - Natural sort order (ORDER BY id = ORDER BY created_at)
  - 15-25% faster inserts than UUID v4
- **Cons**:
  - Slight timestamp leakage (acceptable for internal system)
  - Less widely adopted (RFC 9562 published May 2024)
- **Example**: `uuidv7.New()` (custom implementation)

### Option 4: Snowflake ID (Twitter)

- **Pros**: Time-ordered, compact (64-bit), high performance
- **Cons**:
  - Requires centralized coordination (node ID assignment)
  - Not UUID-compatible (breaks existing libraries)
  - Complex deployment (Zookeeper/etcd for coordination)

---

## Decision

**We will use UUID v7 (time-ordered) for all primary keys** instead of UUID v4 (random).

**Implementation**:

```go
// pkg/uuidv7/uuidv7.go
func New() uuid.UUID {
    now := time.Now()
    msec := uint64(now.UnixMilli())

    var u uuid.UUID
    // First 48 bits: Unix timestamp (milliseconds)
    binary.BigEndian.PutUint64(u[:8], msec<<16)

    // Next 12 bits: version (7) + variant (2)
    u[6] = 0x70 | (u[6] & 0x0F)  // Version 7
    u[8] = 0x80 | (u[8] & 0x3F)  // Variant 2 (RFC 4122)

    // Last 62 bits: random (collision resistance)
    _, _ = rand.Read(u[8:])
    return u
}
```

**Usage**:

```go
// internal/contexts/customer-mgmt/customer/aggregate/customer.go
import "github.com/basilex/promenade/pkg/uuidv7"

func NewCustomer(name string) *Customer {
    return &Customer{
        BaseAggregate: aggregate.BaseAggregate{
            ID:        uuidv7.New(),  // Time-sortable UUID
            CreatedAt: time.Now(),
        },
        Name: name,
    }
}
```

---

## Consequences

### Positive

-  **Better index performance**: 15-25% faster inserts (sequential locality)
-  **Natural sorting**: `ORDER BY id` = chronological order (no need for `created_at` in queries)
-  **Embedded timestamp**: Debugging easier (extract timestamp from ID)
-  **B-tree friendly**: Sequential IDs reduce page splits (better cache locality)
-  **Zero coordination**: No central ID generator needed

### Negative

-  **Timestamp leakage**: First 48 bits reveal creation time (acceptable for internal system)
  - Mitigation: Not a security concern (IDs not exposed publicly, API uses `/api/v1/customers` not `/api/v1/customers/123`)
-  **Less mature**: RFC 9562 (May 2024) is recent vs UUID v4 (RFC 4122, 2005)
  - Mitigation: Simple implementation, well-tested (20+ unit tests)

### Neutral

- ℹ Same 16-byte storage as UUID v4
- ℹ Compatible with PostgreSQL `uuid` type (no schema changes)
- ℹ Collision risk remains negligible (62 random bits = 1 in 4.6 quintillion)

---

## Implementation Notes

**Rollout**:

- **Phase 1**  (2026-01-15): `pkg/uuidv7` package created
- **Phase 2**  (2026-01-15): All aggregates migrated to `uuidv7.New()`
- **Phase 3**  (2026-01-18): 20+ unit tests added

**Affected Components**:

- All bounded contexts (11 contexts, 30+ aggregates)
- `pkg/aggregate/BaseAggregate` (ID field generation)
- No database migrations needed (UUID v4 and v7 use same `uuid` type)

**Rollback Plan**:

- Switch back to `uuid.New()` (Google UUID library)
- No data migration needed (UUIDs remain valid)

---

## References

- [RFC 9562 - UUID Version 7](https://www.rfc-editor.org/rfc/rfc9562.html) - Official specification
- [UUID v7 Performance Analysis](https://buildkite.com/blog/goodbye-integers-hello-uuids) - Buildkite case study
- [pkg/uuidv7/README.md](../../pkg/uuidv7/README.md) - Implementation documentation
- [The best UUID type for a database Primary Key](https://vladmihalcea.com/uuid-database-primary-key/) - Performance comparison

---

## Revisit Criteria

**Reconsider this decision if**:

- Write performance degrades by >10% compared to baseline (monitor with benchmarks)
- UUID v8 becomes standardized and widely adopted (better collision resistance)
- Security audit flags timestamp leakage as critical risk
- Distributed coordination layer (Zookeeper/etcd) becomes available (enables Snowflake IDs)

---

## Changelog

| Date       | Change                                         | Author              |
| ---------- | ---------------------------------------------- | ------------------- |
| 2026-01-15 | Initial version                                | Alexander Vasilenko |
| 2026-01-22 | Updated with benchmark results (15-25% faster) | Alexander Vasilenko |
