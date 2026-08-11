# Fiber Backend API

A modern, robust REST API built with Go 1.26, **Fiber v3**, GORM, and PostgreSQL. Feature-for-feature twin of `echobackend` (same API contract, Echo v5 in that repo) — only the HTTP layer differs.

## Core Features

- **User Management**: Auth (JWT + DB-backed refresh sessions), GitHub OAuth, password reset (SMTP/queue), activity logs, profiles, following.
- **Content Management**: Posts with rich text, tags, comments, bookmarks (+ folders), post views, likes, sitemap & trending.
- **Analytics**: Post analytics + likes-by-month, admin reports (overview/user/post/engagement), holdings dashboards.
- **Real-time Chat**: Conversational AI via OpenRouter with SSE streaming and message history.
- **Finance Tools**: Holdings with live price sync (Yahoo), monthly compare/trends, IDX corporate-action calendar, exchange rates.
- **Security**: JWT auth (DB-checked admin), per-endpoint rate limiting, input validation (422), helmet security headers, no error leakage.
- **Performance**: Valkey/Redis caching, connection pooling, GORM slow-query logging.
- **Infra**: Goose migrations, Docker (multi-stage, non-root, HEALTHCHECK), Docker Compose local services, GitHub Actions CI, Fly.io deploy.

## Tech Stack

- **Runtime**: Go 1.26
- **Web Framework**: [Fiber v3](https://github.com/gofiber/fiber)
- **ORM**: [GORM](https://gorm.io/) v2
- **Database**: PostgreSQL 18+
- **Cache**: Valkey/Redis (optional)
- **Storage**: MinIO / AWS S3 (optional)
- **Queue/Email**: Asynq + SMTP (optional)
- **Migrations**: [Goose](https://github.com/pressly/goose) (Raw SQL)

## Quick Start

```bash
# 1. Setup environment
cp .env.example .env       # set JWT_SECRET (>= 32 chars)

# 2. Start local services (Postgres 18, Valkey, MinIO)
docker compose up -d --wait

# 3. Apply database migrations
goose up

# 4. Run with hot reload (requires air)
air

# 5. Or run normally
go run cmd/main.go
```

The server starts at `http://localhost:8080` (health check at `/health`). A `Makefile` wraps common tasks (`make help`); on Windows use Git Bash/WSL or run the underlying commands directly.

## Environment Variables

The application requires the following mandatory environment variables to start:

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string (DSN) |
| `JWT_SECRET` | Secret key for JWT signing & verification (≥ 32 chars) |

All other configurations (S3, Valkey, SMTP, Asynq, GitHub OAuth, OpenRouter, market data, rate limiting, CORS) are optional and documented in [`.env.example`](.env.example). Many keys accept legacy fallback aliases (first-set wins).

## Architecture

Modular layered architecture with manual dependency injection:

- **`internal/di/`**: Centralized DI container (`container.go`).
- **`internal/handler/`**: Request handling and response formatting via `pkg/response`.
- **`internal/service/`**: Core business logic and service orchestration.
- **`internal/repository/`**: Data access layer using GORM.
- **`internal/model/`**: GORM entities and shared domain models.
- **`internal/dto/`**: Request/response structs and converters.
- **`internal/apperror/`**: Shared application error sentinels.
- **`internal/platform/`**: App-owned infrastructure adapters (cache, database, email, queue, storage).
- **`pkg/`**: Reusable helper packages (applog, market, response, validator).

## API Documentation

Full HTTP API reference lives in [`docs/api/README.md`](docs/api/README.md), with per-module docs for auth, users, posts, tags, chat, holdings, exchange rates, bookmarks, notifications, and admin reports.

### Standardized Responses

All handlers use `pkg/response` helpers to ensure a consistent API contract:
```go
return response.Success(c, "Data retrieved", data)
return response.ValidationError(c, "Invalid input", err)
```

## Database Migrations

Managed via [goose](https://github.com/pressly/goose). Configuration is automatically picked up from `.env`.

```bash
goose up        # Apply all pending migrations
goose down      # Rollback the last migration
goose status    # Check migration history
goose create <migration_name> sql
```

## Deployment

CI via GitHub Actions on push to `main`: test/lint → Docker build/push (`cecep31/fiberbackend:latest`, `sha-*` tags) → Fly.io deploy. Pull a pinned `sha-*` image for reproducible deploys.

```bash
# Local Docker build test
docker build -t cecep31/fiberbackend .

# Run container locally
docker run -p 8080:8080 --env-file .env cecep31/fiberbackend
```

## License

MIT
