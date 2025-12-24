# Visão Geral da Arquitetura Promenade

[🇬🇧 English](../ARCHITECTURE_OVERVIEW.md) | [🇺🇦 Українська](../uk/ARCHITECTURE_OVERVIEW.uk.md) | [🇩🇪 Deutsch](../de/ARCHITECTURE_OVERVIEW.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/ARCHITECTURE_OVERVIEW.es.md)

Este documento fornece uma visão geral de alto nível da arquitetura Promenade, organizada em torno de **princípios de Clean Architecture** e um **sistema de módulos baseado em plugins**.

---

## Camadas da Arquitetura

### 1. Camada de Aplicação

**Localização:** `cmd/api/main.go`

**Responsabilidades:**

- Inicializar infraestrutura (DB, EventBus, Config, Logger)
- Carregar configuração principal
- Inicializar sistema de módulos (descoberta, registro, ciclo de vida)
- Iniciar servidor HTTP

---

### 2. Infraestrutura Principal

**Localização:** `internal/infrastructure/`

**Componentes:**

| Componente    | Propósito                              |
| ------------- | -------------------------------------- |
| Database      | Gerenciamento de conexão PostgreSQL    |
| Event IBus     | Adaptadores Memory/Redis Pub/Sub       |
| Scheduler     | Agendamento de tarefas baseado em cron |
| Config        | Carregador de configuração YAML        |
| Logger        | Logging estruturado com slog           |
| Email Service | Envio assíncrono de e-mail             |

**Fornece:** Serviços compartilhados para todos os módulos (DB, EventBus, Config, JWT, Logger)

---

### 3. Domínio Principal

**Localização:** `internal/domain/entity/`

**Segurança & Autenticação** (sempre habilitado):

- User (apenas autenticação: email, senha, papéis)
- Session (tokens JWT)
- Role (papéis RBAC)
- Permission (formato resource:action)

**Dados de Referência** (estáveis, compartilhados):

- Country (195+ códigos ISO, regiões)
- Currency (170+ códigos ISO 4217)
- Timezone (500+ fusos horários IANA)
- Language (180+ códigos ISO 639)

**Propósito:** Entidades principais requeridas por TODOS os módulos. Sem lógica de negócios.

---

### 4. Casos de Uso Principais

**Localização:** `internal/usecase/`

**Casos de Uso Disponíveis:**

| Caso de Uso        | Propósito                              |
| ------------------ | -------------------------------------- |
| Auth UseCase       | Registro, login, logout                |
| Role UseCase       | CRUD de papéis, atribuição de papéis   |
| Permission UseCase | CRUD de permissões, controle de acesso |
| Country UseCase    | Listar países, obter por código        |
| Currency UseCase   | Listar moedas, obter por código        |
| Purge UseCase      | Orquestrar jobs de limpeza             |

**Nota:** Casos de uso principais NÃO contêm lógica de negócios - apenas infraestrutura e segurança.

---

### 5. Rotas API Principais

**Localização:** `internal/adapter/http/v1/`

**Endpoints:**

| Rota                   | Propósito                         |
| ---------------------- | --------------------------------- |
| /api/v1/health         | Verificações de saúde             |
| /api/v1/auth/\*        | Login, registro, logout           |
| /api/v1/countries/\*   | Dados de referência               |
| /api/v1/currencies/\*  | Dados de referência               |
| /api/v1/roles/\*       | Gerenciamento RBAC                |
| /api/v1/permissions/\* | Gerenciamento RBAC                |
| /api/v1/admin/\*       | Limpeza, gerenciamento do sistema |

---

## Sistema de Módulos

### Registro & Gerenciador de Módulos

**Localização:** `pkg/module/`

**Propósito:** Orquestra ciclo de vida dos módulos

**Recursos do Registro:**

- `Register(module)` - Auto-registro via `init()`
- `GetEnabled(config)` - Filtrar módulos habilitados da config
- `InitializeAll()` - Inicializar em ordem de dependência
- `StartAll()` - Iniciar workers em background
- `StopAll()` - Desligamento gracioso

**Fornece aos Módulos:**

- Conexão DB compartilhada
- EventBus compartilhado
- Gerenciador JWT compartilhado
- Carregador Config compartilhado

---

### Módulo: Posts

**Localização:** `internal/modules/posts/`

**Status:** Habilitado (Gratuito)

**Entidades:** Post, Comment, Like

**Recursos:**

- Criar, atualizar, deletar posts
- Comentários em thread (profundidade máxima configurável)
- Curtir posts e comentários
- Soft delete com retenção configurável

**Configuração:**

- max_content_length: 10000
- comments.max_depth: 10
- purge.user_posts.retention_days: 90
- purge.post_comments.retention_days: 30

**Rotas:** /api/v1/posts/_, /api/v1/comments/_, /api/v1/likes/\*

---

### Módulo: Profiles

**Localização:** `internal/modules/profiles/`

**Status:** Habilitado (Gratuito)

**Entidades:** Profile, Contact

**Recursos:**

- Gerenciamento de perfil de usuário
- Informações de contato (email, telefone, etc.)
- Verificação de contato
- Designação de contato principal

**Configuração:**

- profiles.max_per_user: 1
- contacts.max_per_user: 5
- contacts.verification_required: true

**Rotas:** /api/v1/profiles/_, /api/v1/contacts/_

---

### Módulo: Warehouse

**Localização:** `internal/modules/warehouse/`

**Status:** Desabilitado (Comercial - requer licença)

**Recursos** (quando licenciado):

- Gerenciamento de inventário
- Rastreamento de estoque
- Escaneamento de código de barras
- Localizações de armazém

**Rotas:** /api/v1/warehouse/\* (quando habilitado)

---

## Comunicação Inter-Módulos

### Event IBus

**Localização:** `pkg/bus/`

**Adaptadores:**

| Adaptador | Caso de Uso                     | Recursos                 |
| --------- | ------------------------------- | ------------------------ |
| Memory    | Dev/Test/Instância única        | Rápido, sem dependências |
| Redis     | Production/Múltiplas instâncias | Distribuído, persistente |

**Fluxo de Eventos:**

1. Módulo A publica evento no EventBus
2. EventBus distribui para todos os assinantes
3. Módulos B, C, D processam evento assincronamente

**Exemplos:**

- user.registered → enviar email de boas-vindas (assíncrono)
- post.created → atualizar estatísticas do usuário (assíncrono)
- purge.completed → registrar em auditoria (assíncrono)

---

## Arquitetura de Configuração

### Configuração Principal

```
config/
├── app.dev.yaml     - Infraestrutura principal (dev)
├── app.test.yaml    - Infraestrutura principal (test)
├── app.prod.yaml    - Infraestrutura principal (prod)
└── modules.yaml     - Quais módulos carregar
```

### Configuração de Módulos

```
internal/modules/posts/config/
├── config.dev.yaml  - Configurações do módulo Posts (dev)
├── config.test.yaml - Configurações do módulo Posts (test)
└── config.prod.yaml - Configurações do módulo Posts (prod)

internal/modules/profiles/config/
├── config.dev.yaml  - Configurações do módulo Profiles (dev)
├── config.test.yaml - Configurações do módulo Profiles (test)
└── config.prod.yaml - Configurações do módulo Profiles (prod)
```

**Substituições de Ambiente:** `.env.example` (opcional)

---

## Ciclo de Vida do Módulo

### 1. Auto-Registro (via init())

Módulo registra a si mesmo na importação do pacote:

```go
package posts

func init() {
    module.DefaultRegistry.Register(New())
}
```

### 2. Descoberta & Filtragem

- Ler `config/modules.yaml`
- Filtrar módulos habilitados
- Resolver dependências (ordenação topológica)

### 3. Inicialização (em ordem de dependência)

Para cada módulo:

- Carregar config do módulo de `config/config.*.yaml`
- Configurar repositórios
- Configurar casos de uso
- Configurar handlers
- Registrar handlers de limpeza
- Registrar políticas de retenção
- Registrar permissões

### 4. Registro de Rotas

Para cada módulo:

- Montar rotas do módulo no roteador
- Aplicar middleware (auth, RBAC, etc.)

### 5. Assinatura de Eventos

Para cada módulo:

- Assinar eventos relevantes
- Configurar handlers de eventos assíncronos

### 6. Início (workers em background)

Para cada módulo:

- Iniciar jobs cron
- Iniciar workers em background

### 7. Tempo de Execução

- Módulos processam requisições HTTP
- Publicam/assinam eventos
- Executam tarefas agendadas

### 8. Desligamento (em SIGTERM/SIGINT)

Para cada módulo (ordem reversa):

- Parar workers graciosamente
- Fechar conexões
- Limpar recursos

---

## Princípios de Design Principais

### 1. CORE = INFRAESTRUTURA + DADOS DE REFERÊNCIA

- Core fornece serviços (DB, EventBus, Config, JWT)
- Core contém dados de referência estáveis (países, moedas)
- Core gerencia segurança (auth, RBAC)
- Core NÃO contém lógica de negócios

### 2. MÓDULOS = LÓGICA DE NEGÓCIOS

- Módulos são slices verticais autocontidos
- Módulos possuem suas entidades, casos de uso, adaptadores
- Módulos registram handlers, políticas, permissões
- Módulos podem ser habilitados/desabilitados via config
- Módulos NÃO importam de `internal/domain` ou `internal/usecase`
- Módulos NÃO dependem uns dos outros diretamente (usam eventos)

### 3. ARQUITETURA DE PLUGIN

- Carregamento dinâmico via registro de módulos
- Resolução de dependências (ordenação topológica)
- Gerenciamento de ciclo de vida (init → start → stop)
- Auto-registro via `init()`

### 4. AUTONOMIA DE CONFIGURAÇÃO

- Core: `config/app.*.yaml` (apenas infraestrutura)
- Módulos: `internal/modules/{name}/config/config.*.yaml`
- Cada módulo carrega sua própria config
- Configs específicas de ambiente (dev, test, prod)

### 5. ACOPLAMENTO FRACO

- Comunicação baseada em eventos (pub/sub)
- Padrão de registro (handlers, políticas, permissões)
- Dependências baseadas em interface
- Sem imports diretos módulo-para-módulo

### 6. SUPORTE A LICENCIAMENTO

- Chaves de licença por módulo
- Validação de licença em `Initialize()`
- Degradação graciosa se licença inválida

---

## Benefícios

### Modularidade

- Adicionar novos módulos sem tocar no core
- Remover módulos sem quebrar outros
- Testar módulos independentemente

### Escalabilidade

- Módulos comerciais (warehouse, fleet, finance)
- Habilitação de recursos baseada em licença
- Fácil adicionar novas verticais

### Manutenibilidade

- Limites claros (core vs módulos)
- Responsabilidade única (cada módulo possui seu domínio)
- Clareza de configuração (sem config monolítica)

### Testabilidade

- Teste unitário de módulos isoladamente
- Teste de integração com core real/mock
- Teste smoke de fluxos críticos

### Implantabilidade

- Habilitar apenas módulos necessários por implantação
- Teste A/B de novos módulos
- Rollout gradual de recursos

---

## Status Atual

### Core

- Infraestrutura (DB, EventBus, Scheduler, Config, Logger)
- Segurança (Auth, RBAC, JWT, Sessions)
- Dados de Referência (Countries, Currencies, Timezones, Languages)
- Gerenciamento de Módulos (Registry, Lifecycle, Config Loader)
- Orquestração de Limpeza (baseada em registro, sem conhecimento de entidades)

### Módulos

- **Posts** (posts + comments + likes) - Completamente independente
- **Profiles** (profiles + contacts) - Completamente independente
- **Warehouse** (inventory management) - Comercial, desabilitado

### Conformidade de Arquitetura

- Core contém APENAS infraestrutura + dados de referência
- Módulos são COMPLETAMENTE independentes (sem imports do core)
- Configuração é AUTÔNOMA (cada módulo possui config)
- Licenciamento é SUPORTADO (pronto para módulos comerciais)

### Melhorias Recentes

- Sistema de limpeza refatorado (core = orquestrador, módulos = workers)
- Profiles/Contacts migrados para módulo (removidos do core)
- 15.000+ linhas de código removidas do core
- Independência completa de módulos alcançada

---

## Documentação Relacionada

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md) - Auditoria de conformidade de arquitetura
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md) - Guia de referência rápida
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md) - Criando novos módulos
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md) - Princípios de independência de módulos
- [../internal/CORE.md](../internal/CORE.md) - Detalhes dos componentes principais
