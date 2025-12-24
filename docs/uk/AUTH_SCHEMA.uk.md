# Архітектура Системи Автентифікації

🇬🇧 [English](../AUTH_SCHEMA.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/AUTH_SCHEMA.de.md) | [🇵🇹 Português](../pt/AUTH_SCHEMA.pt.md) | [🇪🇸 Español](../es/AUTH_SCHEMA.es.md)

Повний довідник системи автентифікації в Promenade. Охоплює реєстрацію користувачів, вхід, управління сесіями, обробку токенів та механізми безпеки.

## Зміст

- [Огляд](#огляд)
- [Схема Бази Даних](#схема-бази-даних)
- [Стани та Життєвий Цикл Користувача](#стани-та-життєвий-цикл-користувача)
- [Потік Автентифікації](#потік-автентифікації)
- [Управління Сесіями](#управління-сесіями)
- [Система Токенів](#система-токенів)
- [Механізми Безпеки](#механізми-безпеки)
- [API Ендпоінти](#api-ендпоінти)
- [Обробка Помилок](#обробка-помилок)
- [Конфігурація](#конфігурація)

---

## Огляд

**Компоненти Системи Автентифікації:**

| Компонент       | Призначення                             | Технологія          |
| --------------- | --------------------------------------- | ------------------- |
| User Entity     | Основний обліковий запис з паролем      | PostgreSQL, bcrypt  |
| Session Entity  | Зберігання refresh токенів              | PostgreSQL, SHA-256 |
| JWT Manager     | Генерація/валідація access токенів      | HMAC-SHA256         |
| Auth UseCase    | Бізнес-логіка всіх операцій авторизації | Go                  |
| Auth Middleware | Автентифікація запитів                  | Gin middleware      |
| Event Bus       | Асинхронні сповіщення (email, журнали)  | Memory/Redis        |

**Ключові Функції:**

-  JWT-автентифікація (access + refresh токени)
-  Ротація refresh токенів (найкраща практика безпеки)
-  Управління конкурентними сесіями (макс. 5 на користувача)
-  Машина станів користувача (unverified → active → suspended/banned)
-  Хешування паролів з bcrypt (складність 10)
-  Хешування refresh токенів з SHA-256
-  Автоматичне очищення сесій при зміні стану
-  Асинхронні email сповіщення через шину подій

---

## Схема Бази Даних

### Таблиця `core_users`

```sql
CREATE TABLE core_users (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    email              VARCHAR(255) NOT NULL UNIQUE,
    name               VARCHAR(255) NOT NULL,
    password           VARCHAR(255) NOT NULL,  -- bcrypt хеш
    status             VARCHAR(20) NOT NULL DEFAULT 'unverified',
    email_verified_at  TIMESTAMPTZ,
    suspended_reason   TEXT,
    suspended_until    TIMESTAMPTZ,
    last_login_at      TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON core_users(email);
CREATE INDEX idx_users_status ON core_users(status);
```

### Таблиця `core_user_sessions`

```sql
CREATE TABLE core_user_sessions (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id       UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,  -- SHA-256 хеш
    user_agent    TEXT,
    ip_address    VARCHAR(45),
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON core_user_sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON core_user_sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON core_user_sessions(expires_at);
```

**Додаткові Таблиці (ще не реалізовані):**

- `core_password_reset_tokens` - Процес скидання пароля
- `core_email_verification_tokens` - Процес верифікації email
- `core_login_attempts` - Захист від brute force

---

## Стани та Життєвий Цикл Користувача

### Перелік Статусів Користувача

```go
type UserStatus string

const (
    UserStatusUnverified UserStatus = "unverified" // Зареєстрований, але email не підтверджено
    UserStatusActive     UserStatus = "active"     // Email підтверджено і акаунт активний
    UserStatusSuspended  UserStatus = "suspended"  // Тимчасово заблокований (можна реактивувати)
    UserStatusBanned     UserStatus = "banned"     // Назавжди заблокований
    UserStatusInactive   UserStatus = "inactive"   // Деактивований користувачем (можна реактивувати)
)
```

### Переходи Станів

```
                    Register()
                        │
                        ▼
                 ──────────────
                 │  unverified  │  ──────────────
                 └──────────────                │
                        │                        │
                 VerifyEmail()              Login() дозволено
                        │                        │
                        ▼                        ▼
                 ──────────────         Користувач може увійти
                 │    active    │         (unverified або active)
                 └──────────────
                    │   │   │
        ───────────   │   └───────────
   Suspend()      Ban()           Deactivate()
        │               │                │
        ▼               ▼                ▼
 ───────────   ──────────    ─────────────
 │ suspended │   │  banned  │    │  inactive   │
 └───────────   └──────────    └─────────────
        │                              │
   Reactivate()                   Reactivate()
        │                              │
        └─────────────────────────────
                     │
                     ▼
              ──────────────
              │    active    │
              └──────────────
```

### Логіка CanLogin()

```go
func (u *User) CanLogin() bool {
    return u.Status == UserStatusActive || u.Status == UserStatusUnverified
}
```

**Дозволені Стани:**

-  `active` - Повний доступ
-  `unverified` - Може увійти, але функції можуть бути обмежені

**Заблоковані Стани:**

-  `suspended` - Повертає `ErrUserSuspended`
-  `banned` - Повертає `ErrUserBanned`
-  `inactive` - Повертає `ErrUserNotActive`

---

## Потік Автентифікації

### 1. Потік Реєстрації

```
Клієнт                  API                    UseCase                База Даних         Шина Подій
  │                      │                        │                        │                 │
  │  POST /auth/register │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │                      │  Register(email, name, pwd)                     │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  (не знайдено - OK)    │                 │
  │                      │                        │                        │                 │
  │                      │                        │  bcrypt.Hash(pwd)      │                 │
  │                      │                        │  user.Status = "unverified"              │
  │                      │                        │                        │                 │
  │                      │                        │  Create(user)          │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │                        │                 │
  │                      │                        │  Publish(UserRegisteredEvent)            │
  │                      │                        │─────────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  201 Created         │                        │                        │  EmailWorker    │
  │<─────────────────────│                        │                        │  надсилає       │
  │  {id, email, name}   │                        │                        │  welcome email  │
```

**Ключові Моменти:**

- Пароль хешується через `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)`
- Користувач створюється зі `status = "unverified"`
- UUID v7 генерується для ID користувача (впорядкований за часом)
- Подія публікується асинхронно - реєстрація не чекає на email
- Email worker обробляє подію `user.registered` у фоновому режимі

---

### 2. Потік Входу

```
Клієнт                  API                    UseCase                База Даних         Сесія
  │                      │                        │                        │                 │
  │  POST /auth/login    │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │  {email, password}   │  Login(email, pwd, ua, ip)                      │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  користувач знайдено   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  bcrypt.Compare(pwd, hash)               │
  │                      │                        │  OK                   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  YES                  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  JWTManager.GenerateAccessToken()        │
  │                      │                        │  crypto/rand 32 байти для refresh токену │
  │                      │                        │  SHA-256(refresh_token)                  │
  │                      │                        │                        │                 │
  │                      │                        │  CountUserSessions(user_id)              │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  count = 5 (ліміт!)    │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetOldestSession()    │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  Delete(oldest)        │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │                        │                 │
  │                      │                        │  Create(new session)   │                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │                        │  UpdateLastLogin()     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  200 OK              │  (access_token, refresh_token, user)            │                 │
  │<─────────────────────│                        │                        │                 │
  │  {access_token,      │                        │                        │                 │
  │   refresh_token,     │                        │                        │                 │
  │   user: {...}}       │                        │                        │                 │
```

**Ключові Моменти:**

- **Валідація пароля:** `bcrypt.CompareHashAndPassword(hash, password)`
- **Перевірка стану:** Має бути `active` або `unverified`
- **Access токен:** JWT з TTL 15 хвилин (за замовчуванням)
- **Refresh токен:** 32-байтний випадковий + base64, TTL 7 днів (за замовчуванням)
- **Зберігання refresh токену:** Хешується SHA-256 перед збереженням у БД
- **Ліміт сесій:** Макс. 5 конкурентних сесій на користувача
- **Видалення найстарішої:** Якщо ліміт перевищено, найстаріша сесія видаляється автоматично
- **Останній вхід:** Оновлюється асинхронно (не блокує відповідь)

---

### 3. Потік Оновлення Токена

```
Клієнт                  API                    UseCase                База Даних         Сесія
  │                      │                        │                        │                 │
  │  POST /auth/refresh  │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │  {refresh_token}     │  RefreshToken(token)   │                        │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  SHA-256(token)        │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetByRefreshToken(hash)                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │<───────────────────────────────────────│
  │                      │                        │  сесія знайдена        │                 │
  │                      │                        │                        │                 │
  │                      │                        │  session.IsExpired()?  │                 │
  │                      │                        │  NO                   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetByID(session.user_id)                │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  користувач знайдено   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  YES                  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GenerateAccessToken() │                 │
  │                      │                        │  crypto/rand новий refresh               │
  │                      │                        │  SHA-256(new_refresh)  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  Update(session)       │  ← РОТАЦІЯ ТОКЕНА
  │                      │                        │  - новий refresh hash  │                 │
  │                      │                        │  - нова expires_at     │                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  200 OK              │  (new_access, new_refresh)                      │                 │
  │<─────────────────────│                        │                        │                 │
  │  {access_token,      │                        │                        │                 │
  │   refresh_token}     │                        │                        │                 │
```

**Ключові Моменти:**

- **Ротація Токена:** Старий refresh токен інвалідується, видається новий
- **Безпека:** Одноразові refresh токени запобігають атакам повторного відтворення
- **Повторне використання сесії:** Оновлює існуючу сесію замість delete+create (продуктивність)
- **Перевірка терміну дії:** Прострочені сесії автоматично видаляються
- **Перевірка стану:** Користувач все ще має мати можливість увійти

---

### 4. Потік Виходу

```
Клієнт                  API                    UseCase                Сесія
  │                      │                        │                        │
  │  POST /auth/logout   │                        │                        │
  │─────────────────────>│                        │                        │
  │  {refresh_token}     │  Logout(token)         │                        │
  │                      │───────────────────────>│                        │
  │                      │                        │  SHA-256(token)        │
  │                      │                        │                        │
  │                      │                        │  GetByRefreshToken(hash)
  │                      │                        │───────────────────────>│
  │                      │                        │<───────────────────────│
  │                      │                        │  сесія знайдена        │
  │                      │                        │                        │
  │                      │                        │  Delete(session.id)    │
  │                      │                        │───────────────────────>│
  │                      │                        │                        │
  │                      │<───────────────────────│                        │
  │  200 OK              │                        │                        │
  │<─────────────────────│                        │                        │
  │  {message: "success"}│                        │                        │
```

**Ключові Моменти:**

- **Видалення сесії:** Негайно інвалідує refresh токен
- **Access токен:** Залишається дійсним до закінчення терміну (без стану JWT)
- **Відповідальність клієнта:** Клієнт має відкинути обидва токени

---

## Управління Сесіями

### Ліміт Конкурентних Сесій

```go
const MaxConcurrentSessions = 5
```

**Поведінка:**

- Користувач може мати максимум 5 активних сесій на різних пристроях
- При 6-му вході: найстаріша сесія автоматично видаляється
- Запобігає атакам необмеженої генерації токенів

### Сутність Session

```go
type Session struct {
    ID           UUID      `db:"id"`
    UserID       UUID      `db:"user_id"`
    RefreshToken string    `db:"refresh_token"`  // SHA-256 хешований
    UserAgent    *string   `db:"user_agent"`     // Інформація про браузер/пристрій
    IPAddress    *string   `db:"ip_address"`     // IP клієнта
    ExpiresAt    time.Time `db:"expires_at"`     // Абсолютний термін дії
    CreatedAt    time.Time `db:"created_at"`     // Початок сесії
}
```

### Операції Сесій

| Операція          | Призначення                        | Тригер                   |
| ----------------- | ---------------------------------- | ------------------------ |
| Create            | Нова сесія при вході               | Login                    |
| Update            | Ротація refresh токену             | Оновлення токена         |
| Delete            | Вихід з однієї сесії               | Logout                   |
| DeleteByUserID    | Інвалідація всіх сесій користувача | Suspend/Ban/Зміна пароля |
| GetOldestSession  | Знайти найстарішу для видалення    | Перевищення ліміту сесій |
| CountUserSessions | Перевірка ліміту                   | Login                    |

---

## Система Токенів

### JWT Access Токен

**Властивості:**

- **Алгоритм:** HS256 (HMAC-SHA256)
- **TTL:** 15 хвилин (за замовчуванням, налаштовується)
- **Зберігання:** Тільки на стороні клієнта (не в БД)
- **Валідація:** Перевірка підпису + терміну дії при кожному запиті

**Структура Claims:**

```go
type Claims struct {
    UserID uuidv7.UUID `json:"user_id"`
    Email  string      `json:"email"`
    jwt.RegisteredClaims
}

// RegisteredClaims включає:
// - iat (issued at - виданий)
// - exp (expires at - термін дії)
// - nbf (not before - не раніше)
```

**Приклад JWT:**

```json
{
  "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
  "email": "user@example.com",
  "iat": 1703347200,
  "exp": 1703348100,
  "nbf": 1703347200
}
```

### Refresh Токен

**Властивості:**

- **Алгоритм:** crypto/rand (32 байти) + base64
- **TTL:** 7 днів (за замовчуванням, налаштовується)
- **Зберігання:** База даних (SHA-256 хешований)
- **Валідація:** Порівняння хешу + термін дії

**Генерація:**

```go
func generateRefreshToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

**Хешування (SHA-256):**

```go
func hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return base64.URLEncoding.EncodeToString(hash[:])
}
```

### Патерн Ротації Токенів

**Перевага Безпеки:** Запобігає атакам повторного використання refresh токенів

1. Клієнт надсилає refresh токен
2. Сервер валідує та видає нову пару токенів
3. **Старий refresh токен інвалідується негайно**
4. Клієнт має використовувати новий refresh токен для наступного оновлення

**Запобігання Сценарію Атаки:**

-  Зловмисник краде refresh токен
-  Зловмисник намагається його використати
-  Токен вже оновлено легітимним користувачем → **Атака провалена**

---

## Механізми Безпеки

### Безпека Пароля

| Механізм  | Реалізація                        | Призначення                |
| --------- | --------------------------------- | -------------------------- |
| Хешування | `bcrypt` (складність 10)          | Одностороннє шифрування    |
| Сіль      | Автоматична (внутрішня bcrypt)    | Унікальний хеш для кожного |
| Валідація | `bcrypt.CompareHashAndPassword()` | Порівняння за сталий час   |

**Код:**

```go
// Хешування при реєстрації
func (u *User) HashPassword(password string) error {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedBytes)
    return nil
}

// Перевірка при вході
func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
```

### Безпека Refresh Токена

| Механізм            | Реалізація           | Призначення                    |
| ------------------- | -------------------- | ------------------------------ |
| Випадкова генерація | `crypto/rand` (32 б) | Криптографічно безпечний       |
| Хешування           | SHA-256              | Ніколи не зберігати plaintext  |
| Ротація токена      | Одноразові токени    | Інвалідація після використання |
| Термін дії          | TTL 7 днів           | Обмеження вікна впливу         |

### Безпека JWT

| Механізм       | Реалізація        | Призначення                      |
| -------------- | ----------------- | -------------------------------- |
| Підпис         | HMAC-SHA256       | Захист від підробки              |
| Секретний ключ | Змінна середовища | Перевірка підпису                |
| Короткий TTL   | 15 хвилин         | Мінімізація вікна впливу         |
| Без стану      | Без запиту до БД  | Продуктивність + масштабованість |

### Безпека Адміністративних Дій

**Автоматична Інвалідація Сесій:**

Коли адмін призупиняє/банить користувача:

1. Статус користувача змінюється в БД
2. **Всі сесії користувача видаляються негайно**
3. Користувач не може оновити токени
4. Існуючі access токени закінчуються природньо (макс. 15 хв)

```go
func (uc *authUseCase) SuspendUser(ctx context.Context, userID UUID, reason string, until *time.Time) error {
    // Оновити статус користувача
    if err := uc.userRepo.Suspend(ctx, userID, reason, until); err != nil {
        return err
    }

    // КРИТИЧНО: Інвалідувати всі сесії
    if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
        return err
    }

    // Асинхронне сповіщення
    uc.eventBus.Publish(ctx, bus.TopicUserSuspended, event)
    return nil
}
```

---

## API Ендпоінти

### Публічні Ендпоінти (Без Автентифікації)

#### POST /api/v1/auth/register

Реєстрація нового облікового запису.

**Запит:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Відповідь (201 Created):**

```json
{
  "status": "success",
  "data": {
    "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "unverified",
    "email_verified_at": null,
    "last_login_at": null,
    "created_at": "2024-12-23T10:00:00Z",
    "updated_at": "2024-12-23T10:00:00Z"
  }
}
```

---

#### POST /api/v1/auth/login

Автентифікація користувача та отримання токенів.

**Запит:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Відповідь (200 OK):**

```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "email": "user@example.com",
      "name": "John Doe",
      "status": "active"
    }
  }
}
```

---

#### POST /api/v1/auth/refresh

Оновлення access токена через refresh токен.

**Запит:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Відповідь (200 OK):**

```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "bmV3cmVmcmVzaHRva2VuaGVyZQ...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

**Примітка:** Старий refresh токен інвалідується (ротація токенів).

---

### Захищені Ендпоінти (Потрібна Автентифікація)

#### GET /api/v1/auth/me

Отримати профіль поточного користувача.

**Запит:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Відповідь (200 OK):**

```json
{
  "status": "success",
  "data": {
    "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "active",
    "email_verified_at": "2024-12-20T14:30:00Z",
    "last_login_at": "2024-12-23T10:00:00Z",
    "created_at": "2024-12-15T09:00:00Z",
    "updated_at": "2024-12-23T10:00:00Z"
  }
}
```

---

#### GET /api/v1/auth/sessions

Отримати всі активні сесії поточного користувача.

**Запит:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/sessions \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Відповідь (200 OK):**

```json
{
  "status": "success",
  "data": [
    {
      "id": "01936d6a-9999-7890-a1b2-c3d4e5f67890",
      "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...",
      "ip_address": "192.168.1.100",
      "expires_at": "2024-12-30T10:00:00Z",
      "created_at": "2024-12-23T10:00:00Z"
    },
    {
      "id": "01936d6a-8888-7890-a1b2-c3d4e5f67890",
      "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "user_agent": "Mobile Safari/537.36",
      "ip_address": "192.168.1.101",
      "expires_at": "2024-12-29T15:30:00Z",
      "created_at": "2024-12-22T15:30:00Z"
    }
  ]
}
```

---

#### POST /api/v1/auth/logout

Вихід та інвалідація refresh токена.

**Запит:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Відповідь (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "Successfully logged out"
  }
}
```

---

### Адміністративні Ендпоінти (Потрібні дозволи `users:suspend` / `users:ban`)

#### POST /api/v1/auth/users/:id/suspend

Тимчасово призупинити обліковий запис користувача.

**Запит:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/suspend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Порушення правил спільноти",
    "suspended_until": "2025-01-23T00:00:00Z"
  }'
```

**Відповідь (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User suspended successfully"
  }
}
```

**Наслідки:**

- Статус користувача змінено на `suspended`
- **Всі активні сесії видалено негайно**
- Користувач не може увійти до дати `suspended_until` або ручної реактивації

---

#### POST /api/v1/auth/users/:id/ban

Назавжди заблокувати обліковий запис користувача.

**Запит:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/ban \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Спам та шахрайська діяльність"
  }'
```

**Відповідь (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User banned successfully"
  }
}
```

**Наслідки:**

- Статус користувача змінено на `banned`
- **Всі активні сесії видалено негайно**
- Користувач не може увійти (потрібна ручна реактивація адміністратором)

---

#### POST /api/v1/auth/users/:id/reactivate

Реактивувати призупиненого/забаненого/неактивного користувача.

**Запит:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/reactivate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Відповідь (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User reactivated successfully"
  }
}
```

---

## Обробка Помилок

### Помилки Автентифікації

| Код Помилки           | HTTP Статус | Повідомлення               | Причина                              |
| --------------------- | ----------- | -------------------------- | ------------------------------------ |
| `INVALID_CREDENTIALS` | 401         | Invalid email or password  | Невірна комбінація email/пароль      |
| `EMAIL_EXISTS`        | 409         | Email already exists       | Реєстрація з існуючим email          |
| `USER_NOT_ACTIVE`     | 403         | User account is not active | Статус `inactive`                    |
| `USER_SUSPENDED`      | 403         | User account is suspended  | Статус `suspended`                   |
| `USER_BANNED`         | 403         | User account is banned     | Статус `banned`                      |
| `INVALID_TOKEN`       | 401         | Invalid or expired token   | Валідація токена провалена           |
| `TOKEN_EXPIRED`       | 401         | Token has expired          | JWT або refresh токен прострочений   |
| `UNAUTHORIZED`        | 401         | Authentication required    | Відсутній або невірний Authorization |

### Формат Відповіді Помилки

```json
{
  "status": "error",
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email or password",
    "details": null
  }
}
```

### Помилки Валідації

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "must be a valid email address"
      },
      {
        "field": "password",
        "message": "must be at least 8 characters"
      }
    ]
  }
}
```

---

## Конфігурація

### Змінні Оточення

```bash
# Конфігурація JWT
JWT_SECRET=ваш-256-біт-секретний-ключ-змініть-у-продакшені
JWT_ACCESS_TOKEN_TTL=15m    # Час життя access токена (напр., 15m, 1h)
JWT_REFRESH_TOKEN_TTL=168h  # Час життя refresh токена (напр., 168h = 7 днів)

# База даних
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=promenade
DB_SSLMODE=disable

# Сервер
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

### Файл Конфігурації (config/app.dev.yaml)

```yaml
jwt:
  secret: ${JWT_SECRET}
  access_token_ttl: 15m
  refresh_token_ttl: 168h

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  name: ${DB_NAME}
  sslmode: ${DB_SSLMODE}
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

server:
  port: ${SERVER_PORT}
  host: ${SERVER_HOST}
  read_timeout: 10s
  write_timeout: 10s
```

### Найкращі Практики Безпеки

| Налаштування            | Розробка      | Продакшен                    |
| ----------------------- | ------------- | ---------------------------- |
| `JWT_SECRET`            | Будь-який     | **256-біт випадковий рядок** |
| `JWT_ACCESS_TOKEN_TTL`  | 15m           | 5m - 15m                     |
| `JWT_REFRESH_TOKEN_TTL` | 168h (7 днів) | 7-30 днів                    |
| `DB_SSLMODE`            | disable       | **require або verify-full**  |
| `SERVER_HOST`           | 0.0.0.0       | 0.0.0.0 або конкретний IP    |

**Критичні Налаштування Продакшену:**

1. Згенеруйте безпечний JWT секрет: `openssl rand -base64 32`
2. Використовуйте тільки HTTPS (TLS/SSL сертифікати)
3. Увімкніть SSL бази даних (`DB_SSLMODE=require`)
4. Встановіть короткий TTL для access токена (5-15 хвилин)
5. Моніторте невдалі спроби входу (обмеження швидкості)

---

## Пов'язана Документація

- [AUTHORIZATION.md](AUTHORIZATION.uk.md) - Система дозволів RBAC (що відбувається ПІСЛЯ автентифікації)
- [CREDENTIALS.md](CREDENTIALS.uk.md) - Користувачі та ролі за замовчуванням для розробки/тестування
- [LOGGING.md](LOGGING.uk.md) - Структуроване логування з контекстом автентифікації
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.uk.md) - Архітектура системи та основні модулі
