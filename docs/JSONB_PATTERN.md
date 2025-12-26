# PostgreSQL JSONB Pattern - MUST HAVE

 **English** | [ Українська](uk/JSONB_PATTERN.uk.md) | [ Deutsch](de/JSONB_PATTERN.de.md)

---

##  CRITICAL: This is the ONLY correct way to work with JSONB in Promenade

**DO NOT** invent new patterns, use manual marshaling, or create dual fields (struct + []byte).

**ALWAYS** use `pkg/jsonb` package types.

---

## Why This Pattern Exists

PostgreSQL JSONB columns cannot directly map to Go structs. We **MUST** implement `sql.Scanner` and `driver.Valuer` interfaces to enable automatic serialization/deserialization by `sqlx`.

### What Went Wrong Before

 **Profiles Module** (old pattern - DO NOT COPY):

```go
// BAD: Dual fields + manual marshaling
type UserProfile struct {
    SocialLinks     SocialLinks `db:"-"`                // Go struct
    SocialLinksJSON []byte      `db:"social_links"`     // Raw JSON
}

// BAD: Manual marshaling in repository
func (r *Repository) Create(ctx context.Context, profile *UserProfile) error {
    profile.MarshalSocialLinks()  // MANUAL STEP - ERROR PRONE!
    // ...
}
```

**Problems:**

- Dual fields clutter code
- Manual marshaling required (7+ call sites)
- Easy to forget Marshal/Unmarshal
- NULL handling inconsistent
- Hard to maintain

 **Billing Module** (untested pattern):

```go
// UNKNOWN: No Scanner/Valuer, no integration tests
type Plan struct {
    Features []string `db:"features"` // Might work, might not
}
```

**Problems:**

- No integration tests with real PostgreSQL
- Relies on undefined sqlx behavior
- Breaks on complex types

 **Workflows Module** (broken pattern that started this):

```go
// BROKEN: Mixed pattern
type WorkflowDefinition struct {
    Definition     WorkflowSchema `db:"-"`
    DefinitionJSON []byte         `db:"definition"`
}

type WorkflowSchema struct {
    // Has Scan() but missing Value()
}
```

**Problems:**

- Incomplete Scanner/Valuer implementation
- "invalid input syntax for type json" errors
- 14/19 tests failing

---

## The Correct Pattern: pkg/jsonb

 **Notifications Module** (reference implementation):

```go
import "github.com/basilex/promenade/pkg/jsonb"

type Notification struct {
    Data jsonb.Map `db:"data" json:"data"`
}
```

**No manual marshaling. No dual fields. Just works.™**

---

## Usage Guide

### 1. Choose the Right Type

#### `jsonb.Map` - For flexible key-value data

Use when structure is dynamic or unknown:

```go
import "github.com/basilex/promenade/pkg/jsonb"

type UserProfile struct {
    Metadata jsonb.Map `db:"metadata" json:"metadata"`
}

// Usage
profile := &UserProfile{
    Metadata: jsonb.Map{
        "theme": "dark",
        "notifications": true,
        "last_login": "2025-12-26T10:00:00Z",
    },
}
```

**Use cases:**

- User preferences
- Dynamic configuration
- Event metadata
- API response data

#### `jsonb.Array` - For mixed-type lists

Use when array contains different types:

```go
type Activity struct {
    Data jsonb.Array `db:"data" json:"data"`
}

// Usage
activity := &Activity{
    Data: jsonb.Array{"user_login", 123, true, map[string]any{"ip": "1.2.3.4"}},
}
```

**Use cases:**

- Event logs with mixed data
- Dynamic lists
- API payloads with mixed types

#### `jsonb.JSON[T]` - For strongly-typed structures (RECOMMENDED)

Use when structure is known and consistent:

```go
type WorkflowSchema struct {
    States      []State      `json:"states"`
    Transitions []Transition `json:"transitions"`
}

type WorkflowDefinition struct {
    Definition jsonb.JSON[WorkflowSchema] `db:"definition" json:"definition"`
}

// Usage
definition := &WorkflowDefinition{}
definition.Definition.Set(WorkflowSchema{
    States:      []State{{Name: "pending"}},
    Transitions: []Transition{{From: "pending", To: "active"}},
})
```

**Use cases:**

- Workflow definitions
- JSON schemas
- Strongly-typed configurations
- Complex nested structures

---

## 2. Entity Definition

```go
package entity

import (
    "time"
    "github.com/basilex/promenade/pkg/jsonb"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// WorkflowSchema is our business logic type
type WorkflowSchema struct {
    States      []State      `json:"states"`
    Transitions []Transition `json:"transitions"`
}

type State struct {
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type Transition struct {
    From      string   `json:"from"`
    To        string   `json:"to"`
    Event     string   `json:"event"`
    Condition string   `json:"condition,omitempty"`
}

// WorkflowDefinition entity
type WorkflowDefinition struct {
    ID          uuidv7.UUID                 `db:"id" json:"id"`
    Name        string                      `db:"name" json:"name"`
    Definition  jsonb.JSON[WorkflowSchema]  `db:"definition" json:"definition"`   // JSONB column
    InputSchema *jsonb.JSON[WorkflowSchema] `db:"input_schema" json:"input_schema,omitempty"` // Optional JSONB
    CreatedAt   time.Time                   `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time                   `db:"updated_at" json:"updated_at"`
}

// Validation
func (w *WorkflowDefinition) Validate() error {
    if !w.Definition.Valid {
        return errors.New("definition is required")
    }
    if len(w.Definition.Data.States) == 0 {
        return errors.New("at least one state is required")
    }
    return nil
}
```

**Key Points:**

- Use `jsonb.JSON[T]` for strongly-typed JSONB
- Define business type (`WorkflowSchema`) separately
- Optional fields: Use pointer `*jsonb.JSON[T]`
- NO manual Marshal/Unmarshal methods needed
- NO dual fields (DefinitionJSON + Definition)

---

## 3. Repository Usage

```go
package repository

import (
    "context"
    "github.com/jmoiron/sqlx"
    "your-module/entity"
)

type WorkflowRepository struct {
    db *sqlx.DB
}

// Create - sqlx automatically calls Value() on jsonb.JSON
func (r *WorkflowRepository) Create(ctx context.Context, def *entity.WorkflowDefinition) error {
    query := `
        INSERT INTO workflow_definitions (id, name, definition, input_schema, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    // NO MANUAL MARSHALING NEEDED!
    _, err := r.db.ExecContext(ctx, query,
        def.ID,
        def.Name,
        def.Definition,     // sqlx calls Definition.Value() automatically
        def.InputSchema,    // handles NULL if pointer is nil
        def.CreatedAt,
        def.UpdatedAt,
    )
    return err
}

// GetByID - sqlx automatically calls Scan() on jsonb.JSON
func (r *WorkflowRepository) GetByID(ctx context.Context, id string) (*entity.WorkflowDefinition, error) {
    query := `SELECT * FROM workflow_definitions WHERE id = $1`

    var def entity.WorkflowDefinition

    // NO MANUAL UNMARSHALING NEEDED!
    if err := r.db.GetContext(ctx, &def, query, id); err != nil {
        return nil, err
    }

    return &def, nil
}

// List - works with slices too
func (r *WorkflowRepository) List(ctx context.Context) ([]*entity.WorkflowDefinition, error) {
    query := `SELECT * FROM workflow_definitions ORDER BY created_at DESC`

    var defs []*entity.WorkflowDefinition

    // Automatic scanning for each row
    if err := r.db.SelectContext(ctx, &defs, query); err != nil {
        return nil, err
    }

    return defs, nil
}
```

**Key Points:**

- NO `MarshalDefinition()` / `UnmarshalDefinition()` methods
- NO `DefinitionJSON []byte` temporary fields
- sqlx handles everything automatically via Scanner/Valuer
- NULL handling built-in (pointer types)
- Works with ExecContext, GetContext, SelectContext

---

## 4. HTTP Handler Usage

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "your-module/entity"
    "github.com/basilex/promenade/pkg/jsonb"
    "github.com/basilex/promenade/pkg/response"
)

type CreateWorkflowRequest struct {
    Name        string                     `json:"name" binding:"required"`
    Definition  WorkflowSchema             `json:"definition" binding:"required"`  // Plain struct in API
    InputSchema *WorkflowSchema            `json:"input_schema,omitempty"`        // Optional
}

type WorkflowResponse struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    Definition  WorkflowSchema  `json:"definition"`   // Plain struct in response
    InputSchema *WorkflowSchema `json:"input_schema,omitempty"`
}

func (h *Handler) Create(c *gin.Context) {
    var req CreateWorkflowRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }

    // Convert API DTO to entity
    def := &entity.WorkflowDefinition{
        ID:   uuidv7.New(),
        Name: req.Name,
    }

    // Use Set() to populate JSONB field
    def.Definition.Set(req.Definition)

    if req.InputSchema != nil {
        def.InputSchema = &jsonb.JSON[WorkflowSchema]{}
        def.InputSchema.Set(*req.InputSchema)
    }

    // Repository handles the rest
    if err := h.repo.Create(c.Request.Context(), def); err != nil {
        response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
        return
    }

    // Convert entity to response
    resp := WorkflowResponse{
        ID:         def.ID.String(),
        Name:       def.Name,
        Definition: def.Definition.Data,  // Direct access to typed data
    }

    if def.InputSchema != nil && def.InputSchema.Valid {
        resp.InputSchema = &def.InputSchema.Data
    }

    response.Success(c, http.StatusCreated, resp)
}
```

**Key Points:**

- DTOs use plain structs (API layer)
- Entity layer uses `jsonb.JSON[T]` (database layer)
- Use `.Set()` to populate JSONB fields
- Use `.Data` to access typed data
- Check `.Valid` for NULL fields

---

## 5. Migration Schema

```sql
-- Create table with JSONB columns
CREATE TABLE workflow_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(255) NOT NULL,
    definition JSONB NOT NULL,           -- Required JSONB
    input_schema JSONB,                  -- Optional JSONB (NULL allowed)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index JSONB columns for performance
CREATE INDEX idx_workflow_definitions_definition ON workflow_definitions USING GIN (definition);
CREATE INDEX idx_workflow_definitions_input_schema ON workflow_definitions USING GIN (input_schema);

-- Query JSONB data
-- Find workflows with specific state
SELECT * FROM workflow_definitions
WHERE definition @> '{"states": [{"name": "pending"}]}'::jsonb;

-- Extract values from JSONB
SELECT
    id,
    name,
    definition->'states' as states,
    jsonb_array_length(definition->'states') as state_count
FROM workflow_definitions;
```

---

## NULL Handling

### Required JSONB Field

```go
type Entity struct {
    Data jsonb.JSON[MyStruct] `db:"data" json:"data"`
}

// Always set before saving
entity := &Entity{}
entity.Data.Set(MyStruct{...})
entity.Data.Valid // true

// In database: NOT NULL constraint
```

### Optional JSONB Field

```go
type Entity struct {
    Data *jsonb.JSON[MyStruct] `db:"data" json:"data,omitempty"`
}

// NULL case
entity := &Entity{} // Data is nil
// In database: NULL

// Non-NULL case
entity := &Entity{
    Data: &jsonb.JSON[MyStruct]{},
}
entity.Data.Set(MyStruct{...})
// In database: {"field":"value"}
```

---

## Testing

```go
func TestWorkflowRepository_Create(t *testing.T) {
    db := testhelpers.SetupTestDB(t)
    repo := NewWorkflowRepository(db)

    // Create test data
    def := &entity.WorkflowDefinition{
        ID:   uuidv7.New(),
        Name: "Test Workflow",
    }

    // Set JSONB field
    def.Definition.Set(entity.WorkflowSchema{
        States: []entity.State{
            {Name: "pending", Description: "Initial state"},
            {Name: "active", Description: "Active state"},
        },
        Transitions: []entity.Transition{
            {From: "pending", To: "active", Event: "activate"},
        },
    })

    // Save to database
    err := repo.Create(context.Background(), def)
    require.NoError(t, err)

    // Retrieve and verify
    retrieved, err := repo.GetByID(context.Background(), def.ID.String())
    require.NoError(t, err)

    assert.True(t, retrieved.Definition.Valid)
    assert.Len(t, retrieved.Definition.Data.States, 2)
    assert.Equal(t, "pending", retrieved.Definition.Data.States[0].Name)
}
```

---

## Comparison: Old vs New

### Before (Profiles Pattern - Complex)

```go
// Entity with dual fields
type UserProfile struct {
    SocialLinks     SocialLinks `db:"-"`
    SocialLinksJSON []byte      `db:"social_links"`
}

func (p *UserProfile) MarshalSocialLinks() error {
    data, err := json.Marshal(p.SocialLinks)
    if err != nil {
        return err
    }
    p.SocialLinksJSON = data
    return nil
}

func (p *UserProfile) UnmarshalSocialLinks() error {
    if len(p.SocialLinksJSON) == 0 {
        return nil
    }
    return json.Unmarshal(p.SocialLinksJSON, &p.SocialLinks)
}

// Repository - manual marshaling
func (r *Repository) Create(ctx context.Context, profile *UserProfile) error {
    // MANUAL STEP - easy to forget!
    if err := profile.MarshalSocialLinks(); err != nil {
        return err
    }

    query := `INSERT INTO profiles (..., social_links) VALUES (..., $10)`
    _, err := r.db.ExecContext(ctx, query, ..., profile.SocialLinksJSON)
    return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*UserProfile, error) {
    var profile UserProfile
    err := r.db.GetContext(ctx, &profile, query, id)
    if err != nil {
        return nil, err
    }

    // MANUAL STEP - easy to forget!
    if err := profile.UnmarshalSocialLinks(); err != nil {
        return nil, err
    }

    return &profile, nil
}
```

**Problems:**

- 7+ places calling Marshal/Unmarshal
- Dual fields everywhere
- Easy to forget marshaling step
- Verbose and error-prone

### After (New Pattern - Clean)

```go
// Entity with single field
type UserProfile struct {
    SocialLinks jsonb.JSON[SocialLinks] `db:"social_links" json:"social_links"`
}

// Repository - automatic
func (r *Repository) Create(ctx context.Context, profile *UserProfile) error {
    query := `INSERT INTO profiles (..., social_links) VALUES (..., $10)`
    _, err := r.db.ExecContext(ctx, query, ..., profile.SocialLinks)
    return err  // sqlx calls Value() automatically
}

func (r *Repository) GetByID(ctx context.Context, id string) (*UserProfile, error) {
    var profile UserProfile
    err := r.db.GetContext(ctx, &profile, query, id)
    return &profile, err  // sqlx calls Scan() automatically
}
```

**Benefits:**

- Single field
- No manual marshaling
- Cannot forget steps
- Clean and maintainable

---

## Migration Guide for Existing Code

### Step 1: Add import

```go
import "github.com/basilex/promenade/pkg/jsonb"
```

### Step 2: Replace dual fields with jsonb.JSON

Before:

```go
type MyEntity struct {
    Data     MyStruct `db:"-"`
    DataJSON []byte   `db:"data"`
}
```

After:

```go
type MyEntity struct {
    Data jsonb.JSON[MyStruct] `db:"data" json:"data"`
}
```

### Step 3: Remove Marshal/Unmarshal methods

Delete:

```go
func (e *MyEntity) MarshalData() error { ... }
func (e *MyEntity) UnmarshalData() error { ... }
```

### Step 4: Update repository - remove marshaling calls

Before:

```go
func (r *Repo) Create(ctx context.Context, entity *MyEntity) error {
    if err := entity.MarshalData(); err != nil {  // REMOVE THIS
        return err
    }
    _, err := r.db.ExecContext(ctx, query, ..., entity.DataJSON)  // Change to entity.Data
    return err
}
```

After:

```go
func (r *Repo) Create(ctx context.Context, entity *MyEntity) error {
    _, err := r.db.ExecContext(ctx, query, ..., entity.Data)
    return err
}
```

### Step 5: Update handlers - use Set/Data

Before:

```go
entity.Data = myStruct
entity.MarshalData()
```

After:

```go
entity.Data.Set(myStruct)
```

### Step 6: Run tests

```bash
make test-module-mymodule
```

---

## Common Mistakes

###  Using map[string]interface{} directly

```go
// WRONG - no Scanner/Valuer
type Entity struct {
    Data map[string]interface{} `db:"data"`
}
```

Fix: Use `jsonb.Map`

###  Creating custom JSONB types in modules

```go
// WRONG - reinventing the wheel
type MyJSONB map[string]any
func (m MyJSONB) Value() (driver.Value, error) { ... }
```

Fix: Use `pkg/jsonb.Map` or `pkg/jsonb.JSON[T]`

###  Manual marshaling in repository

```go
// WRONG - unnecessary manual work
data, _ := json.Marshal(entity.Data)
_, err := db.Exec(query, data)
```

Fix: Pass `entity.Data` directly, let sqlx handle it

###  Forgetting to check Valid for nullable fields

```go
// WRONG - panic if NULL
value := entity.Data.Data.SomeField
```

Fix:

```go
if entity.Data != nil && entity.Data.Valid {
    value := entity.Data.Data.SomeField
}
```

---

## Performance Considerations

### JSONB Indexing

```sql
-- GIN index for containment queries (@>, @?, @@)
CREATE INDEX idx_entity_data ON entities USING GIN (data);

-- GIN index with jsonb_path_ops (smaller, faster for @> queries only)
CREATE INDEX idx_entity_data ON entities USING GIN (data jsonb_path_ops);
```

### Query Optimization

```sql
-- Good: Use indexes
SELECT * FROM entities WHERE data @> '{"status": "active"}'::jsonb;

-- Bad: Sequential scan
SELECT * FROM entities WHERE data->>'status' = 'active';
```

### Memory Usage

- `jsonb.JSON[T]` stores typed data in memory (efficient)
- `jsonb.Map` uses `map[string]any` (flexible but uses more memory)
- Choose based on your use case

---

## Checklist for Code Reviews

When reviewing code with JSONB:

- [ ] Using `pkg/jsonb` types (Map, Array, or JSON[T])
- [ ] NO dual fields (struct + []byte)
- [ ] NO manual Marshal/Unmarshal in repository
- [ ] NO custom JSONB types in modules
- [ ] Proper NULL handling for optional fields (pointer types)
- [ ] Using `.Set()` to populate JSONB fields
- [ ] Using `.Data` to access typed data
- [ ] Checking `.Valid` before accessing nullable fields
- [ ] GIN indexes on JSONB columns
- [ ] Integration tests with real PostgreSQL

---

## Resources

- [PostgreSQL JSONB Documentation](https://www.postgresql.org/docs/current/datatype-json.html)
- [Go database/sql/driver Package](https://pkg.go.dev/database/sql/driver)
- [sqlx Documentation](https://jmoiron.github.io/sqlx/)
- [pkg/jsonb/jsonb.go](../pkg/jsonb/jsonb.go) - Implementation
- [pkg/jsonb/jsonb_test.go](../pkg/jsonb/jsonb_test.go) - Usage examples

---

## Questions?

If you're unsure how to implement JSONB for your use case:

1. Check [internal/modules/notifications](../internal/modules/notifications) - Reference implementation
2. Read [pkg/jsonb/jsonb_test.go](../pkg/jsonb/jsonb_test.go) - Test examples
3. Ask in team chat or create an issue

**Remember:** This is the ONLY pattern. No exceptions. No "but in my case...". Use `pkg/jsonb`. 
