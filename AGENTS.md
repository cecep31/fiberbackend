# AGENTS.md - Agent Guidelines for FiberBackend

REST API built with Go 1.25+, Fiber v3, PostgreSQL, and GORM. Clean architecture with handler/service/repository layers using Uber's dig for dependency injection.

## Build/Lint/Test Commands

```bash
# Build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/main cmd/main.go

# Run
air                              # Hot reload development
go run cmd/main.go              # Direct run

# Lint
golangci-lint run               # Run all linters (see .golangci.yml)
gofmt -l .                      # Check formatting
go vet ./...                    # Vet all packages

# Test
go test ./...                   # Run all tests
go test ./pkg/utils/...         # Run tests in specific package
go test -run TestIsValidEmail   # Run single test by name
go test -v ./...                # Verbose output
go test -cover ./...            # With coverage
go test -short ./...            # Short mode (skip long-running tests)

# Dependencies
go mod download && go mod tidy && go mod verify
```

## Code Style Guidelines

### Imports

Group imports: 1) Standard library 2) Local (`fiberbackend/...`) 3) Third-party. Separate groups with blank lines.

```go
import (
	"context"

	"fiberbackend/internal/model"
	"fiberbackend/pkg/response"

	"github.com/gofiber/fiber/v3"
)
```

### Formatting

- Use `gofmt` for formatting
- Max line length: 140 characters
- Tabs for indentation
- No trailing whitespace

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `handler`, `service`)
- **Exported**: PascalCase (e.g., `UserService`, `GetByID`)
- **Unexported**: camelCase (e.g., `userRepo`)
- **Interfaces**: Noun ending in "-er" (e.g., `UserService`)
- **Constants**: PascalCase for exported
- **Test files**: `*_test.go`, functions: `TestXxx`
- **JSON/DB fields**: snake_case (e.g., `first_name`, `created_at`)

### Types and Structs

Define interfaces in consumer package. Order fields: ID, timestamps, other fields, relationships. Use pointers for nullable DB fields. Document exported types.

```go
// User represents the user model in the database
type User struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt *time.Time     `json:"created_at"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Profile   *Profile       `gorm:"foreignKey:UserID"`
}
```

### Error Handling

Use `AppError` from `pkg/utils/errors.go`. Wrap with `fmt.Errorf("...: %w", err)`. Return early on errors. Check sentinel errors with `errors.Is()`. HTTP handlers use `pkg/response/` helpers.

### Response Patterns

Use standardized helpers in handlers:
- `response.Success(c, message, data)` - 200 OK
- `response.Created(c, message, data)` - 201 Created
- `response.BadRequest(c, message, err)` - 400
- `response.Unauthorized(c, message)` - 401
- `response.Forbidden(c, message)` - 403
- `response.NotFound(c, message, err)` - 404
- `response.InternalServerError(c, message, err)` - 500
- `response.HandleBindError(c, err)` - Validation errors

### Architecture

**Layer order**: Handler → Service → Repository → Model

- **Handler**: HTTP handling, input validation, responses
- **Service**: Business logic, orchestration, transactions
- **Repository**: Data access, GORM queries
- **Model**: Structs, DB tags, conversion methods (`ToResponse`, `ToSummary`)

**Dependency Injection**: Use Uber's `dig` container (`internal/di/container.go`). Register: Config → Database → Repositories → Services → Handlers → Routes.

### Middleware

Middleware is defined in `internal/middleware/`. Use ` fiber.Ctx` methods for request context. Auth middleware validates JWT tokens and sets user context. Apply middleware in routes using `app.Use()` or route-level with `config fiber.Config` in routes.

### Database Migrations

SQL migrations in `migrations/` folder. Apply manually via psql:
```bash
psql -d your_database -f migrations/001_*.sql
```

### Comments

All exported identifiers need comments starting with the identifier name. Use complete sentences. Document the "why", not just "what".

### Testing

Table-driven tests with `testify/assert` or `testify/require`. Mock at repository level. Name tests: `Test<FunctionName>` or `Test<Struct>_<Method>`.

```go
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{name: "Valid email", email: "test@example.com", want: true},
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

Use `go-playground/validator` tags. Define request structs in handlers.

```go
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}
```

### Configuration

Use `config/config.go`. Environment variables with defaults. Required validation in `validate()` method. Never commit secrets; `.env` is in `.gitignore`.

## Key Files

- `cmd/main.go` - Entry point
- `config/config.go` - Configuration
- `internal/di/container.go` - DI container
- `internal/middleware/setup.go` - Middleware setup
- `internal/routes/routes.go` - Route definitions
- `pkg/response/response.go` - HTTP response helpers
- `pkg/utils/errors.go` - Application error types
- `.golangci.yml` - Linter config
- `.air.toml` - Hot reload config
