# Guía de Implementación UUID v7

## Visión General

Este proyecto ha sido migrado de UUID v4 a UUID v7 para generación de claves primarias. UUID v7 proporciona beneficios significativos de rendimiento mientras mantiene la unicidad global y capacidades de generación distribuida de UUID v4.

## ¿Por Qué UUID v7?

### Beneficios sobre UUID v4

1. **Ordenado por tiempo**: Los UUIDs están naturalmente ordenados por tiempo de creación
2. **Mejor rendimiento B-tree**: Reduce divisiones de página en índices PostgreSQL (INSERTs 20-50% más rápidos en benchmarks)
3. **Fragmentación reducida**: IDs secuenciales-ish minimizan fragmentación de índice
4. **Amigable con caché**: Mejor localidad de referencia para registros recién creados
5. **Timestamp extraíble**: Puede recuperar tiempo de creación del propio UUID
6. **Aún globalmente único**: Mantiene garantías de unicidad de UUID v4

### Comparación con otras estrategias

| Estrategia       | Pros                                                    | Contras                                            |
| ---------------- | ------------------------------------------------------- | -------------------------------------------------- |
| **UUID v7**      | [+] Ordenado por tiempo, globalmente único, distribuido | Ligeramente mayor que BIGINT (16 bytes)            |
| UUID v4          | Globalmente único, distribuido                          | [X] Aleatorio = localidad de índice pobre          |
| SERIAL/BIGSERIAL | Pequeño, secuencial, rápido                             | [X] Punto único de falla, problemas de replicación |
| ULID             | Similar a v7, codificado en base32                      | Menos estándar, soporte limitado de bibliotecas    |
| Snowflake ID     | Rápido, ordenado por tiempo                             | Requiere servicio de coordinación                  |

## Implementación

### Configuración PostgreSQL

El proyecto incluye dos migraciones:

1. **000005_add_uuidv7_support.sql**: Agrega la función `uuid_generate_v7()` a PostgreSQL
2. **000006_switch_tables_to_uuidv7.sql**: Actualiza valores predeterminados de tablas para usar v7

```sql
-- Generar un UUID v7 en PostgreSQL
SELECT uuid_generate_v7();
-- Ejemplo: 018d2f07-0f3c-7000-8000-123456789abc
```

### Código Go

Use el paquete `pkg/uuidv7`:

```go
import "github.com/basilex/promenade/pkg/uuidv7"

// Generar nuevo UUID v7
id := uuidv7.New()

// Generar con timestamp específico (útil para pruebas)
id := uuidv7.NewWithTime(time.Now())

// Extraer timestamp de creación del UUID
timestamp := uuidv7.ExtractTime(id)

// Verificar si UUID es versión 7
if uuidv7.IsV7(id) {
    // ...
}
```

### Ruta de Migración

Los registros existentes con UUID v4 seguirán funcionando. El sistema soporta UUIDs mixtos:

| Tipo de Registro | Versión UUID | Notas          |
| ---------------- | ------------ | -------------- |
| Registros viejos | UUID v4      | Pre-migración  |
| Registros nuevos | UUID v7      | Post-migración |

## Consideraciones de Rendimiento

### Cuándo Usar UUID v7

[+] **Buen ajuste:**

- Sistemas distribuidos donde múltiples nodos generan IDs
- Despliegues multi-región
- Arquitectura de microservicios
- APIs donde los clientes necesitan generar IDs
- Tablas con alta tasa de INSERT
- Cuando se necesita ordenación natural basada en tiempo

[X] **Reconsidere si:**

- Sistema de base de datos única sin necesidades de distribución
- Almacenamiento extremadamente restringido (UUID = 16 bytes vs BIGINT = 8 bytes)
- Se necesita análisis de brechas secuenciales (use SERIAL en su lugar)

### Optimización de Índice PostgreSQL

UUID v7 funciona mejor con índices B-tree debido a su naturaleza ordenada por tiempo:

```sql
-- Bueno: Clave primaria usa B-tree automáticamente
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7()
);

-- Bueno: Consultas basadas en tiempo se benefician de ordenación UUID v7
CREATE INDEX idx_users_created ON users(created_at DESC);

-- Óptimo: Combine UUID v7 con timestamp para consultas de rango
SELECT * FROM users
WHERE id >= uuid_generate_v7_from_timestamp('2024-01-01'::timestamp)
ORDER BY id DESC;
```

### Monitoreando Rendimiento

Compare rendimiento de INSERT antes/después de migración:

```sql
-- Verificar hinchazón de índice
SELECT schemaname, tablename,
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE tablename IN ('users', 'products', 'roles')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Monitorear rendimiento de INSERT
EXPLAIN ANALYZE
INSERT INTO users (email, name, password)
VALUES ('test@example.com', 'Test', 'hash');
```

## Recomendaciones para Claves Foráneas

### Use UUID v7 para Claves Foráneas También

```go
type Product struct {
    ID        uuidv7.UUID `db:"id"`        // UUID v7 clave primaria
    UserID    uuidv7.UUID `db:"user_id"`   // UUID v7 clave foránea
    Name      string      `db:"name"`
    CreatedAt time.Time   `db:"created_at"`
}
```

### Indexando Claves Foráneas

Siempre indexe claves foráneas para rendimiento de JOIN:

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL
);

-- Crítico: Indexar la clave foránea
CREATE INDEX idx_products_user_id ON products(user_id);
```

### Claves Compuestas con UUID v7

Para relaciones muchos-a-muchos:

```sql
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

-- Índice para búsqueda inversa
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
```

## Estrategias Alternativas (Consideración Futura)

### Opción 1: ULID (Universally Unique Lexicographically Sortable Identifier)

Similar a UUID v7 pero usa codificación base32 (representación de cadena más corta):

```
01ARZ3NDEKTSV4RRFFQ69G5FAV  (ULID - 26 caracteres)
018d2f07-0f3c-7000-8000-123456789abc  (UUID v7 - 36 caracteres)
```

**Cuándo usar**: Si frecuentemente muestra IDs a usuarios y quiere cadenas más cortas.

### Opción 2: Snowflake IDs

Snowflake de Twitter: enteros de 64 bits con timestamp, ID de datacenter y secuencia.

```go
// Ejemplo de estructura (no implementado en este proyecto)
// 41 bits: timestamp
// 10 bits: datacenter/worker ID
// 12 bits: número de secuencia
```

**Cuándo usar**: Sistemas de throughput extremadamente alto (escala Twitter) donde coordinación es aceptable.

### Opción 3: PostgreSQL SERIAL/BIGSERIAL

Enteros auto-incrementales tradicionales:

```sql
CREATE TABLE simple_table (
    id BIGSERIAL PRIMARY KEY,  -- 1, 2, 3, 4, ...
    name VARCHAR(255)
);
```

**Cuándo usar**: Aplicaciones simples de base de datos única sin necesidades de distribución.

### Opción 4: Claves Naturales Compuestas

Use datos de negocio significativos como claves:

```sql
CREATE TABLE country_codes (
    code CHAR(2) PRIMARY KEY,  -- 'US', 'GB', 'FR'
    name VARCHAR(100)
);
```

**Cuándo usar**: Cuando claves naturales son estables, cortas y verdaderamente únicas.

## Pruebas

Execute la suite de pruebas para verificar implementación UUID v7:

```bash
# Probar paquete UUID v7
go test -v ./pkg/uuidv7/

# Ejecutar con benchmarks
go test -bench=. ./pkg/uuidv7/

# Probar integración con base de datos
make test-integration
```

Resultados esperados de benchmark (aproximados):

```
BenchmarkNew-8      3000000    450 ns/op    (UUID v7)
BenchmarkNewV4-8    2800000    480 ns/op    (UUID v4)
```

## Lista de Verificación de Migración

- [x] Agregar función `uuid_generate_v7()` a PostgreSQL
- [x] Actualizar valores predeterminados de tablas para usar UUID v7
- [x] Actualizar código Go para usar paquete `pkg/uuidv7`
- [x] Actualizar todos los use cases (auth, product, role)
- [ ] Ejecutar migraciones: `make migrate-up`
- [ ] Probar generación de UUID: `go test ./pkg/uuidv7/`
- [ ] Monitorear métricas de producción después del despliegue

## Lectura Adicional

- [RFC 9562 - UUID Version 7 (Draft)](https://datatracker.ietf.org/doc/draft-ietf-uuidrev-rfc4122bis/)
- [PostgreSQL UUID Performance](https://www.2ndquadrant.com/en/blog/sequential-uuid-generators/)
- [UUID v7 in Production (Blog post)](https://buildkite.com/blog/goodbye-integers-hello-uuids)

## ¿Preguntas?

Vea `.github/copilot-instructions.md` para convenciones del proyecto o pregunte en el chat del equipo.
