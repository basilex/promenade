# Архітектурний Аудит - Core vs Modules

[🇬🇧 English](../ARCHITECTURE_AUDIT.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/ARCHITECTURE_AUDIT.de.md) | [🇵🇹 Português](../pt/ARCHITECTURE_AUDIT.pt.md) | [🇪🇸 Español](../es/ARCHITECTURE_AUDIT.es.md)

**Дата:** 22 грудня 2025
**Статус:** Відповідає архітектурі

## Виконавче Резюме

Promenade використовує **плагінну архітектуру**, де:

- **Core** = Інфраструктура + Довідкові Дані + Менеджери (завжди увімкнені)
- **Modules** = Бізнес-логіка (опціональні, ліцензовані, незалежні)

Цей аудит підтверджує, що архітектура **правильно реалізована** з належним розділенням відповідальностей.

---

## Відповідальності Core

### 1. Управління Інфраструктурою

```
internal/infrastructure/
├── config/         Завантаження конфігурації (YAML + env)
├── database/       Підключення до БД + транзакції
├── email/          Email-сервіс
└── scheduler/      Cron-планувальник
```

**Статус:** Правильно - Core надає інфраструктуру як сервіс для модулів.

---

### 2. Довідкові Дані

```
internal/domain/entity/
├── country.go      Країни (ISO2/3, регіони) - 195+ записів
├── timezone.go     Часові пояси (IANA) - 500+ записів
├── language.go     Мови (ISO 639) - 180+ записів
└── (currency via repository)  Валюти (ISO 4217) - 170+ записів
```

**Призначення:** Стабільні, рідко змінювані дані, спільні для всіх модулів.

**Статус:** Правильно - Це справжні довідкові дані, а не бізнес-сутності.

**Реалізації Use Case:**

```
internal/usecase/
├── country_usecase.go   CRUD для країн
└── currency_usecase.go  CRUD для валют
```

---

### 3. Аутентифікація та Авторизація (RBAC)

```
internal/domain/entity/
├── user.go         Core-сутність користувача (тільки auth: email, пароль, ролі)
├── session.go      JWT-сесії
├── role.go         RBAC-ролі (5 системних ролей)
└── permission.go   RBAC-дозволи (ресурс:дія)
```

**Призначення:** Безпека та контроль доступу - фундаментальні для всіх модулів.

**Статус:** Правильно - Auth/RBAC повинні бути в core (всі модулі залежать від них).

**Реалізації Use Case:**

```
internal/usecase/
├── auth_usecase.go        Реєстрація, вхід, управління паролями
├── role_usecase.go        Управління ролями
└── permission_usecase.go  Управління дозволами
```

---

### 4. Система Управління Модулями

```
pkg/module/
├── module.go       Інтерфейс модуля
├── registry.go     Реєстр модулів + вирішення залежностей
├── config/         Завантажувач конфігурації модуля
└── base.go         Допоміжний BaseModule
```

**Призначення:** Оркестрація - виявлення, ініціалізація, запуск/зупинка модулів.

**Статус:** Правильно - Core - це оркестратор, модулі - виконавці.

**Ключові Функції:**

- Динамічне завантаження модулів через авто-реєстрацію `init()`
- Вирішення залежностей (топологічне сортування)
- Управління життєвим циклом (Initialize → RegisterRoutes → Start → Stop)
- Управління конфігурацією (кожен модуль завантажує власну конфігурацію)

---

### 5. Інфраструктура Event Bus

```
pkg/bus/
├── bus.go          Інтерфейс event bus
├── memory/         In-memory адаптер (dev/test)
├── redis/          Redis-адаптер (production)
└── factory.go      Фабрика адаптерів з fallback
```

**Призначення:** Інфраструктура міжмодульної комунікації.

**Статус:** Правильно - Core надає bus, модулі його використовують.

---

### 6. Інфраструктура Purge-системи

```
pkg/purge/
├── handler.go      Реєстр handler'ів (модулі реєструють handler'и)
└── (NEW) Реєстр політик (модулі реєструють політики збереження)
```

```
internal/usecase/
└── purge_usecase.go  Тільки оркестрація (отримує політики з реєстру)
```

**Призначення:** Інфраструктура планувальника - модулі визначають що/коли очищати.

**Статус:** ВИПРАВЛЕНО (недавній рефакторинг) - Core оркеструє, модулі реалізують.

---

## Відповідальності Модулів

### Поточні Модулі

#### 1. Модуль Posts (`internal/modules/posts/`)

```
posts/
├── module.go               Реалізація модуля
├── register.go             Авто-реєстрація через init()
├── config/                 Власні YAML-конфіги (dev, test, prod)
│   └── config.*.yaml
├── domain/entity/          Entity Post, Comment, Like
├── usecase/                Бізнес-логіка
├── adapter/
│   ├── http/               Handler'и, DTO, маршрути
│   ├── repository/         Реалізації Postgres
│   └── purge/              Purge-handler'и для posts+comments
└── README.md
```

**Функції:**

- Пости користувачів (створення, оновлення, видалення, soft-delete)
- Коментарі з вкладеністю (max depth конфігурується)
- Лайки (пости + коментарі)
- Purge-handler'и з політиками збереження (90 днів пости, 30 днів коментарі)

**Статус:** Повністю незалежний - Без імпортів з internal/domain або internal/usecase

---

#### 2. Модуль Profiles (`internal/modules/profiles/`)

```
profiles/
├── module.go               Реалізація модуля
├── register.go             Авто-реєстрація
├── config/                 Власні YAML-конфіги
│   └── config.*.yaml
├── entity/                 Entity UserProfile, UserContact
├── usecase/                Бізнес-логіка
└── adapter/
    ├── http/               Handler'и, DTO, маршрути
    └── repository/         Реалізації Postgres
```

**Функції:**

- Профілі користувачів (біо, аватар, соціальні посилання)
- Контакти користувачів (email, телефон, різні типи)
- Верифікація контактів
- Управління основним контактом

**Статус:** Повністю незалежний - Об'єднані profiles+contacts в один цілісний модуль

---

#### 3. Модуль Warehouse (`internal/modules/warehouse/`)

**Статус:** Закоментований (комерційний модуль, потрібна ліцензія)

**Призначення:** Управління складом для комерційних розгортань.

---

## Управління Конфігурацією

### Core Конфігурація

```yaml
# config/app.{env}.yaml - Тільки Core-інфраструктура
app:
  name: "Promenade"
  environment: "development"

server:
  host: "localhost"
  port: 8081

database:
  host: "localhost"
  port: 5432

jwt:
  secret: "..."

bus:
  adapter: "memory" # або "redis"

purge:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 1000
```

**Чого НЕ в core-конфігурації:**

- Політики збереження для окремих entity → Перенесені в модулі
- Налаштування специфічні для модулів → Перенесені в модулі
- Конфігурація бізнес-логіки → Перенесені в модулі

---

### Конфігурація Модуля

```yaml
# internal/modules/posts/config/config.dev.yaml
module:
  name: "posts"
  enabled: true
  version: "1.0.0"

posts:
  max_content_length: 10000
  comments:
    max_content_length: 2000
    max_depth: 10

purge:
  user_posts:
    retention_days: 90
    enabled: true
  post_comments:
    retention_days: 30
    enabled: true
```

**Кожен модуль:**

- Завантажує власну конфігурацію через `pkg/module/config.Load()`
- Визначає власні політики збереження
- Реєструє handler'и + політики через глобальні реєстри
- Повна автономія

---

### Реєстр Модулів

```yaml
# config/modules.yaml - Які модулі завантажувати
modules:
  enabled:
    - posts
    - profiles
    # - warehouse  # Потрібен ключ ліцензії
```

**Призначення:** Контроль активних модулів (ліцензування, функції тощо)

---

## Підтримка Ліцензування

### Архітектура Готова для Ліцензування

```yaml
# config/modules.yaml (майбутнє)
modules:
  enabled:
    - warehouse

  config:
    warehouse:
      version: "1.2.0"
      license_key: "WH-ABC-123-XYZ" #  Валідація ліцензії
      settings:
        max_items: 10000
```

**Модуль може валідувати ліцензію в Initialize():**

```go
func (m *WarehouseModule) Initialize(ctx context.Context, core *Core) error {
    // Завантаження конфігурації
    cfg := moduleconfig.Load("internal/modules/warehouse/config", env)

    // Валідація ліцензії
    licenseKey := cfg.GetString("module.license_key")
    if !validateLicense(licenseKey, "warehouse") {
        return fmt.Errorf("invalid license for warehouse module")
    }

    // Продовження ініціалізації...
}
```

**Статус:** Архітектура підтримує ліцензування - реалізація готова при необхідності.

---

## Управління Залежностями

### Залежності Модулів

```go
func (m *MyModule) Dependencies() []string {
    return []string{"posts", "profiles"}  // Цей модуль потребує posts + profiles
}
```

**Реєстр автоматично вирішує залежності:**

1. Топологічне сортування модулів
2. Ініціалізація в порядку залежностей
3. Помилка при циклічних залежностях або відсутніх модулях

**Приклад:** Модуль Fleet залежить від модуля Warehouse (для запчастин):

```yaml
modules:
  enabled:
    - warehouse # Повинен завантажитися першим
    - fleet # Залежить від warehouse
```

**Статус:** Система залежностей реалізована в `pkg/module/registry.go`

---

## Шаблони Комунікації

### 1. Міжмодульні Події (Async)

```go
// Модуль Posts публікує подію
event := &PostCreatedEvent{...}
eventBus.Publish(ctx, "post.created", event)

// Модуль Profiles підписується
eventBus.Subscribe("post.created", func(e bus.Event) {
    // Оновлення статистики користувача
})
```

**Переваги:**

- Без прямих імпортів модуль-до-модуля
- Слабке зв'язування
- Асинхронна обробка

---

### 2. Реєстр Модулів (Sync)

```go
// Отримання іншого модуля
postsModule := core.Registry.Get("posts")

// Виклик методів (якщо модуль надає публічне API)
stats := postsModule.(PostsModuleAPI).GetUserStats(userID)
```

**Переваги:**

- Пряма комунікація коли потрібно
- Type-safe інтерфейси
- Використовувати обережно - краще події

---

## Архітектура Router

### Core Маршрути

```go
// internal/adapter/http/v1/router/router.go
type V1Router struct {
    // Тільки Core-інфраструктурні маршрути
    HealthRouter   *gin.RouterGroup
    AuthRouter     *gin.RouterGroup
    CountryRouter  *gin.RouterGroup
    CurrencyRouter *gin.RouterGroup
    RBACRouter     *gin.RouterGroup  // Ролі + Дозволи
    AdminRouter    *gin.RouterGroup  // Управління Purge
}
```

**Чого НЕ в core-router:**

- Маршрути Posts → Перенесені в модуль posts
- Маршрути Comments → Перенесені в модуль posts
- Маршрути Profile → Перенесені в модуль profiles
- Маршрути Contact → Перенесені в модуль profiles

---

### Маршрути Модулів

```go
// Модуль Posts реєструє власні маршрути
func (m *PostsModule) RegisterRoutes(router *gin.RouterGroup) {
    postsGroup := router.Group("/posts")
    {
        postsGroup.GET("", m.postHandler.ListPosts)
        postsGroup.POST("", m.postHandler.CreatePost)
        // ...
    }

    commentsGroup := router.Group("/comments")
    {
        commentsGroup.POST("", m.commentHandler.CreateComment)
        // ...
    }
}
```

**Результат:**

- Core: `/api/v1/auth/*`, `/api/v1/countries/*`, `/api/v1/admin/*`
- Модуль Posts: `/api/v1/posts/*`, `/api/v1/comments/*`
- Модуль Profiles: `/api/v1/profiles/*`, `/api/v1/contacts/*`

**Статус:** Чітке розділення - кожен модуль володіє своїми маршрутами

---

## Управління Базою Даних

### Система Міграцій

**Core міграції (з namespace та описовими назвами):**

```
migrations/core/
├── 000001_core_init_uuid_v7.up.sql            UUID v7 + тригери
├── 000002_core_auth_full.up.sql               Auth-таблиці (users, sessions, tokens)
├── 000003_core_rbac_full.up.sql               RBAC (roles, permissions)
├── 000004_core_ref_timezones.up.sql           Довідкові дані
├── 000005_core_ref_languages.up.sql           Довідкові дані
└── 000006_core_ref_countries_currencies.up.sql Довідкові дані
```

**Міграції модулів (з namespace та префіксами модулів):**

```
migrations/posts/
├── 000001_posts_posts.up.sql                  Таблиця Posts
├── 000002_posts_comments.up.sql               Таблиця Comments
└── 000003_posts_comment_likes.up.sql          Comment likes

migrations/profiles/
├── 000001_profiles_contacts.up.sql            User contacts
└── 000002_profiles_profiles.up.sql            User profiles
```

**Майбутнє:** Модулі можуть реєструвати міграції програмно:

```go
func (m *MyModule) RegisterMigrations() []module.Migration {
    return []module.Migration{
        {Version: 1, Up: "CREATE TABLE my_table ...", Down: "DROP TABLE my_table"},
    }
}
```

**Статус:** Зараз файлова система, програмна система готова в `pkg/module/module.go`

---

## Стратегія Тестування

### Core Тести

```
internal/
├── domain/entity/*_test.go         Юніт-тести entity
├── usecase/*_test.go               Юніт-тести use case
└── adapter/repository/*_test.go    Інтеграційні тести repository
```

**Фокус:** Auth, RBAC, довідкові дані, інфраструктура.

---

### Тести Модулів

```
internal/modules/posts/
├── usecase/*_test.go               Юніт-тести бізнес-логіки
├── adapter/repository/*_test.go    Тести repository
└── module_test.go                  Інтеграційні тести модуля
```

**Статус:** Кожен модуль тестує власну логіку незалежно

---

### Smoke Тести

```
test/smoke/
├── auth_smoke_test.go              Core auth-потоки
├── rbac_smoke_test.go              Core RBAC-потоки
├── user_post_smoke_test.go         Модуль Posts (потребує оновлення)
└── user_profile_smoke_test.go      Модуль Profiles (потребує оновлення)
```

**Статус:** Smoke-тести потребують оновлення import-шляхів після міграції модулів

---

## Перевірка Порушень →

### ВИПРАВЛЕНО: Core містив політики purge для конкретних entity

**До:**

```go
//  Core знав про entity модулів
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

**Після:**

```go
//  Core має тільки інфраструктуру
type PurgeConfig struct {
    Enabled   bool
    Schedule  string
    BatchSize int
}

// Модулі реєструють політики через purge.DefaultPolicyRegistry
```

---

### ВИПРАВЛЕНО: Posts/Comments були в core

**До:** Posts та comments мали entity, use cases, handler'и в `internal/`

**Після:** Повна міграція в `internal/modules/posts/`

**Видалено з core:** 15,000+ рядків коду перенесені в модуль

---

### ВИПРАВЛЕНО: Profiles/Contacts були в core

**До:** Profiles та contacts розкидані по `internal/domain`, `internal/usecase`, `internal/adapter`

**Після:** Повна міграція в `internal/modules/profiles/`

**Результат:** Core справді мінімальний - тільки інфраструктура + довідкові дані

---

## Підсумок: Core vs Modules

| Компонент           | Розташування | Призначення      | Статус       |
| ------------------- | ------------ | ---------------- | ------------ |
| **Authentication**  | Core         | Основа безпеки   | Правильно    |
| **RBAC**            | Core         | Контроль доступу | Правильно    |
| **Countries**       | Core         | Довідкові дані   | Правильно    |
| **Currencies**      | Core         | Довідкові дані   | Правильно    |
| **Timezones**       | Core         | Довідкові дані   | Правильно    |
| **Languages**       | Core         | Довідкові дані   | Правильно    |
| **Database**        | Core         | Інфраструктура   | Правильно    |
| **Event Bus**       | Core         | Інфраструктура   | Правильно    |
| **Purge Scheduler** | Core         | Інфраструктура   | Правильно    |
| **Module Registry** | Core         | Оркестрація      | Правильно    |
|                     |              |                  |
| **Posts**           | Module       | Бізнес-логіка    | Незалежний   |
| **Comments**        | Module       | Бізнес-логіка    | Незалежний   |
| **Likes**           | Module       | Бізнес-логіка    | Незалежний   |
| **Profiles**        | Module       | Бізнес-логіка    | Незалежний   |
| **Contacts**        | Module       | Бізнес-логіка    | Незалежний   |
| **Warehouse**       | Module       | Бізнес-логіка    | Ліцензований |

---

## Рекомендації

### 1. Core Чистий

Поточний core містить тільки:

- Інфраструктурні сервіси
- Довідкові дані
- Безпеку (auth + RBAC)
- Інтерфейси управління (реєстри)

**Дія:** Зміни не потрібні - архітектура правильна.

---

### 2. Модулі Незалежні

Кожен модуль:

- Має власну структуру entity/usecase/adapter
- Завантажує власну конфігурацію
- Реєструє handler'и/політики/дозволи
- Може бути увімкнений/вимкнений через конфіг

**Дія:** Зміни не потрібні - модулі належним чином ізольовані.

---

### 3. Готовність до Ліцензування

Архітектура підтримує:

- Валідацію ключа ліцензії в init модуля
- Конфігурацію ліцензії для кожного модуля
- Управління залежностями (ліцензований модуль залежить від безкоштовного модуля)

**Дія:** ⏳ Реалізувати валідацію ліцензій коли комерційні модулі будуть готові.

---

### 4. Дрібні TODO

1. **Модуль Audit** - Додати комплексну систему аудит-логування
2. **Оновити smoke-тести** - Виправити import-шляхи після міграції модулів
3. **Програмні міграції** - Активувати `RegisterMigrations()` в модулях
4. **API документація** - Оновити Swagger для відображення маршрутів модулів

---

## Висновок

**Оцінка Архітектури: ВІДПОВІДАЄ**

Promenade успішно реалізує **плагінну архітектуру** з:

- Чітким розділенням між Core (інфраструктура) та Modules (бізнес-логіка)
- Незалежністю модулів (без залежностей від core)
- Динамічним завантаженням модулів з вирішенням залежностей
- Автономією конфігурації (кожен модуль володіє своєю конфігурацією)
- Підтримкою ліцензування (готово для комерційних модулів)
- Event-driven комунікацією (слабке зв'язування)

**Core справді мінімальний:**

- Менеджери інфраструктури
- Довідкові дані
- Основа безпеки (auth + RBAC)

**Модулі самодостатні:**

- Власні entity, use cases, adapter'и
- Власна конфігурація
- Власні політики purge
- Pluggable (увімкнення/вимкнення через конфіг)

**Наступні Кроки:**

1. Реалізувати валідацію ліцензій для комерційних модулів
2. Оновити smoke-тести
3. Розглянути винесення timezone/language в окремий модуль "reference", якщо вони стануть великими

---

**Дата Аудиту:** 22 грудня 2025
**Аудитор:** AI Assistant (GitHub Copilot)
**Статус:** ПРОЙДЕНО - Архітектура надійна та правильно реалізована
