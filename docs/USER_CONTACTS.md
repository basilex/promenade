# User Contacts Feature

## Overview

The User Contacts feature allows users to manage multiple contact methods (email, phone, various messaging apps) with flexible privacy settings and availability schedules.

## Features

- **Multiple Contact Types**: email, phone, telegram, whatsapp, viber, signal, skype, discord, linkedin, other
- **Privacy Control**: Public/private visibility for each contact
- **Verification Status**: Track which contacts have been verified
- **Primary Contact**: Set one contact as primary per type
- **Active/Inactive**: Toggle contact availability
- **Availability Schedule**: Set time ranges and days when contacts are available
- **Timezone Support**: Store timezone for availability calculations

## Database Schema

**Table:** `user_contacts`

| Column         | Type         | Description                                       |
| -------------- | ------------ | ------------------------------------------------- |
| id             | UUID         | Primary key (UUID v7)                             |
| user_id        | UUID         | Foreign key to users table                        |
| contact_type   | VARCHAR(50)  | Type of contact                                   |
| contact_value  | VARCHAR(255) | Contact data (email, phone, username)             |
| label          | VARCHAR(100) | Optional custom label                             |
| is_verified    | BOOLEAN      | Whether contact is verified                       |
| is_primary     | BOOLEAN      | Whether this is the primary contact for this type |
| is_active      | BOOLEAN      | Whether contact is currently active               |
| is_public      | BOOLEAN      | Whether contact is publicly visible               |
| available_from | TIME         | Start time of availability                        |
| available_to   | TIME         | End time of availability                          |
| available_days | VARCHAR[]    | Days when contact is available                    |
| timezone       | VARCHAR(50)  | Timezone for availability                         |
| notes          | TEXT         | Additional notes                                  |
| created_at     | TIMESTAMP    | Creation timestamp                                |
| updated_at     | TIMESTAMP    | Last update timestamp                             |

**Constraints:**

- `UNIQUE(user_id, contact_type, contact_value)` - Prevent duplicate contacts
- Only one primary contact per type per user

## API Endpoints

All endpoints require authentication (JWT Bearer token).

### Create Contact

```http
POST /api/v1/users/contacts
```

**Request:**

```json
{
  "contact_type": "email",
  "contact_value": "user@example.com",
  "label": "Work Email",
  "is_public": true,
  "available_from": "09:00:00",
  "available_to": "18:00:00",
  "available_days": ["monday", "tuesday", "wednesday", "thursday", "friday"],
  "timezone": "UTC",
  "notes": "Best time to contact"
}
```

**Response:** `201 Created`

### Get User's Contacts

```http
GET /api/v1/users/contacts?include_inactive=false
```

**Response:** `200 OK` - Array of contacts

### Get Contact by ID

```http
GET /api/v1/users/contacts/{id}
```

**Response:** `200 OK` - Single contact

### Get Contacts by Type

```http
GET /api/v1/users/contacts/type/{type}
```

**Response:** `200 OK` - Array of contacts filtered by type

### Get Primary Contact

```http
GET /api/v1/users/contacts/primary/{type}
```

**Response:** `200 OK` - Primary contact for specified type

### Update Contact

```http
PUT /api/v1/users/contacts/{id}
```

**Request:** Same as Create Contact

**Response:** `200 OK` - Updated contact

### Delete Contact

```http
DELETE /api/v1/users/contacts/{id}
```

**Response:** `204 No Content`

### Set Primary Contact

```http
POST /api/v1/users/contacts/{id}/primary
```

**Response:** `200 OK`

### Toggle Contact Active Status

```http
POST /api/v1/users/contacts/{id}/toggle
```

**Response:** `200 OK`

## Business Rules

1. **Privacy**: Users can only see their own private contacts. Public contacts are visible to anyone.
2. **Ownership**: Users can only modify/delete their own contacts.
3. **Primary Contacts**: Only one contact can be primary per type. Setting a new primary automatically unsets the old one.
4. **Verification**: New contacts are unverified by default.
5. **Availability**: Time validation ensures `available_from` is before `available_to`.
6. **Uniqueness**: Same contact value cannot be added twice for the same type.

## Use Cases

### Authorization Layer

- `CreateContact` - Create new contact for authenticated user
- `GetContact` - Get contact with privacy check
- `GetUserContacts` - Get all contacts (own contacts or public contacts of others)
- `UpdateContact` - Update own contact
- `DeleteContact` - Delete own contact
- `SetPrimaryContact` - Set contact as primary
- `VerifyContact` - Mark contact as verified (admin action)
- `ToggleContactActive` - Toggle contact active status

## Example Scenarios

### Adding Work Email

```bash
curl -X POST http://localhost:8081/api/v1/users/contacts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_type": "email",
    "contact_value": "john.doe@company.com",
    "label": "Work",
    "is_public": false,
    "available_from": "09:00:00",
    "available_to": "17:00:00",
    "available_days": ["monday", "tuesday", "wednesday", "thursday", "friday"],
    "timezone": "America/New_York"
  }'
```

### Getting Public Contacts

Public contacts are available without authentication through the public contacts endpoint.

### Setting Primary Phone

```bash
curl -X POST http://localhost:8081/api/v1/users/contacts/{contact-id}/primary \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Testing

Run tests:

```bash
make test                    # All tests
make test-unit               # Unit tests only
make test-integration        # Integration tests only
```

## Future Enhancements

- [ ] Contact verification workflow (email verification links, SMS codes)
- [ ] Contact groups/tags
- [ ] Contact sharing permissions
- [ ] Contact history/changelog
- [ ] Bulk operations
- [ ] Contact import/export
- [ ] Contact search functionality
- [ ] Rate limiting for contact verification attempts
