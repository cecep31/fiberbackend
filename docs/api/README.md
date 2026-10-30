# API Documentation - fiberbackend

HTTP API reference for frontend integration. All business routes are under the `/api` prefix, except the root health check endpoints.

## Base URL

| Environment | URL |
|-------------|-----|
| Local | `http://localhost:<PORT>` (see `PORT` in `.env`) |
| Production | Your deployment / reverse proxy URL |

Full examples: `GET /api/posts`, `POST /api/auth/login`.

## Authentication

Routes that require login must send this header:

```http
Authorization: Bearer <access_token>
```

Tokens are returned by `POST /api/auth/login`, `POST /api/auth/refresh`, or `POST /api/auth/oauth/exchange` after GitHub OAuth. JWT claims include `user_id` (UUID).

Failed auth middleware responses (missing token, invalid token, or non-super-admin user on admin routes) use the standard `success` envelope below, with `success: false` and a generic `error` string:

| Situation | HTTP | Body |
|-----------|------|------|
| Missing / invalid token | 401 | `{"success":false,"message":"...","error":"Unauthorized access"}` |
| Not a super admin on an admin route | 403 | `{"success":false,"message":"...","error":"Access forbidden"}` |

## Standard Response Format

Most handlers use the `pkg/response` envelope:

```json
{
  "success": true,
  "message": "Human-readable message",
  "data": {},
  "meta": {},
  "error": "...",
  "errors": []
}
```

`data`, `meta`, `error`, and `errors` use `omitempty` — they are omitted from the JSON when empty (for example, success responses have no `error`/`errors` keys, and error responses usually have no `data`/`meta` keys).

| Helper | HTTP | Notes |
|--------|------|-------|
| Success | 200 | `success: true`, optional `data` |
| Created | 201 | Same envelope as success |
| Bad request | 400 | `success: false`, `error` contains details |
| Unauthorized | 401 | `error`: `"Unauthorized access"` |
| Forbidden | 403 | `error`: `"Access forbidden"` |
| Not found | 404 | |
| Conflict | 409 | Duplicate resource |
| Validation | 422 | `errors` contains field errors (`field`, `message`, `value`, `tag`) |
| Server error | 500 | Generic client message; details are logged server-side only |

### Pagination (`meta`)

Paginated lists use `SuccessWithMeta`:

```json
{
  "meta": {
    "total_items": 100,
    "offset": 0,
    "limit": 10,
    "total_pages": 10
  }
}
```

Query: `limit` (default varies by endpoint, **maximum 100** on most endpoints — `GET /api/posts` is an exception and currently accepts larger values), `offset` (default `0`).

## Global Limits

- Request body size: **10 MB**. Larger requests return **413** `{"success":false,"message":"Request Entity Too Large","error":"Request body too large"}`.
- CORS: `HTTP_ALLOW_ORIGINS` (default `*`). CORS headers are also set on 404/405/413/429 responses.
- Global rate limit: enabled when `HTTP_RATE_LIMIT_RPS` (alias `RATE_LIMITER_MAX`) > 0. It is a per-client-IP token bucket: `HTTP_RATE_LIMIT_RPS` tokens per second, burst 2× that. `HTTP_RATE_LIMIT_WINDOW_SEC` (alias `RATE_LIMITER_TTL`) only sets how long an idle client is remembered (default 3 minutes); it is not a counting window. A denied request gets **429** with a `Retry-After` header (seconds) and `{"success":false,"message":"Too many attempts. Please try again later.","error":"Rate limit exceeded"}`.
- Client IP: with `HTTP_TRUST_PROXY=true` (alias `TRUST_PROXY`), `X-Forwarded-For` is walked right to left and the first hop outside loopback, private, link-local and CGNAT (`100.64.0.0/10`) ranges is the client IP; otherwise the peer address is used. It keys every per-IP limit and is stored in sessions, activity logs and post views.
- Framework-level errors use the envelope too: unknown route → **404** `{"success":false,"message":"Route not found","error":"Resource not found"}`; known path, unsupported method → **405** `{"success":false,"message":"Method not allowed","error":"Method not allowed"}`; unhandled error or panic → **500** `{"success":false,"message":"Internal server error"}`.
- Security headers on every response: `X-XSS-Protection: 1; mode=block`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`, `Content-Security-Policy: default-src 'self'`, `Referrer-Policy: strict-origin-when-cross-origin`; over HTTPS (TLS, or a forwarded https scheme from a loopback/private proxy) also `Strict-Transport-Security: max-age=3600; includeSubdomains`.
- Auth-specific rate limits use a fixed window per IP. If `VALKEY_URL` is set, counters are stored in Valkey/Redis and work across instances; otherwise they fall back to in-memory per instance:
  `register`, `login`, and `reset-password` **5 / 5 minutes**;
  `forgot-password` **3 / 5 minutes**;
  `refresh` **30 / minute**;
  `oauth/exchange` **10 / minute**.

## Health & Root

| Method | Path | Auth | Response |
|--------|------|------|----------|
| GET | `/` | No | Success envelope with welcome message |
| GET | `/health` | No | `200` `{"status":"ok"}` or `503` `{"status":"unhealthy","reason":"database unreachable"}` |

## Modules

| Module | Base path | Document |
|--------|-----------|----------|
| Auth | `/api/auth` | [auth.md](./auth.md) |
| Users & follow | `/api/users` | [users.md](./users.md) |
| Posts (comments, views, likes) | `/api/posts` | [posts.md](./posts.md) |
| Tags | `/api/tags` | [tags.md](./tags.md) |
| Chat | `/api/chat/conversations`, `/api/chat/messages` | [chat.md](./chat.md) |
| Holdings | `/api/holdings`, `/api/holding-types` | [holdings.md](./holdings.md) |
| Exchange rates | `/api/exchange-rates` | [exchange-rates.md](./exchange-rates.md) |
| Bookmarks | `/api/bookmarks` | [bookmarks.md](./bookmarks.md) |
| Notifications | `/api/notifications` | [notifications.md](./notifications.md) |
| Guilds (channels, messages, SSE) | `/api/guilds` | [guilds.md](./guilds.md) |
| Admin stats | `/api/posts/stats`, `/api/tags/stats`, `/api/users/stats` | [stats.md](./stats.md) |

Post reporting and moderation routes live under `/api/posts` — see the Moderation section in [posts.md](./posts.md). The legacy `/api/reports` module was removed; use the stats endpoints instead.

Debug routes (`/api/debug/pprof/*`) are registered only when `APP_DEBUG=true`; they are not intended for frontend use.

## Type Conventions

- **UUID**: string primary key for users, posts, comments, and conversations.
- **Time**: ISO 8601 / RFC3339 (`2026-05-12T08:00:00Z`).
- **Nullable**: Go pointer fields may be `null` or omitted (`omitempty`).
- **Financial numbers (holdings)**: decimal strings in JSON (for example `"1500000.00"`), not numbers.
