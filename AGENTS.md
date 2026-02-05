# AGENTS.md - Agent Guidelines for FiberBackend

This document provides guidelines for AI agents working on the FiberBackend Go project.

## Project Overview

FiberBackend is a REST API built with Go 1.25+, Fiber v3 framework, PostgreSQL, and GORM. It follows clean architecture with handler/service/repository layers and uses Uber's dig for dependency injection.

## Build/Lint/Test Commands

```bash
# Build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/main cmd/main.go

# Run
air                              # Hot reload development
go run cmd/main.go              # Direct run

# Lint
golangci-lint run               # Run all configured linters
gofmt -l .                      # Check formatting
go vet ./...                    # Vet all packages

# Test
go test ./...                   # Run all tests
go test ./pkg/utils/...         # Run tests in specific package
go test -run TestIsValidEmail   # Run single test by name
go test -v ./...                # Verbose output

# Dependencies
go mod download
go mod tidy
go mod verify
```

## Code Style Guidelines

### Imports

Group imports in this order (separated by blank lines):
1. Standard library imports
2. Local project imports (`fiberbackend/...`)
3. Third-party imports

Example:
```go
import (
	"context"
	"net/http"

	"fiberbackend/internal/model"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)
```

### Formatting

- Use `gofmt` for automatic formatting
- Maximum line length: 140 characters (per .golangci.yml)
- Use tabs for indentation
- No trailing whitespace

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `handler`, `service`, `repository`)
- **Exported identifiers**: PascalCase (e.g., `UserService`, `GetByID`)
- **Unexported identifiers**: camelCase (e.g., `userRepo`, `validateInput`)
- **Interfaces**: Noun describing capability, often ending in "-er" (e.g., `UserService`, `Repository`)
- **Structs**: Nouns, exported if used outside package
- **Methods**: Verbs or verb phrases
- **Constants**: PascalCase for exported, camelCase for unexported
- **Test files**: `*_test.go`, test functions: `TestXxx`, `BenchmarkXxx`
- **Database fields**: snake_case in struct tags (e.g., `json:"first_name"`)
- **JSON keys**: snake_case (e.g., `created_at`, `user_id`)

### Types and Structs

- Define interfaces in the consumer package (usually service or handler)
- Keep struct fields ordered: ID, timestamps, other fields, relationships
- Use pointer types for nullable database fields
- Add struct tags for JSON and GORM
- Document exported types with comments starting with the type name

Example:
```go
// User represents the user model in the database
type User struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt *time.Time     `json:"created_at"`
	UpdatedAt *time.Time     `json:"updated_at"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Profile   *Profile       `gorm:"foreignKey:UserID"`
}
```

### Error Handling

- Use the custom `AppError` type from `pkg/utils/errors.go` for application errors
- Wrap errors with context using `fmt.Errorf("...: %w", err)`
- Return early on errors to reduce nesting
- Check sentinel errors with `errors.Is()` (e.g., `err == service.ErrUserExists`)
- Use `errors.As()` for error type assertions
- HTTP handlers should use response helpers from `pkg/response/`

Example:
```go
func (s *userService) GetByID(ctx context.Context, id string) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	return user.ToResponse(), nil
}
```

### Response Patterns

Use standardized response helpers in handlers:
- `response.Success(c, message, data)` - 200 OK
- `response.Created(c, message, data)` - 201 Created
- `response.BadRequest(c, message, err)` - 400 Bad Request
- `response.Unauthorized(c, message)` - 401 Unauthorized
- `response.Forbidden(c, message)` - 403 Forbidden
- `response.NotFound(c, message, err)` - 404 Not Found
- `response.InternalServerError(c, message, err)` - 500 Internal Server Error
- `response.HandleBindError(c, err)` - Handle validation errors

### Architecture Patterns

**Layer order**: Handler → Service → Repository → Model

- **Handler**: HTTP request handling, input validation, response formatting
- **Service**: Business logic, orchestration, transaction coordination
- **Repository**: Data access, database queries, GORM operations
- **Model**: Struct definitions, database tags, conversion methods (ToResponse, ToSummary)

**Dependency Injection**: 
- Use Uber's `dig` container (see `internal/di/container.go`)
- Register providers in order: Config → Database → Repositories → Services → Handlers → Routes
- Constructor injection: `func NewXxxService(repo Repository) Service`

### Comments

- All exported identifiers must have a comment starting with the identifier name
- Use complete sentences with proper punctuation
- Document the "why" not just the "what"
- Add package documentation in a `doc.go` file or at the top of the main package file

### Testing

- Test files in same package as code being tested or `*_test` package
- Use table-driven tests for multiple test cases
- Use `testify/assert` or `testify/require` for assertions
- Mock external dependencies at repository level
- Test function naming: `Test<FunctionName>` or `Test<Struct>_<Method>`
- Skip test boilerplate rules with `//nolint` if necessary

Example:
```go
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "Valid email", email: "test@example.com", want: true},
		{name: "Invalid email", email: "invalid", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEmail(tt.email); got != tt.want {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

### Validation

- Use `go-playground/validator` tags on struct fields
- Define request structs with validation tags in handlers
- Use `response.HandleBindError()` to return validation errors

Example:
```go
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
```

### Configuration

- Use `config/config.go` for all configuration
- Environment variables with defaults
- Required validation in `validate()` method
- Never commit secrets; use `.env` file (already in `.gitignore`)

## Key Files

- `cmd/main.go` - Application entry point
- `config/config.go` - Configuration management
- `internal/di/container.go` - Dependency injection setup
- `pkg/response/response.go` - HTTP response helpers
- `pkg/utils/errors.go` - Application error types
- `.golangci.yml` - Linter configuration
- `.air.toml` - Hot reload configuration
