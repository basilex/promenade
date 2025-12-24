# Guía de Middleware de Autorización

Guía completa para usar el middleware de autorización RBAC (Role-Based Access Control) en Promenade.

## Tabla de Contenidos

- [Descripción General](#descripción-general)
- [Arquitectura](#arquitectura)
- [Sistema de Permisos](#sistema-de-permisos)
- [Sistema de Roles](#sistema-de-roles)
- [Métodos de Middleware](#métodos-de-middleware)
- [Ejemplos de Uso](#ejemplos-de-uso)
- [Mejores Prácticas](#mejores-prácticas)
- [Patrones Comunes](#patrones-comunes)
- [Manejo de Errores](#manejo-de-errores)
- [Probando Autorización](#probando-autorización)

## Descripción General

El Middleware de Autorización proporciona **control de acceso flexible y granular** para endpoints de API usando un sistema RBAC basado en permisos. Soporta:

- [+] **Verificaciones basadas en permisos** - Control granular en formato `resource:action`
- [+] **Verificaciones basadas en roles** - Verificaciones rápidas de roles de usuario (admin, moderator, etc.)
- [+] **Permisos con comodines** - Patrones `*:*`, `posts:*`, `*:read`
- [+] **Verificaciones compuestas** - RequireAny, RequireAll para lógica compleja
- [+] **Separación clara** - Funciona independientemente del middleware de autenticación

## Arquitectura

**Flujo de Solicitud:**

1. **Solicitud HTTP** llega
   ↓
2. **RequireAuth Middleware**
   - Valida token JWT
   - Establece user_id en contexto
     ↓
3. **RequirePermission Middleware**
   - Obtiene user_id del contexto
   - Consulta roles del usuario
   - Verifica permisos del rol (con soporte de comodines)
   - Permite/Deniega solicitud
     ↓
4. **Handler Function** se ejecuta

## Sistema de Permisos

### Formato de Permisos

Los permisos siguen el patrón `resource:action`:

```
resource:action
   │       │
   │       └─ Acción: create, read, update, delete, manage, *
   └───────── Recurso: posts, users, comments, roles, *
```

### Ejemplos

| Permiso             | Descripción                                           |
| ------------------- | ----------------------------------------------------- |
| `posts:create`      | Puede crear posts                                     |
| `posts:read`        | Puede leer posts                                      |
| `posts:*`           | Puede realizar cualquier acción con posts             |
| `*:read`            | Puede leer cualquier recurso                          |
| `*:*`               | Puede realizar cualquier acción con cualquier recurso |
| `users:ban`         | Puede banear usuarios (acción personalizada)          |
| `comments:moderate` | Puede moderar comentarios                             |

### Comodines en Permisos

Los comodines proporcionan herencia poderosa de permisos:

```go
// Usuario tiene permiso "posts:*"
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "posts:update")  // [+] TRUE
HasPermission(userID, "posts:delete")  // [+] TRUE
HasPermission(userID, "users:create")  // [X] FALSE

// Usuario tiene permiso "*:read"
HasPermission(userID, "posts:read")    // [+] TRUE
HasPermission(userID, "users:read")    // [+] TRUE
HasPermission(userID, "posts:create")  // [X] FALSE

// Usuario tiene permiso "*:*" (admin con acceso total)
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "users:delete")  // [+] TRUE
HasPermission(userID, "anything:anything") // [+] TRUE
```

## Sistema de Roles

### Roles del Sistema

4 roles de sistema predefinidos con diferentes niveles de permisos:

| Rol         | Nombre de Visualización | Permisos                   | Caso de Uso                           |
| ----------- | ----------------------- | -------------------------- | ------------------------------------- |
| `admin`     | Administrador           | `*:*` (todos)              | Acceso total al sistema               |
| `moderator` | Moderador               | Moderación de contenido    | Revisar y moderar contenido           |
| `user`      | Usuario                 | Gestionar propio contenido | Usuarios normales                     |
| `guest`     | Invitado                | Solo lectura               | Usuarios no autenticados/restringidos |

### Distribución de Permisos por Rol

**Admin** (`*:*`):

- Acceso total a todo
- No puede ser eliminado (rol del sistema)

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
posts:create, posts:read, posts:update (propios), posts:delete (propios)
comments:create, comments:read, comments:update (propios), comments:delete (propios)
profiles:read, profiles:update (propios)
```

**Guest**:

```
posts:read, comments:read, profiles:read
```

## Métodos de Middleware

### RequirePermission

Verifica si el usuario tiene **un permiso específico**.

```go
func (m *AuthorizationMiddleware) RequirePermission(permission string) gin.HandlerFunc
```

**Uso:**

```go
router.POST("/posts",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

**Devuelve:**

- `200 OK` - Usuario tiene permiso, continúa al handler
- `401 Unauthorized` - Usuario no autenticado
- `403 Forbidden` - Usuario no tiene permiso
- `500 Internal Server Error` - Error de base de datos al verificar permisos

---

### RequireAnyPermission

Verifica si el usuario tiene **al menos uno** de los permisos especificados (lógica O).

```go
func (m *AuthorizationMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc
```

**Uso:**

```go
// Permitir si el usuario puede leer O moderar comentarios
router.GET("/comments/flagged",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission("comments:read", "comments:moderate"),
    handler.GetFlaggedComments,
)
```

**Casos de Uso:**

- Permisos alternativos (admin O moderator)
- Acceso a funciones con múltiples puntos de entrada
- Elevación gradual de permisos

---

### RequireAllPermissions

Verifica si el usuario tiene **todos** los permisos especificados (lógica Y).

```go
func (m *AuthorizationMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc
```

**Uso:**

```go
// Requiere permisos publish Y schedule
router.POST("/posts/schedule",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions("posts:create", "posts:schedule"),
    handler.SchedulePost,
)
```

**Casos de Uso:**

- Operaciones compuestas que requieren múltiples permisos
- Operaciones sensibles que necesitan verificaciones en múltiples capas
- Combinaciones de funciones

---

### RequireRole

Verifica si el usuario tiene **un rol específico** por nombre.

```go
func (m *AuthorizationMiddleware) RequireRole(roleName string) gin.HandlerFunc
```

**Uso:**

```go
// Solo administradores pueden acceder
router.GET("/admin/dashboard",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.GetAdminDashboard,
)
```

**Nota:** Prefiera `RequirePermission` sobre `RequireRole` para mejor flexibilidad.

---

### RequireAnyRole

Verifica si el usuario tiene **al menos uno** de los roles especificados (lógica O).

```go
func (m *AuthorizationMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc
```

**Uso:**

```go
// Permitir admins O moderators
router.GET("/moderation/queue",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyRole("admin", "moderator"),
    handler.GetModerationQueue,
)
```

**Casos de Uso:**

- Áreas administrativas con múltiples niveles de roles
- Acceso a funciones para roles similares
- Sistemas heredados migrando de roles a permisos

## Ejemplos de Uso

### Ejemplo 1: Protección CRUD Básica

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")

    // Acceso público de lectura (sin auth necesaria)
    posts.GET("", handler.ListPosts)
    posts.GET("/:id", handler.GetPost)

    // Operaciones para autenticados
    posts.Use(r.authMiddleware.RequireAuth())
    {
        // Permisos específicos para cada operación
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

### Ejemplo 2: Endpoints Solo para Admin

```go
func (r *UserRouter) Setup(api *gin.RouterGroup) {
    users := api.Group("/users")
    users.Use(r.authMiddleware.RequireAuth())

    // Operaciones de usuarios normales
    users.GET("/me", handler.GetMe)
    users.PUT("/me", handler.UpdateProfile)

    // Operaciones solo para admin
    admin := users.Group("")
    admin.Use(r.authzMiddleware.RequirePermission("users:manage"))
    {
        admin.GET("", handler.ListAllUsers)
        admin.POST("/:id/ban", handler.BanUser)
        admin.POST("/:id/suspend", handler.SuspendUser)
    }
}
```

### Ejemplo 3: Acceso Flexible con Múltiples Permisos

```go
func (r *CommentRouter) Setup(api *gin.RouterGroup) {
    comments := api.Group("/comments")

    // Ver comentarios - cualquiera de estos permisos funciona
    comments.GET("/:id",
        r.authMiddleware.RequireAuth(),
        r.authzMiddleware.RequireAnyPermission(
            "comments:read",
            "comments:moderate",
            "*:read",
        ),
        handler.GetComment,
    )

    // Moderar comentarios - requiere read Y moderate
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

### Ejemplo 4: Acceso al Panel Basado en Rol

```go
func (r *DashboardRouter) Setup(api *gin.RouterGroup) {
    dashboards := api.Group("/dashboard")
    dashboards.Use(r.authMiddleware.RequireAuth())

    // Panel de usuario - cualquier usuario autenticado
    dashboards.GET("/user", handler.GetUserDashboard)

    // Panel de moderador - moderators y admins
    dashboards.GET("/moderator",
        r.authzMiddleware.RequireAnyRole("moderator", "admin"),
        handler.GetModeratorDashboard,
    )

    // Panel de admin - solo admins
    dashboards.GET("/admin",
        r.authzMiddleware.RequireRole("admin"),
        handler.GetAdminDashboard,
    )
}
```

### Ejemplo 5: Lógica de Negocio Compleja

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")
    posts.Use(r.authMiddleware.RequireAuth())

    // Publicar requiere permisos create y publish
    posts.POST("/:id/publish",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
        ),
        handler.PublishPost,
    )

    // Programar requiere create, publish Y schedule
    posts.POST("/:id/schedule",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
            "posts:schedule",
        ),
        handler.SchedulePost,
    )

    // Destacar requiere rol moderator O admin + permiso feature
    posts.POST("/:id/feature",
        r.authzMiddleware.RequireAnyRole("admin", "moderator"),
        r.authzMiddleware.RequirePermission("posts:feature"),
        handler.FeaturePost,
    )
}
```

### Ejemplo 6: Configuración de Migración

Inicializar autorización en la configuración del router:

```go
// cmd/api/main.go o inicialización del router
func setupRouters(
    authMiddleware *middleware.AuthMiddleware,
    authzMiddleware *middleware.AuthorizationMiddleware,
) *gin.Engine {
    r := gin.New()

    // Rutas públicas
    api := r.Group("/api/v1")

    // Rutas de Auth (sin autorización necesaria)
    authRouter := router.NewAuthRouter(authHandler, authMiddleware)
    authRouter.Setup(api)

    // Rutas protegidas con autorización
    postRouter := router.NewPostRouter(postHandler, authMiddleware, authzMiddleware)
    postRouter.Setup(api)

    userRouter := router.NewUserRouter(userHandler, authMiddleware, authzMiddleware)
    userRouter.Setup(api)

    return r
}
```

## Mejores Prácticas

### 1. Siempre Use RequireAuth Primero

El middleware de autorización requiere contexto de autenticación:

```go
// [+] CORRECTO - Auth antes de autorización
router.POST("/posts",
    authMiddleware.RequireAuth(),           // Primero: autenticación
    authzMiddleware.RequirePermission(...), // Después: autorización
    handler.CreatePost,
)

// [X] INCORRECTO - Autorización sin autenticación
router.POST("/posts",
    authzMiddleware.RequirePermission(...), // Falla - sin user_id
    handler.CreatePost,
)
```

### 2. Prefiera Permisos sobre Roles

Los permisos proporcionan mejor flexibilidad y mantenibilidad:

```go
// [+] MEJOR - Basado en permisos (flexible)
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:delete"),
    handler.DeletePost,
)

// [!] ACEPTABLE pero menos flexible - Basado en roles
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.DeletePost,
)
```

**¿Por qué?**

- Agregar nuevos roles no requiere cambios de código
- Los permisos pueden reasignarse sin tocar el código
- Control más granular

### 3. Use Nombres Descriptivos para Permisos

```go
// [+] BUENO - Intención clara
"posts:create"
"posts:publish"
"posts:feature"
"posts:schedule"

// [X] MALO - Poco claro
"posts:manage"  // ¿Qué significa "manage"?
"posts:admin"   // Demasiado genérico
```

### 4. Aproveche Comodines para Roles de Admin

```go
// En su migración/seed
INSERT INTO permissions (resource, action) VALUES
    ('*', '*'),           -- Admin: todo
    ('posts', '*'),       -- Admin de contenido: todas las operaciones de posts
    ('*', 'read');        -- Visor: leer todo
```

### 5. Agrupe Permisos Relacionados

```go
// Agrupar por área funcional
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Todas las operaciones de escritura de posts requieren posts:* o posts:write
write := posts.Group("")
write.Use(authzMiddleware.RequirePermission("posts:write"))
{
    write.POST("", handler.CreatePost)
    write.PUT("/:id", handler.UpdatePost)
    write.DELETE("/:id", handler.DeletePost)
}

// Operaciones públicas de lectura
posts.GET("", handler.ListPosts)
posts.GET("/:id", handler.GetPost)
```

### 6. Maneje Recursos del Propietario en el Handler

No use middleware de autorización para verificaciones de propietario:

```go
// [+] CORRECTO - Verificar propiedad en el handler
func (h *PostHandler) UpdatePost(c *gin.Context) {
    userID := middleware.GetUserIDOrPanic(c)
    postID := c.Param("id")

    post, err := h.postUC.GetByID(c.Request.Context(), postID)
    if err != nil {
        response.Error(c, http.StatusNotFound, "post not found", err)
        return
    }

    // Verificar propiedad O permiso de admin
    if post.UserID != userID {
        hasAdmin, _ := h.roleUC.HasPermission(c.Request.Context(), userID, "posts:*")
        if !hasAdmin {
            response.Error(c, http.StatusForbidden, "can only update own posts", nil)
            return
        }
    }

    // Continuar actualización...
}

// [X] INCORRECTO - Intentar verificar propiedad en middleware
// Middleware no tiene acceso a los detalles del recurso
```

### 7. Use RequireAny para Permisos de Respaldo

```go
// Permitir operación si el usuario tiene permiso específico O es admin
router.POST("/posts/:id/feature",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission(
        "posts:feature",  // Permiso específico
        "posts:*",        // Acceso total a posts
        "*:*",            // Admin
    ),
    handler.FeaturePost,
)
```

## Patrones Comunes

### Patrón 1: Sobrescritura de Admin

Permitir que los admins eviten verificaciones de propiedad:

```go
// Cualquier permiso de admin sobrescribe propiedad
authzMiddleware.RequireAnyPermission(
    "posts:update",  // Permiso de usuario normal
    "posts:*",       // Admin de posts
    "*:*",           // Admin
)
```

### Patrón 2: Permisos Graduados

Diferentes niveles de permisos para el mismo recurso:

```go
// Nivel 1: Lectura básica
authzMiddleware.RequirePermission("posts:read")

// Nivel 2: Lectura + escritura
authzMiddleware.RequireAllPermissions("posts:read", "posts:write")

// Nivel 3: Acceso total
authzMiddleware.RequirePermission("posts:*")
```

### Patrón 3: Permisos entre Recursos

Operaciones que afectan múltiples recursos:

```go
// Publicar post puede requerir permisos de post Y medios
router.POST("/posts/:id/publish",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions(
        "posts:publish",
        "media:attach",  // Si el post contiene imágenes
    ),
    handler.PublishPost,
)
```

### Patrón 4: Autorización Condicional

Permisos diferentes para diferentes endpoints:

```go
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Posts de borrador - solo permiso create
posts.POST("/drafts",
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreateDraft,
)

// Posts publicados - permisos create + publish
posts.POST("/publish",
    authzMiddleware.RequireAllPermissions("posts:create", "posts:publish"),
    handler.CreateAndPublish,
)
```

## Manejo de Errores

### Códigos de Estado HTTP

| Estado | Significado           | Motivo                                       |
| ------ | --------------------- | -------------------------------------------- |
| 401    | Unauthorized          | Usuario no autenticado (sin JWT)             |
| 403    | Forbidden             | Usuario autenticado pero sin permiso         |
| 500    | Internal Server Error | Error de base de datos al verificar permisos |

### Formato de Respuesta de Error

```json
{
  "error": "insufficient permissions",
  "message": "You don't have permission to perform this action"
}
```

### Manejo del Lado del Cliente

```typescript
// Ejemplo TypeScript/JavaScript
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
    // Redirigir a login
    window.location.href = "/login";
  } else if (response.status === 403) {
    // Mostrar mensaje de "permisos insuficientes"
    showError("You do not have permission to create posts");
  } else if (response.ok) {
    // Éxito
    const post = await response.json();
  }
} catch (error) {
  console.error("Request failed:", error);
}
```

## Probando Autorización

### Pruebas Unitarias de Middleware

```go
func TestAuthorizationMiddleware_RequirePermission(t *testing.T) {
    // Configuración
    mockRoleUC := mocks.NewMockRoleUseCase(t)
    authzMiddleware := middleware.NewAuthorizationMiddleware(mockRoleUC)

    t.Run("allows user with permission", func(t *testing.T) {
        // Crear contexto de prueba con user_id
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Set("user_id", testUserID)

        // Mock de verificación de permiso - devuelve true
        mockRoleUC.EXPECT().
            HasPermission(mock.Anything, testUserID, "posts:create").
            Return(true, nil)

        // Crear cadena de handler
        handler := authzMiddleware.RequirePermission("posts:create")(func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })

        // Ejecutar
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

### Pruebas de Integración

```go
func TestPostEndpoints_Authorization(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    // Crear usuarios de prueba con diferentes roles
    adminUser := helpers.UserFixture(t, testDB.DB)
    regularUser := helpers.UserFixture(t, testDB.DB)

    // Asignar roles
    assignRole(t, testDB, adminUser.ID, "admin")
    assignRole(t, testDB, regularUser.ID, "user")

    // Generar tokens
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

### Pruebas Manuales con curl

```bash
# 1. Login y obtener token
TOKEN=$(curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# 2. Probar endpoint protegido
curl -X POST http://localhost:8081/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Post","content":"Content"}'

# Respuestas esperadas:
# 200 OK - Éxito
# 401 Unauthorized - Token inválido/ausente
# 403 Forbidden - Permisos insuficientes
```

## Solución de Problemas

### Problema: 401 Unauthorized en Endpoint Protegido

**Síntomas:**

```json
{
  "error": "user not authenticated"
}
```

**Causas:**

1. Middleware `RequireAuth()` ausente antes de `RequirePermission()`
2. Token JWT inválido
3. Token expirado

**Solución:**

```go
// Asegurar que middleware de auth se aplica primero
router.POST("/posts",
    authMiddleware.RequireAuth(),  // ← Debe estar antes de autorización
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

### Problema: 403 Forbidden para Admin

**Síntomas:**
Usuario admin recibe 403 en endpoints a los que debería tener acceso.

**Causas:**

1. Permiso comodín (`*:*`) no se está verificando adecuadamente
2. Rol no asignado al usuario
3. Permiso no asignado al rol

**Solución:**

```sql
-- Verificar que admin tiene permiso comodín
SELECT r.name, p.resource, p.action
FROM roles r
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions p ON p.id = rp.permission_id
WHERE r.name = 'admin';

-- Debe devolver: name='admin', resource='*', action='*'

-- Verificar que usuario tiene rol admin
SELECT u.email, r.name
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
WHERE u.id = '<user_uuid>';
```

### Problema: Rendimiento de Base de Datos con Verificaciones de Permisos

**Síntomas:**
Tiempos de respuesta lentos en endpoints protegidos.

**Solución:**
Implementar caché en RoleUseCase:

```go
// Usar Redis/caché en memoria para verificaciones de permisos
func (uc *RoleUseCase) HasPermission(ctx context.Context, userID uuidv7.UUID, permission string) (bool, error) {
    // Verificar caché primero
    cacheKey := fmt.Sprintf("user:%s:permission:%s", userID, permission)
    if cached, found := uc.cache.Get(cacheKey); found {
        return cached.(bool), nil
    }

    // Consulta a base de datos
    hasPermission, err := uc.repo.HasPermission(ctx, userID, permission)
    if err != nil {
        return false, err
    }

    // Cachear por 5 minutos
    uc.cache.Set(cacheKey, hasPermission, 5*time.Minute)

    return hasPermission, nil
}
```

## Resumen

El Middleware de Autorización proporciona control de acceso potente y flexible para su API:

[+] **Basado en permisos** - Control granular en formato `resource:action`  
[+] **Soporte de comodines** - Herencia potente con patrones `*`  
[+] **Verificaciones compuestas** - Lógica Y/O para requisitos complejos  
[+] **Atajos de roles** - Verificaciones rápidas basadas en roles cuando sea necesario  
[+] **Arquitectura limpia** - Separa autorización de autenticación  
[+] **Listo para producción** - Manejo de errores y rendimiento probados

**Referencia Rápida:**

```go
// Verificación de permiso único
RequirePermission("posts:create")

// Cualquiera de múltiples permisos (O)
RequireAnyPermission("posts:read", "posts:*", "*:*")

// Todos los múltiples permisos (Y)
RequireAllPermissions("posts:create", "posts:publish")

// Verificación basada en rol
RequireRole("admin")

// Cualquiera de múltiples roles (O)
RequireAnyRole("admin", "moderator")
```

Para más información, consulte:

- [RBAC Implementation](RBAC_IMPLEMENTATION.md)
- [Testing Guide](TESTING_GUIDE.es.md)
- [API Documentation](../README.es.md)
