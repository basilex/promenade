# Autorisierungs-Middleware Leitfaden

Vollständiger Leitfaden zur Verwendung der RBAC (Role-Based Access Control) Autorisierungs-Middleware in Promenade.

## Inhaltsverzeichnis

- [Überblick](#überblick)
- [Architektur](#architektur)
- [Berechtigungssystem](#berechtigungssystem)
- [Rollensystem](#rollensystem)
- [Middleware-Methoden](#middleware-methoden)
- [Verwendungsbeispiele](#verwendungsbeispiele)
- [Best Practices](#best-practices)
- [Häufige Muster](#häufige-muster)
- [Fehlerbehandlung](#fehlerbehandlung)
- [Autorisierung Testen](#autorisierung-testen)

## Überblick

Die Autorisierungs-Middleware bietet **flexible, feinkörnige Zugriffskontrolle** für API-Endpunkte mit einem berechtigungsbasierten RBAC-System. Unterstützt:

- [+] **Berechtigungsbasierte Prüfungen** - Feinkörnige Kontrolle im Format `resource:action`
- [+] **Rollenbasierte Prüfungen** - Schnelle Prüfungen der Benutzerrollen (admin, moderator usw.)
- [+] **Wildcard-Berechtigungen** - Muster wie `*:*`, `posts:*`, `*:read`
- [+] **Zusammengesetzte Prüfungen** - RequireAny, RequireAll für komplexe Logik
- [+] **Klare Trennung** - Funktioniert unabhängig von der Authentifizierungs-Middleware

## Architektur

**Request-Ablauf:**

1. **HTTP-Request** kommt an
   ↓
2. **RequireAuth Middleware**
   - Validiert JWT-Token
   - Setzt user_id in Kontext
     ↓
3. **RequirePermission Middleware**
   - Ruft user_id aus Kontext ab
   - Fragt Benutzerrollen ab
   - Prüft Rollenberechtigungen (mit Wildcard-Unterstützung)
   - Erlaubt/Verweigert Request
     ↓
4. **Handler Function** wird ausgeführt

## Berechtigungssystem

### Berechtigungsformat

Berechtigungen folgen dem `resource:action` Muster:

```
resource:action
   │       │
   │       └─ Aktion: create, read, update, delete, manage, *
   └───────── Ressource: posts, users, comments, roles, *
```

### Beispiele

| Berechtigung        | Beschreibung                                      |
| ------------------- | ------------------------------------------------- |
| `posts:create`      | Kann Beiträge erstellen                           |
| `posts:read`        | Kann Beiträge lesen                               |
| `posts:*`           | Kann alle Aktionen mit Beiträgen ausführen        |
| `*:read`            | Kann jede Ressource lesen                         |
| `*:*`               | Kann alle Aktionen mit allen Ressourcen ausführen |
| `users:ban`         | Kann Benutzer sperren (benutzerdefinierte Aktion) |
| `comments:moderate` | Kann Kommentare moderieren                        |

### Wildcards in Berechtigungen

Wildcards bieten leistungsstarke Berechtigungsvererbung:

```go
// Benutzer hat Berechtigung "posts:*"
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "posts:update")  // [+] TRUE
HasPermission(userID, "posts:delete")  // [+] TRUE
HasPermission(userID, "users:create")  // [X] FALSE

// Benutzer hat Berechtigung "*:read"
HasPermission(userID, "posts:read")    // [+] TRUE
HasPermission(userID, "users:read")    // [+] TRUE
HasPermission(userID, "posts:create")  // [X] FALSE

// Benutzer hat Berechtigung "*:*" (Admin mit vollem Zugriff)
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "users:delete")  // [+] TRUE
HasPermission(userID, "anything:anything") // [+] TRUE
```

## Rollensystem

### Systemrollen

4 vordefinierte Systemrollen mit unterschiedlichen Berechtigungsstufen:

| Rolle       | Anzeigename   | Berechtigungen           | Anwendungsfall                                 |
| ----------- | ------------- | ------------------------ | ---------------------------------------------- |
| `admin`     | Administrator | `*:*` (alle)             | Voller Systemzugriff                           |
| `moderator` | Moderator     | Inhaltsmoderation        | Inhalte prüfen und moderieren                  |
| `user`      | Benutzer      | Eigene Inhalte verwalten | Normale Benutzer                               |
| `guest`     | Gast          | Nur Lesezugriff          | Nicht authentifizierte/eingeschränkte Benutzer |

### Rollenberechtigungsverteilung

**Admin** (`*:*`):

- Voller Zugriff auf alles
- Kann nicht gelöscht werden (Systemrolle)

**Admin**:

```
users:create, users:read, users:update, users:delete, users:ban, users:suspend
roles:read, roles:assign
permissions:read
posts:*, comments:*, profiles:*
```

**Moderator**:

```
posts:read, posts:update, posts:delete
comments:read, comments:update, comments:delete, comments:moderate
users:read, users:suspend
```

**User**:

```
posts:create, posts:read, posts:update (eigene), posts:delete (eigene)
comments:create, comments:read, comments:update (eigene), comments:delete (eigene)
profiles:read, profiles:update (eigene)
```

**Guest**:

```
posts:read, comments:read, profiles:read
```

## Middleware-Methoden

### RequirePermission

Prüft, ob Benutzer **eine bestimmte Berechtigung** hat.

```go
func (m *AuthorizationMiddleware) RequirePermission(permission string) gin.HandlerFunc
```

**Verwendung:**

```go
router.POST("/posts",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

**Rückgabe:**

- `200 OK` - Benutzer hat Berechtigung, fährt mit Handler fort
- `401 Unauthorized` - Benutzer nicht authentifiziert
- `403 Forbidden` - Benutzer fehlt Berechtigung
- `500 Internal Server Error` - Datenbankfehler bei Berechtigungsprüfung

---

### RequireAnyPermission

Prüft, ob Benutzer **mindestens eine** der angegebenen Berechtigungen hat (ODER-Logik).

```go
func (m *AuthorizationMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc
```

**Verwendung:**

```go
// Erlauben, wenn Benutzer Kommentare lesen ODER moderieren kann
router.GET("/comments/flagged",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission("comments:read", "comments:moderate"),
    handler.GetFlaggedComments,
)
```

**Anwendungsfälle:**

- Alternative Berechtigungen (admin ODER moderator)
- Feature-Zugriff mit mehreren Einstiegspunkten
- Abgestufte Berechtigungserhöhung

---

### RequireAllPermissions

Prüft, ob Benutzer **alle** angegebenen Berechtigungen hat (UND-Logik).

```go
func (m *AuthorizationMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc
```

**Verwendung:**

```go
// Erfordert sowohl publish- als auch schedule-Berechtigungen
router.POST("/posts/schedule",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions("posts:create", "posts:schedule"),
    handler.SchedulePost,
)
```

**Anwendungsfälle:**

- Zusammengesetzte Operationen, die mehrere Berechtigungen erfordern
- Sensible Operationen, die mehrschichtige Prüfungen benötigen
- Feature-Kombinationen

---

### RequireRole

Prüft, ob Benutzer **eine bestimmte Rolle** nach Namen hat.

```go
func (m *AuthorizationMiddleware) RequireRole(roleName string) gin.HandlerFunc
```

**Verwendung:**

```go
// Nur Administratoren können zugreifen
router.GET("/admin/dashboard",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.GetAdminDashboard,
)
```

**Hinweis:** Bevorzugen Sie `RequirePermission` gegenüber `RequireRole` für bessere Flexibilität.

---

### RequireAnyRole

Prüft, ob Benutzer **mindestens eine** der angegebenen Rollen hat (ODER-Logik).

```go
func (m *AuthorizationMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc
```

**Verwendung:**

```go
// Admins ODER Moderatoren erlauben
router.GET("/moderation/queue",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyRole("admin", "moderator"),
    handler.GetModerationQueue,
)
```

**Anwendungsfälle:**

- Verwaltungsbereiche mit mehreren Rollenstufen
- Feature-Zugriff für ähnliche Rollen
- Legacy-Systeme, die von Rollen zu Berechtigungen migrieren

## Verwendungsbeispiele

### Beispiel 1: Grundlegender CRUD-Schutz

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")

    // Öffentlicher Lesezugriff (keine Auth erforderlich)
    posts.GET("", handler.ListPosts)
    posts.GET("/:id", handler.GetPost)

    // Operationen für Authentifizierte
    posts.Use(r.authMiddleware.RequireAuth())
    {
        // Spezifische Berechtigungen für jede Operation
        posts.POST("",
            r.authzMiddleware.RequirePermission("posts:create"),
            handler.CreatePost,
        )

        posts.PUT("/:id",
            r.authzMiddleware.RequirePermission("posts:update"),
            handler.UpdatePost,
        )

        posts.DELETE("/:id",
            r.authzMiddleware.RequirePermission("posts:delete"),
            handler.DeletePost,
        )
    }
}
```

### Beispiel 2: Nur-Admin-Endpunkte

```go
func (r *UserRouter) Setup(api *gin.RouterGroup) {
    users := api.Group("/users")
    users.Use(r.authMiddleware.RequireAuth())

    // Normale Benutzeroperationen
    users.GET("/me", handler.GetMe)
    users.PUT("/me", handler.UpdateProfile)

    // Nur-Admin-Operationen
    admin := users.Group("")
    admin.Use(r.authzMiddleware.RequirePermission("users:manage"))
    {
        admin.GET("", handler.ListAllUsers)
        admin.POST("/:id/ban", handler.BanUser)
        admin.POST("/:id/suspend", handler.SuspendUser)
    }
}
```

### Beispiel 3: Flexibler Zugriff mit Mehrfachberechtigungen

```go
func (r *CommentRouter) Setup(api *gin.RouterGroup) {
    comments := api.Group("/comments")

    // Kommentare anzeigen - jede dieser Berechtigungen funktioniert
    comments.GET("/:id",
        r.authMiddleware.RequireAuth(),
        r.authzMiddleware.RequireAnyPermission(
            "comments:read",
            "comments:moderate",
            "*:read",
        ),
        handler.GetComment,
    )

    // Kommentare moderieren - sowohl read als auch moderate erforderlich
    comments.POST("/:id/moderate",
        r.authMiddleware.RequireAuth(),
        r.authzMiddleware.RequireAllPermissions(
            "comments:read",
            "comments:moderate",
        ),
        handler.ModerateComment,
    )
}
```

### Beispiel 4: Rollenbasierter Dashboard-Zugriff

```go
func (r *DashboardRouter) Setup(api *gin.RouterGroup) {
    dashboards := api.Group("/dashboard")
    dashboards.Use(r.authMiddleware.RequireAuth())

    // Benutzer-Dashboard - jeder authentifizierte Benutzer
    dashboards.GET("/user", handler.GetUserDashboard)

    // Moderator-Dashboard - Moderatoren und Admins
    dashboards.GET("/moderator",
        r.authzMiddleware.RequireAnyRole("moderator", "admin"),
        handler.GetModeratorDashboard,
    )

    // Admin-Dashboard - nur Admins
    dashboards.GET("/admin",
        r.authzMiddleware.RequireRole("admin"),
        handler.GetAdminDashboard,
    )
}
```

### Beispiel 5: Komplexe Geschäftslogik

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")
    posts.Use(r.authMiddleware.RequireAuth())

    // Veröffentlichen erfordert create- und publish-Berechtigungen
    posts.POST("/:id/publish",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
        ),
        handler.PublishPost,
    )

    // Planen erfordert create, publish UND schedule
    posts.POST("/:id/schedule",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
            "posts:schedule",
        ),
        handler.SchedulePost,
    )

    // Hervorheben erfordert moderator ODER admin Rolle + feature-Berechtigung
    posts.POST("/:id/feature",
        r.authzMiddleware.RequireAnyRole("admin", "moderator"),
        r.authzMiddleware.RequirePermission("posts:feature"),
        handler.FeaturePost,
    )
}
```

### Beispiel 6: Migrations-Setup

Autorisierung im Router-Setup initialisieren:

```go
// cmd/api/main.go oder Router-Initialisierung
func setupRouters(
    authMiddleware *middleware.AuthMiddleware,
    authzMiddleware *middleware.AuthorizationMiddleware,
) *gin.Engine {
    r := gin.New()

    // Öffentliche Routen
    api := r.Group("/api/v1")

    // Auth-Routen (keine Autorisierung erforderlich)
    authRouter := router.NewAuthRouter(authHandler, authMiddleware)
    authRouter.Setup(api)

    // Geschützte Routen mit Autorisierung
    postRouter := router.NewPostRouter(postHandler, authMiddleware, authzMiddleware)
    postRouter.Setup(api)

    userRouter := router.NewUserRouter(userHandler, authMiddleware, authzMiddleware)
    userRouter.Setup(api)

    return r
}
```

## Best Practices

### 1. Verwenden Sie Immer Zuerst RequireAuth

Autorisierungs-Middleware benötigt Authentifizierungskontext:

```go
// [+] RICHTIG - Auth vor Autorisierung
router.POST("/posts",
    authMiddleware.RequireAuth(),           // Zuerst: Authentifizierung
    authzMiddleware.RequirePermission(...), // Dann: Autorisierung
    handler.CreatePost,
)

// [X] FALSCH - Autorisierung ohne Authentifizierung
router.POST("/posts",
    authzMiddleware.RequirePermission(...), // Schlägt fehl - keine user_id
    handler.CreatePost,
)
```

### 2. Bevorzugen Sie Berechtigungen Gegenüber Rollen

Berechtigungen bieten bessere Flexibilität und Wartbarkeit:

```go
// [+] BESSER - Berechtigungsbasiert (flexibel)
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:delete"),
    handler.DeletePost,
)

// [!] AKZEPTABEL aber weniger flexibel - Rollenbasiert
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.DeletePost,
)
```

**Warum?**

- Neue Rollen hinzufügen erfordert keine Codeänderungen
- Berechtigungen können ohne Code-Berührung neu zugewiesen werden
- Feinkörnigere Kontrolle

### 3. Verwenden Sie Beschreibende Berechtigungsnamen

```go
// [+] GUT - Klare Absicht
"posts:create"
"posts:publish"
"posts:feature"
"posts:schedule"

// [X] SCHLECHT - Unklar
"posts:manage"  // Was bedeutet "manage"?
"posts:admin"   // Zu allgemein
```

### 4. Nutzen Sie Wildcards für Admin-Rollen

```go
// In Ihrer Migration/Seed
INSERT INTO permissions (resource, action) VALUES
    ('*', '*'),           -- Admin: alles
    ('posts', '*'),       -- Content-Admin: alle Post-Operationen
    ('*', 'read');        -- Viewer: alles lesen
```

### 5. Gruppieren Sie Verwandte Berechtigungen

```go
// Nach Funktionsbereich gruppieren
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Alle Post-Schreiboperationen erfordern posts:* oder posts:write
write := posts.Group("")
write.Use(authzMiddleware.RequirePermission("posts:write"))
{
    write.POST("", handler.CreatePost)
    write.PUT("/:id", handler.UpdatePost)
    write.DELETE("/:id", handler.DeletePost)
}

// Öffentliche Leseoperationen
posts.GET("", handler.ListPosts)
posts.GET("/:id", handler.GetPost)
```

### 6. Behandeln Sie Besitzerressourcen im Handler

Verwenden Sie keine Autorisierungs-Middleware für Besitzerprüfungen:

```go
// [+] RICHTIG - Besitz im Handler prüfen
func (h *PostHandler) UpdatePost(c *gin.Context) {
    userID := middleware.GetUserIDOrPanic(c)
    postID := c.Param("id")

    post, err := h.postUC.GetByID(c.Request.Context(), postID)
    if err != nil {
        response.Error(c, http.StatusNotFound, "post not found", err)
        return
    }

    // Besitz ODER Admin-Berechtigung prüfen
    if post.UserID != userID {
        hasAdmin, _ := h.roleUC.HasPermission(c.Request.Context(), userID, "posts:*")
        if !hasAdmin {
            response.Error(c, http.StatusForbidden, "can only update own posts", nil)
            return
        }
    }

    // Update fortsetzen...
}

// [X] FALSCH - Versuch, Besitz in Middleware zu prüfen
// Middleware hat keinen Zugriff auf Ressourcendetails
```

### 7. Verwenden Sie RequireAny für Fallback-Berechtigungen

```go
// Operation erlauben, wenn Benutzer spezifische Berechtigung ODER Admin ist
router.POST("/posts/:id/feature",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission(
        "posts:feature",  // Spezifische Berechtigung
        "posts:*",        // Voller Post-Zugriff
        "*:*",            // Admin
    ),
    handler.FeaturePost,
)
```

## Häufige Muster

### Muster 1: Admin-Override

Admins erlauben, Besitzerprüfungen zu umgehen:

```go
// Jede Admin-Berechtigung überschreibt Besitz
authzMiddleware.RequireAnyPermission(
    "posts:update",  // Normale Benutzerberechtigung
    "posts:*",       // Post-Admin
    "*:*",           // Admin
)
```

### Muster 2: Abgestufte Berechtigungen

Verschiedene Berechtigungsstufen für dieselbe Ressource:

```go
// Stufe 1: Grundlegendes Lesen
authzMiddleware.RequirePermission("posts:read")

// Stufe 2: Lesen + Schreiben
authzMiddleware.RequireAllPermissions("posts:read", "posts:write")

// Stufe 3: Voller Zugriff
authzMiddleware.RequirePermission("posts:*")
```

### Muster 3: Ressourcenübergreifende Berechtigungen

Operationen, die mehrere Ressourcen betreffen:

```go
// Post veröffentlichen kann sowohl Post- als auch Medien-Berechtigungen erfordern
router.POST("/posts/:id/publish",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions(
        "posts:publish",
        "media:attach",  // Wenn Post Bilder enthält
    ),
    handler.PublishPost,
)
```

### Muster 4: Bedingte Autorisierung

Verschiedene Berechtigungen für verschiedene Endpunkte:

```go
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Entwurfs-Posts - nur create-Berechtigung
posts.POST("/drafts",
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreateDraft,
)

// Veröffentlichte Posts - create + publish Berechtigungen
posts.POST("/publish",
    authzMiddleware.RequireAllPermissions("posts:create", "posts:publish"),
    handler.CreateAndPublish,
)
```

## Fehlerbehandlung

### HTTP-Statuscodes

| Status | Bedeutung             | Grund                                             |
| ------ | --------------------- | ------------------------------------------------- |
| 401    | Unauthorized          | Benutzer nicht authentifiziert (kein JWT)         |
| 403    | Forbidden             | Benutzer authentifiziert, aber Berechtigung fehlt |
| 500    | Internal Server Error | Datenbankfehler bei Berechtigungsprüfung          |

### Fehlerantwortformat

```json
{
  "error": "insufficient permissions",
  "message": "You don't have permission to perform this action"
}
```

### Client-seitige Behandlung

```typescript
// TypeScript/JavaScript Beispiel
try {
  const response = await fetch("/api/v1/posts", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(postData),
  });

  if (response.status === 401) {
    // Zu Login umleiten
    window.location.href = "/login";
  } else if (response.status === 403) {
    // "Unzureichende Berechtigungen" Nachricht anzeigen
    showError("You do not have permission to create posts");
  } else if (response.ok) {
    // Erfolg
    const post = await response.json();
  }
} catch (error) {
  console.error("Request failed:", error);
}
```

## Autorisierung Testen

### Unit-Tests für Middleware

```go
func TestAuthorizationMiddleware_RequirePermission(t *testing.T) {
    // Setup
    mockRoleUC := mocks.NewMockRoleUseCase(t)
    authzMiddleware := middleware.NewAuthorizationMiddleware(mockRoleUC)

    t.Run("allows user with permission", func(t *testing.T) {
        // Test-Kontext mit user_id erstellen
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Set("user_id", testUserID)

        // Mock-Berechtigungsprüfung - gibt true zurück
        mockRoleUC.EXPECT().
            HasPermission(mock.Anything, testUserID, "posts:create").
            Return(true, nil)

        // Handler-Kette erstellen
        handler := authzMiddleware.RequirePermission("posts:create")(func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })

        // Ausführen
        handler(c)

        // Assert
        assert.Equal(t, 200, w.Code)
    })

    t.Run("denies user without permission", func(t *testing.T) {
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Set("user_id", testUserID)

        mockRoleUC.EXPECT().
            HasPermission(mock.Anything, testUserID, "posts:create").
            Return(false, nil)

        handler := authzMiddleware.RequirePermission("posts:create")(func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })

        handler(c)

        assert.Equal(t, 403, w.Code)
    })
}
```

### Integrationstests

```go
func TestPostEndpoints_Authorization(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    // Testbenutzer mit verschiedenen Rollen erstellen
    adminUser := helpers.UserFixture(t, testDB.DB)
    regularUser := helpers.UserFixture(t, testDB.DB)

    // Rollen zuweisen
    assignRole(t, testDB, adminUser.ID, "admin")
    assignRole(t, testDB, regularUser.ID, "user")

    // Tokens generieren
    adminToken := generateToken(t, adminUser.ID)
    userToken := generateToken(t, regularUser.ID)

    t.Run("admin can delete any post", func(t *testing.T) {
        post := createTestPost(t, testDB, regularUser.ID)

        req := httptest.NewRequest("DELETE", "/api/v1/posts/"+post.ID, nil)
        req.Header.Set("Authorization", "Bearer "+adminToken)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, 200, w.Code)
    })

    t.Run("user cannot delete others' posts", func(t *testing.T) {
        post := createTestPost(t, testDB, adminUser.ID)

        req := httptest.NewRequest("DELETE", "/api/v1/posts/"+post.ID, nil)
        req.Header.Set("Authorization", "Bearer "+userToken)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, 403, w.Code)
    })
}
```

### Manuelles Testen mit curl

```bash
# 1. Login und Token erhalten
TOKEN=$(curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# 2. Geschützten Endpunkt testen
curl -X POST http://localhost:8081/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Post","content":"Content"}'

# Erwartete Antworten:
# 200 OK - Erfolg
# 401 Unauthorized - Ungültiges/fehlendes Token
# 403 Forbidden - Unzureichende Berechtigungen
```

## Fehlerbehebung

### Problem: 401 Unauthorized bei geschütztem Endpunkt

**Symptome:**

```json
{
  "error": "user not authenticated"
}
```

**Ursachen:**

1. Fehlende `RequireAuth()` Middleware vor `RequirePermission()`
2. Ungültiges JWT-Token
3. Abgelaufenes Token

**Lösung:**

```go
// Sicherstellen, dass Auth-Middleware zuerst angewendet wird
router.POST("/posts",
    authMiddleware.RequireAuth(),  // ← Muss vor Autorisierung sein
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

### Problem: 403 Forbidden für Admin

**Symptome:**
Admin-Benutzer erhält 403 bei Endpunkten, auf die er Zugriff haben sollte.

**Ursachen:**

1. Wildcard-Berechtigung (`*:*`) wird nicht korrekt geprüft
2. Rolle nicht dem Benutzer zugewiesen
3. Berechtigung nicht der Rolle zugewiesen

**Lösung:**

```sql
-- Prüfen, dass Admin Wildcard-Berechtigung hat
SELECT r.name, p.resource, p.action
FROM roles r
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions p ON p.id = rp.permission_id
WHERE r.name = 'admin';

-- Sollte zurückgeben: name='admin', resource='*', action='*'

-- Prüfen, dass Benutzer Admin-Rolle hat
SELECT u.email, r.name
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
WHERE u.id = '<user_uuid>';
```

### Problem: Datenbankleistung bei Berechtigungsprüfungen

**Symptome:**
Langsame Antwortzeiten bei geschützten Endpunkten.

**Lösung:**
Caching in IRoleUseCase implementieren:

```go
// Redis/In-Memory-Cache für Berechtigungsprüfungen verwenden
func (uc *IRoleUseCase) HasPermission(ctx context.Context, userID uuidv7.UUID, permission string) (bool, error) {
    // Cache zuerst prüfen
    cacheKey := fmt.Sprintf("user:%s:permission:%s", userID, permission)
    if cached, found := uc.cache.Get(cacheKey); found {
        return cached.(bool), nil
    }

    // Datenbankabfrage
    hasPermission, err := uc.repo.HasPermission(ctx, userID, permission)
    if err != nil {
        return false, err
    }

    // 5 Minuten cachen
    uc.cache.Set(cacheKey, hasPermission, 5*time.Minute)

    return hasPermission, nil
}
```

## Zusammenfassung

Die Autorisierungs-Middleware bietet leistungsstarke, flexible Zugriffskontrolle für Ihre API:

[+] **Berechtigungsbasiert** - Feinkörnige Kontrolle im Format `resource:action`
[+] **Wildcard-Unterstützung** - Leistungsstarke Vererbung mit `*` Mustern
[+] **Zusammengesetzte Prüfungen** - UND/ODER-Logik für komplexe Anforderungen
[+] **Rollen-Shortcuts** - Schnelle rollenbasierte Prüfungen bei Bedarf
[+] **Saubere Architektur** - Trennt Autorisierung von Authentifizierung
[+] **Produktionsbereit** - Bewährte Fehlerbehandlung und Leistung

**Schnellreferenz:**

```go
// Einzelne Berechtigungsprüfung
RequirePermission("posts:create")

// Eine von mehreren Berechtigungen (ODER)
RequireAnyPermission("posts:read", "posts:*", "*:*")

// Alle von mehreren Berechtigungen (UND)
RequireAllPermissions("posts:create", "posts:publish")

// Rollenbasierte Prüfung
RequireRole("admin")

// Eine von mehreren Rollen (ODER)
RequireAnyRole("admin", "moderator")
```

Für weitere Informationen siehe:

- [RBAC Implementation](RBAC_IMPLEMENTATION.md)
- [Testing Guide](TESTING_GUIDE.de.md)
- [API Documentation](../README.de.md)
