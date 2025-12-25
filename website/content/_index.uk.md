---
title: "Production-Ready REST API Framework"
description: "Enterprise-рівня backend фреймворк, побудований на принципах Clean Architecture. Революційна модульна плагін-система, де бізнес-логіка живе в незалежних, ліцензованих модулях. Ніяких спільних доменних сутностей, жодного тісного зв'язку - справжня архітектурна автономія."
date: 2025-12-25
features:
  - icon: "🏗️"
    title: "Clean Architecture"
    description: "Суворий багатошарова архітектура з шарами Domain, Use Case, Adapter та Infrastructure. Правило залежностей дотримане."

  - icon: "🧩"
    title: "Модульна система"
    description: "Незалежні бізнес-модулі з власними сутностями, репозиторіями та use cases. Динамічне вмикання/вимикання модулів."

  - icon: "🗄️"
    title: "PostgreSQL + UUID v7"
    description: "Впорядковані за часом UUID для 2x швидших вставок. Чистий SQL з sqlx - без ORM. Паттерн BaseRepository."

  - icon: "🔐"
    title: "Auth & RBAC"
    description: "JWT аутентифікація, контроль доступу на основі ролей з шаблонними дозволами. Включено управління сесіями."

  - icon: "📡"
    title: "Event-Driven"
    description: "Подвійні адаптери event bus (Memory/Redis). Асинхронна комунікація між модулями. Готово до production."

  - icon: "🔄"
    title: "Namespace міграції"
    description: "Кожен модуль має незалежну історію міграцій. Справжня автономія модулів без конфліктів схем."

  - icon: "✅"
    title: "400+ тестів"
    description: "Повний набір тестів з unit, integration та smoke тестами. Паттерн ручних моків для тестування."

  - icon: "📚"
    title: "20+ документів"
    description: "Розширена документація на EN/UK/DE. Архітектурні посібники, туторіали та API довідки."

  - icon: "🚀"
    title: "Production-Ready"
    description: "Використовується в реальних додатках. Доступні комерційні модулі (billing, audit, warehouse)."
---

## Швидкий старт

```bash
git clone https://github.com/basilex/promenade.git
cd promenade
make docker-up
make migrate
make dev
```

Сервер запускається на **http://localhost:8081**

---

## Філософія Promenade

**Promenade - це не просто ще один Go фреймворк** - це повне переосмислення того, як повинні бути структуровані backend-додатки.

🎯 **Core як Оркестратор, Модулі як Робітники** - Core надає інфраструктуру без бізнес-логіки

🔌 **Справжня Незалежність** - Модулі ніколи не імпортують з `internal/domain`

📊 **Namespace Міграції** - Кожен модуль має власну історію без конфліктів

⚡ **Продуктивність** - UUID v7 забезпечує в 2 рази швидші вставки

🔐 **Безпека** - JWT, RBAC, audit логи з HMAC підписами

---

## Чому Обрати Promenade?

### Для Стартапів

✅ Швидкий вихід на ринок ✅ Безкоштовні модулі для MVP ✅ Єдине розгортання

### Для Enterprise

✅ Підтримуваність ✅ Аудитованість (SOC 2, GDPR) ✅ Масштабованість

### Для Команд

✅ Без конфліктів міграцій ✅ Чіткі межі модулів ✅ Багатомовна документація

### Для Розробників

✅ Кращі практики ✅ Фокус на бізнесі ✅ Готові паттерни

---

## Документація

📖 [Огляд Архітектури](docs/architecture/) | [Схема БД](docs/database-schema/) | [Розробка Модулів](docs/module-development/)

🌐 EN/UK/DE | 📝 [Swagger API](/api/v1/docs/swagger/index.html)

---

**Створено з ❤️ використовуючи Clean Architecture та Go**

[Почніть](docs/getting-started/) | [Функції](features/) | [GitHub](https://github.com/basilex/promenade)
