# Development Credentials

Quick reference for default users and access credentials with RBAC roles.

## Default Users

### 1. System Administrator (Main Admin - Oracle Style)

```
Email:    system@promenade.com
Password: passw0rd
Role:     admin
Access:   *:* (full system access)
```

**Use for:**

- System initialization and bootstrap
- RBAC management (roles, permissions)
- User management (ban, suspend, assign roles)
- All administrative operations

### 2. Administrator

```
Email:    admin@promenade.com
Password: passw0rd
Role:     admin
Access:   users:*, posts:*, comments:*, profiles:*, roles:read|list|assign
```

**Use for:**

- User management (create, update, delete, ban)
- Content management (posts, comments)
- Role assignment to users
- Testing admin-level permissions

### 4. Moderator

```
Email:    moderator@promenade.com
Password: passw0rd
Role:     moderator
Access:   posts:read|update|delete|list, comments:*, profiles:read|list
```

**Use for:**

- Content moderation (posts, comments)
- Comment management (approve, delete)
- Testing moderation workflows
- Limited user visibility (read only)

### 5. Regular User

```
Email:    alexander.vasilenko@gmail.com
Password: 03041965
Role:     user
Access:   posts:create|read, comments:create|read, profiles:create|read
```

**Use for:**

- Testing regular user workflows
- Own content creation (posts, comments)
- Profile management
- Basic user operations

### Guest Access (Unauthenticated)

No login required:

```
Role:     guest (implicit)
Access:   posts:read, comments:read, profiles:read
```

**Use for:**

- Public content browsing
- Testing unauthenticated access
- Read-only operations

## Quick Login Examples

```bash
# Login as System Administrator (admin with full access)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq

# Login as Administrator (admin)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@promenade.com","password":"passw0rd"}' | jq

# Login as Moderator (moderator)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@promenade.com","password":"passw0rd"}' | jq

# Login as Regular User (user)
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alexander.vasilenko@gmail.com","password":"03041965"}' | jq

# Save token for reuse (system admin)
export TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq -r '.data.access_token')

# Use token in requests
curl -X GET http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

## Testing RBAC Permissions

```bash
# Test system admin access (should work - full access)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $SYSTEM_TOKEN" | jq

# Test admin access (should work - has users:*)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Test moderator access (should fail - no users:list permission)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $MODERATOR_TOKEN" | jq

# Test user access (should fail - no admin permissions)
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $USER_TOKEN" | jq
```

## Database Access

```bash
# Connect to development database
psql -h localhost -p 5432 -U system -d promenade_dev
# Password: passw0rd

# View all users with their roles
SELECT
    u.email,
    u.name,
    r.name as role,
    r.description
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
ORDER BY u.email;

# View role permissions
SELECT
    r.name as role,
    p.resource,
    p.action,
    p.description
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.resource, p.action;
```

## [!] Security Warning

**These credentials are for DEVELOPMENT ONLY!**

Before deploying to production:

1. Change all default passwords
2. Remove or disable default admin accounts
3. Use environment-specific credentials
4. Enable proper secret management
5. Set up proper authentication providers

## Need Help?

- [Authorization Guide](docs/AUTHORIZATION.md) - RBAC system documentation
- [Testing Guide](docs/TESTING_GUIDE.md) - How to test with authentication
- [API Documentation](README.md#api-examples) - Full API reference
