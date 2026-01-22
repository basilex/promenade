# Database Query Patterns

**Best practices for database queries in Promenade**

---

## Row Closing Pattern

###  Correct Pattern

Always close rows with defer, using anonymous function to ignore Close() error:

```go
rows, err := db.QueryContext(ctx, query, args...)
if err != nil {
    return nil, err
}
defer func() { _ = rows.Close() }() //  Correct

for rows.Next() {
    var item Item
    if err := rows.Scan(&item.ID, &item.Name); err != nil {
        return nil, err
    }
    items = append(items, item)
}

//  Always check rows.Err() after iteration
if err := rows.Err(); err != nil {
    return nil, err
}

return items, nil
```

### Why This Pattern?

1. **`defer func() { _ = rows.Close() }()`** - Anonymous function to explicitly ignore Close() error
   - Close() error is usually not actionable
   - We've already read all data we need
   - Ignoring error is acceptable here

2. **`rows.Err()`** - Check iteration error after loop
   - This catches errors that occurred during iteration
   - This is the important error to check

###  Incorrect Patterns

```go
//  Bad - No defer
rows, err := db.QueryContext(ctx, query)
for rows.Next() { ... }
rows.Close()  // May not execute if panic occurs

//  Bad - Checking Close() error (unnecessary)
defer func() {
    if err := rows.Close(); err != nil {
        return err  // Can't return from defer
    }
}()

//  Bad - Missing rows.Err() check
for rows.Next() { ... }
return items, nil  // Missing iteration error check
```

---

## Query Patterns

### Single Row Query

```go
func (r *Repository) GetByID(ctx context.Context, id uuidv7.UUID) (*Entity, error) {
    query := `SELECT id, name, created_at FROM table WHERE id = $1 AND deleted_at IS NULL`

    var entity Entity
    err := r.getExecutor(ctx).GetContext(ctx, &entity, query, id.String())
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }

    return &entity, nil
}
```

### Multi-Row Query

```go
func (r *Repository) List(ctx context.Context) ([]*Entity, error) {
    query := `SELECT id, name, created_at FROM table WHERE deleted_at IS NULL ORDER BY created_at DESC`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query entities: %w", err)
    }
    defer func() { _ = rows.Close() }()

    var entities []*Entity
    for rows.Next() {
        var entity Entity
        if err := rows.StructScan(&entity); err != nil {
            return nil, fmt.Errorf("failed to scan entity: %w", err)
        }
        entities = append(entities, &entity)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iteration error: %w", err)
    }

    return entities, nil
}
```

### JOIN Query with Grouping

For master-detail relationships (e.g., Budget with BudgetLines):

```go
func (r *Repository) ListWithDetails(ctx context.Context) ([]*Master, error) {
    query := `
        SELECT
            m.id, m.name,
            d.id as detail_id, d.amount
        FROM master m
        LEFT JOIN details d ON m.id = d.master_id
        WHERE m.deleted_at IS NULL
        ORDER BY m.id, d.id`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query: %w", err)
    }
    defer func() { _ = rows.Close() }()

    masterMap := make(map[string]*Master)
    var masterOrder []string

    for rows.Next() {
        var masterID, detailID sql.NullString
        var master Master
        var amount sql.NullInt64

        err := rows.Scan(&masterID, &master.Name, &detailID, &amount)
        if err != nil {
            return nil, fmt.Errorf("failed to scan: %w", err)
        }

        // Get or create master
        m, exists := masterMap[masterID.String]
        if !exists {
            master.ID, _ = uuidv7.Parse(masterID.String)
            master.Details = []*Detail{}
            masterMap[masterID.String] = &master
            masterOrder = append(masterOrder, masterID.String)
            m = &master
        }

        // Add detail if exists (LEFT JOIN may have NULL)
        if detailID.Valid {
            detail := &Detail{
                Amount: amount.Int64,
            }
            detail.ID, _ = uuidv7.Parse(detailID.String)
            m.Details = append(m.Details, detail)
        }
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iteration error: %w", err)
    }

    // Return in deterministic order
    var result []*Master
    for _, id := range masterOrder {
        result = append(result, masterMap[id])
    }

    return result, nil
}
```

---

## Transaction Pattern

```go
func (r *Repository) getExecutor(ctx context.Context) sqlx.ExtContext {
    if tx, ok := database.GetTx(ctx); ok {
        return tx
    }
    return r.db
}

// Use in service layer
func (uc *UseCase) CreateWithDetails(ctx context.Context, ...) error {
    return database.WithTx(ctx, uc.db, func(ctx context.Context) error {
        // Create master
        master, err := uc.masterRepo.Create(ctx, ...)
        if err != nil {
            return err
        }

        // Create details (uses same transaction via ctx)
        for _, detail := range details {
            if err := uc.detailRepo.Create(ctx, detail); err != nil {
                return err  // Rollback on error
            }
        }

        return nil  // Commit
    })
}
```

---

## Performance Tips

### Use Placeholders Correctly

```go
//  Good - PostgreSQL numbered placeholders
query := `SELECT * FROM users WHERE id = $1 AND email = $2`
db.QueryContext(ctx, query, id, email)

//  Bad - String concatenation (SQL injection risk!)
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)
```

### LIMIT Queries

```go
// Always add LIMIT for list queries
query := `SELECT * FROM items ORDER BY created_at DESC LIMIT $1 OFFSET $2`
rows, err := db.QueryContext(ctx, query, limit, offset)
```

### Use Indexes

```sql
-- Ensure queries have supporting indexes
CREATE INDEX idx_items_created_at ON items(created_at DESC) WHERE deleted_at IS NULL;
```

### Avoid N+1 Queries

```go
//  Bad - N+1 queries
orders, _ := orderRepo.List(ctx)
for _, order := range orders {
    items, _ := orderItemRepo.ListByOrderID(ctx, order.ID)  // N queries!
    order.Items = items
}

//  Good - Single JOIN query
orders, _ := orderRepo.ListWithItems(ctx)  // 1 query with LEFT JOIN
```

---

## Common Pitfalls

### 1. Forgetting rows.Err()

```go
//  Missing error check
for rows.Next() { ... }
return items, nil  // May silently skip errors!

//  Always check
for rows.Next() { ... }
if err := rows.Err(); err != nil {
    return nil, err
}
```

### 2. Not Closing Rows

```go
//  Resource leak
rows, _ := db.Query(query)
for rows.Next() { ... }
// rows never closed → connection leak

//  Always defer
rows, err := db.Query(query)
if err != nil { return err }
defer func() { _ = rows.Close() }()
```

### 3. Using sql.NullString Incorrectly

```go
//  Bad - Not checking Valid
var name sql.NullString
rows.Scan(&name)
entity.Name = name.String  // Might be empty even if NULL!

//  Good - Check Valid
if name.Valid {
    entity.Name = name.String
}
```

---

## See Also

- [Database Conventions](database-conventions.md) - Naming, migrations
- [Testing Patterns](testing-patterns.md) - Database testing
- [Security Patterns](security-patterns.md) - SQL injection prevention
