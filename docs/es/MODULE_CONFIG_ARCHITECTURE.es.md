# Arquitectura de Configuración de Módulos

🇬🇧 [English](../MODULE_CONFIG_ARCHITECTURE.md) | [🇺🇦 Українська](../uk/MODULE_CONFIG_ARCHITECTURE.uk.md) | [🇩🇪 Deutsch](../de/MODULE_CONFIG_ARCHITECTURE.de.md) | [🇵🇹 Português](../pt/MODULE_CONFIG_ARCHITECTURE.pt.md) | 🇪🇸 **Español**

## Descripción General

Los módulos en Promenade son **segmentos verticales completamente autónomos** que gestionan su propia configuración. Cada módulo carga su configuración desde su propio árbol de directorios, garantizando una independencia completa del sistema central.

## Principios de Arquitectura

### 1. Autonomía del Módulo

- Cada módulo es responsable de cargar su propia configuración
- Los módulos almacenan configuraciones en su propio directorio: `internal/modules/{nombre_modulo}/config/`
- El núcleo no carga ni gestiona configuraciones de módulos
- El núcleo solo proporciona infraestructura (DB, EventBus, JWT) y contexto de entorno

### 2. Soporte de Entornos

- Cada módulo tiene tres archivos de configuración específicos de entorno:
  - `config.dev.yaml` - Configuración de desarrollo
  - `config.test.yaml` - Configuración de pruebas
  - `config.prod.yaml` - Configuración de producción
- El SDK del módulo carga automáticamente la configuración correcta según la variable `ENVIRONMENT`

### 3. Flujo de Carga de Configuración

```
Inicio de la Aplicación
    ↓
Núcleo carga app.{env}.yaml
    ↓
Núcleo inicializa infraestructura (DB, EventBus, etc.)
    ↓
Registro de módulos descubre módulos
    ↓
Para cada módulo habilitado:
    Se llama Module.Initialize(ctx, core)
        ↓
    Módulo carga internal/modules/{nombre}/config/config.{env}.yaml
        ↓
    Módulo valida configuraciones (licencia, características, etc.)
        ↓
    Módulo inicializa repositorios, casos de uso, manejadores
        ↓
    Module.RegisterRoutes() registra endpoints HTTP
```

## Estructura de Directorios

```
config/
├── app.dev.yaml          # Configuración del núcleo - desarrollo
├── app.test.yaml         # Configuración del núcleo - pruebas
├── app.prod.yaml         # Configuración del núcleo - producción
└── modules.yaml          # Registro de módulos (habilitado/deshabilitado)

internal/modules/
├── posts/
│   ├── config/
│   │   ├── config.dev.yaml   # Configuración del módulo posts para dev
│   │   ├── config.test.yaml  # Configuración del módulo posts para pruebas
│   │   └── config.prod.yaml  # Configuración del módulo posts para prod
│   ├── entity/
│   ├── repository/
│   ├── usecase/
│   ├── handler/
│   └── module.go            # Carga propia configuración en Initialize()
│
├── warehouse/
│   ├── config/
│   │   ├── config.dev.yaml   # Configuración del módulo warehouse para dev
│   │   ├── config.test.yaml  # Configuración del módulo warehouse para pruebas
│   │   └── config.prod.yaml  # Configuración del módulo warehouse para prod
│   └── module.go            # Carga propia configuración en Initialize()
│
└── profiles/
    ├── config/
    │   ├── config.dev.yaml   # Configuración del módulo profiles para dev
    │   ├── config.test.yaml  # Configuración del módulo profiles para pruebas
    │   └── config.prod.yaml  # Configuración del módulo profiles para prod
    └── ...
```

## Ámbitos de Configuración

### Configuración del Núcleo (`config/app.{env}.yaml`)

**Gestionado por**: `internal/infrastructure/config/yaml_config.go`

Contiene:

- Metadatos de la aplicación (nombre, versión, entorno)
- Configuración del servidor (host, puerto, timeouts)
- Conexión a base de datos (host, puerto, credenciales)
- Configuración JWT (secreto, expiración)
- Configuración de logging
- Configuración CORS
- Adaptador de bus de eventos (memory/redis)
- Limitación de tasa
- Servicio de correo electrónico

**Nunca contiene**: Configuraciones de lógica de negocio específicas de módulos

### Configuración del Módulo (`internal/modules/{nombre}/config/config.{env}.yaml`)

**Gestionado por**: Cada módulo usando `pkg/module/config`

Contiene:

- Metadatos del módulo (nombre, versión, flag de habilitación)
- Configuraciones específicas del módulo
- Políticas de purga (días de retención, tamaños de lote)
- Definiciones de permisos
- Flags de características
- Claves de licencia (para módulos comerciales)

**Nunca contiene**: Configuraciones de infraestructura central

## Ejemplo de Implementación

### Carga de Configuración del Módulo Posts

```go
// internal/modules/posts/module.go
package posts

import (
    "context"
    "github.com/basilex/promenade/pkg/module"
    moduleconfig "github.com/basilex/promenade/pkg/module/config"
)

type PostsModule struct {
    *module.BaseModule
    config *moduleconfig.Config
    // ... otros campos
}

func (m *PostsModule) Initialize(ctx context.Context, core *module.Core) error {
    // Llamar implementación base
    if err := m.BaseModule.Initialize(ctx, core); err != nil {
        return err
    }

    // Cargar propia configuración del módulo desde su propio directorio
    config, err := moduleconfig.Load("internal/modules/posts/config", core.Config.AppName)
    if err != nil {
        return fmt.Errorf("failed to load posts config: %w", err)
    }
    m.config = config

    // Usar configuraciones
    retentionDays := m.config.GetRetentionDays("posts", 30)
    batchSize := m.config.GetBatchSize("posts", 100)

    // ... inicializar repositorios, casos de uso, manejadores

    return nil
}
```

### Estructura del Archivo de Configuración del Módulo

```yaml
# internal/modules/posts/config/config.dev.yaml
module:
  enabled: true
  name: "posts"
  version: "1.0.0"

settings:
  max_content_length: 10000
  allow_markdown: true

purge:
  enabled: true
  schedule: "0 2 * * *"
  settings:
    posts:
      retention_days: 30
      batch_size: 100
    comments:
      retention_days: 60
      batch_size: 50

permissions:
  - resource: "posts"
    actions: ["create", "read", "update", "delete"]
  - resource: "comments"
    actions: ["create", "read", "update", "delete"]

features:
  enable_likes: true
  enable_sharing: true
```

## SDK de Configuración del Módulo

### Cargar Configuración

```go
import moduleconfig "github.com/basilex/promenade/pkg/module/config"

// Cargar configuración para entorno actual
config, err := moduleconfig.Load("internal/modules/{nombre}/config", environment)
```

### Acceder a Configuraciones

```go
// Obtener días de retención para purga con fallback predeterminado
retentionDays := config.GetRetentionDays("nombre_entidad", 30)

// Obtener tamaño de lote para purga con fallback predeterminado
batchSize := config.GetBatchSize("nombre_entidad", 100)

// Verificar si característica está habilitada
enabled := config.GetFeature("enable_likes", true)

// Obtener configuración anidada
value := config.GetNestedSetting("settings", "max_content_length")
```

## Migración desde Configuraciones Centralizadas

### Arquitectura Antigua (Obsoleta)

```
config/modules/
├── posts.dev.yaml          Centralizado
├── posts.test.yaml         Centralizado
├── posts.prod.yaml         Centralizado
└── ...

Núcleo carga todas las configuraciones de módulos   Acoplamiento fuerte
Núcleo pasa configuraciones a módulos   Dependencia
```

### Nueva Arquitectura (Actual)

```
internal/modules/posts/config/
├── config.dev.yaml         Propiedad del módulo
├── config.test.yaml        Propiedad del módulo
└── config.prod.yaml        Propiedad del módulo

Módulo carga propia configuración   Autónomo
Módulo gestiona propias configuraciones   Independiente
```

## Beneficios

### 1. Verdadera Independencia del Módulo

- Los módulos pueden desarrollarse, probarse e implementarse de forma independiente
- No se necesitan cambios en el núcleo al agregar/modificar configuraciones de módulos
- Los módulos son verdaderos plugins autónomos

### 2. Mejor Encapsulación

- La configuración está junto con el código que configura
- Propiedad y responsabilidad claras
- Más fácil entender las capacidades del módulo

### 3. Núcleo Simplificado

- El núcleo solo gestiona infraestructura
- Sin lógica de negocio en configuraciones del núcleo
- Separación de responsabilidades más clara

### 4. Pruebas Más Fáciles

- Cada módulo puede tener diferentes configuraciones de prueba
- Sin contaminación de configuración global
- Las pruebas de módulos están aisladas

### 5. Despliegue Flexible

- Habilitar/deshabilitar módulos sin cambios en el núcleo
- Diferentes entornos pueden tener diferentes configuraciones de módulos
- Los módulos comerciales pueden tener validación de licencia

## Módulos Comerciales

Los módulos comerciales (por ejemplo, warehouse) requieren claves de licencia:

```yaml
# internal/modules/warehouse/config/config.prod.yaml
module:
  enabled: true
  name: "warehouse"
  version: "1.2.0"
  license_key: "su-clave-de-licencia-comercial-aqui"

settings:
  max_items: 100000
  enable_barcode_scanner: true
```

El módulo valida la licencia durante `Initialize()`:

```go
func (m *WarehouseModule) verifyLicense() error {
    licenseKey := m.config.GetNestedSetting("module", "license_key").(string)
    if licenseKey == "" {
        return fmt.Errorf("warehouse module requires license key")
    }
    // Validar licencia...
    return nil
}
```

## Mejores Prácticas

### HACER

- Almacenar configuraciones de módulos en `internal/modules/{nombre}/config/`
- Cargar configuración en el método `Initialize()` del módulo
- Usar archivos de configuración específicos de entorno
- Validar configuraciones críticas durante la inicialización
- Usar métodos auxiliares de `pkg/module/config`
- Proporcionar valores predeterminados sensatos para configuraciones opcionales

### NO HACER

- Colocar configuraciones de módulos en `config/modules/` (obsoleto)
- Cargar configuraciones de módulos en el núcleo
- Colocar configuraciones de módulos en la configuración del núcleo
- Codificar valores específicos de entorno
- Omitir validación de configuración
- Acceder a configuraciones de otros módulos

## Depuración

### Verificar qué configuración se cargó:

```go
logger.FromContext(ctx).Info("Module config loaded",
    "module", m.GetMetadata().Name,
    "config_path", "internal/modules/posts/config",
    "environment", os.Getenv("ENVIRONMENT"),
)
```

### Verificar entorno:

```bash
echo $ENVIRONMENT  # Debe ser: development, test o production
```

### Fallos en la carga de configuración:

- Verifique si el archivo existe: `internal/modules/{nombre}/config/config.{env}.yaml`
- Verifique la sintaxis YAML
- Verifique los permisos del archivo
- Asegúrese de que el entorno esté configurado correctamente

## Documentación Relacionada

- [Guía de Desarrollo de Módulos](MODULE_DEVELOPMENT.es.md)
- [Independencia de Módulos](MODULE_INDEPENDENCE.es.md)
- [Infraestructura de Pruebas](TESTING_INFRASTRUCTURE.es.md)
- [Migración de Configuración](CONFIG_MIGRATION.es.md)
