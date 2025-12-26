# Посібник з тестування Promenade

[ English](../TESTING_GUIDE.md) |  **Українська** | [ Deutsch](../de/TESTING_GUIDE.de.md) | [ Português](../pt/TESTING_GUIDE.pt.md) | [ Español](../es/TESTING_GUIDE.es.md)

## Огляд

Комплексна система тестування, що охоплює всі рівні додатку з **388 тестами (100% проходять)**:

- **Юніт-тести** (183) - ізольована бізнес-логіка та валідація сутностей
- **Інтеграційні тести** (91) - операції репозиторіїв з реальним PostgreSQL
- **Smoke-тести** (114) - наскрізні критичні сценарії з реальною базою даних
- **E2E-тести** - HTTP API тести (TODO)

## Швидкий старт

```bash
# Запустити всі тести (юніт + інтеграційні)
make test                  # 274 тести за ~41с

# Окремі набори тестів
make test-unit            # 183 юніт-тести (~5с)
make test-integration     # 91 інтеграційний тест (~36с)
make test-smoke           # 114 smoke-тестів (~4с)

# Покриття та моніторинг
make test-coverage        # HTML звіт про покриття
make test-watch           # Режим спостереження (gotestsum)
```

## Тестова база даних

Інтеграційні тести використовують окрему тестову базу даних на порту **5433**:

```bash
# Запустити тестову БД
make test-db-start

# Зупинити та очистити
make test-db-stop

# Переглянути логи
make test-db-logs
```

**Важливо:** Тестова БД повністю ізольована від dev/prod баз даних.

## Структура тестів

### 1. Інтеграційні тести (Репозиторії)

Розташовані поруч з кодом: `internal/adapter/repository/postgres/*_test.go`

Приклад:

```go
func TestUserRepository_Create(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    repo := postgres.NewUserRepository(testDB.DB)
    ctx := context.Background()

    t.Run("creates user successfully", func(t *testing.T) {
        user := helpers.UserFixture()
        err := repo.Create(ctx, user)
        require.NoError(t, err)

        retrieved, err := repo.GetByID(ctx, user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Email, retrieved.Email)
    })
}
```

**Покриття:**

- [+] IUserRepository: Create, GetByID, GetByEmail, UpdateStatus, Suspend, Ban, Reactivate, VerifyEmail
- [+] ISessionRepository: Create, GetByID, GetByRefreshToken, GetUserSessions, DeleteByUserID, DeleteExpired

### 2. Юніт-тести (Випадки використання)

_TODO: Наступний крок_

Будуть тестувати бізнес-логіку з замокованими репозиторіями:

- Register
- Login
- RefreshToken
- Logout
- SuspendUser
- BanUser
- ReactivateUser
- ChangePassword

### 3. Smoke-тести (Наскрізні критичні сценарії)

Розташовані в: `test/smoke/*_smoke_test.go`

**114 smoke-тестів** перевіряють критичні користувацькі сценарії з реальними операціями бази даних.

Приклад:

```go
func TestAuth_SmokeTest(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping smoke test in short mode")
    }

    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    ctx := context.Background()

    t.Run("[+] Complete_auth_flow", func(t *testing.T) {
        // Register → Login → GetMe → Refresh → Logout
        user, err := authUC.Register(ctx, "test@example.com", "John", "password123")
        require.NoError(t, err)

        tokens, err := authUC.Login(ctx, "test@example.com", "password123", "test-device")
        require.NoError(t, err)
        assert.NotEmpty(t, tokens.AccessToken)
        assert.NotEmpty(t, tokens.RefreshToken)

        // Продовжуємо тестувати повний сценарій...
    })

    t.Logf("[SUCCESS] All auth smoke tests passed!")
}
```

**Запуск Smoke-тестів:**

```bash
# Всі smoke-тести
make test-smoke

# Конкретний smoke-тест
go test -v ./test/smoke -run TestRBAC_SmokeTest
go test -v ./test/smoke -run TestUserPost_SmokeTest

# Пропустити в short режимі
go test -short ./test/smoke  # Smoke-тести пропускаються
```

**Покриття за модулями:**

| Модуль           | Сценарії | Покриття                                                         |
| ---------------- | -------- | ---------------------------------------------------------------- |
| Auth             | 8        | Реєстрація, логін, сесії, оновлення, вихід                       |
| Country/Currency | 12       | Повні CRUD операції                                              |
| UserContact      | 11       | Email, телефон, telegram, основний, верифікація                  |
| UserPost         | 12       | Чернетка, публікація, featured, розклад, перегляди, пошук        |
| UserProfile      | 12       | Приватність, верифікація, бан/розбан, перегляди, пошук           |
| PostComment      | 13       | Нитки, відповіді, вкладені відповіді, м'яке видалення            |
| CommentLikes     | 5        | Вподобання/скасування, пагінація, продуктивність (100 перевірок) |
| RBAC             | 28       | Дозволи, ролі, wildcard, закінчення терміну                      |
| RBAC Integration | 13       | Реальні сценарії дозволів (модератор, адмін та ін.)              |
| **Всього**       | **114**  | **Всі тести проходять [+]**                                      |

**Ключові функції:**

- [+] Реальна інтеграція з PostgreSQL (порт 5433)
- [+] Перевірка критичних шляхів (CRUD сценарії)
- [+] Включені бенчмарки продуктивності
- [+] Швидке виконання (~4 секунди для 114 тестів)
- [+] 100% успішність

### 4. HTTP інтеграційні тести (Обробники)

_TODO: Після розширення smoke-тестів_

Тестування всіх HTTP ендпоінтів через реальний Gin роутер:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `GET /api/auth/me`
- тощо.

## Допоміжні функції тестування

### `test/helpers/database.go`

```go
// Підключитися до тестової БД
testDB := helpers.SetupTestDB(t)
defer testDB.Close()

// Очистити всі таблиці
testDB.CleanupTables(t)

// Транзакційний тест (авто відкат)
testDB.RunInTransaction(t, func(tx *sqlx.Tx) {
    // Ваш код з tx
})
```

### `test/helpers/fixtures.go`

```go
// Стандартний активний користувач
user := helpers.UserFixture()

// Неверифікований користувач
user := helpers.UnverifiedUserFixture()

// Призупинений користувач
user := helpers.SuspendedUserFixture()

// Заблокований користувач
user := helpers.BannedUserFixture()

// Користувацький користувач
user := helpers.UserFixture(func(u *entity.User) {
    u.Email = "custom@test.com"
    u.Status = entity.UserStatusInactive
})

// Сесія
session := helpers.SessionFixture(userID)

// Прострочена сесія
session := helpers.ExpiredSessionFixture(userID)
```

## Найкращі практики

### [+] Робіть

- Використовуйте `testify/require` для критичних перевірок (зупиняє тест)
- Використовуйте `testify/assert` для некритичних перевірок (продовжує тест)
- Завжди робіть очищення: `defer testDB.CleanupTables(t)`
- Тестуйте граничні випадки: прострочені сесії, заблоковані користувачі тощо
- Використовуйте fixtures для консистентних тестових даних

### [X] Не робіть

- Не використовуйте production БД для тестів
- Не створюйте залежності між тестами
- Не забувайте `defer testDB.Close()`
- Не хардкодьте тестові дані - використовуйте fixtures

## Інтеграція CI/CD

Тести готові до CI:

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    make test-db-start
    make test
    make test-db-stop
```

## Покриття

```bash
make test-coverage
open coverage.html
```

Мета: **>80% покриття** для критичних модулів (usecase, repository).

## Що далі

1. [+] Інтеграційні тести репозиторіїв - **ГОТОВО**
2. **TODO** Юніт-тести випадків використання з моками
3. **TODO** Інтеграційні тести HTTP обробників
4. **TODO** E2E тести для повних сценаріїв
5. **TODO** Тести продуктивності/бенчмарки

## Приклади

### Запуск конкретних тестів

```bash
# Один тестовий файл
go test -v ./internal/adapter/repository/postgres/user_repository_test.go

# Один тест
go test -v ./internal/adapter/repository/postgres -run TestUserRepository_Create

# З детектором гонок
go test -race ./...

# З покриттям
go test -cover ./internal/adapter/repository/postgres
```

### Налагодження тестів

```bash
# Детальний вивід
go test -v ./...

# З логами БД
make test-db-logs

# Перевірити стан БД під час тесту
docker exec -it promenade_test_db psql -U system -d promenade_test
```

## Виправлення проблем

### "connection refused"

```bash
make test-db-start
# Почекати 3-5 секунд для готовності БД
```

### "table does not exist"

```bash
make migrate-test-up
```

### "too many open connections"

```bash
make test-db-stop
make test-db-start
```

---

**Питання?** Перевірте `Makefile.test.mk` для всіх доступних команд.
