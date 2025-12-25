---
title: "Архітектура"
description: "Вичерпна документація архітектури з діаграмами"
weight: 10
---

## Огляд Архітектури

Promenade дотримується принципів **Clean Architecture** з **модульною плагін-системою**.

### Системні Шари

```mermaid
graph TB
    A[Application Entry Point] --> B[Core Infrastructure]
    B --> C[Core Domain]
    B --> D[Module System]
    C --> E[Core Use Cases]
    D --> F[Module Domains]
    F --> G[Module Use Cases]
    E --> H[Core Adapters]
    G --> I[Module Adapters]
    H --> J[Database/HTTP/Events]
    I --> J
```

### Core vs Модулі

**Core = Інфраструктура + Безпека + Референсні Дані**

- Завжди увімкнений
- Автентифікація (JWT)
- RBAC (ролі, дозволи)
- Референсні дані (країни, валюти, регіони, міста, способи оплати, часові пояси, мови)
- Event bus, База даних, Logger, Scheduler

**Модулі = Бізнес-Логіка**

- Опціональні, можна вмикати/вимикати
- Повні вертикальні зрізи (entity → usecase → adapter)
- Власні міграції, конфіги, дозволи
- Ліцензуються окремо (безкоштовні або комерційні)

---

## Шари Clean Architecture

### Потік Залежностей

```mermaid
graph LR
    A[Domain Layer] --> B[Use Case Layer]
    B --> C[Adapter Layer]
    C --> D[Infrastructure Layer]

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#a78bfa
    style D fill:#c084fc
```

**Ключове Правило:** Внутрішні шари ніколи не залежать від зовнішніх!

### Відповідальність Шарів

**1. Domain Layer** (`internal/domain/entity/`)

- Чисті бізнес-сутності
- Методи валідації
- Без залежностей від фреймворків
- Приклад: `User`, `Post`, `Comment`

**2. Use Case Layer** (`internal/usecase/`)

- Тільки бізнес-логіка
- Залежить від domain інтерфейсів
- Приклад: `RegisterUser`, `CreatePost`

**3. Adapter Layer** (`internal/adapter/`)

- HTTP обробники
- Реалізації репозиторіїв
- Інтеграція з фреймворком
- Приклад: `UserHandler`, `PostgresUserRepo`

**4. Infrastructure Layer** (`internal/infrastructure/`)

- З'єднання з базою даних
- Event bus
- Зовнішні сервіси
- Приклад: `PostgreSQL`, `Redis`, `SMTP`

---

## Архітектура Модулів

### Структура Модуля

```
internal/modules/posts/
├── module.go              # Реалізація інтерфейсу модуля
├── register.go            # Авто-реєстрація
├── config/                # Власні конфіги на середовище
├── entity/                # Доменні сутності
├── usecase/               # Бізнес-логіка
└── adapter/
    ├── http/              # Обробники, DTO, маршрути
    └── repository/        # Доступ до даних
```

### Життєвий Цикл Модуля

```mermaid
sequenceDiagram
    participant A as Application
    participant R as Registry
    participant M as Module
    participant C as Core

    A->>R: Register modules (init)
    R->>M: Initialize(core)
    M->>C: Use DB, EventBus, etc.
    M->>R: RegisterRoutes()
    M->>R: RegisterPermissions()
    A->>M: Start()
    Note over M: Background workers
    A->>M: Stop() on shutdown
```

### Незалежність Модулів

**Модулі НІКОЛИ не імпортують:**

- `internal/domain` (core domain)
- `internal/usecase` (core use cases)
- `internal/adapter` (core adapters)
- Інші модулі

**Модулі МОЖУТЬ використовувати:**

- `pkg/*` (спільні пакети)
- Core інфраструктуру через `module.Core`
- Події для міжмодульної комунікації

---

## Паттерни Комунікації

### Подієво-Орієнтована Комунікація

```mermaid
graph LR
    A[Posts Module] -->|user.registered| B[Event Bus]
    B -->|Subscribe| C[Email Module]
    B -->|Subscribe| D[Analytics Module]
    B -->|Subscribe| E[Audit Module]

    style A fill:#38bdf8
    style B fill:#fbbf24
    style C fill:#818cf8
    style D fill:#a78bfa
    style E fill:#c084fc
```

**Переваги:**

- Слабка зв'язаність між модулями
- Асинхронна обробка
- Легко додавати нових підписників
- Без прямих залежностей модулів

---

## Архітектура Бази Даних

### UUID v7 Первинні Ключі

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Переваги:**

- ⚡ У 2 рази швидші вставки (локальність B-дерева)
- 📉 Зменшена фрагментація індексів
- 🔍 Сортування за часом створення
- 🆔 Глобально унікальні

### Міграції на Основі Просторів Імен

```
migrations/
├── core/                    # Core завжди виконується першим
│   ├── 000001_init.sql
│   ├── 000002_auth.sql
│   └── 000003_rbac.sql
├── posts/                   # Модуль Posts
│   ├── 000001_posts.sql
│   └── 000002_comments.sql
└── billing/                 # Модуль Billing
    └── 000001_tables.sql
```

**Кожен простір імен має незалежну історію версій!**

---

## Система Очищення

### Дизайн на Основі Реєстрів

```mermaid
graph TB
    A[Core Scheduler] -->|Runs cron| B[Purge Use Case]
    B -->|Gets policies| C[Policy Registry]
    B -->|Gets handlers| D[Handler Registry]

    E[Posts Module] -->|Register| C
    E -->|Register| D

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#fbbf24
    style D fill:#fbbf24
```

**Ключові Моменти:**

- Core знає КОЛИ очищати (розклад, розмір пакета)
- Модулі визначають ЩО очищати (сутності, дні зберігання)
- Справжня незалежність модулів через реєстри

---

## Тестування

### Піраміда Тестів

```mermaid
graph TB
    A[Smoke Tests<br/>5-10 tests<br/>E2E critical flows]
    B[Integration Tests<br/>50+ tests<br/>Real DB]
    C[Unit Tests<br/>400+ tests<br/>Manual mocks]

    A --> B
    B --> C

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#a78bfa
```

**Покриття:** Core: 275 тестів | Posts: 33 | Profiles: 21 | Billing: 375 | Analytics: 11

**Всього: 400+ тестів виконуються за ~20 секунд**

---

## Розгортання

### Розробка

```mermaid
graph LR
    A[Developer] -->|make dev| B[Docker Compose]
    B --> C[PostgreSQL:5432]
    B --> D[Redis:6379]
    E[Go App:8081] --> C
    E --> D
```

### Продакшн

```mermaid
graph TB
    A[GitHub Actions] -->|Deploy| B[Docker Image]
    B --> C[Kubernetes Pod]
    C --> D[PostgreSQL RDS]
    C --> E[Redis ElastiCache]
    G[Load Balancer] --> C

    style A fill:#38bdf8
    style C fill:#818cf8
    style D fill:#fbbf24
    style E fill:#fbbf24
```

---

## Наступні Кроки

- [Посібник з Розробки Модулів](/promenade/docs/module-development)
- [Схема Бази Даних](/promenade/docs/database-schema)
- [Швидкий Старт](/promenade/docs/getting-started)
- [Функції](/promenade/features)
