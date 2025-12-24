# Посібник з Розробки Модулів

[🇬🇧 English](../MODULE_DEVELOPMENT.md) | 🇺🇦 **Українська** | [🇩🇪 Deutsch](../de/MODULE_DEVELOPMENT.de.md) | [🇵🇹 Português](../pt/MODULE_DEVELOPMENT.pt.md) | [🇪🇸 Español](../es/MODULE_DEVELOPMENT.es.md)

Цей посібник пояснює, як розробляти власні модулі для Promenade, використовуючи Плагінну Архітектуру.

## Зміст

- [Огляд](#огляд)
- [Структура Модуля](#структура-модуля)
- [Створення Модуля](#створення-модуля)
- [Життєвий Цикл Модуля](#життєвий-цикл-модуля)
- [Міжмодульна Комунікація](#міжмодульна-комунікація)
- [Найкращі Практики](#найкращі-практики)
- [Приклади](#приклади)

---

## Огляд

Promenade використовує **Плагінну Архітектуру**, яка дозволяє:

- Упаковувати бізнес-логіку як незалежні, багаторазові модулі
- Спільно використовувати основну інфраструктуру (БД, Event Bus, Auth, RBAC)
- Увімкнути/вимкнути модулі через конфігурацію
- Продавати модулі комерційно з ліцензуванням
- Комбінувати кілька модулів (наприклад, Warehouse + Fleet)

### Core vs Модулі

**Core** (завжди увімкнений):

- Автентифікація і JWT
- RBAC і Дозволи
- Event Bus і Сповіщення
- Довідкові Дані (Країни, Валюти)
- Журнал Аудиту

**Модулі** (опційні):

- Posts, Comments, Profiles (соціальні функції)
- Warehouse Management (управління інвентаризацією)
- Fleet Management (управління транспортом)
- Finance (виставлення рахунків, інвойси)
- Власна бізнес-логіка

---

## Структура Модуля

Типовий модуль дотримується Clean Architecture:

```
modules/warehouse/
├── domain/
│   ├── entity/
│   │   └── item.go
│   └── repository/
│       └── item_repository.go
│
├── usecase/
│   └── item_usecase.go
│
├── adapter/
│   ├── http/
│   │   ├── handler/
│   │   │   └── item_handler.go
│   │   └── dto/
│   │       └── item_dto.go
│   └── repository/
│       └── postgres/
│           └── item_repository.go
│
├── migrations/
│   ├── 001_create_warehouse_items.up.sql
│   └── 001_create_warehouse_items.down.sql
│
└── module.go  # Реєстрація модуля
```

---

## Створення Модуля

### Крок 1: Визначення Метаданих Модуля

```go
// modules/warehouse/module.go
package warehouse

import (
	"github.com/basilex/promenade/pkg/module"
)

type WarehouseModule struct {
	*module.BaseModule

	// Залежності
	itemRepo repository.ItemRepository
	itemUC   usecase.ItemUseCase
}

func New() module.Module {
	meta := module.Metadata{
		Name:        "warehouse",
		DisplayName: "Warehouse Management",
		Version:     "1.2.0",
		Author:      "Your Company",
		Description: "Inventory and stock management system",
		License:     "Commercial",
		Tags:        []string{"inventory", "logistics"},
	}

	return &WarehouseModule{
		BaseModule: module.NewBaseModule(meta),
	}
}
```

### Крок 2: Реалізація Інтерфейсу Модуля

```go
// Dependencies (опційно - повернути порожній масив якщо немає)
func (m *WarehouseModule) Dependencies() []string {
	return []string{} // Немає залежностей
}

// Initialize - налаштування репозиторіїв та сценаріїв використання
func (m *WarehouseModule) Initialize(ctx context.Context, core *module.Core) error {
	// Викликати базову реалізацію
	if err := m.BaseModule.Initialize(ctx, core); err != nil {
		return err
	}

	// Ініціалізувати репозиторії
	m.itemRepo = postgres.NewItemRepository(core.DB)

	// Ініціалізувати сценарії використання
	m.itemUC = usecase.NewItemUseCase(m.itemRepo, core.EventBus)

	return nil
}

// RegisterRoutes - реєстрація HTTP endpoints
func (m *WarehouseModule) RegisterRoutes(router *gin.RouterGroup) {
	handler := handler.NewItemHandler(m.itemUC)

	items := router.Group("/warehouse/items")
	{
		items.GET("", handler.List)      // GET /api/v1/warehouse/items
		items.POST("", handler.Create)   // POST /api/v1/warehouse/items
		items.GET("/:id", handler.Get)   // GET /api/v1/warehouse/items/:id
		items.PUT("/:id", handler.Update)
		items.DELETE("/:id", handler.Delete)
	}
}

// RegisterMigrations - повернути міграції бази даних
func (m *WarehouseModule) RegisterMigrations() []module.Migration {
	return []module.Migration{
		{
			Version:     1,
			Description: "Create warehouse_items table",
			Up: `
				CREATE TABLE warehouse_items (
					id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
					name VARCHAR(255) NOT NULL,
					sku VARCHAR(100) UNIQUE NOT NULL,
					quantity INTEGER NOT NULL DEFAULT 0,
					price DECIMAL(10,2) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT NOW(),
					updated_at TIMESTAMP NOT NULL DEFAULT NOW()
				);
			`,
			Down: `DROP TABLE warehouse_items;`,
		},
	}
}

// RegisterEventHandlers - підписатися на події
func (m *WarehouseModule) RegisterEventHandlers(bus bus.Bus) error {
	// Підписатися на події замовлень
	return bus.Subscribe(ctx, "order.created", m.handleOrderCreated)
}

func (m *WarehouseModule) handleOrderCreated(ctx context.Context, event bus.Event) error {
	// Зменшити кількість на складі
	// ...
	return nil
}

// RegisterPermissions - визначити RBAC дозволи
func (m *WarehouseModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "warehouse:items", Action: "read", Description: "View warehouse items"},
		{Resource: "warehouse:items", Action: "create", Description: "Create warehouse items"},
		{Resource: "warehouse:items", Action: "update", Description: "Update warehouse items"},
		{Resource: "warehouse:items", Action: "delete", Description: "Delete warehouse items"},
	}
}

// Start - запустити фонові робітники (опційно)
func (m *WarehouseModule) Start(ctx context.Context) error {
	// Запустити робітника синхронізації інвентаризації
	go m.syncInventoryWorker(ctx)
	return nil
}

// Stop - граціозне завершення (опційно)
func (m *WarehouseModule) Stop(ctx context.Context) error {
	// Зупинити робітників, закрити з'єднання тощо
	return nil
}

// HealthCheck - перевірка здоров'я модуля (опційно)
func (m *WarehouseModule) HealthCheck(ctx context.Context) error {
	// Перевірити підключення до бази даних, зовнішні API тощо
	return nil
}
```

### Крок 3: Реєстрація Модуля

```go
// modules/warehouse/register.go
package warehouse

import "github.com/basilex/promenade/pkg/module"

func init() {
	// Автоматична реєстрація модуля при імпорті
	module.DefaultRegistry.Register(New())
}
```

### Крок 4: Увімкнення Модуля

```yaml
# config/modules.yaml
modules:
  enabled:
    - warehouse

  config:
    warehouse:
      version: "1.2.0"
      license_key: "WH-ABC-123-XYZ" # Опційно
      settings:
        max_items: 10000
```

---

## Життєвий Цикл Модуля

**Послідовність Запуску:**

1. **Запуск Додатку**

   - Прочитати `config/modules.yaml`
   - Завантажити увімкнені модулі
   - Завантажити налаштування для кожного модуля

2. **Ініціалізація Core**

   - З'єднання з базою даних
   - Event bus
   - JWT менеджер
   - Логер

3. **Module.Initialize()** (у порядку залежностей)

   - Вирішити залежності
   - Ініціалізувати репозиторії
   - Ініціалізувати сценарії використання

4. **Module.RegisterRoutes()**

   - Зареєструвати HTTP endpoints

5. **Module.RegisterMigrations()**

   - Застосувати міграції бази даних

6. **Module.RegisterEventHandlers()**

   - Підписатися на події

7. **Module.RegisterPermissions()**

   - Вставити RBAC дозволи

8. **Module.Start()**

   - Запустити фонові робітники
   - Ініціалізувати зовнішні з'єднання

9. **Сервер Працює**

**Послідовність Завершення:**

10. **Module.Stop()** (у зворотному порядку)
    - Зупинити робітників
    - Закрити з'єднання
    - Очистити ресурси

---

## Міжмодульна Комунікація

Модулі повинні спілкуватися через **події** (слабке зв'язування):

### Публікація Подій

```go
// В модулі Warehouse
func (uc *ItemUseCase) CreateItem(ctx context.Context, item *entity.Item) error {
	// Зберегти в базу даних
	if err := uc.repo.Create(ctx, item); err != nil {
		return err
	}

	// Опублікувати подію
	event := &events.ItemCreatedEvent{
		ItemID:   item.ID,
		SKU:      item.SKU,
		Quantity: item.Quantity,
	}
	uc.eventBus.Publish(ctx, "warehouse.item.created", event)

	return nil
}
```

### Підписка на Події

```go
// В модулі Fleet (потребує товари зі складу для запчастин)
func (m *FleetModule) RegisterEventHandlers(bus bus.Bus) error {
	return bus.Subscribe(ctx, "warehouse.item.created", m.handleItemCreated)
}

func (m *FleetModule) handleItemCreated(ctx context.Context, event bus.Event) error {
	itemEvent := event.(*events.ItemCreatedEvent)

	// Оновити інвентаризацію запчастин
	// ...

	return nil
}
```

---

## Найкращі Практики

### РОБИТИ

1. **Використовувати BaseModule** - Вбудувати `module.BaseModule` щоб уникнути реалізації кожного методу
2. **Clean Architecture** - Дотримуватися структури domain/usecase/adapter
3. **UUID v7** - Використовувати `uuidv7.New()` для первинних ключів
4. **Подієво-Орієнтований** - Спілкуватися між модулями через події
5. **RBAC** - Реєструвати дозволи для всіх endpoints
6. **Міграції** - Завжди надавати Up і Down міграції
7. **Перевірки Здоров'я** - Реалізувати `HealthCheck()` для моніторингу
8. **Граціозне Завершення** - Очищати ресурси в `Stop()`

### НЕ РОБИТИ

1. **Прямі Залежності** - Ніколи не імпортувати код інших модулів безпосередньо
2. **Спільні Таблиці** - Кожен модуль володіє своїми таблицями
3. **Модифікації Core** - Не змінювати код core для функцій модуля
4. **Синхронні Виклики** - Уникати блокуючої міжмодульної комунікації
5. **Глобальний Стан** - Не використовувати глобальні змінні (використовувати Core натомість)

---

## Приклади

### Приклад 1: Простий Модуль (Без Залежностей)

```go
package hello

type HelloModule struct {
	*module.BaseModule
}

func New() module.Module {
	return &HelloModule{
		BaseModule: module.NewBaseModule(module.Metadata{
			Name:    "hello",
			Version: "1.0.0",
		}),
	}
}

func (m *HelloModule) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello from module!"})
	})
}
```

### Приклад 2: Модуль з Залежностями

```go
package fleet

type FleetModule struct {
	*module.BaseModule
}

func (m *FleetModule) Dependencies() []string {
	return []string{"warehouse"} // Fleet залежить від Warehouse
}

func (m *FleetModule) Initialize(ctx context.Context, core *module.Core) error {
	// Warehouse має бути ініціалізований першим (гарантовано реєстром)
	warehouseModule := core.Registry.Get("warehouse")
	if warehouseModule == nil {
		return fmt.Errorf("warehouse module required but not loaded")
	}

	// Ініціалізувати логіку, специфічну для fleet
	// ...

	return nil
}
```

---

## Тестування Модулів

```go
// modules/warehouse/module_test.go
func TestWarehouseModule(t *testing.T) {
	// Налаштувати тестовий core
	core := setupTestCore(t)

	// Створити модуль
	mod := warehouse.New()

	// Ініціалізувати
	err := mod.Initialize(context.Background(), core)
	assert.NoError(t, err)

	// Тестувати маршрути
	router := gin.New()
	group := router.Group("/api/v1")
	mod.RegisterRoutes(group)

	// Тестувати endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/warehouse/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
```

---

## Наступні Кроки

1. Переглянути існуючі модулі в `internal/modules/`
2. Використати `make generate-module NAME=mymodule` (TODO)
3. Перевірити `pkg/module/` для повного API довідника
4. Прочитати [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md) для шаблонів Clean Architecture

---

Для питань, див. [README.md](../../README.md) або відкрийте issue.
