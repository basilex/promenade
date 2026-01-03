# Phase 4: Customer Management - Детальний План

**Дата**: 3 січня 2026  
**Статус**: 🎯 Ready to Start  
**Оцінка**: 3 дні (Days 9-11)

---

## 📊 Поточний стан (Baseline Audit)

### ✅ Що ВЖЕ зроблено:

1. **JSONB → TEXT міграція** ✅
   - Customer.tags: `TEXT` (не JSONB)
   - Interaction.attendees: `TEXT` (не JSONB)
   - Обидві колонки вже cross-DB compatible

2. **jsonstore.Field[T] впровадження** ✅
   - Customer: `jsonstore.Field[[]string]` для tags
   - Interaction: `jsonstore.Field[[]uuidv7.UUID]` для attendees
   - Repository вже використовує Get()/Set()

3. **UUID DEFAULT видалено** ✅
   - Всі 4 міграції не мають `DEFAULT uuid_v7()`
   - ID генерується в Go через конструктори

4. **Структура міграцій**:
   ```
   migrations/customer-mgmt/
   ├── 000001_customers.up.sql     (id UUID PRIMARY KEY)
   ├── 000002_companies.up.sql     (id UUID PRIMARY KEY)
   ├── 000003_deals.up.sql         (id UUID PRIMARY KEY)
   └── 000004_interactions.up.sql  (id UUID PRIMARY KEY)
   ```

### ❌ Що ТРЕБА зробити:

1. **Timestamp управління** ❌
   - `DEFAULT CURRENT_TIMESTAMP` ще є в міграціях
   - `updated_at` trigger не використовується (немає trigger)
   - Entity методи не викликають `Touch()`

2. **BaseAggregate Integration** ❌
   - Entity не embed BaseAggregate
   - Немає версіонування (optimistic locking)
   - Немає Touch() викликів в business методах

3. **Testing** ❌
   - Потрібні тести для Touch()
   - Потрібні тести для jsonstore з реальною БД

---

## 🎯 Phase 4 Tasks (Детально)

### Task 4.1: Entity Enhancement (30 min)

**Мета**: Додати BaseAggregate та Touch() до всіх 4 entities

#### 4.1.1 Customer Entity

**Файл**: `internal/contexts/customer-mgmt/customer/entity.go`

**Зміни**:
```go
// Before:
type Customer struct {
    ID         uuidv7.UUID
    Name       string
    // ... fields
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  *time.Time
}

// After:
type Customer struct {
    aggregate.BaseAggregate  // ✅ Embed BaseAggregate
    ID         uuidv7.UUID
    Name       string
    // ... fields
    DeletedAt  *time.Time
}
```

**Business методи що потребують Touch()**:
- `UpdateName(name string)` → `c.Touch()`
- `UpdateEmail(email valueobject.Email)` → `c.Touch()`
- `UpdatePhone(phone *valueobject.Phone)` → `c.Touch()`
- `UpdateStatus(status CustomerStatus)` → `c.Touch()`
- `UpdateTier(tier CustomerTier)` → `c.Touch()`
- `AssignTo(userID uuidv7.UUID)` → `c.Touch()`
- `AddTag(tag string)` → `c.Touch()`
- `RemoveTag(tag string)` → `c.Touch()`
- `RecordContact()` → `c.Touch()`
- `MarkConverted()` → `c.Touch()`
- `Churn(reason string)` → `c.Touch()`

**Constructor update**:
```go
func NewCustomer(name string, email valueobject.Email, source string, assignedTo uuidv7.UUID) (*Customer, error) {
    return &Customer{
        BaseAggregate: aggregate.NewBase(),  // ✅ Initialize
        ID:            uuidv7.New(),
        Name:          name,
        Email:         email,
        // ...
    }, nil
}
```

#### 4.1.2 Company Entity

**Файл**: `internal/contexts/customer-mgmt/company/entity.go`

**Business методи**:
- `UpdateName(name string)` → `c.Touch()`
- `UpdateLegalName(legalName string)` → `c.Touch()`
- `UpdateTaxID(taxID string)` → `c.Touch()`
- `UpdateRegistrationNumber(regNum string)` → `c.Touch()`
- `UpdateIndustry(industry string)` → `c.Touch()`
- `UpdateWebsite(website string)` → `c.Touch()`
- `UpdateEmployeeCount(count int)` → `c.Touch()`
- `UpdateRevenueRange(range string)` → `c.Touch()`
- `SetParentCompany(parentID uuidv7.UUID)` → `c.Touch()`
- `ClearParentCompany()` → `c.Touch()`

#### 4.1.3 Deal Entity

**Файл**: `internal/contexts/customer-mgmt/deal/entity.go`

**Business методи**:
- `UpdateName(name string)` → `d.Touch()`
- `UpdateValue(value valueobject.Money)` → `d.Touch()`
- `AssignToCustomer(customerID uuidv7.UUID)` → `d.Touch()`
- `AssignToSalesRep(salesRepID uuidv7.UUID)` → `d.Touch()`
- `MoveToStage(stage DealStage)` → `d.Touch()`
- `MarkWon(closeDate time.Time)` → `d.Touch()`
- `MarkLost(lostReason string)` → `d.Touch()`

#### 4.1.4 Interaction Entity

**Файл**: `internal/contexts/customer-mgmt/interaction/entity.go`

**Business методи**:
- `UpdateSubject(subject string)` → `i.Touch()`
- `UpdateDescription(description string)` → `i.Touch()`
- `Start()` → `i.Touch()`
- `End()` → `i.Touch()`
- `AddAttendee(attendeeID uuidv7.UUID)` → `i.Touch()`
- `RemoveAttendee(attendeeID uuidv7.UUID)` → `i.Touch()`
- `MarkRequiresFollowup(notes string, dueDate time.Time)` → `i.Touch()`
- `CompleteFollowup()` → `i.Touch()`
- `SetOutcome(outcome InteractionOutcome)` → `i.Touch()`

---

### Task 4.2: Repository Enhancement (45 min)

**Мета**: Оновити repositories для використання BaseAggregate timestamps

#### 4.2.1 Customer Repository

**Файл**: `internal/contexts/customer-mgmt/customer/adapter/repository/postgres/customer_repository.go`

**toEntity() - використати BaseAggregate timestamps**:
```go
func (r *customerRow) toEntity() (*Customer, error) {
    // ... existing parsing
    
    c := &customer.Customer{
        BaseAggregate: aggregate.BaseAggregate{
            CreatedAt: r.CreatedAt,  // ✅ Set from DB
            UpdatedAt: r.UpdatedAt,  // ✅ Set from DB
            Version:   r.Version,    // ✅ Optimistic locking
        },
        ID:    id,
        Name:  r.Name,
        // ... rest
    }
    return c, nil
}
```

**fromEntity() - читати BaseAggregate timestamps**:
```go
func fromEntity(c *customer.Customer) *customerRow {
    row := &customerRow{
        ID:        c.ID.String(),
        Name:      c.Name,
        // ...
        CreatedAt: c.GetCreatedAt(),  // ✅ From BaseAggregate
        UpdatedAt: c.GetUpdatedAt(),  // ✅ From BaseAggregate
        Version:   c.GetVersion(),    // ✅ For optimistic locking
    }
    // ... nullable fields
    return row
}
```

**Update() - використати IncrementVersion()**:
```go
func (r *customerRepository) Update(ctx context.Context, c *customer.Customer) error {
    c.IncrementVersion()  // ✅ Bumps version + updates timestamp
    
    row := fromEntity(c)
    query := `
        UPDATE customer_customers SET
            name = :name,
            email = :email,
            status = :status,
            updated_at = :updated_at,  -- ✅ Explicit from Go
            version = :version          -- ✅ For optimistic locking
        WHERE id = :id
          AND version = :version - 1    -- ✅ Check old version
          AND deleted_at IS NULL
    `
    result, err := r.NamedExec(ctx, query, row)
    // ... check rows affected for optimistic lock violation
}
```

**Аналогічно для**: Company, Deal, Interaction repositories

---

### Task 4.3: Migration Cleanup (15 min)

**Мета**: Видалити `DEFAULT CURRENT_TIMESTAMP` з міграцій

#### 4.3.1 Customer Migration

**Файл**: `migrations/customer-mgmt/000001_customers.up.sql`

**Зміни**:
```sql
-- Before:
created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

-- After:
created_at TIMESTAMP WITH TIME ZONE NOT NULL,
updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
```

**Аналогічно**: Companies, Deals, Interactions

**⚠️ ВАЖЛИВО**: Не створювати нові файли міграцій! Редагувати існуючі.

---

### Task 4.4: Constructor Updates (30 min)

**Мета**: Всі конструктори повинні ініціалізувати BaseAggregate

#### Приклад (Customer):

```go
func NewCustomer(name string, email valueobject.Email, source string, assignedTo uuidv7.UUID) (*Customer, error) {
    if name == "" {
        return nil, fmt.Errorf("name is required")
    }
    
    return &Customer{
        BaseAggregate: aggregate.NewBase(),  // ✅ Version=1, Now timestamps
        ID:            uuidv7.New(),         // ✅ Time-ordered UUID
        Name:          name,
        Email:         email,
        Status:        CustomerStatusLead,   // Default
        Tier:          CustomerTierFree,     // Default
        Source:        source,
        AssignedTo:    assignedTo,
        Tags:          []string{},           // Empty slice
    }, nil
}
```

**Аналогічно**: Company, Deal, Interaction

---

### Task 4.5: Testing (90 min)

#### 4.5.1 Unit Tests (entity_test.go)

**Customer entity tests**:
- `TestCustomer_Touch()` - перевірити що UpdatedAt змінюється
- `TestCustomer_AddTag()` - перевірити Touch() виклик
- `TestCustomer_UpdateStatus()` - перевірити Touch() + правила переходів

**Deal entity tests**:
- `TestDeal_MoveToStage()` - перевірити Touch() + probability update
- `TestDeal_MarkWon()` - перевірити Touch() + CloseDate

**Interaction entity tests**:
- `TestInteraction_AddAttendee()` - перевірити Touch() + attendees update via jsonstore

#### 4.5.2 Integration Tests (repository_test.go)

**Customer repository tests**:
```go
func TestCustomerRepository_OptimisticLocking(t *testing.T) {
    // 1. Create customer
    c := NewCustomer(...)
    repo.Create(ctx, c)  // Version=1
    
    // 2. Fetch twice
    c1, _ := repo.GetByID(ctx, c.ID)  // Version=1
    c2, _ := repo.GetByID(ctx, c.ID)  // Version=1
    
    // 3. Update first
    c1.UpdateName("New Name")
    repo.Update(ctx, c1)  // Version=2, OK
    
    // 4. Try update second (should fail - stale version)
    c2.UpdateName("Other Name")
    err := repo.Update(ctx, c2)  // Version still 1, CONFLICT
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "optimistic lock")
}
```

**Тести для всіх 4 repositories**:
- OptimisticLocking
- TimestampUpdates (created_at ≤ updated_at)
- JSONStore round-trip (tags, attendees)

#### 4.5.3 Integration Tests - JSONStore

**Customer tags test**:
```go
func TestCustomerRepository_TagsJSONStore(t *testing.T) {
    c := NewCustomer(...)
    c.AddTag("vip")
    c.AddTag("high-value")
    
    repo.Create(ctx, c)
    
    // Fetch back
    fetched, _ := repo.GetByID(ctx, c.ID)
    assert.Equal(t, []string{"vip", "high-value"}, fetched.Tags)
}
```

**Interaction attendees test**:
```go
func TestInteractionRepository_AttendeesJSONStore(t *testing.T) {
    attendee1 := uuidv7.New()
    attendee2 := uuidv7.New()
    
    i := NewInteraction(...)
    i.AddAttendee(attendee1)
    i.AddAttendee(attendee2)
    
    repo.Create(ctx, i)
    
    // Fetch back
    fetched, _ := repo.GetByID(ctx, i.ID)
    assert.Len(t, fetched.Attendees, 2)
    assert.Contains(t, fetched.Attendees, attendee1)
    assert.Contains(t, fetched.Attendees, attendee2)
}
```

---

## 📋 Checklist

### Phase 4.1: Entity Enhancement ✅ / ❌
- [ ] Customer: Embed BaseAggregate
- [ ] Customer: Add Touch() to 11 methods
- [ ] Company: Embed BaseAggregate
- [ ] Company: Add Touch() to 10 methods
- [ ] Deal: Embed BaseAggregate
- [ ] Deal: Add Touch() to 7 methods
- [ ] Interaction: Embed BaseAggregate
- [ ] Interaction: Add Touch() to 9 methods

### Phase 4.2: Repository Enhancement ✅ / ❌
- [ ] Customer repository: BaseAggregate integration
- [ ] Customer repository: Optimistic locking in Update()
- [ ] Company repository: BaseAggregate integration
- [ ] Company repository: Optimistic locking in Update()
- [ ] Deal repository: BaseAggregate integration
- [ ] Deal repository: Optimistic locking in Update()
- [ ] Interaction repository: BaseAggregate integration
- [ ] Interaction repository: Optimistic locking in Update()

### Phase 4.3: Migration Cleanup ✅ / ❌
- [ ] 000001_customers.up.sql: Remove CURRENT_TIMESTAMP defaults
- [ ] 000002_companies.up.sql: Remove CURRENT_TIMESTAMP defaults
- [ ] 000003_deals.up.sql: Remove CURRENT_TIMESTAMP defaults
- [ ] 000004_interactions.up.sql: Remove CURRENT_TIMESTAMP defaults

### Phase 4.4: Constructor Updates ✅ / ❌
- [ ] Customer: Initialize BaseAggregate in NewCustomer()
- [ ] Company: Initialize BaseAggregate in NewCompany()
- [ ] Deal: Initialize BaseAggregate in NewDeal()
- [ ] Interaction: Initialize BaseAggregate in NewInteraction()

### Phase 4.5: Testing ✅ / ❌
- [ ] Customer: Touch() unit tests
- [ ] Company: Touch() unit tests
- [ ] Deal: Touch() unit tests
- [ ] Interaction: Touch() unit tests
- [ ] Customer repository: Optimistic locking integration test
- [ ] Company repository: Optimistic locking integration test
- [ ] Deal repository: Optimistic locking integration test
- [ ] Interaction repository: Optimistic locking integration test
- [ ] Customer: Tags JSONStore integration test
- [ ] Interaction: Attendees JSONStore integration test
- [ ] Run full test suite: `make test`
- [ ] Run integration tests: `make test-integration`
- [ ] Verify no regressions

---

## ⏱️ Time Estimates

| Task | Subtasks | Estimated Time | Priority |
|------|----------|----------------|----------|
| 4.1  | Entity Enhancement (4 files) | 30 min | HIGH |
| 4.2  | Repository Enhancement (4 files) | 45 min | HIGH |
| 4.3  | Migration Cleanup (4 files) | 15 min | MEDIUM |
| 4.4  | Constructor Updates (4 files) | 30 min | HIGH |
| 4.5  | Testing (unit + integration) | 90 min | CRITICAL |
| **TOTAL** | **20 subtasks** | **210 min (3.5 hours)** | |

**Realistic estimate with buffer**: 4-5 hours (half day)

---

## 🚨 Risk Assessment

### High Risk:
1. **Optimistic Locking Breaking Change**
   - **Risk**: Concurrent updates may fail
   - **Mitigation**: Comprehensive integration tests + retry logic in handlers
   - **Rollback**: Revert to timestamp-only updates (no version check)

2. **BaseAggregate Embedding**
   - **Risk**: Changes struct layout, may break JSON serialization
   - **Mitigation**: Test DTO conversion thoroughly
   - **Rollback**: Keep CreatedAt/UpdatedAt as direct fields

### Medium Risk:
1. **Migration Timestamp Defaults**
   - **Risk**: Existing code may rely on DB defaults
   - **Mitigation**: Verify all INSERT queries provide timestamps
   - **Rollback**: Add DEFAULT back to migrations

2. **JSONStore Already Working**
   - **Risk**: May introduce regressions if we touch it
   - **Mitigation**: Don't change jsonstore.Field[T] usage, just add tests
   - **Rollback**: N/A (already done)

### Low Risk:
1. **Touch() Method Calls**
   - **Risk**: Forgetting Touch() in some business methods
   - **Mitigation**: Code review + unit tests for UpdatedAt changes
   - **Rollback**: Remove Touch() calls

---

## ✅ Success Criteria

**Phase 4 is complete when**:

1. ✅ All 4 entities embed `BaseAggregate`
2. ✅ All business methods call `Touch()`
3. ✅ All repositories use `IncrementVersion()` in Update()
4. ✅ All migrations have no `DEFAULT CURRENT_TIMESTAMP`
5. ✅ Optimistic locking works (version check in UPDATE)
6. ✅ JSONStore continues working (tags, attendees)
7. ✅ All existing tests pass
8. ✅ New tests added (Touch(), optimistic locking, JSONStore)
9. ✅ Integration tests pass with real database
10. ✅ No regressions in API endpoints

**Test Coverage Target**: 90%+ (maintain current level)

---

## 📖 References

- [DB Agnostic Refactoring Plan](DB_AGNOSTIC_REFACTORING.md) - Main document
- [pkg/aggregate/README.md](../../pkg/aggregate/README.md) - BaseAggregate docs
- [pkg/jsonstore/README.md](../../pkg/jsonstore/README.md) - Field[T] usage
- [Customer Management Guide](../concepts/customer-management.md) - Business rules
- [Deal Management Guide](../concepts/deal-management.md) - Deal lifecycle
- [Interaction Management Guide](../concepts/interaction-management.md) - Interaction tracking

---

**Last Updated**: 3 січня 2026  
**Status**: Ready to Execute  
**Next Action**: Start Task 4.1 (Entity Enhancement)
