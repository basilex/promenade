# Arquitectura Promenade - Referencia Rápida

[🇬🇧 English](../ARCHITECTURE_QUICKREF.md) | [🇺🇦 Українська](../uk/ARCHITECTURE_QUICKREF.uk.md) | [🇩🇪 Deutsch](../de/ARCHITECTURE_QUICKREF.de.md) | [🇵🇹 Português](../pt/ARCHITECTURE_QUICKREF.pt.md) | 🇪🇸 **Español**

## Core vs Módulos: Regla Simple

**CORE** = Infraestructura + Datos de Referencia + Auth

- Siempre habilitado, proporciona servicios

**MÓDULOS** = Lógica de Negocio

- Opcionales, licenciables, independientes

---

## ¿Qué Pertenece al Core?

### PERTENECE al Core:

1. **Servicios de Infraestructura**

   - Gestión de conexión a base de datos
   - Event Bus (adaptadores Memory/Redis)
   - Planificador (tareas cron)
   - Cargador de configuración
   - Logger
   - Servicio de email
   - Gestor JWT

2. **Base de Seguridad**

   - Autenticación de usuarios (login, registro, contraseña)
   - RBAC (roles, permisos, control de acceso)
   - Sesiones (tokens JWT)

3. **Datos de Referencia**

   - Países (145 países, códigos ISO 3166-1, regiones)
   - Monedas (124 monedas, ISO 4217, símbolos)
   - Regiones (30 regiones administrativas: estados, oblasts, provincias, Länder)
   - Ciudades (17 grandes ciudades con coordenadas, población, capitales)
   - Métodos de Pago (40+ métodos: tarjetas, billeteras, cripto, BNPL)
   - Zonas horarias (base de datos de zonas horarias IANA)
   - Idiomas (códigos ISO 639)
   - _Datos estables y raramente cambiantes, compartidos entre módulos_

4. **Interfaces de Gestión**
   - Registro de módulos
   - Registro de handlers de purga
   - Registro de políticas de purga
   - Interfaz del event bus

### NO Pertenece al Core:

- Entidades de negocio (Post, Comment, Profile, etc.)
- Casos de uso de negocio
- Handlers HTTP de negocio
- Rutas de negocio
- Configuración de negocio
- Lógica específica de entidades

**Regla general:** Si es un concepto de negocio que podría venderse por separado, es un MÓDULO.

---

## ¿Qué Pertenece a los Módulos?

### Estructura del Módulo

```
internal/modules/mymodule/
├── module.go              # Implementación del módulo
├── register.go            # Auto-registro via init()
├── config/                # Configuraciones YAML propias por entorno
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
├── entity/                # Entidades de dominio
├── usecase/               # Lógica de negocio
└── adapter/
    ├── http/              # Handlers, DTOs, rutas
    ├── repository/        # Implementaciones Postgres
    └── purge/             # Handlers de purga (si es necesario)
```

### Lista de Verificación del Módulo

- [ ] Tiene estructura propia entity/usecase/adapter
- [ ] Carga configuración propia de `config/config.*.yaml`
- [ ] Registra rutas en `RegisterRoutes()`
- [ ] Registra permisos en `RegisterPermissions()`
- [ ] Registra handlers de purga (si entidades con soft-delete)
- [ ] Sin imports de `internal/domain` o `internal/usecase`
- [ ] Usa solo paquetes `pkg/*`

---

## Flujo de Trabajo de Desarrollo de Módulo

### 1. Crear Módulo

```bash
mkdir -p internal/modules/mymodule/{config,entity,usecase,adapter/http/handler}
```

### 2. Implementar Interfaz del Módulo

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type MyModule struct {
    *module.BaseModule
    db *sqlx.DB
    // ... otros campos
}

func New() module.Module {
    return &MyModule{
        BaseModule: module.NewBaseModule(module.Metadata{
            Name:        "mymodule",
            DisplayName: "My Module",
            Version:     "1.0.0",
            Description: "Does something useful",
        }),
    }
}

func (m *MyModule) Initialize(ctx context.Context, core *module.Core) error {
    // 1. Cargar configuración del módulo
    cfg := moduleconfig.Load("internal/modules/mymodule/config", os.Getenv("ENVIRONMENT"))

    // 2. Configurar repositorios, casos de uso, handlers
    m.db = core.DB

    // 3. Registrar handlers de purga (si es necesario)
    // 4. Registrar políticas de retención (si es necesario)

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

// ... implementar otros métodos de la interfaz
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

### 4. Agregar Configuración

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

### 5. Habilitar en Configuración Principal

```yaml
# config/modules.yaml
modules:
  enabled:
    - posts
    - profiles
    - mymodule # Agregar aquí
```

### 6. Importar en main.go

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/mymodule"  // Agregar aquí
)
```

---

## Reglas de Configuración

### Configuración Core

**Archivo:** `config/app.{dev|test|prod}.yaml`

**Contiene SOLO:**

- Configuraciones de infraestructura (DB, servidor, JWT, logging)
- Configuración del event bus
- Infraestructura de purga (enabled, schedule, batch_size)
- Configuraciones CORS
- Configuraciones del servicio de email

**NO contiene:**

- Días de retención específicos de entidades → Módulos
- Configuraciones específicas de módulos → Módulos
- Configuración de lógica de negocio → Módulos

### Configuración del Módulo

**Archivo:** `internal/modules/{name}/config/config.{dev|test|prod}.yaml`

**Contiene:**

- Metadatos del módulo (nombre, versión)
- Configuraciones específicas del módulo
- Políticas de retención de purga (si aplica)
- Feature flags (si aplica)

**Cargado por:** Cada módulo via `pkg/module/config.Load()`

---

## Reglas del Sistema de Purga

### FORMA ANTIGUA (Core conoce entidades)

```go
//  INCORRECTO - Core tiene configuración específica de entidades
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

### NUEVA FORMA (Core solo orquesta)

**Configuración Core:**

```yaml
purge:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 1000
```

**Configuración Módulo:**

```yaml
purge:
  user_posts:
    retention_days: 90
    enabled: true
```

**Registro del Módulo:**

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

**Orquestación Core:**

```go
// Core obtiene TODAS las políticas del registro
policies := purge.DefaultPolicyRegistry.GetAllPolicies()

// Core crea caso de uso
useCase := usecase.NewPurgeUseCase(
    purge.DefaultRegistry,  // handlers
    policies,               // de módulos
    batchSize,
    eventBus,
)

// Core inicia planificador
scheduler.Start(ctx)
```

**Resultado:** ¡Core no sabe nada sobre `user_posts` o días de retención!

---

## Patrones de Comunicación

### Orientado a Eventos (Preferido)

```go
// Módulo A publica
event := &PostCreatedEvent{PostID: id}
eventBus.Publish(ctx, "post.created", event)

// Módulo B se suscribe
eventBus.Subscribe("post.created", func(e bus.Event) {
    // Manejar asincrónicamente
})
```

**Beneficios:** Acoplamiento débil, procesamiento asíncrono

### Registro Directo (Usar Con Moderación)

```go
// Obtener otro módulo
postsModule := core.Registry.Get("posts")

// Verificar tipo y llamar
if api, ok := postsModule.(PostsAPI); ok {
    stats := api.GetStats(userID)
}
```

**Usar solo cuando:** Se necesita respuesta síncrona, no se pueden usar eventos

---

## Errores Comunes a Evitar

### Importar Paquetes Core en Módulos

```go
//  INCORRECTO
import "github.com/basilex/promenade/internal/domain/entity"
import "github.com/basilex/promenade/internal/usecase"
```

**Corrección:** Definir entidades en el propio paquete `entity/` del módulo.

### Codificar Valores de Negocio en el Código

```go
//  INCORRECTO
const maxCommentLength = 2000
```

**Corrección:** Cargar desde la configuración del módulo.

### Colocar Lógica de Negocio en Core

```go
//  INCORRECTO - PostUseCase en internal/usecase/
```

**Corrección:** Mover al paquete `usecase/` del módulo.

### Core Conociendo Entidades de Módulos

```go
//  INCORRECTO - Core tiene días de retención para posts
type PurgeConfig struct {
    RetentionDaysUserPosts int
}
```

**Corrección:** Módulo registra política de retención via registro.

---

## Estrategia de Pruebas

### Pruebas Core

- Pruebas unitarias de servicios de infraestructura
- Pruebas de integración auth/RBAC
- Pruebas de repositorios de datos de referencia

### Pruebas de Módulos

- Pruebas unitarias de lógica de negocio (casos de uso)
- Pruebas de integración de repositorios
- Pruebas de handlers con casos de uso mock

### Pruebas Smoke

- Flujos críticos end-to-end
- Pruebas de comunicación entre módulos via eventos
- Verificar funcionamiento de rutas de módulos

---

## Comandos Rápidos

```bash
# Build
make build

# Ejecutar dev
make dev

# Ejecutar pruebas
make test

# Ejecutar pruebas de módulo específico
go test ./internal/modules/posts/...

# Ejecutar pruebas smoke
make test-smoke

# Generar boilerplate de módulo
make generate ENTITY=MyEntity

# Crear migración
make migrate-create NAME=add_my_table
```

---

## Árbol de Decisión: ¿Core o Módulo?

```
¿Es infraestructura (DB, logger, event bus)?
└─> SÍ → CORE

¿Es seguridad (auth, RBAC)?
└─> SÍ → CORE

¿Son datos de referencia (países, monedas, regiones, ciudades, métodos de pago)?
└─> SÍ → CORE

¿Es estable y usado por múltiples módulos?
└─> SÍ → Considerar CORE (o pkg compartido)

¿Es lógica de negocio?
└─> SÍ → MÓDULO

¿Puede venderse por separado?
└─> SÍ → MÓDULO

¿Es específico de entidad?
└─> SÍ → MÓDULO

¿En duda?
└─> MÓDULO (más fácil mover a core después que viceversa)
```

---

## Recursos

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.es.md) - Revisión detallada de arquitectura
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md) - Arquitectura visual
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.es.md) - Guía de desarrollo de módulos
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.es.md) - Principios de independencia
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.es.md) - Detalles del sistema de purga
- [../../internal/CORE.md](../../internal/CORE.md) - Documentación de componentes Core

---

**Recuerde:**

- Core = Infraestructura + Datos de Referencia + Auth
- Módulos = Lógica de Negocio (independientes, licenciables)
- Use registros para acoplamiento débil
- Eventos para comunicación asíncrona
- Autonomía de configuración para cada módulo
