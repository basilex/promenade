---
title: "Розробка Модулів"
description: "Повний посібник зі створення модулів Promenade"
weight: 30
---

## Що Таке Модуль?

**Модуль** = Незалежний вертикальний зріз бізнес-логіки з власною:

- Доменною моделлю (entities)
- Бізнес-логікою (use cases)
- Доступом до даних (repositories)
- HTTP API (handlers, routes)
- Конфігурацією
- Міграціями бази даних
- Дозволами RBAC

---

## Структура Модуля

```
internal/modules/mymodule/
├── module.go              # Реалізація інтерфейсу IModule
├── register.go            # Авто-реєстрація через init()
├── config/                # Власні конфіги на середовище
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
├── entity/                # Доменні сутності
│   ├── item.go
│   └── item_test.go
├── usecase/               # Бізнес-логіка
│   ├── item_usecase.go
│   └── item_usecase_test.go
└── adapter/
    ├── http/
    │   ├── handler/       # HTTP обробники
    │   │   └── item_handler.go
    │   └── dto/           # Data Transfer Objects
    │       └── item_dto.go
    └── repository/
        └── postgres/
            └── item_repository.go
```

---

## Життєвий Цикл Модуля

```mermaid
sequenceDiagram
    participant App as Application
    participant Reg as Module Registry
    participant Mod as Module
    participant Core as Core Infrastructure

    Note over App: Startup
    App->>Reg: Load modules via init()
    Reg->>Mod: Initialize(core)
    Mod->>Core: Access DB, EventBus, JWT
    Mod->>Reg: RegisterRoutes(router)
    Mod->>Reg: RegisterPermissions()
    Mod->>Reg: RegisterMigrations()
    App->>Mod: Start() - begin workers

    Note over App: Runtime
    Note over Mod: Handle requests
    Note over Mod: Publish events

    Note over App: Shutdown
    App->>Mod: Stop() - cleanup
    Mod-->>Core: Close connections
```

---

## Крок 1: Реалізувати Інтерфейс IModule

```go
// internal/modules/mymodule/module.go
package mymodule

import (
    "context"
    "github.com/basilex/promenade/pkg/module"
    "github.com/gin-gonic/gin"
)

type MyModule struct {
    *module.BaseModule
    // Власні поля
    db       *sqlx.DB
    usecase  *ItemUseCase
    handler  *ItemHandler
    config   *Config
}

func New() module.IModule {
    return &MyModule{
        BaseModule: module.NewBaseModule(module.Metadata{
            Name:        "mymodule",
            DisplayName: "My Module",
            Version:     "1.0.0",
            Author:      "Your Name",
            Description: "Does something useful",
            License:     "MIT",
            Tags:        []string{"business", "feature"},
        }),
    }
}

func (m *MyModule) Initialize(ctx context.Context, core *module.Core) error {
    // 1. Зберегти core інфраструктуру
    m.db = core.DB

    // 2. Завантажити конфігурацію модуля
    cfg, err := LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    m.config = cfg

    // 3. Створити репозиторії
    repo := NewItemRepository(m.db)

    // 4. Створити use cases
    m.usecase = NewItemUseCase(repo, core.EventBus)

    // 5. Створити handlers
    m.handler = NewItemHandler(m.usecase)

    return nil
}

func (m *MyModule) RegisterRoutes(router *gin.RouterGroup) {
    items := router.Group("/mymodule")
    {
        items.GET("", m.handler.List)
        items.GET("/:id", m.handler.Get)
        items.POST("", m.handler.Create)
        items.PUT("/:id", m.handler.Update)
        items.DELETE("/:id", m.handler.Delete)
    }
}

func (m *MyModule) RegisterPermissions() []module.Permission {
    return []module.Permission{
        {Resource: "mymodule", Action: "read", Description: "View items"},
        {Resource: "mymodule", Action: "create", Description: "Create items"},
        {Resource: "mymodule", Action: "update", Description: "Update items"},
        {Resource: "mymodule", Action: "delete", Description: "Delete items"},
    }
}

func (m *MyModule) Start(ctx context.Context) error {
    // Запустити фонові робітники
    return nil
}

func (m *MyModule) Stop(ctx context.Context) error {
    // Закрити з'єднання, зупинити робітників
    return nil
}

func (m *MyModule) HealthCheck(ctx context.Context) error {
    // Перевірити статус модуля
    return nil
}
```

---

## Крок 2: Авто-Реєстрація

```go
// internal/modules/mymodule/register.go
package mymodule

import (
    "log/slog"
    "github.com/basilex/promenade/pkg/module"
)

func init() {
    if err := module.DefaultRegistry.Register(New()); err != nil {
        slog.Error("Failed to register mymodule", "error", err)
    }
}
```

---

## Крок 3: Конфігурація Модуля

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true
  version: "1.0.0"

mymodule:
  max_items: 100
  allow_public: true
  cache_ttl: 300 # seconds

purge:
  items:
    retention_days: 60
    enabled: true
```

```go
// internal/modules/mymodule/config/config.go
package config

import (
    "github.com/basilex/promenade/pkg/module/config"
)

type Config struct {
    Module struct {
        Name    string `yaml:"name"`
        Enabled bool   `yaml:"enabled"`
        Version string `yaml:"version"`
    } `yaml:"module"`

    MyModule struct {
        MaxItems    int  `yaml:"max_items"`
        AllowPublic bool `yaml:"allow_public"`
        CacheTTL    int  `yaml:"cache_ttl"`
    } `yaml:"mymodule"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := config.Load("internal/modules/mymodule/config", &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

---

## Паттерни Комунікації

### Event-Driven (Рекомендується)

```go
// Публікація події
event := &ItemCreatedEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    ItemID:    item.ID,
    UserID:    userID,
}
core.EventBus.Publish(ctx, "mymodule.item.created", event)

// Підписка на події
func (m *MyModule) RegisterEventHandlers(bus bus.IBus) error {
    return bus.Subscribe(ctx, "user.registered", m.onUserRegistered)
}

func (m *MyModule) onUserRegistered(ctx context.Context, e bus.Event) error {
    evt := e.(*UserRegisteredEvent)
    // Створити default items для нового користувача
    return m.usecase.CreateDefaultItems(ctx, evt.UserID)
}
```

### Direct Registry (Використовуйте обережно)

```go
// Отримати інший модуль
postsModule := core.Registry.Get("posts")
if api, ok := postsModule.(PostsAPI); ok {
    stats := api.GetStats(userID)
}
```

---

## Best Practices

### ✅ DO

- Використовуйте UUID v7 для всіх ID
- Завантажуйте власні конфіги з `config/`
- Реєструйте purge handlers для сутностей з soft delete
- Публікуйте події для важливих операцій
- Пишіть unit tests для use cases
- Документуйте всі публічні API

### ❌ DON'T

- Не імпортуйте `internal/domain` або `internal/usecase`
- Не імпортуйте інші модулі напряму
- Не зберігайте стан в обробниках (stateless)
- Не використовуйте глобальні змінні
- Не забувайте фільтр `deleted_at IS NULL` для soft delete

---

## Тестування Модулів

### Unit Tests

```go
// internal/modules/mymodule/usecase/item_usecase_test.go
func TestItemUseCase_Create(t *testing.T) {
    // Arrange
    mockRepo := &MockItemRepository{}
    mockBus := &MockEventBus{}
    usecase := NewItemUseCase(mockRepo, mockBus)

    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    mockBus.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

    // Act
    item, err := usecase.Create(context.Background(), "Test Item")

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, item.ID)
    mockRepo.AssertExpectations(t)
    mockBus.AssertExpectations(t)
}
```

### Integration Tests

```go
func TestItemRepository_Create(t *testing.T) {
    // Setup test DB
    db := setupTestDB(t)
    defer db.Close()

    repo := NewItemRepository(db)

    // Test
    item := &entity.Item{ID: uuidv7.New(), Name: "Test"}
    err := repo.Create(context.Background(), item)

    assert.NoError(t, err)
}
```

---

## Увімкнення Модуля

### 1. Імпортувати в main.go

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/mymodule"  // Додати тут
)
```

### 2. Додати в config/modules.yaml

```yaml
modules:
  enabled:
    - posts
    - profiles
    - mymodule # Додати тут
```

### 3. Створити міграції

```bash
make migrate-create MODULE=mymodule NAME=init_tables
```

---

## Troubleshooting

**Проблема:** Модуль не завантажується

- ✅ Перевірте import в `cmd/api/main.go`
- ✅ Перевірте `config/modules.yaml` - чи є модуль у `enabled`
- ✅ Перевірте логи на помилки реєстрації

**Проблема:** Міграції не виконуються

- ✅ Перевірте namespace: `migrations/mymodule/`
- ✅ Виконайте `make migrate-status`
- ✅ Перевірте формат імені файлу: `000001_description.up.sql`

**Проблема:** Роути не працюють

- ✅ Перевірте `RegisterRoutes()` викликається
- ✅ Перевірте шлях: `/api/v1/mymodule`
- ✅ Перевірте middleware (auth, permissions)

---

## Приклади Модулів

### 1. Simple CRUD Module

```go
type SimpleModule struct {
    *module.BaseModule
    crud *CRUDUseCase
}

func (m *SimpleModule) RegisterRoutes(r *gin.RouterGroup) {
    r.GET("/items", m.handler.List)
    r.POST("/items", m.handler.Create)
}
```

### 2. Background Worker Module

```go
func (m *WorkerModule) Start(ctx context.Context) error {
    go m.processQueue(ctx)
    return nil
}

func (m *WorkerModule) processQueue(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    for {
        select {
        case <-ticker.C:
            m.usecase.ProcessPendingJobs(ctx)
        case <-ctx.Done():
            return
        }
    }
}
```

### 3. Event-Driven Module

```go
func (m *EventModule) RegisterEventHandlers(bus bus.IBus) error {
    bus.Subscribe(ctx, "user.registered", m.onUserRegistered)
    bus.Subscribe(ctx, "post.created", m.onPostCreated)
    return nil
}
```

---

## Наступні Кроки

- [Незалежність Модулів](/promenade/docs/MODULE_INDEPENDENCE)
- [Архітектура](/promenade/docs/architecture)
- [Паттерни Бази Даних](/promenade/docs/database-schema)
- [Приклади Модулів](https://github.com/basilex/promenade/tree/dev/internal/modules)
