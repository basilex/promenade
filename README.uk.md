# Promenade

[🇬🇧 English](README.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](README.de.md) | [🇵🇹 Português](README.pt.md) | [🇪🇸 Español](README.es.md)

> ** Примітка про переклади**: Деякі технічні документи з внутрішніх директорій (migrations/, internal/, pkg/, test/) поки доступні тільки англійською мовою. **Всі документи з docs/ повністю перекладені українською мовою.** Див. [docs/uk/INDEX.uk.md](docs/uk/INDEX.uk.md) для повного переліку доступних перекладів.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Готовий до продакшену REST API**, побудований на **Clean Architecture**, **модульній плагін-системі** та **namespace-based міграціях бази даних**. Включає автономні бізнес-модулі, PostgreSQL з UUID v7, event-driven архітектуру та комплексну тестову інфраструктуру.

---

## Огляд Архітектури

Promenade використовує **сувору шарову архітектуру**, де **Core оркеструє**, а **Модулі виконують** бізнес-логіку:

**Core Layer** - Оркестратор + Інфраструктура + Спільні Сервіси

- Аутентифікація та Авторизація (RBAC)
- Event IBus (Memory/Redis)
- База Даних та Транзакції
- Логування та Конфігурація
- Реєстр Модулів та Життєвий Цикл
- Довідкові Дані (країни, валюти, регіони, міста, часові зони, мови, способи оплати)

**IModule Layer** - Незалежні Вертикальні Зрізи (Бізнес-Домени)

| Модуль        | Сутності                     | Опис                            | Статус          |
| ------------- | ---------------------------- | ------------------------------- | --------------- |
| Posts         | пости, коментарі, лайки      | Контент створений користувачами | Безкоштовно     |
| Profiles      | контакти, профілі            | Профілі користувачів            | Безкоштовно     |
| **Analytics** | **метрики, звіти, дашборди** | **Аналітика та звітність**      | **Безкоштовно** |
| Warehouse     | продукти, інвентар           | Управління складом (майбутнє)   | В планах        |

Кожен модуль є самодостатнім з:

- Власними domain entities та бізнес-логікою
- Власними міграціями бази даних (namespace-based)
- Власними репозиторіями та use cases
- Власними HTTP handlers та routes
- Власною конфігурацією та життєвим циклом
- Опціонально: Політики очищення, дозволи, події

### Основні Принципи

1. **Core як Оркестратор**

   - Core **керує** життєвим циклом модулів (init, start, stop)
   - Core **надає** спільні сервіси (auth, events, DB, logging)
   - Core **знає КОЛИ** викликати модулі, але не **ЯК** вони працюють
   - Core **ніколи** не імпортує код специфічний для модулів

2. **Модулі як Виконавці**

   - Модулі **реалізують** domain-specific бізнес-логіку
   - Модулі **реєструються** через функції `init()`
   - Модулі є **незалежними** - можна вмикати/вимикати без впливу на інші
   - Модулі **ніколи** не імпортують код інших модулів (тільки `pkg/*`)
   - Кожен модуль = повний вертикальний зріз (entities → handlers)

3. **Шари Clean Architecture**

   ```
   Domain (сутності, інтерфейси) → Use Case (бізнес-логіка)
      ↓                                      ↓
   Adapter (репо, хендлери) → Infrastructure (DB, HTTP, події)
   ```

   - **Правило Залежностей**: Внутрішні шари ніколи не залежать від зовнішніх
   - Use cases залежать тільки від domain інтерфейсів, ніколи від конкретних реалізацій

4. **Namespace-Based Міграції**
   - Кожен namespace (core, posts, profiles) має незалежну історію версій
   - Міграції знаходяться в `migrations/{namespace}/NNNNNN_description.{up|down}.sql`
   - Core міграції виконуються першими, потім увімкнені модулі
   - Справжня автономія модулів - вмикай/вимикай без конфліктів схеми

---

## Швидкий Старт

### Передумови

- **Go 1.25+**
- **Docker & Docker Compose** (для PostgreSQL, Redis)
- **Make** (для автоматизації)

### 1. Клонування та Налаштування

```bash
git clone https://github.com/basilex/promenade.git
cd promenade

# Встановлення залежностей для розробки
make install

# Запуск PostgreSQL + Redis через Docker
make docker-up
```

### 2. Запуск Міграцій

Міграції виконуються **автоматично** при запуску додатку, але ви також можете запустити їх вручну:

```bash
# Перевірка статусу міграцій для всіх namespaces
make migrate-status

# Запуск всіх міграцій (core + увімкнені модулі)
make migrate

# Запуск конкретного namespace
make migrate-core                    # Тільки Core
make migrate-module MODULE=posts     # Конкретний модуль
```

**Повний гайд з міграцій**: див. розділ "Database Migrations" в [англійській версії README](README.md#database-migrations).

### 3. Запуск Додатку

```bash
# Режим розробки (hot reload, debug logging)
make dev

# Або зібрати та запустити бінарник
make build
./bin/promenade
```

Сервер запускається на **http://localhost:8081**

---

## Структура Документації

### Основна Документація

| Документ                                                                       | Опис                                                   |
| ------------------------------------------------------------------------------ | ------------------------------------------------------ |
| **[docs/uk/ARCHITECTURE_OVERVIEW.uk.md](docs/uk/ARCHITECTURE_OVERVIEW.uk.md)** | Візуальні діаграми архітектури, відповідальності шарів |
| **[docs/uk/ARCHITECTURE_QUICKREF.uk.md](docs/uk/ARCHITECTURE_QUICKREF.uk.md)** | Швидкий довідник, дерева рішень, поширені помилки      |
| **[docs/uk/ARCHITECTURE_AUDIT.uk.md](docs/uk/ARCHITECTURE_AUDIT.uk.md)**       | Аудит відповідності архітектури, чеклист перевірки     |

> 📖 **Додаткова документація**: internal/CORE.md та інші технічні документи доступні в [англійській версії](README.md).

### Система Модулів

| Документ                                                                                 | Опис                                               |
| ---------------------------------------------------------------------------------------- | -------------------------------------------------- |
| **[docs/uk/MODULE_DEVELOPMENT.uk.md](docs/uk/MODULE_DEVELOPMENT.uk.md)**                 | Створення нових модулів, кращі практики            |
| **[docs/uk/MODULE_INDEPENDENCE.uk.md](docs/uk/MODULE_INDEPENDENCE.uk.md)**               | Правила автономії модулів, управління залежностями |
| **[docs/uk/MODULE_CONFIG_ARCHITECTURE.uk.md](docs/uk/MODULE_CONFIG_ARCHITECTURE.uk.md)** | Система конфігурації модулів                       |

> 📖 **Додаткова документація**: internal/modules/ та інші технічні модулі описані в [англійській версії](README.md).

### Інфраструктура та Системи

| Документ                                                                 | Опис                                                   |
| ------------------------------------------------------------------------ | ------------------------------------------------------ |
| **[docs/uk/PURGE_ARCHITECTURE.uk.md](docs/uk/PURGE_ARCHITECTURE.uk.md)** | Автоматизована система очищення даних (registry-based) |

> 📖 **Технічна документація**: migrations/, pkg/bus/, test/ та інші технічні компоненти описані в [англійській версії](README.md).

### Гайди для Розробників

| Документ                                                                         | Опис                                    |
| -------------------------------------------------------------------------------- | --------------------------------------- |
| **[docs/uk/TESTING_GUIDE.uk.md](docs/uk/TESTING_GUIDE.uk.md)**                   | Кращі практики тестування, патерни      |
| **[docs/uk/TESTING_INFRASTRUCTURE.uk.md](docs/uk/TESTING_INFRASTRUCTURE.uk.md)** | Налаштування тестової інфраструктури    |
| **[docs/uk/MAKEFILE_ARCHITECTURE.uk.md](docs/uk/MAKEFILE_ARCHITECTURE.uk.md)**   | Система Makefile (модульна архітектура) |

> 📖 **Технічна документація**: test/ та інші інструменти розробки описані в [англійській версії](README.md).

### Технічні Довідники

| Документ                                                       | Опис                                               |
| -------------------------------------------------------------- | -------------------------------------------------- |
| **[docs/uk/UUID_V7_GUIDE.uk.md](docs/uk/UUID_V7_GUIDE.uk.md)** | Реалізація UUID v7 та переваги                     |
| **[docs/uk/SOFT_DELETE.uk.md](docs/uk/SOFT_DELETE.uk.md)**     | Патерн м'якого видалення для контенту користувачів |
| **[docs/uk/AUTHORIZATION.uk.md](docs/uk/AUTHORIZATION.uk.md)** | Система RBAC (4 ролі, wildcard дозволи)            |
| **[docs/uk/LOGGING.uk.md](docs/uk/LOGGING.uk.md)**             | Структуроване логування з slog                     |
| **[docs/uk/VALIDATION.uk.md](docs/uk/VALIDATION.uk.md)**       | Патерни валідації запитів                          |
| **[docs/uk/CREDENTIALS.uk.md](docs/uk/CREDENTIALS.uk.md)**     | Стандартні тестові користувачі та облікові дані    |
| **[docs/uk/INDEX.uk.md](docs/uk/INDEX.uk.md)**                 | Повний індекс документації з навчальними шляхами   |

---

## Система Модулів

### Доступні Модулі

#### **Модуль Posts** (`internal/modules/posts`)

Управління контентом створеним користувачами:

- **Сутності**: Пости, Коментарі, Лайки
- **Функції**: Створення/редагування постів, треди коментарів, система лайків
- **Міграції**: 3 міграції (namespace: `posts`)
- **Конфіг**: `config/modules.yaml` → `posts`

#### **Модуль Profiles** (`internal/modules/profiles`)

Управління профілями та контактами користувачів:

- **Сутності**: UserProfiles, UserContacts
- **Функції**: Управління профілями, контактна інформація
- **Міграції**: 2 міграції (namespace: `profiles`)
- **Конфіг**: `config/modules.yaml` → `profiles`

#### **Модуль Analytics** (`internal/modules/analytics`) - Безкоштовно

Аналітика, метрики та звітність:

- **Статус**: Безкоштовно - Доступний для всіх користувачів
- **Сутності**: Метрики, Звіти, Дашборди
- **Функції**: Збір метрик, кастомні звіти, візуальні дашборди
- **Міграції**: 1 міграція (namespace: `analytics`)
- **Призначення**: Бізнес-аналітика, моніторинг продуктивності, інсайти з даних

#### **Модуль Warehouse** (`internal/modules/warehouse`) - Майбутній Модуль

Управління інвентарем та продуктами (в планах):

- **Статус**: Заплановано - Структура існує як placeholder, ще не реалізована
- **Призначення**: E-commerce, інвентарні системи, роздрібна торгівля

### Структура Модуля

Кожен модуль має послідовну структуру:

```
internal/modules/{module}/
├── module.go           # Реєстрація модуля та життєвий цикл
├── domain/
│   └── entity/         # Domain сутності
├── repository/         # Інтерфейси доступу до даних та реалізації
├── usecase/            # Бізнес-логіка
├── adapter/
│   └── handler/        # HTTP хендлери та DTO
└── README.md           # Документація модуля
```

### Увімкнення/Вимкнення Модулів

Редагуйте `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts # Контент користувачів
    - profiles # Профілі + контакти користувачів
    - analytics # Бізнес аналітика
    # - warehouse  # Майбутнє: Управління складом
```

Модулі завантажуються автоматично при запуску додатку.

---

## Міграції Бази Даних

### Namespace-Based Система

Кожен namespace підтримує **незалежну історію версій**:

```
migrations/
├── core/               # Core інфраструктура (завжди виконується першою)
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000003_core_rbac_full.up.sql
│   ├── 000004_core_ref_timezones.up.sql
│   ├── 000005_core_ref_languages.up.sql
│   ├── 000006_core_ref_countries_currencies.up.sql    # 145 країн, 124 валюти
│   ├── 000007_core_ref_regions_cities.up.sql          # 30 регіонів, 17 міст
│   └── 000008_core_ref_payment_methods.up.sql         # 40+ способів оплати
├── posts/              # Міграції модуля Posts
│   ├── 000001_posts_posts.up.sql
│   ├── 000002_posts_comments.up.sql
│   └── 000003_posts_comment_likes.up.sql
├── profiles/           # Міграції модуля Profiles
│   ├── 000001_profiles_contacts.up.sql
│   └── 000002_profiles_profiles.up.sql
└── analytics/          # Міграції модуля Analytics
    └── 000001_analytics_tables.up.sql
```

### Команди Міграцій

```bash
# Статус для всіх namespaces
make migrate-status

# Запуск всіх (core + увімкнені модулі)
make migrate

# Запуск конкретного namespace
make migrate-core
make migrate-module MODULE=posts

# Відкат
make migrate-rollback MODULE=posts STEPS=1

# Створити нову міграцію
make migrate-create MODULE=posts NAME=add_post_views
make migrate-create-core NAME=add_audit_log
```

**Авто-міграції**: Міграції виконуються автоматично при запуску додатку (спочатку core, потім увімкнені модулі).

---

## Аутентифікація та Авторизація

### Стандартні Тестові Користувачі

| Email                           | Пароль     | Роль      | Дозволи                            |
| ------------------------------- | ---------- | --------- | ---------------------------------- |
| `system@promenade.com`          | `passw0rd` | Admin     | Повний доступ (`*`)                |
| `admin@promenade.com`           | `passw0rd` | Admin     | Управління користувачами/контентом |
| `moderator@promenade.com`       | `passw0rd` | Moderator | Модерація контенту                 |
| `alexander.vasilenko@gmail.com` | `03041965` | User      | Базові операції                    |

**Змініть паролі перед розгортанням у продакшен!**

### Система RBAC

- **4 Системні Ролі**: Admin, Moderator, User, Guest
- **Wildcard Дозволи**: `posts:*` (всі дії з постами), `*` (повний доступ)
- **Формат Ресурс-Дія**: `posts:create`, `users:delete`, `comments:moderate`

**Повний RBAC гайд**: [docs/uk/AUTHORIZATION.uk.md](docs/uk/AUTHORIZATION.uk.md)

---

## Тестування

**400+ тестів** на всіх шарах (100% успішно, ~20 секунд):

```bash
# Запуск всіх тестів
make test               # Всі тести (~20с)

# Запуск за модулями
make test-core          # Тести ядра (275 тестів: 39 entity + 236 usecase)
make test-modules       # Всі тести модулів
make test-module-posts      # Тести модуля Posts (33 тести)
make test-module-profiles   # Тести модуля Profiles (21 тест)
make test-module-analytics  # Тести модуля Analytics (11 тестів)

# Звіт покриття
make test-coverage      # HTML звіт покриття
```

### Покриття Тестами

- **Шар Core**: 275 тестів
  - Доменні сутності: 39 тестів (Country, Currency, Language, Timezone, Permission, Role, User, Session, Purge)
  - Use cases: 236 тестів (Auth, RBAC, CRUD довідкових даних, Операції очищення)
- **Модуль Posts**: 33 тести, 83.3% покриття (PostStatus, життєвий цикл UserPost, валідація, генерація slug)
- **Модуль Profiles**: 21 тест, 80.4% покриття (UserContact, UserProfile, приватність, валідація)
- **Модуль Analytics**: 11 тестів (Metrics, MetricAggregate, операції usecase)
- **Утиліти**: 51 тест, 89.5% середнє покриття (response 100%, validator 80%, logger 83.8%, pagination 94.1%)

### Час Виконання Тестів

- **Тести Core**: 14.7с (entity 4.4с + usecase 10.2с)
- **Модуль Posts**: 1.4с
- **Модуль Profiles**: 1.5с
- **Модуль Analytics**: 2.7с (entity 1.4с + usecase 1.4с)
- **Всього**: ~20 секунд для повного набору тестів

**Гайди з тестування**:

- [docs/uk/TESTING_GUIDE.uk.md](docs/uk/TESTING_GUIDE.uk.md) - Кращі практики

---

## Event IBus

**Dual-adapter event bus** для асинхронних операцій:

### Memory Адаптер

- In-memory Pub/Sub (goroutines + channels)
- **Призначення**: Розробка, тестування, single-instance розгортання
- **Переваги**: Без залежностей, швидкий, простий
- **Конфіг**: `BUS_ADAPTER=memory` (за замовчуванням)

### Redis Адаптер

- Розподілений Pub/Sub через Redis
- **Призначення**: Продакшен multi-instance розгортання
- **Переваги**: Персистентний, масштабований, відмовостійкий
- **Конфіг**: `BUS_ADAPTER=redis` + налаштування підключення Redis
- **Fallback**: Автоматичний перехід на memory якщо Redis недоступний

### Приклад Використання

```go
// Публікація події
event := &UserRegisteredEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    UserID:    user.ID,
    Email:     user.Email,
}
eventBus.Publish(ctx, bus.TopicUserRegistered, event)

// Підписка на події
eventBus.Subscribe(ctx, bus.TopicUserRegistered, func(ctx context.Context, e bus.Event) error {
    evt := e.(*UserRegisteredEvent)
    // Надіслати вітальний email
    return emailService.SendWelcome(ctx, evt.Email)
})
```

---

## Команди Makefile

### Розробка

```bash
make dev                # Запуск dev сервера (hot reload)
make build              # Зібрати production бінарник
make run                # Запустити зібраний бінарник
make lint               # Запуск linter (golangci-lint)
make fmt                # Форматування коду
make config-show        # Показати YAML конфігурацію (використовуйте ENV=dev|test|prod)
```

### Тестування

```bash
make test                      # Всі тести (core + модулі)
make test-core                 # Тільки core тести (domain + usecase)
make test-modules              # Всі тести модулів
make test-module-posts         # Тести модуля Posts
make test-module-profiles      # Тести модуля Profiles
make test-coverage             # Генерація HTML звіту покриття
```

### База Даних

```bash
make migrate                   # Запуск всіх міграцій (core + увімкнені модулі)
make migrate-status            # Показати статус міграцій
make migrate-core              # Мігрувати тільки core
make migrate-module MODULE=posts          # Мігрувати конкретний модуль
make migrate-rollback MODULE=posts STEPS=1  # Відкат
make migrate-create MODULE=posts NAME=xxx  # Створити міграцію модуля
make migrate-create-core NAME=xxx          # Створити core міграцію
```

### Docker

```bash
make docker-up          # Запуск PostgreSQL + Redis
make docker-down        # Зупинка сервісів
make docker-clean       # Видалення контейнерів + volumes
make docker-logs        # Перегляд логів
```

### Swagger

```bash
make swagger-all        # Генерація API docs (v1 + v2)
make swagger-v1         # Генерація тільки v1 docs
make swagger-v2         # Генерація тільки v2 docs
```

---

## Docker

### Налаштування Розробки

```bash
# Запуск сервісів
make docker-up

# Перегляд логів
make docker-logs

# Зупинка сервісів
make docker-down

# Чиста установка (видалення volumes)
make docker-clean
```

### Сервіси

- **PostgreSQL 16**: Порт 5432, користувач `system`, база даних `promenade_dev`
- **Redis 7**: Порт 6379 (для розподіленого event bus)

---

## Ключові Технічні Особливості

### UUID v7 Первинні Ключі

Time-ordered UUID для **2x швидших вставок** ніж UUID v4 та кращої продуктивності B-tree.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    ...
);
```

[docs/uk/UUID_V7_GUIDE.uk.md](docs/uk/UUID_V7_GUIDE.uk.md)

### Патерн М'якого Видалення

Контент створений користувачами (пости, коментарі) використовує timestamp `deleted_at` для безпечного видалення.

```go
// Завжди фільтруйте м'яко видалені записи
WHERE deleted_at IS NULL
```

[docs/uk/SOFT_DELETE.uk.md](docs/uk/SOFT_DELETE.uk.md)

### Автоматизована Система Очищення

Registry-based система де модулі реєструють свої політики утримання:

```go
purge.DefaultPolicyRegistry.RegisterPolicy(purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
})
```

Core планувальник запускає purge jobs через cron. Core не знає про конкретні сутності.

[docs/uk/PURGE_ARCHITECTURE.uk.md](docs/uk/PURGE_ARCHITECTURE.uk.md)

### Структуроване Логування

Context-aware логування з `slog`:

```go
log := logger.FromContext(ctx)  // Включає request_id, user_id
log.Info("User registered", "email", user.Email)
```

[docs/uk/LOGGING.uk.md](docs/uk/LOGGING.uk.md)

---

## API Документація

### Swagger UI

- **v1 API**: http://localhost:8081/api/v1/docs/swagger/index.html
- **v2 API**: http://localhost:8081/api/v2/docs/swagger/index.html

### Health Check

```bash
curl http://localhost:8081/api/v1/health
```

Відповідь:

```json
{
  "status": "success",
  "data": {
    "status": "healthy",
    "database": "connected",
    "timestamp": "2025-12-22T16:40:00Z"
  }
}
```

### Версіювання API

- **v1**: Поточний стабільний API (`internal/adapter/http/v1`)
- **v2**: API наступного покоління (`internal/adapter/http/v2`)

Обидві версії мають:

- Ізольовані handlers та DTOs
- Окрему Swagger документацію
- Незалежну реєстрацію routes

---

## Навчальні Шляхи

### Для Нових Розробників

1. **Старт**: [docs/uk/ARCHITECTURE_QUICKREF.uk.md](docs/uk/ARCHITECTURE_QUICKREF.uk.md) - 15-хвилинний огляд
2. **Практика**: Створіть простий модуль слідуючи [docs/uk/MODULE_DEVELOPMENT.uk.md](docs/uk/MODULE_DEVELOPMENT.uk.md)

### Для DevOps/Розгортання

4. **Конфігурація**: [docs/uk/MODULE_CONFIG_ARCHITECTURE.uk.md](docs/uk/MODULE_CONFIG_ARCHITECTURE.uk.md)

> 📖 **Додаткові шляхи навчання**: Технічні шляхи (DevOps, інфраструктура, інструменти) описані в [англійській версії](README.md).

### Для Архітекторів

1. **Огляд Архітектури**: [docs/uk/ARCHITECTURE_OVERVIEW.uk.md](docs/uk/ARCHITECTURE_OVERVIEW.uk.md)
2. **Аудит та Верифікація**: [docs/uk/ARCHITECTURE_AUDIT.uk.md](docs/uk/ARCHITECTURE_AUDIT.uk.md)
3. **Незалежність Модулів**: [docs/uk/MODULE_INDEPENDENCE.uk.md](docs/uk/MODULE_INDEPENDENCE.uk.md)

**Повний індекс**: [docs/uk/INDEX.uk.md](docs/uk/INDEX.uk.md)

> 📖 **Додаткова інформація**: Технічні деталі (міграції, Docker, інфраструктура) доступні в [англійській версії](README.md).

---

## Внесок у Проєкт

1. Зробіть Fork репозиторію
2. Створіть гілку features (`git checkout -b feature/amazing-feature`)
3. Слідуйте архітектурним принципам (дивіться [docs/uk/ARCHITECTURE_QUICKREF.uk.md](docs/uk/ARCHITECTURE_QUICKREF.uk.md))
4. Пишіть тести (підтримуйте 100% pass rate)
5. Закомітьте зміни (`git commit -m 'Add amazing feature'`)
6. Запушіть до гілки (`git push origin feature/amazing-feature`)
7. Відкрийте Pull Request

---

## Ліцензія

Цей проєкт ліцензований під ліцензією MIT - дивіться файл [LICENSE](LICENSE) для деталей.

---

## Підтримка

- **Документація**: [docs/uk/INDEX.uk.md](docs/uk/INDEX.uk.md)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Email**: alexander.vasilenko@gmail.com

---

**Побудовано з Clean Architecture та Go** 🇺🇦
