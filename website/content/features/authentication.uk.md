---
title: "Автентифікація та RBAC"
description: "JWT автентифікація з рольовим контролем доступу"
weight: 3
---

## Автентифікація та Авторизація

Промислова автентифікація з **JWT токенами** та **RBAC** (Role-Based Access Control - рольовий контроль доступу).

### JWT Автентифікація

```go
// Login повертає access + refresh токени
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "secure123"
}

// Відповідь
{
  "access_token": "eyJhbGc...",   // 15 хвилин
  "refresh_token": "eyJhbGc...",  // 7 днів
  "user": { ... }
}
```

### RBAC Система

**4 Вбудовані Ролі:**

- **Admin** - Повний доступ до системи (`*` дозвіл)
- **Moderator** - Модерація контенту
- **User** - Базові операції
- **Guest** - Тільки читання

**Формат Дозволів:** `ресурс:дія`

```
posts:create
posts:update
posts:delete
users:manage
*  # Wildcard - повний доступ
```

### Захист Middleware

```go
// Вимагати автентифікації
router.Use(authMiddleware.RequireAuth())

// Вимагати конкретний дозвіл
router.POST("/posts",
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

### Управління Сесіями

- Кілька сесій на користувача
- Відкликання сесій
- Відстеження пристроїв
- Моніторинг активності

### Тестові Користувачі за Замовчуванням

```
admin@promenade.com     | passw0rd | Admin
moderator@promenade.com | passw0rd | Moderator
user@promenade.com      | passw0rd | User
```

### Переваги

✅ **Безпечно** - JWT з підписом HMAC-SHA256  
✅ **Гнучко** - Wildcard та деталізовані дозволи  
✅ **Масштабовано** - Stateless токени, опціональне відстеження сесій  
✅ **Готово до продакшн** - Тестовано в реальних проектах

[Детальна документація →](/docs/AUTHORIZATION)
