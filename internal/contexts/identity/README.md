# Identity Context (Bounded Context)

**Domain:** User authentication, profiles, and contact information  
**Ubiquitous Language:** User, Profile, Contact, Session, Credentials  
**Status:** ✅ Production-ready (User, Contact aggregates) | 🚧 Profile aggregate in progress

---

## 📌 Overview

The **Identity Context** manages everything related to user identity, authentication, authorization (RBAC), and personal contact information. This is the **Core Domain** for user management in Promenade CRM, following **Domain-Driven Design (DDD)** principles with clear bounded context separation.

### 🎯 Bounded Context Responsibilities

✅ **What Identity Context DOES:**

- ✅ User registration and authentication (login/logout)
- ✅ Session management (JWT tokens, token refresh)
- ✅ Role-Based Access Control (RBAC) - users, roles, permissions
- ✅ Contact information management (email, phone, address)
- ✅ User profile management (display name, bio, avatar, timezone, language)
- ✅ User credentials and password management (bcrypt hashing)
- ✅ Contact verification (email confirmation, phone OTP)

❌ **What Identity Context DOES NOT DO:**

- ❌ Billing and subscriptions → **Billing Context**
- ❌ Customer relationships (CRM) → **Customer Management Context**
- ❌ Orders and contracts → **Order Management Context**
- ❌ Business analytics → **Analytics Context**
- ❌ Notifications → **Notification Context** (Identity publishes events only)

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

**Entity Structure:**

```go
type Profile struct {
    aggregate.BaseAggregate

    // Identity
    ID        uuidv7.UUID
    UserID    uuidv7.UUID // references User aggregate

    // Personal Info
    FirstName string
    LastName  string
    FullName  string // computed

    // Business Info
    Company   string
    Position  string
    Bio       string

    // Avatar
    AvatarURL string

    // Privacy
    IsPublic  bool

    // Lifecycle
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Business Rules:**

- One profile per user
- Full name is computed: `FirstName + LastName`
- Bio limited to 500 characters
- Avatar URL must be valid HTTPS URL

**Methods:**

```go
func (p *Profile) UpdatePersonalInfo(firstName, lastName string) error
func (p *Profile) UpdateBusinessInfo(company, position string) error
func (p *Profile) SetAvatar(url string) error
func (p *Profile) MakePublic() error
func (p *Profile) MakePrivate() error
```

---

### 3. Contact Aggregate ⭐

**Aggregate Root:** `Contact`  
**Purpose:** Contact information management  
**Status:** 🎯 **Ready for implementation (Day 3)**

**Entity Structure:**

```go
type Contact struct {
    aggregate.BaseAggregate

    // Identity
    ID        uuidv7.UUID
    UserID    uuidv7.UUID // references User aggregate

    // Contact Type
    Type      ContactType // primary, home, work, other
    Label     string      // e.g., "Home Phone", "Work Email"

    // Contact Details (value objects)
    Email     *valueobject.Email   // optional
    Phone     *valueobject.Phone   // optional
    Address   *valueobject.Address // optional

    // Flags
    IsPrimary bool // only one primary contact per user
    IsPublic  bool // visible to other users

    // Lifecycle
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Business Rules:**

- At least one of Email, Phone, or Address must be present
- Only ONE primary contact per user per type
- Primary contacts cannot be deleted (must set another as primary first)
- Public contacts visible to all users, private only to owner

**Methods:**

```go
func (c *Contact) SetAsPrimary() error
func (c *Contact) UpdateEmail(email valueobject.Email) error
func (c *Contact) UpdatePhone(phone valueobject.Phone) error
func (c *Contact) UpdateAddress(address valueobject.Address) error
func (c *Contact) MakePublic() error
func (c *Contact) MakePrivate() error
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

### Contact Management ⭐

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

### Contacts Table ⭐

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

### User Management

```
POST   /api/v1/identity/register        # Register new user
POST   /api/v1/identity/login           # Login
POST   /api/v1/identity/logout          # Logout
POST   /api/v1/identity/refresh         # Refresh token
GET    /api/v1/identity/me              # Get current user
PUT    /api/v1/identity/password        # Change password
```

### Profile Management

```
GET    /api/v1/identity/profile         # Get own profile
PUT    /api/v1/identity/profile         # Update profile
GET    /api/v1/identity/profiles/:id    # Get public profile
```

### Contact Management ⭐ (Day 3)

```
GET    /api/v1/identity/contacts        # List own contacts
POST   /api/v1/identity/contacts        # Create contact
GET    /api/v1/identity/contacts/:id    # Get contact
PUT    /api/v1/identity/contacts/:id    # Update contact
DELETE /api/v1/identity/contacts/:id    # Delete contact
POST   /api/v1/identity/contacts/:id/primary  # Set as primary
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
  ├── entity/
  │   ├── user_profile.go
  │   └── user_contact.go    ← Will migrate to contexts/identity/contact/
  └── ...
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

**Status:** 🎯 Ready for Contact aggregate implementation  
**Next:** Day 3 - Implement Contact aggregate with 40 tests  
**Documentation:** See `docs/BOUNDED_CONTEXTS_GUIDE.md` for complete DDD guide
