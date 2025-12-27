# Contact API - Quick Start

## ✅ Completed

### Contact Aggregate Implementation

- ✅ Entity (Contact, Email, Phone, Address value objects)
- ✅ UseCase (business logic)
- ✅ Repository (PostgreSQL adapter)
- ✅ HTTP Handler (8 REST endpoints)
- ✅ DTOs (request/response mapping)
- ✅ Router (Identity context integration)
- ✅ Tests (70 tests, 100% passing)

### API Endpoints

Base URL: `http://localhost:8081/api/v1/identity/contacts`

| Method | Path                    | Description                                   |
| ------ | ----------------------- | --------------------------------------------- |
| POST   | `/contacts`             | Create new contact (email, phone, or address) |
| GET    | `/contacts`             | List contacts (filter by user_id, type)       |
| GET    | `/contacts/:id`         | Get contact by ID                             |
| PUT    | `/contacts/:id`         | Update contact (label, visibility)            |
| DELETE | `/contacts/:id`         | Delete contact (soft delete)                  |
| PUT    | `/contacts/:id/verify`  | Verify contact (email/phone confirmation)     |
| PUT    | `/contacts/:id/primary` | Set contact as primary for user               |

## 🚀 Running the API

### 1. Start PostgreSQL (Docker)

```bash
# Start Docker containers
make docker-up

# Check containers running
make docker-ps
```

### 2. Run Migrations

```bash
# Migrations run automatically on app startup
# Or manually:
make migrate
```

### 3. Start API Server

```bash
# Development mode (auto-runs migrations)
make dev

# Or build and run
make build
./bin/promenade
```

Server starts on: **http://localhost:8081**

### 4. Test API Endpoints

Use the provided test script:

```bash
# Run all tests
./test_api.sh

# Or manual curl tests:

# Health check
curl http://localhost:8081/health

# API info
curl http://localhost:8081/api/v1

# Create email contact
curl -X POST http://localhost:8081/api/v1/identity/contacts?user_id=<uuid> \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "label": "Work",
    "email": "john@company.com"
  }'
```

## 📊 Test Coverage

Run tests:

```bash
# All unit tests (70 tests)
make test

# Coverage report
make test-coverage
```

Test Results:

- ✅ Entity tests: 13/13 PASS
- ✅ UseCase tests: 8/8 PASS
- ✅ Handler tests: 8/8 PASS (+ subtests)
- 📊 Coverage: Entity+UseCase 75.3%, Handler 38.3%

## 🎯 Next Steps

1. ✅ **Contact Aggregate Complete** - Foundation solid!
2. 🔄 **Phase 2**: User aggregate (registration, authentication)
3. 🔄 **Phase 3**: Customer Management context
4. 🔄 **Phase 4**: Order Management context

## 📝 Notes

- Contact creation requires `user_id` query parameter (JWT auth will replace this)
- All IDs use UUID v7 (time-ordered)
- Contacts support soft delete (`deleted_at`)
- Primary contact: only one per user per type
- Verification: email confirmation, phone OTP, etc.

## 🐛 Troubleshooting

**Database Connection Error**:

```bash
# Check PostgreSQL running
make docker-ps

# Restart containers
make docker-down
make docker-up
```

**Port 8081 Already in Use**:

```bash
# Find and kill process
lsof -ti:8081 | xargs kill -9
```

**Migration Issues**:

```bash
# Check migration status
make migrate-status

# Rollback last migration
make migrate-down
```

## 📚 Documentation

- [Architecture Overview](docs/ARCHITECTURE_OVERVIEW.md)
- [DDD Patterns](docs/ARCHITECTURE_QUICKREF.md)
- [Testing Guide](docs/TESTING_GUIDE.md)
- [UUID v7 Guide](docs/UUID_V7_GUIDE.md)
