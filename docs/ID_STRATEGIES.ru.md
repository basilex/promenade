# Рекомендации по стратегиям первичных ключей для Promenade

## Краткое резюме

**Текущее решение:** UUID v7 (миграция с v4 завершена)

**Почему UUID v7:**

1. Time-ordered = лучшая производительность индексов (20-50% быстрее INSERT)
2. Сохраняет распределенную генерацию UUID v4
3. В 2x быстрее самого v4 в генерации (87ns vs 185ns)
4. Нулевые аллокации памяти
5. Можно извлечь timestamp создания

---

## Сравнение стратегий ID

### 1. UUID v7 (текущий выбор) ⭐

**Плюсы:**

- ✅ Time-ordered для B-tree индексов
- ✅ Распределенная генерация без координации
- ✅ Глобальная уникальность
- ✅ Извлекаемый timestamp
- ✅ Совместимость с UUID v4
- ✅ Производительность лучше v4

**Минусы:**

- ❌ 16 байт (vs 8 байт у BIGINT)
- ❌ 36 символов в строковом виде

**Используй когда:**

- Микросервисная архитектура
- Несколько источников записи
- Мультирегиональность
- Нужна временная сортировка
- API где клиенты генерируют ID

---

### 2. UUID v4 (random) - Устарел

**Плюсы:**

- ✅ Глобальная уникальность
- ✅ Распределенная генерация

**Минусы:**

- ❌ Плохая локальность в индексах
- ❌ Много page splits в B-tree
- ❌ Фрагментация индекса
- ❌ Медленнее v7

**Статус:** Заменен на v7 в этом проекте

---

### 3. PostgreSQL SERIAL / BIGSERIAL

```sql
CREATE TABLE simple (
    id BIGSERIAL PRIMARY KEY,  -- 1, 2, 3, 4...
    name TEXT
);
```

**Плюсы:**

- ✅ Маленький размер (8 байт)
- ✅ Последовательный (отличная локальность)
- ✅ Читаемый людьми
- ✅ Максимальная производительность для single-node

**Минусы:**

- ❌ Single point of failure (PostgreSQL должен генерировать)
- ❌ Проблемы с репликацией (конфликты ID)
- ❌ Сложности при шардировании
- ❌ Утечка информации (можно посчитать количество записей)

**Используй когда:**

- Простое приложение с одной БД
- Не планируется шардирование
- Не критична утечка количества записей
- Нужна максимальная эффективность хранения

---

### 4. ULID (Universally Unique Lexicographically Sortable Identifier)

```
01ARZ3NDEKTSV4RRFFQ69G5FAV  (26 символов)
```

**Плюсы:**

- ✅ Time-ordered как UUID v7
- ✅ Короче строкового UUID (26 vs 36 символов)
- ✅ Base32 encoding (URL-safe без экранирования)
- ✅ Case-insensitive

**Минусы:**

- ❌ Нет нативной поддержки в PostgreSQL
- ❌ Меньше библиотек поддержки
- ❌ Все равно 16 байт в БД

**Используй когда:**

- ID часто передаются в URL
- Важна читаемость
- Хочется избежать дефисов в UUID

**Пример библиотеки:** `github.com/oklog/ulid`

---

### 5. Snowflake ID (Twitter)

```
64-bit integer:
┌─────────────┬──────────┬────────────┐
│ Timestamp   │ Machine  │ Sequence   │
│ 41 bits     │ 10 bits  │ 12 bits    │
└─────────────┴──────────┴────────────┘
```

**Плюсы:**

- ✅ Time-ordered
- ✅ Компактный (8 байт)
- ✅ Очень быстрая генерация
- ✅ Можно извлечь timestamp

**Минусы:**

- ❌ Требует координации (machine ID)
- ❌ Нужен центральный сервис или конфигурация
- ❌ Возможные коллизии при неправильной настройке
- ❌ Ограничение: 4096 ID/ms на машину

**Используй когда:**

- Twitter/Facebook масштаб
- Есть инфраструктура для координации
- Критичен размер (8 байт vs 16 у UUID)

**Пример библиотеки:** `github.com/bwmarrin/snowflake`

---

### 6. Composite Natural Keys

```sql
CREATE TABLE country_codes (
    code CHAR(2) PRIMARY KEY,  -- 'US', 'RU', 'GB'
    name VARCHAR(100)
);

CREATE TABLE user_emails (
    email VARCHAR(255) PRIMARY KEY,
    user_id UUID REFERENCES users(id)
);
```

**Плюсы:**

- ✅ Бизнес-значимые ключи
- ✅ Без surrogate key (меньше JOINов)
- ✅ Самодокументируемые

**Минусы:**

- ❌ Может измениться бизнес-логика
- ❌ Сложности при изменении ключа
- ❌ Проблемы с privacy (email как PK)
- ❌ Больше данных для FK

**Используй когда:**

- Ключ действительно неизменен (ISO коды)
- Справочники и словари
- Небольшие lookup-таблицы

---

## Рекомендации для внешних ключей

### UUID v7 в FK

```go
type Order struct {
    ID        uuid.UUID `db:"id"`         // PK: UUID v7
    UserID    uuid.UUID `db:"user_id"`    // FK: UUID v7
    ProductID uuid.UUID `db:"product_id"` // FK: UUID v7
}
```

```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    product_id UUID NOT NULL REFERENCES products(id)
);

-- ВАЖНО: всегда индексируй FK
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_product_id ON orders(product_id);
```

### Many-to-Many

```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Индекс для обратного поиска
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

---

## Когда менять стратегию?

### Оставаться на UUID v7 если:

- ✅ Микросервисная архитектура
- ✅ Планы на шардирование/партиционирование
- ✅ API с клиентской генерацией ID
- ✅ Мультирегиональность
- ✅ Несколько источников записи

### Рассмотреть SERIAL/BIGSERIAL если:

- Монолитное приложение на одной БД
- Нет планов на масштабирование
- Критичен размер данных (миллиарды записей)
- Нет требований к распределенной генерации

### Рассмотреть ULID если:

- ID часто в URL и важна читаемость
- Хочется избежать дефисов
- Нужна case-insensitive сортировка

### Рассмотреть Snowflake если:

- Очень высокая нагрузка (Twitter scale)
- Критичен размер БД
- Есть инфраструктура координации

---

## Best Practices

### 1. Всегда индексируй внешние ключи

```sql
-- ❌ ПЛОХО
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id)
);

-- ✅ ХОРОШО
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id)
);
CREATE INDEX idx_orders_user_id ON orders(user_id);
```

### 2. Используй ON DELETE для cascade

```sql
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id)
);
```

### 3. Добавляй timestamps

```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Даже с UUID v7 храни явный timestamp для:
-- - Точности (UUID v7 = milliseconds, timestamp = microseconds)
-- - Читаемости запросов
-- - Независимости от формата ID
```

### 4. Партиционирование с UUID v7

UUID v7 отлично подходит для временного партиционирования:

```sql
-- Партиционирование по extracted timestamp
CREATE TABLE logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    message TEXT,
    created_at TIMESTAMP GENERATED ALWAYS AS
        (to_timestamp(('x'||substring(id::text from 1 for 8))::bit(32)::bigint / 1000.0))
        STORED
) PARTITION BY RANGE (created_at);

CREATE TABLE logs_2024_01 PARTITION OF logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

---

## Итоговые рекомендации для Promenade

1. **UUID v7 - текущий стандарт**: Используй везде по умолчанию
2. **Всегда индексируй FK**: Для быстрых JOINов
3. **Храни timestamps явно**: Даже если UUID v7 содержит время
4. **Используй pkg/uuidv7**: Не генерируй вручную
5. **Документируй исключения**: Если используешь другую стратегию, объясни почему

## Миграция

См. `docs/UUID_V7_MIGRATION.md` для инструкций по миграции.

## Вопросы?

Открой issue или спроси в team chat.
