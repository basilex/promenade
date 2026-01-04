# BaseAggregate Field Duplication Fix

**Created:** January 4, 2026  
**Status:** 🚧 IN PROGRESS  
**Priority:** 🔴 CRITICAL - Блокує Billing розробку

---

## Проблема

Усі entity дублюють поля з `BaseAggregate`:
- `ID uuidv7.UUID` - дубльовано в BaseAggregate
- `CreatedAt time.Time` - дубльовано в BaseAggregate
- `UpdatedAt time.Time` - дубльовано в BaseAggregate
- `DeletedAt *time.Time` - ВІДСУТНЄ в BaseAggregate (потрібно додати)

Це призводить до:
- 🐛 Конфлікт полів при серіалізації (два `id` в JSON/DB)
- 🐛 Невизначеність - яке поле використовувати?
- 🐛 Неправильний маппінг до бази даних
- 🐛 Дублювання даних в пам'яті

---

## Виправлення

### ✅ 1. BaseAggregate - ВИПРАВЛЕНО

**Файл:** `pkg/aggregate/aggregate.go`

**Зміни:**
- ✅ Додано `DeletedAt *time.Time` поле
- ✅ Конструктор `NewBaseAggregate()` вже є

```go
type BaseAggregate struct {
    ID        uuidv7.UUID `db:"id" json:"id"`
    Version   int         `db:"version" json:"version"`
    CreatedAt time.Time   `db:"created_at" json:"created_at"`
    UpdatedAt time.Time   `db:"updated_at" json:"updated_at"`
    DeletedAt *time.Time  `db:"deleted_at" json:"deleted_at,omitempty"`
}
```

### ✅ 2. Identity - Contact - ВИПРАВЛЕНО

**Файл:** `internal/contexts/identity/contact/entity.go`

**Зміни:**
- ✅ Видалено дубльовані поля: ID, CreatedAt, UpdatedAt
- ✅ Видалено ініціалізацію ID, now := time.Now() з конструктора
- ✅ Замінено всі `c.UpdatedAt = time.Now()` на `c.Touch()`

**Приклад змін:**
```go
// БУЛО
type Contact struct {
    aggregate.BaseAggregate
    ID     uuidv7.UUID  // ❌ ДУБЛЮВАННЯ
    UserID uuidv7.UUID
    // ...
    CreatedAt time.Time  // ❌ ДУБЛЮВАННЯ
    UpdatedAt time.Time  // ❌ ДУБЛЮВАННЯ
}

// СТАЛО
type Contact struct {
    aggregate.BaseAggregate  // ✅ Містить ID, CreatedAt, UpdatedAt, DeletedAt
    UserID uuidv7.UUID
    // ...
}
```

---

## План Виправлення Інших Entity

### 🔴 Priority 1 - Identity Context (6 entity)

0. ✅ **Contact** - `internal/contexts/identity/contact/entity.go` - **COMPLETE**
   - ✅ Видалено: ID, CreatedAt, UpdatedAt
   - ✅ Замінено: UpdatedAt = time.Now() → Touch()
   - ✅ Repository toEntity() виправлено
   - ✅ DTO тести виправлено
   - ✅ Всі тести пройшли: `go test ./internal/contexts/identity/contact/... -v`

1. ⏳ **Profile** - `internal/contexts/identity/profile/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt
   - Замінити: UpdatedAt = time.Now() → Touch()

2. ⏳ **User** - `internal/contexts/identity/user/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt
   - Замінити: UpdatedAt = time.Now() → Touch()

3. ⏳ **Role** - `internal/contexts/identity/role/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt, DeletedAt
   - Замінити: UpdatedAt = time.Now() → Touch()
   - DeletedAt: вже буде в BaseAggregate

4. ⏳ **Permission** - `internal/contexts/identity/permission/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt
   - Замінити: UpdatedAt = time.Now() → Touch()

### 🟠 Priority 2 - Shared Context (4 entity)

5. ⏳ **Country** - `internal/contexts/shared/country/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt
   - Reference data - потребує акуратного тестування

6. ⏳ **Currency** - `internal/contexts/shared/currency/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt
   - Reference data - потребує акуратного тестування

7. ⏳ **Language** - `internal/contexts/shared/language/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt
   - Reference data - потребує акуратного тестування

8. ⏳ **Timezone** - `internal/contexts/shared/timezone/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt
   - Reference data - потребує акуратного тестування

### 🟡 Priority 3 - Customer Management Context (4 entity)

9. ⏳ **Customer** - `internal/contexts/customer-mgmt/customer/entity.go`
   - Видалити: ID, CreatedAt, UpdatedAt, DeletedAt
   - Замінити: UpdatedAt = time.Now() → Touch()

10. ⏳ **Company** - `internal/contexts/customer-mgmt/company/entity.go`
    - Видалити: ID, CreatedAt, UpdatedAt, DeletedAt
    - Замінити: UpdatedAt = time.Now() → Touch()

11. ⏳ **Deal** - `internal/contexts/customer-mgmt/deal/entity.go`
    - Видалити: ID, CreatedAt, UpdatedAt, DeletedAt
    - Замінити: UpdatedAt = time.Now() → Touch()

12. ⏳ **Interaction** - `internal/contexts/customer-mgmt/interaction/entity.go`
    - Видалити: ID, CreatedAt, UpdatedAt, DeletedAt
    - Замінити: UpdatedAt = time.Now() → Touch()

### 🟢 Priority 4 - Order Management Context (1 entity)

13. ⏳ **Order** - `internal/contexts/order-mgmt/order/entity.go`
    - Видалити: ID, CreatedAt, UpdatedAt, DeletedAt
    - Замінити: UpdatedAt = time.Now() → Touch()
    - OrderLine (вкладена структура): залишити як є

### 🔵 Priority 5 - Billing Context (2 entity)

14. ⏳ **Invoice** - `internal/contexts/billing/invoice/entity.go`
    - Видалити: ID, CreatedAt, UpdatedAt, DeletedAt
    - Замінити: UpdatedAt = time.Now() → Touch()
    - InvoiceLine (вкладена структура): залишити як є

15. ⏳ **Payment** - `internal/contexts/billing/payment/entity.go`
    - Видалити: ID, CreatedAt, UpdatedAt, DeletedAt
    - Замінити: UpdatedAt = time.Now() → Touch()

---

## Checklist на Entity

Для кожного entity:

### 1. Видалити Дубльовані Поля зі Struct
```go
// ❌ Видалити
type Entity struct {
    aggregate.BaseAggregate
    ID        uuidv7.UUID  // ВИДАЛИТИ
    CreatedAt time.Time    // ВИДАЛИТИ
    UpdatedAt time.Time    // ВИДАЛИТИ
    DeletedAt *time.Time   // ВИДАЛИТИ
}

// ✅ Залишити
type Entity struct {
    aggregate.BaseAggregate  // Містить усе
    // ... інші поля
}
```

### 2. Видалити Ініціалізацію з Конструктора
```go
// ❌ Видалити
now := time.Now()
entity := &Entity{
    BaseAggregate: aggregate.NewBaseAggregate(),
    ID:            uuidv7.New(),     // ВИДАЛИТИ
    CreatedAt:     now,               // ВИДАЛИТИ
    UpdatedAt:     now,               // ВИДАЛИТИ
}

// ✅ Залишити
entity := &Entity{
    BaseAggregate: aggregate.NewBaseAggregate(),  // Ініціалізує все
    // ... інші поля
}
```

### 3. Замінити UpdatedAt = time.Now() на Touch()
```go
// ❌ Старий спосіб
func (e *Entity) Update() {
    e.UpdatedAt = time.Now()
}

// ✅ Новий спосіб
func (e *Entity) Update() {
    e.Touch()
}
```

### 4. Перевірити Tests
- Repository tests: переконатись що ID, CreatedAt, UpdatedAt працюють
- Entity tests: перевірити конструктори
- DTO tests: якщо є маппінг ID/timestamps

---

## Після Виправлення

### Тести
```bash
# Після кожного entity
make test-unit
make test-integration

# Перевірити що нічого не зламалось
make test-all
```

### Швидка перевірка
```bash
# Знайти всі дубльовані ID поля
rg "^\s+ID\s+uuidv7\.UUID" internal/contexts --type go

# Знайти всі CreatedAt/UpdatedAt дубльовані
rg "^\s+(CreatedAt|UpdatedAt)\s+time\.Time" internal/contexts --type go

# Знайти всі використання UpdatedAt = time.Now()
rg "UpdatedAt\s*=\s*time\.Now\(\)" internal/contexts --type go
```

---

## Очікувані Проблеми

### 1. Дублювання полів в JSON
**Проблема:** Два поля `id` в JSON response  
**Рішення:** Після видалення дублікатів - проблема зникне

### 2. Дублювання полів в DB mapping
**Проблема:** sqlx не знає яке поле використати  
**Рішення:** Після видалення дублікатів - проблема зникне

### 3. Tests можуть fail
**Проблема:** Tests можуть звертатись до entity.ID напряму  
**Рішення:** Якщо треба - використати entity.BaseAggregate.ID або entity.GetID()

### 4. Repository може мати проблеми з mapping
**Проблема:** SELECT * може не працювати правильно  
**Рішення:** Перевірити всі repository після змін

---

## Статистика

- **Всього entity:** 15
- **Виправлено:** 1 (Contact) ✅
- **Залишилось:** 14 ⏳
- **Прогрес:** 6.7%

**ETA:** ~2-3 години (по 10-15хв на entity)

---

**Наступний крок:** Виправити Identity context entity (Profile, User, Role, Permission)
