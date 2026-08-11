# AGENTS.md - Agent Guidelines for FiberBackend

REST API built with Go 1.26, **Fiber v3**, PostgreSQL, and GORM. Clean architecture with handler/service/repository layers using **manual dependency injection** (`internal/di/container.go`). Synced feature-for-feature with `echobackend` (Echo v5 twin); only the HTTP layer differs (Fiber vs Echo).

## Build/Lint/Test Commands

```bash
# Local dev services (Postgres 18 + Valkey + MinIO via docker-compose.yml)
docker compose up -d --wait # Start services (or: make up). Creates the `custom` schema automatically.
docker compose down         # Stop (or: make down; make down-clean also wipes data)
make help                   # All shortcuts: up, dev, test, lint, check, migrate-*, ...

# Run
air                              # Hot reload development (reads .env automatically)
go run cmd/main.go              # Direct run

# Lint
golangci-lint run               # Run all linters (see .golangci.yml)
gofmt -l .                      # Check formatting
go vet ./...                    # Vet all packages

# Test
go test ./...                   # Run all tests (service + pkg layers only; no DB integration tests)
go test -race ./...             # Race checker
go test -cover ./...            # With coverage

# Migrations (requires .env with GOOSE_* vars)
goose up                        # Apply pending
goose down                      # Rollback one
goose status                    # Check current
goose create <name> sql         # New migration file (always sql, never go)

# Dependencies
go mod download && go mod tidy && go mod verify
```

## Architecture

- **Framework**: Fiber **v3** (not v2). API differences: use `c fiber.Ctx`, `c.Bind().Body(&req)`, `c.Params("id")`, `c.Query("x")`, middleware signature `func(c fiber.Ctx) error`.
- **Entry point**: `cmd/main.go` — loads config → creates DI container → registers routes → starts server with graceful shutdown (10s).
- **DI**: Manual wiring in `internal/di/container.go`. All handler/service/repo/platform instances created there. **No reflection-based DI (e.g. dig).**
- **Layering**: `handler` → `service` → `repository` → `model`.
- **`internal/platform/`**: App-owned infrastructure adapters (`cache`, `database`, `email`, `queue`, `storage`). All fail-open when their optional config is empty.
- **`internal/dto/`**: Request/response structs + converters. `internal/apperror/` for shared app error sentinels.
- **`pkg/`**: Reusable helper packages (`applog`, `market`, `response`, `validator`).
- **API routes**: All under `/api/*`, defined in `internal/routes/*Routes.go`. Auth-protected routes use `r.authMiddleware.Auth()`; admin routes chain `r.authMiddleware.AuthAdmin()`.
- **Auth context**: Auth middleware stores validated JWT claims under `c.Locals("user", claims)`. Handlers read them via `handler.GetUserIDFromClaims(c)`.
- **Health**: `GET /health` — pings DB (200/503). Used by Docker HEALTHCHECK and load balancers.
- **Pagination**: Use `handler.ParsePaginationParams(c, defaultLimit)` — returns `(limit, offset)`, max cap 100. Build meta with `response.CalculatePaginationMeta(total, offset, limit)` and pass via `response.SuccessWithMeta`.
- **Streaming (chat)**: SSE via `c.RequestCtx().SetBodyStreamWriter(...)` — see `internal/handler/chat_conversation_handler.go`.

## Config & Env

- Config loaded via `config.Load()` (`config/config.go`) — reads `.env` (godotenv) then environment variables. Grouped into sections: `App`, `HTTP`, `Auth`, `Database`, `S3`, `Cache`, `Queue`, `OpenRouter`, `GitHub`, `Frontend`, `Email`, `MarketData`.
- **Required**: `DATABASE_URL`, `JWT_SECRET`. App panics if missing (JWT must be ≥ 32 chars).
- Many keys accept **fallback aliases** (legacy names such as `MAX_OPEN_CONNS`, `RATE_LIMITER_MAX`, `DEBUG`). First-set key wins.
- `GOOSE_TABLE=custom.goose_migrations` — non-default goose table location; create the `custom` schema before the first `goose up`.
- Valkey/Redis caching, SMTP email, Asynq queue, S3/MinIO, GitHub OAuth, OpenRouter and market-data are **optional** — leave their keys empty to disable (app runs fail-open).

## Testing

- Tests exist in `internal/service/`, `internal/dto/`, `config/`, `pkg/`. No repository or DB integration tests.
- **No external test dependencies** — service tests use hand-written mocks (`internal/service/mocks_test.go`).
- Running `go test ./...` does not require PostgreSQL.
- Test file pattern: `*_test.go` in the same package (white-box).

## Response Format

All handlers use `pkg/response` for consistent JSON:

```go
response.Success(c, "message", data)        // 200
response.Created(c, "message", data)         // 201
response.ValidationError(c, "msg", err)       // 422
response.BadRequest(c, "msg", err)            // 400
response.NotFound(c, "msg", err)              // 404
response.Unauthorized(c, "msg")              // 401
response.Forbidden(c, "msg")                 // 403
response.Conflict(c, "msg", conflictErr)      // 409
response.InternalServerError(c, "msg", err)  // 500 — err logged server-side only, never sent to client
response.FromValidateError(c, err)            // 422 with structured field errors
response.TooManyRequests(c, "msg")           // 429 (rate limit)
```

Validation failures map to **422** with a structured `errors` array (`response.FromValidateError`). `response.Conflict` takes a string reason, not an `error`.

## CI

`.github/workflows/main.yml` runs on PRs and pushes to `main`:
1. **test** — `go vet ./...`, `go test ./...`, `golangci-lint` (v2.12.2)
2. **docker** (push to `main` only, after test) — build & push `cecep31/fiberbackend:latest`, `:sha-<12-char>`, and `:sha-<full>` (awaiting)
3. **deploy** (after docker) — Fly.io `flyctl deploy --remote-only`

## Migrations

- Goose with **raw SQL** files in `migrations/`. Numbered `001_init_schema.sql` through `012_add_corporate_actions.sql`.
- Uses PostgreSQL features: triggers for count fields, UUID v7 defaults, `ON DELETE CASCADE`, soft deletes via `deleted_at`.
- **New migrations**: `goose create <name> sql` (always `sql`, never `go`).
- **First-time setup**: The local Postgres from `docker compose up` auto-creates the `custom` schema (`scripts/init-db.sql`). For an external Postgres, run `psql "$DATABASE_URL" -c 'CREATE SCHEMA IF NOT EXISTS custom;'` once before the first `goose up`.

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
- Max line length: 140 characters (see `.golangci.yml` `lll` setting)
- Tabs for indentation
- No trailing whitespace

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `handler`, `service`)
- **Exported**: PascalCase (e.g., `UserService`, `GetByID`)
- **Unexported**: camelCase (e.g., `userRepo`)
- **Constants**: PascalCase for exported
- **Test files**: `*_test.go`, functions: `TestXxx`
- **JSON/DB fields**: snake_case (e.g., `first_name`, `created_at`)

### Error Handling

Use sentinel errors from `internal/apperror` (or package-local `var` declarations). Wrap with `fmt.Errorf("...: %w", err)`. Check sentinel errors with `errors.Is()`. HTTP handlers map apperror sentinels to the right `pkg/response` helper — never leak `err.Error()` on 500.

### Validation

Use `go-playground/validator` tags. Request structs live in `internal/dto/`. Validate explicitly after binding:

```go
var req dto.CreatePostRequest
if err := c.Bind().Body(&req); err != nil {
	return response.BadRequest(c, "Failed to create post", err)
}
if err := bindValidate(c, &req); err != nil {
	return err
}
```

## Key Files

- `cmd/main.go` - Entry point
- `config/config.go` - Configuration (grouped, alias-aware)
- `internal/di/container.go` - Manual DI container
- `internal/middleware/setup.go` - Middleware setup
- `internal/routes/routes.go` - Route definitions
- `pkg/response/response.go` - HTTP response helpers
- `internal/apperror/errors.go` - Application error sentinels
- `.golangci.yml` - Linter config
- `.air.toml` - Hot reload config
