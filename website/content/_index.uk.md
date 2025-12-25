---
title: "Production-Ready REST API Framework"
description: "Побудовано з Clean Architecture, модульною плагін-системою та міграціями бази даних на основі namespace. Написано на Go."
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
# Клонувати репозиторій
git clone https://github.com/basilex/promenade.git
cd promenade

# Запустити PostgreSQL + Redis
make docker-up

# Запустити міграції
make migrate

# Запустити додаток
make dev
```

Сервер запуститься на **http://localhost:8081**

## Ключові особливості

### Модульна архітектура

Promenade використовує **систему плагінів** де кожен модуль - це повна вертикальна секція:

- Власні domain сутності
- Власні репозиторії
- Власні use cases
- Власні HTTP handlers
- Власні міграції бази даних

### Доступні модулі

**Безкоштовні:**

- **Posts** - користувацький контент (пости, коментарі, лайки)
- **Profiles** - профілі користувачів та контакти
- **Analytics** - метрики, звіти, дашборди

**Комерційні:**

- **Billing** - управління підписками, рахунки, платежі
- **Audit** - незмінні аудит-логи з криптографічними підписами
- **Warehouse** - управління інвентарем (незабаром)

## Навчальні матеріали

- [Архітектурний огляд](/docs/ARCHITECTURE_OVERVIEW) - візуальні діаграми
- [Розробка модулів](/docs/MODULE_DEVELOPMENT) - створення нових модулів
- [Тестування](/docs/TESTING_GUIDE) - стратегії тестування
- [Посібник з UUID v7](/docs/UUID_V7_GUIDE) - впорядковані за часом ідентифікатори

---

**Побудовано з Clean Architecture та Go**
