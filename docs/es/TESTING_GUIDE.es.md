# Guía de Pruebas Promenade

[🇬🇧 English](../TESTING_GUIDE.md) | [🇺🇦 Українська](../uk/TESTING_GUIDE.uk.md) | [🇩🇪 Deutsch](../de/TESTING_GUIDE.de.md) | [🇵🇹 Português](../pt/TESTING_GUIDE.pt.md) | 🇪🇸 **Español**

## Visión General

Un sistema de pruebas integral que cubre todas las capas de la aplicación con **388 pruebas (100% pasando)**:

- **Pruebas Unitarias** (183) - lógica de negocio aislada y validación de entidades
- **Pruebas de Integración** (91) - operaciones de repositorio con PostgreSQL real
- **Pruebas Smoke** (114) - flujos críticos end-to-end con base de datos real
- **Pruebas E2E** - pruebas de API HTTP (TODO)

## Inicio Rápido

```bash
# Ejecutar todas las pruebas (unitarias + integración)
make test                  # 274 pruebas en ~41s

# Suites de pruebas individuales
make test-unit            # 183 pruebas unitarias (~5s)
make test-integration     # 91 pruebas de integración (~36s)
make test-smoke           # 114 pruebas smoke (~4s)

# Cobertura y monitoreo
make test-coverage        # Reporte de cobertura HTML
make test-watch           # Modo watch (gotestsum)
```

## Base de Datos de Prueba

Las pruebas de integración usan una base de datos de prueba separada en el puerto **5433**:

```bash
# Iniciar BD de prueba
make test-db-start

# Detener y limpiar
make test-db-stop

# Ver logs
make test-db-logs
```

**Importante:** BD de prueba está completamente aislada de las bases dev/prod.

## Estructura de Pruebas

### 1. Pruebas de Integración (Repositorios)

Ubicadas junto al código: `internal/adapter/repository/postgres/*_test.go`

Ejemplo:

```go
func TestUserRepository_Create(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    repo := postgres.NewUserRepository(testDB.DB)
    ctx := context.Background()

    t.Run("creates user successfully", func(t *testing.T) {
        user := helpers.UserFixture()
        err := repo.Create(ctx, user)
        require.NoError(t, err)

        retrieved, err := repo.GetByID(ctx, user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Email, retrieved.Email)
    })
}
```

**Cobertura:**

- [+] UserRepository: Create, GetByID, GetByEmail, UpdateStatus, Suspend, Ban, Reactivate, VerifyEmail
- [+] SessionRepository: Create, GetByID, GetByRefreshToken, GetUserSessions, DeleteByUserID, DeleteExpired

### 2. Pruebas Unitarias (Casos de Uso)

_TODO: Siguiente paso_

Probarán lógica de negocio con repositorios mockeados:

- Register
- Login
- RefreshToken
- Logout
- SuspendUser
- BanUser
- ReactivateUser
- ChangePassword

### 3. Pruebas Smoke (Flujos Críticos End-to-End)

Ubicadas en: `test/smoke/*_smoke_test.go`

**114 pruebas smoke** verifican flujos críticos de usuario con operaciones reales de base de datos.

Ejemplo:

```go
func TestAuth_SmokeTest(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping smoke test in short mode")
    }

    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    ctx := context.Background()

    t.Run("[+] Complete_auth_flow", func(t *testing.T) {
        // Register → Login → GetMe → Refresh → Logout
        user, err := authUC.Register(ctx, "test@example.com", "John", "password123")
        require.NoError(t, err)

        tokens, err := authUC.Login(ctx, "test@example.com", "password123", "test-device")
        require.NoError(t, err)
        assert.NotEmpty(t, tokens.AccessToken)
        assert.NotEmpty(t, tokens.RefreshToken)

        // Continuar probando flujo completo...
    })

    t.Logf("[SUCCESS] All auth smoke tests passed!")
}
```

**Ejecutando Pruebas Smoke:**

```bash
# Todas las pruebas smoke
make test-smoke

# Prueba smoke específica
go test -v ./test/smoke -run TestRBAC_SmokeTest
go test -v ./test/smoke -run TestUserPost_SmokeTest

# Omitir en modo short
go test -short ./test/smoke  # Pruebas smoke son omitidas
```

**Cobertura por Módulo:**

| Módulo           | Escenarios | Cobertura                                                       |
| ---------------- | ---------- | --------------------------------------------------------------- |
| Auth             | 8          | Registro, login, sesiones, actualización, logout                |
| Country/Currency | 12         | Operaciones CRUD completas                                      |
| UserContact      | 11         | Email, teléfono, telegram, principal, verificación              |
| UserPost         | 12         | Borrador, publicación, destacado, programación, views, búsqueda |
| UserProfile      | 12         | Privacidad, verificación, ban/unban, views, búsqueda            |
| PostComment      | 13         | Threading, respuestas, respuestas anidadas, soft delete         |
| CommentLikes     | 5          | Like/unlike, paginación, rendimiento (100 verificaciones)       |
| RBAC             | 28         | Permisos, roles, wildcards, expiración                          |
| RBAC Integration | 13         | Escenarios reales de permisos (moderador, admin, etc.)          |
| **Total**        | **114**    | **Todas las pruebas pasando [+]**                               |

**Características Principales:**

- [+] Integración real con PostgreSQL (puerto 5433)
- [+] Verificación de ruta crítica (flujos CRUD)
- [+] Benchmarks de rendimiento incluidos
- [+] Ejecución rápida (~4 segundos para 114 pruebas)
- [+] Tasa de éxito del 100%

### 4. Pruebas de Integración HTTP (Handlers)

_TODO: Después de expansión de pruebas smoke_

Probar todos los endpoints HTTP vía enrutador Gin real:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `GET /api/auth/me`
- etc.

## Helpers de Prueba

### `test/helpers/database.go`

```go
// Conectar a BD de prueba
testDB := helpers.SetupTestDB(t)
defer testDB.Close()

// Limpiar todas las tablas
testDB.CleanupTables(t)

// Prueba transaccional (auto rollback)
testDB.RunInTransaction(t, func(tx *sqlx.Tx) {
    // Su código con tx
})
```

### `test/helpers/fixtures.go`

```go
// Usuario activo estándar
user := helpers.UserFixture()

// Usuario no verificado
user := helpers.UnverifiedUserFixture()

// Usuario suspendido
user := helpers.SuspendedUserFixture()

// Usuario baneado
user := helpers.BannedUserFixture()

// Usuario personalizado
user := helpers.UserFixture(func(u *entity.User) {
    u.Email = "custom@test.com"
    u.Status = entity.UserStatusInactive
})

// Sesión
session := helpers.SessionFixture(userID)

// Sesión expirada
session := helpers.ExpiredSessionFixture(userID)
```

## Mejores Prácticas

### [+] Hacer

- Use `testify/require` para verificaciones críticas (detiene prueba)
- Use `testify/assert` para verificaciones no críticas (continúa prueba)
- Siempre hacer limpieza: `defer testDB.CleanupTables(t)`
- Probar casos extremos: sesiones expiradas, usuarios baneados, etc.
- Usar fixtures para datos de prueba consistentes

### [X] No Hacer

- No use BD de producción para pruebas
- No cree dependencias entre pruebas
- No olvide `defer testDB.Close()`
- No codifique datos de prueba - use fixtures

## Integración CI/CD

Las pruebas están listas para CI:

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    make test-db-start
    make test
    make test-db-stop
```

## Cobertura

```bash
make test-coverage
open coverage.html
```

Meta: **>80% de cobertura** para módulos críticos (usecase, repository).

## Qué Sigue

1. [+] Pruebas de integración de repositorios - **COMPLETADO**
2. ⏳ Pruebas unitarias de casos de uso con mocks
3. ⏳ Pruebas de integración de handlers HTTP
4. ⏳ Pruebas E2E para flujos completos
5. ⏳ Pruebas de rendimiento/benchmark

## Ejemplos

### Ejecutando Pruebas Específicas

```bash
# Archivo de prueba único
go test -v ./internal/adapter/repository/postgres/user_repository_test.go

# Prueba única
go test -v ./internal/adapter/repository/postgres -run TestUserRepository_Create

# Con detector de race
go test -race ./...

# Con cobertura
go test -cover ./internal/adapter/repository/postgres
```

### Depurando Pruebas

```bash
# Salida verbose
go test -v ./...

# Con logs de BD
make test-db-logs

# Verificar estado de BD durante prueba
docker exec -it promenade_test_db psql -U system -d promenade_test
```

## Solución de Problemas

### "connection refused"

```bash
make test-db-start
# Esperar 3-5 segundos para que BD esté lista
```

### "table does not exist"

```bash
make migrate-test-up
```

### "too many open connections"

```bash
make test-db-stop
make test-db-start
```

---

**¿Preguntas?** Verifique `Makefile.test.mk` para todos los comandos disponibles.
