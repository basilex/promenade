# Visión General de la Arquitectura Promenade

[🇬🇧 English](ARCHITECTURE_OVERVIEW.es.md) | [🇺🇦 Українська](../uk/ARCHITECTURE_OVERVIEW.uk.md) | [🇩🇪 Deutsch](../de/ARCHITECTURE_OVERVIEW.de.md) | [🇵🇹 Português](../pt/ARCHITECTURE_OVERVIEW.pt.md) | 🇪🇸 **Español**

Este documento proporciona una visión general de alto nivel de la arquitectura Promenade, organizada en torno a **principios de Clean Architecture** y un **sistema de módulos basado en plugins**.

---

## Capas de la Arquitectura

### 1. Capa de Aplicación

**Ubicación:** `cmd/api/main.go`

**Responsabilidades:**

- Inicializar infraestructura (DB, EventBus, Config, Logger)
- Cargar configuración principal
- Inicializar sistema de módulos (descubrimiento, registro, ciclo de vida)
- Iniciar servidor HTTP

---

### 2. Infraestructura Principal

**Ubicación:** `internal/infrastructure/`

**Componentes:**

| Componente    | Propósito                             |
| ------------- | ------------------------------------- |
| Database      | Gestión de conexión PostgreSQL        |
| Event Bus     | Adaptadores Memory/Redis Pub/Sub      |
| Scheduler     | Programación de tareas basada en cron |
| Config        | Cargador de configuración YAML        |
| Logger        | Logging estructurado con slog         |
| Email Service | Envío asíncrono de correo electrónico |

**Proporciona:** Servicios compartidos para todos los módulos (DB, EventBus, Config, JWT, Logger)

---

### 3. Dominio Principal

**Ubicación:** `internal/domain/entity/`

**Seguridad & Autenticación** (siempre habilitado):

- User (solo autenticación: email, contraseña, roles)
- Session (tokens JWT)
- Role (roles RBAC)
- Permission (formato resource:action)

**Datos de Referencia** (estables, compartidos):

- Country (195+ códigos ISO, regiones)
- Currency (170+ códigos ISO 4217)
- Timezone (500+ zonas horarias IANA)
- Language (180+ códigos ISO 639)

**Propósito:** Entidades principales requeridas por TODOS los módulos. Sin lógica de negocio.

---

### 4. Casos de Uso Principales

**Ubicación:** `internal/usecase/`

**Casos de Uso Disponibles:**

| Caso de Uso        | Propósito                           |
| ------------------ | ----------------------------------- |
| Auth UseCase       | Registro, login, logout             |
| Role UseCase       | CRUD de roles, asignación de roles  |
| Permission UseCase | CRUD de permisos, control de acceso |
| Country UseCase    | Listar países, obtener por código   |
| Currency UseCase   | Listar monedas, obtener por código  |
| Purge UseCase      | Orquestar jobs de limpieza          |

**Nota:** Los casos de uso principales NO contienen lógica de negocio - solo infraestructura y seguridad.

---

### 5. Rutas API Principales

**Ubicación:** `internal/adapter/http/v1/`

**Endpoints:**

| Ruta                   | Propósito                     |
| ---------------------- | ----------------------------- |
| /api/v1/health         | Verificaciones de salud       |
| /api/v1/auth/\*        | Login, registro, logout       |
| /api/v1/countries/\*   | Datos de referencia           |
| /api/v1/currencies/\*  | Datos de referencia           |
| /api/v1/roles/\*       | Gestión RBAC                  |
| /api/v1/permissions/\* | Gestión RBAC                  |
| /api/v1/admin/\*       | Limpieza, gestión del sistema |

---

## Sistema de Módulos

### Registro & Gestor de Módulos

**Ubicación:** `pkg/module/`

**Propósito:** Orquesta ciclo de vida de módulos

**Características del Registro:**

- `Register(module)` - Auto-registro vía `init()`
- `GetEnabled(config)` - Filtrar módulos habilitados de la config
- `InitializeAll()` - Inicializar en orden de dependencia
- `StartAll()` - Iniciar workers en background
- `StopAll()` - Apagado gracioso

**Proporciona a los Módulos:**

- Conexión DB compartida
- EventBus compartido
- Gestor JWT compartido
- Cargador Config compartido

---

### Módulo: Posts

**Ubicación:** `internal/modules/posts/`

**Estado:** Habilitado (Gratuito)

**Entidades:** Post, Comment, Like

**Características:**

- Crear, actualizar, eliminar posts
- Comentarios en hilo (profundidad máxima configurable)
- Me gusta en posts y comentarios
- Soft delete con retención configurable

**Configuración:**

- max_content_length: 10000
- comments.max_depth: 10
- purge.user_posts.retention_days: 90
- purge.post_comments.retention_days: 30

**Rutas:** /api/v1/posts/_, /api/v1/comments/_, /api/v1/likes/\*

---

### Módulo: Profiles

**Ubicación:** `internal/modules/profiles/`

**Estado:** Habilitado (Gratuito)

**Entidades:** Profile, Contact

**Características:**

- Gestión de perfil de usuario
- Información de contacto (email, teléfono, etc.)
- Verificación de contacto
- Designación de contacto principal

**Configuración:**

- profiles.max_per_user: 1
- contacts.max_per_user: 5
- contacts.verification_required: true

**Rutas:** /api/v1/profiles/_, /api/v1/contacts/_

---

### Módulo: Warehouse

**Ubicación:** `internal/modules/warehouse/`

**Estado:** Deshabilitado (Comercial - requiere licencia)

**Características** (cuando está licenciado):

- Gestión de inventario
- Seguimiento de stock
- Escaneo de código de barras
- Ubicaciones de almacén

**Rutas:** /api/v1/warehouse/\* (cuando habilitado)

---

## Comunicación Inter-Módulos

### Event Bus

**Ubicación:** `pkg/bus/`

**Adaptadores:**

| Adaptador | Caso de Uso                     | Características          |
| --------- | ------------------------------- | ------------------------ |
| Memory    | Dev/Test/Instancia única        | Rápido, sin dependencias |
| Redis     | Production/Múltiples instancias | Distribuido, persistente |

**Flujo de Eventos:**

1. Módulo A publica evento al EventBus
2. EventBus distribuye a todos los suscriptores
3. Módulos B, C, D procesan evento asíncronamente

**Ejemplos:**

- user.registered → enviar email de bienvenida (asíncrono)
- post.created → actualizar estadísticas del usuario (asíncrono)
- purge.completed → registrar en auditoría (asíncrono)

---

## Arquitectura de Configuración

### Configuración Principal

```
config/
├── app.dev.yaml     - Infraestructura principal (dev)
├── app.test.yaml    - Infraestructura principal (test)
├── app.prod.yaml    - Infraestructura principal (prod)
└── modules.yaml     - Qué módulos cargar
```

### Configuración de Módulos

```
internal/modules/posts/config/
├── config.dev.yaml  - Configuración del módulo Posts (dev)
├── config.test.yaml - Configuración del módulo Posts (test)
└── config.prod.yaml - Configuración del módulo Posts (prod)

internal/modules/profiles/config/
├── config.dev.yaml  - Configuración del módulo Profiles (dev)
├── config.test.yaml - Configuración del módulo Profiles (test)
└── config.prod.yaml - Configuración del módulo Profiles (prod)
```

**Sustituciones de Entorno:** `.env.example` (opcional)

---

## Ciclo de Vida del Módulo

### 1. Auto-Registro (vía init())

El módulo se registra a sí mismo en la importación del paquete:

```go
package posts

func init() {
    module.DefaultRegistry.Register(New())
}
```

### 2. Descubrimiento & Filtrado

- Leer `config/modules.yaml`
- Filtrar módulos habilitados
- Resolver dependencias (ordenación topológica)

### 3. Inicialización (en orden de dependencia)

Para cada módulo:

- Cargar config del módulo desde `config/config.*.yaml`
- Configurar repositorios
- Configurar casos de uso
- Configurar handlers
- Registrar handlers de limpieza
- Registrar políticas de retención
- Registrar permisos

### 4. Registro de Rutas

Para cada módulo:

- Montar rutas del módulo en el router
- Aplicar middleware (auth, RBAC, etc.)

### 5. Suscripción a Eventos

Para cada módulo:

- Suscribirse a eventos relevantes
- Configurar handlers de eventos asíncronos

### 6. Inicio (workers en background)

Para cada módulo:

- Iniciar jobs cron
- Iniciar workers en background

### 7. Tiempo de Ejecución

- Módulos procesan peticiones HTTP
- Publican/suscriben a eventos
- Ejecutan tareas programadas

### 8. Apagado (en SIGTERM/SIGINT)

Para cada módulo (orden inverso):

- Detener workers graciosamente
- Cerrar conexiones
- Limpiar recursos

---

## Principios de Diseño Clave

### 1. CORE = INFRAESTRUCTURA + DATOS DE REFERENCIA

- Core proporciona servicios (DB, EventBus, Config, JWT)
- Core contiene datos de referencia estables (países, monedas)
- Core gestiona seguridad (auth, RBAC)
- Core NO contiene lógica de negocio

### 2. MÓDULOS = LÓGICA DE NEGOCIO

- Los módulos son slices verticales autocontenidos
- Los módulos poseen sus entidades, casos de uso, adaptadores
- Los módulos registran handlers, políticas, permisos
- Los módulos pueden habilitarse/deshabilitarse vía config
- Los módulos NO importan de `internal/domain` o `internal/usecase`
- Los módulos NO dependen unos de otros directamente (usan eventos)

### 3. ARQUITECTURA DE PLUGIN

- Carga dinámica vía registro de módulos
- Resolución de dependencias (ordenación topológica)
- Gestión de ciclo de vida (init → start → stop)
- Auto-registro vía `init()`

### 4. AUTONOMÍA DE CONFIGURACIÓN

- Core: `config/app.*.yaml` (solo infraestructura)
- Módulos: `internal/modules/{name}/config/config.*.yaml`
- Cada módulo carga su propia config
- Configs específicas de entorno (dev, test, prod)

### 5. ACOPLAMIENTO DÉBIL

- Comunicación basada en eventos (pub/sub)
- Patrón de registro (handlers, políticas, permisos)
- Dependencias basadas en interfaces
- Sin imports directos módulo-a-módulo

### 6. SOPORTE DE LICENCIAMIENTO

- Claves de licencia por módulo
- Validación de licencia en `Initialize()`
- Degradación graciosa si licencia inválida

---

## Beneficios

### Modularidad

- Agregar nuevos módulos sin tocar el core
- Eliminar módulos sin romper otros
- Probar módulos independientemente

### Escalabilidad

- Módulos comerciales (warehouse, fleet, finance)
- Habilitación de características basada en licencia
- Fácil agregar nuevas verticales

### Mantenibilidad

- Límites claros (core vs módulos)
- Responsabilidad única (cada módulo posee su dominio)
- Claridad de configuración (sin config monolítica)

### Testeabilidad

- Prueba unitaria de módulos aisladamente
- Prueba de integración con core real/mock
- Prueba smoke de flujos críticos

### Desplegabilidad

- Habilitar solo módulos necesarios por despliegue
- Prueba A/B de nuevos módulos
- Rollout gradual de características

---

## Estado Actual

### Core

- Infraestructura (DB, EventBus, Scheduler, Config, Logger)
- Seguridad (Auth, RBAC, JWT, Sessions)
- Datos de Referencia (Countries, Currencies, Timezones, Languages)
- Gestión de Módulos (Registry, Lifecycle, Config Loader)
- Orquestación de Limpieza (basada en registro, sin conocimiento de entidades)

### Módulos

- **Posts** (posts + comments + likes) - Completamente independiente
- **Profiles** (profiles + contacts) - Completamente independiente
- **Warehouse** (inventory management) - Comercial, deshabilitado

### Conformidad de Arquitectura

- Core contiene SOLO infraestructura + datos de referencia
- Módulos son COMPLETAMENTE independientes (sin imports del core)
- Configuración es AUTÓNOMA (cada módulo posee config)
- Licenciamiento es SOPORTADO (listo para módulos comerciales)

### Mejoras Recientes

- Sistema de limpieza refactorizado (core = orquestador, módulos = workers)
- Profiles/Contacts migrados a módulo (eliminados del core)
- 15.000+ líneas de código eliminadas del core
- Independencia completa de módulos lograda

---

## Documentación Relacionada

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md) - Auditoría de conformidad de arquitectura
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md) - Guía de referencia rápida
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md) - Creando nuevos módulos
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md) - Principios de independencia de módulos
- [../internal/CORE.md](../internal/CORE.md) - Detalles de componentes principales
