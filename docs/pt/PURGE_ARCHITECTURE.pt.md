# Arquitetura do Sistema de Limpeza

🇬🇧 [English](PURGE_ARCHITECTURE.pt.md) | [🇺🇦 Українська](../uk/PURGE_ARCHITECTURE.uk.md) | [🇩🇪 Deutsch](../de/PURGE_ARCHITECTURE.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/PURGE_ARCHITECTURE.es.md)

## Visão Geral

O sistema de limpeza foi projetado para limpar automaticamente registros soft-deleted com base em políticas de retenção. Ele segue o **padrão orquestrador**, onde:

- **Núcleo** gerencia a infraestrutura (agendador, caso de uso, barramento de eventos)
- **Módulos** possuem a lógica de negócio (manipuladores, políticas de retenção)

## Princípios de Arquitetura

### 1. Independência de Módulo

Cada módulo é responsável por:

- Registrar manipuladores de limpeza para suas entidades
- Definir políticas de retenção para suas entidades
- Implementar a lógica real de limpeza

O núcleo **nunca** conhece tipos de entidades específicas ou políticas de retenção.

### 2. Padrão de Registro

Dois registros globais permitem independência de módulo:

#### Registro de Manipuladores (`purge.DefaultRegistry`)

```go
// Módulo registra manipulador durante a inicialização
handler := NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)
```

#### Registro de Políticas (`purge.DefaultPolicyRegistry`)

```go
// Módulo registra política de retenção
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

### 3. Orquestração do Núcleo

As responsabilidades do núcleo são limitadas a:

1. **Infraestrutura**: Carregar configuração de limpeza do YAML (enabled, schedule, dry_run, batch_size)
2. **Orquestração**: Coletar políticas do registro, criar caso de uso, iniciar agendador
3. **Execução**: Acionar operações de limpeza no cronograma
4. **Eventos**: Publicar eventos de limpeza (sucesso/falha)

O núcleo **NÃO**:

- Conhece entidades específicas
- Define políticas de retenção
- Implementa lógica de limpeza

## Estrutura de Configuração

### Configuração do Núcleo (`config/app.*.yaml`)

```yaml
# Apenas infraestrutura de limpeza do núcleo
purge:
  enabled: true # Interruptor principal
  schedule: "0 2 * * *" # Agendamento cron (2h diariamente)
  dry_run: false # Modo de visualização
  batch_size: 1000 # Registros por lote
```

### Configuração do Módulo (`internal/modules/{nome}/config/config.*.yaml`)

```yaml
# Políticas de retenção do módulo
purge:
  user_posts:
    retention_days: 90 # Manter por 90 dias
    enabled: true

  post_comments:
    retention_days: 30 # Manter por 30 dias
    enabled: true
```

## Fluxo de Implementação

### 1. Inicialização do Módulo

```go
func (m *PostsModule) Initialize(core *Core) error {
    // Carregar configuração do módulo
    cfg := moduleconfig.Load("internal/modules/posts/config", environment)

    // Obter configurações de retenção
    postsRetentionDays := cfg.GetRetentionDays("purge.user_posts.retention_days")
    commentsRetentionDays := cfg.GetRetentionDays("purge.post_comments.retention_days")

    // Registrar manipulador de limpeza
    postHandler := NewPostPurgeHandler(db)
    purge.DefaultRegistry.Register(postHandler)

    // Registrar política de retenção
    policy := purge.RetentionPolicy{
        EntityName:    "user_posts",
        RetentionDays: postsRetentionDays,
        Enabled:       true,
    }
    purge.DefaultPolicyRegistry.RegisterPolicy(policy)

    return nil
}
```

### 2. Inicialização do Núcleo

```go
func InitPurgeModule(purgeConfig config.PurgeConfig, eventBus bus.Bus) {
    // Obter todos os manipuladores registrados
    handlerRegistry := purge.DefaultRegistry

    // Obter todas as políticas registradas
    policyRegistry := purge.DefaultPolicyRegistry
    policies := policyRegistry.GetAllPolicies()

    // Converter para entidades de domínio
    domainPolicies := convertToEntityPolicies(policies, purgeConfig.Enabled)

    // Criar caso de uso
    useCase := usecase.NewPurgeUseCase(
        handlerRegistry,
        domainPolicies,
        purgeConfig.BatchSize,
        eventBus,
    )

    // Criar e iniciar agendador
    scheduler := scheduler.NewScheduler(useCase, purgeConfig.Schedule, ...)
    scheduler.Start(ctx)
}
```

### 3. Execução de Limpeza

```go
// Agendador aciona limpeza no cronograma cron
func (s *Scheduler) runPurge(ctx context.Context) {
    // Caso de uso itera sobre políticas
    for _, policy := range policies {
        // Obter manipulador do registro
        handler, ok := registry.Get(policy.EntityName)
        if !ok {
            continue // Pular se não houver manipulador
        }

        // Executar limpeza
        cutoffDate := time.Now().AddDate(0, 0, -policy.RetentionDays)
        recordsPurged, err := handler.Purge(ctx, cutoffDate, batchSize, dryRun)

        // Publicar eventos
        if err != nil {
            eventBus.Publish(ctx, bus.TopicPurgeFailed, ...)
        } else {
            eventBus.Publish(ctx, bus.TopicPurgeCompleted, ...)
        }
    }
}
```

## Criando um Novo Manipulador de Limpeza

### 1. Implementar Interface do Manipulador

```go
// internal/modules/mymodule/adapter/purge/handler.go
package purge

import (
    "context"
    "time"
    "github.com/jmoiron/sqlx"
)

type MyEntityPurgeHandler struct {
    db *sqlx.DB
}

func NewMyEntityPurgeHandler(db *sqlx.DB) *MyEntityPurgeHandler {
    return &MyEntityPurgeHandler{db: db}
}

func (h *MyEntityPurgeHandler) EntityName() string {
    return "my_entities"
}

func (h *MyEntityPurgeHandler) Purge(
    ctx context.Context,
    cutoffDate time.Time,
    batchSize int,
    dryRun bool,
) (int64, error) {
    query := `
        SELECT id FROM my_entities
        WHERE deleted_at IS NOT NULL
        AND deleted_at < $1
        LIMIT $2
    `

    var ids []string
    if err := h.db.SelectContext(ctx, &ids, query, cutoffDate, batchSize); err != nil {
        return 0, err
    }

    if dryRun {
        return int64(len(ids)), nil // Apenas visualização
    }

    deleteQuery := `DELETE FROM my_entities WHERE id = ANY($1)`
    result, err := h.db.ExecContext(ctx, deleteQuery, pq.Array(ids))
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}
```

### 2. Registrar no Módulo

```go
// internal/modules/mymodule/module.go
func (m *MyModule) Initialize(core *Core) error {
    // Carregar configuração
    cfg := moduleconfig.Load("internal/modules/mymodule/config", env)
    retentionDays := cfg.GetRetentionDays("purge.my_entities.retention_days")

    // Registrar manipulador
    handler := purge.NewMyEntityPurgeHandler(m.db)
    if err := purge.DefaultRegistry.Register(handler); err != nil {
        return err
    }

    // Registrar política
    policy := purge.RetentionPolicy{
        EntityName:    "my_entities",
        RetentionDays: retentionDays,
        Enabled:       true,
    }
    if err := purge.DefaultPolicyRegistry.RegisterPolicy(policy); err != nil {
        return err
    }

    slog.Info("Registered purge for my_entities", "retention_days", retentionDays)
    return nil
}
```

### 3. Adicionar Configuração do Módulo

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true

purge:
  my_entities:
    retention_days: 60
    enabled: true
```

## Benefícios

### ✅ Independência de Módulo

- Módulos possuem completamente sua lógica de limpeza
- Sem dependências do núcleo em entidades de módulos
- Fácil adicionar/remover módulos

### ✅ Clareza de Configuração

- Núcleo: Configurações de infraestrutura
- Módulos: Políticas de negócio
- Separação clara de responsabilidades

### ✅ Testabilidade

- Manipuladores podem ser testados independentemente
- Políticas podem ser alteradas sem mudanças de código
- Padrão de registro permite mocagem fácil

### ✅ Manutenibilidade

- Mudanças em políticas de retenção não requerem mudanças no núcleo
- Novas entidades automaticamente captadas via registro
- Lógica de orquestração centralizada

## Monitoramento

### Eventos Publicados

- `purge.entity.completed` - Limpeza de entidade bem-sucedida
- `purge.entity.failed` - Limpeza de entidade falhou
- `purge.all.completed` - Ciclo completo de limpeza concluído

### Logs

```
level=info msg="Registered purge for user_posts" retention_days=90
level=info msg="Starting purge operation" entity=user_posts cutoff_date=2024-09-22
level=info msg="Purge operation completed" entity=user_posts records_purged=150 duration=2.3s
```

### Endpoints Administrativos

- `GET /api/v1/admin/purge/policies` - Listar todas as políticas de retenção
- `POST /api/v1/admin/purge/preview/:entity` - Visualizar limpeza para entidade
- `POST /api/v1/admin/purge/execute/:entity` - Acionar limpeza manualmente

## Migração do Sistema Antigo

**Antes** (Núcleo conhecia entidades):

```go
// ❌ Núcleo tinha configuração específica de entidade
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

**Depois** (Núcleo tem apenas infraestrutura):

```go
// ✅ Núcleo tem apenas configurações de infraestrutura
type PurgeConfig struct {
    Enabled   bool
    Schedule  string
    DryRun    bool
    BatchSize int
}
```

Módulos agora registram suas próprias políticas via `purge.DefaultPolicyRegistry`.

## Veja Também

- [Independência de Módulos](MODULE_INDEPENDENCE.pt.md)
- [Desenvolvimento de Módulos](MODULE_DEVELOPMENT.pt.md)
- [Guia de Testes](TESTING_GUIDE.pt.md)
