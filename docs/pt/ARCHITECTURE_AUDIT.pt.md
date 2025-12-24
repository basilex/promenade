# Auditoria de Arquitetura - Core vs Módulos

[🇬🇧 English](../ARCHITECTURE_AUDIT.md) | [🇺🇦 Українська](../uk/ARCHITECTURE_AUDIT.uk.md) | [🇩🇪 Deutsch](../de/ARCHITECTURE_AUDIT.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/ARCHITECTURE_AUDIT.es.md)

---

**Data:** 22 de dezembro de 2025
**Status:** Conformidade Arquitetural

## Resumo Executivo

Promenade segue uma **arquitetura de plugins** onde:

- **Core** = Infraestrutura + Dados de Referência + Gerenciadores (sempre ativado)
- **Módulos** = Lógica de Negócio (opcional, licenciável, independente)

Esta auditoria confirma que a arquitetura está **corretamente implementada** com separação adequada de responsabilidades.

---

## Responsabilidades do Core

### 1. Gerenciamento de Infraestrutura

```
internal/infrastructure/
├── config/         Carregamento de configuração (YAML + env)
├── database/       Conexão de banco de dados + transações
├── email/          Serviço de email
└── scheduler/      Agendador cron
```

**Status:** Correto - Core fornece infraestrutura como serviço para os módulos.

---

### 2. Dados de Referência

```
internal/domain/entity/
├── country.go      Países (ISO2/3, regiões) - 195+ entradas
├── timezone.go     Fusos horários (IANA) - 500+ entradas
├── language.go     Idiomas (ISO 639) - 180+ entradas
└── (currency via repository)  Moedas (ISO 4217) - 170+ entradas
```

**Propósito:** Dados estáveis e raramente alterados, compartilhados entre módulos.

**Status:** Correto - Estes são verdadeiros dados de referência, não entidades de negócio.

**Implementações de Casos de Uso:**

```
internal/usecase/
├── country_usecase.go   CRUD para países
└── currency_usecase.go  CRUD para moedas
```

---

### 3. Autenticação & Autorização (RBAC)

```
internal/domain/entity/
├── user.go         Entidade de usuário principal (apenas auth: email, senha, papéis)
├── session.go      Sessões JWT
├── role.go         Papéis RBAC (5 papéis do sistema)
└── permission.go   Permissões RBAC (recurso:ação)
```

**Propósito:** Segurança e controle de acesso - fundamental para todos os módulos.

**Status:** Correto - Auth/RBAC deve estar no core (todos os módulos dependem dele).

**Implementações de Casos de Uso:**

```
internal/usecase/
├── auth_usecase.go        Registro, login, gerenciamento de senha
├── role_usecase.go        Gerenciamento de papéis
└── permission_usecase.go  Gerenciamento de permissões
```

---

### 4. Sistema de Gerenciamento de Módulos

```
pkg/module/
├── module.go       Interface de módulo
├── registry.go     Registro de módulos + resolução de dependências
├── config/         Carregador de config de módulo
└── base.go         Helper BaseModule
```

**Propósito:** Orquestração - descobrir, inicializar, iniciar/parar módulos.

**Status:** Correto - Core é o orquestrador, módulos são os trabalhadores.

**Funcionalidades Principais:**

- Carregamento dinâmico de módulos via auto-registro `init()`
- Resolução de dependências (ordenação topológica)
- Gerenciamento de ciclo de vida (Initialize → RegisterRoutes → Start → Stop)
- Gerenciamento de configuração (cada módulo carrega sua própria configuração)

---

### 5. Infraestrutura de Barramento de Eventos

```
pkg/bus/
├── bus.go          Interface do barramento de eventos
├── memory/         Adaptador em memória (dev/test)
├── redis/          Adaptador Redis (produção)
└── factory.go      Factory de adaptador com fallback
```

**Propósito:** Infraestrutura de comunicação inter-módulos.

**Status:** Correto - Core fornece o barramento, módulos o utilizam.

---

### 6. Infraestrutura do Sistema de Purga

```
pkg/purge/
├── handler.go      Registro de handlers (módulos registram handlers)
└── (NOVO) Registro de políticas (módulos registram políticas de retenção)
```

```
internal/usecase/
└── purge_usecase.go  Apenas orquestração (obtém políticas do registro)
```

**Propósito:** Infraestrutura de agendador - módulos definem o que/quando purgar.

**Status:** CORRIGIDO (refatoração recente) - Core orquestra, módulos implementam.

---

## Responsabilidades dos Módulos

### Módulos Atuais

#### 1. Módulo Posts (`internal/modules/posts/`)

```
posts/
├── module.go               Implementação do módulo
├── register.go             Auto-registro via init()
├── config/                 Configs YAML próprias (dev, test, prod)
│   └── config.*.yaml
├── domain/entity/          Entidades Post, Comment, Like
├── usecase/                Lógica de negócio
├── adapter/
│   ├── http/               Handlers, DTOs, rotas
│   ├── repository/         Implementações Postgres
│   └── purge/              Handlers de purga para posts+comments
└── README.md
```

**Funcionalidades:**

- Posts de usuário (criar, atualizar, deletar, soft-delete)
- Comentários com threading (profundidade máxima configurável)
- Likes (posts + comentários)
- Handlers de purga com políticas de retenção (90 dias posts, 30 dias comentários)

**Status:** Totalmente independente - Sem importações de internal/domain ou internal/usecase

---

#### 2. Módulo Profiles (`internal/modules/profiles/`)

```
profiles/
├── module.go               Implementação do módulo
├── register.go             Auto-registro
├── config/                 Configs YAML próprias
│   └── config.*.yaml
├── entity/                 Entidades UserProfile, UserContact
├── usecase/                Lógica de negócio
└── adapter/
    ├── http/               Handlers, DTOs, rotas
    └── repository/         Implementações Postgres
```

**Funcionalidades:**

- Perfis de usuário (bio, avatar, links sociais)
- Contatos de usuário (email, telefone, múltiplos tipos)
- Verificação de contato
- Gerenciamento de contato primário

**Status:** Totalmente independente - Profiles+contacts mesclados em um módulo coeso

---

#### 3. Módulo Warehouse (`internal/modules/warehouse/`)

**Status:** Comentado (módulo comercial, licença obrigatória)

**Propósito:** Gerenciamento de inventário para implantações comerciais.

---

## Gerenciamento de Configuração

### Configuração do Core

```yaml
# config/app.{env}.yaml - Apenas infraestrutura do core
app:
  name: "Promenade"
  environment: "development"

server:
  host: "localhost"
  port: 8081

database:
  host: "localhost"
  port: 5432

jwt:
  secret: "..."

bus:
  adapter: "memory" # ou "redis"

purge:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 1000
```

**O que NÃO está na configuração do core:**

- Políticas de retenção específicas de entidades → Movidas para módulos
- Configurações específicas de módulos → Movidas para módulos
- Configuração de lógica de negócio → Movida para módulos

---

### Configuração de Módulos

```yaml
# internal/modules/posts/config/config.dev.yaml
module:
  name: "posts"
  enabled: true
  version: "1.0.0"

posts:
  max_content_length: 10000
  comments:
    max_content_length: 2000
    max_depth: 10

purge:
  user_posts:
    retention_days: 90
    enabled: true
  post_comments:
    retention_days: 30
    enabled: true
```

**Cada módulo:**

- Carrega sua própria configuração via `pkg/module/config.Load()`
- Define suas próprias políticas de retenção
- Registra handlers + políticas via registros globais
- Autonomia total

---

### Registro de Módulos

```yaml
# config/modules.yaml - Quais módulos carregar
modules:
  enabled:
    - posts
    - profiles
    # - warehouse  # Requer chave de licença
```

**Propósito:** Controlar quais módulos estão ativos (licenciamento, recursos, etc.)

---

## Suporte a Licenciamento

### Arquitetura Pronta para Licenciamento

```yaml
# config/modules.yaml (futuro)
modules:
  enabled:
    - warehouse

  config:
    warehouse:
      version: "1.2.0"
      license_key: "WH-ABC-123-XYZ" # Validação de licença
      settings:
        max_items: 10000
```

**Módulo pode validar licença em Initialize():**

```go
func (m *WarehouseModule) Initialize(ctx context.Context, core *Core) error {
    // Carregar config
    cfg := moduleconfig.Load("internal/modules/warehouse/config", env)

    // Validar licença
    licenseKey := cfg.GetString("module.license_key")
    if !validateLicense(licenseKey, "warehouse") {
        return fmt.Errorf("invalid license for warehouse module")
    }

    // Continuar inicialização...
}
```

**Status:** Arquitetura suporta licenciamento - implementação pronta quando necessário.

---

## Gerenciamento de Dependências

### Dependências de Módulos

```go
func (m *MyModule) Dependencies() []string {
    return []string{"posts", "profiles"}  // Este módulo precisa de posts + profiles
}
```

**Registro resolve dependências automaticamente:**

1. Ordenação topológica dos módulos
2. Inicializar em ordem de dependência
3. Erro se dependências circulares ou módulos ausentes

**Exemplo:** Módulo Fleet depende do módulo Warehouse (para peças sobressalentes):

```yaml
modules:
  enabled:
    - warehouse # Deve carregar primeiro
    - fleet # Depende de warehouse
```

**Status:** Sistema de dependências implementado em `pkg/module/registry.go`

---

## Padrões de Comunicação

### 1. Eventos Inter-Módulos (Assíncrono)

```go
// Módulo posts publica evento
event := &PostCreatedEvent{...}
eventBus.Publish(ctx, "post.created", event)

// Módulo profiles assina
eventBus.Subscribe("post.created", func(e bus.Event) {
    // Atualizar estatísticas do usuário
})
```

**Benefícios:**

- Sem importações diretas módulo-para-módulo
- Acoplamento fraco
- Processamento assíncrono

---

### 2. Registro de Módulos (Síncrono)

```go
// Obter outro módulo
postsModule := core.Registry.Get("posts")

// Chamar métodos (se o módulo expõe API pública)
stats := postsModule.(PostsModuleAPI).GetUserStats(userID)
```

**Benefícios:**

- Comunicação direta quando necessário
- Interfaces type-safe
- Usar com moderação - preferir eventos

---

## Arquitetura de Roteador

### Rotas do Core

```go
// internal/adapter/http/v1/router/router.go
type V1Router struct {
    // Apenas rotas de infraestrutura do core
    HealthRouter   *gin.RouterGroup
    AuthRouter     *gin.RouterGroup
    CountryRouter  *gin.RouterGroup
    CurrencyRouter *gin.RouterGroup
    RBACRouter     *gin.RouterGroup  // Roles + Permissions
    AdminRouter    *gin.RouterGroup  // Gerenciamento de purga
}
```

**O que NÃO está no roteador do core:**

- Rotas de posts → Movidas para o módulo posts
- Rotas de comentários → Movidas para o módulo posts
- Rotas de perfil → Movidas para o módulo profiles
- Rotas de contato → Movidas para o módulo profiles

---

### Rotas de Módulos

```go
// Módulo posts registra suas próprias rotas
func (m *PostsModule) RegisterRoutes(router *gin.RouterGroup) {
    postsGroup := router.Group("/posts")
    {
        postsGroup.GET("", m.postHandler.ListPosts)
        postsGroup.POST("", m.postHandler.CreatePost)
        // ...
    }

    commentsGroup := router.Group("/comments")
    {
        commentsGroup.POST("", m.commentHandler.CreateComment)
        // ...
    }
}
```

**Resultado:**

- Core: `/api/v1/auth/*`, `/api/v1/countries/*`, `/api/v1/admin/*`
- Módulo Posts: `/api/v1/posts/*`, `/api/v1/comments/*`
- Módulo Profiles: `/api/v1/profiles/*`, `/api/v1/contacts/*`

**Status:** Separação limpa - cada módulo possui suas rotas

---

## Gerenciamento de Banco de Dados

### Sistema de Migrações

**Migrações do core (baseadas em namespace com nomes descritivos):**

```
migrations/core/
├── 000001_core_init_uuid_v7.up.sql            UUID v7 + triggers
├── 000002_core_auth_full.up.sql               Tabelas de auth (users, sessions, tokens)
├── 000003_core_rbac_full.up.sql               RBAC (roles, permissions)
├── 000004_core_ref_timezones.up.sql           Dados de referência
├── 000005_core_ref_languages.up.sql           Dados de referência
└── 000006_core_ref_countries_currencies.up.sql Dados de referência
```

**Migrações de módulos (baseadas em namespace com prefixos de módulo):**

```
migrations/posts/
├── 000001_posts_posts.up.sql                  Tabela posts
├── 000002_posts_comments.up.sql               Tabela comments
└── 000003_posts_comment_likes.up.sql          Comment likes

migrations/profiles/
├── 000001_profiles_contacts.up.sql            Contatos de usuário
└── 000002_profiles_profiles.up.sql            Perfis de usuário
```

**Futuro:** Módulos podem registrar migrações programaticamente:

```go
func (m *MyModule) RegisterMigrations() []module.Migration {
    return []module.Migration{
        {Version: 1, Up: "CREATE TABLE my_table ...", Down: "DROP TABLE my_table"},
    }
}
```

**Status:** Atualmente baseado em arquivos, sistema programático pronto em `pkg/module/module.go`

---

## Estratégia de Testes

### Testes do Core

```
internal/
├── domain/entity/*_test.go         Testes unitários de entidade
├── usecase/*_test.go               Testes unitários de casos de uso
└── adapter/repository/*_test.go    Testes de integração de repositório
```

**Foco:** Auth, RBAC, dados de referência, infraestrutura.

---

### Testes de Módulos

```
internal/modules/posts/
├── usecase/*_test.go               Testes unitários de lógica de negócio
├── adapter/repository/*_test.go    Testes de repositório
└── module_test.go                  Testes de integração de módulo
```

**Status:** Cada módulo testa sua própria lógica independentemente

---

### Testes Smoke

```
test/smoke/
├── auth_smoke_test.go              Fluxos de auth do core
├── rbac_smoke_test.go              Fluxos RBAC do core
├── user_post_smoke_test.go         Módulo posts (precisa atualização)
└── user_profile_smoke_test.go      Módulo profiles (precisa atualização)
```

**Status:** Testes smoke precisam de atualizações de caminhos de importação após migração de módulos

---

## Verificação de Violações →

### CORRIGIDO: Core tinha políticas de purga específicas de entidades

**Antes:**

```go
//  Core sabia sobre entidades de módulos
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

**Depois:**

```go
//  Core tem apenas infraestrutura
type PurgeConfig struct {
    Enabled   bool
    Schedule  string
    BatchSize int
}

// Módulos registram políticas via purge.DefaultPolicyRegistry
```

---

### CORRIGIDO: Posts/Comentários estavam no core

**Antes:** Posts e comentários tinham entidades, casos de uso, handlers em `internal/`

**Depois:** Migração completa para `internal/modules/posts/`

**Deletado do core:** 15.000+ linhas de código movidas para módulo

---

### CORRIGIDO: Perfis/Contatos estavam no core

**Antes:** Perfis e contatos espalhados por `internal/domain`, `internal/usecase`, `internal/adapter`

**Depois:** Migração completa para `internal/modules/profiles/`

**Resultado:** Core verdadeiramente minimalista - apenas infraestrutura + dados de referência

---

## Resumo: Core vs Módulos

| Componente          | Localização | Propósito             | Status       |
| ------------------- | ----------- | --------------------- | ------------ |
| **Autenticação**    | Core        | Fundação de segurança | Correto      |
| **RBAC**            | Core        | Controle de acesso    | Correto      |
| **Países**          | Core        | Dados de referência   | Correto      |
| **Moedas**          | Core        | Dados de referência   | Correto      |
| **Fusos Horários**  | Core        | Dados de referência   | Correto      |
| **Idiomas**         | Core        | Dados de referência   | Correto      |
| **Banco de Dados**  | Core        | Infraestrutura        | Correto      |
| **Event IBus**       | Core        | Infraestrutura        | Correto      |
| **Agendador Purga** | Core        | Infraestrutura        | Correto      |
| **Registro Módulo** | Core        | Orquestração          | Correto      |
|                     |             |                       |
| **Posts**           | Módulo      | Lógica de negócio     | Independente |
| **Comentários**     | Módulo      | Lógica de negócio     | Independente |
| **Likes**           | Módulo      | Lógica de negócio     | Independente |
| **Perfis**          | Módulo      | Lógica de negócio     | Independente |
| **Contatos**        | Módulo      | Lógica de negócio     | Independente |
| **Warehouse**       | Módulo      | Lógica de negócio     | Licenciável  |

---

## Recomendações

### 1. Core está Limpo

O core atual contém apenas:

- Serviços de infraestrutura
- Dados de referência
- Segurança (auth + RBAC)
- Interfaces de gerenciamento (registros)

**Ação:** Nenhuma mudança necessária - arquitetura está correta.

---

### 2. Módulos são Independentes

Cada módulo:

- Tem sua própria estrutura entity/usecase/adapter
- Carrega sua própria configuração
- Registra handlers/políticas/permissões
- Pode ser habilitado/desabilitado via config

**Ação:** Nenhuma mudança necessária - módulos estão adequadamente isolados.

---

### 3. Pronto para Licenciamento

A arquitetura suporta:

- Validação de chave de licença na inicialização do módulo
- Configuração de licença por módulo
- Gerenciamento de dependências (módulo licenciado depende de módulo gratuito)

**Ação:** ⏳ Implementar validação de licença quando módulos comerciais estiverem prontos.

---

### 4. TODOs Menores

1. **Módulo audit** - Adicionar sistema abrangente de log de auditoria
2. **Atualizar testes smoke** - Corrigir caminhos de importação após migração de módulos
3. **Migrações programáticas** - Ativar `RegisterMigrations()` nos módulos
4. **Documentação API** - Atualizar Swagger para refletir rotas de módulos

---

## Conclusão

**Avaliação da Arquitetura: CONFORME**

Promenade implementa com sucesso uma **arquitetura de plugins** com:

- Separação limpa entre Core (infraestrutura) e Módulos (lógica de negócio)
- Independência de módulos (sem dependências do core)
- Carregamento dinâmico de módulos com resolução de dependências
- Autonomia de configuração (cada módulo possui sua config)
- Suporte a licenciamento (pronto para módulos comerciais)
- Comunicação orientada a eventos (acoplamento fraco)

**Core é verdadeiramente minimalista:**

- Gerenciadores de infraestrutura
- Dados de referência
- Fundação de segurança (auth + RBAC)

**Módulos são autocontidos:**

- Próprias entidades, casos de uso, adaptadores
- Própria configuração
- Próprias políticas de purga
- Plugáveis (habilitar/desabilitar via config)

**Próximos Passos:**

1. Implementar validação de licença para módulos comerciais
2. Atualizar testes smoke
3. Considerar extração de timezone/language para módulo "reference" separado se crescerem muito

---

**Data da Auditoria:** 22 de dezembro de 2025
**Auditor:** AI Assistant (GitHub Copilot)
**Status:** APROVADO - Arquitetura é sólida e corretamente implementada
