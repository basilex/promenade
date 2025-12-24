# Arquitetura do Sistema de Autenticação

🇬🇧 [English](AUTH_SCHEMA.pt.md) | [🇺🇦 Українська](../uk/AUTH_SCHEMA.uk.md) | [🇩🇪 Deutsch](../de/AUTH_SCHEMA.de.md) | 🇵🇹 **Português** | [🇪🇸 Español](../es/AUTH_SCHEMA.es.md)

Referência completa para o sistema de autenticação no Promenade. Abrange registro de usuários, login, gerenciamento de sessões, manipulação de tokens e mecanismos de segurança.

## Índice

- [Visão Geral](#visão-geral)
- [Esquema do Banco de Dados](#esquema-do-banco-de-dados)
- [Estados e Ciclo de Vida do Usuário](#estados-e-ciclo-de-vida-do-usuário)
- [Fluxo de Autenticação](#fluxo-de-autenticação)
- [Gerenciamento de Sessões](#gerenciamento-de-sessões)
- [Sistema de Tokens](#sistema-de-tokens)
- [Mecanismos de Segurança](#mecanismos-de-segurança)
- [Endpoints da API](#endpoints-da-api)
- [Tratamento de Erros](#tratamento-de-erros)
- [Configuração](#configuração)

---

## Visão Geral

**Componentes do Sistema de Autenticação:**

| Componente      | Propósito                                 | Tecnologia          |
| --------------- | ----------------------------------------- | ------------------- |
| User Entity     | Conta principal com senha                 | PostgreSQL, bcrypt  |
| Session Entity  | Armazenamento de refresh tokens           | PostgreSQL, SHA-256 |
| JWT Manager     | Geração/validação de access tokens        | HMAC-SHA256         |
| Auth UseCase    | Lógica de negócio de todas operações auth | Go                  |
| Auth Middleware | Autenticação de requisições               | Gin middleware      |
| Event Bus       | Notificações assíncronas (e-mail, logs)   | Memory/Redis        |

**Recursos Principais:**

- ✅ Autenticação baseada em JWT (access + refresh tokens)
- ✅ Rotação de refresh tokens (melhor prática de segurança)
- ✅ Gerenciamento de sessões concorrentes (máx. 5 por usuário)
- ✅ Máquina de estados do usuário (unverified → active → suspended/banned)
- ✅ Hashing de senhas com bcrypt (custo 10)
- ✅ Hashing de refresh tokens com SHA-256
- ✅ Limpeza automática de sessões em mudanças de estado
- ✅ Notificações por e-mail assíncronas via event bus

---

## Esquema do Banco de Dados

### Tabela `core_users`

```sql
CREATE TABLE core_users (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    email              VARCHAR(255) NOT NULL UNIQUE,
    name               VARCHAR(255) NOT NULL,
    password           VARCHAR(255) NOT NULL,  -- hash bcrypt
    status             VARCHAR(20) NOT NULL DEFAULT 'unverified',
    email_verified_at  TIMESTAMPTZ,
    suspended_reason   TEXT,
    suspended_until    TIMESTAMPTZ,
    last_login_at      TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON core_users(email);
CREATE INDEX idx_users_status ON core_users(status);
```

### Tabela `core_user_sessions`

```sql
CREATE TABLE core_user_sessions (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id       UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,  -- hash SHA-256
    user_agent    TEXT,
    ip_address    VARCHAR(45),
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON core_user_sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON core_user_sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON core_user_sessions(expires_at);
```

**Tabelas Adicionais (ainda não implementadas):**

- `core_password_reset_tokens` - Processo de redefinição de senha
- `core_email_verification_tokens` - Processo de verificação de e-mail
- `core_login_attempts` - Proteção contra força bruta

---

## Estados e Ciclo de Vida do Usuário

### Enumeração de Status do Usuário

```go
type UserStatus string

const (
    UserStatusUnverified UserStatus = "unverified" // Registrado, mas e-mail não verificado
    UserStatusActive     UserStatus = "active"     // E-mail verificado e conta ativa
    UserStatusSuspended  UserStatus = "suspended"  // Suspenso temporariamente (pode ser reativado)
    UserStatusBanned     UserStatus = "banned"     // Banido permanentemente
    UserStatusInactive   UserStatus = "inactive"   // Desativado pelo usuário (pode ser reativado)
)
```

### Transições de Estado

```
                    Register()
                        │
                        ▼
                 ┌──────────────┐
                 │  unverified  │  ──────────────┐
                 └──────────────┘                │
                        │                        │
                 VerifyEmail()              Login() permitido
                        │                        │
                        ▼                        ▼
                 ┌──────────────┐         Usuário pode fazer login
                 │    active    │         (unverified ou active)
                 └──────────────┘
                    │   │   │
        ┌───────────┘   │   └───────────┐
   Suspend()      Ban()           Deactivate()
        │               │                │
        ▼               ▼                ▼
 ┌───────────┐   ┌──────────┐    ┌─────────────┐
 │ suspended │   │  banned  │    │  inactive   │
 └───────────┘   └──────────┘    └─────────────┘
        │                              │
   Reactivate()                   Reactivate()
        │                              │
        └────────────┬─────────────────┘
                     │
                     ▼
              ┌──────────────┐
              │    active    │
              └──────────────┘
```

### Lógica CanLogin()

```go
func (u *User) CanLogin() bool {
    return u.Status == UserStatusActive || u.Status == UserStatusUnverified
}
```

**Status Permitidos:**

- ✅ `active` - Acesso completo
- ✅ `unverified` - Pode fazer login, mas funcionalidades podem estar restritas

**Status Bloqueados:**

- ❌ `suspended` - Retorna `ErrUserSuspended`
- ❌ `banned` - Retorna `ErrUserBanned`
- ❌ `inactive` - Retorna `ErrUserNotActive`

---

## Fluxo de Autenticação

### 1. Fluxo de Registro

```
Cliente                 API                    UseCase                Banco de Dados    Event Bus
  │                      │                        │                        │                 │
  │  POST /auth/register │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │                      │  Register(email, name, pwd)                     │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  (não encontrado - OK) │                 │
  │                      │                        │                        │                 │
  │                      │                        │  bcrypt.Hash(pwd)      │                 │
  │                      │                        │  user.Status = "unverified"              │
  │                      │                        │                        │                 │
  │                      │                        │  Create(user)          │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │                        │                 │
  │                      │                        │  Publish(UserRegisteredEvent)            │
  │                      │                        │─────────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  201 Created         │                        │                        │  EmailWorker    │
  │<─────────────────────│                        │                        │  envia e-mail   │
  │  {id, email, name}   │                        │                        │  de boas-vindas │
```

**Pontos Importantes:**

- Senha é hasheada via `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)`
- Usuário criado com `status = "unverified"`
- UUID v7 gerado para ID do usuário (ordenado temporalmente)
- Evento publicado assincronamente - registro não espera e-mail
- Email worker processa evento `user.registered` em background

---

### 2. Fluxo de Login

```
Cliente                 API                    UseCase                Banco de Dados    Sessão
  │                      │                        │                        │                 │
  │  POST /auth/login    │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │  {email, password}   │  Login(email, pwd, ua, ip)                      │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  usuário encontrado    │                 │
  │                      │                        │                        │                 │
  │                      │                        │  bcrypt.Compare(pwd, hash)               │
  │                      │                        │  OK ✓                  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  SIM ✓                 │                 │
  │                      │                        │                        │                 │
  │                      │                        │  JWTManager.GenerateAccessToken()        │
  │                      │                        │  crypto/rand 32 bytes para refresh token │
  │                      │                        │  SHA-256(refresh_token)                  │
  │                      │                        │                        │                 │
  │                      │                        │  CountUserSessions(user_id)              │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  count = 5 (limite!)   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetOldestSession()    │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  Delete(oldest)        │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │                        │                 │
  │                      │                        │  Create(new session)   │                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │                        │  UpdateLastLogin()     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  200 OK              │  (access_token, refresh_token, user)            │                 │
  │<─────────────────────│                        │                        │                 │
  │  {access_token,      │                        │                        │                 │
  │   refresh_token,     │                        │                        │                 │
  │   user: {...}}       │                        │                        │                 │
```

**Pontos Importantes:**

- **Validação de senha:** `bcrypt.CompareHashAndPassword(hash, password)`
- **Verificação de estado:** Deve ser `active` ou `unverified`
- **Access token:** JWT com TTL de 15 minutos (padrão)
- **Refresh token:** 32 bytes aleatórios + base64, TTL de 7 dias (padrão)
- **Armazenamento de refresh token:** Hasheado com SHA-256 antes de salvar no BD
- **Limite de sessões:** Máx. 5 sessões concorrentes por usuário
- **Deletar mais antiga:** Se limite excedido, sessão mais antiga é deletada automaticamente
- **Último login:** Atualizado assincronamente (não bloqueia resposta)

---

### 3. Fluxo de Atualização de Token

```
Cliente                 API                    UseCase                Banco de Dados    Sessão
  │                      │                        │                        │                 │
  │  POST /auth/refresh  │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │  {refresh_token}     │  RefreshToken(token)   │                        │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  SHA-256(token)        │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetByRefreshToken(hash)                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │<───────────────────────────────────────│
  │                      │                        │  sessão encontrada     │                 │
  │                      │                        │                        │                 │
  │                      │                        │  session.IsExpired()?  │                 │
  │                      │                        │  NÃO ✓                 │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetByID(session.user_id)                │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  usuário encontrado    │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  SIM ✓                 │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GenerateAccessToken() │                 │
  │                      │                        │  crypto/rand novo refresh                │
  │                      │                        │  SHA-256(new_refresh)  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  Update(session)       │  ← ROTAÇÃO DE TOKEN
  │                      │                        │  - novo refresh hash   │                 │
  │                      │                        │  - nova expires_at     │                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  200 OK              │  (new_access, new_refresh)                      │                 │
  │<─────────────────────│                        │                        │                 │
  │  {access_token,      │                        │                        │                 │
  │   refresh_token}     │                        │                        │                 │
```

**Pontos Importantes:**

- **Rotação de Token:** Refresh token antigo invalidado, novo emitido
- **Segurança:** Tokens únicos previnem ataques de replay
- **Reutilização de sessão:** Atualiza sessão existente em vez de delete+create (performance)
- **Verificação de expiração:** Sessões expiradas automaticamente deletadas
- **Verificação de estado:** Usuário ainda deve poder fazer login

---

### 4. Fluxo de Logout

```
Cliente                 API                    UseCase                Sessão
  │                      │                        │                        │
  │  POST /auth/logout   │                        │                        │
  │─────────────────────>│                        │                        │
  │  {refresh_token}     │  Logout(token)         │                        │
  │                      │───────────────────────>│                        │
  │                      │                        │  SHA-256(token)        │
  │                      │                        │                        │
  │                      │                        │  GetByRefreshToken(hash)
  │                      │                        │───────────────────────>│
  │                      │                        │<───────────────────────│
  │                      │                        │  sessão encontrada     │
  │                      │                        │                        │
  │                      │                        │  Delete(session.id)    │
  │                      │                        │───────────────────────>│
  │                      │                        │                        │
  │                      │<───────────────────────│                        │
  │  200 OK              │                        │                        │
  │<─────────────────────│                        │                        │
  │  {message: "success"}│                        │                        │
```

**Pontos Importantes:**

- **Deleção de sessão:** Invalida refresh token imediatamente
- **Access token:** Permanece válido até expiração (JWT sem estado)
- **Responsabilidade do cliente:** Cliente deve descartar ambos os tokens

---

## Gerenciamento de Sessões

### Limite de Sessões Concorrentes

```go
const MaxConcurrentSessions = 5
```

**Comportamento:**

- Usuário pode ter máximo de 5 sessões ativas em diferentes dispositivos
- No 6º login: sessão mais antiga é automaticamente deletada
- Previne ataques de geração ilimitada de tokens

### Entidade Session

```go
type Session struct {
    ID           UUID      `db:"id"`
    UserID       UUID      `db:"user_id"`
    RefreshToken string    `db:"refresh_token"`  // hasheado SHA-256
    UserAgent    *string   `db:"user_agent"`     // Informação do navegador/dispositivo
    IPAddress    *string   `db:"ip_address"`     // IP do cliente
    ExpiresAt    time.Time `db:"expires_at"`     // Expiração absoluta
    CreatedAt    time.Time `db:"created_at"`     // Início da sessão
}
```

### Operações de Sessão

| Operação          | Propósito                          | Gatilho                      |
| ----------------- | ---------------------------------- | ---------------------------- |
| Create            | Nova sessão no login               | Login                        |
| Update            | Rotação de refresh token           | Atualização de token         |
| Delete            | Logout de sessão única             | Logout                       |
| DeleteByUserID    | Invalidar todas sessões do usuário | Suspend/Ban/Mudança de senha |
| GetOldestSession  | Encontrar mais antiga para deletar | Excedeu limite de sessões    |
| CountUserSessions | Verificar limite                   | Login                        |

---

## Sistema de Tokens

### JWT Access Token

**Propriedades:**

- **Algoritmo:** HS256 (HMAC-SHA256)
- **TTL:** 15 minutos (padrão, configurável)
- **Armazenamento:** Apenas no cliente (não no BD)
- **Validação:** Verificação de assinatura + expiração em cada requisição

**Estrutura de Claims:**

```go
type Claims struct {
    UserID uuidv7.UUID `json:"user_id"`
    Email  string      `json:"email"`
    jwt.RegisteredClaims
}

// RegisteredClaims inclui:
// - iat (issued at - emitido em)
// - exp (expires at - expira em)
// - nbf (not before - não antes)
```

**Exemplo JWT:**

```json
{
  "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
  "email": "user@example.com",
  "iat": 1703347200,
  "exp": 1703348100,
  "nbf": 1703347200
}
```

### Refresh Token

**Propriedades:**

- **Algoritmo:** crypto/rand (32 bytes) + base64
- **TTL:** 7 dias (padrão, configurável)
- **Armazenamento:** Banco de dados (hasheado SHA-256)
- **Validação:** Comparação de hash + expiração

**Geração:**

```go
func generateRefreshToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

**Hashing (SHA-256):**

```go
func hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return base64.URLEncoding.EncodeToString(hash[:])
}
```

### Padrão de Rotação de Tokens

**Vantagem de Segurança:** Previne ataques de replay com refresh tokens

1. Cliente envia refresh token
2. Servidor valida e emite novo par de tokens
3. **Refresh token antigo invalidado imediatamente**
4. Cliente deve usar novo refresh token para próxima atualização

**Prevenção de Cenário de Ataque:**

- ❌ Atacante rouba refresh token
- ❌ Atacante tenta usá-lo
- ✅ Token já foi rotacionado pelo usuário legítimo → **Ataque falha**

---

## Mecanismos de Segurança

### Segurança de Senha

| Mecanismo | Implementação                     | Propósito                     |
| --------- | --------------------------------- | ----------------------------- |
| Hashing   | `bcrypt` (custo 10)               | Criptografia unidirecional    |
| Salt      | Automático (interno bcrypt)       | Hash único para cada          |
| Validação | `bcrypt.CompareHashAndPassword()` | Comparação em tempo constante |

**Código:**

```go
// Hashing no registro
func (u *User) HashPassword(password string) error {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedBytes)
    return nil
}

// Verificação no login
func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
```

### Segurança de Refresh Token

| Mecanismo         | Implementação        | Propósito                      |
| ----------------- | -------------------- | ------------------------------ |
| Geração aleatória | `crypto/rand` (32 B) | Criptograficamente seguro      |
| Hashing           | SHA-256              | Nunca armazenar em texto plano |
| Rotação de token  | Tokens únicos        | Invalidação após uso           |
| Expiração         | TTL de 7 dias        | Janela de impacto limitada     |

### Segurança JWT

| Mecanismo     | Implementação        | Propósito                    |
| ------------- | -------------------- | ---------------------------- |
| Assinatura    | HMAC-SHA256          | Proteção contra falsificação |
| Chave secreta | Variável de ambiente | Verificação de assinatura    |
| TTL curto     | 15 minutos           | Janela de impacto minimizada |
| Sem estado    | Sem acesso ao BD     | Performance + escalabilidade |

### Segurança de Ações Administrativas

**Invalidação Automática de Sessões:**

Quando admin suspende/bane usuário:

1. Status do usuário mudado no BD
2. **Todas as sessões do usuário deletadas imediatamente**
3. Usuário não pode atualizar tokens
4. Access tokens existentes expiram naturalmente (máx. 15 min)

```go
func (uc *authUseCase) SuspendUser(ctx context.Context, userID UUID, reason string, until *time.Time) error {
    // Atualizar status do usuário
    if err := uc.userRepo.Suspend(ctx, userID, reason, until); err != nil {
        return err
    }

    // CRÍTICO: Invalidar todas as sessões
    if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
        return err
    }

    // Notificação assíncrona
    uc.eventBus.Publish(ctx, bus.TopicUserSuspended, event)
    return nil
}
```

---

## Endpoints da API

### Endpoints Públicos (Sem Autenticação)

#### POST /api/v1/auth/register

Registrar nova conta.

**Requisição:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Resposta (201 Created):**

```json
{
  "status": "success",
  "data": {
    "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "unverified",
    "email_verified_at": null,
    "last_login_at": null,
    "created_at": "2024-12-23T10:00:00Z",
    "updated_at": "2024-12-23T10:00:00Z"
  }
}
```

---

#### POST /api/v1/auth/login

Autenticar usuário e obter tokens.

**Requisição:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Resposta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "email": "user@example.com",
      "name": "John Doe",
      "status": "active"
    }
  }
}
```

---

#### POST /api/v1/auth/refresh

Atualizar access token via refresh token.

**Requisição:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Resposta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "bmV3cmVmcmVzaHRva2VuaGVyZQ...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

**Nota:** Refresh token antigo é invalidado (rotação de tokens).

---

### Endpoints Protegidos (Autenticação Necessária)

#### GET /api/v1/auth/me

Obter perfil do usuário atual.

**Requisição:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Resposta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "active",
    "email_verified_at": "2024-12-20T14:30:00Z",
    "last_login_at": "2024-12-23T10:00:00Z",
    "created_at": "2024-12-15T09:00:00Z",
    "updated_at": "2024-12-23T10:00:00Z"
  }
}
```

---

#### GET /api/v1/auth/sessions

Obter todas as sessões ativas do usuário atual.

**Requisição:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/sessions \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Resposta (200 OK):**

```json
{
  "status": "success",
  "data": [
    {
      "id": "01936d6a-9999-7890-a1b2-c3d4e5f67890",
      "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...",
      "ip_address": "192.168.1.100",
      "expires_at": "2024-12-30T10:00:00Z",
      "created_at": "2024-12-23T10:00:00Z"
    },
    {
      "id": "01936d6a-8888-7890-a1b2-c3d4e5f67890",
      "user_id": "01936d6a-8f7c-7890-a1b2-c3d4e5f67890",
      "user_agent": "Mobile Safari/537.36",
      "ip_address": "192.168.1.101",
      "expires_at": "2024-12-29T15:30:00Z",
      "created_at": "2024-12-22T15:30:00Z"
    }
  ]
}
```

---

#### POST /api/v1/auth/logout

Fazer logout e invalidar refresh token.

**Requisição:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Resposta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "Successfully logged out"
  }
}
```

---

### Endpoints Administrativos (Permissões `users:suspend` / `users:ban` necessárias)

#### POST /api/v1/auth/users/:id/suspend

Suspender temporariamente conta do usuário.

**Requisição:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/suspend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Violação das regras da comunidade",
    "suspended_until": "2025-01-23T00:00:00Z"
  }'
```

**Resposta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User suspended successfully"
  }
}
```

**Consequências:**

- Status do usuário mudado para `suspended`
- **Todas as sessões ativas deletadas imediatamente**
- Usuário não pode fazer login até `suspended_until` ou reativação manual

---

#### POST /api/v1/auth/users/:id/ban

Banir permanentemente conta do usuário.

**Requisição:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/ban \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Spam e atividades fraudulentas"
  }'
```

**Resposta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User banned successfully"
  }
}
```

**Consequências:**

- Status do usuário mudado para `banned`
- **Todas as sessões ativas deletadas imediatamente**
- Usuário não pode fazer login (reativação manual por admin necessária)

---

#### POST /api/v1/auth/users/:id/reactivate

Reativar usuário suspenso/banido/inativo.

**Requisição:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/reactivate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Resposta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User reactivated successfully"
  }
}
```

---

## Tratamento de Erros

### Erros de Autenticação

| Código de Erro        | Status HTTP | Mensagem                   | Motivo                            |
| --------------------- | ----------- | -------------------------- | --------------------------------- |
| `INVALID_CREDENTIALS` | 401         | Invalid email or password  | Combinação e-mail/senha incorreta |
| `EMAIL_EXISTS`        | 409         | Email already exists       | Registro com e-mail existente     |
| `USER_NOT_ACTIVE`     | 403         | User account is not active | Status `inactive`                 |
| `USER_SUSPENDED`      | 403         | User account is suspended  | Status `suspended`                |
| `USER_BANNED`         | 403         | User account is banned     | Status `banned`                   |
| `INVALID_TOKEN`       | 401         | Invalid or expired token   | Validação de token falhou         |
| `TOKEN_EXPIRED`       | 401         | Token has expired          | JWT ou refresh token expirado     |
| `UNAUTHORIZED`        | 401         | Authentication required    | Authorization ausente ou inválido |

### Formato de Resposta de Erro

```json
{
  "status": "error",
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email or password",
    "details": null
  }
}
```

### Erros de Validação

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "must be a valid email address"
      },
      {
        "field": "password",
        "message": "must be at least 8 characters"
      }
    ]
  }
}
```

---

## Configuração

### Variáveis de Ambiente

```bash
# Configuração JWT
JWT_SECRET=sua-chave-secreta-256-bit-mudar-em-producao
JWT_ACCESS_TOKEN_TTL=15m    # Tempo de vida do access token (ex: 15m, 1h)
JWT_REFRESH_TOKEN_TTL=168h  # Tempo de vida do refresh token (ex: 168h = 7 dias)

# Banco de dados
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=promenade
DB_SSLMODE=disable

# Servidor
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

### Arquivo de Configuração (config/app.dev.yaml)

```yaml
jwt:
  secret: ${JWT_SECRET}
  access_token_ttl: 15m
  refresh_token_ttl: 168h

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  name: ${DB_NAME}
  sslmode: ${DB_SSLMODE}
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m

server:
  port: ${SERVER_PORT}
  host: ${SERVER_HOST}
  read_timeout: 10s
  write_timeout: 10s
```

### Melhores Práticas de Segurança

| Configuração            | Desenvolvimento | Produção                     |
| ----------------------- | --------------- | ---------------------------- |
| `JWT_SECRET`            | Qualquer        | **String aleatória 256-bit** |
| `JWT_ACCESS_TOKEN_TTL`  | 15m             | 5m - 15m                     |
| `JWT_REFRESH_TOKEN_TTL` | 168h (7 dias)   | 7-30 dias                    |
| `DB_SSLMODE`            | disable         | **require ou verify-full**   |
| `SERVER_HOST`           | 0.0.0.0         | 0.0.0.0 ou IP específico     |

**Configurações Críticas de Produção:**

1. Gerar JWT secret seguro: `openssl rand -base64 32`
2. Usar apenas HTTPS (certificados TLS/SSL)
3. Habilitar SSL do banco de dados (`DB_SSLMODE=require`)
4. Definir TTL curto para access token (5-15 minutos)
5. Monitorar tentativas de login falhadas (rate limiting)

---

## Documentação Relacionada

- [AUTHORIZATION.md](AUTHORIZATION.pt.md) - Sistema de permissões RBAC (o que acontece APÓS autenticação)
- [CREDENTIALS.md](CREDENTIALS.pt.md) - Usuários e roles padrão para desenvolvimento/testes
- [LOGGING.md](LOGGING.pt.md) - Logging estruturado com contexto de autenticação
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.pt.md) - Arquitetura do sistema e módulos principais
