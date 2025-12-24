# Promenade

[🇬🇧 English](README.md) | [🇺🇦 Українська](README.uk.md) | [🇩🇪 Deutsch](README.de.md) | [🇵🇹 Português](README.pt.md) | 🇪🇸 **Español**

> **📝 Nota sobre traducciones**: Algunos documentos técnicos (migrations/, internal/, pkg/, test/) y guías especializadas (SOFT*DELETE.md, LOGGING.md, VALIDATION.md, CREDENTIALS.md, MAKEFILE_ARCHITECTURE.md, MIGRATION_ARCHITECTURE.md, REDIS_BUS_TESTING.md, MOCK*\*.md) están disponibles solo en inglés por el momento. Los principales documentos de arquitectura están completamente traducidos al español. Consulte [docs/es/INDEX.es.md](docs/es/INDEX.es.md) para la lista de traducciones disponibles.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**API REST lista para producción** construida con **Clean Architecture**, **sistema modular de plugins** y **migraciones de base de datos basadas en namespace**. Presenta módulos de negocio autónomos, PostgreSQL con UUID v7, arquitectura orientada a eventos e infraestructura completa de pruebas.

---

## Visión General de la Arquitectura

Promenade sigue una **arquitectura de capas estricta** donde el **Core orquesta** y los **Módulos ejecutan** la lógica de negocio:

**Capa Core** - Orquestador + Infraestructura + Servicios Compartidos

- Autenticación & Autorización (RBAC)
- Event Bus (Memory/Redis)
- Base de Datos & Transacciones
- Logging & Configuración
- Registro de Módulos & Ciclo de Vida
- Datos de Referencia (países, monedas, regiones, ciudades, zonas horarias, idiomas, métodos de pago)

**Capa de Módulos** - Slices Verticales Independientes (Áreas de Dominio)

| Módulo        | Entidades                        | Descripción                     | Estado       |
| ------------- | -------------------------------- | ------------------------------- | ------------ |
| Posts         | posts, comments, likes           | Contenido generado por usuarios | Gratuito     |
| Profiles      | contacts, profiles               | Perfiles de usuarios            | Gratuito     |
| **Analytics** | **metrics, reports, dashboards** | **Analytics e informes**        | **Gratuito** |
| Warehouse     | products, inventory              | Gestión de inventario (futuro)  | Planificado  |

Cada módulo es autocontenido con:

- Propias entidades de dominio & lógica de negocio
- Propias migraciones de base de datos (basadas en namespace)
- Propios repositorios & casos de uso
- Propios handlers HTTP & rutas
- Propia configuración & ciclo de vida
- Opcional: Políticas de purga, permisos, eventos

### Principios Fundamentales

1. **Core como Orquestador**

   - Core **gestiona** el ciclo de vida de los módulos (init, start, stop)
   - Core **proporciona** servicios compartidos (auth, eventos, DB, logging)
   - Core **sabe CUÁNDO** llamar a módulos, pero no **CÓMO** funcionan
   - Core **nunca** importa código específico de módulos

2. **Módulos como Workers**

   - Módulos **implementan** lógica de negocio específica del dominio
   - Módulos se **registran** vía funciones `init()`
   - Módulos son **independientes** - pueden activarse/desactivarse sin afectar a otros
   - Módulos **nunca** importan código de otros módulos (solo `pkg/*`)
   - Cada módulo = slice vertical completo (entidades → handlers)

3. **Capas Clean Architecture**

   ```
   Domain (entidades, interfaces) → Use Case (lógica de negocio)
      ↓                                      ↓
   Adapter (repos, handlers) → Infrastructure (DB, HTTP, eventos)
   ```

   - **Regla de Dependencia**: Las capas internas nunca dependen de capas externas
   - Los casos de uso dependen solo de interfaces de dominio, nunca de implementaciones concretas

4. **Migraciones Basadas en Namespace**
   - Cada namespace (core, posts, profiles) tiene historial de versión independiente
   - Las migraciones están en `migrations/{namespace}/NNNNNN_descripcion.{up|down}.sql`
   - Las migraciones del Core se ejecutan primero, luego los módulos habilitados
   - Verdadera autonomía de módulos - habilitar/deshabilitar sin conflictos de esquema

---

## Inicio Rápido

### Requisitos Previos

- **Go 1.25+**
- **Docker & Docker Compose** (para PostgreSQL, Redis)
- **Make** (para automatización)

### 1. Clonar & Configurar

```bash
git clone https://github.com/basilex/promenade.git
cd promenade

# Instalar dependencias de desarrollo
make install

# Iniciar PostgreSQL + Redis vía Docker
make docker-up
```

### 2. Ejecutar Migraciones

Las migraciones se ejecutan **automáticamente** al iniciar la aplicación, pero también puede ejecutarlas manualmente:

```bash
# Verificar estado de migraciones para todos los namespaces
make migrate-status

# Ejecutar todas las migraciones (core + módulos habilitados)
make migrate

# Ejecutar namespace específico
make migrate-core                    # Solo Core
make migrate-module MODULE=posts     # Módulo específico
```

Vea [migrations/README.md](migrations/README.md) _(en inglés)_ para guía detallada de migraciones.

### 3. Iniciar Aplicación

```bash
# Modo de desarrollo (hot reload, debug logging)
make dev

# O compilar y ejecutar binario
make build
./bin/promenade
```

El servidor inicia en **http://localhost:8081**

---

## Estructura de Documentación

### Documentación Principal

| Documento                                                                      | Descripción                                                                   |
| ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| **[docs/es/ARCHITECTURE_OVERVIEW.es.md](docs/es/ARCHITECTURE_OVERVIEW.es.md)** | Diagramas visuales de arquitectura, responsabilidades de capas, ciclo de vida |
| **[docs/es/ARCHITECTURE_QUICKREF.es.md](docs/es/ARCHITECTURE_QUICKREF.es.md)** | Referencia rápida, árboles de decisión, errores comunes                       |
| **[docs/es/ARCHITECTURE_AUDIT.es.md](docs/es/ARCHITECTURE_AUDIT.es.md)**       | Auditoría de cumplimiento de arquitectura, checklist de verificación          |
| **[internal/CORE.md](internal/CORE.md)** _(en inglés)_                         | Responsabilidades y límites del Core                                          |

### Sistema de Módulos

| Documento                                                                                | Descripción                                                 |
| ---------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| **[internal/modules/README.md](internal/modules/README.md)** _(en inglés)_               | Visión general del sistema de módulos, estructura, registro |
| **[docs/es/MODULE_DEVELOPMENT.es.md](docs/es/MODULE_DEVELOPMENT.es.md)**                 | Creación de nuevos módulos, mejores prácticas               |
| **[docs/es/MODULE_INDEPENDENCE.es.md](docs/es/MODULE_INDEPENDENCE.es.md)**               | Reglas de autonomía de módulos, gestión de dependencias     |
| **[docs/es/MODULE_CONFIG_ARCHITECTURE.es.md](docs/es/MODULE_CONFIG_ARCHITECTURE.es.md)** | Sistema de configuración de módulos                         |

### Infraestructura & Sistemas

| Documento                                                                          | Descripción                                                 |
| ---------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| **[migrations/README.md](migrations/README.md)** _(en inglés)_                     | Sistema de migración basado en namespace, uso CLI           |
| **[docs/MIGRATION_ARCHITECTURE.md](docs/MIGRATION_ARCHITECTURE.md)** _(en inglés)_ | Diseño e implementación del sistema de migración            |
| **[docs/es/PURGE_ARCHITECTURE.es.md](docs/es/PURGE_ARCHITECTURE.es.md)**           | Sistema automatizado de purga de datos (basado en registro) |
| **[pkg/bus/README.md](pkg/bus/README.md)** _(en inglés)_                           | Event bus (adaptadores Memory/Redis)                        |
| **[docs/REDIS_BUS_TESTING.md](docs/REDIS_BUS_TESTING.md)** _(en inglés)_           | Probando el Redis event bus                                 |

### Guías de Desarrollo

| Documento                                                                        | Descripción                                 |
| -------------------------------------------------------------------------------- | ------------------------------------------- |
| **[docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md)** _(en inglés)_ | Sistema Makefile (comandos dev, test, prod) |
| **[test/README.md](test/README.md)** _(en inglés)_                               | Infraestructura de pruebas (200+ pruebas)   |
| **[docs/es/TESTING_GUIDE.es.md](docs/es/TESTING_GUIDE.es.md)**                   | Mejores prácticas de pruebas, patrones      |
| **[docs/es/TESTING_INFRASTRUCTURE.es.md](docs/es/TESTING_INFRASTRUCTURE.es.md)** | Configuración de infraestructura de pruebas |

### Referencias Técnicas

| Documento                                                      | Descripción                                               |
| -------------------------------------------------------------- | --------------------------------------------------------- |
| **[docs/es/UUID_V7_GUIDE.es.md](docs/es/UUID_V7_GUIDE.es.md)** | Implementación y beneficios del UUID v7                   |
| **[docs/SOFT_DELETE.md](docs/SOFT_DELETE.md)**                 | Patrón soft delete para contenido de usuarios             |
| **[docs/es/AUTHORIZATION.es.md](docs/es/AUTHORIZATION.es.md)** | Sistema RBAC (4 roles, permisos wildcard)                 |
| **[docs/LOGGING.md](docs/LOGGING.md)**                         | Logging estructurado con slog                             |
| **[docs/VALIDATION.md](docs/VALIDATION.md)**                   | Patrones de validación de peticiones                      |
| **[docs/CREDENTIALS.md](docs/CREDENTIALS.md)**                 | Usuarios de prueba por defecto y credenciales             |
| **[docs/es/INDEX.es.md](docs/es/INDEX.es.md)**                 | Índice completo de documentación con rutas de aprendizaje |

---

## Sistema de Módulos

### Módulos Disponibles

#### **Módulo Posts** (`internal/modules/posts`)

Gestión de contenido generado por usuarios:

- **Entidades**: Posts, Comments, Likes
- **Funcionalidades**: Crear/editar posts, hilos de comentarios, sistema de likes
- **Migraciones**: 3 migraciones (namespace: `posts`)
- **Configuración**: `config/modules.yaml` → `posts`

**Documentación completa**: [internal/modules/posts/README.md](internal/modules/posts/README.md) _(en inglés)_

#### **Módulo Profiles** (`internal/modules/profiles`)

Gestión de perfiles de usuarios y contactos:

- **Entidades**: UserProfiles, UserContacts
- **Funcionalidades**: Gestión de perfiles, información de contacto
- **Migraciones**: 2 migraciones (namespace: `profiles`)
- **Configuración**: `config/modules.yaml` → `profiles`

**Documentación completa**: [internal/modules/profiles/README.md](internal/modules/profiles/README.md) _(en inglés)_

#### **Módulo Analytics** (`internal/modules/analytics`) - Gratuito

Analytics, métricas e informes:

- **Estado**: Gratuito - Disponible para todos los usuarios
- **Entidades**: Metrics, Reports, Dashboards
- **Funcionalidades**: Recolección de métricas, informes personalizados, dashboards visuales
- **Migraciones**: 1 migración (namespace: `analytics`)
- **Caso de Uso**: Business intelligence, monitoreo de rendimiento, insights de datos

**Documentación completa**: [internal/modules/analytics/README.md](internal/modules/analytics/README.md) _(en inglés)_

#### **Módulo Warehouse** (`internal/modules/warehouse`) - Módulo Futuro

Gestión de inventario y productos (planificado):

- **Estado**: Planificado - La estructura existe como placeholder, aún no implementado
- **Caso de Uso**: E-commerce, sistemas de inventario, retail

**Documentación planificada**: [internal/modules/warehouse/README.md](internal/modules/warehouse/README.md) _(en inglés)_

### Estructura de Módulo

Cada módulo sigue una estructura consistente:

```
internal/modules/{module}/
├── module.go           # Registro de módulo & ciclo de vida
├── domain/
│   └── entity/         # Entidades de dominio
├── repository/         # Interfaces de acceso a datos & implementaciones
├── usecase/            # Lógica de negocio
├── adapter/
│   └── handler/        # Handlers HTTP & DTOs
└── README.md           # Documentación específica del módulo
```

### Habilitando/Deshabilitando Módulos

Edite `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts # Contenido generado por usuarios
    - profiles # Perfiles de usuarios + contactos
    - analytics # Business analytics (requiere licencia)
    # - warehouse  # Futuro: Gestión de inventario
```

Los módulos se cargan automáticamente al iniciar la aplicación.

---

## Migraciones de Base de Datos

### Sistema Basado en Namespace

Cada namespace mantiene **historial de versión independiente**:

```
migrations/
├── core/               # Infraestructura core (siempre ejecuta primero)
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000003_core_rbac_full.up.sql
│   ├── 000004_core_ref_timezones.up.sql
│   ├── 000005_core_ref_languages.up.sql
│   ├── 000006_core_ref_countries_currencies.up.sql    # 145 países, 124 monedas
│   ├── 000007_core_ref_regions_cities.up.sql          # 30 regiones, 17 ciudades
│   └── 000008_core_ref_payment_methods.up.sql         # 40+ métodos de pago
├── posts/              # Migraciones del módulo Posts
│   ├── 000001_posts_posts.up.sql
│   ├── 000002_posts_comments.up.sql
│   └── 000003_posts_comment_likes.up.sql
├── profiles/           # Migraciones del módulo Profiles
│   ├── 000001_profiles_contacts.up.sql
│   └── 000002_profiles_profiles.up.sql
└── analytics/          # Migraciones del módulo Analytics (comercial)
    └── 000001_analytics_tables.up.sql
```

### Comandos de Migración

```bash
# Estado para todos los namespaces
make migrate-status

# Ejecutar todas (core + módulos habilitados)
make migrate

# Ejecutar namespace específico
make migrate-core
make migrate-module MODULE=posts

# Rollback
make migrate-rollback MODULE=posts STEPS=1

# Crear nueva migración
make migrate-create MODULE=posts NAME=add_post_views
make migrate-create-core NAME=add_audit_log
```

**Auto-migraciones**: Las migraciones se ejecutan automáticamente al iniciar la aplicación (core primero, luego módulos habilitados).

**Guía completa**: [migrations/README.md](migrations/README.md) _(en inglés)_

---

## Autenticación & Autorización

### Usuarios de Prueba por Defecto

| Email                           | Contraseña | Rol       | Permisos                      |
| ------------------------------- | ---------- | --------- | ----------------------------- |
| `system@promenade.com`          | `passw0rd` | Admin     | Acceso completo (`*`)         |
| `admin@promenade.com`           | `passw0rd` | Admin     | Gestión de usuarios/contenido |
| `moderator@promenade.com`       | `passw0rd` | Moderator | Moderación de contenido       |
| `alexander.vasilenko@gmail.com` | `03041965` | User      | Operaciones básicas           |

**¡Cambie las contraseñas antes del despliegue en producción!**

### Sistema RBAC

- **4 Roles del Sistema**: Admin, Moderator, User, Guest
- **Permisos Wildcard**: `posts:*` (todas las acciones de posts), `*` (acceso completo)
- **Formato Recurso-Acción**: `posts:create`, `users:delete`, `comments:moderate`

**Guía completa RBAC**: [docs/es/AUTHORIZATION.es.md](docs/es/AUTHORIZATION.es.md)

---

## Pruebas

**388 pruebas** en todas las capas (100% aprobadas, ~41 segundos):

```bash
# Ejecutar todas las pruebas (unit + integration + smoke)
make test               # Todas las pruebas (~41s)

# Ejecutar por tipo
make test-unit          # Solo pruebas unitarias (183 pruebas, ~5s)
make test-integration   # Pruebas de integración (91 pruebas, ~36s)
make test-smoke         # Smoke tests (114 pruebas, ~4s)

# Informe de cobertura
make test-coverage
```

### Estructura de Pruebas

- **Pruebas Core**: Entidades de dominio (Country, Currency, Language, Timezone, Permission, Role, User, Session, políticas de Purga)
- **Core Use Cases**: Auth, RBAC, CRUD de datos de referencia, operaciones de Purga
- **Pruebas de Módulos**: Posts (Comment, Post, PostStatus), Profiles (UserContact, UserProfile, ContactType, Gender), Analytics (Metrics, Reports, validación de licencia)
- **Pruebas de Integración**: Operaciones de repositorio con PostgreSQL real en puerto 5433
- **Helpers de Prueba**: `test/helpers/` y `test/integration/` para fixtures, configuración de base de datos, gestión de transacciones

**Guías de pruebas**:

- [test/README.md](test/README.md) _(en inglés)_ - Infraestructura de pruebas
- [docs/es/TESTING_GUIDE.es.md](docs/es/TESTING_GUIDE.es.md) - Mejores prácticas

---

## Event Bus

**Event bus de adaptador dual** para operaciones asíncronas:

### Adaptador Memory

- Pub/Sub en memoria (goroutines + channels)
- **Caso de uso**: Desarrollo, pruebas, despliegues de instancia única
- **Ventajas**: Cero dependencias, rápido, simple
- **Configuración**: `BUS_ADAPTER=memory` (por defecto)

### Adaptador Redis

- Pub/Sub distribuido vía Redis
- **Caso de uso**: Despliegues de producción multi-instancia
- **Ventajas**: Persistente, escalable, tolerante a fallos
- **Configuración**: `BUS_ADAPTER=redis` + configuraciones de conexión Redis
- **Fallback**: Auto-fallback a memory si Redis no disponible

### Ejemplo de Uso

```go
// Publicar evento
event := &UserRegisteredEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    UserID:    user.ID,
    Email:     user.Email,
}
eventBus.Publish(ctx, bus.TopicUserRegistered, event)

// Suscribirse a eventos
eventBus.Subscribe(ctx, bus.TopicUserRegistered, func(ctx context.Context, e bus.Event) error {
    evt := e.(*UserRegisteredEvent)
    // Enviar email de bienvenida
    return emailService.SendWelcome(ctx, evt.Email)
})
```

**Guía completa**: [pkg/bus/README.md](pkg/bus/README.md) _(en inglés)_

---

## Comandos Makefile

### Desarrollo

```bash
make dev                # Iniciar servidor dev (hot reload)
make build              # Compilar binario de producción
make run                # Ejecutar binario compilado
make lint               # Ejecutar linter (golangci-lint)
make fmt                # Formatear código
make config-show        # Mostrar configuración YAML (use ENV=dev|test|prod)
```

### Pruebas

```bash
make test                      # Todas las pruebas (core + módulos)
make test-core                 # Solo pruebas core (domain + usecase)
make test-modules              # Todas las pruebas de módulos
make test-module-posts         # Pruebas del módulo Posts
make test-module-profiles      # Pruebas del módulo Profiles
make test-coverage             # Generar informe de cobertura HTML
```

### Base de Datos

```bash
make migrate                   # Ejecutar todas las migraciones (core + módulos habilitados)
make migrate-status            # Mostrar estado de migración
make migrate-core              # Migrar solo core
make migrate-module MODULE=posts          # Migrar módulo específico
make migrate-rollback MODULE=posts STEPS=1  # Rollback
make migrate-create MODULE=posts NAME=xxx  # Crear migración de módulo
make migrate-create-core NAME=xxx          # Crear migración core
```

### Docker

```bash
make docker-up          # Iniciar PostgreSQL + Redis
make docker-down        # Detener servicios
make docker-clean       # Eliminar containers + volumes
make docker-logs        # Ver logs
```

### Swagger

```bash
make swagger-all        # Generar documentación API (v1 + v2)
make swagger-v1         # Generar solo docs v1
make swagger-v2         # Generar solo docs v2
```

**Guía completa Makefile**: [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md)

---

## Docker

### Configuración de Desarrollo

```bash
# Iniciar servicios
make docker-up

# Ver logs
make docker-logs

# Detener servicios
make docker-down

# Limpiar todo (eliminar volumes)
make docker-clean
```

### Servicios

- **PostgreSQL 16**: Puerto 5432, usuario `system`, base de datos `promenade_dev`
- **Redis 7**: Puerto 6379 (para event bus distribuido)

**Guía Docker**: [docker/README.md](docker/README.md) _(en inglés)_

---

## Funcionalidades Técnicas Principales

### Claves Primarias UUID v7

UUIDs ordenados por tiempo para **inserciones 2x más rápidas** que UUID v4 y mejor rendimiento de B-tree.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    ...
);
```

[docs/es/UUID_V7_GUIDE.es.md](docs/es/UUID_V7_GUIDE.es.md)

### Patrón Soft Delete

El contenido generado por usuarios (posts, comentarios) usa timestamp `deleted_at` para eliminación segura.

```go
// Siempre filtrar registros soft-deleted
WHERE deleted_at IS NULL
```

[docs/SOFT_DELETE.md](docs/SOFT_DELETE.md)

### Sistema Automatizado de Purga

Sistema basado en registro donde los módulos registran sus políticas de retención:

```go
purge.DefaultPolicyRegistry.RegisterPolicy(purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
})
```

El Scheduler del Core ejecuta jobs de purga vía cron. El Core no conoce entidades específicas.

[docs/es/PURGE_ARCHITECTURE.es.md](docs/es/PURGE_ARCHITECTURE.es.md)

### Logging Estructurado

Logging con contexto usando `slog`:

```go
log := logger.FromContext(ctx)  // Incluye request_id, user_id
log.Info("User registered", "email", user.Email)
```

[docs/LOGGING.md](docs/LOGGING.md)

---

## Documentación de la API

### Swagger UI

- **v1 API**: http://localhost:8081/api/v1/docs/swagger/index.html
- **v2 API**: http://localhost:8081/api/v2/docs/swagger/index.html

### Health Check

```bash
curl http://localhost:8081/api/v1/health
```

Respuesta:

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

### Versionado de API

- **v1**: API estable actual (`internal/adapter/http/v1`)
- **v2**: API de próxima generación (`internal/adapter/http/v2`)

Ambas versiones tienen:

- Handlers y DTOs aislados
- Documentación Swagger separada
- Registro de rutas independiente

---

## Estructura del Proyecto

```
promenade/
├── cmd/
│   ├── api/                    # Punto de entrada principal de la aplicación
│   └── migrate/                # Herramienta CLI de migración
├── internal/
│   ├── domain/                 # Dominio core (entidades, interfaces)
│   │   ├── entity/             # Entidades de dominio (User, Session)
│   │   ├── event/              # Eventos de dominio (UserRegistered, etc.)
│   │   └── repository/         # Interfaces de repositorio
│   ├── usecase/                # Casos de uso core (auth, RBAC)
│   ├── adapter/                # Adaptadores (HTTP, repositorios)
│   │   ├── http/
│   │   │   ├── v1/             # API v1 (handlers, DTOs, rutas)
│   │   │   └── v2/             # API v2
│   │   └── repository/postgres/ # Implementaciones PostgreSQL
│   ├── infrastructure/         # Infraestructura (DB, config, scheduler)
│   │   ├── database/
│   │   ├── config/
│   │   ├── notification/
│   │   └── scheduler/
│   └── modules/                # Módulos de negocio (plugins)
│       ├── posts/              # Posts + comentarios + likes
│       ├── profiles/           # Perfiles de usuarios + contactos
│       ├── analytics/          # Analytics + informes (Comercial, activo)
│       └── warehouse/          # Gestión de inventario (futuro)
├── pkg/                        # Paquetes compartidos (reutilizables)
│   ├── bus/                    # Event bus (memory/redis)
│   ├── jwt/                    # Gestor JWT
│   ├── logger/                 # Logger estructurado
│   ├── migration/              # Gestor de migraciones
│   ├── module/                 # Registro de módulos
│   ├── purge/                  # Registro de purga
│   ├── response/               # Helpers de respuesta HTTP
│   ├── uuidv7/                 # Generador UUID v7
│   └── validator/              # Validación de peticiones
├── migrations/                 # Migraciones basadas en namespace
│   ├── core/                   # Migraciones core (auth, RBAC, datos ref)
│   ├── posts/                  # Migraciones del módulo Posts
│   └── profiles/               # Migraciones del módulo Profiles
├── test/                       # Infraestructura de pruebas
│   ├── helpers/                # Helpers de prueba (fixtures, configuración DB)
│   ├── integration/            # Pruebas de integración
│   ├── smoke/                  # Smoke tests
│   └── mocks/                  # Implementaciones mock
├── config/                     # Archivos de configuración
│   ├── app.dev.yaml            # Configuración entorno dev
│   ├── app.test.yaml           # Configuración entorno test
│   ├── app.prod.yaml           # Configuración producción
│   └── modules.yaml            # Habilitar/deshabilitar módulo + configuraciones
├── docs/                       # Documentación
├── scripts/                    # Scripts auxiliares
├── templates/                  # Plantillas de email
└── docker/                     # Configuraciones Docker
```

---

## Configuración

### Configuraciones Específicas de Entorno

Prioridad de configuración (YAML primero, `.env` como fallback):

1. `config/app.{dev|test|prod}.yaml` - Configuraciones de infraestructura core
2. `config/modules.yaml` - Habilitar/deshabilitar módulo + configuraciones específicas de módulo
3. `.env.{env}.local` / `.env.{env}` / `.env` - Soporte legacy

### Ejemplo: `config/app.dev.yaml`

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
  adapter: "memory" # o "redis"
  worker_pool_size: 4

purge:
  enabled: true
  schedule: "0 2 * * *" # Diariamente a las 2 AM
  batch_size: 1000
```

### Ejemplo: `config/modules.yaml`

```yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics # Módulo comercial (requiere licencia)
    # - warehouse  # Futuro: Gestión de inventario

  config:
    posts:
      version: "1.0.0"
      settings:
        max_post_length: 10000
        max_comment_depth: 10

    analytics:
      version: "1.0.0"
      license_key: "" # Definir vía variable de entorno ANALYTICS_LICENSE_KEY
      settings:
        metrics_retention_days: 90
```

**Guía de configuración**: [docs/es/MODULE_CONFIG_ARCHITECTURE.es.md](docs/es/MODULE_CONFIG_ARCHITECTURE.es.md)

---

## Creando Nuevos Módulos

### Paso 1: Crear Estructura del Módulo

```bash
mkdir -p internal/modules/mymodule/{domain/entity,repository,usecase,adapter/handler}
```

### Paso 2: Implementar Interfaz de Módulo

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type Module struct{}

func (m *Module) Name() string { return "mymodule" }

func (m *Module) Initialize(ctx context.Context, core module.Core) error {
    // Registrar rutas, permisos, handlers de purga
    return nil
}

func (m *Module) Start(ctx context.Context) error {
    // Iniciar workers en background
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

### Paso 3: Crear Migraciones

```bash
make migrate-create MODULE=mymodule NAME=create_tables
```

### Paso 4: Habilitar Módulo

Añadir a `config/modules.yaml`:

```yaml
modules:
  enabled:
    - mymodule
```

**Guía completa**: [docs/es/MODULE_DEVELOPMENT.es.md](docs/es/MODULE_DEVELOPMENT.es.md)

---

## Rutas de Aprendizaje

### Para Nuevos Desarrolladores

1. **Inicio**: [docs/es/ARCHITECTURE_QUICKREF.es.md](docs/es/ARCHITECTURE_QUICKREF.es.md) - Visión general de 15 minutos
2. **Conceptos Core**: [internal/CORE.md](internal/CORE.md) _(en inglés)_ - Responsabilidades del Core
3. **Sistema de Módulos**: [internal/modules/README.md](internal/modules/README.md) _(en inglés)_
4. **Práctica**: Crear un módulo simple siguiendo [docs/es/MODULE_DEVELOPMENT.es.md](docs/es/MODULE_DEVELOPMENT.es.md)

### Para DevOps/Despliegue

1. **Makefile**: [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md)
2. **Migraciones**: [migrations/README.md](migrations/README.md) _(en inglés)_
3. **Docker**: [docker/README.md](docker/README.md) _(en inglés)_
4. **Configuración**: [docs/es/MODULE_CONFIG_ARCHITECTURE.es.md](docs/es/MODULE_CONFIG_ARCHITECTURE.es.md)

### Para Arquitectos

1. **Visión General de Arquitectura**: [docs/es/ARCHITECTURE_OVERVIEW.es.md](docs/es/ARCHITECTURE_OVERVIEW.es.md)
2. **Auditoría & Verificación**: [docs/es/ARCHITECTURE_AUDIT.es.md](docs/es/ARCHITECTURE_AUDIT.es.md)
3. **Independencia de Módulos**: [docs/es/MODULE_INDEPENDENCE.es.md](docs/es/MODULE_INDEPENDENCE.es.md)
4. **Sistema de Migración**: [docs/MIGRATION_ARCHITECTURE.md](docs/MIGRATION_ARCHITECTURE.md)

**Índice completo**: [docs/es/INDEX.es.md](docs/es/INDEX.es.md)

---

## Contribuyendo

1. Hacer fork del repositorio
2. Crear branch de feature (`git checkout -b feature/amazing-feature`)
3. Seguir principios de arquitectura (vea [docs/es/ARCHITECTURE_QUICKREF.es.md](docs/es/ARCHITECTURE_QUICKREF.es.md))
4. Escribir pruebas (mantener 100% de tasa de aprobación)
5. Hacer commit de cambios (`git commit -m 'Add amazing feature'`)
6. Push a branch (`git push origin feature/amazing-feature`)
7. Abrir Pull Request

---

## Licencia

Este proyecto está licenciado bajo la Licencia MIT - vea el archivo [LICENSE](LICENSE) para detalles.

---

## Soporte

- **Documentación**: [docs/es/INDEX.es.md](docs/es/INDEX.es.md)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Email**: alexander.vasilenko@gmail.com

---

**Construido con Clean Architecture y Go**
