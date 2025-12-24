# Guia de Desenvolvimento de Módulos

[🇬🇧 English](../MODULE_DEVELOPMENT.md) | [🇺🇦 Українська](../uk/MODULE_DEVELOPMENT.uk.md) | [🇩🇪 Deutsch](../de/MODULE_DEVELOPMENT.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/MODULE_DEVELOPMENT.es.md)

Este guia explica como desenvolver módulos personalizados para Promenade usando a Arquitetura de Plugin.

## Índice

- [Visão Geral](#visão-geral)
- [Estrutura do Módulo](#estrutura-do-módulo)
- [Criando um Módulo](#criando-um-módulo)
- [Ciclo de Vida do Módulo](#ciclo-de-vida-do-módulo)
- [Comunicação Inter-Módulos](#comunicação-inter-módulos)
- [Melhores Práticas](#melhores-práticas)
- [Exemplos](#exemplos)

---

## Visão Geral

Promenade usa uma **Arquitetura de Plugin** que permite:

- Empacotar lógica de negócios como módulos independentes e reutilizáveis
- Compartilhar infraestrutura central (DB, Event Bus, Auth, RBAC)
- Habilitar/desabilitar módulos via configuração
- Vender módulos comercialmente com licenciamento
- Combinar múltiplos módulos (ex: Warehouse + Fleet)

### Core vs Módulos

**Core** (sempre habilitado):

- Autenticação & JWT
- RBAC & Permissões
- Event Bus & Notificações
- Dados de Referência (Países, Moedas)
- Registro de Auditoria

**Módulos** (opcionais):

- Posts, Comments, Profiles (recursos sociais)
- Warehouse Management (gerenciamento de inventário)
- Fleet Management (gerenciamento de veículos)
- Finance (faturamento, faturas)
- Lógica de negócios personalizada

---

## Estrutura do Módulo

Um módulo típico segue Clean Architecture:

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
└── module.go  # Registro do módulo
```

---

## Criando um Módulo

### Passo 1: Definir Metadados do Módulo

```go
// modules/warehouse/module.go
package warehouse

import (
	"github.com/basilex/promenade/pkg/module"
)

type WarehouseModule struct {
	*module.BaseModule

	// Dependências
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

### Passo 2: Implementar Interface do Módulo

```go
// Dependencies (opcional - retornar array vazio se não houver)
func (m *WarehouseModule) Dependencies() []string {
	return []string{} // Sem dependências
}

// Initialize - configurar repositórios e casos de uso
func (m *WarehouseModule) Initialize(ctx context.Context, core *module.Core) error {
	// Chamar implementação base
	if err := m.BaseModule.Initialize(ctx, core); err != nil {
		return err
	}

	// Inicializar repositórios
	m.itemRepo = postgres.NewItemRepository(core.DB)

	// Inicializar casos de uso
	m.itemUC = usecase.NewItemUseCase(m.itemRepo, core.EventBus)

	return nil
}

// RegisterRoutes - registrar endpoints HTTP
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

// RegisterMigrations - retornar migrações de banco de dados
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

// RegisterEventHandlers - assinar eventos
func (m *WarehouseModule) RegisterEventHandlers(bus bus.Bus) error {
	// Assinar eventos de pedidos
	return bus.Subscribe(ctx, "order.created", m.handleOrderCreated)
}

func (m *WarehouseModule) handleOrderCreated(ctx context.Context, event bus.Event) error {
	// Reduzir quantidade em estoque
	// ...
	return nil
}

// RegisterPermissions - definir permissões RBAC
func (m *WarehouseModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "warehouse:items", Action: "read", Description: "View warehouse items"},
		{Resource: "warehouse:items", Action: "create", Description: "Create warehouse items"},
		{Resource: "warehouse:items", Action: "update", Description: "Update warehouse items"},
		{Resource: "warehouse:items", Action: "delete", Description: "Delete warehouse items"},
	}
}

// Start - iniciar workers em background (opcional)
func (m *WarehouseModule) Start(ctx context.Context) error {
	// Iniciar worker de sincronização de inventário
	go m.syncInventoryWorker(ctx)
	return nil
}

// Stop - encerramento gracioso (opcional)
func (m *WarehouseModule) Stop(ctx context.Context) error {
	// Parar workers, fechar conexões, etc.
	return nil
}

// HealthCheck - verificar saúde do módulo (opcional)
func (m *WarehouseModule) HealthCheck(ctx context.Context) error {
	// Verificar conectividade do banco de dados, APIs externas, etc.
	return nil
}
```

### Passo 3: Registrar Módulo

```go
// modules/warehouse/register.go
package warehouse

import "github.com/basilex/promenade/pkg/module"

func init() {
	// Auto-registrar módulo ao importar
	module.DefaultRegistry.Register(New())
}
```

### Passo 4: Habilitar Módulo

```yaml
# config/modules.yaml
modules:
  enabled:
    - warehouse

  config:
    warehouse:
      version: "1.2.0"
      license_key: "WH-ABC-123-XYZ" # Opcional
      settings:
        max_items: 10000
```

---

## Ciclo de Vida do Módulo

**Sequência de Inicialização:**

1. **Inicialização da Aplicação**

   - Ler `config/modules.yaml`
   - Carregar módulos habilitados
   - Carregar configurações por módulo

2. **Inicializar Core**

   - Conexão com banco de dados
   - Event bus
   - Gerenciador JWT
   - Logger

3. **Module.Initialize()** (em ordem de dependência)

   - Resolver dependências
   - Inicializar repositórios
   - Inicializar casos de uso

4. **Module.RegisterRoutes()**

   - Registrar endpoints HTTP

5. **Module.RegisterMigrations()**

   - Aplicar migrações de banco de dados

6. **Module.RegisterEventHandlers()**

   - Assinar eventos

7. **Module.RegisterPermissions()**

   - Inserir permissões RBAC

8. **Module.Start()**

   - Iniciar workers em background
   - Inicializar conexões externas

9. **Servidor Executando**

**Sequência de Encerramento:**

10. **Module.Stop()** (em ordem reversa)
    - Parar workers
    - Fechar conexões
    - Limpar recursos

---

## Comunicação Inter-Módulos

Módulos devem comunicar via **eventos** (acoplamento fraco):

### Publicando Eventos

```go
// No módulo Warehouse
func (uc *ItemUseCase) CreateItem(ctx context.Context, item *entity.Item) error {
	// Salvar no banco de dados
	if err := uc.repo.Create(ctx, item); err != nil {
		return err
	}

	// Publicar evento
	event := &events.ItemCreatedEvent{
		ItemID:   item.ID,
		SKU:      item.SKU,
		Quantity: item.Quantity,
	}
	uc.eventBus.Publish(ctx, "warehouse.item.created", event)

	return nil
}
```

### Assinando Eventos

```go
// No módulo Fleet (precisa de itens do warehouse para peças de reposição)
func (m *FleetModule) RegisterEventHandlers(bus bus.Bus) error {
	return bus.Subscribe(ctx, "warehouse.item.created", m.handleItemCreated)
}

func (m *FleetModule) handleItemCreated(ctx context.Context, event bus.Event) error {
	itemEvent := event.(*events.ItemCreatedEvent)

	// Atualizar inventário de peças de reposição
	// ...

	return nil
}
```

---

## Melhores Práticas

### FAZER

1. **Usar BaseModule** - Incorporar `module.BaseModule` para evitar implementar cada método
2. **Clean Architecture** - Seguir estrutura domain/usecase/adapter
3. **UUID v7** - Usar `uuidv7.New()` para chaves primárias
4. **Orientado a Eventos** - Comunicar entre módulos via eventos
5. **RBAC** - Registrar permissões para todos os endpoints
6. **Migrações** - Sempre fornecer migrações Up e Down
7. **Health Checks** - Implementar `HealthCheck()` para monitoramento
8. **Encerramento Gracioso** - Limpar recursos em `Stop()`

### NÃO FAZER

1. **Dependências Diretas** - Nunca importar código de outros módulos diretamente
2. **Tabelas Compartilhadas** - Cada módulo possui suas tabelas
3. **Modificações no Core** - Não modificar código core para recursos do módulo
4. **Chamadas Síncronas** - Evitar comunicação inter-módulos bloqueante
5. **Estado Global** - Não usar variáveis globais (usar Core em vez disso)

---

## Exemplos

### Exemplo 1: Módulo Simples (Sem Dependências)

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

### Exemplo 2: Módulo com Dependências

```go
package fleet

type FleetModule struct {
	*module.BaseModule
}

func (m *FleetModule) Dependencies() []string {
	return []string{"warehouse"} // Fleet depende de Warehouse
}

func (m *FleetModule) Initialize(ctx context.Context, core *module.Core) error {
	// Warehouse deve ser inicializado primeiro (garantido pelo registry)
	warehouseModule := core.Registry.Get("warehouse")
	if warehouseModule == nil {
		return fmt.Errorf("warehouse module required but not loaded")
	}

	// Inicializar lógica específica do fleet
	// ...

	return nil
}
```

---

## Testando Módulos

```go
// modules/warehouse/module_test.go
func TestWarehouseModule(t *testing.T) {
	// Configurar core de teste
	core := setupTestCore(t)

	// Criar módulo
	mod := warehouse.New()

	// Inicializar
	err := mod.Initialize(context.Background(), core)
	assert.NoError(t, err)

	// Testar rotas
	router := gin.New()
	group := router.Group("/api/v1")
	mod.RegisterRoutes(group)

	// Testar endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/warehouse/items", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
```

---

## Próximos Passos

1. Ver módulos existentes em `internal/modules/`
2. Usar `make generate-module NAME=mymodule` (TODO)
3. Verificar `pkg/module/` para referência completa da API
4. Ler [ARCHITECTURE_OVERVIEW.md](../ARCHITECTURE_OVERVIEW.md) para padrões de Clean Architecture

---

Para perguntas, veja [README.md](../../README.md) ou abra uma issue.
