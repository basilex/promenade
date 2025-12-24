# Керівництво з Реалізації UUID v7

## Огляд

Цей проєкт було мігровано з UUID v4 на UUID v7 для генерації первинних ключів. UUID v7 забезпечує значні переваги продуктивності, зберігаючи глобальну унікальність та можливості розподіленої генерації UUID v4.

## Чому UUID v7?

### Переваги над UUID v4

1. **Впорядковані за часом**: UUID природно відсортовані за часом створення
2. **Краща продуктивність B-tree**: Зменшує розділення сторінок в індексах PostgreSQL (INSERT на 20-50% швидше в бенчмарках)
3. **Зменшена фрагментація**: Послідовні-ish ID мінімізують фрагментацію індексів
4. **Дружній до кешу**: Краща локальність посилань для нещодавно створених записів
5. **Витягуваний timestamp**: Можна отримати час створення з самого UUID
6. **Все ще глобально унікальний**: Зберігає гарантії унікальності UUID v4

### Порівняння з іншими стратегіями

| Стратегія        | Плюси                                            | Мінуси                                         |
| ---------------- | ------------------------------------------------ | ---------------------------------------------- |
| **UUID v7**      | [+] Впорядкований за часом, глобально унікальний | Трохи більший ніж BIGINT (16 байт)             |
| UUID v4          | Глобально унікальний, розподілений               | [X] Випадковий = погана локальність індексу    |
| SERIAL/BIGSERIAL | Малий, послідовний, швидкий                      | [X] Єдина точка відмови, проблеми реплікації   |
| ULID             | Схожий на v7, кодування base32                   | Менш стандартний, обмежена підтримка бібліотек |
| Snowflake ID     | Швидкий, впорядкований за часом                  | Потребує сервісу координації                   |

## Реалізація

### Налаштування PostgreSQL

Проєкт включає дві міграції:

1. **000005_add_uuidv7_support.sql**: Додає функцію `uuid_generate_v7()` до PostgreSQL
2. **000006_switch_tables_to_uuidv7.sql**: Оновлює значення таблиць за замовчуванням на v7

```sql
-- Згенерувати UUID v7 в PostgreSQL
SELECT uuid_generate_v7();
-- Приклад: 018d2f07-0f3c-7000-8000-123456789abc
```

### Код Go

Використовуйте пакет `pkg/uuidv7`:

```go
import "github.com/basilex/promenade/pkg/uuidv7"

// Згенерувати новий UUID v7
id := uuidv7.New()

// Згенерувати з конкретним timestamp (корисно для тестування)
id := uuidv7.NewWithTime(time.Now())

// Витягнути timestamp створення з UUID
timestamp := uuidv7.ExtractTime(id)

// Перевірити чи UUID є версії 7
if uuidv7.IsV7(id) {
    // ...
}
```

### Шлях Міграції

Існуючі записи з UUID v4 продовжать працювати. Система підтримує змішані UUID:

| Тип Запису   | Версія UUID | Примітки       |
| ------------ | ----------- | -------------- |
| Старі записи | UUID v4     | До міграції    |
| Нові записи  | UUID v7     | Після міграції |

## Міркування Щодо Продуктивності

### Коли Використовувати UUID v7

[+] **Добре підходить:**

- Розподілені системи де кілька вузлів генерують ID
- Мульти-регіональні розгортання
- Архітектура мікросервісів
- API де клієнти потребують генерувати ID
- Таблиці з високою частотою INSERT
- Коли потрібне природне впорядкування за часом

[X] **Перегляньте якщо:**

- Однобазова система без потреб розподілення
- Зберігання надзвичайно обмежене (UUID = 16 байт проти BIGINT = 8 байт)
- Потрібен аналіз послідовних пропусків (використовуйте SERIAL)

### Оптимізація Індексів PostgreSQL

UUID v7 працює найкраще з індексами B-tree завдяки впорядкуванню за часом:

```sql
-- Добре: Первинний ключ автоматично використовує B-tree
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7()
);

-- Добре: Запити на основі часу виграють від впорядкування UUID v7
CREATE INDEX idx_users_created ON users(created_at DESC);

-- Оптимально: Комбінуйте UUID v7 з timestamp для запитів діапазону
SELECT * FROM users
WHERE id >= uuid_generate_v7_from_timestamp('2024-01-01'::timestamp)
ORDER BY id DESC;
```

### Моніторинг Продуктивності

Порівняйте продуктивність INSERT до/після міграції:

```sql
-- Перевірити роздування індексу
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE tablename IN ('users', 'products', 'roles')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Моніторити продуктивність INSERT
EXPLAIN ANALYZE
INSERT INTO users (email, name, password)
VALUES ('test@example.com', 'Test', 'hash');
```

## Рекомендації для Зовнішніх Ключів

### Використовуйте UUID v7 і для Зовнішніх Ключів

```go
type Product struct {
    ID        uuidv7.UUID `db:"id"`        // UUID v7 первинний ключ
    UserID    uuidv7.UUID `db:"user_id"`   // UUID v7 зовнішній ключ
    Name      string      `db:"name"`
    CreatedAt time.Time   `db:"created_at"`
}
```

### Індексування Зовнішніх Ключів

Завжди індексуйте зовнішні ключі для продуктивності JOIN:

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL
);

-- Критично: Індексувати зовнішній ключ
CREATE INDEX idx_products_user_id ON products(user_id);
```

### Складені Ключі з UUID v7

Для відношень many-to-many:

```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Індекс для зворотного пошуку
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

## Альтернативні Стратегії (Майбутнє Розгляд)

### Варіант 1: ULID (Universally Unique Lexicographically Sortable Identifier)

Схожий на UUID v7, але використовує кодування base32 (коротше строкове представлення):

```
01ARZ3NDEKTSV4RRFFQ69G5FAV  (ULID - 26 символів)
018d2f07-0f3c-7000-8000-123456789abc  (UUID v7 - 36 символів)
```

**Коли використовувати**: Якщо ви часто показуєте ID користувачам і хочете коротші рядки.

### Варіант 2: Snowflake IDs

Snowflake від Twitter: 64-бітні цілі числа з timestamp, datacenter ID і послідовністю.

```go
// Приклад структури (не реалізовано в цьому проєкті)
// 41 біт: timestamp
// 10 біт: datacenter/worker ID
// 12 біт: номер послідовності
```

**Коли використовувати**: Системи з надзвичайно високою пропускною здатністю (масштаб Twitter) де координація прийнятна.

### Варіант 3: PostgreSQL SERIAL/BIGSERIAL

Традиційні автоінкрементні цілі числа:

```sql
CREATE TABLE simple_table (
    id BIGSERIAL PRIMARY KEY,  -- 1, 2, 3, 4, ...
    name VARCHAR(255)
);
```

**Коли використовувати**: Прості однобазові застосунки без потреб розподілення.

### Варіант 4: Складені Природні Ключі

Використовуйте значущі бізнес-дані як ключі:

```sql
CREATE TABLE country_codes (
    code CHAR(2) PRIMARY KEY,  -- 'US', 'GB', 'FR'
    name VARCHAR(100)
);
```

**Коли використовувати**: Коли природні ключі стабільні, короткі і справді унікальні.

## Тестування

Запустіть набір тестів для перевірки реалізації UUID v7:

```bash
# Тестувати пакет UUID v7
go test -v ./pkg/uuidv7/

# Запустити з бенчмарками
go test -bench=. ./pkg/uuidv7/

# Тестувати інтеграцію з базою даних
make test-integration
```

Очікувані результати бенчмарків (приблизно):

```
BenchmarkNew-8      3000000    450 ns/op    (UUID v7)
BenchmarkNewV4-8    2800000    480 ns/op    (UUID v4)
```

## Чеклист Міграції

- [x] Додати функцію `uuid_generate_v7()` до PostgreSQL
- [x] Оновити значення таблиць за замовчуванням на UUID v7
- [x] Оновити код Go для використання пакету `pkg/uuidv7`
- [x] Оновити всі use cases (auth, product, role)
- [ ] Запустити міграції: `make migrate-up`
- [ ] Тестувати генерацію UUID: `go test ./pkg/uuidv7/`
- [ ] Моніторити метрики продакшн після розгортання

## Додаткове Читання

- [RFC 9562 - UUID Version 7 (Draft)](https://datatracker.ietf.org/doc/draft-ietf-uuidrev-rfc4122bis/)
- [PostgreSQL UUID Performance](https://www.2ndquadrant.com/en/blog/sequential-uuid-generators/)
- [UUID v7 in Production (Blog post)](https://buildkite.com/blog/goodbye-integers-hello-uuids)

## Питання?

Дивіться `.github/copilot-instructions.md` для конвенцій проєкту або запитуйте в командному чаті.
