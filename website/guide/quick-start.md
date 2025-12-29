# Quick Start

**Get Promenade running in 5 minutes** - from zero to first API call.

---

## Prerequisites

Before you begin, ensure you have:

- **Go 1.23+** installed
- **Docker & Docker Compose** (for PostgreSQL)
- **Make** (comes with macOS/Linux)
- **Git** (to clone the repository)

---

## 1. Clone & Start Infrastructure

```bash
# Clone repository
git clone https://github.com/basilex/promenade.git
cd promenade

# Start PostgreSQL (localhost:5432)
make docker-up

# Wait for PostgreSQL to be ready (3 seconds)
```

✅ PostgreSQL is now running on `localhost:5432`

---

## 2. Run Migrations

```bash
# Migrations run automatically on app startup
# Or manually:
make migrate
```

This creates all tables in correct order:
- **Core** (UUID v7 extensions, auth, RBAC)
- **Shared** (Countries, Currencies, Languages, Timezones)
- **Identity** (Users, Contacts, Profiles, Roles, Permissions)
- **Customer Management** (Customers)

---

## 3. Start API Server

```bash
# Full dev environment (Docker + migrations + API)
make dev

# Or build and run separately
make build
./bin/promenade
```

✅ Server starts on **http://localhost:8081**

---

## 4. Health Check

```bash
curl http://localhost:8081/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-12-29T12:00:00Z"
}
```

---

## 5. Your First API Calls

### Register a User

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "name": "John Doe",
    "password": "SecurePass123"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC1234567890DEFGHIJK",
    "email": "john@example.com",
    "name": "John Doe",
    "status": "active",
    "created_at": "2025-12-29T12:00:00Z"
  }
}
```

### Login

```bash
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
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
    "expires_at": "2025-12-29T12:15:00Z",
    "token_type": "Bearer",
    "user": {
      "id": "01JGABC1234567890DEFGHIJK",
      "email": "john@example.com",
      "name": "John Doe"
    }
  }
}
```

💡 **Copy the `access_token` - you'll need it for protected endpoints!**

### Create Profile (Protected)

```bash
# Replace YOUR_ACCESS_TOKEN with token from login response
curl -X POST http://localhost:8081/api/v1/identity/profiles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "display_name": "John D.",
    "first_name": "John",
    "last_name": "Doe",
    "bio": "Software developer and DDD enthusiast",
    "timezone": "Europe/Kyiv",
    "language": "uk",
    "country": "UA"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "01JGABC9876543210ZYXWVUTS",
    "user_id": "01JGABC1234567890DEFGHIJK",
    "display_name": "John D.",
    "first_name": "John",
    "last_name": "Doe",
    "bio": "Software developer and DDD enthusiast",
    "timezone": "Europe/Kyiv",
    "language": "uk",
    "country": "UA",
    "is_public": false
  }
}
```

### Add Email Contact

```bash
curl -X POST http://localhost:8081/api/v1/identity/contacts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "type": "email",
    "email": "john.work@company.com",
    "label": "Work",
    "is_primary": false
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "01JGABCXYZ1234567890ABCDE",
    "user_id": "01JGABC1234567890DEFGHIJK",
    "type": "email",
    "label": "Work",
    "email": "john.work@company.com",
    "is_primary": false,
    "is_verified": false,
    "is_public": false
  }
}
```

### Get Reference Data

```bash
# Get all countries
curl http://localhost:8081/api/v1/shared/countries

# Get all currencies
curl http://localhost:8081/api/v1/shared/currencies

# Get all languages
curl http://localhost:8081/api/v1/shared/languages

# Get all timezones
curl http://localhost:8081/api/v1/shared/timezones
```

---

## 6. Common Workflows

### Complete User Setup

```bash
# 1. Register
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/identity/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","name":"Alice","password":"Pass123!"}' | \
  jq -r '.data.access_token')

# 2. Create Profile
curl -X POST http://localhost:8081/api/v1/identity/profiles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"display_name":"Alice","timezone":"Europe/London","language":"en","country":"GB"}'

# 3. Add Work Email
curl -X POST http://localhost:8081/api/v1/identity/contacts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"email","email":"alice.work@company.com","label":"Work","is_primary":true}'

# 4. Add Phone
curl -X POST http://localhost:8081/api/v1/identity/contacts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type":"phone","phone":"+441234567890","label":"Mobile"}'

# 5. Get My Profile
curl http://localhost:8081/api/v1/identity/profiles/me \
  -H "Authorization: Bearer $TOKEN"
```

### Customer Management

```bash
# Create Customer
curl -X POST http://localhost:8081/api/v1/customer-mgmt/customers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "email": "contact@acme.com",
    "phone": "+441234567890",
    "segment": "enterprise",
    "status": "active"
  }'

# List Customers
curl http://localhost:8081/api/v1/customer-mgmt/customers \
  -H "Authorization: Bearer $TOKEN"

# Get Customer by ID
curl http://localhost:8081/api/v1/customer-mgmt/customers/01JGABC... \
  -H "Authorization: Bearer $TOKEN"
```

---

## 7. Development Workflow

### File Watching (Manual)

```bash
# Terminal 1: Watch and restart on changes
fswatch -o internal/ cmd/ pkg/ | xargs -n1 -I{} make run

# Terminal 2: Watch tests
fswatch -o internal/ | xargs -n1 -I{} make test-unit
```

### Database Reset

```bash
# Drop and recreate database (WARNING: all data lost!)
make db-reset

# Fresh database with migrations
make db-fresh

# Or full clean restart
make dev-fresh
```

### Run Tests

```bash
# All tests (240+ tests, ~40s)
make test

# By type
make test-unit              # Unit tests only (~5s)
make test-smoke             # Smoke tests (~0.4s)
make test-integration       # Integration tests (~6s)

# Specific context
go test ./internal/contexts/identity/... -v

# With coverage
make test-coverage
```

---

## 8. Configuration

### Environment Variables

```bash
# Development (default)
ENVIRONMENT=dev ./bin/promenade

# Production
ENVIRONMENT=production ./bin/promenade

# Custom database
DB_HOST=localhost DB_PORT=5432 DB_NAME=promenade_custom ./bin/promenade

# Custom JWT secret (production!)
JWT_SECRET=$(openssl rand -base64 32) ./bin/promenade
```

### Configuration Files

- `config/app.dev.yaml` - Development (Memory bus, localhost DB)
- `config/app.test.yaml` - Testing (isolated test DB)
- `config/app.prod.yaml` - Production (Redis bus, production DB)

**Example (`config/app.dev.yaml`)**:
```yaml
app:
  name: "Promenade Platform"
  environment: "development"
  version: "0.1.0"

server:
  host: "0.0.0.0"
  port: 8081

database:
  postgres:
    host: "localhost"
    port: 5432
    user: "system"
    password: "passw0rd"
    database: "promenade_dev"

bus:
  adapter: "memory"  # fast in-process event bus
```

---

## 9. What's Next?

### Learn the Architecture

- [Architecture Guide](/guide/architecture) - Domain-Driven Design principles
- [Testing Strategy](/guide/testing) - Three-tier testing approach
- [RBAC Guide](/guide/rbac) - Role-based access control

### Explore Contexts

- [Shared Context](/contexts/shared) - Reference data (Countries, Currencies, etc.)
- [Identity Context](/contexts/identity) - User management & authentication
- [Customer Context](/contexts/customer) - Customer management

### API Reference

- [API Reference](/guide/api-reference) - Complete endpoint documentation
- [Contributing](/guide/contributing) - Development workflow

---

## Common Issues

### PostgreSQL Connection Failed

```bash
# Check if PostgreSQL is running
make docker-ps

# Restart PostgreSQL
make docker-down && make docker-up
```

### Port 8081 Already in Use

```bash
# Find process using port
lsof -i :8081

# Kill process
kill -9 <PID>

# Or use different port
SERVER_PORT=8082 ./bin/promenade
```

### Migration Errors

```bash
# Check migration status
make migrate-status

# Reset database (WARNING: destructive!)
make db-fresh
```

### JWT Token Expired

```bash
# Use refresh token to get new access token
curl -X POST http://localhost:8081/api/v1/identity/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

---

## Make Commands Reference

```bash
# Development
make dev               # Full dev environment (Docker + migrations + API)
make dev-fresh         # Fresh start with clean database
make build             # Build binary
make run               # Build and run
make fmt               # Format code
make lint              # Run linters

# Testing
make test              # All tests (~40s)
make test-unit         # Unit tests (~5s)
make test-smoke        # Smoke tests (~0.4s)
make test-integration  # Integration tests (~6s)
make test-coverage     # HTML coverage report

# Database
make docker-up         # Start PostgreSQL
make docker-down       # Stop PostgreSQL
make docker-ps         # Show running containers
make migrate           # Run all migrations
make db-reset          # Drop and recreate database
make db-fresh          # Fresh database with migrations

# Documentation
make swagger-all       # Generate API documentation
```

---

## Success! 🎉

You now have Promenade running locally with:

- ✅ PostgreSQL database with all migrations
- ✅ API server on http://localhost:8081
- ✅ User registration and authentication
- ✅ Profile and contact management
- ✅ Reference data (countries, currencies, etc.)

**Next Steps:**
1. Explore the [API Reference](/guide/api-reference) for all available endpoints
2. Read the [Architecture Guide](/guide/architecture) to understand DDD principles
3. Check out [Testing Strategy](/guide/testing) to write your own tests
4. See [RBAC Guide](/guide/rbac) for authorization patterns

Happy coding! 🚀
