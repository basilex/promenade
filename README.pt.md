# Promenade

[🇬🇧 English](README.md) | [🇺🇦 Українська](README.uk.md) | [🇩🇪 Deutsch](README.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](README.es.md)

> **📝 Nota sobre traduções**: Alguns documentos técnicos (migrations/, internal/, pkg/, test/) e guias especializados (SOFT*DELETE.md, LOGGING.md, VALIDATION.md, CREDENTIALS.md, MAKEFILE_ARCHITECTURE.md, MIGRATION_ARCHITECTURE.md, REDIS_BUS_TESTING.md, MOCK*\*.md) estão disponíveis apenas em inglês no momento. Os principais documentos de arquitetura estão completamente traduzidos para português. Consulte [docs/pt/INDEX.pt.md](docs/pt/INDEX.pt.md) para a lista de traduções disponíveis.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**API REST pronta para produção** construída com **Clean Architecture**, **sistema modular de plugins** e **migrações de banco de dados baseadas em namespace**. Apresenta módulos de negócios autônomos, PostgreSQL com UUID v7, arquitetura orientada a eventos e infraestrutura abrangente de testes.

---

## Visão Geral da Arquitetura

Promenade segue uma **arquitetura de camadas estrita** onde o **Core orquestra** e os **Módulos executam** a lógica de negócios:

**Camada Core** - Orquestrador + Infraestrutura + Serviços Compartilhados

- Autenticação & Autorização (RBAC)
- Event Bus (Memory/Redis)
- Banco de Dados & Transações
- Logging & Configuração
- Registro de Módulos & Ciclo de Vida
- Dados de Referência (países, moedas, regiões, cidades, fusos horários, idiomas, métodos de pagamento)

**Camada de Módulos** - Slices Verticais Independentes (Áreas de Domínio)

| Módulo        | Entidades                        | Descrição                     | Status       |
| ------------- | -------------------------------- | ----------------------------- | ------------ |
| Posts         | posts, comments, likes           | Conteúdo gerado por usuários  | Gratuito     |
| Profiles      | contacts, profiles               | Perfis de usuários            | Gratuito     |
| **Analytics** | **metrics, reports, dashboards** | **Analytics e relatórios**    | **Gratuito** |
| Warehouse     | products, inventory              | Gestão de inventário (futuro) | Planejado    |

Cada módulo é autocontido com:

- Próprias entidades de domínio & lógica de negócios
- Próprias migrações de banco de dados (baseadas em namespace)
- Próprios repositórios & casos de uso
- Próprios handlers HTTP & rotas
- Própria configuração & ciclo de vida
- Opcional: Políticas de purga, permissões, eventos

### Princípios Fundamentais

1. **Core como Orquestrador**

   - Core **gerencia** o ciclo de vida dos módulos (init, start, stop)
   - Core **fornece** serviços compartilhados (auth, eventos, DB, logging)
   - Core **sabe QUANDO** chamar módulos, mas não **COMO** eles funcionam
   - Core **nunca** importa código específico de módulos

2. **Módulos como Workers**

   - Módulos **implementam** lógica de negócios específica do domínio
   - Módulos se **registram** via funções `init()`
   - Módulos são **independentes** - podem ser ativados/desativados sem afetar outros
   - Módulos **nunca** importam código de outros módulos (apenas `pkg/*`)
   - Cada módulo = slice vertical completo (entidades → handlers)

3. **Camadas Clean Architecture**

   ```
   Domain (entidades, interfaces) → Use Case (lógica de negócios)
      ↓                                      ↓
   Adapter (repos, handlers) → Infrastructure (DB, HTTP, eventos)
   ```

   - **Regra de Dependência**: Camadas internas nunca dependem de camadas externas
   - Casos de uso dependem apenas de interfaces de domínio, nunca de implementações concretas

4. **Migrações Baseadas em Namespace**
   - Cada namespace (core, posts, profiles) tem histórico de versão independente
   - Migrações ficam em `migrations/{namespace}/NNNNNN_descricao.{up|down}.sql`
   - Migrações do Core executam primeiro, depois módulos habilitados
   - Verdadeira autonomia de módulos - habilitar/desabilitar sem conflitos de esquema

---

## Início Rápido

### Pré-requisitos

- **Go 1.25+**
- **Docker & Docker Compose** (para PostgreSQL, Redis)
- **Make** (para automação)

### 1. Clonar & Configurar

```bash
git clone https://github.com/basilex/promenade.git
cd promenade

# Instalar dependências de desenvolvimento
make install

# Iniciar PostgreSQL + Redis via Docker
make docker-up
```

### 2. Executar Migrações

As migrações executam **automaticamente** no início da aplicação, mas você também pode executá-las manualmente:

```bash
# Verificar status das migrações para todos os namespaces
make migrate-status

# Executar todas as migrações (core + módulos habilitados)
make migrate

# Executar namespace específico
make migrate-core                    # Apenas Core
make migrate-module MODULE=posts     # Módulo específico
```

### 3. Iniciar Aplicação

```bash
# Modo de desenvolvimento (hot reload, debug logging)
make dev

# Ou compilar e executar binário
make build
./bin/promenade
```

Servidor inicia em **http://localhost:8081**

---

## Estrutura de Documentação

### Documentação Principal

| Documento                                                                      | Descrição                                                                     |
| ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| **[docs/pt/ARCHITECTURE_OVERVIEW.pt.md](docs/pt/ARCHITECTURE_OVERVIEW.pt.md)** | Diagramas visuais de arquitetura, responsabilidades de camadas, ciclo de vida |
| **[docs/pt/ARCHITECTURE_QUICKREF.pt.md](docs/pt/ARCHITECTURE_QUICKREF.pt.md)** | Referência rápida, árvores de decisão, erros comuns                           |
| **[docs/pt/ARCHITECTURE_AUDIT.pt.md](docs/pt/ARCHITECTURE_AUDIT.pt.md)**       | Auditoria de conformidade de arquitetura, checklist de verificação            |
### Sistema de Módulos

| Documento                                                                                | Descrição                                              |
| ---------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| **[docs/pt/MODULE_DEVELOPMENT.pt.md](docs/pt/MODULE_DEVELOPMENT.pt.md)**                 | Criação de novos módulos, melhores práticas            |
| **[docs/pt/MODULE_INDEPENDENCE.pt.md](docs/pt/MODULE_INDEPENDENCE.pt.md)**               | Regras de autonomia de módulos, gestão de dependências |
| **[docs/pt/MODULE_CONFIG_ARCHITECTURE.pt.md](docs/pt/MODULE_CONFIG_ARCHITECTURE.pt.md)** | Sistema de configuração de módulos                     |

### Infraestrutura & Sistemas

| Documento                                                                          | Descrição                                                         |
| ---------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| **[docs/pt/PURGE_ARCHITECTURE.pt.md](docs/pt/PURGE_ARCHITECTURE.pt.md)**           | Sistema automatizado de expurgação de dados (baseado em registro) |
### Guias de Desenvolvimento

| Documento                                                                        | Descrição                                   |
| -------------------------------------------------------------------------------- | ------------------------------------------- |
| **[docs/pt/TESTING_GUIDE.pt.md](docs/pt/TESTING_GUIDE.pt.md)**                   | Melhores práticas de testes, padrões        |
| **[docs/pt/TESTING_INFRASTRUCTURE.pt.md](docs/pt/TESTING_INFRASTRUCTURE.pt.md)** | Configuração de infraestrutura de testes    |

### Referências Técnicas

| Documento                                                      | Descrição                                                  |
| -------------------------------------------------------------- | ---------------------------------------------------------- |
| **[docs/pt/UUID_V7_GUIDE.pt.md](docs/pt/UUID_V7_GUIDE.pt.md)** | Implementação e benefícios do UUID v7                      |
| **[docs/SOFT_DELETE.md](docs/SOFT_DELETE.md)**                 | Padrão soft delete para conteúdo de usuários               |
| **[docs/pt/AUTHORIZATION.pt.md](docs/pt/AUTHORIZATION.pt.md)** | Sistema RBAC (4 papéis, permissões wildcard)               |
| **[docs/LOGGING.md](docs/LOGGING.md)**                         | Logging estruturado com slog                               |
| **[docs/VALIDATION.md](docs/VALIDATION.md)**                   | Padrões de validação de requisições                        |
| **[docs/CREDENTIALS.md](docs/CREDENTIALS.md)**                 | Usuários de teste padrão e credenciais                     |
| **[docs/pt/INDEX.pt.md](docs/pt/INDEX.pt.md)**                 | Índice completo de documentação com trilhas de aprendizado |

---

## Sistema de Módulos

### Módulos Disponíveis

#### **Módulo Posts** (`internal/modules/posts`)

Gestão de conteúdo gerado por usuários:

- **Entidades**: Posts, Comments, Likes
- **Funcionalidades**: Criar/editar posts, threads de comentários, sistema de likes
- **Migrações**: 3 migrações (namespace: `posts`)
- **Configuração**: `config/modules.yaml` → `posts`

#### **Módulo Profiles** (`internal/modules/profiles`)

Gestão de perfis de usuários e contatos:

- **Entidades**: UserProfiles, UserContacts
- **Funcionalidades**: Gestão de perfis, informações de contato
- **Migrações**: 2 migrações (namespace: `profiles`)
- **Configuração**: `config/modules.yaml` → `profiles`

#### **Módulo Analytics** (`internal/modules/analytics`) - Gratuito

Analytics, métricas e relatórios:

- **Status**: Gratuito - Disponível para todos os usuários
- **Entidades**: Metrics, Reports, Dashboards
- **Funcionalidades**: Coleta de métricas, relatórios personalizados, dashboards visuais
- **Migrações**: 1 migração (namespace: `analytics`)
- **Caso de Uso**: Business intelligence, monitoramento de performance, insights de dados

#### **Módulo Warehouse** (`internal/modules/warehouse`) - Módulo Futuro

Gestão de inventário e produtos (planejado):

- **Status**: Planejado - Estrutura existe como placeholder, ainda não implementado
- **Caso de Uso**: E-commerce, sistemas de inventário, varejo

### Estrutura de Módulo

Cada módulo segue uma estrutura consistente:

```
internal/modules/{module}/
├── module.go           # Registro de módulo & ciclo de vida
├── domain/
│   └── entity/         # Entidades de domínio
├── repository/         # Interfaces de acesso a dados & implementações
├── usecase/            # Lógica de negócios
├── adapter/
│   └── handler/        # Handlers HTTP & DTOs
└── README.md           # Documentação específica do módulo
```

### Habilitando/Desabilitando Módulos

Edite `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts # Conteúdo gerado por usuários
    - profiles # Perfis de usuários + contatos
    - analytics # Business analytics (requer licença)
    # - warehouse  # Futuro: Gestão de inventário
```

Módulos carregam automaticamente na inicialização da aplicação.

---

## Migrações de Banco de Dados

### Sistema Baseado em Namespace

Cada namespace mantém **histórico de versão independente**:

```
migrations/
├── core/               # Infraestrutura core (sempre executa primeiro)
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000003_core_rbac_full.up.sql
│   ├── 000004_core_ref_timezones.up.sql
│   ├── 000005_core_ref_languages.up.sql
│   ├── 000006_core_ref_countries_currencies.up.sql    # 145 países, 124 moedas
│   ├── 000007_core_ref_regions_cities.up.sql          # 30 regiões, 17 cidades
│   └── 000008_core_ref_payment_methods.up.sql         # 40+ métodos de pagamento
├── posts/              # Migrações do módulo Posts
│   ├── 000001_posts_posts.up.sql
│   ├── 000002_posts_comments.up.sql
│   └── 000003_posts_comment_likes.up.sql
├── profiles/           # Migrações do módulo Profiles
│   ├── 000001_profiles_contacts.up.sql
│   └── 000002_profiles_profiles.up.sql
└── analytics/          # Migrações do módulo Analytics (comercial)
    └── 000001_analytics_tables.up.sql
```

### Comandos de Migração

```bash
# Status para todos os namespaces
make migrate-status

# Executar todas (core + módulos habilitados)
make migrate

# Executar namespace específico
make migrate-core
make migrate-module MODULE=posts

# Rollback
make migrate-rollback MODULE=posts STEPS=1

# Criar nova migração
make migrate-create MODULE=posts NAME=add_post_views
make migrate-create-core NAME=add_audit_log
```

**Auto-migrações**: Migrações executam automaticamente no início da aplicação (core primeiro, depois módulos habilitados).

---

## Autenticação & Autorização

### Usuários de Teste Padrão

| Email                           | Senha      | Papel     | Permissões                  |
| ------------------------------- | ---------- | --------- | --------------------------- |
| `system@promenade.com`          | `passw0rd` | Admin     | Acesso completo (`*`)       |
| `admin@promenade.com`           | `passw0rd` | Admin     | Gestão de usuários/conteúdo |
| `moderator@promenade.com`       | `passw0rd` | Moderator | Moderação de conteúdo       |
| `alexander.vasilenko@gmail.com` | `03041965` | User      | Operações básicas           |

**Altere as senhas antes do deploy em produção!**

### Sistema RBAC

- **4 Papéis do Sistema**: Admin, Moderator, User, Guest
- **Permissões Wildcard**: `posts:*` (todas as ações de posts), `*` (acesso completo)
- **Formato Recurso-Ação**: `posts:create`, `users:delete`, `comments:moderate`

**Guia completo RBAC**: [docs/pt/AUTHORIZATION.pt.md](docs/pt/AUTHORIZATION.pt.md)

---

## Testes

**388 testes** em todas as camadas (100% aprovados, ~41 segundos):

```bash
# Executar todos os testes (unit + integration + smoke)
make test               # Todos os testes (~41s)

# Executar por tipo
make test-unit          # Apenas testes unitários (183 testes, ~5s)
make test-integration   # Testes de integração (91 testes, ~36s)
make test-smoke         # Smoke tests (114 testes, ~4s)

# Relatório de cobertura
make test-coverage
```

### Estrutura de Testes

- **Testes Core**: Entidades de domínio (Country, Currency, Language, Timezone, Permission, Role, User, Session, políticas de Purga)
- **Core Use Cases**: Auth, RBAC, CRUD de dados de referência, operações de Purga
- **Testes de Módulos**: Posts (Comment, Post, PostStatus), Profiles (UserContact, UserProfile, ContactType, Gender), Analytics (Metrics, Reports, validação de licença)
- **Testes de Integração**: Operações de repositório com PostgreSQL real na porta 5433
- **Helpers de Teste**: `test/helpers/` e `test/integration/` para fixtures, configuração de banco de dados, gestão de transações

**Guias de teste**:

- [docs/pt/TESTING_GUIDE.pt.md](docs/pt/TESTING_GUIDE.pt.md) - Melhores práticas

---

## Event Bus

**Event bus de adaptador duplo** para operações assíncronas:

### Adaptador Memory

- Pub/Sub em memória (goroutines + channels)
- **Caso de uso**: Desenvolvimento, testes, deployments de instância única
- **Vantagens**: Zero dependências, rápido, simples
- **Configuração**: `BUS_ADAPTER=memory` (padrão)

### Adaptador Redis

- Pub/Sub distribuído via Redis
- **Caso de uso**: Deployments de produção multi-instância
- **Vantagens**: Persistente, escalável, tolerante a falhas
- **Configuração**: `BUS_ADAPTER=redis` + configurações de conexão Redis
- **Fallback**: Auto-fallback para memory se Redis indisponível

### Exemplo de Uso

```go
// Publicar evento
event := &UserRegisteredEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    UserID:    user.ID,
    Email:     user.Email,
}
eventBus.Publish(ctx, bus.TopicUserRegistered, event)

// Assinar eventos
eventBus.Subscribe(ctx, bus.TopicUserRegistered, func(ctx context.Context, e bus.Event) error {
    evt := e.(*UserRegisteredEvent)
    // Enviar email de boas-vindas
    return emailService.SendWelcome(ctx, evt.Email)
})
```

---

## Comandos Makefile

### Desenvolvimento

```bash
make dev                # Iniciar servidor dev (hot reload)
make build              # Compilar binário de produção
make run                # Executar binário compilado
make lint               # Executar linter (golangci-lint)
make fmt                # Formatar código
make config-show        # Mostrar configuração YAML (use ENV=dev|test|prod)
```

### Testes

```bash
make test                      # Todos os testes (core + módulos)
make test-core                 # Apenas testes core (domain + usecase)
make test-modules              # Todos os testes de módulos
make test-module-posts         # Testes do módulo Posts
make test-module-profiles      # Testes do módulo Profiles
make test-coverage             # Gerar relatório de cobertura HTML
```

### Banco de Dados

```bash
make migrate                   # Executar todas as migrações (core + módulos habilitados)
make migrate-status            # Mostrar status de migração
make migrate-core              # Migrar apenas core
make migrate-module MODULE=posts          # Migrar módulo específico
make migrate-rollback MODULE=posts STEPS=1  # Rollback
make migrate-create MODULE=posts NAME=xxx  # Criar migração de módulo
make migrate-create-core NAME=xxx          # Criar migração core
```

### Docker

```bash
make docker-up          # Iniciar PostgreSQL + Redis
make docker-down        # Parar serviços
make docker-clean       # Remover containers + volumes
make docker-logs        # Ver logs
```

### Swagger

```bash
make swagger-all        # Gerar documentação API (v1 + v2)
make swagger-v1         # Gerar apenas docs v1
make swagger-v2         # Gerar apenas docs v2
```

**Guia completo Makefile**: [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md)

---

## Docker

### Configuração de Desenvolvimento

```bash
# Iniciar serviços
make docker-up

# Ver logs
make docker-logs

# Parar serviços
make docker-down

# Limpar tudo (remover volumes)
make docker-clean
```

### Serviços

- **PostgreSQL 16**: Porta 5432, usuário `system`, banco de dados `promenade_dev`
- **Redis 7**: Porta 6379 (para event bus distribuído)

---

## Funcionalidades Técnicas Principais

### Chaves Primárias UUID v7

UUIDs ordenados por tempo para **inserções 2x mais rápidas** que UUID v4 e melhor performance de B-tree.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    ...
);
```

[docs/pt/UUID_V7_GUIDE.pt.md](docs/pt/UUID_V7_GUIDE.pt.md)

### Padrão Soft Delete

Conteúdo gerado por usuários (posts, comentários) usa timestamp `deleted_at` para exclusão segura.

```go
// Sempre filtrar registros soft-deleted
WHERE deleted_at IS NULL
```

[docs/SOFT_DELETE.md](docs/SOFT_DELETE.md)

### Sistema Automatizado de Purga

Sistema baseado em registro onde módulos registram suas políticas de retenção:

```go
purge.DefaultPolicyRegistry.RegisterPolicy(purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
})
```

Scheduler do Core executa jobs de purga via cron. Core não conhece entidades específicas.

[docs/pt/PURGE_ARCHITECTURE.pt.md](docs/pt/PURGE_ARCHITECTURE.pt.md)

### Logging Estruturado

Logging com contexto usando `slog`:

```go
log := logger.FromContext(ctx)  // Inclui request_id, user_id
log.Info("User registered", "email", user.Email)
```

[docs/LOGGING.md](docs/LOGGING.md)

---

## Documentação da API

### Swagger UI

- **v1 API**: http://localhost:8081/api/v1/docs/swagger/index.html
- **v2 API**: http://localhost:8081/api/v2/docs/swagger/index.html

### Health Check

```bash
curl http://localhost:8081/api/v1/health
```

Resposta:

```json
{
  "status": "success",
  "data": {
    "status": "healthy",
    "database": "connected",
    "timestamp": "2025-12-22T16:40:00Z"
  }
}
```

### Versionamento de API

- **v1**: API estável atual (`internal/adapter/http/v1`)
- **v2**: API de próxima geração (`internal/adapter/http/v2`)

Ambas as versões têm:

- Handlers e DTOs isolados
- Documentação Swagger separada
- Registro de rotas independente

---

## Estrutura do Projeto

```
promenade/
├── cmd/
│   ├── api/                    # Ponto de entrada principal da aplicação
│   └── migrate/                # Ferramenta CLI de migração
├── internal/
│   ├── domain/                 # Domínio core (entidades, interfaces)
│   │   ├── entity/             # Entidades de domínio (User, Session)
│   │   ├── event/              # Eventos de domínio (UserRegistered, etc.)
│   │   └── repository/         # Interfaces de repositório
│   ├── usecase/                # Casos de uso core (auth, RBAC)
│   ├── adapter/                # Adaptadores (HTTP, repositórios)
│   │   ├── http/
│   │   │   ├── v1/             # API v1 (handlers, DTOs, rotas)
│   │   │   └── v2/             # API v2
│   │   └── repository/postgres/ # Implementações PostgreSQL
│   ├── infrastructure/         # Infraestrutura (DB, config, scheduler)
│   │   ├── database/
│   │   ├── config/
│   │   ├── notification/
│   │   └── scheduler/
│   └── modules/                # Módulos de negócios (plugins)
│       ├── posts/              # Posts + comentários + likes
│       ├── profiles/           # Perfis de usuários + contatos
│       ├── analytics/          # Analytics + relatórios (Comercial, ativo)
│       └── warehouse/          # Gestão de inventário (futuro)
├── pkg/                        # Pacotes compartilhados (reutilizáveis)
│   ├── bus/                    # Event bus (memory/redis)
│   ├── jwt/                    # Gerenciador JWT
│   ├── logger/                 # Logger estruturado
│   ├── migration/              # Gerenciador de migrações
│   ├── module/                 # Registro de módulos
│   ├── purge/                  # Registro de purga
│   ├── response/               # Helpers de resposta HTTP
│   ├── uuidv7/                 # Gerador UUID v7
│   └── validator/              # Validação de requisições
├── migrations/                 # Migrações baseadas em namespace
│   ├── core/                   # Migrações core (auth, RBAC, dados ref)
│   ├── posts/                  # Migrações do módulo Posts
│   └── profiles/               # Migrações do módulo Profiles
├── test/                       # Infraestrutura de testes
│   ├── helpers/                # Helpers de teste (fixtures, configuração DB)
│   ├── integration/            # Testes de integração
│   ├── smoke/                  # Smoke tests
│   └── mocks/                  # Implementações mock
├── config/                     # Arquivos de configuração
│   ├── app.dev.yaml            # Configuração ambiente dev
│   ├── app.test.yaml           # Configuração ambiente test
│   ├── app.prod.yaml           # Configuração produção
│   └── modules.yaml            # Habilitar/desabilitar módulo + configurações
├── docs/                       # Documentação
├── scripts/                    # Scripts auxiliares
├── templates/                  # Templates de email
└── docker/                     # Configurações Docker
```

---

## Configuração

### Configurações Específicas de Ambiente

Prioridade de configuração (YAML primeiro, `.env` como fallback):

1. `config/app.{dev|test|prod}.yaml` - Configurações de infraestrutura core
2. `config/modules.yaml` - Habilitar/desabilitar módulo + configurações específicas de módulo
3. `.env.{env}.local` / `.env.{env}` / `.env` - Suporte legado

### Exemplo: `config/app.dev.yaml`

```yaml
app:
  name: "Promenade API"
  environment: "development"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8081

database:
  host: "localhost"
  port: 5432
  user: "system"
  password: "passw0rd"
  database: "promenade_dev"

jwt:
  secret: "your-super-secret-jwt-key-change-in-production"
  access_token_duration: 15m
  refresh_token_duration: 168h

bus:
  adapter: "memory" # ou "redis"
  worker_pool_size: 4

purge:
  enabled: true
  schedule: "0 2 * * *" # Diariamente às 2h
  batch_size: 1000
```

### Exemplo: `config/modules.yaml`

```yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics # Módulo comercial (requer licença)
    # - warehouse  # Futuro: Gestão de inventário

  config:
    posts:
      version: "1.0.0"
      settings:
        max_post_length: 10000
        max_comment_depth: 10

    analytics:
      version: "1.0.0"
      license_key: "" # Definir via variável de ambiente ANALYTICS_LICENSE_KEY
      settings:
        metrics_retention_days: 90
```

**Guia de configuração**: [docs/pt/MODULE_CONFIG_ARCHITECTURE.pt.md](docs/pt/MODULE_CONFIG_ARCHITECTURE.pt.md)

---

## Criando Novos Módulos

### Passo 1: Criar Estrutura do Módulo

```bash
mkdir -p internal/modules/mymodule/{domain/entity,repository,usecase,adapter/handler}
```

### Passo 2: Implementar Interface de Módulo

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type Module struct{}

func (m *Module) Name() string { return "mymodule" }

func (m *Module) Initialize(ctx context.Context, core module.Core) error {
    // Registrar rotas, permissões, handlers de purga
    return nil
}

func (m *Module) Start(ctx context.Context) error {
    // Iniciar workers em background
    return nil
}

func (m *Module) Stop(ctx context.Context) error {
    // Graceful shutdown
    return nil
}

func init() {
    module.DefaultRegistry.Register(&Module{})
}
```

### Passo 3: Criar Migrações

```bash
make migrate-create MODULE=mymodule NAME=create_tables
```

### Passo 4: Habilitar Módulo

Adicionar a `config/modules.yaml`:

```yaml
modules:
  enabled:
    - mymodule
```

**Guia completo**: [docs/pt/MODULE_DEVELOPMENT.pt.md](docs/pt/MODULE_DEVELOPMENT.pt.md)

---

## Trilhas de Aprendizado

### Para Novos Desenvolvedores

1. **Início**: [docs/pt/ARCHITECTURE_QUICKREF.pt.md](docs/pt/ARCHITECTURE_QUICKREF.pt.md) - Visão geral de 15 minutos
4. **Prática**: Criar um módulo simples seguindo [docs/pt/MODULE_DEVELOPMENT.pt.md](docs/pt/MODULE_DEVELOPMENT.pt.md)

### Para DevOps/Deployment

1. **Makefile**: [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md)
4. **Configuração**: [docs/pt/MODULE_CONFIG_ARCHITECTURE.pt.md](docs/pt/MODULE_CONFIG_ARCHITECTURE.pt.md)

### Para Arquitetos

1. **Visão Geral da Arquitetura**: [docs/pt/ARCHITECTURE_OVERVIEW.pt.md](docs/pt/ARCHITECTURE_OVERVIEW.pt.md)
2. **Auditoria & Verificação**: [docs/pt/ARCHITECTURE_AUDIT.pt.md](docs/pt/ARCHITECTURE_AUDIT.pt.md)
3. **Independência de Módulos**: [docs/pt/MODULE_INDEPENDENCE.pt.md](docs/pt/MODULE_INDEPENDENCE.pt.md)
4. **Sistema de Migração**: [docs/MIGRATION_ARCHITECTURE.md](docs/MIGRATION_ARCHITECTURE.md)

**Índice completo**: [docs/pt/INDEX.pt.md](docs/pt/INDEX.pt.md)

---

## Contribuindo

1. Fazer fork do repositório
2. Criar branch de feature (`git checkout -b feature/amazing-feature`)
3. Seguir princípios de arquitetura (veja [docs/pt/ARCHITECTURE_QUICKREF.pt.md](docs/pt/ARCHITECTURE_QUICKREF.pt.md))
4. Escrever testes (manter 100% de taxa de aprovação)
5. Commitar mudanças (`git commit -m 'Add amazing feature'`)
6. Push para branch (`git push origin feature/amazing-feature`)
7. Abrir Pull Request

---

## Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

---

## Suporte

- **Documentação**: [docs/pt/INDEX.pt.md](docs/pt/INDEX.pt.md)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Email**: alexander.vasilenko@gmail.com

---

**Construído com Clean Architecture e Go**
