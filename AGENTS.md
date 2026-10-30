# AGENTS.md - Agent Guidelines for FiberBackend

REST API built with Go 1.26, **Fiber v3**, PostgreSQL, and GORM. Clean architecture with handler/service/repository layers using **manual dependency injection** (`internal/di/container.go`). Synced feature-for-feature with `echobackend` (Echo v5 twin); only the HTTP layer differs (Fiber vs Echo).

## Build/Lint/Test Commands

```bash
# Local dev services (Postgres 18 + Valkey + MinIO via docker-compose.yml)
docker compose up -d --wait # Start services (or: make up). Creates the `custom` schema automatically.
docker compose down         # Stop (or: make down; make down-clean also wipes data)
make help                   # All shortcuts: up, dev, test, lint, check, migrate-*, ...
# Windows gets no `make` by default — use Git Bash/WSL/Chocolatey make, or run the underlying go/docker/goose commands directly.

# Run
air                              # Hot reload development (reads .env automatically)
go run cmd/main.go              # Direct run

# Lint (golangci-lint v2 — see .golangci.yml)
golangci-lint run ./...         # All linters + gofmt/goimports formatter checks
golangci-lint fmt ./...         # Auto-format (gofmt + goimports); same as `make fmt`
go vet ./...                    # Vet all packages

# Test
go test ./...                   # Run all tests — no Postgres/Redis required (service + pkg layers only, hand-written mocks)
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
- **`internal/platform/`**: App-owned infrastructure adapters (`cache`, `database`, `email`, `market`, `openrouter`, `realtime`, `storage`). All fail-open when their optional config is empty.
- **`internal/dto/`**: Request/response structs + converters. `internal/apperror/` for shared app error sentinels.
- **`pkg/`**: Reusable helper packages (`antispam`, `applog`, `password`, `response`, `slug`, `uid`, `validator`).
- **API routes**: All under `/api/*`, defined in `internal/routes/*_routes.go`. Auth-protected routes use `r.authMiddleware.Auth()`; admin routes must chain `Auth()` **first**, then `r.authMiddleware.AuthAdmin()` (`AuthAdmin` reads the claims set by `Auth()` and returns 401 without them). `OptionalAuth()` exists for public-but-personalized endpoints. The `is_super_admin` claim is a fast path only — the DB (`GetAdminByID`) is authoritative.
- **Auth context**: Auth middleware stores validated JWT claims under `c.Locals("user", claims)`. Handlers read them via `handler.GetUserIDFromClaims(c)`.
- **Refresh tokens rotate** (RFC 9700 §4.14): `POST /api/auth/refresh` returns a **new** refresh token and retires the old one. Rows in `sessions` form a family (`family_id`) per login; rotated rows are kept, not deleted, so a replayed token is detectable — outside `REFRESH_TOKEN_GRACE_PERIOD` it revokes the whole family. `REFRESH_TOKEN_EXPIRY` is a sliding inactivity window, `REFRESH_TOKEN_ABSOLUTE_EXPIRY` a hard ceiling inherited unchanged by every rotation. Never return the same refresh token twice, and never store one unhashed.
- **Guild chat** (Discord-style): channels + messages under `/api/guilds/:slug/channels`, realtime via one SSE stream per guild (`GET /api/guilds/:slug/events`). Events go through `internal/platform/realtime.Hub` — Redis pub/sub fan-out when Redis is up, in-process otherwise. SSE handlers stream via `c.RequestCtx().SetBodyStreamWriter(...)`; the handler returns as soon as the writer is registered, so any resource held by the stream (e.g. the realtime subscription) must be closed **inside** the writer callback, not via `defer` in the handler. `cmd/main.go` calls `container.CloseStreams()` before `app.ShutdownWithContext` so open streams don't stall graceful shutdown. Message history paginates by `?before=<uuidv7 id>` cursor, not offset.
- **Moderation**: post reports + admin hide/unhide under `/api/posts/:id/reports` and `/api/posts/:id/moderation` (`PostReportService`). Hiding a post sets `hidden_at`; posts and likes are hard-deleted (no `deleted_at`).
- **Health**: `GET /health` — pings DB (200/503). Used by Docker HEALTHCHECK and load balancers.
- **Pagination**: Use `handler.ParsePaginationParams(c, defaultLimit)` — returns `(limit, offset)`, max cap 100. Build meta with `response.CalculatePaginationMeta(total, offset, limit)` and pass via `response.SuccessWithMeta`.
- **Streaming (chat)**: SSE via `c.RequestCtx().SetBodyStreamWriter(...)` — see `internal/handler/chat_conversation_handler.go` and `guild_channel_handler.go`.

## Config & Env

- Config loaded via `config.Load()` (`config/config.go`) — reads `.env` (stdlib loader, best-effort) then environment variables. Grouped into sections: `App`, `HTTP`, `Auth`, `Database`, `S3`, `Cache`, `OpenRouter`, `GitHub`, `Frontend`, `Email`, `MarketData`.
- **Required**: `DATABASE_URL`, `JWT_SECRET`. App panics if missing (JWT must be ≥ 32 chars).
- **Auth durations** accept duration strings: `JWT_EXPIRY` (default 15m), `REFRESH_TOKEN_EXPIRY` (sliding inactivity window, default 72h), `REFRESH_TOKEN_ABSOLUTE_EXPIRY` (hard ceiling, default 720h), `REFRESH_TOKEN_GRACE_PERIOD` (rotation replay window, default 1m).
- **Pool knobs** (all optional): `DB_POOL_MAX_OPEN_CONNS`, `DB_POOL_MAX_IDLE_CONNS`, `DB_POOL_CONN_MAX_LIFETIME`, `DB_POOL_CONN_IDLE_TIME` (aliases: `MAX_OPEN_CONNS`, `MAX_IDLE_CONNS`, `CONN_MAX_LIFETIME`, `CONN_MAX_IDLE_TIME`).
- Many keys accept **fallback aliases** (legacy names such as `MAX_OPEN_CONNS`, `RATE_LIMITER_MAX`, `DEBUG`). First-set key wins. RapidAPI market data accepts `RAPIDAPI_KEY` (aliases: `RAPIDAPI_IDX_KEY`, `RAPIDAPI_QUOTE_KEY`).
- `GOOSE_TABLE=custom.goose_migrations` — non-default goose table location; create the `custom` schema before the first `goose up`.
- Valkey/Redis caching, SMTP email, S3/MinIO, GitHub OAuth, OpenRouter and market-data are **optional** — leave their keys empty to disable (app runs fail-open).

## Testing

- Tests live in `internal/service/`, `internal/handler/`, `internal/middleware/`, `internal/dto/`, `internal/platform/`, `config/`, `pkg/`. No repository or DB integration tests.
- **No external test dependencies** — service tests use hand-written mocks (`internal/service/mocks_test.go`). Handler/middleware tests drive requests through a real Fiber app via `app.Test(req)` (SSE tests use `app.Listener` + a real HTTP client).
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

- Goose with **raw SQL** files in `migrations/`. Numbered `001_init_schema.sql` through `015_add_users_lower_email_index.sql`, plus timestamped files (`20260916032847_add_core_feed_indexes.sql` … `20260930100000_add_post_reports.sql`) — keep numbering and timestamps in lockstep with `echobackend/migrations/`.
- Uses PostgreSQL features: triggers for count fields, UUID v7 defaults, `ON DELETE CASCADE`, soft deletes via `deleted_at`.
- Soft deletes apply to `users`, `post_views`, `post_comments`, `user_follows`, `files`, `chat_conversations` only. **`posts` and `post_likes` are hard-deleted since migration 014**. **`guilds`, `guild_members`, `guild_channels` and `guild_channel_messages` are hard-deleted too** (deleting a guild cascades).
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

- `.golangci.yml` is **v2 format**: it enables `gofmt` + `goimports` as *formatters* and `default: standard` linters plus extras (errorlint, revive, gocritic, bodyclose, nilerr, noctx, ...). `golangci-lint run ./...` catches formatting, not just lint. CI pins golangci-lint **v2.12.2**.
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

### Logging

Use `pkg/applog` (a thin wrapper over `log/slog`) instead of bare `fmt.Println`/`log`. Define a package-level component logger and call its `.Debug/.Info/.Warn/.Error` methods:

```go
var log = applog.Component("auth") // adds "component" field; safe as a package var
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
