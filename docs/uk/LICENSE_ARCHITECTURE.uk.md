# Архітектура Системи Ліцензування

Promenade використовує систему ліцензування на основі підписів для комерційних модулів. Цей документ описує архітектуру, реалізацію та патерни використання.

---

## Огляд

### Принципи Дизайну

1. **Незалежність Модулів**: Кожен модуль керує власним ліцензуванням
2. **Безпека на Основі Підписів**: HMAC-SHA256 запобігає підробці
3. **Плавна Деградація**: Пільгові періоди для прострочених ліцензій
4. **Специфічність для Середовища**: Різні правила валідації для кожного середовища
5. **Зрозумілість для Людини**: Ліцензійні ключі структуровані та розбірні

### Формат Ліцензії

```
PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}
```

Приклад:

```
PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Компоненти:

- **Префікс**: Завжди `PROMENADE`
- **Модуль**: Назва модуля ВЕЛИКИМИ ЛІТЕРАМИ (ANALYTICS, WAREHOUSE, AUDITLOG)
- **Рівень**: Рівень ліцензії (BASIC, PRO, ENTERPRISE)
- **Термін**: Дата у форматі YYYYMMDD
- **Підпис**: HMAC-SHA256 перших 4 частин, base64 URL-кодування

---

## Архітектура

### Системні Компоненти

**Потік Валідації Ліцензій:**

1. **Запуск Додатку**
   - Реєстр Модулів (`pkg/module`) ініціалізується
2. **Для Кожного Увімкненого Модуля**
   - Викликається `IModule.Initialize()`
3. **Валідатор Ліцензій** (`module/license/`)
   - `Parse()` - Вилучення компонентів ліцензії
   - `Validate()` - Перевірка підпису та терміну
   - `HealthCheck()` - Постійна валідація
4. **Статус Ліцензії** (результат)
   - Дійсна - Модуль працює нормально
   - Прострочена (пільга) - Видано попередження
   - Недійсна - Модуль вимкнено

### Потік Валідації

1. **Розбір**: Вилучення компонентів з рядка ліцензії
2. **Перевірка Модуля**: Переконатися, що ліцензія відповідає назві модуля
3. **Перевірка Підпису**: Валідація HMAC-SHA256
4. **Перевірка Терміну**: Валідація дати + пільговий період
5. **Повернення Статусу**: Дійсна, прострочена (пільга) або недійсна

### Зберігання

Ліцензії зберігаються як змінні середовища:

- `{MODULE}_LICENSE_KEY`: Ліцензійний ключ
- `LICENSE_SECRET`: Секрет для перевірки підпису (продакшн)

Приклад:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret-key"
```

---

## Реалізація

### Інтеграція Модуля

Кожен комерційний модуль реалізує валідацію ліцензії у своєму методі `Initialize()`:

```go
func (m *AnalyticsModule) Initialize(cfg interface{}) error {
 // Завантажити конфігурацію модуля
 config, ok := cfg.(*Config)
 if !ok {
 return ErrInvalidConfig
 }

 // Валідувати ліцензію, якщо потрібно
 if config.LicenseRequired {
 if err := m.validateLicense(config); err != nil {
 return fmt.Errorf("license validation failed: %w", err)
 }
 }

 return nil
}

func (m *AnalyticsModule) validateLicense(config *Config) error {
 licenseKey := config.LicenseKey
 if licenseKey == "" {
 licenseKey = os.Getenv("ANALYTICS_LICENSE_KEY")
 }

 if licenseKey == "" {
 return ErrLicenseRequired
 }

 secret := os.Getenv("LICENSE_SECRET")
 if secret == "" {
 secret = "default-dev-secret"
 }

 license, err := ParseLicense(licenseKey)
 if err != nil {
 return err
 }

 validationOptions := ValidationOptions{
 ValidateExpiry: config.ValidateExpiry,
 ValidateSignature: config.ValidateSignature,
 GracePeriodDays: config.GracePeriodDays,
 }

 return license.Validate("ANALYTICS", secret, validationOptions)
}
```

### Валідація Ліцензії

Пакет license надає основну логіку валідації:

```go
// Parse вилучає компоненти з рядка ліцензії
func ParseLicense(licenseKey string) (*License, error)

// Validate перевіряє валідність ліцензії
func (l *License) Validate(
 expectedModule string,
 secret string,
 options ValidationOptions,
) error

// Generate створює нову ліцензію (для тестування/інструментів)
func GenerateLicense(
 module, tier, expiry, secret string,
) (string, error)
```

### Перевірки Здоров'я

Перевірка здоров'я кожного модуля включає статус ліцензії:

```go
func (m *AnalyticsModule) HealthCheck(ctx context.Context) error {
 if m.license != nil {
 if m.license.IsExpired() && !m.license.InGracePeriod(7) {
 return ErrLicenseExpired
 }
 }
 return nil
}
```

---

## Конфігурація

### Налаштування для Різних Середовищ

#### Розробка (`config.dev.yaml`)

```yaml
license_required: false # Опціонально для розробки
validate_expiry: false # Ігнорувати прострочення
validate_signature: true # Все ще перевіряти підписи
grace_period_days: 30 # Довгий пільговий період
```

#### Тестування (`config.test.yaml`)

```yaml
license_required: false # Без ліцензії в CI/CD
validate_expiry: false
validate_signature: false
grace_period_days: 0
```

#### Продакшн (`config.prod.yaml`)

```yaml
license_required: true # Обов'язково
validate_expiry: true # Строге прострочення
validate_signature: true # Повна безпека
grace_period_days: 7 # Обмежена пільга
validate_on_request: true # Опціональна валідація на запит
```

### Конфігурація Модуля

У `config/modules.yaml`:

```yaml
modules:
  analytics:
  enabled: true
  version: "1.0.0"
  description: "Advanced analytics and reporting"
  license_key: "" # Встановити через змінну середовища
  settings:
  metrics_retention_days: 90
  max_reports_per_user: 10
  max_dashboards_per_user: 5
```

---

## Рівні та Функції

### Рівень BASIC

- Базовий збір метрик (100 метрик/день)
- Базові звіти (JSON, CSV)
- 5 дашбордів на користувача
- 30-денне зберігання даних
- Підтримка email

### Рівень PRO

- Необмежені метрики
- Розширені звіти (PDF, XLSX)
- Необмежені дашборди
- 90-денне зберігання даних
- Заплановані звіти
- Пріоритетна підтримка

### Рівень ENTERPRISE

- Всі функції PRO
- Власне зберігання (1+ років)
- Мультитенантна підтримка
- Власний брендинг
- Доступ до API
- Виділена підтримка
- Розгортання на власних серверах

---

## Використання

### Генерація Ліцензій

Використовуйте наданий скрипт:

```bash
# Згенерувати PRO ліцензію дійсну 365 днів
./scripts/generate-license.sh analytics PRO 365

# Вивід:
# PROMENADE-ANALYTICS-PRO-20261231-K8mF3pL9qT2xN7vR4zW6yH1jC5eS8bD0aG3hM6nP9
```

Встановити змінну середовища:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
export LICENSE_SECRET="your-production-secret"
```

### Тестування Валідації Ліцензії

Модуль analytics включає комплексні тести:

```bash
# Запустити тести валідації ліцензій
go test ./internal/modules/analytics/license/... -v

# Тестові сценарії:
# - Розбір дійсної ліцензії
# - Виявлення невірного формату
# - Перевірка підпису
# - Обробка прострочення
# - Логіка пільгового періоду
# - Виявлення невідповідності модуля
```

### Інструмент Генерації Ліцензій

Збудувати генератор ліцензій:

```bash
make build-license-generator

# Або вручну:
go build -o ./bin/license-generator ./cmd/license-generator/main.go
```

Згенерувати ліцензію програмно:

```bash
./bin/license-generator \
 -module=ANALYTICS \
 -tier=PRO \
 -expiry=20261231 \
 -secret="your-secret-key"
```

---

## Безпека

### Підписи HMAC-SHA256

- **Алгоритм**: HMAC з SHA-256
- **Ключ**: Зберігається у змінній середовища `LICENSE_SECRET`
- **Кодування**: Base64 URL-кодування (безпечне для URL, без padding)
- **Перевірка**: Порівняння зі сталим часом для запобігання атакам по часу

### Управління Секретами

**Розробка**:

```bash
# Стандартний тестовий секрет (прийнятно для локальної розробки)
LICENSE_SECRET="default-dev-secret-change-in-production"
```

**Продакшн**:

```bash
# Згенерувати сильний секрет (32+ байти)
openssl rand -base64 32

# Зберігати у безпечному сховищі (AWS Secrets Manager, HashiCorp Vault тощо)
# Встановити через змінну середовища (ніколи не комітити в git)
```

### Стратегія Ротації

1. Згенерувати новий секрет
2. Підписати нові ліцензії новим секретом
3. Підтримувати обидва старий та новий секрети під час перехідного періоду
4. Відкинути старий секрет після пільгового періоду

---

## Обробка Помилок

### Типи Помилок

```go
var (
 ErrLicenseRequired = errors.New("license key is required")
 ErrInvalidFormat = errors.New("invalid license format")
 ErrInvalidSignature = errors.New("invalid license signature")
 ErrLicenseExpired = errors.New("license has expired")
 ErrModuleMismatch = errors.New("license module mismatch")
 ErrInvalidTier = errors.New("invalid license tier")
)
```

### Плавна Деградація

1. **Відсутня Ліцензія** (dev/test): Модуль завантажується зі зменшеними функціями
2. **Прострочена Ліцензія**: Пільговий період дозволяє продовження роботи з попередженнями
3. **Недійсна Ліцензія**: Модуль не ініціалізується в продакшні, логує в dev

### Повідомлення для Користувачів

```go
// Помилка продакшну (строго)
return fmt.Errorf("analytics module requires valid license: %w", err)

// Попередження розробки (дозволяюче)
logger.Warn("Analytics license validation failed, continuing in dev mode",
 "error", err)
```

---

## Стратегія Монетизації

### Поточний Стан

| Модуль        | Статус           | Рівень             | Причина                            |
| ------------- | ---------------- | ------------------ | ---------------------------------- |
| posts         | Безкоштовний     | N/A                | Основні соціальні функції          |
| profiles      | Безкоштовний     | N/A                | Основні функції користувача        |
| **analytics** | ** Комерційний** | **PRO/ENTERPRISE** | **Активний: Аналітика як преміум** |
| warehouse     |  Запланований  | TBD                | Майбутнє: Управління інвентарем    |

### Дорожня Карта

**Фаза 1 (Поточна - Активна)**:

- Модуль analytics є **комерційним** (рівні PRO/ENTERPRISE)
- Фокус на бізнес-метриках, звітах, дашбордах
- Цільова аудиторія: МСБ, підприємства, що потребують інсайтів з даних
- Статус: Реалізовано та увімкнено

**Фаза 2 (Q2 2026 - Запланована)**:

- Випуск **модуля Audit Log** (комерційний)
- Функції: Відповідність, GDPR, детальні журнали аудиту
- Цільова аудиторія: Регульовані галузі, підприємства

**Фаза 3 (Після Audit Log)**:

- Переведення **Analytics на безкоштовний рівень**
- Analytics стає доступним для всіх користувачів
- Audit Log залишається комерційним для потреб відповідності

### Обґрунтування

1. **Analytics Перший**: Працює з існуючими даними, негайна цінність
2. **Audit Log Преміум**: Відповідність є вимогою підприємств
3. **Безкоштовний Analytics**: Ширше впровадження, продаж Audit Log

---

## Моніторинг та Спостережуваність

### Метрики Ліцензій

Відстежувати використання ліцензій через логи:

```go
logger.Info("License validated",
 "module", "analytics",
 "tier", license.Tier,
 "expiry", license.ExpiryDate,
 "days_remaining", daysRemaining)
```

### Ендпоінт Здоров'я

Статус ліцензії доступний через перевірку здоров'я:

```bash
curl http://localhost:8080/api/v1/analytics/health

# Відповідь:
{
 "status": "healthy",
 "license": {
 "tier": "PRO",
 "expiry": "2026-12-31",
 "days_remaining": 365,
 "in_grace_period": false
 }
}
```

### Сповіщення

Налаштувати моніторинг для:

- Прострочених ліцензій (закінчення пільгового періоду)
- Спроб використання недійсних ліцензій
- Помилок перевірки ліцензій

---

## Тестування

### Unit Тести

Кожен модуль включає комплексні тести ліцензій:

```bash
# Запустити всі тести ліцензій
go test ./internal/modules/*/license/... -v

# Запустити тести конкретного модуля
go test ./internal/modules/analytics/license/... -v
```

Покриття тестами:

- Розбір дійсної ліцензії
- Виявлення невірного формату
- Перевірка підпису
- Обробка прострочення
- Логіка пільгового періоду
- Невідповідність модуля
- Валідація рівня

### Інтеграційні Тести

Тестувати ініціалізацію модуля з різними станами ліцензій:

```go
func TestModuleInitialization(t *testing.T) {
 tests := []struct {
 name string
 licenseKey string
 expectError bool
 }{
 {"valid license", validLicense, false},
 {"expired license", expiredLicense, true},
 {"invalid signature", tamperedLicense, true},
 {"missing license", "", true},
 }
 // ...
}
```

---

## Усунення Неполадок

### Поширені Проблеми

**1. Помилка Потреби у Ліцензії**

```
Error: license validation failed: license key is required
```

Рішення:

```bash
export ANALYTICS_LICENSE_KEY="PROMENADE-ANALYTICS-PRO-20261231-..."
```

**2. Невірний Підпис**

```
Error: invalid license signature
```

Причини:

- Невірний `LICENSE_SECRET`
- Підроблений ліцензійний ключ
- Ключ скопійований неправильно

Рішення: Перегенерувати ліцензію з правильним секретом.

**3. Прострочена Ліцензія**

```
Warning: License expired 3 days ago, grace period active
```

Рішення: Оновити ліцензію до закінчення пільгового періоду (7 днів за замовчуванням).

**4. Невідповідність Модуля**

```
Error: license module mismatch: expected ANALYTICS, got WAREHOUSE
```

Рішення: Використовувати правильну ліцензію для кожного модуля.

### Режим Налагодження

Увімкнути детальне логування ліцензій:

```yaml
# config/app.dev.yaml
log:
  level: debug

modules:
  analytics:
  settings:
  debug_license: true # Логувати всі кроки валідації
```

### Інструмент Валідації Ліцензій

Перевірити валідність ліцензії вручну:

```bash
# Валідувати ліцензію
go run ./cmd/license-generator/main.go \
 -validate \
 -key="PROMENADE-ANALYTICS-PRO-20261231-..." \
 -secret="your-secret-key"

# Вивід:
# License valid
# IModule: ANALYTICS
# Tier: PRO
# Expiry: 2026-12-31
# Days remaining: 365
```

---

## Найкращі Практики

### Для Розробників

1. **Ніколи не комітити секрети**: Використовувати `.env` файли (gitignored)
2. **Тестувати з недійсними ліцензіями**: Переконатися, що обробка помилок працює
3. **Документувати функції рівнів**: Чітка матриця функцій для кожного рівня
4. **Плавна деградація**: Не крашитися на проблемах з ліцензією в dev
5. **Структуроване логування**: Логувати події ліцензій для моніторингу

### Для Операторів

1. **Регулярно ротувати секрети**: Оновлювати `LICENSE_SECRET` щокварталу
2. **Моніторити дати прострочення**: Сповіщати за 30 днів до прострочення
3. **Безпечне зберігання секретів**: Використовувати vault рішення (AWS, HashiCorp)
4. **Відстежувати використання ліцензій**: Моніторити, які рівні активні
5. **Планувати поновлення**: Поновлювати до закінчення пільгового періоду

### Для Авторів Модулів

1. **Слідувати патернам**: Використовувати існуючий модуль analytics як шаблон
2. **Документувати функції**: Чітка документація функцій на основі рівнів
3. **Ретельно тестувати**: Комплексні тести валідації ліцензій
4. **Перевірки здоров'я**: Включати статус ліцензії в ендпоінти здоров'я
5. **Повідомлення про помилки**: Зрозумілі, дієві повідомлення про помилки

---

## Посилання

- [README Модуля Analytics](../internal/modules/analytics/README.uk.md)
- [Пакет License](../internal/modules/analytics/license/)
- [Посібник з Розробки Модулів](./MODULE_DEVELOPMENT.uk.md)
- [Архітектура Конфігурації](./MODULE_CONFIG_ARCHITECTURE.uk.md)

---

## Додаток

### Специфікація Формату Ліцензії

```
Формат: PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}

Обмеження:
- MODULE: [A-Z0-9]+ (великі літери, без пробілів)
- TIER: BASIC|PRO|ENTERPRISE
- EXPIRY: YYYYMMDD (дійсна дата)
- SIGNATURE: [A-Za-z0-9_-]+ (base64 URL-кодування)

Максимальна довжина: 256 символів
Мінімальна довжина: 50 символів
```

### Реалізація HMAC-SHA256

```go
func generateSignature(data, secret string) string {
 h := hmac.New(sha256.New, []byte(secret))
 h.Write([]byte(data))
 signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
 return strings.TrimRight(signature, "=") // Видалити padding
}

func verifySignature(data, signature, secret string) bool {
 expected := generateSignature(data, secret)
 return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Чекліст Тестування

- [ ] Розібрати дійсну ліцензію
- [ ] Відхилити невірний формат
- [ ] Перевірити підпис
- [ ] Перевірити прострочення
- [ ] Тестувати пільговий період
- [ ] Виявити невідповідність модуля
- [ ] Валідувати рівні
- [ ] Тестувати відсутню ліцензію
- [ ] Тестувати підроблену ліцензію
- [ ] Інтеграція з модулем
- [ ] Перевірка здоров'я включає статус
- [ ] Повідомлення про помилки зрозумілі

---

**Остання Оновлення**: 2024-12-19
**Версія**: 1.0.0
**Статус**: Готовий до Продакшну
