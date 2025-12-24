[🇬🇧 English](../MAKEFILE_ARCHITECTURE.md) | [🇺🇦 Українська](../uk/MAKEFILE_ARCHITECTURE.uk.md) | [🇩🇪 Deutsch](../de/MAKEFILE_ARCHITECTURE.de.md) | [🇵🇹 Português](../pt/MAKEFILE_ARCHITECTURE.pt.md) | 🇪🇸 **Español**

---

# Arquitectura del Makefile

Sistema modular de Makefile para separación limpia de responsabilidades y escalabilidad.

## Estructura

```
Makefile             (63 líneas)  - Archivo principal: variables, carga env, help
Makefile.dev.mk      (64 líneas)  - Flujo de trabajo de desarrollo
Makefile.test.mk     (57 líneas)  - Infraestructura de pruebas
Makefile.prod.mk     (90 líneas)  - Operaciones de producción/DevOps
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Total:              274 líneas
```

## Filosofía

**Makefile Principal** - Solo común:

- Carga de variables de entorno (`.env.development`)
- Variables compartidas (`APP_NAME`, `VERSION`, `DB_URL`, `MIGRATE`)
- Inclusión de módulos (`include Makefile.*.mk`)
- Comando help agrupado (muestra todos los módulos)

**Makefile.dev.mk** - Flujo de trabajo del desarrollador:

```bash
make install             # Instalar herramientas (swag, migrate, golangci-lint)
make dev                 # Iniciar servidor dev (postgres + migrations + app)
make build               # Construir binario (con generación de swagger)
make run                 # Ejecutar binario compilado
make lint                # Ejecutar golangci-lint
make fmt                 # Formatear código (go fmt + gofmt -s)
make deps-update         # Actualizar dependencias
make config-show         # Mostrar configuración env actual
```

**Makefile.test.mk** - Pruebas:

```bash
make test                # Pruebas unit + integration
make test-unit           # Solo pruebas unit (domain + usecase)
make test-integration    # Pruebas integration (BD real en puerto 5433)
make test-smoke          # Pruebas smoke (flujos críticos end-to-end)
make test-coverage       # Generar reporte HTML de cobertura
make test-db-start       # Iniciar base de datos de prueba
make test-db-stop        # Detener base de datos de prueba
```

**Makefile.prod.mk** - Operaciones DevOps:

```bash
# Docker
make docker-build        # Construir imagen (VERSION=0.1.0 ENV=dev)
make docker-run          # Construir + ejecutar contenedores
make docker-up           # Iniciar servicios
make docker-down         # Detener servicios
make docker-logs         # Ver logs
make docker-restart      # Reiniciar contenedores
make docker-ps           # Mostrar contenedores en ejecución
make docker-clean        # Eliminar contenedores + volumes

# Migrations
make migrate-create      # Crear migration (NAME=xxx)
make migrate-up          # Aplicar migrations
make migrate-down        # Revertir última migration
make migrate-force       # Forzar versión (VERSION=N)
make migrate-version     # Mostrar versión actual
make migrate-status      # Mostrar estado

# Documentación
make swagger-all         # Generar documentación Swagger v1 + v2

# Limpieza
make clean               # Eliminar artefactos (bin/, docs/, coverage)
```

## Beneficios

1. **Modularidad** - Cada archivo tiene una única responsabilidad
2. **Escalabilidad** - Fácil añadir `Makefile.{stage,ci,deploy}.mk`
3. **Legibilidad** - Separación clara por contexto
4. **Mantenibilidad** - Archivos pequeños enfocados vs monolito de 188 líneas
5. **Amigable para el equipo** - Desarrolladores/QA/DevOps ven solo comandos relevantes

## Ejemplos de Uso

**Desarrollo:**

```bash
make help          # Ver todos los comandos disponibles
make dev           # Iniciar desarrollo (más común)
make build         # Construir para pruebas locales
make fmt lint      # Formatear y lint antes del commit
```

**Pruebas:**

```bash
make test          # Ejecutar suite completa de pruebas antes del PR
make test-unit     # Retroalimentación rápida durante desarrollo
make test-smoke    # Verificar flujos críticos después de cambios
```

**DevOps:**

```bash
make docker-run    # Desplegar en Docker local
make migrate-up    # Aplicar migrations de base de datos
make swagger-all   # Regenerar documentación de API
make clean         # Limpiar antes de despliegue nuevo
```

## Añadiendo Nuevos Comandos

1. Identificar contexto: dev/test/prod
2. Editar `Makefile.{context}.mk` apropiado
3. Añadir `## Comentario` para mostrar en help
4. Ejecutar `make help` para verificar

Ejemplo:

```makefile
# En Makefile.dev.mk
watch: ## Observar y recargar en cambios de archivo
	air -c .air.toml
```

## Variables

Todas las variables compartidas están en el `Makefile` principal:

- `APP_NAME` - Nombre de la aplicación
- `VERSION` - Versión de construcción (predeterminado: 0.1.0)
- `ENV` - Entorno (dev/test/prod)
- `DB_URL` - Cadena de conexión PostgreSQL
- `MIGRATE` - Comando migrate con DB URL
- `DOCKER_COMPOSE` - Comando Docker Compose

Sobrescribir con:

```bash
make docker-build VERSION=1.2.3 ENV=prod
make migrate-up DB_NAME=promenade_staging
```

## Migración desde Estructura Antigua

Antes (monolito de 188 líneas):

```
Makefile  ← Todo mezclado junto
```

Después (274 líneas modular):

```
Makefile          ← Común (63 líneas)
Makefile.dev.mk   ← Desarrollo (64 líneas)
Makefile.test.mk  ← Pruebas (57 líneas)
Makefile.prod.mk  ← Producción (90 líneas)
```

**Sin breaking changes** - ¡Todos los comandos funcionan exactamente igual que antes!
