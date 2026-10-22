# Stats Module - Admin Statistics

Admin statistics endpoints that replaced the legacy `/api/reports` module. **All routes require a Bearer token and super admin access.**

## Route Summary

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/posts/stats` | Post statistics + top posts |
| GET | `/api/posts/stats/engagement` | Engagement metrics |
| GET | `/api/tags/stats` | Per-tag statistics |
| GET | `/api/users/stats` | User statistics + growth trend |

---

## Query Parameters

Optional date filters use `startDate` and `endDate` in **`YYYY-MM-DD`** format.

| Endpoint | Params |
|----------|--------|
| GET `/api/posts/stats` | `startDate`, `endDate`, `limit` (default 10, max 100), `tagId` (numeric tag id) |
| GET `/api/posts/stats/engagement` | `startDate`, `endDate` |
| GET `/api/tags/stats` | `limit` (default 10, max 100) |
| GET `/api/users/stats` | `startDate`, `endDate`, `limit` (default 10, max 100) |

---

## GET `/api/posts/stats`

**Success - 200** - `data` (`PostStatsResponse`):

| Field | Type |
|-------|------|
| `totalPosts` | number |
| `newPostsThisPeriod` | number |
| `totalViews` | number |
| `totalLikes` | number |
| `totalComments` | number |
| `newPostsToday` | number |
| `avgEngagementRate` | number |
| `topPosts` | `PostPerformanceData[]` |

`PostPerformanceData`: `id`, `title`, `slug`, `views`, `likes`, `comments`, `engagementRate`, `author` (`id`, `username`, `firstName`, `lastName`), `createdAt`.

## GET `/api/posts/stats/engagement`

**Success - 200** - `data` (`EngagementMetricsResponse`):

| Field | Type |
|-------|------|
| `totalEngagements` | number |
| `avgLikesPerPost` | number |
| `avgCommentsPerPost` | number |
| `avgViewsPerPost` | number |
| `periodComparison` | `{ "current": number, "previous": number, "changePercent": number }` |

## GET `/api/tags/stats`

**Success - 200** - `data`: array of `{ "id": number, "name": string, "postCount": number, "totalViews": number, "totalLikes": number }` (ordered by post activity).

## GET `/api/users/stats`

**Success - 200** - `data` (`UserStatsResponse`):

| Field | Type |
|-------|------|
| `totalUsers` | number |
| `newUsersThisPeriod` | number |
| `activeUsers` | number |
| `newUsersToday` | number |
| `activeUsersThisWeek` | number |
| `topContributors` | `{ "id", "username", "firstName", "lastName", "postCount", "totalViews", "totalLikes" }[]` |
| `growthTrend` | `{ "date": "YYYY-MM-DD", "newUsers": number, "cumulativeUsers": number }[]` |
