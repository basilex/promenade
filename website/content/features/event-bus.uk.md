---
title: "Подієво-Орієнтована Архітектура"
description: "Подвійні адаптери event bus для асинхронної комунікації"
weight: 6
---

## Система Event Bus

Promenade включає **event bus з двома адаптерами** для асинхронної, подієво-орієнтованої комунікації.

### Два Адаптери

**Memory Адаптер** (за замовчуванням):

- In-memory Pub/Sub
- Goroutines + channels
- Ідеально для dev/test
- Нуль залежностей

**Redis Адаптер** (продакшн):

- Розподілений Pub/Sub
- Підтримка кількох інстансів
- Постійна черга повідомлень
- Відмовостійкість

### Конфігурація

```yaml
# config/app.yaml
bus:
  adapter: "memory" # або "redis"
  worker_pool_size: 4

  # Redis-специфічні
  redis:
    host: "localhost"
    port: 6379
    password: ""
    db: 0
```

### Публікація Подій

```go
// Визначити подію
type UserRegisteredEvent struct {
    bus.BaseEvent
    UserID string
    Email  string
}

// Опублікувати
event := &UserRegisteredEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    UserID:    user.ID,
    Email:     user.Email,
}
eventBus.Publish(ctx, "user.registered", event)
```

### Підписка на Події

```go
// Підписатися під час ініціалізації модуля
func (m *EmailModule) Initialize(ctx context.Context, core *module.Core) error {
    return core.EventBus.Subscribe(ctx, "user.registered",
        func(ctx context.Context, e bus.Event) error {
            evt := e.(*UserRegisteredEvent)
            return m.sendWelcomeEmail(ctx, evt.Email)
        },
    )
}
```

### Вбудовані Події

- `user.registered` - Реєстрація нового користувача
- `user.deleted` - Видалення облікового запису
- `post.created` - Опублікування нового поста
- `comment.created` - Додавання нового коментаря
- `purge.completed` - Завершення роботи з очищення

### Комунікація Модулів

Модулі використовують події для **слабкого зв'язку**:

```
Posts Module                Email Module
    |                           |
    | user.registered           |
    |-------------------------->|
    |                  Надіслати welcome email
    |                           |
```

### Переваги

✅ **Асинхронно** - Не блокує основний потік  
✅ **Розв'язано** - Модулі не знають один про одного  
✅ **Масштабовано** - Redis для розподілених систем  
✅ **Надійно** - Автоматичні повтори та обробка помилок

[Детальний посібник →](/pkg/bus/README.md)
