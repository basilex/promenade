---
title: "Модулі"
description: "Повний посібник з модульної архітектури Promenade та доступних модулів"
aliases:
  - /uk/docs/modules/
---

## Система Модулів

Революційна модульна архітектура Promenade, де бізнес-логіка живе в незалежних, ліцензованих модулях.

---

## Основна Концепція

**Core = Оркестратор** (Інфраструктура, без бізнес-логіки)

- Автентифікація і Авторизація (JWT, RBAC)
- Event Bus (Memory/Redis адаптери)
- Управління БД (Connection pool, транзакції)
- Довідкові Дані (Країни, валюти, регіони, міста, методи оплати)
- Логування, Конфігурація, Планувальник

**Модулі = Виконавці** (Бізнес-домени)

- Повні вертикальні зрізи (entity → usecase → adapter)
- Незалежний життєвий цикл (вмикання/вимикання в конфігу)
- Власні міграції (на основі namespace)
- Власні дозволи (інтеграція RBAC)
- Event-driven комунікація

---

## Доступні Модулі

### Безкоштовні Модулі

<div class="docs-grid">

<div class="feature-card">

#### Posts Модуль

Управління контентом, створеним користувачами.

**Функції:**

- Створення/редагування постів з markdown
- Вкладені коментарі (threading)
- Система лайків
- Soft delete з політиками очищення
- Генерація slug

**Сутності:** `UserPost`, `Comment`, `CommentLike`

[Подивитись Код](https://github.com/basilex/promenade/tree/dev/internal/modules/posts)

</div>

<div class="feature-card">

#### Profiles Модуль

Управління профілями та контактами користувачів.

**Функції:**

- Профілі користувачів з біо, аватаром, локацією
- Множинні методи зв'язку (email, телефон, соцмережі)
- Налаштування приватності (public/private/friends)
- Верифікація контактів

**Сутності:** `UserProfile`, `UserContact`

[Подивитись Код](https://github.com/basilex/promenade/tree/dev/internal/modules/profiles)

</div>

<div class="feature-card">

#### Analytics Модуль

Збір метрик та звітність.

**Функції:**

- Запис користувацьких метрик
- Агрегація часових рядів
- Дані дашборда
- Зберігання 90 днів

**Сутності:** `Metric`, `MetricAggregate`

[Подивитись Код](https://github.com/basilex/promenade/tree/dev/internal/modules/analytics)

</div>

</div>

---

### Комерційні Модулі

<div class="docs-grid">

<div class="feature-card">

#### Billing Модуль 💰

Готова система управління підписками.

**Функції:**

- Гнучкі тарифні плани (місяць/квартал/рік)
- Пробні періоди
- Життєвий цикл підписки (активна/призупинена/скасована)
- Генерація рахунків
- Відстеження платежів
- Множинні методи оплати

**Сутності:** `Plan`, `Subscription`, `Invoice`, `Payment`

**Покриття:** 375 тестів, 100% покриття entity + usecase

[Подивитись Код](https://github.com/basilex/promenade/tree/dev/internal/modules/billing)

</div>

<div class="feature-card">

#### Audit Модуль 🔒

Незмінні логи аудиту для відповідності вимогам.

**Функції:**

- Незмінний слід аудиту
- HMAC-SHA256 підписи
- SOC 2 / GDPR відповідність
- Виявлення підробок
- Довготривале зберігання

**Випадки використання:** Фінанси, Медицина, Регульовані індустрії

</div>

<div class="feature-card">

#### Warehouse Модуль 📦

Управління інвентаризацією та продуктами (заплановано).

**Функції:**

- Каталог продуктів
- Відстеження запасів
- Коригування інвентаря
- Підтримка декількох локацій

**Статус:** Заплановано на Q1 2026

</div>

</div>

---

## Розробка Модулів

### Створіть Власний Модуль

Слідуйте нашому [Посібнику з Розробки Модулів](/uk/docs/module-development) для створення власних модулів.

**Швидка Структура:**

```
internal/modules/mymodule/
├── module.go              # Реалізація інтерфейсу IModule
├── register.go            # Авто-реєстрація через init()
├── config/                # Конфіги для різних середовищ
├── entity/                # Доменні сутності
├── usecase/               # Бізнес-логіка
└── adapter/
    ├── http/              # HTTP хендлери та DTO
    └── repository/        # PostgreSQL реалізації
```

### Можливості Модулів

✅ **Авто-реєстрація** - функція `init()` реєструє модуль  
✅ **Власна конфігурація** - YAML конфіги для кожного середовища  
✅ **Власні міграції** - На основі namespace, незалежна історія  
✅ **Власні дозволи** - Інтеграція RBAC  
✅ **Event комунікація** - Publish/subscribe до доменних подій  
✅ **Фонові воркери** - Cron jobs, процесори черг  
✅ **Health checks** - Graceful startup/shutdown

---

## Незалежність Модулів

**Модулі НІКОЛИ не імпортують:**

- `internal/domain` (core domain)
- `internal/usecase` (core use cases)
- `internal/adapter` (core adapters)
- Інші модулі

**Модулі МОЖУТЬ використовувати:**

- `pkg/*` (спільні пакети)
- Core інфраструктуру через `module.Core`
- Події для комунікації модуль-модуль

Це забезпечує справжню архітектурну автономію.

---

## Увімкнення Модулів

### Конфігурація

Редагуйте `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics
    - billing # Потрібна ліцензія
    - audit # Потрібна ліцензія
    # - warehouse  # Ще не доступний
```

### Імпорт в main.go

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/analytics"
    _ "github.com/basilex/promenade/internal/modules/billing"
)
```

### Запуск Міграцій

```bash
make migrate                     # Всі увімкнені модулі
make migrate-module MODULE=posts # Конкретний модуль
```

---

## Комерційне Ліцензування

Комерційні модулі (billing, audit, warehouse) потребують ліцензійних ключів.

**Згенерувати Ліцензію:**

```bash
./scripts/generate-license.sh billing PRO 365
```

**Встановити Змінну Середовища:**

```bash
export BILLING_LICENSE_KEY="PROMENADE-BILLING-PRO-20261225-xxxxx"
```

**Або налаштувати в YAML:**

```yaml
# config/modules.yaml
modules:
  config:
    billing:
      license_key: "PROMENADE-BILLING-PRO-20261225-xxxxx"
```

Контакт: alexander.vasilenko@gmail.com для корпоративного ліцензування.

---

## Дізнатися Більше

- [Посібник з Розробки Модулів](/uk/docs/module-development) - Повний туторіал
- [Огляд Архітектури](/uk/docs/architecture) - Дизайн системи
- [Схема Бази Даних](/uk/docs/database-schema) - Таблиці модулів
- [Приклади Модулів](https://github.com/basilex/promenade/tree/dev/internal/modules) - Вихідний код

---

## Підтримка

🐛 [Повідомити про Проблему](https://github.com/basilex/promenade/issues)  
💬 [Обговорення](https://github.com/basilex/promenade/discussions)  
📧 alexander.vasilenko@gmail.com
