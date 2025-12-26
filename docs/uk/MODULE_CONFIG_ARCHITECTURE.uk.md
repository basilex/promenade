# Архітектура Конфігурації Модулів

 [English](../MODULE_CONFIG_ARCHITECTURE.md) |  **Українська** | [ Deutsch](../de/MODULE_CONFIG_ARCHITECTURE.de.md) | [ Português](../pt/MODULE_CONFIG_ARCHITECTURE.pt.md) | [ Español](../es/MODULE_CONFIG_ARCHITECTURE.es.md)

## Огляд

Модулі в Promenade є **повністю автономними вертикальними зрізами**, які керують власною конфігурацією. Кожен модуль завантажує свою конфігурацію з власного дерева каталогів, забезпечуючи повну незалежність від основної системи.

## Принципи Архітектури

### 1. Автономія Модулів

- Кожен модуль відповідає за завантаження власної конфігурації
- Модулі зберігають конфігурації у власному каталозі: `internal/modules/{назва_модуля}/config/`
- Ядро не завантажує і не керує конфігураціями модулів
- Ядро надає лише інфраструктуру (DB, EventBus, JWT) та контекст оточення

### 2. Підтримка Оточень

- Кожен модуль має три файли конфігурації для різних оточень:
  - `config.dev.yaml` - Налаштування для розробки
  - `config.test.yaml` - Налаштування для тестування
  - `config.prod.yaml` - Налаштування для продакшену
- SDK модуля автоматично завантажує правильну конфігурацію на основі змінної `ENVIRONMENT`

### 3. Потік Завантаження Конфігурації

```
Запуск Додатка
    ↓
Ядро завантажує app.{env}.yaml
    ↓
Ядро ініціалізує інфраструктуру (DB, EventBus тощо)
    ↓
Реєстр модулів виявляє модулі
    ↓
Для кожного увімкненого модуля:
    Викликається IModule.Initialize(ctx, core)
        ↓
    Модуль завантажує internal/modules/{назва}/config/config.{env}.yaml
        ↓
    Модуль перевіряє налаштування (ліцензія, функції тощо)
        ↓
    Модуль ініціалізує репозиторії, use cases, обробники
        ↓
    IModule.RegisterRoutes() реєструє HTTP ендпоінти
```

## Структура Каталогів

```
config/
 app.dev.yaml          # Конфігурація ядра - розробка
 app.test.yaml         # Конфігурація ядра - тести
 app.prod.yaml         # Конфігурація ядра - продакшен
 modules.yaml          # Реєстр модулів (увімкнено/вимкнено)

internal/modules/
 posts/
    config/
       config.dev.yaml   # Конфігурація модуля posts для розробки
       config.test.yaml  # Конфігурація модуля posts для тестів
       config.prod.yaml  # Конфігурація модуля posts для продакшену
    entity/
    repository/
    usecase/
    handler/
    module.go            # Завантажує власну конфігурацію в Initialize()

 warehouse/
    config/
       config.dev.yaml   # Конфігурація модуля warehouse для розробки
       config.test.yaml  # Конфігурація модуля warehouse для тестів
       config.prod.yaml  # Конфігурація модуля warehouse для продакшену
    module.go            # Завантажує власну конфігурацію в Initialize()

 profiles/
     config/
        config.dev.yaml   # Конфігурація модуля profiles для розробки
        config.test.yaml  # Конфігурація модуля profiles для тестів
        config.prod.yaml  # Конфігурація модуля profiles для продакшену
     ...
```

## Області Конфігурації

### Конфігурація Ядра (`config/app.{env}.yaml`)

**Керується**: `internal/infrastructure/config/yaml_config.go`

Містить:

- Метадані додатка (назва, версія, оточення)
- Налаштування сервера (хост, порт, таймаути)
- Підключення до бази даних (хост, порт, облікові дані)
- Налаштування JWT (секрет, термін дії)
- Конфігурація логування
- Налаштування CORS
- Адаптер шини подій (memory/redis)
- Обмеження швидкості запитів
- Сервіс електронної пошти

**Ніколи не містить**: Налаштування бізнес-логіки модулів

### Конфігурація Модуля (`internal/modules/{назва}/config/config.{env}.yaml`)

**Керується**: Кожним модулем за допомогою `pkg/module/config`

Містить:

- Метадані модуля (назва, версія, прапорець увімкнення)
- Специфічні налаштування модуля
- Політики очищення (дні зберігання, розміри пакетів)
- Визначення дозволів
- Прапорці функцій
- Ліцензійні ключі (для комерційних модулів)

**Ніколи не містить**: Налаштування основної інфраструктури

## Приклад Реалізації

### Завантаження Конфігурації Модуля Posts

```go
// internal/modules/posts/module.go
package posts

import (
    "context"
    "github.com/basilex/promenade/pkg/module"
    moduleconfig "github.com/basilex/promenade/pkg/module/config"
)

type PostsModule struct {
    *module.BaseModule
    config *moduleconfig.Config
    // ... інші поля
}

func (m *PostsModule) Initialize(ctx context.Context, core *module.Core) error {
    // Викликаємо базову реалізацію
    if err := m.BaseModule.Initialize(ctx, core); err != nil {
        return err
    }

    // Завантажуємо власну конфігурацію модуля з власного каталогу
    config, err := moduleconfig.Load("internal/modules/posts/config", core.Config.AppName)
    if err != nil {
        return fmt.Errorf("failed to load posts config: %w", err)
    }
    m.config = config

    // Використовуємо налаштування конфігурації
    retentionDays := m.config.GetRetentionDays("posts", 30)
    batchSize := m.config.GetBatchSize("posts", 100)

    // ... ініціалізуємо репозиторії, use cases, обробники

    return nil
}
```

### Структура Файлу Конфігурації Модуля

```yaml
# internal/modules/posts/config/config.dev.yaml
module:
  enabled: true
  name: "posts"
  version: "1.0.0"

settings:
  max_content_length: 10000
  allow_markdown: true

purge:
  enabled: true
  schedule: "0 2 * * *"
  settings:
    posts:
      retention_days: 30
      batch_size: 100
    comments:
      retention_days: 60
      batch_size: 50

permissions:
  - resource: "posts"
    actions: ["create", "read", "update", "delete"]
  - resource: "comments"
    actions: ["create", "read", "update", "delete"]

features:
  enable_likes: true
  enable_sharing: true
```

## SDK Конфігурації Модуля

### Завантаження Конфігурації

```go
import moduleconfig "github.com/basilex/promenade/pkg/module/config"

// Завантажити конфігурацію для поточного оточення
config, err := moduleconfig.Load("internal/modules/{назва}/config", environment)
```

### Доступ до Налаштувань

```go
// Отримати дні зберігання для очищення зі значенням за замовчуванням
retentionDays := config.GetRetentionDays("назва_сутності", 30)

// Отримати розмір пакета для очищення зі значенням за замовчуванням
batchSize := config.GetBatchSize("назва_сутності", 100)

// Перевірити, чи функція увімкнена
enabled := config.GetFeature("enable_likes", true)

// Отримати вкладене налаштування
value := config.GetNestedSetting("settings", "max_content_length")
```

## Міграція з Централізованих Конфігурацій

### Стара Архітектура (Застаріла)

```
config/modules/
 posts.dev.yaml          Централізовано
 posts.test.yaml         Централізовано
 posts.prod.yaml         Централізовано
 ...

Ядро завантажує всі конфігурації модулів   Тісний зв'язок
Ядро передає конфігурації модулям   Залежність
```

### Нова Архітектура (Поточна)

```
internal/modules/posts/config/
 config.dev.yaml         Належить модулю
 config.test.yaml        Належить модулю
 config.prod.yaml        Належить модулю

Модуль завантажує власну конфігурацію   Автономний
Модуль керує власними налаштуваннями   Незалежний
```

## Переваги

### 1. Справжня Незалежність Модулів

- Модулі можна розробляти, тестувати та розгортати незалежно
- Не потрібні зміни в ядрі при додаванні/зміні конфігурацій модулів
- Модулі є справді автономними плагінами

### 2. Краще Інкапсулювання

- Конфігурація знаходиться разом з кодом, який вона налаштовує
- Чітке володіння та відповідальність
- Легше зрозуміти можливості модуля

### 3. Спрощене Ядро

- Ядро керує лише інфраструктурою
- Немає бізнес-логіки в конфігураціях ядра
- Чіткіше розділення відповідальностей

### 4. Простіше Тестування

- Кожен модуль може мати різні тестові конфігурації
- Немає забруднення глобальної конфігурації
- Тести модулів ізольовані

### 5. Гнучке Розгортання

- Увімкнення/вимкнення модулів без змін в ядрі
- Різні оточення можуть мати різні конфігурації модулів
- Комерційні модулі можуть мати перевірку ліцензій

## Комерційні Модулі

Комерційні модулі (наприклад, warehouse) вимагають ліцензійних ключів:

```yaml
# internal/modules/warehouse/config/config.prod.yaml
module:
  enabled: true
  name: "warehouse"
  version: "1.2.0"
  license_key: "ваш-комерційний-ліцензійний-ключ-тут"

settings:
  max_items: 100000
  enable_barcode_scanner: true
```

Модуль перевіряє ліцензію під час виконання `Initialize()`:

```go
func (m *WarehouseModule) verifyLicense() error {
    licenseKey := m.config.GetNestedSetting("module", "license_key").(string)
    if licenseKey == "" {
        return fmt.Errorf("warehouse module requires license key")
    }
    // Перевірка ліцензії...
    return nil
}
```

## Найкращі Практики

### РОБІТЬ

- Зберігайте конфігурації модулів в `internal/modules/{назва}/config/`
- Завантажуйте конфігурацію в методі `Initialize()` модуля
- Використовуйте файли конфігурації для різних оточень
- Перевіряйте критичні налаштування під час ініціалізації
- Використовуйте допоміжні методи `pkg/module/config`
- Надавайте розумні значення за замовчуванням для опціональних налаштувань

### НЕ РОБІТЬ

- Розміщувати конфігурації модулів в `config/modules/` (застаріло)
- Завантажувати конфігурації модулів в ядрі
- Розміщувати налаштування модулів в конфігурації ядра
- Жорстко кодувати значення для конкретних оточень
- Пропускати перевірку конфігурації
- Отримувати доступ до конфігурацій інших модулів

## Налагодження

### Перевірка, яка конфігурація завантажена:

```go
logger.FromContext(ctx).Info("IModule config loaded",
    "module", m.GetMetadata().Name,
    "config_path", "internal/modules/posts/config",
    "environment", os.Getenv("ENVIRONMENT"),
)
```

### Перевірка оточення:

```bash
echo $ENVIRONMENT  # Має бути: development, test або production
```

### Помилки завантаження конфігурації:

- Перевірте, чи існує файл: `internal/modules/{назва}/config/config.{env}.yaml`
- Перевірте синтаксис YAML
- Перевірте права доступу до файлу
- Переконайтеся, що оточення встановлено правильно

## Пов'язана Документація

- [Посібник з Розробки Модулів](MODULE_DEVELOPMENT.uk.md)
- [Незалежність Модулів](MODULE_INDEPENDENCE.uk.md)
- [Інфраструктура Тестування](TESTING_INFRASTRUCTURE.uk.md)
- [Міграція Конфігурації](CONFIG_MIGRATION.uk.md)
