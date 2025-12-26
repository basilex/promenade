# Стрес і навантажувальні тести

Тестування продуктивності та навантаження для Promenade API за допомогою **wrk** - сучасного інструменту бенчмаркінгу HTTP.

## Що таке стрес-тести?

Стрес-тести доводять API до межі можливостей, щоб знайти:

-  **Максимальну пропускну здатність** (запитів за секунду)
-  **Затримку під навантаженням** (p50, p95, p99)
-  **Точки відмови** (коли відбувається збій?)
-  **Використання ресурсів** (CPU, пам'ять, з'єднання)

## Передумови

### Встановлення wrk

**macOS:**

```bash
brew install wrk
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt-get install wrk
```

**З вихідного коду:**

```bash
git clone https://github.com/wg/wrk.git
cd wrk
make
sudo cp wrk /usr/local/bin/
```

Перевірка встановлення:

```bash
wrk --version
```

## Швидкий старт

```bash
# 1. Запустити API
make dev

# 2. Запустити стрес-тести (в іншому терміналі)
cd test/stress
./stress_test.sh

# Або через Makefile
make stress-test
```

## Сценарії тестування

### 1. Навантажувальний тест Health Check

**Мета:** Базова продуктивність, мінімальна логіка

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

**Очікувані результати:**

- RPS: 10,000+
- Затримка (p99): < 50ms
- Нуль помилок

### 2. Навантажувальний тест автентифікації

**Мета:** Тестування auth flow під навантаженням

```bash
wrk -t4 -c100 -d30s -s scenarios/auth.lua http://localhost:8081
```

**Очікувані результати:**

- RPS: 500-1,000
- Затримка (p99): < 200ms
- Рівень помилок: < 1%

### 3. Навантажувальний тест Posts CRUD

**Мета:** Операції, інтенсивні за базою даних

```bash
wrk -t4 -c100 -d30s -s scenarios/posts.lua http://localhost:8081
```

**Очікувані результати:**

- RPS: 300-500
- Затримка (p99): < 300ms
- Рівень помилок: < 1%

### 4. Симуляція одночасних користувачів

**Мета:** Реалістична поведінка користувачів

```bash
wrk -t8 -c200 -d60s -s scenarios/user_journey.lua http://localhost:8081
```

**Очікувані результати:**

- Одночасних користувачів: 200
- Тривалість сесії: 60s
- Реалістичний час очікування між запитами

## Рівні навантаження

### Легке навантаження (Розігрів)

```bash
wrk -t2 -c10 -d10s http://localhost:8081/api/v1/health
```

- 2 потоки, 10 з'єднань, 10 секунд
- Мета: ~1,000 RPS

### Середнє навантаження

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

- 4 потоки, 100 з'єднань, 30 секунд
- Мета: ~10,000 RPS

### Важке навантаження

```bash
wrk -t8 -c500 -d60s http://localhost:8081/api/v1/health
```

- 8 потоків, 500 з'єднань, 60 секунд
- Мета: Знайти межі

### Стрес-тест (Знайти точку відмови)

```bash
wrk -t12 -c1000 -d120s http://localhost:8081/api/v1/health
```

- 12 потоків, 1000 з'єднань, 2 хвилини
- Мета: Зламати систему, знайти вузькі місця

## Параметри wrk пояснення

```bash
wrk -t4 -c100 -d30s -s script.lua http://localhost:8081/api/v1/endpoint
    ↑    ↑    ↑      ↑
                   Lua скрипт для складних сценаріїв
             Тривалість (10s, 30s, 1m, 2h)
         З'єднання (одночасні запити)
     Потоки (ядра CPU для використання)
```

**Рекомендовано:**

- **Потоки (-t):** Кількість ядер CPU (2-8)
- **З'єднання (-c):** 10-1000 (починати з малого, збільшувати)
- **Тривалість (-d):** 10s-60s (довше для production тестів)

## Lua скрипти

### Базовий POST запит

```lua
-- scenarios/auth_login.lua
wrk.method = "POST"
wrk.body   = '{"email":"test@example.com","password":"password123"}'
wrk.headers["Content-Type"] = "application/json"
```

### Динамічні запити зі станом

```lua
-- scenarios/posts_crud.lua
local counter = 0
local token = "Bearer YOUR_TOKEN_HERE"

request = function()
    counter = counter + 1
    local path = "/api/v1/posts"
    local body = string.format('{"title":"Post %d","content":"Test"}', counter)

    return wrk.format("POST", path, {
        ["Content-Type"] = "application/json",
        ["Authorization"] = token
    }, body)
end

response = function(status, headers, body)
    if status ~= 201 then
        print("Помилка: " .. status .. " - " .. body)
    end
end
```

## Інтерпретація результатів

### Приклад виводу

```
Running 30s test @ http://localhost:8081/api/v1/health
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.12ms    2.34ms  45.23ms   89.34%
    Req/Sec     4.89k   456.23    5.67k    76.45%
  586234 requests in 30.00s, 123.45MB read
Requests/sec:  19541.13
Transfer/sec:      4.11MB
```

**Ключові метрики:**

- **Latency Avg:** Середній час відповіді (менше краще)
- **Latency Stdev:** Стабільність (менше краще)
- **Latency Max:** Найгірший випадок (стежити за стрибками)
- **Req/Sec:** Пропускна здатність на потік
- **Requests/sec:** Загальна пропускна здатність (RPS)
- **Transfer/sec:** Використана пропускна здатність мережі

### Що добре?

| Метрика             | Добре    | Попередження | Критично |
| ------------------- | -------- | ------------ | -------- |
| **Health endpoint** | >10k RPS | <5k RPS      | <1k RPS  |
| **Auth endpoints**  | >500 RPS | <200 RPS     | <100 RPS |
| **CRUD endpoints**  | >300 RPS | <100 RPS     | <50 RPS  |
| **p99 Latency**     | <100ms   | <500ms       | >1s      |
| **Рівень помилок**  | 0%       | <1%          | >5%      |

### Червоні прапорці 

- **Високий Stdev:** Непослідовна продуктивність (досліджувати)
- **Max >> Avg:** Іноді дуже повільні запити (проблема з кешуванням?)
- **Помилки:** Вичерпано пул з'єднань БД?
- **Лінійна деградація:** Не масштабується зі з'єднаннями

## Моніторинг під час тестів

### Термінал 1: Запуск API

```bash
make dev
```

### Термінал 2: Моніторинг ресурсів

```bash
# Спостереження за CPU/Пам'яттю
watch -n 1 'ps aux | grep promenade | grep -v grep'

# Або використовувати htop
htop -p $(pgrep promenade)
```

### Термінал 3: Моніторинг БД

```bash
# PostgreSQL з'єднання
watch -n 1 'psql -U promenade -d promenade_dev -c "SELECT count(*) FROM pg_stat_activity;"'

# Навантаження БД
watch -n 1 'psql -U promenade -d promenade_dev -c "SELECT * FROM pg_stat_database WHERE datname = '\''promenade_dev'\'';"'
```

### Термінал 4: Запуск wrk

```bash
wrk -t4 -c100 -d30s http://localhost:8081/api/v1/health
```

## Поради з оптимізації

### Якщо RPS низький

1. **Перевірити пул з'єднань БД:**

   ```yaml
   # config/app.dev.yaml
   database:
     max_open_conns: 100 # Збільшити
     max_idle_conns: 25 # Збільшити
   ```

2. **Увімкнути HTTP/2**

3. **Додати індекси:**

   ```sql
   CREATE INDEX idx_posts_user_id ON user_posts(user_id);
   ```

4. **Використовувати кешування:**
   - Redis для кешу сесій
   - In-memory кеш для довідкових даних

### Якщо затримка висока

1. **Профілювати код:**

   ```bash
   go tool pprof http://localhost:8081/debug/pprof/profile
   ```

2. **Перевірити повільні запити:**

   ```sql
   SELECT query, mean_exec_time, calls
   FROM pg_stat_statements
   ORDER BY mean_exec_time DESC
   LIMIT 10;
   ```

3. **Оптимізувати JSON серіалізацію**

### Якщо виникають помилки

1. **Досягнуто ліміт з'єднань БД:**

   ```
   Error: pq: sorry, too many clients already
   ```

   → Збільшити `max_connections` в PostgreSQL

2. **Context deadline exceeded:**
   → Збільшити timeout в конфігурації

3. **Out of memory:**
   → Зменшити пул з'єднань або додати RAM

## Ресурси

- [wrk GitHub](https://github.com/wg/wrk)
- [wrk Lua Scripting](https://github.com/wg/wrk/blob/master/SCRIPTING)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Go Performance Tips](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)

---

**Створено:** 25 грудня 2025
**Підтримується:** Promenade DevOps Team
