[🇬🇧 English](../LOGGING.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/LOGGING.de.md) | [🇵🇹 Português](../pt/LOGGING.pt.md) | [🇪🇸 Español](../es/LOGGING.es.md)

# Посібник зі структурованого логування

## Огляд

Promenade використовує **slog** (structured logging з Go 1.21+) для високопродуктивного, машиночитаного логування.

## Ключові переваги

[+] **Структуроване логування** - JSON формат для автоматизації
[+] **Контекстні поля** - request_id, user_id, trace_id
[+] **Рівні логування** - Debug, Info, Warn, Error
[+] **Продуктивність** - Нульова алокація де можливо
[+] **stdlib** - Без зовнішніх залежностей

## Конфігурація

### На основі середовища

```go
// Development: text format with source file/line
logger.Init(logger.Config{
    Level:      "debug",
    Format:     "text",
    AddSource:  true,
    TimeFormat: time.RFC3339,
})

// Production: JSON format without source
logger.Init(logger.Config{
    Level:      "info",
    Format:     "json",
    AddSource:  false,
    TimeFormat: time.RFC3339,
})
```

### Рівні логування

- `debug` - Детальна інформація для відлагодження
- `info` - Загальна інформація (за замовчуванням)
- `warn` - Попередження
- `error` - Помилки

## Базове використання

### Просте логування

```go
import (
    "log/slog"
    "github.com/basilex/promenade/pkg/logger"
)

// Simple messages
logger.Info("Server started")
logger.Error("Failed to connect")

// With attributes
logger.Info("User registered",
    slog.String("email", "user@example.com"),
    slog.String("user_id", userID),
)

logger.Error("Database query failed",
    slog.String("query", "SELECT * FROM users"),
    slog.Duration("duration", queryTime),
    slog.Any("error", err),
)
```

### Логування з контекстом

```go
import "context"

// Context logging (automatically adds request_id, user_id, trace_id)
logger.InfoContext(ctx, "Processing request",
    slog.String("action", "create_user"),
)

logger.ErrorContext(ctx, "Failed to create user",
    slog.Any("error", err),
    slog.String("email", req.Email),
)
```

### Logger з полями

```go
// Create logger with preset fields
userLogger := logger.Default().WithFields(map[string]any{
    "user_id": userID,
    "tenant_id": tenantID,
})

// All subsequent logs will include these fields
userLogger.Info("User action performed")
userLogger.Warn("Rate limit approaching")
```

### Logger з помилкою

```go
logger.Default().WithError(err).Error("Operation failed",
    slog.String("operation", "update_profile"),
)
```

## Приклади за шарами

### HTTP Handler

```go
func (h *UserHandler) Create(c *gin.Context) {
    requestID := c.GetString("request_id")

    // Log request start
    logger.Info("Creating user",
        slog.String("request_id", requestID),
        slog.String("ip", c.ClientIP()),
    )

    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        logger.Warn("Invalid request body",
            slog.String("request_id", requestID),
            slog.Any("error", err),
        )
        response.Error(c, http.StatusBadRequest, "invalid request", err)
        return
    }

    user, err := h.userUseCase.Create(c.Request.Context(), &req)
    if err != nil {
        logger.Error("Failed to create user",
            slog.String("request_id", requestID),
            slog.String("email", req.Email),
            slog.Any("error", err),
        )
        response.Error(c, http.StatusInternalServerError, "failed to create user", err)
        return
    }

    logger.Info("User created successfully",
        slog.String("request_id", requestID),
        slog.String("user_id", user.ID.String()),
        slog.String("email", user.Email),
    )

    response.Success(c, http.StatusCreated, user)
}
```

### Use Case

```go
func (uc *UserUseCase) Create(ctx context.Context, req *dto.CreateUserRequest) (*entity.User, error) {
    logger.DebugContext(ctx, "UseCase: Creating user",
        slog.String("email", req.Email),
    )

    // Check if user exists
    existing, err := uc.userRepo.FindByEmail(ctx, req.Email)
    if err != nil && err != entity.ErrNotFound {
        logger.ErrorContext(ctx, "Failed to check existing user",
            slog.String("email", req.Email),
            slog.Any("error", err),
        )
        return nil, err
    }

    if existing != nil {
        logger.WarnContext(ctx, "User already exists",
            slog.String("email", req.Email),
        )
        return nil, ErrUserAlreadyExists
    }

    // Create user
    user := &entity.User{
        Email: req.Email,
        Name:  req.Name,
    }

    if err := uc.userRepo.Create(ctx, user); err != nil {
        logger.ErrorContext(ctx, "Failed to create user in repository",
            slog.String("email", req.Email),
            slog.Any("error", err),
        )
        return nil, err
    }

    logger.InfoContext(ctx, "User created in use case",
        slog.String("user_id", user.ID.String()),
    )

    return user, nil
}
```

### Repository

```go
func (r *IUserRepository) Create(ctx context.Context, user *entity.User) error {
    query := `
        INSERT INTO users (email, name, status)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at
    `

    logger.DebugContext(ctx, "Executing user insert query",
        slog.String("email", user.Email),
    )

    err := r.db.QueryRowContext(ctx, query,
        user.Email,
        user.Name,
        user.Status,
    ).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

    if err != nil {
        logger.ErrorContext(ctx, "Database query failed",
            slog.String("query", "INSERT users"),
            slog.String("email", user.Email),
            slog.Any("error", err),
        )
        return err
    }

    logger.DebugContext(ctx, "User inserted successfully",
        slog.String("user_id", user.ID.String()),
    )

    return nil
}
```

### З'єднання з базою даних

```go
func NewPostgresConnection(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
    logger.Info("Connecting to database",
        slog.String("host", cfg.Host),
        slog.Int("port", cfg.Port),
        slog.String("database", cfg.DBName),
    )

    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        logger.Error("Failed to connect to database",
            slog.String("host", cfg.Host),
            slog.Any("error", err),
        )
        return nil, err
    }

    logger.Info("Database connection established",
        slog.Int("max_open_conns", cfg.MaxOpenConns),
        slog.Int("max_idle_conns", cfg.MaxIdleConns),
    )

    return db, nil
}
```

## Формати виводу

### Development

```
time=2025-12-16T16:35:58+02:00 level=INFO source=/path/file.go:191 msg="HTTP request" method=GET path=/api/users status=200 duration=15ms ip=192.168.1.100 request_id=abc-123
```

### Production (JSON)

```json
{
  "time": "2025-12-16T16:35:58+02:00",
  "level": "INFO",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/api/users",
  "status": 200,
  "duration": 15000000,
  "ip": "192.168.1.100",
  "request_id": "abc-123"
}
```

## Middleware

### Logger Middleware

Автоматично логує всі HTTP-запити зі структурованими полями:

```go
time=2025-12-16T16:35:58+02:00 level=INFO msg="HTTP request"
    method=GET
    path=/api/countries
    query="region=asia&page_size=5"
    status=200
    duration=17.994375ms
    ip=::1
    user_agent=curl/8.7.1
    request_id=b7508a75-04f0-4fef-89c7-ac1b8804d379
    response_size=734
```

### Recovery Middleware

Логує паніки з повним стеком:

```go
logger.Error("PANIC recovered",
    slog.Any("panic", err),
    slog.String("stack", string(debug.Stack())),
    slog.String("request_id", requestID),
    slog.String("method", c.Request.Method),
    slog.String("path", c.Request.URL.Path),
)
```

## Найкращі практики

### [+] РОБІТЬ

```go
// Use structured fields
logger.Info("User logged in",
    slog.String("user_id", userID),
    slog.String("ip", ip),
)

// Add context to errors
logger.ErrorContext(ctx, "Payment failed",
    slog.String("payment_id", paymentID),
    slog.String("amount", amount),
    slog.Any("error", err),
)

// Use proper types
slog.Int("count", 42)
slog.Duration("latency", time.Since(start))
slog.Bool("is_admin", user.IsAdmin)
```

### [X] НЕ РОБІТЬ

```go
// DON'T use string interpolation
logger.Info(fmt.Sprintf("User %s logged in", userID)) // [X]

// DON'T log sensitive data
logger.Info("User authenticated",
    slog.String("password", password), // [X] НІКОЛИ!
)

// DON'T log in loops unnecessarily
for _, item := range items {
    logger.Debug("Processing item") // [X] Може створити мільйони логів
}
```

## Інтеграція з системами моніторингу

JSON формат легко парситься системами:

- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **DataDog**
- **New Relic**
- **Splunk**

Приклад Kibana запиту:

```
request_id:"abc-123" AND status:>=400
```

## Продуктивність

slog оптимізований для продуктивності:

- Нульова алокація для базових типів
- Ліниве обчислення атрибутів
- Ефективна JSON серіалізація

Benchmark:

```
BenchmarkSlog-8     1000000    1156 ns/op    0 allocs/op
```

## Міграція зі старого логування

Старий код:

```go
log.Printf("User %s registered", email)
log.Fatalf("Failed to start: %v", err)
```

Новий код:

```go
logger.Info("User registered", slog.String("email", email))
logger.Fatal("Failed to start", slog.Any("error", err))
```
