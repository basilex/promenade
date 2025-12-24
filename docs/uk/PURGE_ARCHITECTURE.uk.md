# Архітектура Системи Очищення

🇬🇧 [English](PURGE_ARCHITECTURE.uk.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/PURGE_ARCHITECTURE.de.md) | [🇵🇹 Português](../pt/PURGE_ARCHITECTURE.pt.md) | [🇪🇸 Español](../es/PURGE_ARCHITECTURE.es.md)

## Огляд

Система очищення призначена для автоматичного видалення м'яко видалених записів на основі політик зберігання. Вона використовує **патерн оркестратора**, де:

- **Ядро** керує інфраструктурою (планувальник, use case, шина подій)
- **Модулі** володіють бізнес-логікою (обробники, політики зберігання)

## Принципи Архітектури

### 1. Незалежність Модулів

Кожен модуль відповідає за:

- Реєстрацію обробників очищення для своїх сутностей
- Визначення політик зберігання для своїх сутностей
- Реалізацію фактичної логіки очищення

Ядро **ніколи** не знає про конкретні типи сутностей або політики зберігання.

### 2. Патерн Реєстрів

Два глобальних реєстри забезпечують незалежність модулів:

#### Реєстр Обробників (`purge.DefaultRegistry`)

```go
// Модуль реєструє обробник під час ініціалізації
handler := NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)
```

#### Реєстр Політик (`purge.DefaultPolicyRegistry`)

```go
// Модуль реєструє політику зберігання
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

### 3. Оркестрація Ядра

Відповідальність ядра обмежується:

1. **Інфраструктурою**: Завантаження конфігурації очищення з YAML (enabled, schedule, dry_run, batch_size)
2. **Оркестрацією**: Збір політик з реєстру, створення use case, запуск планувальника
3. **Виконанням**: Запуск операцій очищення за розкладом
4. **Подіями**: Публікація подій очищення (успіх/невдача)

Ядро **НЕ**:

- Знає про конкретні сутності
- Визначає політики зберігання
- Реалізує логіку очищення

## Структура Конфігурації

### Конфігурація Ядра (`config/app.*.yaml`)

```yaml
# Тільки інфраструктура очищення ядра
purge:
  enabled: true # Головний перемикач
  schedule: "0 2 * * *" # Розклад cron (щодня о 2 ночі)
  dry_run: false # Режим попереднього перегляду
  batch_size: 1000 # Записів на пакет
```

### Конфігурація Модуля (`internal/modules/{назва}/config/config.*.yaml`)

```yaml
# Політики зберігання модуля
purge:
  user_posts:
    retention_days: 90 # Зберігати 90 днів
    enabled: true

  post_comments:
    retention_days: 30 # Зберігати 30 днів
    enabled: true
```

## Потік Реалізації

### 1. Ініціалізація Модуля

```go
func (m *PostsModule) Initialize(core *Core) error {
    // Завантажити конфігурацію модуля
    cfg := moduleconfig.Load("internal/modules/posts/config", environment)

    // Отримати налаштування зберігання
    postsRetentionDays := cfg.GetRetentionDays("purge.user_posts.retention_days")
    commentsRetentionDays := cfg.GetRetentionDays("purge.post_comments.retention_days")

    // Зареєструвати обробник очищення
    postHandler := NewPostPurgeHandler(db)
    purge.DefaultRegistry.Register(postHandler)

    // Зареєструвати політику зберігання
    policy := purge.RetentionPolicy{
        EntityName:    "user_posts",
        RetentionDays: postsRetentionDays,
        Enabled:       true,
    }
    purge.DefaultPolicyRegistry.RegisterPolicy(policy)

    return nil
}
```

### 2. Ініціалізація Ядра

```go
func InitPurgeModule(purgeConfig config.PurgeConfig, eventBus bus.Bus) {
    // Отримати всі зареєстровані обробники
    handlerRegistry := purge.DefaultRegistry

    // Отримати всі зареєстровані політики
    policyRegistry := purge.DefaultPolicyRegistry
    policies := policyRegistry.GetAllPolicies()

    // Перетворити на доменні сутності
    domainPolicies := convertToEntityPolicies(policies, purgeConfig.Enabled)

    // Створити use case
    useCase := usecase.NewPurgeUseCase(
        handlerRegistry,
        domainPolicies,
        purgeConfig.BatchSize,
        eventBus,
    )

    // Створити та запустити планувальник
    scheduler := scheduler.NewScheduler(useCase, purgeConfig.Schedule, ...)
    scheduler.Start(ctx)
}
```

### 3. Виконання Очищення

```go
// Планувальник запускає очищення за розкладом cron
func (s *Scheduler) runPurge(ctx context.Context) {
    // Use case проходить по політикам
    for _, policy := range policies {
        // Отримати обробник з реєстру
        handler, ok := registry.Get(policy.EntityName)
        if !ok {
            continue // Пропустити, якщо немає обробника
        }

        // Виконати очищення
        cutoffDate := time.Now().AddDate(0, 0, -policy.RetentionDays)
        recordsPurged, err := handler.Purge(ctx, cutoffDate, batchSize, dryRun)

        // Опублікувати події
        if err != nil {
            eventBus.Publish(ctx, bus.TopicPurgeFailed, ...)
        } else {
            eventBus.Publish(ctx, bus.TopicPurgeCompleted, ...)
        }
    }
}
```

## Створення Нового Обробника Очищення

### 1. Реалізувати Інтерфейс Обробника

```go
// internal/modules/mymodule/adapter/purge/handler.go
package purge

import (
    "context"
    "time"
    "github.com/jmoiron/sqlx"
)

type MyEntityPurgeHandler struct {
    db *sqlx.DB
}

func NewMyEntityPurgeHandler(db *sqlx.DB) *MyEntityPurgeHandler {
    return &MyEntityPurgeHandler{db: db}
}

func (h *MyEntityPurgeHandler) EntityName() string {
    return "my_entities"
}

func (h *MyEntityPurgeHandler) Purge(
    ctx context.Context,
    cutoffDate time.Time,
    batchSize int,
    dryRun bool,
) (int64, error) {
    query := `
        SELECT id FROM my_entities
        WHERE deleted_at IS NOT NULL
        AND deleted_at < $1
        LIMIT $2
    `

    var ids []string
    if err := h.db.SelectContext(ctx, &ids, query, cutoffDate, batchSize); err != nil {
        return 0, err
    }

    if dryRun {
        return int64(len(ids)), nil // Тільки попередній перегляд
    }

    deleteQuery := `DELETE FROM my_entities WHERE id = ANY($1)`
    result, err := h.db.ExecContext(ctx, deleteQuery, pq.Array(ids))
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}
```

### 2. Зареєструвати в Модулі

```go
// internal/modules/mymodule/module.go
func (m *MyModule) Initialize(core *Core) error {
    // Завантажити конфігурацію
    cfg := moduleconfig.Load("internal/modules/mymodule/config", env)
    retentionDays := cfg.GetRetentionDays("purge.my_entities.retention_days")

    // Зареєструвати обробник
    handler := purge.NewMyEntityPurgeHandler(m.db)
    if err := purge.DefaultRegistry.Register(handler); err != nil {
        return err
    }

    // Зареєструвати політику
    policy := purge.RetentionPolicy{
        EntityName:    "my_entities",
        RetentionDays: retentionDays,
        Enabled:       true,
    }
    if err := purge.DefaultPolicyRegistry.RegisterPolicy(policy); err != nil {
        return err
    }

    slog.Info("Registered purge for my_entities", "retention_days", retentionDays)
    return nil
}
```

### 3. Додати Конфігурацію Модуля

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true

purge:
  my_entities:
    retention_days: 60
    enabled: true
```

## Переваги

### ✅ Незалежність Модулів

- Модулі повністю володіють своєю логікою очищення
- Немає залежностей ядра від сутностей модулів
- Легко додавати/видаляти модулі

### ✅ Чіткість Конфігурації

- Ядро: Налаштування інфраструктури
- Модулі: Бізнес-політики
- Чітке розділення відповідальностей

### ✅ Тестованість

- Обробники можна тестувати незалежно
- Політики можна змінювати без змін коду
- Патерн реєстру дозволяє легко створювати моки

### ✅ Супроводжуваність

- Зміни політик зберігання не потребують змін в ядрі
- Нові сутності автоматично підхоплюються через реєстр
- Централізована логіка оркестрації

## Моніторинг

### Опубліковані Події

- `purge.entity.completed` - Очищення сутності успішне
- `purge.entity.failed` - Очищення сутності невдале
- `purge.all.completed` - Повний цикл очищення завершено

### Логи

```
level=info msg="Registered purge for user_posts" retention_days=90
level=info msg="Starting purge operation" entity=user_posts cutoff_date=2024-09-22
level=info msg="Purge operation completed" entity=user_posts records_purged=150 duration=2.3s
```

### Адміністративні Ендпоінти

- `GET /api/v1/admin/purge/policies` - Список всіх політик зберігання
- `POST /api/v1/admin/purge/preview/:entity` - Попередній перегляд очищення для сутності
- `POST /api/v1/admin/purge/execute/:entity` - Запустити очищення вручну

## Міграція зі Старої Системи

**Раніше** (Ядро знало про сутності):

```go
// ❌ Ядро мало конфігурацію для конкретних сутностей
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

**Тепер** (Ядро має лише інфраструктуру):

```go
// ✅ Ядро має лише налаштування інфраструктури
type PurgeConfig struct {
    Enabled   bool
    Schedule  string
    DryRun    bool
    BatchSize int
}
```

Модулі тепер реєструють свої власні політики через `purge.DefaultPolicyRegistry`.

## Див. Також

- [Незалежність Модулів](MODULE_INDEPENDENCE.uk.md)
- [Розробка Модулів](MODULE_DEVELOPMENT.uk.md)
- [Посібник з Тестування](TESTING_GUIDE.uk.md)
