---
title: "Посібники"
description: "Повна колекція посібників та туторіалів для Promenade"
aliases:
  - /uk/docs/guides/
---

## Посібники та Туторіали

Комплексні посібники для вивчення та роботи з Promenade.

---

## Початок Роботи

<div class="docs-grid">

### [Швидкий Старт](/uk/docs/getting-started)

Запустіться за 5 хвилин. Встановіть залежності, запустіть сервіси та перевірте інсталяцію.

### [Огляд Архітектури](/uk/docs/architecture)

Зрозумійте основні принципи: Clean Architecture, Core vs Modules, та дизайн системи.

### [Налаштування БД](/uk/docs/database-schema)

Дізнайтеся про конфігурацію PostgreSQL, UUID v7, міграції та дизайн схеми.

</div>

---

## Посібники з Розробки

<div class="docs-grid">

### [Розробка Модулів](/uk/docs/module-development)

Повний посібник зі створення власних бізнес-модулів з прикладами та практиками.

### [Стратегія Тестування](/uk/docs/getting-started#testing)

Як писати unit, integration та smoke тести. Вивчіть паттерн ручних моків.

### [API Інтеграція](/api/v1/docs/swagger/index.html)

Досліджуйте REST API ендпоінти, автентифікацію та паттерни інтеграції через Swagger UI.

</div>

---

## Архітектурні Посібники

<div class="docs-grid">

### [Clean Architecture](/uk/features/clean-architecture)

Глибоке занурення в шари Domain, Use Case, Adapter та Infrastructure.

### [Система Модулів](/uk/features/module-system)

Як працює система модулів: реєстрація, життєвий цикл та паттерни комунікації.

### [Event-Driven Architecture](/uk/features/event-bus)

Використання event bus для асинхронної комунікації між модулями (Memory/Redis адаптери).

</div>

---

## Посібники з Бази Даних

<div class="docs-grid">

### [Схема Бази Даних](/uk/docs/database-schema)

Повна документація схеми з ER діаграмами для всіх таблиць.

### [UUID v7 Посібник](/uk/features/database#uuid-v7)

Чому UUID v7 в 2 рази швидший за UUID v4 та як його використовувати.

### [Паттерн Soft Delete](/uk/features/database#soft-delete)

Реалізація м'якого видалення для контенту користувачів з автоматичними політиками очищення.

</div>

---

## Просунуті Теми

<div class="docs-grid">

### [RBAC & Дозволи](/uk/features/authentication)

Контроль доступу на основі ролей з wildcard дозволами та 4 системними ролями.

### [Довідкові Дані](/uk/docs/database-schema#reference-data)

145 країн, 124 валюти, 30 регіонів, 17 міст, 40+ методів оплати.

### [Namespace Міграції](/uk/docs/database-schema#migrations)

Незалежна історія міграцій для кожного модуля без конфліктів.

</div>

---

## Зовнішні Ресурси

- [GitHub Репозиторій](https://github.com/basilex/promenade) - Вихідний код та issues
- [Приклади Модулів](https://github.com/basilex/promenade/tree/dev/internal/modules) - Posts, Profiles, Analytics, Billing
- [API Документація](/api/v1/docs/swagger/index.html) - Інтерактивний Swagger UI

---

## Потрібна Допомога?

- Перегляньте [Індекс Документації](/uk/docs/) для всіх доступних посібників
- Відвідайте [GitHub Discussions](https://github.com/basilex/promenade/discussions) для запитань
- Повідомляйте про проблеми на [GitHub Issues](https://github.com/basilex/promenade/issues)
