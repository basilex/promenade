# Promenade Testing-Leitfaden

[ English](../TESTING_GUIDE.md) | [ Українська](../uk/TESTING_GUIDE.uk.md) |  **Deutsch** | [ Português](../pt/TESTING_GUIDE.pt.md) | [ Español](../es/TESTING_GUIDE.es.md)

## Überblick

Ein umfassendes Testsystem, das alle Anwendungsebenen mit **388 Tests (100% bestanden)** abdeckt:

- **Unit-Tests** (183) - isolierte Geschäftslogik und Entity-Validierung
- **Integrationstests** (91) - Repository-Operationen mit echtem PostgreSQL
- **Smoke-Tests** (114) - End-to-End kritische Flows mit echter Datenbank
- **E2E-Tests** - HTTP API Tests (TODO)

## Schnellstart

```bash
# Alle Tests ausführen (Unit + Integration)
make test                  # 274 Tests in ~41s

# Einzelne Test-Suiten
make test-unit            # 183 Unit-Tests (~5s)
make test-integration     # 91 Integrationstests (~36s)
make test-smoke           # 114 Smoke-Tests (~4s)

# Coverage und Monitoring
make test-coverage        # HTML-Coverage-Report
make test-watch           # Watch-Modus (gotestsum)
```

## Test-Datenbank

Integrationstests verwenden eine separate Test-Datenbank auf Port **5433**:

```bash
# Test-DB starten
make test-db-start

# Stoppen und aufräumen
make test-db-stop

# Logs anzeigen
make test-db-logs
```

**Wichtig:** Test-DB ist vollständig von dev/prod-Datenbanken isoliert.

## Test-Struktur

### 1. Integrationstests (Repositories)

Neben dem Code lokalisiert: `internal/adapter/repository/postgres/*_test.go`

Beispiel:

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

**Abdeckung:**

- [+] IUserRepository: Create, GetByID, GetByEmail, UpdateStatus, Suspend, Ban, Reactivate, VerifyEmail
- [+] ISessionRepository: Create, GetByID, GetByRefreshToken, GetUserSessions, DeleteByUserID, DeleteExpired

### 2. Unit-Tests (Use Cases)

_TODO: Nächster Schritt_

Werden Geschäftslogik mit gemockten Repositories testen:

- Register
- Login
- RefreshToken
- Logout
- SuspendUser
- BanUser
- ReactivateUser
- ChangePassword

### 3. Smoke-Tests (End-to-End kritische Flows)

Lokalisiert in: `test/smoke/*_smoke_test.go`

**114 Smoke-Tests** verifizieren kritische Benutzer-Flows mit echten Datenbankoperationen.

Beispiel:

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

        // Vollständigen Flow weiter testen...
    })

    t.Logf("[SUCCESS] All auth smoke tests passed!")
}
```

**Smoke-Tests ausführen:**

```bash
# Alle Smoke-Tests
make test-smoke

# Spezifischer Smoke-Test
go test -v ./test/smoke -run TestRBAC_SmokeTest
go test -v ./test/smoke -run TestUserPost_SmokeTest

# Im Short-Modus überspringen
go test -short ./test/smoke  # Smoke-Tests werden übersprungen
```

**Abdeckung nach Modul:**

| Modul            | Szenarien | Abdeckung                                                   |
| ---------------- | --------- | ----------------------------------------------------------- |
| Auth             | 8         | Registrierung, Login, Sessions, Refresh, Logout             |
| Country/Currency | 12        | Vollständige CRUD-Operationen                               |
| UserContact      | 11        | Email, Telefon, Telegram, Primary, Verifizierung            |
| UserPost         | 12        | Entwurf, Veröffentlichung, Featured, Planung, Views, Suche  |
| UserProfile      | 12        | Privatsphäre, Verifizierung, Ban/Unban, Views, Suche        |
| PostComment      | 13        | Threading, Antworten, verschachtelte Antworten, Soft Delete |
| CommentLikes     | 5         | Like/Unlike, Pagination, Performance (100 Checks)           |
| RBAC             | 28        | Berechtigungen, Rollen, Wildcards, Ablauf                   |
| RBAC Integration | 13        | Realistische Berechtigungsszenarien (Moderator, Admin usw.) |
| **Gesamt**       | **114**   | **Alle Tests bestanden [+]**                                |

**Hauptmerkmale:**

- [+] Echte PostgreSQL-Integration (Port 5433)
- [+] Kritische Pfadverifizierung (CRUD-Flows)
- [+] Performance-Benchmarks enthalten
- [+] Schnelle Ausführung (~4 Sekunden für 114 Tests)
- [+] 100% Erfolgsquote

### 4. HTTP-Integrationstests (Handler)

_TODO: Nach Smoke-Tests-Erweiterung_

Alle HTTP-Endpunkte über echten Gin-Router testen:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `GET /api/auth/me`
- usw.

## Test-Helfer

### `test/helpers/database.go`

```go
// Mit Test-DB verbinden
testDB := helpers.SetupTestDB(t)
defer testDB.Close()

// Alle Tabellen bereinigen
testDB.CleanupTables(t)

// Transaktionstest (Auto-Rollback)
testDB.RunInTransaction(t, func(tx *sqlx.Tx) {
    // Ihr Code mit tx
})
```

### `test/helpers/fixtures.go`

```go
// Standard aktiver Benutzer
user := helpers.UserFixture()

// Nicht verifizierter Benutzer
user := helpers.UnverifiedUserFixture()

// Suspendierter Benutzer
user := helpers.SuspendedUserFixture()

// Gebannter Benutzer
user := helpers.BannedUserFixture()

// Benutzerdefinierter Benutzer
user := helpers.UserFixture(func(u *entity.User) {
    u.Email = "custom@test.com"
    u.Status = entity.UserStatusInactive
})

// Session
session := helpers.SessionFixture(userID)

// Abgelaufene Session
session := helpers.ExpiredSessionFixture(userID)
```

## Best Practices

### [+] Tun

- Verwenden Sie `testify/require` für kritische Prüfungen (stoppt Test)
- Verwenden Sie `testify/assert` für nicht-kritische Prüfungen (setzt Test fort)
- Immer aufräumen: `defer testDB.CleanupTables(t)`
- Edge Cases testen: abgelaufene Sessions, gebannte Benutzer usw.
- Fixtures für konsistente Testdaten verwenden

### [X] Nicht tun

- Verwenden Sie keine Production-DB für Tests
- Erstellen Sie keine Abhängigkeiten zwischen Tests
- Vergessen Sie nicht `defer testDB.Close()`
- Keine Test-Daten hardcoden - Fixtures verwenden

## CI/CD-Integration

Tests sind CI-ready:

```yaml
# .github/workflows/test.yml
- name: Run tests
  run: |
    make test-db-start
    make test
    make test-db-stop
```

## Coverage

```bash
make test-coverage
open coverage.html
```

Ziel: **>80% Coverage** für kritische IModule (usecase, repository).

## Was kommt als Nächstes

1. [+] Repository-Integrationstests - **FERTIG**
2. **TODO** Use-Case-Unit-Tests mit Mocks
3. **TODO** HTTP-Handler-Integrationstests
4. **TODO** E2E-Tests für vollständige Flows
5. **TODO** Performance-/Benchmark-Tests

## Beispiele

### Spezifische Tests ausführen

```bash
# Einzelne Testdatei
go test -v ./internal/adapter/repository/postgres/user_repository_test.go

# Einzelner Test
go test -v ./internal/adapter/repository/postgres -run TestUserRepository_Create

# Mit Race-Detector
go test -race ./...

# Mit Coverage
go test -cover ./internal/adapter/repository/postgres
```

### Tests debuggen

```bash
# Ausführliche Ausgabe
go test -v ./...

# Mit DB-Logs
make test-db-logs

# DB-Status während Test prüfen
docker exec -it promenade_test_db psql -U system -d promenade_test
```

## Fehlerbehebung

### "connection refused"

```bash
make test-db-start
# 3-5 Sekunden warten bis DB bereit ist
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

**Fragen?** Siehe `Makefile.test.mk` für alle verfügbaren Befehle.
