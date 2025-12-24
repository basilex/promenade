# Guia de Testes Promenade

[🇬🇧 English](../TESTING_GUIDE.md) | [🇺🇦 Українська](../uk/TESTING_GUIDE.uk.md) | [🇩🇪 Deutsch](../de/TESTING_GUIDE.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/TESTING_GUIDE.es.md)

## Visão Geral

Um sistema de testes abrangente cobrindo todas as camadas da aplicação com **388 testes (100% passando)**:

- **Testes Unitários** (183) - lógica de negócios isolada e validação de entidades
- **Testes de Integração** (91) - operações de repositório com PostgreSQL real
- **Testes Smoke** (114) - fluxos críticos end-to-end com banco de dados real
- **Testes E2E** - testes de API HTTP (TODO)

## Início Rápido

```bash
# Executar todos os testes (unitários + integração)
make test                  # 274 testes em ~41s

# Suítes de testes individuais
make test-unit            # 183 testes unitários (~5s)
make test-integration     # 91 testes de integração (~36s)
make test-smoke           # 114 testes smoke (~4s)

# Cobertura e monitoramento
make test-coverage        # Relatório de cobertura HTML
make test-watch           # Modo watch (gotestsum)
```

## Banco de Dados de Teste

Testes de integração usam um banco de dados de teste separado na porta **5433**:

```bash
# Iniciar BD de teste
make test-db-start

# Parar e limpar
make test-db-stop

# Ver logs
make test-db-logs
```

**Importante:** BD de teste é completamente isolado dos bancos dev/prod.

## Estrutura de Testes

### 1. Testes de Integração (Repositórios)

Localizados ao lado do código: `internal/adapter/repository/postgres/*_test.go`

Exemplo:

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

### 2. Testes Unitários (Casos de Uso)

_TODO: Próximo passo_

Irão testar lógica de negócios com repositórios mockados:

- Register
- Login
- RefreshToken
- Logout
- SuspendUser
- BanUser
- ReactivateUser
- ChangePassword

### 3. Testes Smoke (Fluxos Críticos End-to-End)

Localizados em: `test/smoke/*_smoke_test.go`

**114 testes smoke** verificam fluxos críticos de usuário com operações reais de banco de dados.

Exemplo:

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

        // Continuar testando fluxo completo...
    })

    t.Logf("[SUCCESS] All auth smoke tests passed!")
}
```

**Executando Testes Smoke:**

```bash
# Todos os testes smoke
make test-smoke

# Teste smoke específico
go test -v ./test/smoke -run TestRBAC_SmokeTest
go test -v ./test/smoke -run TestUserPost_SmokeTest

# Pular no modo short
go test -short ./test/smoke  # Testes smoke são pulados
```

**Cobertura por Módulo:**

| Módulo           | Cenários | Cobertura                                                 |
| ---------------- | -------- | --------------------------------------------------------- |
| Auth             | 8        | Registro, login, sessões, atualização, logout             |
| Country/Currency | 12       | Operações CRUD completas                                  |
| UserContact      | 11       | Email, telefone, telegram, principal, verificação         |
| UserPost         | 12       | Rascunho, publicação, destaque, agendamento, views, busca |
| UserProfile      | 12       | Privacidade, verificação, ban/unban, views, busca         |
| PostComment      | 13       | Threading, respostas, respostas aninhadas, soft delete    |
| CommentLikes     | 5        | Like/unlike, paginação, performance (100 verificações)    |
| RBAC             | 28       | Permissões, papéis, wildcards, expiração                  |
| RBAC Integration | 13       | Cenários reais de permissões (moderador, admin, etc.)     |
| **Total**        | **114**  | **Todos os testes passando [+]**                          |

**Recursos Principais:**

- [+] Integração real com PostgreSQL (porta 5433)
- [+] Verificação de caminho crítico (fluxos CRUD)
- [+] Benchmarks de performance incluídos
- [+] Execução rápida (~4 segundos para 114 testes)
- [+] Taxa de sucesso de 100%

### 4. Testes de Integração HTTP (Handlers)

_TODO: Após expansão dos testes smoke_

Testar todos os endpoints HTTP via roteador Gin real:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `GET /api/auth/me`
- etc.

## Helpers de Teste

### `test/helpers/database.go`

```go
// Conectar ao BD de teste
testDB := helpers.SetupTestDB(t)
defer testDB.Close()

// Limpar todas as tabelas
testDB.CleanupTables(t)

// Teste transacional (auto rollback)
testDB.RunInTransaction(t, func(tx *sqlx.Tx) {
    // Seu código com tx
})
```

### `test/helpers/fixtures.go`

```go
// Usuário ativo padrão
user := helpers.UserFixture()

// Usuário não verificado
user := helpers.UnverifiedUserFixture()

// Usuário suspenso
user := helpers.SuspendedUserFixture()

// Usuário banido
user := helpers.BannedUserFixture()

// Usuário personalizado
user := helpers.UserFixture(func(u *entity.User) {
    u.Email = "custom@test.com"
    u.Status = entity.UserStatusInactive
})

// Sessão
session := helpers.SessionFixture(userID)

// Sessão expirada
session := helpers.ExpiredSessionFixture(userID)
```

## Melhores Práticas

### [+] Fazer

- Use `testify/require` para verificações críticas (para o teste)
- Use `testify/assert` para verificações não-críticas (continua o teste)
- Sempre fazer limpeza: `defer testDB.CleanupTables(t)`
- Testar casos extremos: sessões expiradas, usuários banidos, etc.
- Usar fixtures para dados de teste consistentes

### [X] Não Fazer

- Não use BD de produção para testes
- Não crie dependências entre testes
- Não esqueça `defer testDB.Close()`
- Não codifique dados de teste diretamente - use fixtures

## Integração CI/CD

Testes estão prontos para CI:

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

## O Que Vem a Seguir

1. [+] Testes de integração de repositórios - **CONCLUÍDO**
2. ⏳ Testes unitários de casos de uso com mocks
3. ⏳ Testes de integração de handlers HTTP
4. ⏳ Testes E2E para fluxos completos
5. ⏳ Testes de performance/benchmark

## Exemplos

### Executando Testes Específicos

```bash
# Arquivo de teste único
go test -v ./internal/adapter/repository/postgres/user_repository_test.go

# Teste único
go test -v ./internal/adapter/repository/postgres -run TestUserRepository_Create

# Com detector de race
go test -race ./...

# Com cobertura
go test -cover ./internal/adapter/repository/postgres
```

### Depurando Testes

```bash
# Saída verbosa
go test -v ./...

# Com logs do BD
make test-db-logs

# Verificar estado do BD durante teste
docker exec -it promenade_test_db psql -U system -d promenade_test
```

## Solução de Problemas

### "connection refused"

```bash
make test-db-start
# Aguarde 3-5 segundos para o BD ficar pronto
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

**Dúvidas?** Verifique `Makefile.test.mk` para todos os comandos disponíveis.
