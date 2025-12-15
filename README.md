# Promenade - Clean Architecture REST API

Production-ready REST API with Clean Architecture, PostgreSQL, JWT authentication, and API versioning.

### 🚀 Quick Start

#### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make

#### Installation

```bash
## Clone the repository
git clone https://github.com/basilex/promenade.git
cd promenade

## Initialize project (installs tools, starts Docker, runs migrations)
./scripts/init.sh

## Start development server
make dev
```

## Access

- API: http://localhost:8081
- Swagger V1: http://localhost:8081/swagger/v1/index.html
- Swagger V2: http://localhost:8081/swagger/v2/index.html

## Documentation

### Available Commands

```bash
make help              # Show all available commands
make dev               # Run in development mode
make build             # Build binary
make test              # Run all tests
make test-coverage     # Generate coverage report
make docker-up         # Start all services
make docker-down       # Stop all services
make migrate-up        # Run migrations
make migrate-down      # Rollback migrations
make swagger-all       # Generate Swagger docs
```

### Project Structure

```Code
promenade/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── domain/                 # Business logic (entities, interfaces)
│   ├── usecase/                # Application business rules
│   ├── adapter/                # Adapters (HTTP, Repository)
│   └── infrastructure/         # Infrastructure (DB, Config)
├── pkg/                        # Reusable packages
├── test/                       # Tests
├── migrations/                 # Database migrations
├── docker/                     # Docker configuration
└── docs/                       # API documentation (auto-generated)
```

## Architecture

This project follows Clean Architecture principles:

Domain Layer: Business entities and rules (independent)
Use Case Layer: Application-specific business rules
Adapter Layer: Interface adapters (HTTP handlers, repositories)
Infrastructure Layer: Frameworks and drivers (Database, Config)

### Key Features

- Clean Architecture with clear separation of concerns
- API Versioning (v1, v2) with backward compatibility
- JWT Authentication
- Rate Limiting
- PostgreSQL with sqlx (no ORM!)
- Transaction support
- Swagger documentation
- Comprehensive middleware (auth, logging, recovery, CORS)
- Unit, integration, and E2E tests
- Docker support
- Database migrations
- Graceful shutdown

### Configuration

#### Environment Files

- `.env` - Default configuration (committed)
- `.env.development` - Development settings (committed)
- `.env.local` - Your personal overrides (NOT committed)
- `.env.production` - Production secrets (NOT committed)

#### Setup

```bash
# Copy example file
cp .env.local.example .env.local

# Edit your local settings
vim .env.local

# Run
make env-dev
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Generate coverage report
make test-coverage
```

### Deployment

```bash
# Build Docker image
make docker-build

# Run in production
docker run -p 8081:8081 --env-file .env.production promenade: latest

```

### API Examples

#### Create User (v1)

```bash
curl -X POST http://localhost:8081/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"John Doe","password":"password123"}'
```

#### Create User (v2 - with structured profile)

```bash
curl -X POST http://localhost:8081/api/v2/users \
  -H "Content-Type:  application/json" \
  -d '{"email":"user@example.com","first_name":"John","last_name":"Doe","password":"password123"}'
```

#### Get User with Auth

```bash
curl -X GET http://localhost:8081/api/v2/users/{id} \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Contributing

- Fork the repository
- Create your feature branch (git checkout -b feature/amazing-feature)
- Commit your changes (git commit -m 'Add amazing feature')
- Push to the branch (git push origin feature/amazing-feature)
- Open a Pull Request

### License

MIT License - see LICENSE file for details
