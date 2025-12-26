---
title: "Початок Роботи"
description: "Швидкий старт з Promenade"
weight: 5
---

## Швидкий Старт

Почніть працювати з Promenade за 5 хвилин.

### Передумови

- **Go 1.25+**
- **Docker & Docker Compose**
- **Make**

### 1. Клонуйте Репозиторій

\`\`\`bash
git clone https://github.com/basilex/promenade.git
cd promenade
\`\`\`

### 2. Встановіть Залежності

\`\`\`bash
make install
\`\`\`

### 3. Запустіть PostgreSQL та Redis

\`\`\`bash
make docker-up
\`\`\`

### 4. Запустіть Міграції

\`\`\`bash
make migrate
\`\`\`

### 5. Запустіть Додаток

\`\`\`bash
make dev
\`\`\`

Сервер запуститься на **http://localhost:8081**

### 6. Перевірте Здоров'я

\`\`\`bash
curl http://localhost:8081/api/v1/health
\`\`\`

### 7. Swagger Документація

Відкрийте http://localhost:8081/api/v1/docs/swagger/index.html

---

## Наступні Кроки

- [Архітектура](/promenade/docs/architecture)
- [Розробка Модулів](/promenade/docs/module-development)
- [База Даних](/promenade/docs/database-schema)
