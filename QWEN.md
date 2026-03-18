# FiberBackend - Modern Go REST API

## Project Overview

FiberBackend is a modern REST API for a blog/content management system built with Go and the Fiber web framework. The application connects to PostgreSQL for data persistence and includes features like user management, content management, social features, real-time chat, and more.

### Key Technologies
- **Go**: 1.25+
- **Fiber Framework**: v3.0.0
- **GORM**: v1.31.1 for database ORM
- **PostgreSQL**: 14+ via pgx driver
- **JWT Authentication**: github.com/golang-jwt/jwt/v5
- **Dependency Injection**: Manual DI with cleanup management
- **Validation**: go-playground/validator/v10

### Architecture
The project follows a clean architecture pattern with separation of concerns:
- **cmd/**: Application entry point (`main.go`)
- **config/**: Configuration management with environment variable loading
- **internal/**: Private application code
  - **handler/**: HTTP request handlers
  - **service/**: Business logic layer
  - **repository/**: Data access layer
  - **model/**: Database models and DTOs
  - **middleware/**: HTTP middleware (authentication, logging, security)
  - **routes/**: Route definitions and setup
  - **di/**: Dependency injection container setup
- **test/**: Test files
- **pkg/**: Shared utilities
  - **database/**: Database connection and wrapper
  - **response/**: Standardized response helpers
  - **utils/**: General utility functions
  - **validator/**: Custom validator implementation

## Building and Running

### Prerequisites
- Go 1.25+
- PostgreSQL 14+
- Docker (optional, for containerized deployment)

### Local Development Setup

1. **Clone and setup:**
   ```bash
   git clone <your-repo-url>
   cd fiberbackend
   cp .env.example .env
   ```

2. **Configure environment:**
   Edit `.env` file with your configuration:
   ```env
   # Server
   PORT=8080

   # Database
   DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
   MAX_OPEN_CONNS=30
   MAX_IDLE_CONNS=2
   CONN_MAX_LIFETIME=30m

   # Authentication
   JWT_SECRET=your-secret-key

   # Rate Limiting
   RATE_LIMITER_MAX=0
   RATE_LIMITER_TTL=60

   # Debug
   DEBUG=false
   ```

3. **Run:**
   ```bash
   # Install dependencies
   go mod download

   # Run migrations (if using migration tool)
   # psql -d your_database_name -f migrations/*.sql

   # Start the server
   go run cmd/main.go
   ```

   Server starts at `http://localhost:8080`

### Development Commands

```bash
# Build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/main cmd/main.go

# Run with hot reload (using Air)
air

# Run directly
go run cmd/main.go

# Run tests
go test ./...
go test -v ./...           # Verbose output
go test -cover ./...       # With coverage
go test -short ./...       # Skip long-running tests

# Code quality
golangci-lint run          # Run all linters
gofmt -l .                 # Check formatting
go vet ./...               # Vet all packages
```

### Using Air for Hot Reload

The project includes `.air.toml` configuration for the Air live reload tool:
```bash
# Install air if not already installed
go install github.com/air-verse/air@latest

# Run with hot reload
air
```

### Docker Deployment

```bash
# Build and run
docker build -t fiberbackend .
docker run -p 8080:8080 fiberbackend
```

### Fly.io Deployment

The project includes `fly.toml` for easy deployment to Fly.io:
```bash
# Deploy to Fly.io
flyctl deploy
```

## Features

- **User Management**: Complete user system with authentication, profiles, and following
- **Content Management**: Posts with rich text, images, tags, and versioning
- **Social Features**: User follows, post likes, comments, and bookmarks
- **Real-time Chat**: Conversational AI with message history and token tracking
- **Analytics**: Post view tracking and statistics
- **Security**: JWT authentication, rate limiting, and input validation
- **Performance**: Database connection pooling, caching, and optimized queries
- **Monitoring**: Comprehensive logging and metrics

## API Structure

The API follows REST conventions with versioning:
- Base path: `/v1`
- Authentication: JWT tokens in Authorization header
- Response format: JSON with standardized structure

Available API groups:
- `/v1/users` - User management
- `/v1/posts` - Content management
- `/v1/auth` - Authentication endpoints
- `/v1/tags` - Tag management
- `/v1/chat/conversations` - Chat functionality
- Debug routes available when `DEBUG=true`

## Development Conventions

### Code Style
- **Imports**: Group imports: 1) Standard library 2) Local (`fiberbackend/...`) 3) Third-party
- **Formatting**: Use `gofmt`, max line length 140 characters, tabs for indentation
- **Naming**: Packages lowercase, exported PascalCase, unexported camelCase
- **JSON/DB fields**: snake_case (e.g., `first_name`, `created_at`)

### Architecture Patterns
- **Layer order**: Handler → Service → Repository → Model
- **Handler**: HTTP handling, input validation, responses
- **Service**: Business logic, orchestration, transactions
- **Repository**: Data access, GORM queries
- **Model**: Structs, DB tags, conversion methods (`ToResponse`, `ToSummary`)

### Dependency Injection
Uses manual DI with cleanup manager (`internal/di/container.go`):
- Registration order: Config → Database → Repositories → Services → Handlers → Routes
- Cleanup manager handles resource cleanup on shutdown

### Error Handling
- Use `AppError` from `pkg/utils/errors.go`
- Wrap errors with `fmt.Errorf("...: %w", err)`
- Return early on errors
- HTTP handlers use `pkg/response/` helpers

### Response Patterns
Standardized response helpers in handlers:
- `response.Success(c, message, data)` - 200 OK
- `response.Created(c, message, data)` - 201 Created
- `response.BadRequest(c, message, err)` - 400
- `response.Unauthorized(c, message)` - 401
- `response.Forbidden(c, message)` - 403
- `response.NotFound(c, message, err)` - 404
- `response.InternalServerError(c, message, err)` - 500
- `response.HandleBindError(c, err)` - Validation errors

### Validation
Uses `go-playground/validator` tags on request structs:
```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}
```

### Testing
- Table-driven tests with `testify/assert` or `testify/require`
- Mock at repository level
- Test file naming: `*_test.go`, functions: `TestXxx`

### Database Migrations
SQL migrations in `migrations/` folder. Apply manually via psql:
```bash
psql -d your_database -f migrations/001_*.sql
```

## Important Files

- `cmd/main.go` - Application entry point with graceful shutdown
- `config/config.go` - Centralized configuration management
- `internal/di/container.go` - Dependency injection container setup
- `internal/routes/routes.go` - API route definitions
- `internal/middleware/setup.go` - Security and utility middleware
- `pkg/database/setup.go` - Database connection with retry logic
- `pkg/response/response.go` - Standardized HTTP response helpers
- `pkg/utils/errors.go` - Application error types
- `.air.toml` - Development hot reload configuration
- `.golangci.yml` - Linter configuration
- `Dockerfile` - Multi-stage Docker build
- `fly.toml` - Fly.io deployment configuration
- `AGENTS.md` - Development guidelines for AI agents

## Project Structure Notes

The project implements a layered architecture with clear separation between:
1. HTTP layer (handlers and routes)
2. Business logic layer (services)
3. Data access layer (repositories)
4. Data models (models)

The dependency injection system in the `internal/di` package manages the lifecycle of all services and ensures loose coupling between components.
