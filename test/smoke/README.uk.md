# Smoke Тести

Smoke тести перевіряють, що критичний функціонал API працює коректно в робочому оточенні. Це **чорноскриньковий інтеграційні тести**, які виконують реальні HTTP запити до API.

## Що таке Smoke Тести?

Smoke тести - це швидкі, основні тести, які перевіряють базову функціональність додатку:

-  Чи запускається API?
-  Чи доступні endpoints?
-  Чи працюють критичні workflow end-to-end?
-  Чи підключена база даних?

**Що НЕ покривається smoke тестами:**

-  Граничні випадки
-  Тестування продуктивності/навантаження
-  Логіка unit тестів
-  Деталі обробки помилок

## Швидкий старт

### Передумови

- **curl** - HTTP клієнт
- **jq** - JSON процесор
- Запущений екземпляр API (порт 8081 за замовчуванням)

```bash
# Перевірити чи встановлені інструменти
which curl && which jq
```

### Запуск тестів

```bash
# 1. Запустити API (якщо ще не запущено)
make dev

# 2. Запустити smoke тести (в іншому терміналі)
make test-smoke

# Або запустити напряму
cd test/smoke
./smoke_test.sh
```

### Очікуваний вивід

```

   Promenade API - Smoke Tests


  API:         http://localhost:8081
  Environment: dev
  Date:        2025-12-25 14:30:00

ℹ Перевірка необхідних інструментів...
 Всі необхідні інструменти знайдено
ℹ Очікування готовності API...
 API готове!


  Запуск Smoke Тестів


  Перевірка здоров'я

ℹ Тестування GET /health...
 Перевірка здоров'я пройдена
 Тест пройдено: 01_health

...


  Підсумок тестування

  Всього тестів:  5
  Пройдено:       5
  Провалено:      0

 Всі smoke тести пройшли успішно!
```

## Структура тестів

### Тести (у порядку виконання)

1. **01_health.sh** - Перевірка health endpoint
2. **02_auth.sh** - Процес автентифікації (логін → отримання профілю)
3. **03_posts.sh** - CRUD операції з постами
4. **04_profiles.sh** - CRUD операції з профілями
5. **05_analytics.sh** - Метрики аналітики

### Конфігурація

**config.sh** - Конфігурація тестів:

- URL API (за замовчуванням: `http://localhost:8081`)
- Тестові облікові дані
- HTTP таймаути

**helpers.sh** - Допоміжні функції:

- HTTP обгортки (GET, POST, PUT, DELETE)
- JSON перевірки
- Кольоровий вивід
- Утиліти очищення

## Налаштування

### Змінні оточення

```bash
# URL API (за замовчуванням: http://localhost:8081)
export PROMENADE_API_URL="http://localhost:8081"

# Назва оточення (опціонально)
export PROMENADE_ENV="dev"
```

### Користувацький URL API

```bash
# Тестування проти staging
PROMENADE_API_URL="https://staging.example.com" ./smoke_test.sh

# Тестування проти production
PROMENADE_API_URL="https://api.example.com" ./smoke_test.sh
```

## Запуск окремих тестів

Кожен тест можна запустити незалежно:

```bash
cd test/smoke

# Запуск одного тесту
./tests/01_health.sh

# Запуск конкретних тестів
./tests/02_auth.sh
./tests/03_posts.sh
```

**Примітка:** Тести 03-05 потребують автентифікації, тому тест 02 запуститься автоматично за потреби.

## Написання нових тестів

### Шаблон

```bash
#!/bin/bash

# Smoke Test: Моя функція

source "$(dirname "$0")/../helpers.sh"

test_my_feature() {
    print_section "Моя функція"

    # Логіка тесту тут
    print_info "Тестування GET /my-endpoint..."
    local response=$(http_get "$API_BASE/my-endpoint" 200)

    if [ $? -ne 0 ]; then
        print_error "Тест провалено"
        return 1
    fi

    assert_json_field "$response" ".success" "true" || return 1

    print_success "Тест пройдено"
    return 0
}

# Запуск тесту
test_my_feature
exit $?
```

### Додавання до runner

Відредагувати `smoke_test.sh`:

```bash
run_test "tests/06_my_feature.sh"
```

## Інтеграція з CI/CD

### Приклад GitHub Actions

```yaml
name: Smoke Tests

on: [push, pull_request]

jobs:
  smoke-tests:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.25"

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq

      - name: Run migrations
        run: make migrate

      - name: Build API
        run: make build

      - name: Start API
        run: ./bin/promenade &
        env:
          ENVIRONMENT: test

      - name: Wait for API
        run: sleep 5

      - name: Run smoke tests
        run: make test-smoke

      - name: Stop API
        run: pkill promenade
```

## Усунення несправностей

### API не готовий

```bash
# Перевірити чи працює API
curl http://localhost:8081/api/v1/health

# Перевірити логи
tail -f /path/to/logs/promenade.log
```

### Провал тестів

```bash
# Запустити з детальним виводом
bash -x ./test/smoke/tests/01_health.sh

# Перевірити окремі запити
curl -v http://localhost:8081/api/v1/health
```

### Відмовлено в доступі

```bash
# Зробити скрипти виконуваними
chmod +x test/smoke/*.sh
chmod +x test/smoke/tests/*.sh
```

### Відсутні залежності

```bash
# macOS
brew install curl jq

# Ubuntu/Debian
apt-get install curl jq

# CentOS/RHEL
yum install curl jq
```

## Кращі практики

1. **Тримайте тести швидкими** - Smoke тести повинні виконуватися < 2 хвилин
2. **Тестуйте лише критичні шляхи** - Не кожен endpoint потребує smoke тесту
3. **Використовуйте унікальні тестові дані** - Часові мітки в тестових email/назвах запобігають конфліктам
4. **Очищайте після тестів** - Видаляйте тестові дані щоб уникнути забруднення
5. **Робіть тести ідемпотентними** - Повинні проходити навіть при множинному запуску
6. **Використовуйте перевірки** - Перевіряйте структуру відповіді, не лише status коди
7. **Швидкий провал** - Зупиняйтеся на першій помилці щоб заощадити час

## Що далі?

Після проходження smoke тестів, розгляньте:

- **Навантажувальне тестування** - Дивіться `test/stress/` для wrk-базованих стрес тестів
- **End-to-End тести** - Cypress/Playwright для UI тестування
- **Тестування безпеки** - OWASP ZAP, Burp Suite
- **Chaos тестування** - Симуляція збоїв

## Підтримка

- **Проблеми** - Повідомляйте про баги в smoke тестах
- **Питання** - Запитуйте в командному чаті
- **Покращення** - Надсилайте PR з новими тестами

---

**Створено:** 25 грудня 2025
**Підтримується:** Promenade DevOps Team
