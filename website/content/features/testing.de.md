---
title: "Test-Infrastruktur"
description: "400+ Tests mit manuellen Mocks"
weight: 5
---

## Umfassendes Testen

Promenade hat **400+ Tests**, die alle Schichten mit einem **manuellen Mocks-Ansatz** abdecken.

### Testabdeckung

**Kern-Tests (275 Tests):**

- Domain-Entitäten: 39 Tests
- Use Cases: 236 Tests (Auth, RBAC, Referenzdaten)

**Modul-Tests:**

- Posts: 33 Tests (83.3% Abdeckung)
- Profiles: 21 Tests (80.4% Abdeckung)
- Analytics: 11 Tests
- Billing: 375 Tests (100% Abdeckung)

**Utilities:** 51 Tests, 89.5% durchschnittliche Abdeckung

### Test-Typen

```bash
# Alle Tests (~20 Sekunden)
make test

# Nur Kern-Tests
make test-core

# Modul-Tests
make test-module-posts
make test-module-billing

# Abdeckungsbericht
make test-coverage
```

### Manuelles Mocks-Muster

**Kein mockgen** - einfache, explizite Mocks:

```go
// Mock-Repository inline
type mockUserRepo struct {
    users map[string]*User
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
    if user, ok := m.users[id]; ok {
        return user, nil
    }
    return nil, ErrUserNotFound
}

// Verwendung in Tests
func TestRegisterUser(t *testing.T) {
    repo := &mockUserRepo{users: make(map[string]*User)}
    usecase := NewUserUseCase(repo)

    user, err := usecase.Register(ctx, "test@example.com", "password")
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### Integrationstests

Tests mit **echtem PostgreSQL**:

```bash
# Test-Datenbank starten
make test-db-start

# Integrationstests ausführen
make test-integration

# Test-Datenbank stoppen
make test-db-stop
```

### Vorteile

✅ **Schnell** - Komplette Test-Suite in ~20 Sekunden  
✅ **Explizit** - Keine Generator-Magie, verständliche Mocks  
✅ **Zuverlässig** - Integrationstests mit echter DB  
✅ **Abdeckung** - 89.5% durchschnittliche Code-Abdeckung

[Test-Leitfaden →](/docs/TESTING_GUIDE)
