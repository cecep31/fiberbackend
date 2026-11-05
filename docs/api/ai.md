# AI Module - `/api/ai`

Stateless AI helpers for the post editor. All routes **require** `Authorization: Bearer <token>` and are **super-admin only** for now (rollout; checked against the database). Nothing is read from or written to the database.

## Route Summary

| Method | Path | Description |
|--------|------|-------------|
| POST | `/generate-article` | Generate a draft blog article (SSE) |

---

## Data Types

### `GeneratedArticle`

| Field | Type |
|-------|------|
| `content` | string — the full generated article as HTML (starts with one `<h1>` title) |
| `model` | string — the OpenRouter model that produced it (`OPENROUTER_DEFAULT_MODEL`) |

---

## POST `/api/ai/generate-article`

Streams an AI-written draft article for the create/edit-post page. The output is HTML limited to `<h1>`, `<h2>`, `<h3>`, `<p>`, `<ul>`, `<ol>`, `<li>`, `<strong>`, `<em>`, `<blockquote>`, `<code>`, `<pre>` and `<a href>`, written in the language of the prompt. The client should sanitize it before inserting it into the editor.

**Rate limit:** **20 / hour** per client IP (fixed window, `ai:generate-article`; Valkey-backed when `VALKEY_URL` is set). Exceeding it returns **429**.

**Body** (unknown fields are ignored; strings are trimmed before validation)

| Field | Required | Validation |
|-------|----------|------------|
| `prompt` | Yes | 3-2000 characters — what the article should be about |
| `tone` | No | `informative` (default), `professional`, `casual`, `storytelling` |
| `length` | No | `short` (~400 words), `medium` (default, ~800 words), `long` (~1500 words) |
| `title` | No | max 200 characters — working title; ignored when blank |

There is no `model` or `temperature` field: the configured `OPENROUTER_DEFAULT_MODEL` is always used with temperature `0.7`.

**Response**

- **401** `"User not authenticated"` - no authenticated user.
- **403** `"Insufficient privileges"` - the user is not a super admin.
- **400** - body could not be decoded.
- **422** - validation failed (`errors` lists the fields).
- **503** - `{"success":false,"message":"AI article generation is not available","error":"Service Unavailable"}` when AI is not configured (`OPENROUTER_API_KEY` blank or `OPENROUTER_DEFAULT_MODEL` unset).
- Otherwise **200** - `Content-Type: text/event-stream`. Each frame is `data: <json>\n\n` (no `event:` lines).

SSE event order:

| `type` | `data` |
|--------|--------|
| `ai_chunk` | string (HTML text chunk) |
| `ai_complete` | `GeneratedArticle` |
| `error` | error message (see below) |

There is no initial event: zero or more `ai_chunk` events, then exactly one of `ai_complete` or `error`. The stream **always** ends with `data: [DONE]`, including after an `error` event. `ai_complete.data.content` is the concatenation of every chunk.

`error` data: the upstream provider's message, or `"OpenRouter returned an empty response"` when the stream produced no content. If the client disconnects mid-stream the upstream request is cancelled. The whole upstream request is bounded by `OPENROUTER_TIMEOUT_SECONDS` (default 90).

Example:

```text
data: {"data":"<h1>Why Go","type":"ai_chunk"}

data: {"data":" generics matter</h1>…","type":"ai_chunk"}

data: {"data":{"content":"<h1>Why Go generics matter</h1>…","model":"openrouter/free"},"type":"ai_complete"}

data: [DONE]
```
