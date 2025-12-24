# Архітектура Promenade - Швидка Довідка

[🇬🇧 English](../ARCHITECTURE_QUICKREF.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/ARCHITECTURE_QUICKREF.de.md) | [🇵🇹 Português](../pt/ARCHITECTURE_QUICKREF.pt.md) | [🇪🇸 Español](../es/ARCHITECTURE_QUICKREF.es.md)

## Core vs Модулі: Просте Правило

**CORE** = Інфраструктура + Довідкові Дані + Автентифікація

- Завжди увімкнений, надає сервіси

**МОДУЛІ** = Бізнес-Логіка

- Опційні, ліцензійні, незалежні

---

## Що Належить до Core?

### НАЛЕЖИТЬ до Core:

1. **Інфраструктурні Сервіси**

   - Управління з'єднанням з БД
   - Event Bus (Memory/Redis адаптери)
   - Планувальник (cron завдання)
   - Завантажувач конфігурації
   - Логер
   - Email сервіс
   - JWT менеджер

2. **Основа Безпеки**

   - Автентифікація користувачів (логін, реєстрація, пароль)
   - RBAC (ролі, дозволи, контроль доступу)
   - Сесії (JWT токени)

3. **Довідкові Дані**

   - Країни (145 країн, коди ISO 3166-1, регіони)
   - Валюти (124 валюти, ISO 4217, символи)
   - Регіони (30 адміністративних регіонів: штати, області, провінції, землі)
   - Міста (17 великих міст з координатами, населенням, столиці)
   - Методи Оплати (40+ методів: картки, гаманці, крипто, BNPL)
   - Часові пояси (база даних часових поясів IANA)
   - Мови (коди ISO 639)
   - _Стабільні, рідко змінювані дані, спільні для модулів_

4. **Управлінські Інтерфейси**
   - Реєстр модулів
   - Реєстр обробників очищення
   - Реєстр політик очищення
   - Інтерфейс event bus

### НЕ Належить до Core:

- Бізнес-сутності (Post, Comment, Profile і т.д.)
- Бізнес-сценарії використання
- Бізнес HTTP обробники
- Бізнес-маршрути
- Бізнес-конфігурація
- Логіка, специфічна для сутностей

**Емпіричне правило:** Якщо це бізнес-концепція, яку можна продати окремо, це МОДУЛЬ.

---

## Що Належить до Модулів?

### Структура Модуля

```
internal/modules/mymodule/
├── module.go              # Реалізація модуля
├── register.go            # Автореєстрація через init()
├── config/                # Власні YAML конфіги для кожного середовища
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
├── entity/                # Доменні сутності
├── usecase/               # Бізнес-логіка
└── adapter/
    ├── http/              # Обробники, DTO, маршрути
    ├── repository/        # Реалізації для Postgres
    └── purge/             # Обробники очищення (якщо потрібно)
```

### Чек-лист Модуля

- [ ] Має власну структуру entity/usecase/adapter
- [ ] Завантажує власну конфігурацію з `config/config.*.yaml`
- [ ] Реєструє маршрути в `RegisterRoutes()`
- [ ] Реєструє дозволи в `RegisterPermissions()`
- [ ] Реєструє обробники очищення (якщо сутності з м'яким видаленням)
- [ ] Немає імпортів з `internal/domain` або `internal/usecase`
- [ ] Використовує тільки пакети `pkg/*`

---

## Робочий Процес Розробки Модуля

### 1. Створення Модуля

```bash
mkdir -p internal/modules/mymodule/{config,entity,usecase,adapter/http/handler}
```

### 2. Реалізація Інтерфейсу Модуля

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type MyModule struct {
    *module.BaseModule
    db *sqlx.DB
    // ... інші поля
}

func New() module.Module {
    return &MyModule{
        BaseModule: module.NewBaseModule(module.Metadata{
            Name:        "mymodule",
            DisplayName: "My Module",
            Version:     "1.0.0",
            Description: "Does something useful",
        }),
    }
}

func (m *MyModule) Initialize(ctx context.Context, core *module.Core) error {
    // 1. Завантажити конфігурацію модуля
    cfg := moduleconfig.Load("internal/modules/mymodule/config", os.Getenv("ENVIRONMENT"))

    // 2. Налаштувати репозиторії, сценарії використання, обробники
    m.db = core.DB

    // 3. Зареєструвати обробники очищення (якщо потрібно)
    // 4. Зареєструвати політики зберігання (якщо потрібно)

    return nil
}

func (m *MyModule) RegisterRoutes(router *gin.RouterGroup) {
    group := router.Group("/mymodule")
    {
        group.GET("", m.handler.List)
        group.POST("", m.handler.Create)
    }
}

func (m *MyModule) RegisterPermissions() []module.Permission {
    return []module.Permission{
        {Resource: "mymodule", Action: "read", Description: "View items"},
        {Resource: "mymodule", Action: "create", Description: "Create items"},
    }
}

// ... реалізувати інші методи інтерфейсу
```

### 3. Автореєстрація

```go
// internal/modules/mymodule/register.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

func init() {
    module.DefaultRegistry.Register(New())
}
```

### 4. Додавання Конфігурації

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true
  version: "1.0.0"

mymodule:
  max_items: 100
  allow_public: true

purge:
  my_entities:
    retention_days: 60
    enabled: true
```

### 5. Увімкнення в Головній Конфігурації

```yaml
# config/modules.yaml
modules:
  enabled:
    - posts
    - profiles
    - mymodule # Додати тут
```

### 6. Імпорт в main.go

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/mymodule"  // Додати тут
)
```

---

## Правила Конфігурації

### Конфігурація Core

**Файл:** `config/app.{dev|test|prod}.yaml`

**Містить ТІЛЬКИ:**

- Налаштування інфраструктури (БД, сервер, JWT, логування)
- Конфігурація event bus
- Інфраструктура очищення (enabled, schedule, batch_size)
- Налаштування CORS
- Налаштування email сервісу

**НЕ містить:**

- Дні зберігання для конкретних сутностей → Модулі
- Налаштування, специфічні для модулів → Модулі
- Конфігурація бізнес-логіки → Модулі

### Конфігурація Модуля

**Файл:** `internal/modules/{name}/config/config.{dev|test|prod}.yaml`

**Містить:**

- Метадані модуля (name, version)
- Налаштування, специфічні для модуля
- Політики зберігання для очищення (якщо застосовно)
- Прапорці функцій (якщо застосовно)

**Завантажується:** Кожним модулем через `pkg/module/config.Load()`

---

## Правила Системи Очищення

### СТАРИЙ СПОСІБ (Core знає про сутності)

```go
//  НЕПРАВИЛЬНО - Core має конфігурацію для конкретних сутностей
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

### НОВИЙ СПОСІБ (Core тільки оркеструє)

**Конфігурація Core:**

```yaml
purge:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 1000
```

**Конфігурація Модуля:**

```yaml
purge:
  user_posts:
    retention_days: 90
    enabled: true
```

**Реєстрація Модуля:**

```go
// Модуль реєструє обробник
handler := purge.NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)

// Модуль реєструє політику
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

**Оркестрація Core:**

```go
// Core отримує ВСІ політики з реєстру
policies := purge.DefaultPolicyRegistry.GetAllPolicies()

// Core створює сценарій використання
useCase := usecase.NewPurgeUseCase(
    purge.DefaultRegistry,  // обробники
    policies,               // з модулів
    batchSize,
    eventBus,
)

// Core запускає планувальник
scheduler.Start(ctx)
```

**Результат:** Core нічого не знає про `user_posts` або дні зберігання!

---

## Шаблони Комунікації

### Подієво-Орієнтований (Рекомендується)

```go
// Модуль A публікує
event := &PostCreatedEvent{PostID: id}
eventBus.Publish(ctx, "post.created", event)

// Модуль B підписується
eventBus.Subscribe("post.created", func(e bus.Event) {
    // Обробити асинхронно
})
```

**Переваги:** Слабке зв'язування, асинхронна обробка

### Прямий Реєстр (Використовувати Обережно)

```go
// Отримати інший модуль
postsModule := core.Registry.Get("posts")

// Перевірити тип і викликати
if api, ok := postsModule.(PostsAPI); ok {
    stats := api.GetStats(userID)
}
```

**Використовувати тільки коли:** Потрібна синхронна відповідь, не можна використати події

---

## Поширені Помилки, Яких Слід Уникати

### Імпорт Пакетів Core в Модулях

```go
//  НЕПРАВИЛЬНО
import "github.com/basilex/promenade/internal/domain/entity"
import "github.com/basilex/promenade/internal/usecase"
```

**Виправлення:** Визначити сутності у власному пакеті модуля `entity/`.

### Жорстке Кодування Бізнес-Значень у Коді

```go
//  НЕПРАВИЛЬНО
const maxCommentLength = 2000
```

**Виправлення:** Завантажити з конфігурації модуля.

### Розміщення Бізнес-Логіки в Core

```go
//  НЕПРАВИЛЬНО - PostUseCase в internal/usecase/
```

**Виправлення:** Перемістити у пакет `usecase/` модуля.

### Core Знає про Сутності Модулів

```go
//  НЕПРАВИЛЬНО - Core має дні зберігання для постів
type PurgeConfig struct {
    RetentionDaysUserPosts int
}
```

**Виправлення:** Модуль реєструє політику зберігання через реєстр.

---

## Стратегія Тестування

### Тести Core

- Юніт-тести інфраструктурних сервісів
- Інтеграційні тести auth/RBAC
- Тести репозиторіїв довідкових даних

### Тести Модулів

- Юніт-тести бізнес-логіки (сценарії використання)
- Інтеграційні тести репозиторіїв
- Тести обробників з mock сценаріями використання

### Smoke Тести

- Наскрізні критичні потоки
- Тести міжмодульної комунікації через події
- Перевірка роботи маршрутів модулів

---

## Швидкі Команди

```bash
# Збірка
make build

# Запуск в режимі розробки
make dev

# Запуск тестів
make test

# Запуск тестів конкретного модуля
go test ./internal/modules/posts/...

# Запуск smoke тестів
make test-smoke

# Генерація шаблону модуля
make generate ENTITY=MyEntity

# Створення міграції
make migrate-create NAME=add_my_table
```

---

## Дерево Рішень: Core чи Модуль?

```
Це інфраструктура (БД, логер, event bus)?
└─> ТАК → CORE

Це безпека (auth, RBAC)?
└─> ТАК → CORE

Це довідкові дані (країни, валюти, регіони, міста, методи оплати)?
└─> ТАК → CORE

Це стабільне і використовується кількома модулями?
└─> ТАК → Розглянути CORE (або спільний pkg)

Це бізнес-логіка?
└─> ТАК → МОДУЛЬ

Це можна продати окремо?
└─> ТАК → МОДУЛЬ

Це специфічне для сутності?
└─> ТАК → МОДУЛЬ

Якщо сумніви?
└─> МОДУЛЬ (легше перемістити в core пізніше, ніж навпаки)
```

---

## Ресурси

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.uk.md) - Детальний огляд архітектури
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md) - Візуальна архітектура
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.uk.md) - Посібник з розробки модулів
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.uk.md) - Принципи незалежності
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.uk.md) - Деталі системи очищення
- [../../internal/CORE.md](../../internal/CORE.md) - Документація компонентів Core

---

**Пам'ятайте:**

- Core = Інфраструктура + Довідкові Дані + Автентифікація
- Модулі = Бізнес-Логіка (незалежні, ліцензійні)
- Використовуйте реєстри для слабкого зв'язування
- Події для асинхронної комунікації
- Автономія конфігурації для кожного модуля
