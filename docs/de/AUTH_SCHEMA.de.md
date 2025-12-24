# Authentifizierungssystem-Architektur

🇬🇧 [English](AUTH_SCHEMA.de.md) | [🇺🇦 Українська](../uk/AUTH_SCHEMA.uk.md) | 🇩🇪 **Deutsch** | [🇵🇹 Português](../pt/AUTH_SCHEMA.pt.md) | [🇪🇸 Español](../es/AUTH_SCHEMA.es.md)

Vollständige Referenz für das Authentifizierungssystem in Promenade. Behandelt Benutzerregistrierung, Login, Session-Management, Token-Verarbeitung und Sicherheitsmechanismen.

## Inhaltsverzeichnis

- [Übersicht](#übersicht)
- [Datenbankschema](#datenbankschema)
- [Benutzerstatus und Lebenszyklus](#benutzerstatus-und-lebenszyklus)
- [Authentifizierungsablauf](#authentifizierungsablauf)
- [Session-Verwaltung](#session-verwaltung)
- [Token-System](#token-system)
- [Sicherheitsmechanismen](#sicherheitsmechanismen)
- [API-Endpunkte](#api-endpunkte)
- [Fehlerbehandlung](#fehlerbehandlung)
- [Konfiguration](#konfiguration)

---

## Übersicht

**Komponenten des Authentifizierungssystems:**

| Komponente      | Zweck                                        | Technologie         |
| --------------- | -------------------------------------------- | ------------------- |
| User Entity     | Haupt-Account mit Passwort                   | PostgreSQL, bcrypt  |
| Session Entity  | Speicherung von Refresh-Tokens               | PostgreSQL, SHA-256 |
| JWT Manager     | Generierung/Validierung von Access-Tokens    | HMAC-SHA256         |
| Auth UseCase    | Geschäftslogik aller Auth-Operationen        | Go                  |
| Auth Middleware | Anfrage-Authentifizierung                    | Gin Middleware      |
| Event Bus       | Asynchrone Benachrichtigungen (E-Mail, Logs) | Memory/Redis        |

**Hauptfunktionen:**

- ✅ JWT-basierte Authentifizierung (Access + Refresh Tokens)
- ✅ Refresh Token Rotation (Security Best Practice)
- ✅ Concurrent Session Management (max. 5 pro Benutzer)
- ✅ Benutzer-Zustandsautomat (unverified → active → suspended/banned)
- ✅ Passwort-Hashing mit bcrypt (Kosten 10)
- ✅ Refresh Token Hashing mit SHA-256
- ✅ Automatische Session-Bereinigung bei Statusänderungen
- ✅ Asynchrone E-Mail-Benachrichtigungen über Event Bus

---

## Datenbankschema

### Tabelle `core_users`

```sql
CREATE TABLE core_users (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    email              VARCHAR(255) NOT NULL UNIQUE,
    name               VARCHAR(255) NOT NULL,
    password           VARCHAR(255) NOT NULL,  -- bcrypt Hash
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

### Tabelle `core_user_sessions`

```sql
CREATE TABLE core_user_sessions (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id       UUID NOT NULL REFERENCES core_users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,  -- SHA-256 Hash
    user_agent    TEXT,
    ip_address    VARCHAR(45),
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON core_user_sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON core_user_sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON core_user_sessions(expires_at);
```

**Zusätzliche Tabellen (noch nicht implementiert):**

- `core_password_reset_tokens` - Passwort-Zurücksetzungsprozess
- `core_email_verification_tokens` - E-Mail-Verifizierungsprozess
- `core_login_attempts` - Brute-Force-Schutz

---

## Benutzerstatus und Lebenszyklus

### Benutzerstatus-Aufzählung

```go
type UserStatus string

const (
    UserStatusUnverified UserStatus = "unverified" // Registriert, aber E-Mail nicht verifiziert
    UserStatusActive     UserStatus = "active"     // E-Mail verifiziert und Account aktiv
    UserStatusSuspended  UserStatus = "suspended"  // Temporär gesperrt (kann reaktiviert werden)
    UserStatusBanned     UserStatus = "banned"     // Dauerhaft gesperrt
    UserStatusInactive   UserStatus = "inactive"   // Vom Benutzer deaktiviert (kann reaktiviert werden)
)
```

### Zustandsübergänge

```
                    Register()
                        │
                        ▼
                 ┌──────────────┐
                 │  unverified  │  ──────────────┐
                 └──────────────┘                │
                        │                        │
                 VerifyEmail()            Login() erlaubt
                        │                        │
                        ▼                        ▼
                 ┌──────────────┐         Benutzer kann sich anmelden
                 │    active    │         (unverified oder active)
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

### CanLogin() Logik

```go
func (u *User) CanLogin() bool {
    return u.Status == UserStatusActive || u.Status == UserStatusUnverified
}
```

**Erlaubte Status:**

- ✅ `active` - Voller Zugriff
- ✅ `unverified` - Kann sich anmelden, aber Funktionen können eingeschränkt sein

**Gesperrte Status:**

- ❌ `suspended` - Gibt `ErrUserSuspended` zurück
- ❌ `banned` - Gibt `ErrUserBanned` zurück
- ❌ `inactive` - Gibt `ErrUserNotActive` zurück

---

## Authentifizierungsablauf

### 1. Registrierungsablauf

```
Client                  API                    UseCase                Datenbank         Event Bus
  │                      │                        │                        │                 │
  │  POST /auth/register │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │                      │  Register(email, name, pwd)                     │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  (nicht gefunden - OK) │                 │
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
  │<─────────────────────│                        │                        │  sendet         │
  │  {id, email, name}   │                        │                        │  Willkommens-E-Mail│
```

**Wichtige Punkte:**

- Passwort wird mit `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)` gehasht
- Benutzer wird mit `status = "unverified"` erstellt
- UUID v7 wird für Benutzer-ID generiert (zeitlich sortiert)
- Event wird asynchron veröffentlicht - Registrierung wartet nicht auf E-Mail
- Email Worker verarbeitet `user.registered` Event im Hintergrund

---

### 2. Login-Ablauf

```
Client                  API                    UseCase                Datenbank         Session
  │                      │                        │                        │                 │
  │  POST /auth/login    │                        │                        │                 │
  │─────────────────────>│                        │                        │                 │
  │  {email, password}   │  Login(email, pwd, ua, ip)                      │                 │
  │                      │───────────────────────>│                        │                 │
  │                      │                        │  GetByEmail(email)     │                 │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  Benutzer gefunden     │                 │
  │                      │                        │                        │                 │
  │                      │                        │  bcrypt.Compare(pwd, hash)               │
  │                      │                        │  OK ✓                  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  JA ✓                  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  JWTManager.GenerateAccessToken()        │
  │                      │                        │  crypto/rand 32 Bytes für Refresh Token  │
  │                      │                        │  SHA-256(refresh_token)                  │
  │                      │                        │                        │                 │
  │                      │                        │  CountUserSessions(user_id)              │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  count = 5 (Limit!)    │                 │
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

**Wichtige Punkte:**

- **Passwortvalidierung:** `bcrypt.CompareHashAndPassword(hash, password)`
- **Statusprüfung:** Muss `active` oder `unverified` sein
- **Access Token:** JWT mit 15 Minuten TTL (Standard)
- **Refresh Token:** 32 Byte zufällig + base64, 7 Tage TTL (Standard)
- **Refresh Token Speicherung:** SHA-256 gehasht vor DB-Speicherung
- **Session-Limit:** Max. 5 gleichzeitige Sessions pro Benutzer
- **Älteste löschen:** Bei Überschreitung wird älteste Session automatisch gelöscht
- **Letzter Login:** Wird asynchron aktualisiert (blockiert Antwort nicht)

---

### 3. Token-Aktualisierungsablauf

```
Client                  API                    UseCase                Datenbank         Session
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
  │                      │                        │  Session gefunden      │                 │
  │                      │                        │                        │                 │
  │                      │                        │  session.IsExpired()?  │                 │
  │                      │                        │  NEIN ✓                │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GetByID(session.user_id)                │
  │                      │                        │───────────────────────>│                 │
  │                      │                        │<───────────────────────│                 │
  │                      │                        │  Benutzer gefunden     │                 │
  │                      │                        │                        │                 │
  │                      │                        │  user.CanLogin()?      │                 │
  │                      │                        │  JA ✓                  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  GenerateAccessToken() │                 │
  │                      │                        │  crypto/rand neuer Refresh               │
  │                      │                        │  SHA-256(new_refresh)  │                 │
  │                      │                        │                        │                 │
  │                      │                        │  Update(session)       │  ← TOKEN ROTATION
  │                      │                        │  - neuer Refresh Hash  │                 │
  │                      │                        │  - neue expires_at     │                 │
  │                      │                        │───────────────────────────────────────>│
  │                      │                        │                        │                 │
  │                      │<───────────────────────│                        │                 │
  │  200 OK              │  (new_access, new_refresh)                      │                 │
  │<─────────────────────│                        │                        │                 │
  │  {access_token,      │                        │                        │                 │
  │   refresh_token}     │                        │                        │                 │
```

**Wichtige Punkte:**

- **Token-Rotation:** Alter Refresh Token wird ungültig, neuer wird ausgegeben
- **Sicherheit:** Einmalige Refresh Tokens verhindern Replay-Angriffe
- **Session-Wiederverwendung:** Aktualisiert bestehende Session statt Delete+Create (Performance)
- **Ablaufprüfung:** Abgelaufene Sessions werden automatisch gelöscht
- **Statusprüfung:** Benutzer muss weiterhin Login-Berechtigung haben

---

### 4. Logout-Ablauf

```
Client                  API                    UseCase                Session
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
  │                      │                        │  Session gefunden      │
  │                      │                        │                        │
  │                      │                        │  Delete(session.id)    │
  │                      │                        │───────────────────────>│
  │                      │                        │                        │
  │                      │<───────────────────────│                        │
  │  200 OK              │                        │                        │
  │<─────────────────────│                        │                        │
  │  {message: "success"}│                        │                        │
```

**Wichtige Punkte:**

- **Session-Löschung:** Ungültig macht Refresh Token sofort
- **Access Token:** Bleibt bis Ablauf gültig (zustandsloses JWT)
- **Client-Verantwortung:** Client muss beide Tokens verwerfen

---

## Session-Verwaltung

### Concurrent Session Limit

```go
const MaxConcurrentSessions = 5
```

**Verhalten:**

- Benutzer kann maximal 5 aktive Sessions auf verschiedenen Geräten haben
- Bei 6. Login: Älteste Session wird automatisch gelöscht
- Verhindert unbegrenzte Token-Generierung-Angriffe

### Session Entity

```go
type Session struct {
    ID           UUID      `db:"id"`
    UserID       UUID      `db:"user_id"`
    RefreshToken string    `db:"refresh_token"`  // SHA-256 gehasht
    UserAgent    *string   `db:"user_agent"`     // Browser/Geräteinformationen
    IPAddress    *string   `db:"ip_address"`     // Client-IP
    ExpiresAt    time.Time `db:"expires_at"`     // Absoluter Ablauf
    CreatedAt    time.Time `db:"created_at"`     // Session-Start
}
```

### Session-Operationen

| Operation         | Zweck                          | Auslöser                     |
| ----------------- | ------------------------------ | ---------------------------- |
| Create            | Neue Session beim Login        | Login                        |
| Update            | Refresh Token Rotation         | Token-Aktualisierung         |
| Delete            | Logout einzelne Session        | Logout                       |
| DeleteByUserID    | Alle Benutzersessions ungültig | Suspend/Ban/Passwortwechsel  |
| GetOldestSession  | Älteste zum Löschen finden     | Session-Limit-Überschreitung |
| CountUserSessions | Limit prüfen                   | Login                        |

---

## Token-System

### JWT Access Token

**Eigenschaften:**

- **Algorithmus:** HS256 (HMAC-SHA256)
- **TTL:** 15 Minuten (Standard, konfigurierbar)
- **Speicherung:** Nur clientseitig (nicht in DB)
- **Validierung:** Signaturprüfung + Ablauf bei jeder Anfrage

**Claims-Struktur:**

```go
type Claims struct {
    UserID uuidv7.UUID `json:"user_id"`
    Email  string      `json:"email"`
    jwt.RegisteredClaims
}

// RegisteredClaims enthält:
// - iat (issued at - Ausstellungszeit)
// - exp (expires at - Ablaufzeit)
// - nbf (not before - Nicht vor)
```

**Beispiel-JWT:**

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

**Eigenschaften:**

- **Algorithmus:** crypto/rand (32 Bytes) + base64
- **TTL:** 7 Tage (Standard, konfigurierbar)
- **Speicherung:** Datenbank (SHA-256 gehasht)
- **Validierung:** Hash-Vergleich + Ablauf

**Generierung:**

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

### Token-Rotationsmuster

**Sicherheitsvorteil:** Verhindert Replay-Angriffe mit Refresh Tokens

1. Client sendet Refresh Token
2. Server validiert und gibt neue Token-Paare aus
3. **Alter Refresh Token wird sofort ungültig**
4. Client muss neuen Refresh Token für nächste Aktualisierung verwenden

**Angriffszenario-Prävention:**

- ❌ Angreifer stiehlt Refresh Token
- ❌ Angreifer versucht ihn zu verwenden
- ✅ Token wurde bereits vom legitimen Benutzer rotiert → **Angriff schlägt fehl**

---

## Sicherheitsmechanismen

### Passwortsicherheit

| Mechanismus | Implementierung                   | Zweck                      |
| ----------- | --------------------------------- | -------------------------- |
| Hashing     | `bcrypt` (Kosten 10)              | Einweg-Verschlüsselung     |
| Salt        | Automatisch (bcrypt-intern)       | Eindeutiger Hash für jeden |
| Validierung | `bcrypt.CompareHashAndPassword()` | Zeitkonstanter Vergleich   |

**Code:**

```go
// Hashing bei Registrierung
func (u *User) HashPassword(password string) error {
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedBytes)
    return nil
}

// Prüfung beim Login
func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
```

### Refresh Token Sicherheit

| Mechanismus        | Implementierung      | Zweck                        |
| ------------------ | -------------------- | ---------------------------- |
| Zufallsgenerierung | `crypto/rand` (32 B) | Kryptographisch sicher       |
| Hashing            | SHA-256              | Niemals Klartext speichern   |
| Token-Rotation     | Einmaltoken          | Ungültigkeit nach Verwendung |
| Ablauf             | 7 Tage TTL           | Begrenztes Einflussfenster   |

### JWT-Sicherheit

| Mechanismus | Implementierung   | Zweck                        |
| ----------- | ----------------- | ---------------------------- |
| Signatur    | HMAC-SHA256       | Fälschungsschutz             |
| Secret Key  | Umgebungsvariable | Signaturverifikation         |
| Kurze TTL   | 15 Minuten        | Minimiertes Einflussfenster  |
| Zustandslos | Kein DB-Zugriff   | Performance + Skalierbarkeit |

### Sicherheit administrativer Aktionen

**Automatische Session-Invalidierung:**

Wenn Admin Benutzer sperrt/bannt:

1. Benutzerstatus wird in DB geändert
2. **Alle Benutzersessions werden sofort gelöscht**
3. Benutzer kann Tokens nicht aktualisieren
4. Bestehende Access Tokens laufen natürlich ab (max. 15 Min)

```go
func (uc *authUseCase) SuspendUser(ctx context.Context, userID UUID, reason string, until *time.Time) error {
    // Benutzerstatus aktualisieren
    if err := uc.userRepo.Suspend(ctx, userID, reason, until); err != nil {
        return err
    }

    // KRITISCH: Alle Sessions ungültig machen
    if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
        return err
    }

    // Asynchrone Benachrichtigung
    uc.eventBus.Publish(ctx, bus.TopicUserSuspended, event)
    return nil
}
```

---

## API-Endpunkte

### Öffentliche Endpunkte (Keine Authentifizierung erforderlich)

#### POST /api/v1/auth/register

Neuen Account registrieren.

**Anfrage:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Antwort (201 Created):**

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

Benutzer authentifizieren und Tokens erhalten.

**Anfrage:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Antwort (200 OK):**

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

Access Token über Refresh Token aktualisieren.

**Anfrage:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Antwort (200 OK):**

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

**Hinweis:** Alter Refresh Token wird ungültig (Token-Rotation).

---

### Geschützte Endpunkte (Authentifizierung erforderlich)

#### GET /api/v1/auth/me

Aktuelles Benutzerprofil abrufen.

**Anfrage:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Antwort (200 OK):**

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

Alle aktiven Sessions des aktuellen Benutzers abrufen.

**Anfrage:**

```bash
curl -X GET http://localhost:8080/api/v1/auth/sessions \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Antwort (200 OK):**

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

Logout und Refresh Token ungültig machen.

**Anfrage:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "refresh_token": "dGhpc2lzYXJhbmRvbXRva2Vu..."
  }'
```

**Antwort (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "Successfully logged out"
  }
}
```

---

### Administrative Endpunkte (Berechtigungen `users:suspend` / `users:ban` erforderlich)

#### POST /api/v1/auth/users/:id/suspend

Benutzeraccount temporär sperren.

**Anfrage:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/suspend \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Verstoß gegen Community-Regeln",
    "suspended_until": "2025-01-23T00:00:00Z"
  }'
```

**Antwort (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User suspended successfully"
  }
}
```

**Auswirkungen:**

- Benutzerstatus auf `suspended` geändert
- **Alle aktiven Sessions sofort gelöscht**
- Benutzer kann sich nicht anmelden bis `suspended_until` oder manuelle Reaktivierung

---

#### POST /api/v1/auth/users/:id/ban

Benutzeraccount dauerhaft sperren.

**Anfrage:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/ban \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "reason": "Spam und betrügerische Aktivitäten"
  }'
```

**Antwort (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User banned successfully"
  }
}
```

**Auswirkungen:**

- Benutzerstatus auf `banned` geändert
- **Alle aktiven Sessions sofort gelöscht**
- Benutzer kann sich nicht anmelden (manuelle Reaktivierung durch Admin erforderlich)

---

#### POST /api/v1/auth/users/:id/reactivate

Gesperrten/gebannten/inaktiven Benutzer reaktivieren.

**Anfrage:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/users/01936d6a-8f7c-7890-a1b2-c3d4e5f67890/reactivate \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Antwort (200 OK):**

```json
{
  "status": "success",
  "data": {
    "message": "User reactivated successfully"
  }
}
```

---

## Fehlerbehandlung

### Authentifizierungsfehler

| Fehlercode            | HTTP-Status | Nachricht                  | Grund                                   |
| --------------------- | ----------- | -------------------------- | --------------------------------------- |
| `INVALID_CREDENTIALS` | 401         | Invalid email or password  | Falsche E-Mail/Passwort-Kombination     |
| `EMAIL_EXISTS`        | 409         | Email already exists       | Registrierung mit existierender E-Mail  |
| `USER_NOT_ACTIVE`     | 403         | User account is not active | Status `inactive`                       |
| `USER_SUSPENDED`      | 403         | User account is suspended  | Status `suspended`                      |
| `USER_BANNED`         | 403         | User account is banned     | Status `banned`                         |
| `INVALID_TOKEN`       | 401         | Invalid or expired token   | Token-Validierung fehlgeschlagen        |
| `TOKEN_EXPIRED`       | 401         | Token has expired          | JWT oder Refresh Token abgelaufen       |
| `UNAUTHORIZED`        | 401         | Authentication required    | Fehlendes oder ungültiges Authorization |

### Fehlerantwortformat

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

### Validierungsfehler

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

## Konfiguration

### Umgebungsvariablen

```bash
# JWT-Konfiguration
JWT_SECRET=ihr-256-bit-geheimer-schluessel-aendern-in-produktion
JWT_ACCESS_TOKEN_TTL=15m    # Access Token Lebensdauer (z.B. 15m, 1h)
JWT_REFRESH_TOKEN_TTL=168h  # Refresh Token Lebensdauer (z.B. 168h = 7 Tage)

# Datenbank
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=promenade
DB_SSLMODE=disable

# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
```

### Konfigurationsdatei (config/app.dev.yaml)

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

### Best Practices für Sicherheit

| Einstellung             | Entwicklung   | Produktion                    |
| ----------------------- | ------------- | ----------------------------- |
| `JWT_SECRET`            | Beliebig      | **256-Bit zufälliger String** |
| `JWT_ACCESS_TOKEN_TTL`  | 15m           | 5m - 15m                      |
| `JWT_REFRESH_TOKEN_TTL` | 168h (7 Tage) | 7-30 Tage                     |
| `DB_SSLMODE`            | disable       | **require oder verify-full**  |
| `SERVER_HOST`           | 0.0.0.0       | 0.0.0.0 oder spezifische IP   |

**Kritische Produktionseinstellungen:**

1. Sicheren JWT Secret generieren: `openssl rand -base64 32`
2. Nur HTTPS verwenden (TLS/SSL-Zertifikate)
3. Datenbank-SSL aktivieren (`DB_SSLMODE=require`)
4. Kurze TTL für Access Token setzen (5-15 Minuten)
5. Fehlgeschlagene Login-Versuche überwachen (Rate Limiting)

---

## Verwandte Dokumentation

- [AUTHORIZATION.md](AUTHORIZATION.de.md) - RBAC-Berechtigungssystem (was NACH Authentifizierung passiert)
- [CREDENTIALS.md](CREDENTIALS.de.md) - Standard-Benutzer und Rollen für Entwicklung/Tests
- [LOGGING.md](LOGGING.de.md) - Strukturiertes Logging mit Authentifizierungskontext
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.de.md) - Systemarchitektur und Kernmodule
