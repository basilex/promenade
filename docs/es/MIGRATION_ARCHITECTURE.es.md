# Arquitectura de Migraciones de Módulos

[🇬🇧 English](../MIGRATION_ARCHITECTURE.md) | [🇺🇦 Українська](../uk/MIGRATION_ARCHITECTURE.uk.md) | [🇩🇪 Deutsch](../de/MIGRATION_ARCHITECTURE.de.md) | [🇵🇹 Português](../pt/MIGRATION_ARCHITECTURE.pt.md) | 🇪🇸 **Español**

---

## Problema

El sistema actual de migraciones viola la independencia de módulos:

- Todas las migraciones en una sola carpeta `migrations/`
- Numeración secuencial global (000001, 000002, ...)
- Migraciones del núcleo mezcladas con migraciones de módulos
- No hay forma de habilitar/deshabilitar migraciones de módulos de forma independiente

## Solución: Migraciones Basadas en Espacios de Nombres

### 1. Estructura de Directorios

```
migrations/
├── core/                          # Migraciones de infraestructura del núcleo
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_auth_tables.up.sql
│   ├── 000002_auth_tables.down.sql
│   └── ...
│
├── posts/                         # Migraciones del módulo posts
│   ├── 000001_create_posts.up.sql
│   ├── 000001_create_posts.down.sql
│   ├── 000002_create_comments.up.sql
│   └── ...
│
└── profiles/                      # Migraciones del módulo profiles
    ├── 000001_create_profiles.up.sql
    └── ...
```

**Cada espacio de nombres tiene versionado independiente:**

- Core: 1, 2, 3, 4, ...
- Posts: 1, 2, 3, ...
- Profiles: 1, 2, ...

---

### 2. Esquema de Tabla de Migraciones

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     BIGINT       NOT NULL,
    namespace   VARCHAR(50)  NOT NULL,  -- 'core', 'posts', 'profiles', etc.
    dirty       BOOLEAN      NOT NULL DEFAULT FALSE,
    applied_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (namespace, version)
);

CREATE INDEX idx_schema_migrations_namespace ON schema_migrations(namespace);
```

**Datos de ejemplo:**

```
| version | namespace | dirty | applied_at          |
|---------|-----------|-------|---------------------|
| 1       | core      | false | 2024-12-22 10:00:00 |
| 2       | core      | false | 2024-12-22 10:00:01 |
| 3       | core      | false | 2024-12-22 10:00:02 |
| 1       | posts     | false | 2024-12-22 10:00:03 |
| 2       | posts     | false | 2024-12-22 10:00:04 |
| 1       | profiles  | false | 2024-12-22 10:00:05 |
```

---

### 3. Gestor de Migraciones (pkg/migration/)

#### Interfaz

```go
package migration

type Manager interface {
    // MigrateNamespace aplica todas las migraciones pendientes para un espacio de nombres
    MigrateNamespace(ctx context.Context, namespace string) error

    // MigrateAll aplica todas las migraciones pendientes (núcleo + módulos habilitados)
    MigrateAll(ctx context.Context, enabledModules []string) error

    // Rollback revierte N migraciones para un espacio de nombres
    Rollback(ctx context.Context, namespace string, steps int) error

    // Version devuelve la versión actual para un espacio de nombres
    Version(ctx context.Context, namespace string) (int, error)

    // Status devuelve el estado de migración para todos los espacios de nombres
    Status(ctx context.Context) (map[string]MigrationStatus, error)
}

type MigrationStatus struct {
    Namespace      string
    CurrentVersion int
    PendingCount   int
    Dirty          bool
}
```

#### Implementación

```go
package migration

type manager struct {
    db            *sqlx.DB
    migrationsDir string  // "migrations/" por defecto
}

func NewManager(db *sqlx.DB, migrationsDir string) Manager {
    return &manager{db: db, migrationsDir: migrationsDir}
}

func (m *manager) MigrateNamespace(ctx context.Context, namespace string) error {
    // 1. Verificar si existe la tabla schema_migrations
    // 2. Obtener versión actual para el espacio de nombres
    // 3. Leer archivos de migración de migrations/{namespace}/
    // 4. Aplicar migraciones pendientes en transacción
    // 5. Actualizar tabla schema_migrations
}

func (m *manager) MigrateAll(ctx context.Context, enabledModules []string) error {
    // 1. Siempre migrar core primero
    if err := m.MigrateNamespace(ctx, "core"); err != nil {
        return err
    }

    // 2. Migrar cada módulo habilitado
    for _, module := range enabledModules {
        if err := m.MigrateNamespace(ctx, module); err != nil {
            return err
        }
    }

    return nil
}
```

---

### 4. Integración de Módulos

Los módulos pueden proporcionar migraciones de forma programática opcionalmente:

```go
// pkg/module/module.go
type Module interface {
    // ... métodos existentes ...

    // RegisterMigrations devuelve migraciones embebidas para este módulo
    // Devolver nil para usar migraciones basadas en archivos de migrations/{namespace}/
    RegisterMigrations() []Migration
}

type Migration struct {
    Version     int
    Description string
    Up          string  // SQL para aplicar
    Down        string  // SQL para revertir
}
```

Ejemplo en módulo:

```go
// internal/modules/posts/module.go
func (m *PostsModule) RegisterMigrations() []Migration {
    // Opción 1: Devolver nil para usar migraciones basadas en archivos
    return nil

    // Opción 2: Embeber migraciones en código (para bibliotecas)
    return []Migration{
        {
            Version:     1,
            Description: "Create posts table",
            Up:          `CREATE TABLE user_posts (...)`,
            Down:        `DROP TABLE user_posts`,
        },
    }
}
```

---

### 5. Uso en main.go

```go
// cmd/api/main.go
func main() {
    // ... configuración ...

    // Inicializar gestor de migraciones
    migrationMgr := migration.NewManager(db, "migrations")

    // Obtener módulos habilitados de la configuración
    enabledModules := cfg.Modules.Enabled  // ["posts", "profiles"]

    // Ejecutar migraciones para core + módulos habilitados
    if err := migrationMgr.MigrateAll(ctx, enabledModules); err != nil {
        log.Fatal("Failed to run migrations", "error", err)
    }

    // ... continuar inicio ...
}
```

---

### 6. Comandos CLI

```bash
# Migrar solo core
make migrate-core

# Migrar módulo específico
make migrate-module MODULE=posts

# Migrar todo (core + módulos habilitados)
make migrate-all

# Revertir migraciones de módulo
make migrate-rollback MODULE=posts STEPS=1

# Mostrar estado de migraciones
make migrate-status
```

**Makefile:**

```makefile
migrate-core:
	go run cmd/migrate/main.go up --namespace=core

migrate-module:
	go run cmd/migrate/main.go up --namespace=$(MODULE)

migrate-all:
	go run cmd/migrate/main.go up --all

migrate-rollback:
	go run cmd/migrate/main.go down --namespace=$(MODULE) --steps=$(STEPS)

migrate-status:
	go run cmd/migrate/main.go status
```

---

### 7. Reorganización de Migraciones

**Actual (plano - OBSOLETO, usar con espacios de nombres en su lugar):**

```
migrations/
├── 000001_init_schema_deps.up.sql         # core
├── 000002_create_auth_schema.up.sql       # core
├── 000003_create_countries_currencies.up.sql  # core
├── 000004_create_user_contacts.up.sql     # módulo profiles
├── 000005_create_user_profiles.up.sql     # módulo profiles
├── 000006_create_user_posts.up.sql        # módulo posts
├── 000007_create_post_comments.up.sql     # módulo posts
├── 000008_create_comment_likes_table.up.sql  # módulo posts
├── 000009_create_rbac_tables.up.sql       # core
├── 000014_create_timezones_table.up.sql   # core
└── 000015_create_languages_table.up.sql   # core
```

**Nuevo (con espacios de nombres y nombres descriptivos):**

```
migrations/
├── core/
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000001_core_init_uuid_v7.down.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000002_core_auth_full.down.sql
│   ├── 000003_core_rbac_full.up.sql            # era 000009
│   ├── 000003_core_rbac_full.down.sql
│   ├── 000004_core_ref_timezones.up.sql        # era 000014
│   ├── 000004_core_ref_timezones.down.sql
│   ├── 000005_core_ref_languages.up.sql        # era 000015
│   ├── 000005_core_ref_languages.down.sql
│   ├── 000006_core_ref_countries_currencies.up.sql  # era 000003
│   └── 000006_core_ref_countries_currencies.down.sql
│
├── posts/
│   ├── 000001_posts_posts.up.sql               # era 000006_create_user_posts
│   ├── 000001_posts_posts.down.sql
│   ├── 000002_posts_comments.up.sql            # era 000007_create_post_comments
│   ├── 000002_posts_comments.down.sql
│   ├── 000003_create_comment_likes.up.sql    # era 000008
│   └── 000003_create_comment_likes.down.sql
│
└── profiles/
    ├── 000001_create_user_contacts.up.sql    # era 000004
    ├── 000001_create_user_contacts.down.sql
    ├── 000002_create_user_profiles.up.sql    # era 000005
    └── 000002_create_user_profiles.down.sql
```

---

### 8. Beneficios

**Independencia de Módulos**

- Cada módulo posee sus migraciones
- Habilitar/deshabilitar módulos sin conflictos de migración
- Límites de propiedad claros

**Control de Versiones**

- Cada espacio de nombres tiene versionado independiente
- No hay conflictos de numeración global
- Fácil entender en qué versión está un módulo

**Despliegue Flexible**

- Desplegar solo con módulos necesarios
- Agregar nuevos módulos sin tocar migraciones existentes
- Revertir migraciones de módulos de forma independiente

**Experiencia del Desarrollador**

- Claro dónde colocar nuevas migraciones
- No adivinar el siguiente número global
- Comandos de migración específicos por módulo

---

### 9. Flujo de Trabajo de Migraciones

#### Crear una Nueva Migración

```bash
# Migración del núcleo
make migrate-create-core NAME=add_audit_tables

# Migración de módulo
make migrate-create MODULE=posts NAME=add_post_views
```

**Archivos generados:**

```
migrations/posts/
├── 000004_add_post_views.up.sql    # Auto-incrementado
└── 000004_add_post_views.down.sql
```

#### Aplicar Migraciones

```bash
# Desarrollo: Migrar todo
make migrate-all

# Producción: Migrar core + módulos específicos
MODULES=posts,profiles make migrate-all

# Selectivo: Migrar solo nuevo módulo
make migrate-module MODULE=warehouse
```

---

### 10. Compatibilidad Hacia Atrás

Para despliegues existentes:

1. **Script de migración único** que:
   - Respalda la tabla actual `schema_migrations`
   - Crea nueva `schema_migrations` con espacio de nombres
   - Mapea versiones antiguas a versiones con espacios de nombres
   - Marca todas como aplicadas

```sql
-- Respaldo
CREATE TABLE schema_migrations_backup AS SELECT * FROM schema_migrations;

-- Eliminar tabla antigua
DROP TABLE schema_migrations;

-- Crear nueva tabla con espacio de nombres
CREATE TABLE schema_migrations (...);

-- Insertar versiones mapeadas
INSERT INTO schema_migrations (namespace, version, dirty, applied_at)
VALUES
    ('core', 1, false, NOW()),  -- era 000001
    ('core', 2, false, NOW()),  -- era 000002
    ('core', 3, false, NOW()),  -- era 000003
    ('profiles', 1, false, NOW()),  -- era 000004
    ('profiles', 2, false, NOW()),  -- era 000005
    ('posts', 1, false, NOW()),  -- era 000006
    ...
```

2. **Script de reorganización de migraciones** que mueve archivos a carpetas de espacios de nombres

---

### 11. Plan de Implementación

**Fase 1: Infraestructura (Semana 1)**

1. Crear paquete `pkg/migration/`
2. Implementar interfaz `Manager`
3. Agregar soporte de espacios de nombres a tabla schema_migrations
4. Escribir pruebas

**Fase 2: CLI y Herramientas (Semana 1)**

1. Crear herramienta CLI `cmd/migrate/main.go`
2. Agregar comandos Makefile
3. Actualizar documentación

**Fase 3: Migración (Semana 2)**

1. Reorganizar migraciones existentes en espacios de nombres
2. Crear migración de compatibilidad hacia atrás
3. Actualizar main.go para usar nuevo gestor
4. Probar en staging

**Fase 4: Integración de Módulos (Semana 2)**

1. Agregar `RegisterMigrations()` a interfaz de módulo
2. Actualizar módulos existentes
3. Documentación y ejemplos

---

### 12. Alternativa: Fork de golang-migrate

Si queremos usar la biblioteca `golang-migrate`:

```go
import "github.com/golang-migrate/migrate/v4"

// Driver de fuente personalizado que lee de carpetas de espacios de nombres
type NamespaceSource struct {
    namespace string
    basePath  string
}

func (s *NamespaceSource) First() (version uint, err error) {
    // Leer de migrations/{namespace}/
}

// Registrar fuente personalizada
migrate.Register("namespace", &NamespaceSource{})
```

---

## Resumen

**Mejor Solución:** Gestor de migraciones personalizado con soporte de espacios de nombres.

**¿Por qué?**

- Control total sobre la lógica de espacios de nombres
- Versionado independiente de módulos
- Fácil de implementar habilitación/deshabilitación de módulos
- Límites de propiedad claros
- Sin restricciones de bibliotecas externas

**Ruta de Migración:**

1. Implementar gestor `pkg/migration/`
2. Agregar espacio de nombres a schema_migrations
3. Reorganizar migraciones existentes
4. Actualizar código de inicio
5. Documentar flujo de trabajo

**Esfuerzo Estimado:** 2-3 días para implementación completa + pruebas
