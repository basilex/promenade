[ English](../CREDENTIALS.md) |  **Українська** | [ Deutsch](CREDENTIALS.de.md) | [ Português](CREDENTIALS.pt.md) | [ Español](CREDENTIALS.es.md)

---

# Облікові дані для розробки

Швидкий довідник про користувачів за замовчуванням та облікові дані доступу з ролями RBAC.

## Користувачі за замовчуванням

### 1. Системний адміністратор (Головний адмін - Oracle Style)

```
Email:    system@promenade.com
Password: passw0rd
Role:     admin
Access:   *:* (повний доступ до системи)
```

**Використовуйте для:**

- Ініціалізації та bootstrap системи
- Управління RBAC (ролі, дозволи)
- Управління користувачами (блокування, призупинення, призначення ролей)
- Всіх адміністративних операцій

### 2. Адміністратор

```
Email:    admin@promenade.com
Password: passw0rd
Role:     admin
Access:   users:*, posts:*, comments:*, profiles:*, roles:read|list|assign
```

**Використовуйте для:**

- Управління користувачами (створення, оновлення, видалення, блокування)
- Управління контентом (пости, коментарі)
- Призначення ролей користувачам
- Тестування дозволів рівня адміністратора

### 4. Модератор

```
Email:    moderator@promenade.com
Password: passw0rd
Role:     moderator
Access:   posts:read|update|delete|list, comments:*, profiles:read|list
```

**Використовуйте для:**

- Модерації контенту (пости, коментарі)
- Управління коментарями (схвалення, видалення)
- Тестування робочих процесів модерації
- Обмеженої видимості користувачів (тільки читання)

### 5. Звичайний користувач

```
Email:    alexander.vasilenko@gmail.com
Password: 03041965
Role:     user
Access:   posts:create|read, comments:create|read, profiles:create|read
```

**Використовуйте для:**

- Тестування робочих процесів звичайного користувача
- Створення власного контенту (пости, коментарі)
- Управління профілем
- Базових операцій користувача

### Гостьовий доступ (неавтентифікований)

Без необхідності входу:

```
Role:     guest (неявна)
Access:   posts:read, comments:read, profiles:read
```

**Використовуйте для:**

- Перегляду публічного контенту
- Тестування неавтентифікованого доступу
- Операцій лише для читання

## Приклади швидкого входу

```bash
# Вхід як системний адміністратор (admin з повним доступом)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq

# Вхід як адміністратор (admin)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@promenade.com","password":"passw0rd"}' | jq

# Вхід як модератор (moderator)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@promenade.com","password":"passw0rd"}' | jq

# Вхід як звичайний користувач (user)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alexander.vasilenko@gmail.com","password":"03041965"}' | jq

# Збереження токена для повторного використання (системний адміністратор)
export TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq -r '.data.access_token')

# Використання токена в запитах
curl -X GET http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Тестування дозволів RBAC

```bash
# Тестування доступу системного адміністратора (має працювати - повний доступ)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $SYSTEM_TOKEN" | jq

# Тестування доступу адміністратора (має працювати - має users:*)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Тестування доступу модератора (має не вдатися - немає дозволу users:list)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $MODERATOR_TOKEN" | jq

# Тестування доступу користувача (має не вдатися - немає дозволів адміністратора)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $USER_TOKEN" | jq
```

## Доступ до бази даних

```bash
# Підключення до бази даних розробки
psql -h localhost -p 5432 -U system -d promenade_dev
# Password: passw0rd

# Перегляд усіх користувачів з їхніми ролями
SELECT
    u.email,
    u.name,
    r.name as role,
    r.description
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
ORDER BY u.email;

# Перегляд дозволів ролей
SELECT
    r.name as role,
    p.resource,
    p.action,
    p.description
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.resource, p.action;
```

## [!] Попередження безпеки

**Ці облікові дані призначені ТІЛЬКИ ДЛЯ РОЗРОБКИ!**

Перед розгортанням у виробництво:

1. Змініть усі паролі за замовчуванням
2. Видаліть або вимкніть облікові записи адміністратора за замовчуванням
3. Використовуйте облікові дані, специфічні для середовища
4. Увімкніть належне управління секретами
5. Налаштуйте належних провайдерів автентифікації

## Потрібна допомога?

- [Посібник з авторизації](docs/AUTHORIZATION.md) - Документація системи RBAC
- [Посібник з тестування](docs/TESTING_GUIDE.md) - Як тестувати з автентифікацією
- [Документація API](README.md#api-examples) - Повний довідник API
