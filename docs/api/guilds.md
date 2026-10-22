# Guilds Module - `/api/guilds`

Discord-style guilds: channels, messages, and a realtime SSE stream per guild. Deleting a guild **cascades** to its memberships, channels, and messages (hard delete, no undo).

## Route Summary

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `` | No | List public guilds |
| POST | `` | User | Create guild (creator becomes owner/member) |
| GET | `/me` | User | Guilds the caller belongs to |
| GET | `/:slug` | Optional | Guild detail (`is_member`/`my_role` when authenticated) |
| PATCH | `/:slug` | Owner | Update guild |
| DELETE | `/:slug` | Owner | Delete guild (cascade) |
| POST | `/:slug/join` | User | Join guild |
| DELETE | `/:slug/leave` | User | Leave guild |
| GET | `/:slug/members` | Optional | List members |
| GET | `/:slug/channels` | Optional | List channels |
| POST | `/:slug/channels` | Member | Create channel |
| PATCH | `/:slug/channels/:channelId` | Member | Update channel |
| DELETE | `/:slug/channels/:channelId` | Member | Delete channel |
| GET | `/:slug/channels/:channelId/messages` | User | List messages (cursor pagination) |
| POST | `/:slug/channels/:channelId/messages` | Member | Send message |
| PATCH | `/:slug/channels/:channelId/messages/:messageId` | Author | Edit message |
| DELETE | `/:slug/channels/:channelId/messages/:messageId` | Author or owner | Delete message |
| GET | `/:slug/events` | User (member) | SSE realtime stream |

## Data Types

### `GuildResponse`

`id`, `name`, `slug`, `description`, `avatar_url`, `is_public`, `member_count`, `owner_id`, `owner` (`UserBrief`), and when authenticated: `is_member`, `my_role` (`owner`/`admin`/`member`); plus `created_at`, `updated_at`.

### `GuildMemberResponse`

`id`, `guild_id`, `user_id`, `role`, `user` (`UserBrief`), `created_at`.

### `GuildChannelResponse`

`id`, `guild_id`, `name`, `topic`, `position`, `created_by`, `created_at`, `updated_at`.

### `GuildMessageResponse`

`id`, `guild_id`, `channel_id`, `author_id`, `author` (`UserBrief`), `content` (max 4000 chars), `reply_to_id`, `reply_to` (`{ id, author_id, author, content }`), `edited_at`, `created_at`.

## Request Bodies

| Request | Fields |
|---------|--------|
| Create guild | `name` (3–100, required), `slug` (3–100), `description` (≤2000), `avatar_url` (url, ≤2048), `is_public` |
| Update guild | `name`, `description`, `avatar_url`, `is_public` (all optional) |
| Create channel | `name` (1–100, required), `topic` (≤1024), `position` (≥0) |
| Update channel | `name`, `topic`, `position` (all optional) |
| Send / edit message | `content` (required, ≤4000); send also accepts `reply_to_id` (uuid) |

## Message Pagination

`GET /:slug/channels/:channelId/messages?limit=&before=` — **cursor, not offset**. `limit` default 50, max 100. Pass the oldest loaded message's id as `before` to page further back. `meta` is `{ "limit", "has_more", "next_before"? }`.

## GET `/:slug/events` (SSE)

Members-only event stream (`text/event-stream`). The server sends `: connected` on attach and a `: ping` heartbeat every 25s (used to re-check membership — losing membership ends the stream).

Event frames:

```
data: {"type":"message.created","guild_id":"...","channel_id":"...","data":{ /* GuildMessageResponse */ }}

```

Event `type` values: `channel.created`, `channel.updated` (`data` = `GuildChannelResponse`), `channel.deleted` (`data` = `{ "id" }`), `message.created`, `message.updated` (`data` = `GuildMessageResponse`), `message.deleted` (`data` = `{ "id", "channel_id" }`). Events fan out through Redis pub/sub when configured, otherwise in-process.

## Errors

| Status | When |
|--------|------|
| 401 | Not authenticated (member-only routes without token) |
| 403 | Not a guild member (send/edit/delete) or not the author/owner |
| 404 | Guild, channel, or message not found (hidden from non-members) |
| 422 | Validation failure |
