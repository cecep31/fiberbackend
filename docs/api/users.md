# Users & Follow Module - `/api/users`

User profiles, admin user lists, follow/unfollow, and social statistics.

Every `:id` path segment must be a canonical lowercase UUID (`8-4-4-4-12`). Anything else returns **400** `{"success": false, "message": "Invalid user ID", "error": "invalid user ID format"}` before the database is queried.

## Data Types

### `UserResponse` (admin routes: `GET /`, `GET /:id`)

| Field | Type | Description |
|-------|------|-------------|
| `id` | string (UUID) | |
| `email` | string \| null | Omitted on public/general routes; present only on admin routes or `/me` |
| `name` | string | `first last` when both are set, otherwise whichever one is set; `""` when neither is |
| `username` | string \| null | |
| `image` | string \| null | Avatar URL |
| `first_name` | string \| null | |
| `last_name` | string \| null | |
| `followers_count` | number | |
| `following_count` | number | |
| `is_following` | boolean \| null | Present only on routes with auth context, for example admin `GET /:id` |
| `is_super_admin` | boolean \| null | Present only on admin routes (`GET /`, `GET /:id`) |
| `profile` | object \| null | Not loaded on `GET /` (admin list); available on other routes |
| `created_at` | string (ISO) \| null | |
| `updated_at` | string (ISO) \| null | |
| `deleted_at` | string (ISO) \| null | Admin routes only; set when the user has been soft-deleted |
| `last_logged_at` | string (ISO) \| null | Admin routes only (`GET /`, `GET /:id`); last recorded login timestamp, omitted when never set |

### `CurrentUserResponse` (`GET /me`)

Same fields as `UserResponse`, except:

- Always includes `is_super_admin` (not omitted).
- Always includes `email`.
- Never includes `deleted_at`, `is_following`, or `last_logged_at`.
- Includes `profile` when loaded.

### `PublicUserResponse` (`GET /username/:username`, follow lists)

Public profile shape of `UserResponse`. Omits `email`, `is_super_admin`, `deleted_at`, and `last_logged_at`. `is_following` is always omitted on `GET /username/:username` because the route has no auth middleware.

### `Profile`

| Field | Type |
|-------|------|
| `id` | number |
| `user_id` | string (UUID) |
| `bio` | string \| null |
| `website` | string \| null |
| `phone` | string \| null |
| `location` | string \| null |
| `created_at` | string \| null |
| `updated_at` | string \| null |

---

## Profiles & Admin

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/:id` | Bearer + **super admin** | By UUID |
| GET | `/username/:username` | No | By username |
| GET | `/me` | Bearer | User from token |
| GET | `` | Bearer + **super admin** | User list (paginated); soft-delete filter via query |
| PUT | `/:id` | Bearer + **super admin** | Update a user |
| POST | `/me/image` | Bearer | Upload or replace the caller's avatar |
| DELETE | `/:id` | Bearer + **super admin** | Soft-delete user |
| POST | `/:id/restore` | Bearer + **super admin** | Restore a soft-deleted user |

### GET `/api/users/me`

**Success - 200** - `data`: one `CurrentUserResponse`.

### GET `/api/users` (admin)

**Query**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `limit` | number | 10 | Max 100 |
| `offset` | number | 0 | |
| `deleted` | string | *(empty)* | Soft-delete filter: `false` or empty = active users only; `true` = deleted users only; `all` = all users |

Any `deleted` value other than `true`, `false`, `all`, or empty returns **400 Bad Request**.

**Success - 200** - `data`: `UserResponse[]` (`[]` for an empty page, never `null`), `meta`: pagination.

Examples:

```http
GET /api/users?deleted=true&limit=10&offset=0
GET /api/users?deleted=all
```

### GET `/api/users/:id` (admin)

**Query**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `deleted` | string | *(empty)* | `true` = fetch a soft-deleted user by UUID |

Without `deleted=true`, only **active** users are returned. Deleted users are hidden unless the query above is used.

**Success - 200** - `data`: `UserResponse`. `is_following` is set when the requester is logged in (the admin route already uses auth). `deleted_at` is set when the user has been soft-deleted.

Example:

```http
GET /api/users/<uuid>?deleted=true
```

### GET `/api/users/username/:username`

**Success - 200** - `data`: `UserResponse` (public profile shape).

### PUT `/api/users/:id` (admin)

**Body**

| Field | Type | Required | Behaviour |
|-------|------|----------|-----------|
| `username` | string | Yes | 3-30 chars |
| `email` | string | Yes | Email, max 255 |
| `first_name` | string \| null | No | Max 100. Omitted or `null` stores NULL; a string (even `""`) is stored |
| `last_name` | string \| null | No | Max 100. Omitted or `null` stores NULL; a string (even `""`) is stored |
| `is_super_admin` | boolean \| null | No | Omitted or `null` keeps the stored value; `true`/`false` sets it |

**Success - 200** - `data`: updated `UserResponse`. **404** unknown user, **409** email/username taken, **422** validation.

### POST `/api/users/me/image`

Multipart field `avatar`: JPEG, PNG, GIF or WebP (sniffed from the bytes), max 5 MB.

**Success - 200** - message `Profile picture updated successfully`, `data`: `{ "image": "users/avatars/<key>" }`.

**Errors**

| Status | Condition |
|--------|-----------|
| 400 | No file, unsupported type, or storage not configured |
| 413 | File over 5 MB, or `Content-Length` over 5 MB + 64 KB (checked before parsing). `message` and `error` are both `Avatar is too large (max 5 MB)` |

### DELETE `/api/users/:id` (admin)

Soft-delete a user (sets `deleted_at`; the row is not permanently removed from the database).

**Success - 200** - `data`: `null`.

### POST `/api/users/:id/restore` (admin)

Restore a soft-deleted user (`deleted_at` -> `null`).

**Success - 200** - `data`: reactivated `UserResponse`.

**Errors**

| Status | Condition |
|--------|-----------|
| 404 | User not found or not soft-deleted |
| 409 | Email or username is already used by another active user |

Example:

```http
POST /api/users/<uuid>/restore
```

---

## Follow

| Method | Path | Auth |
|--------|------|------|
| POST | `/follow` | Bearer |
| DELETE | `/:id/follow` | Bearer |
| GET | `/:id/follow-status` | Bearer |
| GET | `/:id/mutual-follows` | Bearer |
| GET | `/:id/followers` | No |
| GET | `/:id/following` | No |
| GET | `/:id/follow-stats` | No |

### POST `/api/users/follow`

**Body**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `user_id` | string | Yes | UUID |

**Success - 200** - `data`:

```json
{
  "is_following": true,
  "message": "Message from service"
}
```

**Errors** (message `Failed to follow user`)

| Status | `error` |
|--------|---------|
| 400 | `cannot follow yourself` |
| 404 | `user not found` (follower or target) |
| 409 | `already following this user` |

### DELETE `/api/users/:id/follow`

Unfollow the user with the UUID in the path.

**Success - 200** - `data`: `FollowResponse` (same shape as follow).

**Errors:** **400** `not following this user` (message `Failed to unfollow user`).

### GET `/api/users/:id/follow-status`

**Success - 200**

```json
{
  "data": { "is_following": false }
}
```

### GET `/api/users/:id/mutual-follows`

**Success - 200** - `data`: `UserResponse[]`: users both the caller and `:id` follow, ordered by `username`, not paginated.

### GET `/api/users/:id/followers` and `/:id/following`

**Query:** `limit` (default 10, max 100), `offset`.

**Success - 200** - `data`: `UserResponse[]`, newest follow first, `meta`: pagination.

Follow lists (including mutual follows) only use live follow rows, so an unfollowed user drops out. Soft-deleted users are excluded from the rows and from `total`. Each item embeds `profile`; `is_following` is never set.

### GET `/api/users/:id/follow-stats`

**Success - 200** - `data`:

```json
{
  "user_id": "uuid",
  "followers_count": 10,
  "following_count": 5
}
```

Counts are computed live from `user_follows` rows that are not soft-deleted. **404** (`error`: `user not found`) when no active user has this ID.

The `followers_count`/`following_count` columns on `users` (used by `UserResponse`) are kept by DB triggers, which since migration `20261029120000` also handle soft unfollows.
