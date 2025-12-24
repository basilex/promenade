# Guia de Middleware de Autorização

Guia completo para uso do middleware de autorização RBAC (Role-Based Access Control) no Promenade.

## Índice

- [Visão Geral](#visão-geral)
- [Arquitetura](#arquitetura)
- [Sistema de Permissões](#sistema-de-permissões)
- [Sistema de Funções](#sistema-de-funções)
- [Métodos de Middleware](#métodos-de-middleware)
- [Exemplos de Uso](#exemplos-de-uso)
- [Melhores Práticas](#melhores-práticas)
- [Padrões Comuns](#padrões-comuns)
- [Tratamento de Erros](#tratamento-de-erros)
- [Testando Autorização](#testando-autorização)

## Visão Geral

O Middleware de Autorização fornece **controle de acesso flexível e granular** para endpoints de API usando um sistema RBAC baseado em permissões. Suporta:

- [+] **Verificações baseadas em permissões** - Controle granular no formato `resource:action`
- [+] **Verificações baseadas em funções** - Verificações rápidas de funções de usuário (admin, moderator etc.)
- [+] **Permissões com wildcards** - Padrões `*:*`, `posts:*`, `*:read`
- [+] **Verificações compostas** - RequireAny, RequireAll para lógica complexa
- [+] **Separação clara** - Funciona independentemente do middleware de autenticação

## Arquitetura

**Fluxo de Requisição:**

1. **Requisição HTTP** chega
   ↓
2. **RequireAuth Middleware**
   - Valida token JWT
   - Define user_id no contexto
     ↓
3. **RequirePermission Middleware**
   - Obtém user_id do contexto
   - Consulta funções do usuário
   - Verifica permissões da função (com suporte a wildcards)
   - Permite/Nega requisição
     ↓
4. **Handler Function** executa

## Sistema de Permissões

### Formato de Permissões

Permissões seguem o padrão `resource:action`:

```
resource:action
   │       │
   │       └─ Ação: create, read, update, delete, manage, *
   └───────── Recurso: posts, users, comments, roles, *
```

### Exemplos

| Permissão           | Descrição                                        |
| ------------------- | ------------------------------------------------ |
| `posts:create`      | Pode criar posts                                 |
| `posts:read`        | Pode ler posts                                   |
| `posts:*`           | Pode executar qualquer ação com posts            |
| `*:read`            | Pode ler qualquer recurso                        |
| `*:*`               | Pode executar qualquer ação com qualquer recurso |
| `users:ban`         | Pode banir usuários (ação personalizada)         |
| `comments:moderate` | Pode moderar comentários                         |

### Wildcards em Permissões

Wildcards fornecem herança poderosa de permissões:

```go
// Usuário tem permissão "posts:*"
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "posts:update")  // [+] TRUE
HasPermission(userID, "posts:delete")  // [+] TRUE
HasPermission(userID, "users:create")  // [X] FALSE

// Usuário tem permissão "*:read"
HasPermission(userID, "posts:read")    // [+] TRUE
HasPermission(userID, "users:read")    // [+] TRUE
HasPermission(userID, "posts:create")  // [X] FALSE

// Usuário tem permissão "*:*" (admin com acesso total)
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "users:delete")  // [+] TRUE
HasPermission(userID, "anything:anything") // [+] TRUE
```

## Sistema de Funções

### Funções do Sistema

4 funções de sistema predefinidas com diferentes níveis de permissões:

| Função      | Nome de Exibição | Permissões                 | Caso de Uso                         |
| ----------- | ---------------- | -------------------------- | ----------------------------------- |
| `admin`     | Administrador    | `*:*` (todas)              | Acesso total ao sistema             |
| `moderator` | Moderador        | Moderação de conteúdo      | Revisar e moderar conteúdo          |
| `user`      | Usuário          | Gerenciar próprio conteúdo | Usuários normais                    |
| `guest`     | Convidado        | Apenas leitura             | Usuários não autenticados/restritos |

### Distribuição de Permissões por Função

**Admin** (`*:*`):

- Acesso total a tudo
- Não pode ser excluído (função de sistema)

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
posts:create, posts:read, posts:update (próprios), posts:delete (próprios)
comments:create, comments:read, comments:update (próprios), comments:delete (próprios)
profiles:read, profiles:update (próprios)
```

**Guest**:

```
posts:read, comments:read, profiles:read
```

## Métodos de Middleware

### RequirePermission

Verifica se o usuário tem **uma permissão específica**.

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

**Retorna:**

- `200 OK` - Usuário tem permissão, prossegue para o handler
- `401 Unauthorized` - Usuário não autenticado
- `403 Forbidden` - Usuário não tem permissão
- `500 Internal Server Error` - Erro de banco de dados ao verificar permissões

---

### RequireAnyPermission

Verifica se o usuário tem **pelo menos uma** das permissões especificadas (lógica OU).

```go
func (m *AuthorizationMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc
```

**Uso:**

```go
// Permitir se o usuário puder ler OU moderar comentários
router.GET("/comments/flagged",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission("comments:read", "comments:moderate"),
    handler.GetFlaggedComments,
)
```

**Casos de Uso:**

- Permissões alternativas (admin OU moderator)
- Acesso a recursos com múltiplos pontos de entrada
- Elevação gradual de permissões

---

### RequireAllPermissions

Verifica se o usuário tem **todas** as permissões especificadas (lógica E).

```go
func (m *AuthorizationMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc
```

**Uso:**

```go
// Requer permissões de publish E schedule
router.POST("/posts/schedule",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions("posts:create", "posts:schedule"),
    handler.SchedulePost,
)
```

**Casos de Uso:**

- Operações compostas que requerem múltiplas permissões
- Operações sensíveis que precisam de verificações em múltiplas camadas
- Combinações de recursos

---

### RequireRole

Verifica se o usuário tem **uma função específica** pelo nome.

```go
func (m *AuthorizationMiddleware) RequireRole(roleName string) gin.HandlerFunc
```

**Uso:**

```go
// Apenas administradores podem acessar
router.GET("/admin/dashboard",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.GetAdminDashboard,
)
```

**Nota:** Prefira `RequirePermission` em vez de `RequireRole` para melhor flexibilidade.

---

### RequireAnyRole

Verifica se o usuário tem **pelo menos uma** das funções especificadas (lógica OU).

```go
func (m *AuthorizationMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc
```

**Uso:**

```go
// Permitir admins OU moderators
router.GET("/moderation/queue",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyRole("admin", "moderator"),
    handler.GetModerationQueue,
)
```

**Casos de Uso:**

- Áreas administrativas com múltiplos níveis de funções
- Acesso a recursos para funções semelhantes
- Sistemas legados migrando de funções para permissões

## Exemplos de Uso

### Exemplo 1: Proteção CRUD Básica

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")

    // Acesso público de leitura (sem auth necessária)
    posts.GET("", handler.ListPosts)
    posts.GET("/:id", handler.GetPost)

    // Operações para autenticados
    posts.Use(r.authMiddleware.RequireAuth())
    {
        // Permissões específicas para cada operação
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

### Exemplo 2: Endpoints Somente para Admin

```go
func (r *UserRouter) Setup(api *gin.RouterGroup) {
    users := api.Group("/users")
    users.Use(r.authMiddleware.RequireAuth())

    // Operações de usuários normais
    users.GET("/me", handler.GetMe)
    users.PUT("/me", handler.UpdateProfile)

    // Operações somente para admin
    admin := users.Group("")
    admin.Use(r.authzMiddleware.RequirePermission("users:manage"))
    {
        admin.GET("", handler.ListAllUsers)
        admin.POST("/:id/ban", handler.BanUser)
        admin.POST("/:id/suspend", handler.SuspendUser)
    }
}
```

### Exemplo 3: Acesso Flexível com Múltiplas Permissões

```go
func (r *CommentRouter) Setup(api *gin.RouterGroup) {
    comments := api.Group("/comments")

    // Visualizar comentários - qualquer uma dessas permissões funciona
    comments.GET("/:id",
        r.authMiddleware.RequireAuth(),
        r.authzMiddleware.RequireAnyPermission(
            "comments:read",
            "comments:moderate",
            "*:read",
        ),
        handler.GetComment,
    )

    // Moderar comentários - requer read E moderate
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

### Exemplo 4: Acesso ao Dashboard Baseado em Função

```go
func (r *DashboardRouter) Setup(api *gin.RouterGroup) {
    dashboards := api.Group("/dashboard")
    dashboards.Use(r.authMiddleware.RequireAuth())

    // Dashboard do usuário - qualquer usuário autenticado
    dashboards.GET("/user", handler.GetUserDashboard)

    // Dashboard do moderador - moderators e admins
    dashboards.GET("/moderator",
        r.authzMiddleware.RequireAnyRole("moderator", "admin"),
        handler.GetModeratorDashboard,
    )

    // Dashboard do admin - apenas admins
    dashboards.GET("/admin",
        r.authzMiddleware.RequireRole("admin"),
        handler.GetAdminDashboard,
    )
}
```

### Exemplo 5: Lógica de Negócio Complexa

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")
    posts.Use(r.authMiddleware.RequireAuth())

    // Publicar requer permissões create e publish
    posts.POST("/:id/publish",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
        ),
        handler.PublishPost,
    )

    // Agendar requer create, publish E schedule
    posts.POST("/:id/schedule",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
            "posts:schedule",
        ),
        handler.SchedulePost,
    )

    // Destacar requer função moderator OU admin + permissão feature
    posts.POST("/:id/feature",
        r.authzMiddleware.RequireAnyRole("admin", "moderator"),
        r.authzMiddleware.RequirePermission("posts:feature"),
        handler.FeaturePost,
    )
}
```

### Exemplo 6: Configuração de Migração

Inicializar autorização na configuração do router:

```go
// cmd/api/main.go ou inicialização do router
func setupRouters(
    authMiddleware *middleware.AuthMiddleware,
    authzMiddleware *middleware.AuthorizationMiddleware,
) *gin.Engine {
    r := gin.New()

    // Rotas públicas
    api := r.Group("/api/v1")

    // Rotas de Auth (sem autorização necessária)
    authRouter := router.NewAuthRouter(authHandler, authMiddleware)
    authRouter.Setup(api)

    // Rotas protegidas com autorização
    postRouter := router.NewPostRouter(postHandler, authMiddleware, authzMiddleware)
    postRouter.Setup(api)

    userRouter := router.NewUserRouter(userHandler, authMiddleware, authzMiddleware)
    userRouter.Setup(api)

    return r
}
```

## Melhores Práticas

### 1. Sempre Use RequireAuth Primeiro

Middleware de autorização requer contexto de autenticação:

```go
// [+] CORRETO - Auth antes de autorização
router.POST("/posts",
    authMiddleware.RequireAuth(),           // Primeiro: autenticação
    authzMiddleware.RequirePermission(...), // Depois: autorização
    handler.CreatePost,
)

// [X] INCORRETO - Autorização sem autenticação
router.POST("/posts",
    authzMiddleware.RequirePermission(...), // Falha - sem user_id
    handler.CreatePost,
)
```

### 2. Prefira Permissões em vez de Funções

Permissões fornecem melhor flexibilidade e manutenibilidade:

```go
// [+] MELHOR - Baseado em permissões (flexível)
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:delete"),
    handler.DeletePost,
)

// [!] ACEITÁVEL mas menos flexível - Baseado em funções
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.DeletePost,
)
```

**Por quê?**

- Adicionar novas funções não requer alterações de código
- Permissões podem ser reatribuídas sem tocar no código
- Controle mais granular

### 3. Use Nomes Descritivos para Permissões

```go
// [+] BOM - Intenção clara
"posts:create"
"posts:publish"
"posts:feature"
"posts:schedule"

// [X] RUIM - Pouco claro
"posts:manage"  // O que significa "manage"?
"posts:admin"   // Muito genérico
```

### 4. Aproveite Wildcards para Funções de Admin

```go
// Em sua migração/seed
INSERT INTO permissions (resource, action) VALUES
    ('*', '*'),           -- Admin: tudo
    ('posts', '*'),       -- Admin de conteúdo: todas as operações de posts
    ('*', 'read');        -- Visualizador: ler tudo
```

### 5. Agrupe Permissões Relacionadas

```go
// Agrupar por área funcional
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Todas as operações de escrita de posts requerem posts:* ou posts:write
write := posts.Group("")
write.Use(authzMiddleware.RequirePermission("posts:write"))
{
    write.POST("", handler.CreatePost)
    write.PUT("/:id", handler.UpdatePost)
    write.DELETE("/:id", handler.DeletePost)
}

// Operações públicas de leitura
posts.GET("", handler.ListPosts)
posts.GET("/:id", handler.GetPost)
```

### 6. Trate Recursos de Proprietário no Handler

Não use middleware de autorização para verificações de proprietário:

```go
// [+] CORRETO - Verificar propriedade no handler
func (h *PostHandler) UpdatePost(c *gin.Context) {
    userID := middleware.GetUserIDOrPanic(c)
    postID := c.Param("id")

    post, err := h.postUC.GetByID(c.Request.Context(), postID)
    if err != nil {
        response.Error(c, http.StatusNotFound, "post not found", err)
        return
    }

    // Verificar propriedade OU permissão de admin
    if post.UserID != userID {
        hasAdmin, _ := h.roleUC.HasPermission(c.Request.Context(), userID, "posts:*")
        if !hasAdmin {
            response.Error(c, http.StatusForbidden, "can only update own posts", nil)
            return
        }
    }

    // Continuar atualização...
}

// [X] INCORRETO - Tentar verificar propriedade no middleware
// Middleware não tem acesso aos detalhes do recurso
```

### 7. Use RequireAny para Permissões de Fallback

```go
// Permitir operação se usuário tem permissão específica OU é admin
router.POST("/posts/:id/feature",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission(
        "posts:feature",  // Permissão específica
        "posts:*",        // Acesso total a posts
        "*:*",            // Admin
    ),
    handler.FeaturePost,
)
```

## Padrões Comuns

### Padrão 1: Sobrescrita de Admin

Permitir que admins contornem verificações de propriedade:

```go
// Qualquer permissão de admin sobrescreve propriedade
authzMiddleware.RequireAnyPermission(
    "posts:update",  // Permissão de usuário normal
    "posts:*",       // Admin de posts
    "*:*",           // Admin
)
```

### Padrão 2: Permissões Graduadas

Diferentes níveis de permissões para o mesmo recurso:

```go
// Nível 1: Leitura básica
authzMiddleware.RequirePermission("posts:read")

// Nível 2: Leitura + escrita
authzMiddleware.RequireAllPermissions("posts:read", "posts:write")

// Nível 3: Acesso total
authzMiddleware.RequirePermission("posts:*")
```

### Padrão 3: Permissões entre Recursos

Operações que afetam múltiplos recursos:

```go
// Publicar post pode requerer permissões de post E mídia
router.POST("/posts/:id/publish",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions(
        "posts:publish",
        "media:attach",  // Se o post contém imagens
    ),
    handler.PublishPost,
)
```

### Padrão 4: Autorização Condicional

Permissões diferentes para endpoints diferentes:

```go
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Posts de rascunho - apenas permissão create
posts.POST("/drafts",
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreateDraft,
)

// Posts publicados - permissões create + publish
posts.POST("/publish",
    authzMiddleware.RequireAllPermissions("posts:create", "posts:publish"),
    handler.CreateAndPublish,
)
```

## Tratamento de Erros

### Códigos de Status HTTP

| Status | Significado           | Motivo                                         |
| ------ | --------------------- | ---------------------------------------------- |
| 401    | Unauthorized          | Usuário não autenticado (sem JWT)              |
| 403    | Forbidden             | Usuário autenticado mas sem permissão          |
| 500    | Internal Server Error | Erro de banco de dados ao verificar permissões |

### Formato de Resposta de Erro

```json
{
  "error": "insufficient permissions",
  "message": "You don't have permission to perform this action"
}
```

### Tratamento no Lado do Cliente

```typescript
// Exemplo TypeScript/JavaScript
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
    // Redirecionar para login
    window.location.href = "/login";
  } else if (response.status === 403) {
    // Mostrar mensagem de "permissões insuficientes"
    showError("You do not have permission to create posts");
  } else if (response.ok) {
    // Sucesso
    const post = await response.json();
  }
} catch (error) {
  console.error("Request failed:", error);
}
```

## Testando Autorização

### Testes Unitários de Middleware

```go
func TestAuthorizationMiddleware_RequirePermission(t *testing.T) {
    // Configuração
    mockRoleUC := mocks.NewMockRoleUseCase(t)
    authzMiddleware := middleware.NewAuthorizationMiddleware(mockRoleUC)

    t.Run("allows user with permission", func(t *testing.T) {
        // Criar contexto de teste com user_id
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Set("user_id", testUserID)

        // Mock de verificação de permissão - retorna true
        mockRoleUC.EXPECT().
            HasPermission(mock.Anything, testUserID, "posts:create").
            Return(true, nil)

        // Criar cadeia de handler
        handler := authzMiddleware.RequirePermission("posts:create")(func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })

        // Executar
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

### Testes de Integração

```go
func TestPostEndpoints_Authorization(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    // Criar usuários de teste com diferentes funções
    adminUser := helpers.UserFixture(t, testDB.DB)
    regularUser := helpers.UserFixture(t, testDB.DB)

    // Atribuir funções
    assignRole(t, testDB, adminUser.ID, "admin")
    assignRole(t, testDB, regularUser.ID, "user")

    // Gerar tokens
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

### Testes Manuais com curl

```bash
# 1. Login e obter token
TOKEN=$(curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# 2. Testar endpoint protegido
curl -X POST http://localhost:8081/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Post","content":"Content"}'

# Respostas esperadas:
# 200 OK - Sucesso
# 401 Unauthorized - Token inválido/ausente
# 403 Forbidden - Permissões insuficientes
```

## Solução de Problemas

### Problema: 401 Unauthorized em Endpoint Protegido

**Sintomas:**

```json
{
  "error": "user not authenticated"
}
```

**Causas:**

1. Middleware `RequireAuth()` ausente antes de `RequirePermission()`
2. Token JWT inválido
3. Token expirado

**Solução:**

```go
// Garantir que middleware de auth é aplicado primeiro
router.POST("/posts",
    authMiddleware.RequireAuth(),  // ← Deve estar antes da autorização
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

### Problema: 403 Forbidden para Admin

**Sintomas:**
Usuário admin recebe 403 em endpoints que deveria ter acesso.

**Causas:**

1. Permissão wildcard (`*:*`) não está sendo verificada adequadamente
2. Função não atribuída ao usuário
3. Permissão não atribuída à função

**Solução:**

```sql
-- Verificar que admin tem permissão wildcard
SELECT r.name, p.resource, p.action
FROM roles r
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions p ON p.id = rp.permission_id
WHERE r.name = 'admin';

-- Deve retornar: name='admin', resource='*', action='*'

-- Verificar que usuário tem função admin
SELECT u.email, r.name
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
WHERE u.id = '<user_uuid>';
```

### Problema: Performance do Banco de Dados com Verificações de Permissões

**Sintomas:**
Tempos de resposta lentos em endpoints protegidos.

**Solução:**
Implementar cache no RoleUseCase:

```go
// Usar Redis/cache em memória para verificações de permissões
func (uc *RoleUseCase) HasPermission(ctx context.Context, userID uuidv7.UUID, permission string) (bool, error) {
    // Verificar cache primeiro
    cacheKey := fmt.Sprintf("user:%s:permission:%s", userID, permission)
    if cached, found := uc.cache.Get(cacheKey); found {
        return cached.(bool), nil
    }

    // Consulta ao banco de dados
    hasPermission, err := uc.repo.HasPermission(ctx, userID, permission)
    if err != nil {
        return false, err
    }

    // Cache por 5 minutos
    uc.cache.Set(cacheKey, hasPermission, 5*time.Minute)

    return hasPermission, nil
}
```

## Resumo

O Middleware de Autorização fornece controle de acesso poderoso e flexível para sua API:

[+] **Baseado em permissões** - Controle granular no formato `resource:action`  
[+] **Suporte a wildcards** - Herança poderosa com padrões `*`  
[+] **Verificações compostas** - Lógica E/OU para requisitos complexos  
[+] **Atalhos de funções** - Verificações rápidas baseadas em funções quando necessário  
[+] **Arquitetura limpa** - Separa autorização de autenticação  
[+] **Pronto para produção** - Tratamento de erros e performance comprovados

**Referência Rápida:**

```go
// Verificação de permissão única
RequirePermission("posts:create")

// Qualquer uma de múltiplas permissões (OU)
RequireAnyPermission("posts:read", "posts:*", "*:*")

// Todas as múltiplas permissões (E)
RequireAllPermissions("posts:create", "posts:publish")

// Verificação baseada em função
RequireRole("admin")

// Qualquer uma de múltiplas funções (OU)
RequireAnyRole("admin", "moderator")
```

Para mais informações, consulte:

- [RBAC Implementation](RBAC_IMPLEMENTATION.md)
- [Testing Guide](TESTING_GUIDE.pt.md)
- [API Documentation](../README.pt.md)
