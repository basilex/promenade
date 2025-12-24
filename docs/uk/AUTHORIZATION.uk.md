# Посібник з Middleware Авторизації

Повний посібник з використання RBAC (Role-Based Access Control) middleware авторизації в Promenade.

## Зміст

- [Огляд](#огляд)
- [Архітектура](#архітектура)
- [Система Дозволів](#система-дозволів)
- [Система Ролей](#система-ролей)
- [Методи Middleware](#методи-middleware)
- [Приклади Використання](#приклади-використання)
- [Найкращі Практики](#найкращі-практики)
- [Типові Патерни](#типові-патерни)
- [Обробка Помилок](#обробка-помилок)
- [Тестування Авторизації](#тестування-авторизації)

## Огляд

Middleware Авторизації надає **гнучкий, детальний контроль доступу** для API ендпоінтів, використовуючи систему RBAC на основі дозволів. Підтримує:

- [+] **Перевірки на основі дозволів** - Детальний контроль у форматі `resource:action`
- [+] **Перевірки на основі ролей** - Швидкі перевірки ролей користувачів (admin, moderator тощо)
- [+] **Дозволи з wildcards** - Шаблони `*:*`, `posts:*`, `*:read`
- [+] **Композитні перевірки** - RequireAny, RequireAll для складної логіки
- [+] **Чітке розділення** - Працює незалежно від middleware автентифікації

## Архітектура

**Потік Запиту:**

1. **HTTP Запит** надходить
   ↓
2. **RequireAuth Middleware**
   - Валідує JWT токен
   - Встановлює user_id у контекст
     ↓
3. **RequirePermission Middleware**
   - Отримує user_id з контексту
   - Запитує ролі користувача
   - Перевіряє дозволи ролі (з підтримкою wildcards)
   - Дозволяє/Забороняє запит
     ↓
4. **Handler Function** виконується

## Система Дозволів

### Формат Дозволів

Дозволи слідують шаблону `resource:action`:

```
resource:action
   │       │
   │       └─ Дія: create, read, update, delete, manage, *
   └───────── Ресурс: posts, users, comments, roles, *
```

### Приклади

| Дозвіл              | Опис                                              |
| ------------------- | ------------------------------------------------- |
| `posts:create`      | Може створювати пости                             |
| `posts:read`        | Може читати пости                                 |
| `posts:*`           | Може виконувати будь-яку дію з постами            |
| `*:read`            | Може читати будь-який ресурс                      |
| `*:*`               | Може виконувати будь-яку дію з будь-яким ресурсом |
| `users:ban`         | Може банити користувачів (власна дія)             |
| `comments:moderate` | Може модерувати коментарі                         |

### Wildcards у Дозволах

Wildcards надають потужне успадкування дозволів:

```go
// Користувач має дозвіл "posts:*"
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "posts:update")  // [+] TRUE
HasPermission(userID, "posts:delete")  // [+] TRUE
HasPermission(userID, "users:create")  // [X] FALSE

// Користувач має дозвіл "*:read"
HasPermission(userID, "posts:read")    // [+] TRUE
HasPermission(userID, "users:read")    // [+] TRUE
HasPermission(userID, "posts:create")  // [X] FALSE

// Користувач має дозвіл "*:*" (адмін з повним доступом)
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "users:delete")  // [+] TRUE
HasPermission(userID, "anything:anything") // [+] TRUE
```

## Система Ролей

### Системні Ролі

4 попередньо визначені системні ролі з різними рівнями дозволів:

| Роль        | Відображуване Ім'я | Дозволи                      | Випадок Використання                |
| ----------- | ------------------ | ---------------------------- | ----------------------------------- |
| `admin`     | Адміністратор      | `*:*` (всі)                  | Повний системний доступ             |
| `moderator` | Модератор          | Модерація контенту           | Перегляд і модерація контенту       |
| `user`      | Користувач         | Управління власним контентом | Звичайні користувачі                |
| `guest`     | Гість              | Доступ тільки для читання    | Неавторизовані/обмежені користувачі |

### Розподіл Дозволів за Ролями

**Admin** (`*:*`):

- Повний доступ до всього
- Не може бути видалений (системна роль)

**Admin**:

```
users:create, users:read, users:update, users:delete, users:ban, users:suspend
roles:read, roles:assign
permissions:read
posts:*, comments:*, profiles:*
```

**Moderator**:

```
posts:read, posts:update, posts:delete
comments:read, comments:update, comments:delete, comments:moderate
users:read, users:suspend
```

**User**:

```
posts:create, posts:read, posts:update (власні), posts:delete (власні)
comments:create, comments:read, comments:update (власні), comments:delete (власні)
profiles:read, profiles:update (власні)
```

**Guest**:

```
posts:read, comments:read, profiles:read
```

## Методи Middleware

### RequirePermission

Перевірити, чи має користувач **один конкретний дозвіл**.

```go
func (m *AuthorizationMiddleware) RequirePermission(permission string) gin.HandlerFunc
```

**Використання:**

```go
router.POST("/posts",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

**Повертає:**

- `200 OK` - Користувач має дозвіл, переходить до handler
- `401 Unauthorized` - Користувач не автентифікований
- `403 Forbidden` - Користувачу бракує дозволу
- `500 Internal Server Error` - Помилка бази даних при перевірці дозволів

---

### RequireAnyPermission

Перевірити, чи має користувач **принаймні один** з вказаних дозволів (логіка АБО).

```go
func (m *AuthorizationMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc
```

**Використання:**

```go
// Дозволити, якщо користувач може читати АБО модерувати коментарі
router.GET("/comments/flagged",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission("comments:read", "comments:moderate"),
    handler.GetFlaggedComments,
)
```

**Випадки Використання:**

- Альтернативні дозволи (admin АБО moderator)
- Доступ до функцій з кількома точками входу
- Поступове підвищення дозволів

---

### RequireAllPermissions

Перевірити, чи має користувач **всі** вказані дозволи (логіка ТА).

```go
func (m *AuthorizationMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc
```

**Використання:**

```go
// Вимагати дозволів publish ТА schedule
router.POST("/posts/schedule",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions("posts:create", "posts:schedule"),
    handler.SchedulePost,
)
```

**Випадки Використання:**

- Складені операції, що вимагають кількох дозволів
- Чутливі операції, що потребують багаторівневих перевірок
- Комбінації функцій

---

### RequireRole

Перевірити, чи має користувач **конкретну роль** за іменем.

```go
func (m *AuthorizationMiddleware) RequireRole(roleName string) gin.HandlerFunc
```

**Використання:**

```go
// Тільки адміністратори можуть отримати доступ
router.GET("/admin/dashboard",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.GetAdminDashboard,
)
```

**Примітка:** Надавайте перевагу `RequirePermission` замість `RequireRole` для кращої гнучкості.

---

### RequireAnyRole

Перевірити, чи має користувач **принаймні одну** з вказаних ролей (логіка АБО).

```go
func (m *AuthorizationMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc
```

**Використання:**

```go
// Дозволити адмінам АБО модераторам
router.GET("/moderation/queue",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyRole("admin", "moderator"),
    handler.GetModerationQueue,
)
```

**Випадки Використання:**

- Адміністративні зони з кількома рівнями ролей
- Доступ до функцій для подібних ролей
- Застарілі системи, що переходять з ролей на дозволи

## Приклади Використання

### Приклад 1: Базовий Захист CRUD

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")

    // Публічний доступ для читання (не потребує auth)
    posts.GET("", handler.ListPosts)
    posts.GET("/:id", handler.GetPost)

    // Операції для автентифікованих
    posts.Use(r.authMiddleware.RequireAuth())
    {
        // Конкретні дозволи для кожної операції
        posts.POST("",
            r.authzMiddleware.RequirePermission("posts:create"),
            handler.CreatePost,
        )

        posts.PUT("/:id",
            r.authzMiddleware.RequirePermission("posts:update"),
            handler.UpdatePost,
        )

        posts.DELETE("/:id",
            r.authzMiddleware.RequirePermission("posts:delete"),
            handler.DeletePost,
        )
    }
}
```

### Приклад 2: Ендпоінти Тільки для Адміністраторів

```go
func (r *UserRouter) Setup(api *gin.RouterGroup) {
    users := api.Group("/users")
    users.Use(r.authMiddleware.RequireAuth())

    // Операції звичайних користувачів
    users.GET("/me", handler.GetMe)
    users.PUT("/me", handler.UpdateProfile)

    // Операції тільки для адміністраторів
    admin := users.Group("")
    admin.Use(r.authzMiddleware.RequirePermission("users:manage"))
    {
        admin.GET("", handler.ListAllUsers)
        admin.POST("/:id/ban", handler.BanUser)
        admin.POST("/:id/suspend", handler.SuspendUser)
    }
}
```

### Приклад 3: Гнучкий Доступ з Кількома Дозволами

```go
func (r *CommentRouter) Setup(api *gin.RouterGroup) {
    comments := api.Group("/comments")

    // Перегляд коментарів - працює будь-який з цих дозволів
    comments.GET("/:id",
        r.authMiddleware.RequireAuth(),
        r.authzMiddleware.RequireAnyPermission(
            "comments:read",
            "comments:moderate",
            "*:read",
        ),
        handler.GetComment,
    )

    // Модерація коментарів - потрібні обидва дозволи read та moderate
    comments.POST("/:id/moderate",
        r.authMiddleware.RequireAuth(),
        r.authzMiddleware.RequireAllPermissions(
            "comments:read",
            "comments:moderate",
        ),
        handler.ModerateComment,
    )
}
```

### Приклад 4: Доступ до Панелі на Основі Ролі

```go
func (r *DashboardRouter) Setup(api *gin.RouterGroup) {
    dashboards := api.Group("/dashboard")
    dashboards.Use(r.authMiddleware.RequireAuth())

    // Панель користувача - будь-який автентифікований користувач
    dashboards.GET("/user", handler.GetUserDashboard)

    // Панель модератора - модератори та адміни
    dashboards.GET("/moderator",
        r.authzMiddleware.RequireAnyRole("moderator", "admin"),
        handler.GetModeratorDashboard,
    )

    // Панель адміна - тільки адміни
    dashboards.GET("/admin",
        r.authzMiddleware.RequireRole("admin"),
        handler.GetAdminDashboard,
    )
}
```

### Приклад 5: Складна Бізнес-Логіка

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")
    posts.Use(r.authMiddleware.RequireAuth())

    // Публікація вимагає дозволів create та publish
    posts.POST("/:id/publish",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
        ),
        handler.PublishPost,
    )

    // Планування вимагає create, publish ТА schedule
    posts.POST("/:id/schedule",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
            "posts:schedule",
        ),
        handler.SchedulePost,
    )

    // Виділення вимагає ролі moderator АБО admin + дозволу feature
    posts.POST("/:id/feature",
        r.authzMiddleware.RequireAnyRole("admin", "moderator"),
        r.authzMiddleware.RequirePermission("posts:feature"),
        handler.FeaturePost,
    )
}
```

### Приклад 6: Налаштування Міграції

Ініціалізація авторизації в налаштуванні роутера:

```go
// cmd/api/main.go або ініціалізація роутера
func setupRouters(
    authMiddleware *middleware.AuthMiddleware,
    authzMiddleware *middleware.AuthorizationMiddleware,
) *gin.Engine {
    r := gin.New()

    // Публічні роути
    api := r.Group("/api/v1")

    // Auth роути (авторизація не потрібна)
    authRouter := router.NewAuthRouter(authHandler, authMiddleware)
    authRouter.Setup(api)

    // Захищені роути з авторизацією
    postRouter := router.NewPostRouter(postHandler, authMiddleware, authzMiddleware)
    postRouter.Setup(api)

    userRouter := router.NewUserRouter(userHandler, authMiddleware, authzMiddleware)
    userRouter.Setup(api)

    return r
}
```

## Найкращі Практики

### 1. Завжди Використовуйте RequireAuth Спочатку

Middleware авторизації потребує контексту автентифікації:

```go
// [+] ПРАВИЛЬНО - Auth перед авторизацією
router.POST("/posts",
    authMiddleware.RequireAuth(),           // Спочатку: автентифікація
    authzMiddleware.RequirePermission(...), // Потім: авторизація
    handler.CreatePost,
)

// [X] НЕПРАВИЛЬНО - Авторизація без автентифікації
router.POST("/posts",
    authzMiddleware.RequirePermission(...), // Провалиться - немає user_id
    handler.CreatePost,
)
```

### 2. Надавайте Перевагу Дозволам Замість Ролей

Дозволи забезпечують кращу гнучкість та підтримуваність:

```go
// [+] КРАЩЕ - На основі дозволів (гнучко)
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:delete"),
    handler.DeletePost,
)

// [!] ПРИЙНЯТНО але менш гнучко - На основі ролей
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.DeletePost,
)
```

**Чому?**

- Додавання нових ролей не вимагає змін коду
- Дозволи можна перепризначити без торкання коду
- Більш детальний контроль

### 3. Використовуйте Описові Назви Дозволів

```go
// [+] ДОБРЕ - Чіткий намір
"posts:create"
"posts:publish"
"posts:feature"
"posts:schedule"

// [X] ПОГАНО - Нечітко
"posts:manage"  // Що означає "manage"?
"posts:admin"   // Занадто загально
```

### 4. Використовуйте Wildcards для Ролей Адміністраторів

```go
// У вашій міграції/seed
INSERT INTO permissions (resource, action) VALUES
    ('*', '*'),           -- Admin: все
    ('posts', '*'),       -- Контент адмін: всі операції з постами
    ('*', 'read');        -- Переглядач: читання всього
```

### 5. Групуйте Пов'язані Дозволи

```go
// Групування за функціональною областю
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Всі операції запису постів вимагають posts:* або posts:write
write := posts.Group("")
write.Use(authzMiddleware.RequirePermission("posts:write"))
{
    write.POST("", handler.CreatePost)
    write.PUT("/:id", handler.UpdatePost)
    write.DELETE("/:id", handler.DeletePost)
}

// Публічні операції читання
posts.GET("", handler.ListPosts)
posts.GET("/:id", handler.GetPost)
```

### 6. Обробляйте Ресурси Власника в Handler

Не використовуйте middleware авторизації для перевірок власника:

```go
// [+] ПРАВИЛЬНО - Перевіряйте власність у handler
func (h *PostHandler) UpdatePost(c *gin.Context) {
    userID := middleware.GetUserIDOrPanic(c)
    postID := c.Param("id")

    post, err := h.postUC.GetByID(c.Request.Context(), postID)
    if err != nil {
        response.Error(c, http.StatusNotFound, "post not found", err)
        return
    }

    // Перевірка власності АБО дозволу адміна
    if post.UserID != userID {
        hasAdmin, _ := h.roleUC.HasPermission(c.Request.Context(), userID, "posts:*")
        if !hasAdmin {
            response.Error(c, http.StatusForbidden, "can only update own posts", nil)
            return
        }
    }

    // Продовження оновлення...
}

// [X] НЕПРАВИЛЬНО - Спроба перевірити власність у middleware
// Middleware не має доступу до деталей ресурсу
```

### 7. Використовуйте RequireAny для Запасних Дозволів

```go
// Дозволити операцію, якщо користувач має конкретний дозвіл АБО є адміном
router.POST("/posts/:id/feature",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission(
        "posts:feature",  // Конкретний дозвіл
        "posts:*",        // Повний доступ до постів
        "*:*",            // Адмін
    ),
    handler.FeaturePost,
)
```

## Типові Патерни

### Патерн 1: Перевизначення Адміна

Дозволити адмінам обходити перевірки власності:

```go
// Будь-який дозвіл адміна перевизначає власність
authzMiddleware.RequireAnyPermission(
    "posts:update",  // Дозвіл звичайного користувача
    "posts:*",       // Адмін постів
    "*:*",           // Адмін
)
```

### Патерн 2: Поступові Дозволи

Різні рівні дозволів для одного ресурсу:

```go
// Рівень 1: Базове читання
authzMiddleware.RequirePermission("posts:read")

// Рівень 2: Читання + запис
authzMiddleware.RequireAllPermissions("posts:read", "posts:write")

// Рівень 3: Повний доступ
authzMiddleware.RequirePermission("posts:*")
```

### Патерн 3: Міжресурсні Дозволи

Операції, що впливають на кілька ресурсів:

```go
// Публікація посту може вимагати дозволів і для посту, і для медіа
router.POST("/posts/:id/publish",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions(
        "posts:publish",
        "media:attach",  // Якщо пост містить зображення
    ),
    handler.PublishPost,
)
```

### Патерн 4: Умовна Авторизація

Різні дозволи для різних ендпоінтів:

```go
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Чернетки постів - тільки дозвіл create
posts.POST("/drafts",
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreateDraft,
)

// Опубліковані пости - дозволи create + publish
posts.POST("/publish",
    authzMiddleware.RequireAllPermissions("posts:create", "posts:publish"),
    handler.CreateAndPublish,
)
```

## Обробка Помилок

### HTTP Коди Статусу

| Статус | Значення              | Причина                                         |
| ------ | --------------------- | ----------------------------------------------- |
| 401    | Unauthorized          | Користувач не автентифікований (немає JWT)      |
| 403    | Forbidden             | Користувач автентифікований, але бракує дозволу |
| 500    | Internal Server Error | Помилка бази даних при перевірці дозволів       |

### Формат Відповіді Помилки

```json
{
  "error": "insufficient permissions",
  "message": "You don't have permission to perform this action"
}
```

### Обробка на Стороні Клієнта

```typescript
// Приклад TypeScript/JavaScript
try {
  const response = await fetch("/api/v1/posts", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(postData),
  });

  if (response.status === 401) {
    // Перенаправлення на логін
    window.location.href = "/login";
  } else if (response.status === 403) {
    // Показати повідомлення "insufficient permissions"
    showError("You do not have permission to create posts");
  } else if (response.ok) {
    // Успіх
    const post = await response.json();
  }
} catch (error) {
  console.error("Request failed:", error);
}
```

## Тестування Авторизації

### Unit Тестування Middleware

```go
func TestAuthorizationMiddleware_RequirePermission(t *testing.T) {
    // Налаштування
    mockRoleUC := mocks.NewMockRoleUseCase(t)
    authzMiddleware := middleware.NewAuthorizationMiddleware(mockRoleUC)

    t.Run("allows user with permission", func(t *testing.T) {
        // Створення тестового контексту з user_id
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Set("user_id", testUserID)

        // Mock перевірки дозволу - повертає true
        mockRoleUC.EXPECT().
            HasPermission(mock.Anything, testUserID, "posts:create").
            Return(true, nil)

        // Створення ланцюга handler
        handler := authzMiddleware.RequirePermission("posts:create")(func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })

        // Виконання
        handler(c)

        // Перевірка
        assert.Equal(t, 200, w.Code)
    })

    t.Run("denies user without permission", func(t *testing.T) {
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Set("user_id", testUserID)

        mockRoleUC.EXPECT().
            HasPermission(mock.Anything, testUserID, "posts:create").
            Return(false, nil)

        handler := authzMiddleware.RequirePermission("posts:create")(func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })

        handler(c)

        assert.Equal(t, 403, w.Code)
    })
}
```

### Інтеграційне Тестування

```go
func TestPostEndpoints_Authorization(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    // Створення тестових користувачів з різними ролями
    adminUser := helpers.UserFixture(t, testDB.DB)
    regularUser := helpers.UserFixture(t, testDB.DB)

    // Призначення ролей
    assignRole(t, testDB, adminUser.ID, "admin")
    assignRole(t, testDB, regularUser.ID, "user")

    // Генерація токенів
    adminToken := generateToken(t, adminUser.ID)
    userToken := generateToken(t, regularUser.ID)

    t.Run("admin can delete any post", func(t *testing.T) {
        post := createTestPost(t, testDB, regularUser.ID)

        req := httptest.NewRequest("DELETE", "/api/v1/posts/"+post.ID, nil)
        req.Header.Set("Authorization", "Bearer "+adminToken)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, 200, w.Code)
    })

    t.Run("user cannot delete others' posts", func(t *testing.T) {
        post := createTestPost(t, testDB, adminUser.ID)

        req := httptest.NewRequest("DELETE", "/api/v1/posts/"+post.ID, nil)
        req.Header.Set("Authorization", "Bearer "+userToken)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, 403, w.Code)
    })
}
```

### Ручне Тестування з curl

```bash
# 1. Логін та отримання токена
TOKEN=$(curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# 2. Тестування захищеного ендпоінту
curl -X POST http://localhost:8081/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Post","content":"Content"}'

# Очікувані відповіді:
# 200 OK - Успіх
# 401 Unauthorized - Невірний/відсутній токен
# 403 Forbidden - Недостатньо дозволів
```

## Усунення Неполадок

### Проблема: 401 Unauthorized на Захищеному Ендпоінті

**Симптоми:**

```json
{
  "error": "user not authenticated"
}
```

**Причини:**

1. Відсутній middleware `RequireAuth()` перед `RequirePermission()`
2. Невірний JWT токен
3. Прострочений токен

**Рішення:**

```go
// Переконайтеся, що auth middleware застосовано першим
router.POST("/posts",
    authMiddleware.RequireAuth(),  // ← Має бути перед авторизацією
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

### Проблема: 403 Forbidden для Адміна

**Симптоми:**
Користувач адміна отримує 403 на ендпоінтах, до яких має доступ.

**Причини:**

1. Wildcard дозвіл (`*:*`) не перевіряється належним чином
2. Роль не призначена користувачу
3. Дозвіл не призначений ролі

**Рішення:**

```sql
-- Перевірте, що адмін має wildcard дозвіл
SELECT r.name, p.resource, p.action
FROM roles r
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions p ON p.id = rp.permission_id
WHERE r.name = 'admin';

-- Має повернути: name='admin', resource='*', action='*'

-- Перевірте, що користувач має роль адміна
SELECT u.email, r.name
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
WHERE u.id = '<user_uuid>';
```

### Проблема: Продуктивність Бази Даних з Перевірками Дозволів

**Симптоми:**
Повільний час відповіді на захищених ендпоінтах.

**Рішення:**
Впровадьте кешування в RoleUseCase:

```go
// Використовуйте Redis/кеш у пам'яті для перевірок дозволів
func (uc *RoleUseCase) HasPermission(ctx context.Context, userID uuidv7.UUID, permission string) (bool, error) {
    // Спочатку перевірте кеш
    cacheKey := fmt.Sprintf("user:%s:permission:%s", userID, permission)
    if cached, found := uc.cache.Get(cacheKey); found {
        return cached.(bool), nil
    }

    // Запит до бази даних
    hasPermission, err := uc.repo.HasPermission(ctx, userID, permission)
    if err != nil {
        return false, err
    }

    // Кешувати на 5 хвилин
    uc.cache.Set(cacheKey, hasPermission, 5*time.Minute)

    return hasPermission, nil
}
```

## Підсумок

Middleware Авторизації надає потужний, гнучкий контроль доступу для вашого API:

[+] **На основі дозволів** - Детальний контроль у форматі `resource:action`  
[+] **Підтримка wildcards** - Потужне успадкування з шаблонами `*`  
[+] **Композитні перевірки** - Логіка ТА/АБО для складних вимог  
[+] **Скорочення для ролей** - Швидкі перевірки на основі ролей за потреби  
[+] **Чиста архітектура** - Відокремлює авторизацію від автентифікації  
[+] **Готовність до продакшену** - Перевірена обробка помилок та продуктивність

**Швидка Довідка:**

```go
// Перевірка одного дозволу
RequirePermission("posts:create")

// Будь-який з кількох дозволів (АБО)
RequireAnyPermission("posts:read", "posts:*", "*:*")

// Всі з кількох дозволів (ТА)
RequireAllPermissions("posts:create", "posts:publish")

// Перевірка на основі ролі
RequireRole("admin")

// Будь-яка з кількох ролей (АБО)
RequireAnyRole("admin", "moderator")
```

Для додаткової інформації дивіться:

- [RBAC Implementation](RBAC_IMPLEMENTATION.md)
- [Testing Guide](TESTING_GUIDE.uk.md)
- [API Documentation](../README.uk.md)
