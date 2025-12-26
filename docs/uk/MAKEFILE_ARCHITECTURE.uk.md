[ English](../MAKEFILE_ARCHITECTURE.md) |  **Українська** | [ Deutsch](MAKEFILE_ARCHITECTURE.de.md) | [ Português](MAKEFILE_ARCHITECTURE.pt.md) | [ Español](MAKEFILE_ARCHITECTURE.es.md)

---

# Архітектура Makefile

Модульна система Makefile для чіткого розділення відповідальності та масштабованості.

## Структура

```
Makefile             (63 рядки)  - Головний файл: змінні, завантаження env, help
Makefile.dev.mk      (64 рядки)  - Робочий процес розробки
Makefile.test.mk     (57 рядків)  - Тестова інфраструктура
Makefile.prod.mk     (90 рядків)  - Виробничі/DevOps операції

Всього:              274 рядки
```

## Філософія

**Головний Makefile** - Тільки загальні речі:

- Завантаження змінних середовища (`.env.development`)
- Спільні змінні (`APP_NAME`, `VERSION`, `DB_URL`, `MIGRATE`)
- Включення модулів (`include Makefile.*.mk`)
- Згрупована команда help (показує всі модулі)

**Makefile.dev.mk** - Робочий процес розробника:

```bash
make install             # Встановити інструменти (swag, migrate, golangci-lint)
make dev                 # Запустити dev сервер (postgres + migrations + app)
make build               # Зібрати бінарний файл (з генерацією swagger)
make run                 # Запустити скомпільований бінарний файл
make lint                # Запустити golangci-lint
make fmt                 # Форматувати код (go fmt + gofmt -s)
make deps-update         # Оновити залежності
make config-show         # Показати поточну конфігурацію env
```

**Makefile.test.mk** - Тестування:

```bash
make test                # Unit + integration тести
make test-unit           # Тільки unit тести (domain + usecase)
make test-integration    # Integration тести (реальна БД на порту 5433)
make test-smoke          # Smoke тести (end-to-end критичні потоки)
make test-coverage       # Згенерувати HTML звіт покриття
make test-db-start       # Запустити тестову базу даних
make test-db-stop        # Зупинити тестову базу даних
```

**Makefile.prod.mk** - DevOps операції:

```bash
# Docker
make docker-build        # Зібрати образ (VERSION=0.1.0 ENV=dev)
make docker-run          # Зібрати + запустити контейнери
make docker-up           # Запустити сервіси
make docker-down         # Зупинити сервіси
make docker-logs         # Переглянути логи
make docker-restart      # Перезапустити контейнери
make docker-ps           # Показати запущені контейнери
make docker-clean        # Видалити контейнери + volumes

# Migrations
make migrate-create      # Створити міграцію (NAME=xxx)
make migrate-up          # Застосувати міграції
make migrate-down        # Відкотити останню міграцію
make migrate-force       # Примусово встановити версію (VERSION=N)
make migrate-version     # Показати поточну версію
make migrate-status      # Показати статус

# Документація
make swagger-all         # Згенерувати v1 + v2 Swagger документацію

# Очищення
make clean               # Видалити артефакти (bin/, docs/, coverage)
```

## Переваги

1. **Модульність** - Кожен файл має одну відповідальність
2. **Масштабованість** - Легко додати `Makefile.{stage,ci,deploy}.mk`
3. **Читабельність** - Чітке розділення за контекстом
4. **Підтримуваність** - Малі сфокусовані файли замість 188-рядкового моноліта
5. **Зручність для команди** - Розробники/QA/DevOps бачать тільки релевантні команди

## Приклади використання

**Розробка:**

```bash
make help          # Побачити всі доступні команди
make dev           # Запустити розробку (найчастіше використовується)
make build         # Зібрати для локального тестування
make fmt lint      # Форматувати та перевірити перед комітом
```

**Тестування:**

```bash
make test          # Запустити повний набір тестів перед PR
make test-unit     # Швидкий зворотній зв'язок під час розробки
make test-smoke    # Перевірити критичні потоки після змін
```

**DevOps:**

```bash
make docker-run    # Розгорнути в локальний Docker
make migrate-up    # Застосувати міграції бази даних
make swagger-all   # Перегенерувати документацію API
make clean         # Чистий стан перед свіжим розгортанням
```

## Додавання нових команд

1. Визначте контекст: dev/test/prod
2. Редагуйте відповідний `Makefile.{context}.mk`
3. Додайте `## Коментар` для відображення в help
4. Запустіть `make help` для перевірки

Приклад:

```makefile
# У Makefile.dev.mk
watch: ## Спостерігати та перезавантажувати при зміні файлів
	air -c .air.toml
```

## Змінні

Всі спільні змінні знаходяться в головному `Makefile`:

- `APP_NAME` - Назва застосунку
- `VERSION` - Версія збірки (за замовчуванням: 0.1.0)
- `ENV` - Середовище (dev/test/prod)
- `DB_URL` - Рядок підключення PostgreSQL
- `MIGRATE` - Команда migrate з DB URL
- `DOCKER_COMPOSE` - Команда Docker Compose

Перевизначення:

```bash
make docker-build VERSION=1.2.3 ENV=prod
make migrate-up DB_NAME=promenade_staging
```

## Міграція зі старої структури

До (моноліт 188 рядків):

```
Makefile  ← Все змішано разом
```

Після (модульна структура 274 рядки):

```
Makefile          ← Загальне (63 рядки)
Makefile.dev.mk   ← Розробка (64 рядки)
Makefile.test.mk  ← Тестування (57 рядків)
Makefile.prod.mk  ← Виробництво (90 рядків)
```

**Без breaking changes** - Всі команди працюють точно так само!
