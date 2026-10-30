# Holdings Module - `/api/holdings` & `/api/holding-types`

Investment portfolio records per user per month/year. **All holdings and holding-types routes require a Bearer token.**

## Data Type: `Holding`

| Field | Type | Description |
|-------|------|-------------|
| `id` | number | int64 |
| `user_id` | string (UUID) | |
| `name` | string | |
| `symbol` | string \| null | |
| `platform` | string | |
| `holding_type_id` | number | |
| `holding_type` | object \| null | `HoldingType` |
| `currency` | string | 3 letters, for example `IDR` |
| `invested_amount` | string | Decimal |
| `current_value` | string | Decimal |
| `gain_amount` | string | Read-only (DB) |
| `gain_percent` | string | Read-only |
| `units` | string \| null | |
| `avg_buy_price` | string \| null | |
| `current_price` | string \| null | |
| `last_updated` | string \| null | |
| `notes` | string \| null | |
| `month` | number | 1-12 |
| `year` | number | |
| `created_at` | string | |
| `updated_at` | string | |

### `HoldingType`

```json
{ "id": 1, "code": "STOCK", "name": "Saham", "notes": null }
```

---

## Holdings - Route Summary

| Method | Path | Description |
|--------|------|-------------|
| GET | `` | List + month/year filters |
| GET | `/summary` | Portfolio aggregate |
| GET | `/trends` | Multi-year trends |
| GET | `/compare` | Compare two months |
| GET | `/monthly` | Monthly series |
| GET | `/calendar` | Corporate actions calendar (dividend & RUPS) |
| POST | `` | Create holding |
| POST | `/duplicate` | Copy one month to another |
| POST | `/sync` | Sync prices |
| GET | `/:id` | Detail by ID |
| PUT | `/:id` | Update |
| DELETE | `/:id` | Delete |

## GET `/api/holdings`

**Query**

| Param | Default | Description |
|-------|---------|-------------|
| `month` | Current month | 1-12 |
| `year` | Current year | |
| `sortBy` | `created_at` | `created_at`, `updated_at`, `name`, `platform`, `invested_amount`, `current_value`, `holding_type` |
| `order` | `desc` | `asc` / `desc` |

**Success - 200** - `data`: `Holding[]`.

---

## GET `/api/holdings/summary`

**Query:** `month`, `year` (optional).

**Success - 200** - `data` (`HoldingSummaryResponse`):

| Field | Type |
|-------|------|
| `totalInvested` | string |
| `totalCurrentValue` | string |
| `totalProfitLoss` | string |
| `totalProfitLossPercentage` | string |
| `holdingsCount` | number |
| `typeBreakdown` | breakdown array by type, sorted by type name ascending (`[]` when empty) |
| `platformBreakdown` | breakdown array by platform, sorted by platform ascending (`[]` when empty) |

---

## GET `/api/holdings/trends`

**Query:** `years` - comma-separated years, for example `2024,2025`.

**Success - 200** - `data`: `HoldingTrendResponse[]` (`date`, `invested`, `current`, `profitLoss`, `profitLossPercentage`). `[]` when there are no holdings for those years.

---

## GET `/api/holdings/compare`

**Query:** `fromMonth`, `fromYear`, `toMonth`, `toYear` (default period: previous month -> current month).

**Success - 200** - comparison of `fromMonth` / `toMonth`, `summary`, `typeComparison`, `platformComparison`. `typeComparison` / `platformComparison` are sorted by name ascending and are `[]` when both months are empty; `summary.from` / `summary.to` breakdowns follow the summary ordering.

---

## GET `/api/holdings/monthly`

**Query**

| Param | Default | Description |
|-------|---------|-------------|
| `startMonth`, `startYear` | Current month/year | One endpoint of the desired range |
| `endMonth`, `endYear` | 11 months before start | The other endpoint of the desired range when omitted |

The two endpoints can be passed in either chronological order. The handler normalizes them internally so that `start` becomes the latest month and `end` the oldest month in the returned series. If `end` is omitted, it defaults to 11 months before `start` (returning a 12-month series). If only one `end` component is given, the missing one is copied from `start`. Months outside 1-12 (or not integers) are ignored.

**Success - 200** - `data`: `HoldingMonthlyDataResponse[]`.

| HTTP | Condition |
|------|-----------|
| 400 | `startYear` / `endYear` is not an integer or is outside 1990 to current year + 1 - message `Invalid monthly range`, error `end month/year must be on or before start month/year` |
| 400 | Resolved range spans more than 240 months (20 years) - message `Failed to get monthly data`, error `requested month/year range is too large` |

---

## GET `/api/holdings/calendar`

Corporate actions calendar (dividend & RUPS events) for all stocks in the IDX exchange, for a single calendar month. Requires a Bearer token. The data is market-wide (not personalized); every authenticated user receives the same calendar.

Results are persisted in Postgres (`corporate_actions` table). If the requested month already has stored rows, they are served directly from the database. Otherwise the IDX API is queried, the results are upserted into Postgres, and the stored rows are returned. The `cached` field is `true` when the response was served straight from Postgres without calling IDX for this request.

**Query Parameters**

| Param | Default | Description |
|-------|---------|-------------|
| `month` | Current month | 1-12 |
| `year` | Current year | |

**Success - 200** - `data` (`CorporateActionCalendarResponse`):

| Field | Type | Description |
|-------|------|-------------|
| `from` | string | First day of the requested month (`YYYY-MM-DD`) |
| `to` | string | Last day of the requested month (`YYYY-MM-DD`) |
| `actions` | array | List of corporate action items (sorted chronologically) |
| `total` | number | Total count of actions |
| `cached` | boolean | True if the response was retrieved from Postgres without a fresh IDX fetch |

Each item in `actions` has the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `symbol` | string | Stock ticker symbol (e.g. `BBCA`) |
| `name` | string | Company name |
| `type` | string | Event type: `dividend` or `rups` |
| `date` | string | Ex-date for dividends or meeting date for RUPS (`YYYY-MM-DD`) |
| `pay_date` | string \| null | Dividend payment date (null for RUPS) |
| `amount` | number \| null | Gross dividend amount per share (null for RUPS) |
| `currency` | string | Currency code (e.g. `IDR`) |
| `note` | string | Meeting agenda for RUPS or comments |
| `market` | string | Exchange identifier (`IDX`) |

---

## POST `/api/holdings`

**Body (`CreateHoldingRequest`)**

| Field | Required | Validation |
|-------|----------|------------|
| `name` | Yes | |
| `platform` | Yes | |
| `holding_type_id` | Yes | |
| `currency` | Yes | length 3 |
| `invested_amount` | Yes | decimal string |
| `current_value` | Yes | decimal string |
| `month` | Yes | 1-12 |
| `year` | Yes | min 2000 |
| `units`, `avg_buy_price`, `current_price` | No | decimal string when present |
| `symbol`, `last_updated`, `notes` | No | |

Decimal strings are trimmed and must be a plain decimal (optional sign, digits, optional fractional part); exponents, `NaN` and `Infinity` are rejected. The trimmed value is stored.

**Success - 201** - `data`: `Holding[]` (one element).

| HTTP | Condition |
|------|-----------|
| 400 | Unknown `holding_type_id` - error `holding type not found` |
| 400 | Invalid decimal - message `Failed to create holding`, error `Invalid decimal value for <field>` (fields checked in order `invested_amount`, `current_value`, `units`, `avg_buy_price`, `current_price`; first failure wins) |

---

## POST `/api/holdings/duplicate`

**Body**

| Field | Required |
|-------|----------|
| `fromMonth`, `fromYear`, `toMonth`, `toYear` | Yes (1-12 / 1900-2100) |
| `overwrite` | boolean |

**Success - 201** - `data`: `DuplicateResultItem[]` (`id`, `name`, `month`, `year`).

**Error 400** when source month = destination month.

---

## POST `/api/holdings/sync`

Sync prices for the user's active period.

**Success - 200** - `data`: `{ "syncedCount", "month", "year" }`.

---

## GET `/api/holdings/:id`

**Path:** numeric `id`.

| HTTP | Condition |
|------|-----------|
| 404 | Holding not found |
| 403 | Not the owner |

---

## PUT `/api/holdings/:id`

**Body:** `UpdateHoldingRequest` - optional fields; the handler does not always call global body validation.

Amount fields that are present follow the same decimal rules as create; a failure returns **400** with message `Failed to update holding` and error `Invalid decimal value for <field>`.

**Success - 200** - `data`: `[Holding]`.

---

## DELETE `/api/holdings/:id`

**Success - 200** - `data`: `null`.

---

## GET `/api/holding-types`

**Success - 200** - `data`: `HoldingType[]`.

Asset type list (stocks, mutual funds, etc.) for form dropdowns.
