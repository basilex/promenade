# Arquitetura Promenade - Referência Rápida

[🇬🇧 English](../ARCHITECTURE_QUICKREF.md) | [🇺🇦 Українська](../uk/ARCHITECTURE_QUICKREF.uk.md) | [🇩🇪 Deutsch](../de/ARCHITECTURE_QUICKREF.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/ARCHITECTURE_QUICKREF.es.md)

## Core vs Módulos: Regra Simples

**CORE** = Infraestrutura + Dados de Referência + Auth

- Sempre habilitado, fornece serviços

**MÓDULOS** = Lógica de Negócios

- Opcionais, licenciáveis, independentes

---

## O Que Pertence ao Core?

### PERTENCE ao Core:

1. **Serviços de Infraestrutura**

   - Gerenciamento de conexão com banco de dados
   - Event IBus (adaptadores Memory/Redis)
   - Agendador (tarefas cron)
   - Carregador de configuração
   - Logger
   - Serviço de e-mail
   - Gerenciador JWT

2. **Base de Segurança**

   - Autenticação de usuários (login, registro, senha)
   - RBAC (papéis, permissões, controle de acesso)
   - Sessões (tokens JWT)

3. **Dados de Referência**

   - Países (145 países, códigos ISO 3166-1, regiões)
   - Moedas (124 moedas, ISO 4217, símbolos)
   - Regiões (30 regiões administrativas: estados, oblasts, províncias, Länder)
   - Cidades (17 grandes cidades com coordenadas, população, capitais)
   - Métodos de Pagamento (40+ métodos: cartões, carteiras, cripto, BNPL)
   - Fusos horários (banco de dados de fusos horários IANA)
   - Idiomas (códigos ISO 639)
   - _Dados estáveis e raramente alterados, compartilhados entre módulos_

4. **Interfaces de Gerenciamento**
   - Registro de módulos
   - Registro de handlers de purga
   - Registro de políticas de purga
   - Interface do event bus

### NÃO Pertence ao Core:

- Entidades de negócios (Post, Comment, Profile, etc.)
- Casos de uso de negócios
- Handlers HTTP de negócios
- Rotas de negócios
- Configuração de negócios
- Lógica específica de entidades

**Regra geral:** Se é um conceito de negócio que poderia ser vendido separadamente, é um MÓDULO.

---

## O Que Pertence aos Módulos?

### Estrutura do Módulo

```
internal/modules/mymodule/
├── module.go              # Implementação do módulo
├── register.go            # Auto-registro via init()
├── config/                # Configurações YAML próprias por ambiente
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
├── entity/                # Entidades de domínio
├── usecase/               # Lógica de negócios
└── adapter/
    ├── http/              # Handlers, DTOs, rotas
    ├── repository/        # Implementações Postgres
    └── purge/             # Handlers de purga (se necessário)
```

### Checklist do Módulo

- [ ] Tem estrutura própria entity/usecase/adapter
- [ ] Carrega configuração própria de `config/config.*.yaml`
- [ ] Registra rotas em `RegisterRoutes()`
- [ ] Registra permissões em `RegisterPermissions()`
- [ ] Registra handlers de purga (se entidades com soft-delete)
- [ ] Sem imports de `internal/domain` ou `internal/usecase`
- [ ] Usa apenas pacotes `pkg/*`

---

## Fluxo de Trabalho de Desenvolvimento de Módulo

### 1. Criar Módulo

```bash
mkdir -p internal/modules/mymodule/{config,entity,usecase,adapter/http/handler}
```

### 2. Implementar Interface do Módulo

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type MyModule struct {
    *module.BaseModule
    db *sqlx.DB
    // ... outros campos
}

func New() module.IModule {
    return &MyModule{
        BaseModule: module.NewBaseModule(module.Metadata{
            Name:        "mymodule",
            DisplayName: "My IModule",
            Version:     "1.0.0",
            Description: "Does something useful",
        }),
    }
}

func (m *MyModule) Initialize(ctx context.Context, core *module.Core) error {
    // 1. Carregar configuração do módulo
    cfg := moduleconfig.Load("internal/modules/mymodule/config", os.Getenv("ENVIRONMENT"))

    // 2. Configurar repositórios, casos de uso, handlers
    m.db = core.DB

    // 3. Registrar handlers de purga (se necessário)
    // 4. Registrar políticas de retenção (se necessário)

    return nil
}

func (m *MyModule) RegisterRoutes(router *gin.RouterGroup) {
    group := router.Group("/mymodule")
    {
        group.GET("", m.handler.List)
        group.POST("", m.handler.Create)
    }
}

func (m *MyModule) RegisterPermissions() []module.Permission {
    return []module.Permission{
        {Resource: "mymodule", Action: "read", Description: "View items"},
        {Resource: "mymodule", Action: "create", Description: "Create items"},
    }
}

// ... implementar outros métodos da interface
```

### 3. Auto-Registro

```go
// internal/modules/mymodule/register.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

func init() {
    module.DefaultRegistry.Register(New())
}
```

### 4. Adicionar Configuração

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true
  version: "1.0.0"

mymodule:
  max_items: 100
  allow_public: true

purge:
  my_entities:
    retention_days: 60
    enabled: true
```

### 5. Habilitar na Configuração Principal

```yaml
# config/modules.yaml
modules:
  enabled:
    - posts
    - profiles
    - mymodule # Adicionar aqui
```

### 6. Importar em main.go

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/mymodule"  // Adicionar aqui
)
```

---

## Regras de Configuração

### Configuração Core

**Arquivo:** `config/app.{dev|test|prod}.yaml`

**Contém APENAS:**

- Configurações de infraestrutura (DB, servidor, JWT, logging)
- Configuração do event bus
- Infraestrutura de purga (enabled, schedule, batch_size)
- Configurações CORS
- Configurações do serviço de e-mail

**NÃO contém:**

- Dias de retenção específicos de entidades → Módulos
- Configurações específicas de módulos → Módulos
- Configuração de lógica de negócios → Módulos

### Configuração do Módulo

**Arquivo:** `internal/modules/{name}/config/config.{dev|test|prod}.yaml`

**Contém:**

- Metadados do módulo (nome, versão)
- Configurações específicas do módulo
- Políticas de retenção de purga (se aplicável)
- Feature flags (se aplicável)

**Carregado por:** Cada módulo via `pkg/module/config.Load()`

---

## Regras do Sistema de Purga

### FORMA ANTIGA (Core conhece entidades)

```go
//  ERRADO - Core tem configuração específica de entidades
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

### NOVA FORMA (Core apenas orquestra)

**Configuração Core:**

```yaml
purge:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 1000
```

**Configuração Módulo:**

```yaml
purge:
  user_posts:
    retention_days: 90
    enabled: true
```

**Registro do Módulo:**

```go
// Módulo registra handler
handler := purge.NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)

// Módulo registra política
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

**Orquestração Core:**

```go
// Core obtém TODAS as políticas do registro
policies := purge.DefaultPolicyRegistry.GetAllPolicies()

// Core cria caso de uso
useCase := usecase.NewPurgeUseCase(
    purge.DefaultRegistry,  // handlers
    policies,               // dos módulos
    batchSize,
    eventBus,
)

// Core inicia agendador
scheduler.Start(ctx)
```

**Resultado:** Core não sabe nada sobre `user_posts` ou dias de retenção!

---

## Padrões de Comunicação

### Orientado a Eventos (Preferido)

```go
// Módulo A publica
event := &PostCreatedEvent{PostID: id}
eventBus.Publish(ctx, "post.created", event)

// Módulo B assina
eventBus.Subscribe("post.created", func(e bus.Event) {
    // Processar assincronamente
})
```

**Benefícios:** Acoplamento fraco, processamento assíncrono

### Registro Direto (Usar Com Moderação)

```go
// Obter outro módulo
postsModule := core.Registry.Get("posts")

// Verificar tipo e chamar
if api, ok := postsModule.(PostsAPI); ok {
    stats := api.GetStats(userID)
}
```

**Usar apenas quando:** Resposta síncrona necessária, não pode usar eventos

---

## Erros Comuns a Evitar

### Importar Pacotes Core em Módulos

```go
//  ERRADO
import "github.com/basilex/promenade/internal/domain/entity"
import "github.com/basilex/promenade/internal/usecase"
```

**Correção:** Definir entidades no próprio pacote `entity/` do módulo.

### Codificar Valores de Negócio no Código

```go
//  ERRADO
const maxCommentLength = 2000
```

**Correção:** Carregar da configuração do módulo.

### Colocar Lógica de Negócios no Core

```go
//  ERRADO - PostUseCase em internal/usecase/
```

**Correção:** Mover para o pacote `usecase/` do módulo.

### Core Conhecendo Entidades de Módulos

```go
//  ERRADO - Core tem dias de retenção para posts
type PurgeConfig struct {
    RetentionDaysUserPosts int
}
```

**Correção:** Módulo registra política de retenção via registro.

---

## Estratégia de Testes

### Testes Core

- Testes unitários de serviços de infraestrutura
- Testes de integração auth/RBAC
- Testes de repositórios de dados de referência

### Testes de Módulos

- Testes unitários de lógica de negócios (casos de uso)
- Testes de integração de repositórios
- Testes de handlers com casos de uso mock

### Testes Smoke

- Fluxos críticos ponta a ponta
- Testes de comunicação entre módulos via eventos
- Verificar funcionamento de rotas de módulos

---

## Comandos Rápidos

```bash
# Build
make build

# Executar dev
make dev

# Executar testes
make test

# Executar testes de módulo específico
go test ./internal/modules/posts/...

# Executar testes smoke
make test-smoke

# Gerar boilerplate de módulo
make generate ENTITY=MyEntity

# Criar migração
make migrate-create NAME=add_my_table
```

---

## Árvore de Decisão: Core ou Módulo?

```
É infraestrutura (DB, logger, event bus)?
└─> SIM → CORE

É segurança (auth, RBAC)?
└─> SIM → CORE

São dados de referência (países, moedas, regiões, cidades, métodos de pagamento)?
└─> SIM → CORE

É estável e usado por múltiplos módulos?
└─> SIM → Considerar CORE (ou pkg compartilhado)

É lógica de negócios?
└─> SIM → MÓDULO

Pode ser vendido separadamente?
└─> SIM → MÓDULO

É específico de entidade?
└─> SIM → MÓDULO

Em dúvida?
└─> MÓDULO (mais fácil mover para core depois do que vice-versa)
```

---

## Recursos

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.pt.md) - Revisão detalhada de arquitetura
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md) - Arquitetura visual
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.pt.md) - Guia de desenvolvimento de módulos
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.pt.md) - Princípios de independência
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.pt.md) - Detalhes do sistema de purga
- [../../internal/CORE.md](../../internal/CORE.md) - Documentação de componentes Core

---

**Lembre-se:**

- Core = Infraestrutura + Dados de Referência + Auth
- Módulos = Lógica de Negócios (independentes, licenciáveis)
- Use registros para acoplamento fraco
- Eventos para comunicação assíncrona
- Autonomia de configuração para cada módulo
