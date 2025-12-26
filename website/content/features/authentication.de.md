---
title: "Authentifizierung & RBAC"
description: "JWT-Authentifizierung mit rollenbasierter Zugriffskontrolle"
weight: 3
---

## Authentifizierung & Autorisierung

Produktionsreife Authentifizierung mit **JWT-Tokens** und **RBAC** (Role-Based Access Control).

### JWT-Authentifizierung

```go
// Login gibt Access + Refresh Tokens zurück
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "secure123"
}

// Antwort
{
  "access_token": "eyJhbGc...",   // 15 Minuten
  "refresh_token": "eyJhbGc...",  // 7 Tage
  "user": { ... }
}
```

### RBAC-System

**4 Integrierte Rollen:**

- **Admin** - Vollständiger Systemzugriff (`*` Berechtigung)
- **Moderator** - Inhaltsmoderation
- **User** - Grundlegende Operationen
- **Guest** - Nur-Lese-Zugriff

**Berechtigungsformat:** `ressource:aktion`

```
posts:create
posts:update
posts:delete
users:manage
*  # Wildcard - Vollzugriff
```

### Middleware-Schutz

```go
// Authentifizierung erforderlich
router.Use(authMiddleware.RequireAuth())

// Spezifische Berechtigung erforderlich
router.POST("/posts",
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

### Sitzungsverwaltung

- Mehrere Sitzungen pro Benutzer
- Sitzungswiderruf
- Geräteverfolgung
- Aktivitätsüberwachung

### Standard-Testbenutzer

```
admin@promenade.com     | passw0rd | Admin
moderator@promenade.com | passw0rd | Moderator
user@promenade.com      | passw0rd | User
```

### Vorteile

✅ **Sicher** - JWT mit HMAC-SHA256-Signatur  
✅ **Flexibel** - Wildcard und granulare Berechtigungen  
✅ **Skalierbar** - Zustandslose Tokens, optionales Session-Tracking  
✅ **Produktionsreif** - In echten Projekten getestet

[Ausführliche Dokumentation →](/promenade/docs/AUTHORIZATION)
