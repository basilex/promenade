---
title: "База Даних та Міграції"
description: "PostgreSQL з UUID v7 та міграції на основі просторів імен"
weight: 4
---

## Архітектура Бази Даних

Promenade використовує **PostgreSQL 16** з **UUID v7** первинними ключами та **міграціями на основі просторів імен**.

### UUID v7 - Впорядковані за Часом ID

На відміну від випадкових UUID v4, **UUID v7 впорядковані за часом**:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),  -- Не uuid_generate_v4()!
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Переваги:**

- ⚡ **У 2 рази швидші вставки** - краща локальність B-дерева
- 📉 **Зменшена фрагментація індексів**
- 🔍 **Сортування за часом створення**
- 🆔 **Все ще глобально унікальні**

### Міграції на Основі Просторів Імен

Кожен модуль має **незалежну історію міграцій**:

```
migrations/
├── core/                    # Ядро інфраструктури
│   ├── 000001_auth.sql
│   └── 000002_rbac.sql
├── posts/                   # Модуль постів
│   ├── 000001_posts.sql
│   └── 000002_comments.sql
└── billing/                 # Модуль білінгу
    └── 000001_tables.sql
```

### Справжня Автономія Модулів

```bash
# Мігрувати всі увімкнені модулі
make migrate

# Мігрувати конкретний модуль
make migrate-module MODULE=posts

# Створити нову міграцію
make migrate-create MODULE=posts NAME=add_views_count
```

### Без ORM - Чистий SQL

Promenade використовує **sqlx** (не ORM):

```go
// Чисті, явні запити
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
    return &user, r.Get(ctx, &user, query, email)
}
```

**Чому Без ORM:**

- Повний контроль SQL
- Немає прихованих N+1 запитів
- Явне налаштування продуктивності
- Нульові витрати на абстракцію

### Шаблон BaseRepository

```go
type BaseRepository struct {
    db *sqlx.DB
}

// Спільні методи: Get, Select, Exec, NamedExec
// Кожен репозиторій вбудовує BaseRepository
```

### Переваги

✅ **Швидко** - UUID v7 покращує продуктивність вставок  
✅ **Незалежно** - Міграції модулів не конфліктують  
✅ **Явно** - Без магії ORM, повний контроль SQL  
✅ **Транзакційно** - Підтримка транзакцій з урахуванням контексту

[Детальний посібник →](/promenade/docs/UUID_V7_GUIDE)
