---
title: "Тестова Інфраструктура"
description: "400+ тестів з ручним підходом до моків"
weight: 5
---

## Комплексне Тестування

Promenade має **400+ тестів**, що покривають всі шари з підходом **ручних моків**.

### Покриття Тестами

**Тести Ядра (275 тестів):**

- Доменні сутності: 39 тестів
- Use cases: 236 тестів (auth, RBAC, референсні дані)

**Тести Модулів:**

- Posts: 33 тести (83.3% покриття)
- Profiles: 21 тест (80.4% покриття)
- Analytics: 11 тестів
- Billing: 375 тестів (100% покриття)

**Утиліти:** 51 тест, 89.5% середнє покриття

### Типи Тестів

```bash
# Всі тести (~20 секунд)
make test

# Тільки тести ядра
make test-core

# Тести модулів
make test-module-posts
make test-module-billing

# Звіт покриття
make test-coverage
```

### Шаблон Ручних Моків

**Без mockgen** - прості, явні моки:

```go
// Мок репозиторію inline
type mockUserRepo struct {
    users map[string]*User
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, ErrUserNotFound
}

// Використання в тестах
func TestRegisterUser(t *testing.T) {
    repo := &mockUserRepo{users: make(map[string]*User)}
    usecase := NewUserUseCase(repo)

    user, err := usecase.Register(ctx, "test@example.com", "password")
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### Інтеграційні Тести

Тести з **реальним PostgreSQL**:

```bash
# Запустити тестову базу даних
make test-db-start

# Запустити інтеграційні тести
make test-integration

# Зупинити тестову базу даних
make test-db-stop
```

### Переваги

✅ **Швидко** - Повний набір тестів за ~20 секунд  
✅ **Явно** - Немає магії генераторів, зрозумілі моки  
✅ **Надійно** - Інтеграційні тести з реальною БД  
✅ **Покриття** - 89.5% середнє покриття коду

[Посібник з тестування →](/docs/TESTING_GUIDE)
