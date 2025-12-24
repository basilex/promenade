# Índice de Documentación de Promenade

[🇬🇧 English](INDEX.es.md) | [🇺🇦 Українська](../uk/INDEX.uk.md) | [🇩🇪 Deutsch](../de/INDEX.de.md) | [🇵🇹 Português](../pt/INDEX.pt.md) | 🇪🇸 **Español**

Este directorio contiene documentación completa sobre la arquitectura de la aplicación Promenade, flujos de trabajo de desarrollo y mejores prácticas.

> 🌍 **¡Nuevo!** La documentación está ahora disponible en varios idiomas. Ver [TRANSLATIONS.md](../TRANSLATIONS.md) para el estado de traducción y directrices de contribución.

---

## Empiece Aquí

### ¿Nuevo en Promenade?

1. **[ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md)** - Diagramas visuales de arquitectura y visión general de componentes
2. **[ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.es.md)** - Guía de referencia rápida para desarrolladores
3. **[README.md](../../README.md)** - README principal del proyecto

### Revisión de Arquitectura

- **[ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.es.md)** - Auditoría completa de cumplimiento de arquitectura (Core vs Módulos)

---

## Conceptos Principales

### Arquitectura & Diseño

| Documento                                               | Descripción                                | Cuándo Leer                           |
| ------------------------------------------------------- | ------------------------------------------ | ------------------------------------- |
| [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md) | Arquitectura visual completa con diagramas | Comprender estructura del sistema     |
| [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.es.md)       | Informe de cumplimiento de arquitectura    | Verificar principios de diseño        |
| [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.es.md) | Referencia rápida para patrones comunes    | Desarrollo diario                     |
| [../../internal/CORE.md](../../internal/CORE.md)        | Documentación de componentes Core          | Comprender responsabilidades del core |

### Módulos

| Documento                                                            | Descripción                            | Cuándo Leer                      |
| -------------------------------------------------------------------- | -------------------------------------- | -------------------------------- |
| [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.es.md)                    | Guía completa para crear módulos       | Construir nuevos módulos         |
| [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.es.md)                  | Principios de independencia de módulos | Comprender límites de módulos    |
| [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.es.md)    | Gestión de configuración para módulos  | Configurar configs de módulo     |
| [../../internal/modules/README.md](../../internal/modules/README.md) | Estructura de directorios de módulos   | Visión general rápida de módulos |

---

## Guías Técnicas

### Base de Datos & Persistencia

| Documento                               | Descripción                       | Cuándo Leer                   |
| --------------------------------------- | --------------------------------- | ----------------------------- |
| [UUID_V7_GUIDE.md](UUID_V7_GUIDE.es.md) | Usando UUIDs ordenados por tiempo | Trabajar con claves primarias |
| [SOFT_DELETE.md](../SOFT_DELETE.md)     | Patrones y trampas de soft delete | Implementar soft delete       |

### Infraestructura

| Documento                                         | Descripción                              | Cuándo Leer                                 |
| ------------------------------------------------- | ---------------------------------------- | ------------------------------------------- |
| [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.es.md) | Diseño del sistema de purga automatizado | Implementar políticas de retención          |
| [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md)   | Pruebas con Redis Event Bus              | Probar características orientadas a eventos |
| [LOGGING.md](../LOGGING.md)                       | Logging estructurado con contexto        | Agregar logging al código                   |

### Seguridad & Autenticación

| Documento                                             | Descripción                                     | Cuándo Leer                       |
| ----------------------------------------------------- | ----------------------------------------------- | --------------------------------- |
| [AUTH_SCHEMA.md](AUTH_SCHEMA.es.md)                   | Sistema de autenticación (registro, login, JWT) | Comprender flujo de autenticación |
| [AUTHORIZATION.md](AUTHORIZATION.es.md)               | Sistema de permisos RBAC                        | Implementar autorización          |
| [CREDENTIALS.md](../CREDENTIALS.md)                   | Usuarios y roles predeterminados para dev/test  | Probar con usuarios predefinidos  |
| [LICENSE_ARCHITECTURE.md](LICENSE_ARCHITECTURE.es.md) | Sistema de licencias de módulos con HMAC-SHA256 | Implementar módulos comerciales   |

---

## Pruebas

| Documento                                                     | Descripción                                 | Cuándo Leer                        |
| ------------------------------------------------------------- | ------------------------------------------- | ---------------------------------- |
| [TESTING_GUIDE.md](TESTING_GUIDE.es.md)                       | Estrategia completa de pruebas              | Escribir pruebas                   |
| [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.es.md)     | Configuración de infraestructura de pruebas | Configurar entorno de pruebas      |
| [MOCK_GENERATION_STANDARD.md](../MOCK_GENERATION_STANDARD.md) | Enfoque unificado de generación de mocks    | Trabajar con mocks de repositorio  |
| [MOCK_STANDARDIZATION.md](../MOCK_STANDARDIZATION.md)         | Resumen de estandarización de mocks         | Comprender unificación de mocks    |
| [../../test/README.md](../../test/README.md)                  | Estructura de directorios de pruebas        | Comprender organización de pruebas |

---

## Flujos de Trabajo de Desarrollo

### Build & Deploy

| Documento                                               | Descripción                           | Cuándo Leer              |
| ------------------------------------------------------- | ------------------------------------- | ------------------------ |
| [MAKEFILE_ARCHITECTURE.md](../MAKEFILE_ARCHITECTURE.md) | Documentación del sistema Makefile    | Usar comandos make       |
| [../../docker/README.md](../../docker/README.md)        | Configuración y deployment con Docker | Containerizar aplicación |

### Validación & Calidad

| Documento                         | Descripción                       | Cuándo Leer                  |
| --------------------------------- | --------------------------------- | ---------------------------- |
| [VALIDATION.md](../VALIDATION.md) | Patrones de validación de entrada | Agregar reglas de validación |

---

## Documentación de la API

### API V1

- [v1/v1_docs.go](../v1/v1_docs.go) - Documentación de la API V1
- [v1/v1_swagger.yaml](../v1/v1_swagger.yaml) - Especificación Swagger V1 (YAML)
- [v1/v1_swagger.json](../v1/v1_swagger.json) - Especificación Swagger V1 (JSON)

### API V2

- [v2/v2_docs.go](../v2/v2_docs.go) - Documentación de la API V2
- [v2/v2_swagger.yaml](../v2/v2_swagger.yaml) - Especificación Swagger V2 (YAML)
- [v2/v2_swagger.json](../v2/v2_swagger.json) - Especificación Swagger V2 (JSON)

---

## Por Tema

### Arquitectura Core

- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md) - Visión general visual
- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.es.md) - Revisión de cumplimiento
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.es.md) - Referencia rápida
- [../../internal/CORE.md](../../internal/CORE.md) - Componentes Core

### Sistema de Módulos

- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.es.md) - Guía de desarrollo
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.es.md) - Principios de independencia
- [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.es.md) - Configuración
- [../../internal/modules/README.md](../../internal/modules/README.md) - Índice de módulos

### Gestión de Datos

- [UUID_V7_GUIDE.md](UUID_V7_GUIDE.es.md) - Claves primarias
- [SOFT_DELETE.md](../SOFT_DELETE.md) - Patrones de soft delete
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.es.md) - Purga automatizada

### Seguridad & Auth

- [AUTH_SCHEMA.md](AUTH_SCHEMA.es.md) - Autenticación
- [AUTHORIZATION.md](AUTHORIZATION.es.md) - Sistema RBAC
- [CREDENTIALS.md](../CREDENTIALS.md) - Manejo de credenciales

### Pruebas & Calidad

- [TESTING_GUIDE.md](TESTING_GUIDE.es.md) - Estrategia de pruebas
- [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.es.md) - Configuración de pruebas
- [VALIDATION.md](../VALIDATION.md) - Validación de entrada
- [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md) - Pruebas de event bus

### Infraestructura

- [LOGGING.md](../LOGGING.md) - Logging estructurado
- [MAKEFILE_ARCHITECTURE.md](../MAKEFILE_ARCHITECTURE.md) - Sistema de build
- [../../docker/README.md](../../docker/README.md) - Configuración Docker

---

## Rutas de Aprendizaje

### Ruta 1: Comprendiendo el Sistema (Nuevo Desarrollador)

1. [README.md](../../README.md) - Visión general del proyecto
2. [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md) - Arquitectura del sistema
3. [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.es.md) - Patrones comunes
4. [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.es.md) - Construir características
5. [TESTING_GUIDE.md](TESTING_GUIDE.es.md) - Probar su código

### Ruta 2: Construyendo un Nuevo Módulo

1. [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.es.md) - Guía de creación de módulos
2. [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.es.md) - Principios de diseño
3. [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.es.md) - Configuración
4. [UUID_V7_GUIDE.md](UUID_V7_GUIDE.es.md) - Claves primarias
5. [SOFT_DELETE.md](../SOFT_DELETE.md) - Si usa soft delete
6. [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.es.md) - Si implementa purga

### Ruta 3: Revisión de Arquitectura (Technical Lead)

1. [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.es.md) - Análisis del estado actual
2. [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md) - Diagramas visuales
3. [../../internal/CORE.md](../../internal/CORE.md) - Límites del Core
4. [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.es.md) - Aislamiento de módulos

### Ruta 4: Implementación de Seguridad

1. [AUTH_SCHEMA.md](AUTH_SCHEMA.es.md) - Flujos de autenticación
2. [AUTHORIZATION.md](AUTHORIZATION.es.md) - Permisos RBAC
3. [CREDENTIALS.md](../CREDENTIALS.md) - Seguridad de credenciales

### Ruta 5: Pruebas & Calidad

1. [TESTING_GUIDE.md](TESTING_GUIDE.es.md) - Estrategia de pruebas
2. [TESTING_INFRASTRUCTURE.md](TESTING_INFRASTRUCTURE.es.md) - Configuración de pruebas
3. [VALIDATION.md](../VALIDATION.md) - Validación de entrada
4. [REDIS_BUS_TESTING.md](../REDIS_BUS_TESTING.md) - Pruebas de eventos

---

## Búsquedas Rápidas

### ¿Cómo puedo...

**...crear un nuevo módulo?**
→ [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.es.md)

**...agregar configuración a mi módulo?**
→ [MODULE_CONFIG_ARCHITECTURE.md](MODULE_CONFIG_ARCHITECTURE.es.md)

**...implementar soft delete?**
→ [SOFT_DELETE.md](../SOFT_DELETE.md)

**...agregar políticas de retención?**
→ [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.es.md)

**...usar UUIDs correctamente?**
→ [UUID_V7_GUIDE.md](UUID_V7_GUIDE.es.md)

**...implementar autenticación?**
→ [AUTH_SCHEMA.md](AUTH_SCHEMA.es.md)

**...agregar permisos RBAC?**
→ [AUTHORIZATION.md](AUTHORIZATION.es.md)

**...escribir pruebas?**
→ [TESTING_GUIDE.md](TESTING_GUIDE.es.md)

**...agregar logging?**
→ [LOGGING.md](../LOGGING.md)

**...validar entrada?**
→ [VALIDATION.md](../VALIDATION.md)

**...comprender la arquitectura?**
→ [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md)

**...verificar si mi código sigue los principios?**
→ [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.es.md)

**...obtener una referencia rápida?**
→ [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.es.md)

---

## Actualizaciones Recientes

### 22 de Diciembre de 2025

- Creada documentación completa de arquitectura:
- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.es.md) - Revisión completa de cumplimiento
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md) - Diagramas visuales
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.es.md) - Referencia rápida
- Documentación del sistema de purga actualizada:
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.es.md) - Enfoque basado en registro
- Independencia de módulos verificada (15,000+ líneas eliminadas del core)

---

## 🤝 Contribuyendo

Al agregar nueva documentación:

1. **Elija el tipo correcto:**

   - `ARCHITECTURE_*.md` - Arquitectura y patrones de diseño
   - `MODULE_*.md` - Documentación del sistema de módulos
   - `*_GUIDE.md` - Guías prácticas y tutoriales
   - `*_SCHEMA.md` - Esquemas y estructuras de datos
   - `README.md` - Visiones generales de directorios

2. **Actualice este índice:**

   - Agregue a la sección relevante
   - Actualice "Actualizaciones Recientes"
   - Agregue a "Búsquedas Rápidas" si es aplicable

3. **Haga referencias cruzadas:**

   - Enlace a documentos relacionados
   - Actualice docs relacionados con enlaces de vuelta

4. **Mantenga actualizado:**
   - Actualice cuando la arquitectura cambie
   - Archive docs desactualizados con prefijo `DEPRECATED_`

---

## 📞 Soporte

- **¿Preguntas sobre arquitectura?** → Lea [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md)
- **¿Preguntas sobre módulos?** → Lea [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.es.md)
- **¿Preguntas sobre pruebas?** → Lea [TESTING_GUIDE.md](TESTING_GUIDE.es.md)
- **¿Otras preguntas?** → Verifique este índice o el [README.md](../../README.md) principal

---

**Versión de Documentación:** 2.0  
**Última Actualización:** 22 de Diciembre de 2025  
**Mantenido por:** Equipo de Desarrollo Promenade
