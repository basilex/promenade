# Profiles IModule

User profile and contact information management.

---

## Overview

The **profiles** module extends core user authentication with rich profile information and multiple contact methods. It demonstrates how modules can build upon core entities (auth.users) while maintaining independence.

**Key Features**:

- Rich user profiles with bio, avatar, location
- Multiple contact methods per user (phone, email, social, website)
- Location with country/timezone references
- Privacy settings for profile visibility
- Social media integration
- Profile completeness tracking

**Status**: Enabled by default

**Namespace**: `profiles` (for migrations)

---

## Entities

### UserProfile

Extended user information beyond authentication.

**Fields**:

- `id` (UUID v7) - Primary key
- `user_id` (UUID) - User reference (FK to auth.users) - UNIQUE
- `bio` (TEXT) - User biography/description
- `avatar_url` (VARCHAR 500) - Profile picture URL
- `location` (VARCHAR 255) - Free-text location
- `country_code` (CHAR 2) - ISO country code (FK to countries)
- `timezone_id` (INTEGER) - Timezone reference (FK to timezones)
- `language_code` (CHAR 2) - Preferred language (FK to languages)
- `date_of_birth` (DATE) - Birth date (nullable)
- `is_public` (BOOLEAN) - Profile visibility flag
- `created_at` (TIMESTAMP) - Creation timestamp
- `updated_at` (TIMESTAMP) - Last update timestamp

**Validations**:

- Bio: Max length configurable (default: 500 chars)
- Avatar URL: Must be valid URL format
- Country code: Must exist in `countries` table
- Timezone: Must exist in `timezones` table
- Language: Must exist in `languages` table
- Date of birth: Must be in the past

**Business Rules**:

- One profile per user (1:1 relationship)
- Default `is_public = true` (can be changed)
- Location/country/timezone/language are optional
- Avatar URL stored only; no file upload handling (use external CDN)

**File**: [domain/entity/user_profile.go](domain/entity/user_profile.go)

---

### UserContact

Multiple contact methods for a user.

**Fields**:

- `id` (UUID v7) - Primary key
- `user_id` (UUID) - User reference (FK to auth.users)
- `contact_type` (VARCHAR 50) - Contact type enum
- `contact_value` (VARCHAR 255) - Contact value (email, phone, URL)
- `is_primary` (BOOLEAN) - Primary contact flag
- `is_verified` (BOOLEAN) - Verification status
- `created_at` (TIMESTAMP) - Creation timestamp
- `updated_at` (TIMESTAMP) - Last update timestamp

**Contact Types**:

- `email` - Email address
- `phone` - Phone number
- `github` - GitHub profile URL
- `linkedin` - LinkedIn profile URL
- `twitter` - Twitter/X handle
- `website` - Personal website URL
- `other` - Other contact method

**Validations**:

- Contact type: Must be one of allowed types
- Contact value: Required, validated per type (email format, URL format, etc.)
- Only one primary contact per type per user

**Business Rules**:

- Multiple contacts per user allowed
- One primary contact per type (e.g., one primary email)
- Verification status tracked separately (future: verification flow)
- No soft delete (hard delete only)

**File**: [domain/entity/user_contact.go](domain/entity/user_contact.go)

---

## Database Schema

### Tables

**`user_contacts`** (Namespace: `profiles`, Migration: 000001)

```sql
CREATE TABLE user_contacts (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    contact_type VARCHAR(50) NOT NULL,
    contact_value VARCHAR(255) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_contacts_user_id ON user_contacts(user_id);
CREATE INDEX idx_user_contacts_type ON user_contacts(contact_type);
CREATE UNIQUE INDEX idx_user_contacts_primary
    ON user_contacts(user_id, contact_type)
    WHERE is_primary = true;
```

**`user_profiles`** (Namespace: `profiles`, Migration: 000002)

```sql
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL UNIQUE REFERENCES auth.users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url VARCHAR(500),
    location VARCHAR(255),
    country_code CHAR(2) REFERENCES countries(code),
    timezone_id INTEGER REFERENCES timezones(id),
    language_code CHAR(2) REFERENCES languages(code),
    date_of_birth DATE,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_date_of_birth_past CHECK (date_of_birth < CURRENT_DATE)
);

CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX idx_user_profiles_country ON user_profiles(country_code);
CREATE INDEX idx_user_profiles_timezone ON user_profiles(timezone_id);
```

---

## Migrations

**Namespace**: `profiles`

| Version | Name                   | Description                                                         |
| ------- | ---------------------- | ------------------------------------------------------------------- |
| 000001  | `create_user_contacts` | Creates user_contacts table with contact type enum and verification |
| 000002  | `create_user_profiles` | Creates user_profiles table with references to core reference data  |

**Dependencies**:

- Requires core migrations (auth.users, countries, timezones, languages)

**Run migrations**:

```bash
make migrate-module MODULE=profiles
```

**Rollback**:

```bash
make migrate-rollback MODULE=profiles STEPS=1
```

---

## Configuration

Configuration in `config/modules.yaml`:

```yaml
modules:
  enabled:
    - profiles

  config:
    profiles:
      version: "1.0.0"
      settings:
        max_bio_length: 500 # Max characters in bio
        allow_custom_avatar: true # Allow custom avatar URLs
        require_location: false # Require location in profile
        default_public: true # Default profile visibility
        verify_contacts: true # Enable contact verification
```

**Environment Variables**:

- None (uses main app config for DB, logger, event bus)

---

## API Endpoints

### Profiles

#### Create/Update Profile

```http
PUT /api/v1/users/{user_id}/profile
Authorization: Bearer {token}
Content-Type: application/json

{
  "bio": "Software engineer passionate about Go and clean architecture.",
  "avatar_url": "https://cdn.example.com/avatars/user123.jpg",
  "location": "San Francisco, CA",
  "country_code": "US",
  "timezone_id": 1,
  "language_code": "en",
  "date_of_birth": "1990-01-01",
  "is_public": true
}
```

**Response**:

```json
{
  "status": "success",
  "data": {
    "id": "01931234-5678-7abc-def0-123456789abc",
    "user_id": "01931234-0000-7000-8000-000000000001",
    "bio": "Software engineer...",
    "avatar_url": "https://cdn.example.com/avatars/user123.jpg",
    "location": "San Francisco, CA",
    "country": {
      "code": "US",
      "name": "United States"
    },
    "timezone": {
      "id": 1,
      "name": "America/Los_Angeles"
    },
    "is_public": true,
    "created_at": "2025-12-22T10:00:00Z",
    "updated_at": "2025-12-22T10:00:00Z"
  }
}
```

#### Get Profile

```http
GET /api/v1/users/{user_id}/profile
```

**Notes**:

- Returns 404 if profile doesn't exist
- Respects `is_public` flag (only owner or admin can see private profiles)

#### Delete Profile

```http
DELETE /api/v1/users/{user_id}/profile
Authorization: Bearer {token}
```

---

### Contacts

#### Add Contact

```http
POST /api/v1/users/{user_id}/contacts
Authorization: Bearer {token}
Content-Type: application/json

{
  "contact_type": "email",
  "contact_value": "user@example.com",
  "is_primary": true,
  "is_verified": false
}
```

#### List Contacts

```http
GET /api/v1/users/{user_id}/contacts
Authorization: Bearer {token}
```

**Response**:

```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "contact_type": "email",
      "contact_value": "user@example.com",
      "is_primary": true,
      "is_verified": true,
      "created_at": "2025-12-22T10:00:00Z"
    },
    {
      "id": "...",
      "contact_type": "github",
      "contact_value": "https://github.com/username",
      "is_primary": false,
      "is_verified": false,
      "created_at": "2025-12-22T10:05:00Z"
    }
  ]
}
```

#### Update Contact

```http
PUT /api/v1/contacts/{id}
Authorization: Bearer {token}
Content-Type: application/json

{
  "contact_value": "newemail@example.com",
  "is_primary": true
}
```

#### Delete Contact

```http
DELETE /api/v1/contacts/{id}
Authorization: Bearer {token}
```

#### Verify Contact (Future)

```http
POST /api/v1/contacts/{id}/verify
Authorization: Bearer {token}
Content-Type: application/json

{
  "verification_code": "ABC123"
}
```

---

## Events

### Published Events

#### `profile.created`

```go
type ProfileCreatedEvent struct {
    ProfileID  string
    UserID     string
}
```

#### `profile.updated`

```go
type ProfileUpdatedEvent struct {
    ProfileID  string
    UserID     string
    Fields     []string  // Changed fields
}
```

#### `profile.deleted`

```go
type ProfileDeletedEvent struct {
    ProfileID  string
    UserID     string
}
```

#### `contact.added`

```go
type ContactAddedEvent struct {
    ContactID   string
    UserID      string
    ContactType string
}
```

#### `contact.verified`

```go
type ContactVerifiedEvent struct {
    ContactID   string
    UserID      string
    ContactType string
}
```

### Subscribed Events

- `user.deleted` - Cascades profile/contacts deletion via DB foreign keys (ON DELETE CASCADE)
- `user.created` - Could auto-create default profile (future feature)

---

## Purge Policy

**No automatic purge** - Profiles are not soft-deleted.

When a user is deleted from `auth.users`, profiles and contacts are **hard deleted** via CASCADE constraint.

---

## Testing

### Unit Tests

```bash
# Run module unit tests
go test ./internal/modules/profiles/usecase/... -v
```

### Integration Tests

```bash
# Run module integration tests
make test-integration
```

**Test Files**:

- `usecase/profile_usecase_test.go` - Business logic tests
- `usecase/contact_usecase_test.go` - Contact management tests
- `repository/postgres/profile_repository_test.go` - Database tests
- `test/smoke/profiles_smoke_test.go` - End-to-end smoke tests (if exists)

---

## File Structure

```
internal/modules/profiles/
 module.go                          # IModule registration & lifecycle

 domain/
    entity/
        user_profile.go            # UserProfile entity with validation
        user_contact.go            # UserContact entity

 repository/
    profile_repository.go          # Repository interface
    contact_repository.go
    postgres/
        profile_repository.go      # PostgreSQL implementation
        contact_repository.go

 usecase/
    profile_usecase.go             # Profile business logic
    profile_usecase_test.go
    contact_usecase.go             # Contact business logic
    contact_usecase_test.go

 adapter/
     handler/
         profile_handler.go         # HTTP handlers
         contact_handler.go
         dto/
            profile_dto.go         # Request/Response DTOs
            contact_dto.go
         router/
             profiles_router.go     # Route registration
```

---

## Usage Examples

### Creating a Profile

```go
import (
    "github.com/basilex/promenade/internal/modules/profiles/usecase"
    "github.com/basilex/promenade/internal/modules/profiles/domain/entity"
)

// In Initialize()
profileUseCase := usecase.NewProfileUseCase(profileRepo, eventBus, logger)

// Create profile
profile, err := profileUseCase.CreateProfile(ctx, entity.UserProfile{
    UserID:      userID,
    Bio:         "Software engineer",
    AvatarURL:   "https://cdn.example.com/avatar.jpg",
    Location:    "San Francisco",
    CountryCode: "US",
    IsPublic:    true,
})
```

### Adding Multiple Contacts

```go
contacts := []entity.UserContact{
    {
        UserID:       userID,
        ContactType:  "email",
        ContactValue: "user@example.com",
        IsPrimary:    true,
        IsVerified:   false,
    },
    {
        UserID:       userID,
        ContactType:  "github",
        ContactValue: "https://github.com/username",
        IsPrimary:    false,
        IsVerified:   false,
    },
}

for _, contact := range contacts {
    _, err := contactUseCase.AddContact(ctx, contact)
    if err != nil {
        return err
    }
}
```

### Profile Completeness Check

```go
func (u *ProfileUseCase) CalculateCompleteness(profile *entity.UserProfile) int {
    completeness := 0
    if profile.Bio != "" { completeness += 20 }
    if profile.AvatarURL != "" { completeness += 20 }
    if profile.Location != "" { completeness += 15 }
    if profile.CountryCode != "" { completeness += 15 }
    if profile.TimezoneID != nil { completeness += 15 }
    if profile.LanguageCode != "" { completeness += 15 }
    return completeness  // 0-100%
}
```

---

## Permissions

**RBAC Permissions** (if enabled):

- `profiles:read` - Read profiles (auto-granted to all authenticated users)
- `profiles:update` - Update own profile
- `profiles:delete` - Delete own profile
- `profiles:update:any` - Update any user's profile (admin)
- `profiles:delete:any` - Delete any user's profile (admin)

**Default Roles**:

- **User**: `profiles:read`, `profiles:update`, `profiles:delete`
- **Admin**: All permissions + `profiles:update:any`, `profiles:delete:any`

---

## Future Enhancements

Potential features (not yet implemented):

- [ ] Contact verification flow (email/SMS codes)
- [ ] Profile badges/achievements
- [ ] Privacy controls per field (hide email, show location, etc.)
- [ ] Profile themes/customization
- [ ] Activity feed on profile
- [ ] Profile view analytics
- [ ] Profile QR code generation
- [ ] Export profile data (GDPR compliance)

---

##  Related Documentation

- **[../README.md](../README.md)** - IModule system overview
- **[../../README.md](../../README.md)** - Main project README
- **[../../docs/MODULE_DEVELOPMENT.md](../../docs/MODULE_DEVELOPMENT.md)** - IModule development guide
- **[../../docs/VALIDATION.md](../../docs/VALIDATION.md)** - Validation patterns
- **[../../migrations/README.md](../../migrations/README.md)** - Migration system

---

**IModule Status**: Production-ready | Tested | 2 migrations | RBAC-enabled
