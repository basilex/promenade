---
title: "Документація"
description: "Повна технічна документація для Promenade"
---

## Технічна Документація

Вичерпні посібники, діаграми та довідники для архітектури та розробки Promenade.

---

## Архітектура

<div class="docs-grid">

### [Огляд Архітектури](/docs/architecture)

Повна системна архітектура з діаграмами, шарами та паттернами комунікації.

- Шари Clean Architecture
- Дизайн модульної системи
- Подієво-орієнтована комунікація
- Архітектура розгортання

### [Схема Бази Даних](/docs/database-schema)

Детальна схема бази даних з ER діаграмами та зв'язками.

- Основні таблиці (користувачі, сесії, RBAC)
- Референсні дані (країни, валюти, регіони, міста)
- Таблиці модулів (пости, профілі, аналітика, білінг)
- Індекси та продуктивність

### [Розробка Модулів](/docs/module-development)

Покрокова інструкція зі створення нових модулів.

- Структура та життєвий цикл модуля
- Паттерни комунікації
- Найкращі практики
- Приклади та усунення проблем

</div>

---

## Швидкі Посилання

### Для Розробників

- [Посібник Початку Роботи](/docs/getting-started)
- [Розробка Модулів](/docs/module-development)
- [Посібник з Тестування](/docs/TESTING_GUIDE)
- [API Документація](/api/v1/docs/swagger/index.html)

### Для Архітекторів

- [Огляд Архітектури](/docs/architecture)
- [Схема Бази Даних](/docs/database-schema)
- [Незалежність Модулів](/docs/MODULE_INDEPENDENCE)
- [Принципи Clean Architecture](/features/clean-architecture)

### API та Інтеграція

- [Автентифікація та RBAC](/features/authentication)
- [Система Event Bus](/features/event-bus)
- [REST API v1](/api/v1/docs/swagger/index.html)
- [REST API v2](/api/v2/docs/swagger/index.html)

---

## Детальний Огляд Функцій

<div class="features-grid">

<div class="feature-card">

### [Чиста Архітектура](/features/clean-architecture)

Строга шарова архітектура з чітким розділенням відповідальності.

</div>

<div class="feature-card">

### [Модульна Система](/features/module-system)

Незалежні бізнес-модулі з плагін-архітектурою.

</div>

<div class="feature-card">

### [Автентифікація та RBAC](/features/authentication)

JWT автентифікація з рольовим контролем доступу.

</div>

<div class="feature-card">

### [База Даних та Міграції](/features/database)

PostgreSQL з UUID v7 та міграції на основі просторів імен.

</div>

<div class="feature-card">

### [Тестова Інфраструктура](/features/testing)

400+ тестів з ручним підходом до моків.

</div>

<div class="feature-card">

### [Подієво-Орієнтована Архітектура](/features/event-bus)

Подвійні адаптери event bus для асинхронної комунікації.

</div>

</div>

---

## Зовнішні Ресурси

- [GitHub Репозиторій](https://github.com/basilex/promenade)
- [Трекер Проблем](https://github.com/basilex/promenade/issues)
- [Посібник з Внесення Змін](https://github.com/basilex/promenade/blob/dev/CONTRIBUTING.md)
- [MIT Ліцензія](https://github.com/basilex/promenade/blob/dev/LICENSE)

---

<style>
.docs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin: 2rem 0;
}

.docs-grid > div {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.docs-grid > div:hover {
  border-color: var(--color-primary);
  transform: translateY(-2px);
}

.docs-grid h3 {
  margin-top: 0;
  margin-bottom: 0.5rem;
}

.docs-grid h3 a {
  color: var(--color-primary);
  text-decoration: none;
}

.docs-grid p {
  margin-bottom: 0.5rem;
  color: var(--color-text-secondary);
}

.docs-grid ul {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}
</style>
