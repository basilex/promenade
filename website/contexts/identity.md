# Identity Context

**Identity Context** manages user authentication, authorization, and contact information in Promenade Platform.

## Overview

The Identity Context is a critical bounded context responsible for:

- **User Management** - Registration, authentication, account lifecycle
- **Contact Management** - Email, phone, address information
- **Profile Management** - User profiles with personal and business info
- **Authentication** - JWT-based token authentication
- **Authorization** - Role-Based Access Control (RBAC)

## Aggregates

### User Aggregate

**Purpose**: User account lifecycle, authentication, password management

**Entity**:
```go
type User struct {
    ID              uuid.UUID
    Email           valueobject.Email  // Unique email address
    PasswordHash    string             // Bcrypt hash
    Name            string
    Status          UserStatus         // active, suspended, banned
    FailedLoginCount int
    LastLoginAt     *time.Time
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

**Lifecycle**:
```
New → Active → Suspended → Active
             → Banned
```

**Use Cases**:
- `Register(email, name, password)` - Create new user
- `Authenticate(email, password)` - Login with credentials
- `ChangePassword(userID, oldPassword, newPassword)` - Update password
- `SuspendUser(userID)` - Temporarily disable account
- `BanUser(userID)` - Permanently block account
- `UnlockUser(userID)` - Reset failed login count

**Domain Events**:
- `user.registered` - New user created
- `user.activated` - User activated
- `user.suspended` - User suspended
- `user.banned` - User banned
- `user.password_changed` - Password updated

### Contact Aggregate

**Purpose**: User contact information (email, phone, address)

**Entity**:
```go
type Contact struct {
    ID         uuid.UUID
    UserID     uuid.UUID
    Type       ContactType      // email, phone, address
    Label      string          // "Work", "Home", "Personal"
    
    // Value Objects (only one populated based on Type)
    Email      *valueobject.Email
    Phone      *valueobject.Phone
    Address    *valueobject.Address
    
    IsPrimary  bool  // Only one primary per user per type
    IsVerified bool  // Email confirmation, phone OTP
    IsPublic   bool  // Visible in public profile
}
```

**Factory Methods**:
```go
func NewEmailContact(userID uuid.UUID, email, label string) (*Contact, error)
func NewPhoneContact(userID uuid.UUID, phone, label string) (*Contact, error)
func NewAddressContact(userID uuid.UUID, address Address, label string) (*Contact, error)
```

**Use Cases**:
- `CreateEmailContact(userID, email, label)` - Add email
- `CreatePhoneContact(userID, phone, label)` - Add phone
- `CreateAddressContact(userID, address, label)` - Add address
- `VerifyContact(contactID)` - Mark as verified
- `SetAsPrimary(contactID)` - Set as primary contact
- `UpdateVisibility(contactID, isPublic)` - Change visibility
- `DeleteContact(contactID)` - Soft delete contact

**Domain Events**:
- `contact.created` - New contact added
- `contact.verified` - Contact verified
- `contact.updated` - Contact modified
- `contact.deleted` - Contact removed

### Profile Aggregate

**Purpose**: User profile information (personal, business, social)

**Entity**:
```go
type Profile struct {
    ID          uuid.UUID
    UserID      uuid.UUID  // 1:1 with User
    DisplayName string     // Public name
    
    // Personal Info
    FirstName   string
    LastName    string
    Gender      Gender  // male, female, other, not_specify
    DateOfBirth *time.Time
    
    // Business Info
    Bio       string  // Max 500 chars
    AvatarURL string
    
    // Localization
    Timezone string  // IANA (e.g., "Europe/Kyiv")
    Language string  // ISO 639-1 (e.g., "uk")
    Country  string  // ISO 3166-1 (e.g., "UA")
    
    // Social Links
    Website, LinkedIn, Twitter, GitHub string
    
    // Status
    IsPublic   bool  // Profile visibility
    IsActive   bool  // Activation status
    IsBanned   bool  // Moderation flag
    IsVerified bool  // Verified badge
}
```

**Use Cases**:
- `CreateProfile(userID, displayName)` - Create user profile
- `UpdatePersonalInfo(profileID, firstName, lastName, gender, dob)` - Update personal data
- `UpdateBusinessInfo(profileID, bio, avatarURL)` - Update business info
- `UpdateLocalization(profileID, timezone, language, country)` - Update localization
- `UpdateSocialLinks(profileID, website, linkedin, twitter, github)` - Update social links
- `UpdateVisibility(profileID, isPublic)` - Change visibility
- `VerifyProfile(profileID)` - Add verified badge
- `BanProfile(profileID)` - Moderation action

**Domain Events**:
- `profile.created` - New profile created
- `profile.updated` - Profile modified
- `profile.verified` - Profile verified
- `profile.banned` - Profile banned

## API Endpoints

### User Management

```http
POST   /api/v1/identity/users/register       # Register new user
POST   /api/v1/identity/users/login          # Login with credentials
POST   /api/v1/identity/auth/refresh         # Refresh access token
POST   /api/v1/identity/auth/revoke          # Logout (revoke token)
GET    /api/v1/identity/users/me             # Get current user 
PUT    /api/v1/identity/users/me/password    # Change password 
GET    /api/v1/identity/users                # List all users 
GET    /api/v1/identity/users/:id            # Get user by ID 
PUT    /api/v1/identity/users/:id/suspend    # Suspend user  (admin)
PUT    /api/v1/identity/users/:id/ban        # Ban user  (admin)
```

### Contact Management

```http
POST   /api/v1/identity/contacts             # Create contact 
GET    /api/v1/identity/contacts             # List user contacts 
GET    /api/v1/identity/contacts/:id         # Get contact 
PUT    /api/v1/identity/contacts/:id         # Update contact 
DELETE /api/v1/identity/contacts/:id         # Delete contact 
PUT    /api/v1/identity/contacts/:id/verify  # Verify contact 
PUT    /api/v1/identity/contacts/:id/primary # Set as primary 
```

### Profile Management

```http
POST   /api/v1/identity/profiles             # Create profile 
GET    /api/v1/identity/profiles/:id         # Get profile 
PUT    /api/v1/identity/profiles/:id         # Update profile 
DELETE /api/v1/identity/profiles/:id         # Delete profile 
PUT    /api/v1/identity/profiles/:id/visibility # Update visibility 
GET    /api/v1/identity/public/profiles      # List public profiles
```

 = Requires JWT authentication

## Value Objects

### Email

```go
type Email struct {
    value string
}

func NewEmail(email string) (Email, error) {
    trimmed := strings.TrimSpace(strings.ToLower(email))
    if !emailRegex.MatchString(trimmed) {
        return Email{}, fmt.Errorf("invalid email format")
    }
    return Email{value: trimmed}, nil
}
```

### Phone

```go
type Phone struct {
    countryCode string
    number      string
}

func NewPhone(countryCode, number string) (Phone, error) {
    // Validation with google/libphonenumber
    if !isValidPhone(countryCode, number) {
        return Phone{}, fmt.Errorf("invalid phone number")
    }
    return Phone{countryCode: countryCode, number: number}, nil
}
```

### Address

```go
type Address struct {
    street     string
    city       string
    state      string
    postalCode string
    country    string
}

func NewAddress(street, city, state, postalCode, country string) (Address, error) {
    if country == "" {
        return Address{}, fmt.Errorf("country is required")
    }
    return Address{
        street:     street,
        city:       city,
        state:      state,
        postalCode: postalCode,
        country:    country,
    }, nil
}
```

## Database Schema

### identity_users

```sql
CREATE TABLE identity_users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    failed_login_count INTEGER NOT NULL DEFAULT 0,
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_identity_users_email ON identity_users(email);
CREATE INDEX idx_identity_users_status ON identity_users(status);
```

### identity_contacts

```sql
CREATE TABLE identity_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES identity_users(id),
    type VARCHAR(50) NOT NULL,
    label VARCHAR(100) NOT NULL,
    email VARCHAR(255) NULL,
    phone_country_code VARCHAR(10) NULL,
    phone_number VARCHAR(50) NULL,
    address_street VARCHAR(255) NULL,
    address_city VARCHAR(100) NULL,
    address_state VARCHAR(100) NULL,
    address_postal_code VARCHAR(20) NULL,
    address_country VARCHAR(100) NULL,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_identity_contacts_user_id ON identity_contacts(user_id);
CREATE INDEX idx_identity_contacts_type ON identity_contacts(type);
CREATE UNIQUE INDEX idx_identity_contacts_primary ON identity_contacts(user_id, type) 
    WHERE is_primary = true AND deleted_at IS NULL;
```

### identity_profiles

```sql
CREATE TABLE identity_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID UNIQUE NOT NULL REFERENCES identity_users(id),
    display_name VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NULL,
    last_name VARCHAR(100) NULL,
    gender VARCHAR(20) NULL,
    date_of_birth DATE NULL,
    bio TEXT NULL,
    avatar_url VARCHAR(500) NULL,
    timezone VARCHAR(50) NULL,
    language VARCHAR(10) NULL,
    country VARCHAR(10) NULL,
    website VARCHAR(500) NULL,
    linkedin VARCHAR(500) NULL,
    twitter VARCHAR(500) NULL,
    github VARCHAR(500) NULL,
    is_public BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_banned BOOLEAN NOT NULL DEFAULT false,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_identity_profiles_user_id ON identity_profiles(user_id);
CREATE INDEX idx_identity_profiles_display_name ON identity_profiles(display_name);
```

## Testing

### Unit Tests (In-Place)

```go
// internal/contexts/identity/contact/entity_test.go
func TestContact_NewEmailContact(t *testing.T) {
    userID := uuidv7.New()
    contact, err := NewEmailContact(userID, "test@example.com", "Work")
    
    assert.NoError(t, err)
    assert.Equal(t, ContactTypeEmail, contact.Type)
    assert.NotNil(t, contact.Email)
}
```

### Smoke Tests (Mock-Based)

```go
// test/smoke/contexts/identity/contact/handler_test.go
func TestHandler_CreateEmailContact(t *testing.T) {
    mockUC := new(MockContactUseCase)
    handler := NewContactHandler(mockUC)
    
    mockUC.On("CreateEmailContact", mock.Anything, mock.Anything, "test@example.com", "Work", false).
        Return(&contact.Contact{}, nil)
    
    // Test HTTP handler with mock
}
```

### Integration Tests (Real DB)

```go
// test/integration/contexts/identity/contact/repository_test.go
func TestContactRepository_Create(t *testing.T) {
    db := integration.SetupTestDB(t)
    repo := postgres.NewContactRepository(db)
    
    contact, _ := contact.NewEmailContact(userID, "test@example.com", "Work")
    err := repo.Create(ctx, contact)
    
    assert.NoError(t, err)
}
```

**Test Coverage**:
- User: 32 unit + 11 smoke + 14 integration = **57 tests**
- Contact: 35 unit + 7 smoke + 9 integration = **51 tests**
- Profile: 35 unit + 8 smoke + 17 integration = **60 tests**

## Integration Points

### Event Publishing

```go
// After user registration
event := bus.NewBaseEvent("user.registered", user.ID)
eventBus.Publish(ctx, bus.TopicUserRegistered, event)

// After contact verification
event := bus.NewBaseEvent("contact.verified", contact.ID)
eventBus.Publish(ctx, bus.TopicContactVerified, event)
```

### Event Subscriptions

```go
// Customer context subscribes to user registration
eventBus.Subscribe(bus.TopicUserRegistered, func(ctx context.Context, e bus.Event) error {
    return customerUC.CreateFromUser(ctx, e.AggregateID())
})
```

## Configuration

```yaml
# config/app.postgres-dev.yaml or config/app.sqlite-dev.yaml
jwt:
  secret: "your-secret-key-at-least-32-characters"
  access_token_duration: 15m
  refresh_token_duration: 168h  # 7 days
  issuer: "promenade-platform"

database:
  redis:
    databases:
      revocation: 0  # JWT token revocation
```

## Next Steps

- [Shared Context](/contexts/shared) - Reference data (Country, Currency, Language)
- [Customer Management](/contexts/customer) - Customer, Company, Deal aggregates
- [JWT Package](/packages/jwt) - Token authentication details
- [Testing Guide](/guide/testing) - Write tests for Identity context
