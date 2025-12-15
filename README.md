# Promenade

A RESTful API service for resources management built with Go.

## Features

- **RESTful API** - Full CRUD operations for resource management
- **In-memory Storage** - Fast, thread-safe in-memory data storage
- **Structured Logging** - Request/response logging with timestamps
- **Health Check** - Built-in health check endpoint
- **Validation** - Input validation and error handling
- **Concurrent Safe** - Thread-safe operations using mutex locks

## API Endpoints

### Health Check
- **GET** `/health` - Check service health

### Resources
- **GET** `/api/resources` - List all resources
- **GET** `/api/resources/{id}` - Get a specific resource
- **POST** `/api/resources` - Create a new resource
- **PUT** `/api/resources/{id}` - Update a resource
- **DELETE** `/api/resources/{id}` - Delete a resource

## Resource Model

```json
{
  "id": "uuid-string",
  "name": "Resource Name",
  "description": "Resource Description",
  "type": "resource-type",
  "status": "active",
  "created_at": "2025-12-15T09:47:37.908315034Z",
  "updated_at": "2025-12-15T09:47:37.908315124Z"
}
```

## Quick Start

### Prerequisites
- Go 1.24 or higher

### Installation

1. Clone the repository:
```bash
git clone https://github.com/basilex/promenade.git
cd promenade
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o bin/promenade ./cmd/server
```

4. Run the server:
```bash
./bin/promenade
```

The server will start on port 8080 by default. You can change the port by setting the `PORT` environment variable:
```bash
PORT=3000 ./bin/promenade
```

## Usage Examples

### Create a Resource
```bash
curl -X POST http://localhost:8080/api/resources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Database Server",
    "description": "PostgreSQL production database",
    "type": "database",
    "status": "active"
  }'
```

### List All Resources
```bash
curl http://localhost:8080/api/resources
```

### Get a Specific Resource
```bash
curl http://localhost:8080/api/resources/{id}
```

### Update a Resource
```bash
curl -X PUT http://localhost:8080/api/resources/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "status": "maintenance"
  }'
```

### Delete a Resource
```bash
curl -X DELETE http://localhost:8080/api/resources/{id}
```

## Development

### Run Tests
```bash
go test ./... -v
```

### Run Tests with Coverage
```bash
go test ./... -cover
```

### Project Structure
```
promenade/
├── cmd/
│   └── server/          # Main application entry point
│       └── main.go
├── internal/
│   ├── handlers/        # HTTP request handlers
│   │   ├── resource.go
│   │   └── resource_test.go
│   ├── models/          # Data models
│   │   └── resource.go
│   └── storage/         # Data storage layer
│       ├── memory.go
│       └── memory_test.go
├── go.mod
├── go.sum
└── README.md
```

## License

See [LICENSE](LICENSE) file for details.