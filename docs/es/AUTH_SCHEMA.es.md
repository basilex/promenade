# Arquitectura del Sistema de Autenticación

🇬🇧 [English](../AUTH_SCHEMA.md) | [🇺🇦 Українська](../uk/AUTH_SCHEMA.uk.md) | [🇩🇪 Deutsch](../de/AUTH_SCHEMA.de.md) | [🇵🇹 Português](../pt/AUTH_SCHEMA.pt.md) | 🇪🇸 **Español**

Referencia completa para el sistema de autenticación en Promenade. Cubre registro de usuarios, inicio de sesión, gestión de sesiones, manejo de tokens y mecanismos de seguridad.

## Índice

- [Descripción General](#descripción-general)
- [Esquema de Base de Datos](#esquema-de-base-de-datos)
- [Estados y Ciclo de Vida del Usuario](#estados-y-ciclo-de-vida-del-usuario)
- [Flujo de Autenticación](#flujo-de-autenticación)
- [Gestión de Sesiones](#gestión-de-sesiones)
- [Sistema de Tokens](#sistema-de-tokens)
- [Mecanismos de Seguridad](#mecanismos-de-seguridad)
- [Endpoints de la API](#endpoints-de-la-api)
- [Manejo de Errores](#manejo-de-errores)
- [Configuración](#configuración)

---

## Descripción General

**Componentes del Sistema de Autenticación:**

| Componente      | Propósito                                                   | Tecnología          |
| --------------- | ----------------------------------------------------------- | ------------------- |
| User Entity     | Cuenta principal con contraseña                             | PostgreSQL, bcrypt  |
| Session Entity  | Almacenamiento de refresh tokens                            | PostgreSQL, SHA-256 |
| JWT Manager     | Generación/validación de access tokens                      | HMAC-SHA256         |
| Auth UseCase    | Lógica de negocio de todas las operaciones de autenticación | Go                  |
| Auth Middleware | Autenticación de solicitudes                                | Gin middleware      |
| Event Bus       | Notificaciones asíncronas (email, logs)                     | Memory/Redis        |

**Características Principales:**

-  Autenticación basada en JWT (access + refresh tokens)
-  Rotación de refresh tokens (mejor práctica de seguridad)
-  Gestión de sesiones concurrentes (máx. 5 por usuario)
-  Máquina de estados del usuario (unverified → active → suspended/banned)
-  Hashing de contraseñas con bcrypt (costo 10)
-  Hashing de refresh tokens con SHA-256
-  Limpieza automática de sesiones en cambios de estado
-  Notificaciones por email asíncronas vía event bus

---

## Esquema de Base de Datos

### Tabla `core_users`

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

### Tabla `core_user_sessions`

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

**Tablas Adicionales (aún no implementadas):**

- `core_password_reset_tokens` - Proceso de restablecimiento de contraseña
- `core_email_verification_tokens` - Proceso de verificación de email
- `core_login_attempts` - Protección contra fuerza bruta

---

## Estados y Ciclo de Vida del Usuario

### Enumeración de Estado del Usuario

```go
type UserStatus string

const (
    UserStatusUnverified UserStatus = "unverified" // Registrado, pero email no verificado
    UserStatusActive     UserStatus = "active"     // Email verificado y cuenta activa
    UserStatusSuspended  UserStatus = "suspended"  // Suspendido temporalmente (puede reactivarse)
    UserStatusBanned     UserStatus = "banned"     // Baneado permanentemente
    UserStatusInactive   UserStatus = "inactive"   // Desactivado por el usuario (puede reactivarse)
)
```

### Transiciones de Estado

```
                    Register()
                        │
                        ▼
                 ──────────────
                 │  unverified  │  ──────────────
                 └──────────────                │
                        │                        │
                 VerifyEmail()              Login() permitido
                        │                        │
                        ▼                        ▼
                 ──────────────         Usuario puede iniciar sesión
                 │    active    │         (unverified o active)
                 └──────────────
                    │   │   │
        ───────────   │   └───────────
   Suspend()      Ban()           Deactivate()
        │               │                │
        ▼               ▼                ▼
 ───────────   ──────────    ─────────────
 │ suspended │   │  banned  │    │  inactive   │
 └───────────   └──────────    └─────────────
        │                              │
   Reactivate()                   Reactivate()
        │                              │
        └─────────────────────────────
                     │
                     ▼
              ──────────────
              │    active    │
              └──────────────
```

### Lógica CanLogin()

```go
func (u *User) CanLogin() bool {
    return u.Status == UserStatusActive || u.Status == UserStatusUnverified
}
```

**Estados Permitidos:**

-  `active` - Acceso completo
-  `unverified` - Puede iniciar sesión, pero las funcionalidades pueden estar restringidas

**Estados Bloqueados:**

-  `suspended` - Devuelve `ErrUserSuspended`
-  `banned` - Devuelve `ErrUserBanned`
-  `inactive` - Devuelve `ErrUserNotActive`

---

## Flujo de Autenticación

### 1. Flujo de Registro

```
Cliente                 API                    UseCase                Base de Datos    Event Bus
  │                      │                        │                        │                 │
  │  POST /auth/register │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │                      │  Register(email, name, pwd)                     │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  (no encontrado - OK)  │                 │
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
  │<─────────────────────│                        │                        │  envía email    │
  │  {id, email, name}   │                        │                        │  de bienvenida  │
```

**Puntos Importantes:**

- Contraseña hasheada vía `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)`
- Usuario creado con `status = "unverified"`
- UUID v7 generado para ID de usuario (ordenado temporalmente)
- Evento publicado asíncronamente - registro no espera email
- Email worker procesa evento `user.registered` en segundo plano

---

### 2. Flujo de Inicio de Sesión

```
Cliente                 API                    UseCase                Base de Datos    Sesión
  │                      │                        │                        │                 │
  │  POST /auth/login    │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │  {email, password}   │  Login(email, pwd, ua, ip)                      │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  usuario encontrado    │                 │
  │                      │                        │                        │                 │
  │                      │                        │  bcrypt.Compare(pwd, hash)               │
  │                      │                        │  OK                   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  SÍ                   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  JWTManager.GenerateAccessToken()        │
  │                      │                        │  crypto/rand 32 bytes para refresh token │
  │                      │                        │  SHA-256(refresh_token)                  │
  │                      │                        │                        │                 │
  │                      │                        │  CountUserSessions(user_id)              │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  count = 5 (¡límite!)  │                 │
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

**Puntos Importantes:**

- **Validación de contraseña:** `bcrypt.CompareHashAndPassword(hash, password)`
- **Verificación de estado:** Debe ser `active` o `unverified`
- **Access token:** JWT con TTL de 15 minutos (predeterminado)
- **Refresh token:** 32 bytes aleatorios + base64, TTL de 7 días (predeterminado)
- **Almacenamiento de refresh token:** Hasheado SHA-256 antes de guardar en BD
- **Límite de sesiones:** Máx. 5 sesiones concurrentes por usuario
- **Eliminar más antigua:** Si se excede límite, sesión más antigua se elimina automáticamente
- **Último inicio de sesión:** Se actualiza asíncronamente (no bloquea respuesta)

---

### 3. Flujo de Actualización de Token

```
Cliente                 API                    UseCase                Base de Datos    Sesión
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
  │                      │                        │  sesión encontrada     │                 │
  │                      │                        │                        │                 │
  │                      │                        │  session.IsExpired()?  │                 │
  │                      │                        │  NO                   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetByID(session.user_id)                │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  usuario encontrado    │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  SÍ                   │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GenerateAccessToken() │                 │
  │                      │                        │  crypto/rand nuevo refresh               │
  │                      │                        │  SHA-256(new_refresh)  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  Update(session)       │  ← ROTACIÓN TOKEN
  │                      │                        │  - nuevo refresh hash  │                 │
  │                      │                        │  - nueva expires_at    │                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  200 OK              │  (new_access, new_refresh)                      │                 │
  │<─────────────────────│                        │                        │                 │
  │  {access_token,      │                        │                        │                 │
  │   refresh_token}     │                        │                        │                 │
```

**Puntos Importantes:**

- **Rotación de Token:** Refresh token antiguo invalidado, nuevo emitido
- **Seguridad:** Tokens únicos previenen ataques de replay
- **Reutilización de sesión:** Actualiza sesión existente en lugar de delete+create (rendimiento)
- **Verificación de expiración:** Sesiones expiradas eliminadas automáticamente
- **Verificación de estado:** Usuario aún debe poder iniciar sesión

---

### 4. Flujo de Cierre de Sesión

```
Cliente                 API                    UseCase                Sesión
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
  │                      │                        │  sesión encontrada     │
  │                      │                        │                        │
  │                      │                        │  Delete(session.id)    │
  │                      │                        │───────────────────────>│
  │                      │                        │                        │
  │                      │<───────────────────────│                        │
  │  200 OK              │                        │                        │
  │<─────────────────────│                        │                        │
  │  {message: "success"}│                        │                        │
```

**Puntos Importantes:**

- **Eliminación de sesión:** Invalida refresh token inmediatamente
- **Access token:** Permanece válido hasta expiración (JWT sin estado)
- **Responsabilidad del cliente:** Cliente debe descartar ambos tokens

---

## Gestión de Sesiones

### Límite de Sesiones Concurrentes

```go
const MaxConcurrentSessions = 5
```

**Comportamiento:**

- Usuario puede tener máximo 5 sesiones activas en diferentes dispositivos
- En el 6º inicio de sesión: sesión más antigua se elimina automáticamente
- Previene ataques de generación ilimitada de tokens

### Entidad Session

```go
type Session struct {
    ID           UUID      `db:"id"`
    UserID       UUID      `db:"user_id"`
    RefreshToken string    `db:"refresh_token"`  // hasheado SHA-256
    UserAgent    *string   `db:"user_agent"`     // Información del navegador/dispositivo
    IPAddress    *string   `db:"ip_address"`     // IP del cliente
    ExpiresAt    time.Time `db:"expires_at"`     // Expiración absoluta
    CreatedAt    time.Time `db:"created_at"`     // Inicio de sesión
}
```

### Operaciones de Sesión

| Operación         | Propósito                               | Disparador                       |
| ----------------- | --------------------------------------- | -------------------------------- |
| Create            | Nueva sesión en inicio de sesión        | Login                            |
| Update            | Rotación de refresh token               | Actualización de token           |
| Delete            | Cerrar sesión única                     | Logout                           |
| DeleteByUserID    | Invalidar todas las sesiones de usuario | Suspend/Ban/Cambio de contraseña |
| GetOldestSession  | Encontrar más antigua para eliminar     | Excedió límite de sesiones       |
| CountUserSessions | Verificar límite                        | Login                            |

---

## Sistema de Tokens

### JWT Access Token

**Propiedades:**

- **Algoritmo:** HS256 (HMAC-SHA256)
- **TTL:** 15 minutos (predeterminado, configurable)
- **Almacenamiento:** Solo en el cliente (no en BD)
- **Validación:** Verificación de firma + expiración en cada solicitud

**Estructura de Claims:**

```go
type Claims struct {
    UserID uuidv7.UUID `json:"user_id"`
    Email  string      `json:"email"`
    jwt.RegisteredClaims
}

// RegisteredClaims incluye:
// - iat (issued at - emitido en)
// - exp (expires at - expira en)
// - nbf (not before - no antes de)
```

**Ejemplo JWT:**

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

**Propiedades:**

- **Algoritmo:** crypto/rand (32 bytes) + base64
- **TTL:** 7 días (predeterminado, configurable)
- **Almacenamiento:** Base de datos (hasheado SHA-256)
- **Validación:** Comparación de hash + expiración

**Generación:**

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

### Patrón de Rotación de Tokens

**Ventaja de Seguridad:** Previene ataques de replay con refresh tokens

1. Cliente envía refresh token
2. Servidor valida y emite nuevo par de tokens
3. **Refresh token antiguo invalidado inmediatamente**
4. Cliente debe usar nuevo refresh token para próxima actualización

**Prevención de Escenario de Ataque:**

-  Atacante roba refresh token
-  Atacante intenta usarlo
-  Token ya fue rotado por usuario legítimo → **Ataque falla**

---

## Mecanismos de Seguridad

### Seguridad de Contraseña

| Mecanismo  | Implementación                    | Propósito                       |
| ---------- | --------------------------------- | ------------------------------- |
| Hashing    | `bcrypt` (costo 10)               | Encriptación unidireccional     |
| Salt       | Automático (interno bcrypt)       | Hash único para cada            |
| Validación | `bcrypt.CompareHashAndPassword()` | Comparación en tiempo constante |

**Código:**

```go
// Hashing en registro
func (u *User) HashPassword(password string) error {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedBytes)
    return nil
}

// Verificación en inicio de sesión
func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
```

### Seguridad de Refresh Token

| Mecanismo            | Implementación       | Propósito                      |
| -------------------- | -------------------- | ------------------------------ |
| Generación aleatoria | `crypto/rand` (32 B) | Criptográficamente seguro      |
| Hashing              | SHA-256              | Nunca almacenar en texto plano |
| Rotación de token    | Tokens únicos        | Invalidación después de uso    |
| Expiración           | TTL de 7 días        | Ventana de impacto limitada    |

### Seguridad JWT

| Mecanismo     | Implementación      | Propósito                       |
| ------------- | ------------------- | ------------------------------- |
| Firma         | HMAC-SHA256         | Protección contra falsificación |
| Clave secreta | Variable de entorno | Verificación de firma           |
| TTL corto     | 15 minutos          | Ventana de impacto minimizada   |
| Sin estado    | Sin acceso a BD     | Rendimiento + escalabilidad     |

### Seguridad de Acciones Administrativas

**Invalidación Automática de Sesiones:**

Cuando admin suspende/banea usuario:

1. Estado del usuario cambiado en BD
2. **Todas las sesiones del usuario eliminadas inmediatamente**
3. Usuario no puede actualizar tokens
4. Access tokens existentes expiran naturalmente (máx. 15 min)

```go
func (uc *authUseCase) SuspendUser(ctx context.Context, userID UUID, reason string, until *time.Time) error {
    // Actualizar estado del usuario
    if err := uc.userRepo.Suspend(ctx, userID, reason, until); err != nil {
        return err
    }

    // CRÍTICO: Invalidar todas las sesiones
    if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
        return err
    }

    // Notificación asíncrona
    uc.eventBus.Publish(ctx, bus.TopicUserSuspended, event)
    return nil
}
```

---

## Endpoints de la API

### Endpoints Públicos (Sin Autenticación)

#### POST /api/v1/auth/register

Registrar nueva cuenta.

**Solicitud:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Respuesta (201 Created):**

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

Autenticar usuario y obtener tokens.

**Solicitud:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Respuesta (200 OK):**

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

Actualizar access token vía refresh token.

**Solicitud:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Respuesta (200 OK):**

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

**Nota:** Refresh token antiguo es invalidado (rotación de tokens).

---

### Endpoints Protegidos (Autenticación Requerida)

#### GET /api/v1/auth/me

Obtener perfil del usuario actual.

**Solicitud:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Respuesta (200 OK):**

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

Obtener todas las sesiones activas del usuario actual.

**Solicitud:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/sessions \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Respuesta (200 OK):**

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

Cerrar sesión e invalidar refresh token.

**Solicitud:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Respuesta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "Successfully logged out"
  }
}
```

---

### Endpoints Administrativos (Permisos `users:suspend` / `users:ban` requeridos)

#### POST /api/v1/auth/users/:id/suspend

Suspender temporalmente cuenta de usuario.

**Solicitud:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/suspend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Violación de reglas de la comunidad",
    "suspended_until": "2025-01-23T00:00:00Z"
  }'
```

**Respuesta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User suspended successfully"
  }
}
```

**Consecuencias:**

- Estado del usuario cambiado a `suspended`
- **Todas las sesiones activas eliminadas inmediatamente**
- Usuario no puede iniciar sesión hasta `suspended_until` o reactivación manual

---

#### POST /api/v1/auth/users/:id/ban

Banear permanentemente cuenta de usuario.

**Solicitud:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/ban \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Spam y actividades fraudulentas"
  }'
```

**Respuesta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User banned successfully"
  }
}
```

**Consecuencias:**

- Estado del usuario cambiado a `banned`
- **Todas las sesiones activas eliminadas inmediatamente**
- Usuario no puede iniciar sesión (reactivación manual por admin necesaria)

---

#### POST /api/v1/auth/users/:id/reactivate

Reactivar usuario suspendido/baneado/inactivo.

**Solicitud:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/reactivate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Respuesta (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User reactivated successfully"
  }
}
```

---

## Manejo de Errores

### Errores de Autenticación

| Código de Error       | Estado HTTP | Mensaje                    | Motivo                                  |
| --------------------- | ----------- | -------------------------- | --------------------------------------- |
| `INVALID_CREDENTIALS` | 401         | Invalid email or password  | Combinación email/contraseña incorrecta |
| `EMAIL_EXISTS`        | 409         | Email already exists       | Registro con email existente            |
| `USER_NOT_ACTIVE`     | 403         | User account is not active | Estado `inactive`                       |
| `USER_SUSPENDED`      | 403         | User account is suspended  | Estado `suspended`                      |
| `USER_BANNED`         | 403         | User account is banned     | Estado `banned`                         |
| `INVALID_TOKEN`       | 401         | Invalid or expired token   | Validación de token falló               |
| `TOKEN_EXPIRED`       | 401         | Token has expired          | JWT o refresh token expirado            |
| `UNAUTHORIZED`        | 401         | Authentication required    | Authorization ausente o inválido        |

### Formato de Respuesta de Error

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

### Errores de Validación

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

## Configuración

### Variables de Entorno

```bash
# Configuración JWT
JWT_SECRET=tu-clave-secreta-256-bit-cambiar-en-produccion
JWT_ACCESS_TOKEN_TTL=15m    # Tiempo de vida del access token (ej: 15m, 1h)
JWT_REFRESH_TOKEN_TTL=168h  # Tiempo de vida del refresh token (ej: 168h = 7 días)

# Base de datos
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

### Archivo de Configuración (config/app.dev.yaml)

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

### Mejores Prácticas de Seguridad

| Configuración           | Desarrollo    | Producción                   |
| ----------------------- | ------------- | ---------------------------- |
| `JWT_SECRET`            | Cualquiera    | **String aleatorio 256-bit** |
| `JWT_ACCESS_TOKEN_TTL`  | 15m           | 5m - 15m                     |
| `JWT_REFRESH_TOKEN_TTL` | 168h (7 días) | 7-30 días                    |
| `DB_SSLMODE`            | disable       | **require o verify-full**    |
| `SERVER_HOST`           | 0.0.0.0       | 0.0.0.0 o IP específica      |

**Configuraciones Críticas de Producción:**

1. Generar JWT secret seguro: `openssl rand -base64 32`
2. Usar solo HTTPS (certificados TLS/SSL)
3. Habilitar SSL de base de datos (`DB_SSLMODE=require`)
4. Establecer TTL corto para access token (5-15 minutos)
5. Monitorear intentos de inicio de sesión fallidos (rate limiting)

---

## Documentación Relacionada

- [AUTHORIZATION.md](AUTHORIZATION.es.md) - Sistema de permisos RBAC (lo que sucede DESPUÉS de la autenticación)
- [CREDENTIALS.md](CREDENTIALS.es.md) - Usuarios y roles predeterminados para desarrollo/pruebas
- [LOGGING.md](LOGGING.es.md) - Logging estructurado con contexto de autenticación
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.es.md) - Arquitectura del sistema y módulos principales
