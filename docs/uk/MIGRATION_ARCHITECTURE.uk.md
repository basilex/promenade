# Архітектура IModule Migration

[ English](../MIGRATION_ARCHITECTURE.md) |  **Українська** | [ Deutsch](MIGRATION_ARCHITECTURE.de.md) | [ Português](MIGRATION_ARCHITECTURE.pt.md) | [ Español](MIGRATION_ARCHITECTURE.es.md)

---

## Проблема

Поточна система migration порушує незалежність модулів:

- Усі migration в одній папці `migrations/`
- Глобальна послідовна нумерація (000001, 000002, ...)
- Core migration змішані з module migration
- Немає можливості увімкнути/вимкнути module migration незалежно

## Рішення: Namespace-Based Migrations

### 1. Структура директорій

```
migrations/
 core/                          # Core infrastructure migrations
    000001_init_schema.up.sql
    000001_init_schema.down.sql
    000002_auth_tables.up.sql
    000002_auth_tables.down.sql
    ...

 posts/                         # Posts module migrations
    000001_create_posts.up.sql
    000001_create_posts.down.sql
    000002_create_comments.up.sql
    ...

 profiles/                      # Profiles module migrations
     000001_create_profiles.up.sql
     ...
```

**Кожен namespace має незалежну версійність:**

- Core: 1, 2, 3, 4, ...
- Posts: 1, 2, 3, ...
- Profiles: 1, 2, ...

---

### 2. Схема таблиці migration

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     BIGINT       NOT NULL,
    namespace   VARCHAR(50)  NOT NULL,  -- 'core', 'posts', 'profiles', etc.
    dirty       BOOLEAN      NOT NULL DEFAULT FALSE,
    applied_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (namespace, version)
);

CREATE INDEX idx_schema_migrations_namespace ON schema_migrations(namespace);
```

**Приклад даних:**

```
| version | namespace | dirty | applied_at          |
|---------|-----------|-------|---------------------|
| 1       | core      | false | 2024-12-22 10:00:00 |
| 2       | core      | false | 2024-12-22 10:00:01 |
| 3       | core      | false | 2024-12-22 10:00:02 |
| 1       | posts     | false | 2024-12-22 10:00:03 |
| 2       | posts     | false | 2024-12-22 10:00:04 |
| 1       | profiles  | false | 2024-12-22 10:00:05 |
```

---

### 3. Migration Manager (pkg/migration/)

#### Інтерфейс

```go
package migration

type Manager interface {
    // MigrateNamespace applies all pending migrations for a namespace
    MigrateNamespace(ctx context.Context, namespace string) error

    // MigrateAll applies all pending migrations (core + enabled modules)
    MigrateAll(ctx context.Context, enabledModules []string) error

    // Rollback rolls back N migrations for a namespace
    Rollback(ctx context.Context, namespace string, steps int) error

    // Version returns current version for a namespace
    Version(ctx context.Context, namespace string) (int, error)

    // Status returns migration status for all namespaces
    Status(ctx context.Context) (map[string]MigrationStatus, error)
}

type MigrationStatus struct {
    Namespace      string
    CurrentVersion int
    PendingCount   int
    Dirty          bool
}
```

#### Реалізація

```go
package migration

type manager struct {
    db            *sqlx.DB
    migrationsDir string  // "migrations/" by default
}

func NewManager(db *sqlx.DB, migrationsDir string) Manager {
    return &manager{db: db, migrationsDir: migrationsDir}
}

func (m *manager) MigrateNamespace(ctx context.Context, namespace string) error {
    // 1. Check if schema_migrations table exists
    // 2. Get current version for namespace
    // 3. Read migration files from migrations/{namespace}/
    // 4. Apply pending migrations in transaction
    // 5. Update schema_migrations table
}

func (m *manager) MigrateAll(ctx context.Context, enabledModules []string) error {
    // 1. Always migrate core first
    if err := m.MigrateNamespace(ctx, "core"); err != nil {
        return err
    }

    // 2. Migrate each enabled module
    for _, module := range enabledModules {
        if err := m.MigrateNamespace(ctx, module); err != nil {
            return err
        }
    }

    return nil
}
```

---

### 4. Інтеграція з IModule

IModule можуть опціонально надавати migration програмно:

```go
// pkg/module/module.go
type IModule interface {
    // ... existing methods ...

    // RegisterMigrations returns embedded migrations for this module
    // Return nil to use file-based migrations from migrations/{namespace}/
    RegisterMigrations() []Migration
}

type Migration struct {
    Version     int
    Description string
    Up          string  // SQL to apply
    Down        string  // SQL to rollback
}
```

Приклад у module:

```go
// internal/modules/posts/module.go
func (m *PostsModule) RegisterMigrations() []Migration {
    // Option 1: Return nil to use file-based migrations
    return nil

    // Option 2: Embed migrations in code (for libraries)
    return []Migration{
        {
            Version:     1,
            Description: "Create posts table",
            Up:          `CREATE TABLE user_posts (...)`,
            Down:        `DROP TABLE user_posts`,
        },
    }
}
```

---

### 5. Використання в main.go

```go
// cmd/api/main.go
func main() {
    // ... setup ...

    // Initialize migration manager
    migrationMgr := migration.NewManager(db, "migrations")

    // Get enabled modules from config
    enabledModules := cfg.Modules.Enabled  // ["posts", "profiles"]

    // Run migrations for core + enabled modules
    if err := migrationMgr.MigrateAll(ctx, enabledModules); err != nil {
        log.Fatal("Failed to run migrations", "error", err)
    }

    // ... continue startup ...
}
```

---

### 6. CLI команди

```bash
# Migrate core only
make migrate-core

# Migrate specific module
make migrate-module MODULE=posts

# Migrate all (core + enabled modules)
make migrate-all

# Rollback module migrations
make migrate-rollback MODULE=posts STEPS=1

# Show migration status
make migrate-status
```

**Makefile:**

```makefile
migrate-core:
	go run cmd/migrate/main.go up --namespace=core

migrate-module:
	go run cmd/migrate/main.go up --namespace=$(MODULE)

migrate-all:
	go run cmd/migrate/main.go up --all

migrate-rollback:
	go run cmd/migrate/main.go down --namespace=$(MODULE) --steps=$(STEPS)

migrate-status:
	go run cmd/migrate/main.go status
```

---

### 7. Реорганізація Migration

**Поточна структура (плоска - ЗАСТАРІЛА, використовуйте namespace):**

```
migrations/
 000001_init_schema_deps.up.sql         # core
 000002_create_auth_schema.up.sql       # core
 000003_create_countries_currencies.up.sql  # core
 000004_create_user_contacts.up.sql     # profiles module
 000005_create_user_profiles.up.sql     # profiles module
 000006_create_user_posts.up.sql        # posts module
 000007_create_post_comments.up.sql     # posts module
 000008_create_comment_likes_table.up.sql  # posts module
 000009_create_rbac_tables.up.sql       # core
 000014_create_timezones_table.up.sql   # core
 000015_create_languages_table.up.sql   # core
```

**Нова структура (namespace з описовими назвами):**

```
migrations/
 core/
    000001_core_init_uuid_v7.up.sql
    000001_core_init_uuid_v7.down.sql
    000002_core_auth_full.up.sql
    000002_core_auth_full.down.sql
    000003_core_rbac_full.up.sql            # was 000009
    000003_core_rbac_full.down.sql
    000004_core_ref_timezones.up.sql        # was 000014
    000004_core_ref_timezones.down.sql
    000005_core_ref_languages.up.sql        # was 000015
    000005_core_ref_languages.down.sql
    000006_core_ref_countries_currencies.up.sql  # was 000003
    000006_core_ref_countries_currencies.down.sql

 posts/
    000001_posts_posts.up.sql               # was 000006_create_user_posts
    000001_posts_posts.down.sql
    000002_posts_comments.up.sql            # was 000007_create_post_comments
    000002_posts_comments.down.sql
    000003_create_comment_likes.up.sql    # was 000008
    000003_create_comment_likes.down.sql

 profiles/
     000001_create_user_contacts.up.sql    # was 000004
     000001_create_user_contacts.down.sql
     000002_create_user_profiles.up.sql    # was 000005
     000002_create_user_profiles.down.sql
```

---

### 8. Переваги

**Незалежність IModule**

- Кожен module володіє своїми migration
- Увімкнення/вимкнення module без конфліктів migration
- Чіткі межі відповідальності

**Контроль версій**

- Кожен namespace має незалежну версійність
- Немає конфліктів глобальної нумерації
- Легко зрозуміти, на якій версії знаходиться module

**Гнучке розгортання**

- Розгортання лише з необхідними module
- Додавання нових module без зміни існуючих migration
- Незалежний rollback migration module

**Досвід розробника**

- Зрозуміло, куди додавати нові migration
- Не потрібно вгадувати наступний глобальний номер
- IModule-специфічні команди migration

---

### 9. Робочий процес Migration

#### Створення нової Migration

```bash
# Core migration
make migrate-create-core NAME=add_audit_tables

# IModule migration
make migrate-create MODULE=posts NAME=add_post_views
```

**Згенеровані файли:**

```
migrations/posts/
 000004_add_post_views.up.sql    # Auto-incremented
 000004_add_post_views.down.sql
```

#### Застосування Migration

```bash
# Development: Migrate everything
make migrate-all

# Production: Migrate core + specific modules
MODULES=posts,profiles make migrate-all

# Selective: Migrate only new module
make migrate-module MODULE=warehouse
```

---

### 10. Зворотна сумісність

Для існуючих розгортань:

1. **Одноразовий скрипт migration**, який:
   - Створює резервну копію поточної таблиці `schema_migrations`
   - Створює нову `schema_migrations` з namespace
   - Відображає старі версії на namespace версії
   - Позначає всі як застосовані

```sql
-- Backup
CREATE TABLE schema_migrations_backup AS SELECT * FROM schema_migrations;

-- Drop old table
DROP TABLE schema_migrations;

-- Create new table with namespace
CREATE TABLE schema_migrations (...);

-- Insert mapped versions
INSERT INTO schema_migrations (namespace, version, dirty, applied_at)
VALUES
    ('core', 1, false, NOW()),  -- was 000001
    ('core', 2, false, NOW()),  -- was 000002
    ('core', 3, false, NOW()),  -- was 000003
    ('profiles', 1, false, NOW()),  -- was 000004
    ('profiles', 2, false, NOW()),  -- was 000005
    ('posts', 1, false, NOW()),  -- was 000006
    ...
```

2. **Скрипт реорганізації migration**, який переміщує файли до папок namespace

---

### 11. План впровадження

**Фаза 1: Інфраструктура (Тиждень 1)**

1. Створити пакет `pkg/migration/`
2. Реалізувати інтерфейс `Manager`
3. Додати підтримку namespace до таблиці schema_migrations
4. Написати тести

**Фаза 2: CLI та інструменти (Тиждень 1)**

1. Створити CLI інструмент `cmd/migrate/main.go`
2. Додати команди Makefile
3. Оновити документацію

**Фаза 3: Migration (Тиждень 2)**

1. Реорганізувати існуючі migration у namespace
2. Створити migration зворотної сумісності
3. Оновити main.go для використання нового manager
4. Тестування на staging

**Фаза 4: Інтеграція з IModule (Тиждень 2)**

1. Додати `RegisterMigrations()` до інтерфейсу module
2. Оновити існуючі module
3. Документація та приклади

---

### 12. Альтернатива: Fork golang-migrate

Якщо ми хочемо використати бібліотеку `golang-migrate`:

```go
import "github.com/golang-migrate/migrate/v4"

// Custom source driver that reads from namespace folders
type NamespaceSource struct {
    namespace string
    basePath  string
}

func (s *NamespaceSource) First() (version uint, err error) {
    // Read from migrations/{namespace}/
}

// Register custom source
migrate.Register("namespace", &NamespaceSource{})
```

---

## Підсумок

**Найкраще рішення:** Власний migration manager з підтримкою namespace.

**Чому?**

- Повний контроль над логікою namespace
- Незалежна версійність module
- Легка реалізація увімкнення/вимкнення module
- Чіткі межі відповідальності
- Немає обмежень зовнішніх бібліотек

**Шлях міграції:**

1. Реалізувати `pkg/migration/` manager
2. Додати namespace до schema_migrations
3. Реорганізувати існуючі migration
4. Оновити код запуску
5. Задокументувати робочий процес

**Оцінка зусиль:** 2-3 дні для повної реалізації + тестування
