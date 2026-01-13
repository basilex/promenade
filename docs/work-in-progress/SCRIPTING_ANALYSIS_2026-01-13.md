# Scripting Context Analysis Report

**Date**: January 13, 2026  
**Purpose**: Complete analysis before Phase 3 Week 2 implementation  
**Analyst**: Promenade Team

---

## Current State Summary

### ✅ Already Implemented (Week 1 - HTTP Layer)

**Phase 3 Week 1 COMPLETE** (January 8, 2026):

1. **LUA Engine** (`pkg/scripting/`) - DONE ✅
   - `engine.go` (210 lines) - Sandbox execution
   - `sandbox.go` (90 lines) - Security restrictions
   - `stdlib.go` (185 lines) - Standard Library with real UseCases
   - 21 tests passing (18 unit + 3 benchmarks)

2. **Script Aggregate** (`internal/contexts/scripting/script/`) - DONE ✅
   - `entity.go` (153 lines) - Script entity with status lifecycle
   - `execution.go` - ScriptExecution entity for audit trail
   - `repository.go` - IRepository interface (27 methods)
   - `usecase.go` (443 lines) - Complete business logic
   - `usecase_test.go` - Unit tests

3. **HTTP Layer** (`script/adapter/http/`) - DONE ✅
   - `handler.go` (523 lines) - 10 REST endpoints
   - `dto.go` - Request/Response DTOs
   - 12 smoke tests passing (100% pass rate)

4. **Repository** (`script/adapter/repository/postgres/`) - DONE ✅
   - `script_repository.go` (506 lines) - PostgreSQL implementation
   - `base_repository.go` - BaseRepository pattern
   - CRUD + querying complete

5. **Router** (`internal/contexts/scripting/router.go`) - DONE ✅
   - Context router with LUA engine integration
   - Registered in `cmd/api/server.go`
   - 10 endpoints operational

6. **Database Migration** (`migrations/scripting/000001`) - DONE ✅
   - `scripting_scripts` table (11 columns)
   - `scripting_script_executions` table (9 columns)
   - 9 indexes created

---

## Issues Identified

### 🚨 Problem #1: JSONB vs Database-Agnostic Pattern

**Current State**:
- Migration uses `JSONB` directly (PostgreSQL-specific)
- Entity uses `map[string]interface{}` for metadata
- Repository manually marshals/unmarshals JSON

**Pattern Violation**:
```sql
-- ❌ CURRENT (PostgreSQL-only)
metadata JSONB DEFAULT '{}'::jsonb
```

**Expected Pattern** (from other contexts):
```go
// ✅ Database-agnostic with jsonstore.Field[T]
// Example from Product aggregate:
Tags jsonstore.Field[[]string] `db:"tags"`

// Example from Subscription aggregate:
Metadata jsonstore.Field[map[string]string] `db:"metadata"`
```

**Impact**:
- Migration works on PostgreSQL ✅
- Migration will FAIL on SQLite ❌
- Migration will FAIL on MySQL ❌
- Breaks "db-agnostic" architecture principle

---

### 🚨 Problem #2: Manual JSON Marshaling in Repository

**Current Code** (`script_repository.go`):
```go
// ❌ Manual JSON handling (fragile, verbose)
if r.Metadata.Valid {
    var metadata map[string]interface{}
    if err := json.Unmarshal([]byte(r.Metadata.String), &metadata); err != nil {
        return nil, fmt.Errorf("failed to parse metadata: %w", err)
    }
    s.Metadata = metadata
}

// When saving:
metadataJSON, err := json.Marshal(s.Metadata)
if err != nil {
    return fmt.Errorf("failed to marshal metadata: %w", err)
}
row.Metadata = sql.NullString{String: string(metadataJSON), Valid: true}
```

**Expected Pattern** (from other contexts):
```go
// ✅ Automatic handling with jsonstore.Field
type productRow struct {
    Tags jsonstore.Field[[]string] `db:"tags"`
}

// No manual marshaling needed!
// Field[T] implements sql.Scanner and driver.Valuer
```

---

### 🚨 Problem #3: No Versioning Implementation

**Current State**:
- `version` column exists in `scripting_scripts` (increments on update)
- **NO version history table** (unlike original Phase 3 plan)
- **NO version snapshots** (cannot rollback to old script)
- **NO change_log** field

**Missing Features**:
1. Version history table (`scripting_script_versions`)
2. Snapshot storage (old script code + metadata)
3. Rollback capability
4. Change tracking

**Impact**:
- Cannot audit script changes ❌
- Cannot rollback to previous version ❌
- Compliance risk (no immutable audit trail) ❌

---

### 🚨 Problem #4: Missing Script Classification

**Current State**:
- Only `status` field (draft/active/inactive/archived)
- **NO script_type** (validation, workflow, report, pricing)
- **NO entity_type** (customer, order, deal, invoice)

**Original Phase 3 Plan Had**:
```sql
script_type VARCHAR(50) NOT NULL,  -- 'validation', 'workflow', 'report'
entity_type VARCHAR(50),           -- 'customer', 'order', 'deal'
```

**Impact**:
- Cannot filter scripts by type ❌
- Cannot organize by entity ❌
- Harder to manage in UI ❌

---

## Recommended Solution

### Option A: Add Missing Fields to Existing Table (Lightweight)

**Action**: Create migration `000002_enhance_scripts.up.sql`

```sql
-- Add missing classification fields
ALTER TABLE scripting_scripts 
ADD COLUMN script_type VARCHAR(50),
ADD COLUMN entity_type VARCHAR(50),
ADD CONSTRAINT chk_script_type CHECK (
    script_type IN ('validation', 'workflow', 'report', 'pricing', 'notification', 'custom')
);

-- Add version history table
CREATE TABLE scripting_script_versions (
    id UUID PRIMARY KEY,
    script_id UUID NOT NULL REFERENCES scripting_scripts(id),
    version INTEGER NOT NULL,
    code TEXT NOT NULL,
    metadata TEXT NOT NULL,  -- JSON as TEXT (db-agnostic)
    change_log TEXT,
    created_by UUID,
    created_at TIMESTAMP NOT NULL,
    UNIQUE(script_id, version)
);

CREATE INDEX idx_script_versions_script ON scripting_script_versions(script_id);
```

**Pros**:
- Non-breaking (backward compatible)
- Minimal code changes
- Fast to implement (1-2 hours)

**Cons**:
- Still uses JSONB in 000001 (PostgreSQL lock-in)
- Doesn't fix db-agnostic pattern

---

### Option B: Refactor to Database-Agnostic Pattern (Complete)

**Action**: Replace manual JSON with `jsonstore.Field[T]`

**Steps**:
1. Update `entity.go`:
```go
// ❌ OLD
type Script struct {
    Metadata map[string]interface{}
}

// ✅ NEW
type Script struct {
    Metadata jsonstore.Field[map[string]string]
}
```

2. Update `script_repository.go`:
```go
// ❌ OLD
type scriptRow struct {
    Metadata sql.NullString `db:"metadata"`
}

// ✅ NEW
type scriptRow struct {
    Metadata jsonstore.Field[map[string]string] `db:"metadata"`
}

// No manual marshaling needed! Field[T] handles it automatically
```

3. **CRITICAL**: Fix migration `000001_create_scripts_table.up.sql`:
```sql
-- ❌ OLD (PostgreSQL-only)
metadata JSONB DEFAULT '{}'::jsonb

-- ✅ NEW (Database-agnostic)
metadata TEXT  -- Works on Postgres, SQLite, MySQL
```

4. Add version history (Option A steps)

**Pros**:
- ✅ Database-agnostic (Postgres/SQLite/MySQL)
- ✅ Follows Promenade patterns
- ✅ Cleaner code (no manual marshaling)
- ✅ Type-safe with generics

**Cons**:
- Requires migration rollback + rewrite
- More changes (entity + repository)
- Takes longer (2-3 hours)

---

## Decision Matrix

| Criteria              | Option A (Add Fields) | Option B (Refactor)   |
| --------------------- | --------------------- | --------------------- |
| **Database Agnostic** | ❌ No                 | ✅ Yes                |
| **Pattern Compliance**| 🟡 Partial            | ✅ Full               |
| **Implementation**    | ✅ Fast (1-2h)        | 🟡 Medium (2-3h)      |
| **Breaking Changes**  | ✅ None               | 🟡 Migration rollback |
| **Future-proof**      | 🟡 Partial            | ✅ Yes                |
| **Code Quality**      | 🟡 Mixed patterns     | ✅ Clean              |

---

## Recommendation: **Option B (Refactor)**

**Reasoning**:
1. **Principle над швидкістю** - db-agnostic is core principle
2. **Зараз простіше** - Scripting context not in production yet
3. **Уникнути technical debt** - Option A = technical debt
4. **Consistency** - All other contexts use `jsonstore.Field[T]`
5. **Testing готовий** - Можна швидко перевірити

---

## Implementation Plan (Option B)

### Step 1: Rollback Migration (if already run)
```bash
make migrate-down NAMESPACE=scripting STEPS=1
```

### Step 2: Fix Migration File
```sql
-- Change JSONB → TEXT in migrations/scripting/000001_create_scripts_table.up.sql
metadata TEXT DEFAULT '{}'  -- Database-agnostic
```

### Step 3: Update Entity
```go
// entity.go
type Script struct {
    aggregate.BaseAggregate
    Name        string
    Description string
    Code        string
    Version     int
    Status      ScriptStatus
    ScriptType  string  // NEW: validation, workflow, etc
    EntityType  string  // NEW: customer, order, etc
    Metadata    jsonstore.Field[map[string]string]  // CHANGED
}
```

### Step 4: Update Repository
```go
// script_repository.go
type scriptRow struct {
    // ... other fields
    Metadata jsonstore.Field[map[string]string] `db:"metadata"`
}

// Remove manual JSON marshaling code!
// Field[T].Scan() and Field[T].Value() handle it automatically
```

### Step 5: Add Version History
Create `000002_add_script_versioning.up.sql`:
```sql
ALTER TABLE scripting_scripts 
ADD COLUMN script_type VARCHAR(50),
ADD COLUMN entity_type VARCHAR(50);

CREATE TABLE scripting_script_versions (
    id UUID PRIMARY KEY,
    script_id UUID REFERENCES scripting_scripts(id),
    version INTEGER NOT NULL,
    code TEXT NOT NULL,
    metadata TEXT NOT NULL,
    change_log TEXT,
    created_by UUID,
    created_at TIMESTAMP NOT NULL,
    UNIQUE(script_id, version)
);
```

### Step 6: Update Tests
- Fix entity tests (metadata access)
- Fix repository tests (database writes)
- Smoke tests should still pass

### Step 7: Re-run Migration
```bash
make migrate-scripting
```

---

## Timeline Estimate

**Option B (Recommended)**:
- Step 1-2: Migration fix (30 min)
- Step 3: Entity update (30 min)
- Step 4: Repository update (1 hour)
- Step 5: Version history (30 min)
- Step 6: Test fixes (30 min)
- Step 7: Testing (30 min)
- **Total**: ~4 hours

**Option A (Quick)**:
- New migration (30 min)
- Test (30 min)
- **Total**: ~1 hour

---

## Risk Assessment

**Option A Risks**:
- 🔴 Technical debt (JSONB lock-in)
- 🟡 Pattern inconsistency
- 🟢 Low implementation risk

**Option B Risks**:
- 🟢 No technical debt
- 🟢 Pattern aligned
- 🟡 Medium implementation risk (migration rollback)

---

## Conclusion

**Recommended Action**: **Option B (Refactor to Database-Agnostic)**

**Next Steps**:
1. Get approval for Option B
2. Start with Step 1 (rollback migration if needed)
3. Implement Steps 2-7
4. Test on PostgreSQL
5. Test on SQLite (validation!)
6. Commit changes
7. Continue Phase 3 Week 2 (UI Metadata)

---

**Status**: Awaiting decision  
**Priority**: HIGH (blocks Phase 3 Week 2)  
**Impact**: Foundation for all script storage
