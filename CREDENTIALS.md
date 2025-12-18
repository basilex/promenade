# 🔑 Development Credentials

Quick reference for default users and access credentials.

## Default Users

### Superadmin (Full System Access)

```
Email:    system@promenade.com
Password: passw0rd
Role:     superadmin
Access:   *:* (all permissions)
```

**Use for:**

- System administration
- RBAC management (roles, permissions)
- User management (ban, suspend, assign roles)
- Testing admin-only endpoints

### Admin (Administrative Access)

```
Email:    alexander.vasilenko@gmail.com
Password: 03041965
Role:     admin
Access:   Most resources (users, posts, comments, profiles)
```

**Use for:**

- Content moderation
- User management (except role assignment)
- Testing moderator workflows

### Regular User (Own Content)

Any registered user automatically gets the `user` role:

```
Role:     user
Access:   posts:*, comments:*, profiles:* (own content)
```

## Quick Login

```bash
# Login as Superadmin
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq

# Login as Admin
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alexander.vasilenko@gmail.com","password":"03041965"}' | jq

# Save token for reuse
export TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"system@promenade.com","password":"passw0rd"}' | jq -r '.data.access_token')

# Use token
curl -X GET http://localhost:8081/api/v1/admin/users \
  -H "Authorization: Bearer $TOKEN"
```

## Database Access

```bash
# Connect to development database
psql -h localhost -p 5432 -U system -d promenade_dev
# Password: passw0rd

# View users and roles
SELECT u.email, u.name, r.name as role
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id;
```

## ⚠️ Security Warning

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
