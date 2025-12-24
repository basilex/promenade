# Arquitectura del Sistema de Purga

🇬🇧 [English](../PURGE_ARCHITECTURE.md) | [🇺🇦 Українська](../uk/PURGE_ARCHITECTURE.uk.md) | [🇩🇪 Deutsch](../de/PURGE_ARCHITECTURE.de.md) | [🇵🇹 Português](../pt/PURGE_ARCHITECTURE.pt.md) | 🇪🇸 **Español**

## Descripción General

El sistema de purga está diseñado para limpiar automáticamente registros soft-deleted basándose en políticas de retención. Sigue el **patrón orquestador**, donde:

- **Núcleo** gestiona la infraestructura (planificador, caso de uso, bus de eventos)
- **Módulos** poseen la lógica de negocio (manejadores, políticas de retención)

## Principios de Arquitectura

### 1. Independencia de Módulo

Cada módulo es responsable de:

- Registrar manejadores de purga para sus entidades
- Definir políticas de retención para sus entidades
- Implementar la lógica real de purga

El núcleo **nunca** conoce tipos de entidades específicos o políticas de retención.

### 2. Patrón de Registro

Dos registros globales permiten la independencia de módulos:

#### Registro de Manejadores (`purge.DefaultRegistry`)

```go
// Módulo registra manejador durante la inicialización
handler := NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)
```

#### Registro de Políticas (`purge.DefaultPolicyRegistry`)

```go
// Módulo registra política de retención
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

### 3. Orquestación del Núcleo

Las responsabilidades del núcleo están limitadas a:

1. **Infraestructura**: Cargar configuración de purga desde YAML (enabled, schedule, dry_run, batch_size)
2. **Orquestación**: Recopilar políticas del registro, crear caso de uso, iniciar planificador
3. **Ejecución**: Activar operaciones de purga según cronograma
4. **Eventos**: Publicar eventos de purga (éxito/fallo)

El núcleo **NO**:

- Conoce entidades específicas
- Define políticas de retención
- Implementa lógica de purga

## Estructura de Configuración

### Configuración del Núcleo (`config/app.*.yaml`)

```yaml
# Solo infraestructura de purga del núcleo
purge:
  enabled: true # Interruptor principal
  schedule: "0 2 * * *" # Cronograma cron (2 AM diariamente)
  dry_run: false # Modo de vista previa
  batch_size: 1000 # Registros por lote
```

### Configuración del Módulo (`internal/modules/{nombre}/config/config.*.yaml`)

```yaml
# Políticas de retención del módulo
purge:
  user_posts:
    retention_days: 90 # Mantener por 90 días
    enabled: true

  post_comments:
    retention_days: 30 # Mantener por 30 días
    enabled: true
```

## Flujo de Implementación

### 1. Inicialización del Módulo

```go
func (m *PostsModule) Initialize(core *Core) error {
    // Cargar configuración del módulo
    cfg := moduleconfig.Load("internal/modules/posts/config", environment)

    // Obtener configuraciones de retención
    postsRetentionDays := cfg.GetRetentionDays("purge.user_posts.retention_days")
    commentsRetentionDays := cfg.GetRetentionDays("purge.post_comments.retention_days")

    // Registrar manejador de purga
    postHandler := NewPostPurgeHandler(db)
    purge.DefaultRegistry.Register(postHandler)

    // Registrar política de retención
    policy := purge.RetentionPolicy{
        EntityName:    "user_posts",
        RetentionDays: postsRetentionDays,
        Enabled:       true,
    }
    purge.DefaultPolicyRegistry.RegisterPolicy(policy)

    return nil
}
```

### 2. Inicialización del Núcleo

```go
func InitPurgeModule(purgeConfig config.PurgeConfig, eventBus bus.Bus) {
    // Obtener todos los manejadores registrados
    handlerRegistry := purge.DefaultRegistry

    // Obtener todas las políticas registradas
    policyRegistry := purge.DefaultPolicyRegistry
    policies := policyRegistry.GetAllPolicies()

    // Convertir a entidades de dominio
    domainPolicies := convertToEntityPolicies(policies, purgeConfig.Enabled)

    // Crear caso de uso
    useCase := usecase.NewPurgeUseCase(
        handlerRegistry,
        domainPolicies,
        purgeConfig.BatchSize,
        eventBus,
    )

    // Crear e iniciar planificador
    scheduler := scheduler.NewScheduler(useCase, purgeConfig.Schedule, ...)
    scheduler.Start(ctx)
}
```

### 3. Ejecución de Purga

```go
// Planificador activa purga según cronograma cron
func (s *Scheduler) runPurge(ctx context.Context) {
    // Caso de uso itera sobre políticas
    for _, policy := range policies {
        // Obtener manejador del registro
        handler, ok := registry.Get(policy.EntityName)
        if !ok {
            continue // Omitir si no hay manejador
        }

        // Ejecutar purga
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

## Creación de un Nuevo Manejador de Purga

### 1. Implementar Interfaz del Manejador

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
        return int64(len(ids)), nil // Solo vista previa
    }

    deleteQuery := `DELETE FROM my_entities WHERE id = ANY($1)`
    result, err := h.db.ExecContext(ctx, deleteQuery, pq.Array(ids))
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}
```

### 2. Registrar en el Módulo

```go
// internal/modules/mymodule/module.go
func (m *MyModule) Initialize(core *Core) error {
    // Cargar configuración
    cfg := moduleconfig.Load("internal/modules/mymodule/config", env)
    retentionDays := cfg.GetRetentionDays("purge.my_entities.retention_days")

    // Registrar manejador
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

### 3. Agregar Configuración del Módulo

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

## Beneficios

### ✅ Independencia de Módulo

- Los módulos poseen completamente su lógica de purga
- Sin dependencias del núcleo en entidades de módulos
- Fácil agregar/eliminar módulos

### ✅ Claridad de Configuración

- Núcleo: Configuraciones de infraestructura
- Módulos: Políticas de negocio
- Separación clara de responsabilidades

### ✅ Testabilidad

- Los manejadores pueden probarse independientemente
- Las políticas pueden cambiarse sin cambios de código
- El patrón de registro permite simulación fácil

### ✅ Mantenibilidad

- Los cambios en políticas de retención no requieren cambios en el núcleo
- Nuevas entidades detectadas automáticamente vía registro
- Lógica de orquestación centralizada

## Monitoreo

### Eventos Publicados

- `purge.entity.completed` - Purga de entidad exitosa
- `purge.entity.failed` - Purga de entidad fallida
- `purge.all.completed` - Ciclo completo de purga completado

### Logs

```
level=info msg="Registered purge for user_posts" retention_days=90
level=info msg="Starting purge operation" entity=user_posts cutoff_date=2024-09-22
level=info msg="Purge operation completed" entity=user_posts records_purged=150 duration=2.3s
```

### Endpoints Administrativos

- `GET /api/v1/admin/purge/policies` - Listar todas las políticas de retención
- `POST /api/v1/admin/purge/preview/:entity` - Vista previa de purga para entidad
- `POST /api/v1/admin/purge/execute/:entity` - Activar purga manualmente

## Migración desde el Sistema Antiguo

**Antes** (Núcleo conocía entidades):

```go
// ❌ Núcleo tenía configuración específica de entidad
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

**Después** (Núcleo solo tiene infraestructura):

```go
// ✅ Núcleo solo tiene configuraciones de infraestructura
type PurgeConfig struct {
    Enabled   bool
    Schedule  string
    DryRun    bool
    BatchSize int
}
```

Los módulos ahora registran sus propias políticas vía `purge.DefaultPolicyRegistry`.

## Ver También

- [Independencia de Módulos](MODULE_INDEPENDENCE.es.md)
- [Desarrollo de Módulos](MODULE_DEVELOPMENT.es.md)
- [Guía de Pruebas](TESTING_GUIDE.es.md)
