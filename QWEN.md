# FiberBackend - Modern Go REST API

## Project Overview

FiberBackend is a modern REST API for a blog/content management system built with Go and the Fiber web framework. Despite the README incorrectly mentioning the Echo framework, the project actually uses the Fiber framework (v3.0.0) as evidenced in the go.mod file. The application connects to PostgreSQL for data persistence and includes features like user management, content management, social features, real-time chat, and file storage.

### Key Technologies
- **Go**: 1.25+
- **Fiber Framework**: v3.0.0 (not Echo as incorrectly mentioned in README)
- **GORM**: v1.31.1 for database ORM
- **PostgreSQL**: 14+ via pgx driver
- **JWT Authentication**: github.com/golang-jwt/jwt/v5
- **File Storage**: MinIO/S3 compatible storage
- **Dependency Injection**: go.uber.org/dig
- **Validation**: go-playground/validator

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
- **migrations/**: Database migration files
- **test/**: Test files
- **pkg/**: Shared utilities
  - **database/**: Database connection and wrapper
  - **response/**: Standardized response helpers
  - **storage/**: File storage utilities
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

   # File Storage (MinIO/S3)
   MINIO_ENDPOINT=localhost:9000
   MINIO_ACCESS_KEY=minioadmin
   MINIO_SECRET_KEY=minioadmin
   MINIO_BUCKET=minio-bucket

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
# Build and run
make build
make dev

# Run tests
make test

# Code quality
make fmt
make vet
make lint
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
- **File Storage**: MinIO/S3 integration for file uploads and management
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

- **Code Style**: Follows Go idiomatic conventions
- **Testing**: Unit and integration tests using testify package
- **Dependency Injection**: Uses Uber's dig for dependency injection
- **Configuration**: Environment-based configuration with defaults
- **Logging**: Structured logging with request ID correlation
- **Error Handling**: Consistent error response format
- **Security**: Built-in middleware for CORS, helmet, rate limiting, etc.

## Project Structure Notes

The project correctly implements a layered architecture with clear separation between:
1. HTTP layer (handlers and routes)
2. Business logic layer (services)
3. Data access layer (repositories)
4. Data models (models)

The dependency injection system in the `internal/di` package manages the lifecycle of all services and ensures loose coupling between components.

## Important Files

- `cmd/main.go` - Application entry point with graceful shutdown
- `config/config.go` - Centralized configuration management
- `internal/routes/routes.go` - API route definitions
- `internal/middleware/setup.go` - Security and utility middleware
- `pkg/database/setup.go` - Database connection with retry logic
- `.air.toml` - Development hot reload configuration
- `Dockerfile` - Multi-stage Docker build
- `fly.toml` - Fly.io deployment configuration