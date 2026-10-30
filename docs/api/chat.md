# Chat Module - `/api/chat/conversations` & `/api/chat/messages`

Per-user chat conversation CRUD, messages, and streaming AI responses. All routes **require** `Authorization: Bearer <token>`.

## Route Summary

### Conversations - `/api/chat/conversations`

| Method | Path | Description |
|--------|------|-------------|
| POST | `` | Create conversation |
| POST | `/stream` | Create conversation + first message (SSE) |
| GET | `` | List user conversations |
| GET | `/:id` | Conversation detail |
| PUT | `/:id` | Update title / pin |
| DELETE | `/:id` | Delete conversation |
| POST | `/:conversationId/messages` | Send message |
| POST | `/:conversationId/messages/stream` | Send message + stream AI response (SSE) |
| GET | `/:conversationId/messages` | List messages in a conversation |

### Messages - `/api/chat/messages`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/:messageId` | Message detail |
| DELETE | `/:messageId` | Delete message |

---

## Data Types

### `ChatConversationResponse`

| Field | Type |
|-------|------|
| `id` | string (UUID) |
| `title` | string |
| `user_id` | string (UUID) |
| `is_pinned` | boolean |
| `pinned_at` | string \| null |
| `message_count` | number |
| `chat_messages` | `ChatMessageResponse[]` (only on GET `/:id`, where it is always an array — `[]` for an empty conversation; chronological order; key omitted elsewhere) |
| `created_at` | string (ISO) |
| `updated_at` | string (ISO) |

### `ChatMessageResponse`

| Field | Type |
|-------|------|
| `id` | string (UUID) |
| `conversation_id` | string (UUID) |
| `user_id` | string (UUID) |
| `role` | string |
| `content` | string |
| `model` | string \| null |
| `prompt_tokens` | number |
| `completion_tokens` | number |
| `total_tokens` | number |
| `created_at` | string (ISO) |
| `updated_at` | string (ISO) |

---

## POST `/api/chat/conversations`

**Body**

| Field | Required | Validation |
|-------|----------|------------|
| `title` | Yes | max 255 characters |

**Success - 201** - `data`: `ChatConversationResponse`.

---

## POST `/api/chat/conversations/stream`

Create a new conversation and send the first message; the AI response is streamed via SSE when AI is available (`OPENROUTER_API_KEY` non-blank and a model resolves from `model` or `OPENROUTER_DEFAULT_MODEL`).

**Body**

| Field | Required | Validation |
|-------|----------|------------|
| `content` | Yes | 1-10000 characters |
| `title` | No | 1-255 characters |
| `model` | No | max 100 characters |
| `temperature` | No | 0-2 |

**Response**

- If streaming is unavailable: **201** - message `"Message created successfully"`, `data`: array containing the user `ChatMessageResponse`. The created conversation is kept and is not auto-titled.
- If streaming is active: **200** - `Content-Type: text/event-stream`.

SSE event order:

| `type` | `data` |
|--------|--------|
| `conversation_created` | `{ conversation_id, user_message }` |
| `ai_chunk` | string (text chunk) |
| `ai_complete` | `ChatMessageResponse` (AI message) |
| `error` | error message (see below) |

After the first event come zero or more `ai_chunk` events, then exactly one of `ai_complete` or `error`. The stream **always** ends with `data: [DONE]`, including after an `error` event.

`error` data: the upstream provider's message, `"OpenRouter returned an empty response"`, or `"Failed to save AI response"` when the reply could not be stored (internal error details are logged, never sent). If the client disconnects mid-stream the upstream request is cancelled and no partial reply is saved.

---

## GET `/api/chat/conversations`

**Query:** `limit`, `offset` (default limit 10, max 100).

**Success - 200** - `data`: conversation array (always an array, `[]` when empty; newest `updated_at` first; no `chat_messages`), `meta`: pagination.

---

## GET `/api/chat/conversations/:id`

**Success - 200** - `data`: one `ChatConversationResponse` including `chat_messages` (all messages, ascending `created_at`).

Only conversations owned by the token user are returned.

---

## PUT `/api/chat/conversations/:id`

**Body**

| Field | Required | Validation |
|-------|----------|------------|
| `title` | No | max 255; empty or whitespace-only leaves the title unchanged, otherwise stored as sent |
| `is_pinned` | No | boolean |

**Success - 200** - `data`: updated conversation.

---

## DELETE `/api/chat/conversations/:id`

**Success - 200** - `data`: `null`.

---

## POST `/api/chat/conversations/:conversationId/messages`

**Body**

| Field | Required | Validation |
|-------|----------|------------|
| `content` | Yes | 1-10000 characters |
| `role` | No | max 20 characters |
| `model` | No | max 100 characters |
| `temperature` | No | 0-2 |

**Success - 201** - `data`: message array (user + AI response when available).

The message is stored and the conversation's `updated_at` bumped. Only when `role` is `user` and AI is available is the AI reply generated (and a `"New conversation"` auto-titled from its first message); otherwise just the user message is returned.

---

## POST `/api/chat/conversations/:conversationId/messages/stream`

Same as the non-stream endpoint, but the AI response is streamed via SSE. When `role` is not `user` or AI is unavailable it answers **201** `"Message created successfully"` with `data`: `[userMessage]` instead.

SSE events (same ordering and `[DONE]` rule as `POST /stream`):

| `type` | `data` |
|--------|--------|
| `user_message` | `ChatMessageResponse` |
| `ai_chunk` | string |
| `ai_complete` | `ChatMessageResponse` |
| `error` | generic error message |

---

## GET `/api/chat/conversations/:conversationId/messages`

**Success - 200** - `data`: `ChatMessageResponse[]`.

---

## GET `/api/chat/messages/:messageId`

**Success - 200** - `data`: one `ChatMessageResponse`.

404 when the message is not the token user's or its conversation has been deleted; 403 when the conversation belongs to another user.

---

## DELETE `/api/chat/messages/:messageId`

**Success - 200** - `data`: deleted message. Same lookup rules as GET; the conversation's `updated_at` is bumped.

---

## Common Errors

| HTTP | Condition |
|------|-----------|
| 400 | Malformed UUID path parameter: `"Invalid conversation ID"` (`:id`, `:conversationId`) or `"Invalid message ID"` (`:messageId`), checked before the body |
| 404 | Conversation / message not found |
| 403 | Not the conversation owner |
| 422 | Body validation failed |
| 500 | Other server error |
