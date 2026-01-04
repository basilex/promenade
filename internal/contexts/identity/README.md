# Identity Context (Bounded Context)

**Domain:** User authentication, profiles, and contact information  
**Ubiquitous Language:** User, Profile, Contact, Session, Credentials  
**Status:** ✅ Production-ready (Contact, Profile, User aggregates)

---

## Overview

The **Identity Context** manages everything related to user identity, authentication, authorization (RBAC), and personal contact information. This is the **Core Domain** for user management in Promenade Platform, following **Domain-Driven Design (DDD)** principles with clear bounded context separation.

###  Bounded Context Responsibilities

 **What Identity Context DOES:**

- ✅ Contact information management (email, phone, address)
- ✅ Contact verification (email confirmation, phone OTP)
- ✅ User profile management (display name, bio, avatar, personal info, social links)
- ✅ Profile visibility control (public/private profiles)
- ✅ User registration and authentication (login/logout)
- ✅ User credentials and password management (bcrypt hashing)
- ✅ Account status management (active, suspended, banned, locked)
- 🔄 Session management (JWT tokens, token refresh) - next
- 🔄 Role-Based Access Control (RBAC) - users, roles, permissions - next

 **What Identity Context DOES NOT DO:**

-  Billing and subscriptions → **Billing Context**
-  Customer relationships (CRM) → **Customer Management Context**
-  Orders and contracts → **Order Management Context**
-  Business analytics → **Analytics Context**
-  Notifications → **Notification Context** (Identity publishes events only)

---

## Aggregates

### 1. User Aggregate

**Aggregate Root:** `User`  
**Purpose:** Authentication and authorization

**Entity Structure:**

```go
type User struct {
    aggregate.BaseAggregate

    // Identity
    ID           uuidv7.UUID
    Email        valueobject.Email
    PasswordHash string

    // Status
    Status       UserStatus // active, suspended, deleted

    // RBAC
    Roles        []Role
    Permissions  []Permission

    // Sessions
    sessions     []Session  // private, accessed via methods

    // Lifecycle
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    *time.Time // soft delete
}
```

**Business Rules:**

- Email must be unique
- Password must meet complexity requirements
- User can have multiple active sessions
- Suspended users cannot login
- Soft delete preserves audit trail

**Methods:**

```go
func (u *User) Login(password string) (*Session, error)
func (u *User) Logout(sessionID string) error
func (u *User) Suspend(reason string, until *time.Time) error
func (u *User) Activate() error
func (u *User) ChangePassword(oldPassword, newPassword string) error
func (u *User) AddRole(role Role) error
func (u *User) HasPermission(permission string) bool
```

---

### 2. Profile Aggregate 

**Aggregate Root:** `Profile`  
**Purpose:** User profile information (personal and business)  
**Status:**  **Production-ready** (Day 3)

**Entity Structure:**

```go
type Profile struct {
    aggregate.BaseAggregate

    // Identity
    ID          uuidv7.UUID
    UserID      uuidv7.UUID // references User aggregate
    DisplayName string      // public display name

    // Personal Info
    FirstName  string
    LastName   string
    MiddleName string
    Gender     Gender // male, female, other, not_specify
    DateOfBirth *time.Time

    // Business Info
    Bio       string // max 500 characters
    AvatarURL string // profile picture

    // Localization
    Timezone string // IANA timezone (e.g., "Europe/Kyiv")
    Language string // ISO 639-1 code (e.g., "uk", "en")
    Country  string // ISO 3166-1 alpha-2 (e.g., "UA", "GB")

    // Social Links
    Website   string
    LinkedIn  string
    Twitter   string
    GitHub    string
    Facebook  string
    Instagram string

    // Status Flags
    IsPublic   bool // profile visibility
    IsActive   bool // profile activation status
    IsBanned   bool // moderation flag
    IsVerified bool // verified profile badge

    // Lifecycle
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time // soft delete
}
```

**Business Rules:**

- One profile per user (1:1 relationship with User)
- Display name is required (1-100 characters)
- Bio limited to 500 characters
- Social links must be valid HTTPS URLs
- Gender enum: male, female, other, not_specify
- Public profiles visible to all, private only to owner
- Banned profiles hidden from public listings
- Soft delete preserves audit trail

**Methods:**

```go
// Creation
func NewProfile(userID uuidv7.UUID, displayName string) (*Profile, error)

// Updates
func (p *Profile) UpdateDisplayName(displayName string) error
func (p *Profile) UpdateBio(bio string) error
func (p *Profile) UpdateAvatar(avatarURL string) error
func (p *Profile) UpdatePersonalInfo(firstName, lastName, middleName string) error
func (p *Profile) UpdateGender(gender Gender) error
func (p *Profile) UpdateDateOfBirth(dateOfBirth *time.Time) error
func (p *Profile) UpdateLocalization(timezone, language, country string) error
func (p *Profile) UpdateSocialLinks(website, linkedin, twitter, github, facebook, instagram string) error

// Visibility
func (p *Profile) SetPublic()
func (p *Profile) SetPrivate()

// Status
func (p *Profile) Activate()
func (p *Profile) Deactivate()

// Validation
func (p *Profile) Validate() error
```

**HTTP Endpoints:**

- `POST /api/v1/identity/profiles` - Create profile
- `GET /api/v1/identity/profiles/:id` - Get profile by ID
- `GET /api/v1/identity/profiles/user/:user_id` - Get profile by user ID
- `GET /api/v1/identity/profiles` - List public profiles (paginated)
- `PUT /api/v1/identity/profiles/:id/display-name` - Update display name
- `PUT /api/v1/identity/profiles/:id/bio` - Update bio
- `PUT /api/v1/identity/profiles/:id/avatar` - Update avatar
- `PUT /api/v1/identity/profiles/:id/personal-info` - Update personal info
- `PUT /api/v1/identity/profiles/:id/gender` - Update gender
- `PUT /api/v1/identity/profiles/:id/date-of-birth` - Update date of birth
- `PUT /api/v1/identity/profiles/:id/localization` - Update localization
- `PUT /api/v1/identity/profiles/:id/social-links` - Update social links
- `PUT /api/v1/identity/profiles/:id/public` - Set public
- `PUT /api/v1/identity/profiles/:id/private` - Set private
- `DELETE /api/v1/identity/profiles/:id` - Soft delete profile

**Database:**

- Table: `identity_profiles`
- Migration: `migrations/identity/000005_profiles.up.sql`
- Indexes: user_id, is_public, created_at

**Tests:**

- ✅ Unit tests: `entity_test.go` (35 tests, 96% coverage)
- ✅ Unit tests: `usecase_test.go` (52 tests, 70% coverage)
- ✅ Integration tests: `test/integration/contexts/identity/profile/repository_test.go` (6 test functions, 17 subtests)

---

### 3. Contact Aggregate 

**Aggregate Root:** `Contact`  
**Purpose:** Contact information management  
**Status:**  **Production-ready** (Day 2)

**Entity Structure:**

```go
type Contact struct {
    aggregate.BaseAggregate

    // Identity
    ID     uuidv7.UUID
    UserID uuidv7.UUID // references User aggregate

    // Contact Type
    Type  ContactType // email, phone, address
    Label string      // e.g., "Home Phone", "Work Email"

    // Contact Details (value objects - only one populated based on Type)
    Email   *valueobject.Email   // for email contacts
    Phone   *valueobject.Phone   // for phone contacts
    Address *valueobject.Address // for address contacts

    // Flags
    IsPrimary  bool // only one primary contact per user per type
    IsVerified bool // email confirmed, phone OTP verified
    IsPublic   bool // visible to other users

    // Lifecycle
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time // soft delete
}
```

**Business Rules:**

- Exactly ONE of Email, Phone, or Address must be present (enforced by factory methods)
- Only ONE primary contact per user per type
- Primary contacts cannot be deleted (must set another as primary first)
- Public contacts visible to all users, private only to owner
- Email contacts require verification
- Phone contacts support OTP verification

**Methods:**

```go
// Factory methods (enforce one contact type rule)
func NewEmailContact(userID uuidv7.UUID, email, label string) (*Contact, error)
func NewPhoneContact(userID uuidv7.UUID, phone, label string) (*Contact, error)
func NewAddressContact(userID uuidv7.UUID, street, city, country, postalCode, label string) (*Contact, error)

// Updates
func (c *Contact) SetEmail(email string) error
func (c *Contact) SetPhone(phone string) error
func (c *Contact) SetAddress(street, city, country, postalCode string) error
func (c *Contact) UpdateLabel(label string)

// Primary flag
func (c *Contact) SetAsPrimary()
func (c *Contact) UnsetPrimary()

// Verification
func (c *Contact) Verify()
func (c *Contact) Unverify()

// Visibility
func (c *Contact) SetPublic()
func (c *Contact) SetPrivate()

// Validation
func (c *Contact) Validate() error
```

**Implementation Plan (Day 3):**

1. Create `contact/entity.go` with Contact aggregate
2. Create `contact/repository.go` interface
3. Create `contact/repository_impl.go` (PostgreSQL)
4. Create `contact/usecase.go` with business logic
5. Create `contact/handler.go` with HTTP endpoints
6. Write 40 comprehensive tests

---

## Value Objects

Identity context uses shared value objects from `pkg/valueobject/`:

- **Email** - RFC-compliant email validation
- **Phone** - E.164 format phone numbers
- **Address** - Multi-line addresses with validation

---

## Use Cases

### User Management

```go
type RegisterUserUseCase struct {
    userRepo    UserRepository
    profileRepo ProfileRepository
    eventBus    bus.IBus
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, req RegisterUserRequest) (*User, error) {
    // 1. Validate email uniqueness
    // 2. Create User aggregate
    // 3. Hash password
    // 4. Create default Profile
    // 5. Persist both
    // 6. Publish user.registered event
}
```

### Profile Management

```go
type UpdateProfileUseCase struct {
    profileRepo ProfileRepository
    eventBus    bus.IBus
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, req UpdateProfileRequest) (*Profile, error) {
    // 1. Load profile
    // 2. Validate ownership
    // 3. Update fields
    // 4. Persist
    // 5. Publish profile.updated event
}
```

### Contact Management 

```go
type CreateContactUseCase struct {
    contactRepo ContactRepository
    eventBus    bus.IBus
}

func (uc *CreateContactUseCase) Execute(ctx context.Context, req CreateContactRequest) (*Contact, error) {
    // 1. Validate user exists
    // 2. Create Contact aggregate
    // 3. Validate business rules (at least one field)
    // 4. Handle primary contact logic
    // 5. Persist
    // 6. Publish contact.created event
}
```

---

## Domain Events

### Published Events

```go
// User events
type UserRegisteredEvent struct {
    UserID    uuidv7.UUID
    Email     string
    Timestamp time.Time
}

type UserSuspendedEvent struct {
    UserID    uuidv7.UUID
    Reason    string
    Until     *time.Time
    Timestamp time.Time
}

// Profile events
type ProfileUpdatedEvent struct {
    ProfileID uuidv7.UUID
    UserID    uuidv7.UUID
    Changes   map[string]interface{}
    Timestamp time.Time
}

// Contact events
type ContactCreatedEvent struct {
    ContactID uuidv7.UUID
    UserID    uuidv7.UUID
    Type      string
    IsPrimary bool
    Timestamp time.Time
}
```

### Event Handlers

Other contexts can subscribe to Identity events:

```go
// Billing context subscribes to user registration
eventBus.Subscribe("user.registered", func(event UserRegisteredEvent) {
    // Create default free subscription for new user
    billingService.CreateDefaultSubscription(event.UserID)
})

// Customer Management context subscribes to profile updates
eventBus.Subscribe("profile.updated", func(event ProfileUpdatedEvent) {
    // Update customer cache if profile is a customer
    customerCache.Invalidate(event.UserID)
})
```

---

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
```

### Profiles Table

```sql
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    company VARCHAR(255),
    position VARCHAR(255),
    bio TEXT,
    avatar_url TEXT,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX idx_profiles_user_id ON user_profiles(user_id);
```

### Contacts Table 

```sql
CREATE TABLE user_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- primary, home, work, other
    label VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    address JSONB,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_contacts_user_id ON user_contacts(user_id);
CREATE INDEX idx_contacts_is_primary ON user_contacts(is_primary);

-- Unique constraint: only one primary contact per user per type
CREATE UNIQUE INDEX idx_contacts_primary_unique
ON user_contacts(user_id, type)
WHERE is_primary = true;
```

---

## API Endpoints

### Authentication (Public)

```
POST   /api/v1/identity/users/register       # Register new user
POST   /api/v1/identity/users/login          # Login (returns JWT tokens)
POST   /api/v1/identity/auth/refresh         # Refresh access token
```

**Example: User Registration**

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "019b65dc-ec43-740e-a1bd-bbac7179cbd3",
    "email": "user@example.com",
    "status": "active",
    "email_verified": false,
    "created_at": "2025-12-28T18:48:55Z",
    "updated_at": "2025-12-28T18:48:55Z"
  }
}
```

**Example: User Login**

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-12-28T19:04:06Z",
    "token_type": "Bearer",
    "user": {
      "id": "019b65dc-ec43-740e-a1bd-bbac7179cbd3",
      "email": "user@example.com",
      "status": "active",
      "email_verified": false,
      "last_login_at": "2025-12-28T18:49:06Z"
    }
  }
}
```

**Example: Refresh Access Token**

```bash
curl -X POST http://localhost:8081/api/v1/identity/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

**Response:** Same as login response with new token pair.

---

### User Management (Protected with JWT)

All endpoints below require `Authorization: Bearer <access_token>` header.

```
GET    /api/v1/identity/users                 # List users
GET    /api/v1/identity/users/:id             # Get user by ID
GET    /api/v1/identity/users/email/:email    # Get user by email
POST   /api/v1/identity/users/:id/verify-email      # Verify email
POST   /api/v1/identity/users/:id/change-password   # Change password
POST   /api/v1/identity/users/:id/suspend           # Suspend user
POST   /api/v1/identity/users/:id/ban               # Ban user
POST   /api/v1/identity/users/:id/activate          # Activate user
POST   /api/v1/identity/users/:id/unlock            # Unlock user
```

**Example: Get User with JWT**

```bash
curl -X GET http://localhost:8081/api/v1/identity/users/:id \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "019b65dc-ec43-740e-a1bd-bbac7179cbd3",
    "email": "user@example.com",
    "status": "active",
    "email_verified": false,
    "last_login_at": "2025-12-28T18:49:06Z"
  }
}
```

---

### Profile Management

**Public Routes:**
```
GET    /api/v1/identity/profiles              # List public profiles
GET    /api/v1/identity/profiles/:id          # Get profile by ID (public)
GET    /api/v1/identity/profiles/user/:user_id # Get profile by user ID
```

**Protected Routes (JWT required):**
```
POST   /api/v1/identity/profiles              # Create profile
DELETE /api/v1/identity/profiles/:id          # Delete profile
PUT    /api/v1/identity/profiles/:id/display-name    # Update display name
PUT    /api/v1/identity/profiles/:id/bio             # Update bio
PUT    /api/v1/identity/profiles/:id/avatar          # Update avatar
PUT    /api/v1/identity/profiles/:id/personal-info   # Update personal info
PUT    /api/v1/identity/profiles/:id/gender          # Update gender
PUT    /api/v1/identity/profiles/:id/date-of-birth   # Update date of birth
PUT    /api/v1/identity/profiles/:id/localization    # Update localization
PUT    /api/v1/identity/profiles/:id/social-links    # Update social links
PUT    /api/v1/identity/profiles/:id/public          # Set profile public
PUT    /api/v1/identity/profiles/:id/private         # Set profile private
```

---

### Contact Management (Protected with JWT)

All contact endpoints require JWT authentication.

```
POST   /api/v1/identity/contacts              # Create contact
GET    /api/v1/identity/contacts              # List contacts
GET    /api/v1/identity/contacts/:id          # Get contact
PUT    /api/v1/identity/contacts/:id          # Update contact
DELETE /api/v1/identity/contacts/:id          # Delete contact
PUT    /api/v1/identity/contacts/:id/verify   # Verify contact
PUT    /api/v1/identity/contacts/:id/primary  # Set as primary
```

**Example: Create Contact**

```bash
curl -X POST http://localhost:8081/api/v1/identity/contacts \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "019b65dc-ec43-740e-a1bd-bbac7179cbd3",
    "type": "email",
    "email": "work@example.com",
    "label": "Work",
    "is_primary": false
  }'
```

---

## Testing Strategy

### Aggregate Tests (Domain Logic)

```go
// contact_test.go
func TestContact_Validate(t *testing.T)
func TestContact_SetAsPrimary(t *testing.T)
func TestContact_UpdateEmail(t *testing.T)
```

### Repository Tests (Integration)

```go
// repository_impl_test.go
func TestContactRepository_Create(t *testing.T)
func TestContactRepository_GetByUserID(t *testing.T)
func TestContactRepository_GetPrimaryContact(t *testing.T)
```

### Use Case Tests (Business Logic)

```go
// usecase_test.go
func TestCreateContactUseCase_Execute(t *testing.T)
func TestSetPrimaryContactUseCase_Execute(t *testing.T)
```

### Handler Tests (HTTP)

```go
// handler_test.go
func TestContactHandler_Create(t *testing.T)
func TestContactHandler_List(t *testing.T)
```

**Target:** 40 tests for Contact aggregate implementation

---

## Migration from Modules

Current module structure:

```
internal/modules/profiles/
   entity/
      user_profile.go
      user_contact.go    ← Will migrate to contexts/identity/contact/
   ...
```

Migration plan:

1. **Phase 1:** Implement Contact aggregate in `contexts/identity/contact/`
2. **Phase 2:** Update routes to use new context handlers
3. **Phase 3:** Keep module handlers as facade (backward compatibility)
4. **Phase 4:** Deprecate module after full migration

---

## Dependencies

**Shared Kernel:**

- `pkg/aggregate` - BaseAggregate pattern
- `pkg/valueobject` - Email, Phone, Address
- `pkg/bus` - Event bus for domain events

**Infrastructure:**

- `internal/infrastructure/database` - Database connection
- `pkg/jwt` - JWT token management

**External:**

- PostgreSQL - Data persistence
- Redis - Session storage (optional)

---

## Next Steps

### Day 3: Contact Aggregate Implementation

1. **Create files:**

   - `contact/entity.go` - Contact aggregate root
   - `contact/repository.go` - Repository interface
   - `contact/repository_impl.go` - PostgreSQL implementation
   - `contact/usecase.go` - Business logic (Create, Update, Delete, SetPrimary)
   - `contact/handler.go` - HTTP handlers
   - `contact/dto.go` - Request/Response DTOs

2. **Tests:**

   - 10 aggregate tests (domain logic)
   - 10 repository tests (database operations)
   - 10 use case tests (business logic)
   - 10 handler tests (HTTP endpoints)
   - **Total:** 40 tests

3. **Migration:**
   - Create migration: `000009_identity_contacts.up.sql`
   - Register routes in main.go
   - Update Swagger documentation

---

**Status:**  Ready for Contact aggregate implementation  
**Next:** Day 3 - Implement Contact aggregate with 40 tests  
**Documentation:** See `docs/BOUNDED_CONTEXTS_GUIDE.md` for complete DDD guide
