# Bot Dashboard Hub — Backend API Specification

> **Version:** 1.0.0  
> **Base URL:** `https://api.your-domain.com/v2`  
> **Auth:** All endpoints (except `/auth/login`) require the `Authorization: Bearer <admin_token>` header.

---

## Overview

This API powers the **Bot Dashboard Hub** — a management panel for scheduling automated bot messages (e.g. to Chatwork rooms). Core features:

- **Auth** — Passcode-based gateway, returns a session token
- **Projects** — Group of schedules; protected by a per-project secret key
- **Schedules** — Cron-based message dispatch tasks linked to a project
- **Run Logs** — Execution history for every schedule run

---

## Data Models

### `Project`

| Field            | Type                     | Description                                     |
| ---------------- | ------------------------ | ----------------------------------------------- |
| `id`             | `string`                 | Unique identifier (e.g. `p1`)                   |
| `name`           | `string`                 | Human-readable project name                     |
| `description`    | `string`                 | Short description                               |
| `status`         | `"active" \| "inactive"` | Whether the project is active                   |
| `secretKey`      | `string`                 | Auto-generated key, format: `sk_proj_<12chars>` |
| `createdAt`      | `string`                 | ISO 8601 date (e.g. `2025-12-01`)               |
| `schedulesCount` | `number`                 | Count of schedules in this project              |

### `Schedule`

| Field         | Type                            | Description                                                                                                    |
| ------------- | ------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `id`          | `string`                        | Unique identifier (e.g. `s1`)                                                                                  |
| `projectId`   | `string`                        | Reference to parent `Project.id`                                                                               |
| `name`        | `string`                        | Human-readable name for schedule                                                                               |
| `projectName` | `string`                        | Denormalized project name for display                                                                          |
| `roomId`      | `string`                        | Chatwork room ID to message                                                                                    |
| `apiKey`      | `string`                        | Chatwork API key (write-only; masked in GET responses as `cwk_***hidden***`). Mutually exclusive with `botId`. |
| `botId`       | `number \| null`                | Link to a managed `ChatworkBot.id`. If set, the managed bot's token is used.                                   |
| `cron`        | `string`                        | Cron expression (e.g. `0 2 * * 1-5`)                                                                           |
| `message`     | `string`                        | Message body (supports Chatwork markup: `[info]`, `[title]`, `[code]`)                                         |
| `status`      | `"active" \| "paused"`          | Whether this schedule is running                                                                               |
| `lastRun`     | `string \| null`                | ISO 8601 datetime of last execution                                                                            |
| `lastStatus`  | `"success" \| "failed" \| null` | Result of the last run                                                                                         |
| `createdAt`   | `string`                        | ISO 8601 date                                                                                                  |

### `RunLog`

| Field         | Type                    | Description                   |
| ------------- | ----------------------- | ----------------------------- |
| `id`          | `string`                | Unique identifier (e.g. `r1`) |
| `scheduleId`  | `string`                | Reference to `Schedule.id`    |
| `projectName` | `string`                | Denormalized project name     |
| `status`      | `"success" \| "failed"` | Result of this run            |
| `timestamp`   | `string`                | ISO 8601 datetime             |
| `message`     | `string`                | Human-readable result message |

---

## Error Format

All errors return a consistent JSON body:

```json
{
  "error": "short_error_code",
  "message": "Human-readable description"
}
```

Common HTTP status codes:
| Code | Meaning |
|---|---|
| `200` | OK |
| `201` | Created |
| `204` | No Content (successful delete) |
| `400` | Bad Request — validation failed |
| `401` | Unauthorized — missing or invalid token |
| `403` | Forbidden — secret key mismatch |
| `404` | Not Found |
| `500` | Internal Server Error |

---

## Pagination

List endpoints support optional query params:
| Param | Default | Description |
|---|---|---|
| `page` | `1` | Page number (1-indexed) |
| `limit` | `20` | Items per page (max `100`) |

List response wrapper:

```json
{
  "data": [...],
  "total": 42,
  "page": 1,
  "limit": 20
}
```

---

## Endpoints

---

### Auth

#### `POST /auth/login`

Authenticate with the admin passcode. Returns a session token.

**Request body:**

```json
{
  "passcode": "botadmin2026"
}
```

**Response `200`:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": "2026-03-07T11:00:00Z"
}
```

**Errors:** `401` — Invalid passcode

---

#### `POST /auth/logout`

Invalidate the current session token.

**Response `204`** — No body.

---

### Projects

#### `GET /projects`

List all projects.

**Query params:** `page`, `limit`, `status` (`active | inactive`)

**Response `200`:**

```json
{
  "data": [
    {
      "id": "p1",
      "name": "Daily Standup Reminder",
      "description": "Send daily standup reminder to dev team",
      "status": "active",
      "secretKey": "sk_proj_a1b2c3d4e5f6",
      "createdAt": "2025-12-01",
      "schedulesCount": 3
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 20
}
```

---

#### `POST /projects`

Create a new project. `secretKey` is generated server-side.

**Request body:**

```json
{
  "name": "My Bot Project",
  "description": "Optional description",
  "status": "active"
}
```

**Response `201`:** Full `Project` object including generated `secretKey`.

**Errors:** `400` — `name` is required

---

#### `GET /projects/:projectId`

Get a single project by ID.

**Response `200`:** Single `Project` object.

**Errors:** `404` — Project not found

---

#### `PATCH /projects/:projectId`

Update project metadata. Cannot change `secretKey` via this endpoint.

**Request body (all fields optional):**

```json
{
  "name": "Updated Name",
  "description": "New description",
  "status": "inactive"
}
```

**Response `200`:** Updated `Project` object.

---

#### `DELETE /projects/:projectId`

Delete a project and all its schedules.

**Response `204`** — No body.

---

#### `POST /projects/:projectId/access`

Validate a project's secret key. Used by users to "enter" a project.

**Request body:**

```json
{
  "secretKey": "sk_proj_a1b2c3d4e5f6"
}
```

**Response `200`:**

```json
{
  "projectId": "p1",
  "name": "Daily Standup Reminder",
  "granted": true
}
```

**Errors:** `403` — Invalid secret key

---

### Schedules

> Requests may use `Authorization: Bearer <admin_token>` **or** `X-Project-Key: sk_proj_<key>` to scope access to a single project.

#### `GET /projects/:projectId/schedules`

List all schedules for a project.

**Query params:** `page`, `limit`, `status` (`active | paused`)

**Response `200`:**

```json
{
  "data": [
    {
      "id": "s1",
      "projectId": "p1",
      "name": "Daily Standup Reminder",
      "projectName": "Daily Standup Reminder",
      "roomId": "123456",
      "apiKey": "cwk_***hidden***",
      "botId": null,
      "cron": "0 2 * * 1-5",
      "message": "[info][title]🤖 Daily Reminder[/title]...[/info]",
      "status": "active",
      "lastRun": "2026-03-04T02:00:00Z",
      "lastStatus": "success",
      "createdAt": "2025-12-02"
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 20
}
```

> `apiKey` is always masked in GET responses.

---

#### `POST /projects/:projectId/schedules`

Create a new schedule.

**Request body:**

```json
{
  "name": "Morning Standup",
  "roomId": "123456",
  "apiKey": "cwk_xxxxxxxxxxxx", // optional if botId is provided
  "botId": 1, // optional if apiKey is provided
  "cron": "0 2 * * *",
  "message": "[info][title]🤖 Reminder[/title]Your daily update.[/info]",
  "status": "active"
}
```

**Response `201`:** Full `Schedule` object (with `apiKey` masked).

**Errors:** `400` — `name`, `roomId`, `apiKey`, or `cron` is missing/invalid

---

#### `POST /projects/:projectId/schedules/test`

Send a test message to Chatwork using the provided parameters (used by frontend to test before saving, or test an existing schedule).

**Request body:**

```json
{
  "roomId": "123456",
  "apiKey": "cwk_xxxxxxxxxxxx",
  "message": "[info][title]🤖 Test Message[/title]This is a test![/info]",
  "scheduleId": 12
}
```

> **Note**: If testing an existing schedule where `apiKey` is masked on the frontend (`cwk_***hidden***`), you **must** supply `scheduleId`. The backend will securely resolve the real API key from the database.

**Response `200`:**

```json
{
  "success": true,
  "message": "Test message sent successfully"
}
```

**Errors:**

- `400` — Validation failed (missing fields, or `apiKey` masked without `scheduleId`).
- `404` — `scheduleId` provided but not found.
- `502` — Failed to send message to Chatwork API.

---

#### `GET /projects/:projectId/schedules/:scheduleId`

Get a single schedule.

**Response `200`:** Single `Schedule` object (with `apiKey` masked).

---

#### `PATCH /projects/:projectId/schedules/:scheduleId`

Update a schedule. Omitting `apiKey` leaves the existing key unchanged.

**Request body (all fields optional):**

```json
{
  "name": "Updated name",
  "roomId": "654321",
  "apiKey": "cwk_newkey",
  "botId": 2, // can be set to null to switch back to apiKey
  "cron": "0 3 * * *",
  "message": "Updated message",
  "status": "paused"
}
```

**Response `200`:** Updated `Schedule` object.

---

#### `PATCH /projects/:projectId/schedules/:scheduleId/toggle`

Toggle schedule status between `active` ↔ `paused`.

**Response `200`:**

```json
{
  "id": "s1",
  "status": "paused"
}
```

---

#### `DELETE /projects/:projectId/schedules/:scheduleId`

Delete a schedule.

**Response `204`** — No body.

---

### Run Logs

#### `GET /run-logs`

List all run logs across all projects (admin only).

**Query params:** `page`, `limit`, `status` (`success | failed`), `projectId`, `scheduleId`, `from` (ISO date), `to` (ISO date)

**Response `200`:**

```json
{
  "data": [
    {
      "id": "r1",
      "scheduleId": "s1",
      "projectName": "Daily Standup Reminder",
      "status": "success",
      "timestamp": "2026-03-04T02:00:12Z",
      "message": "Message sent to room 123456"
    }
  ],
  "total": 12,
  "page": 1,
  "limit": 20
}
```

---

#### `GET /projects/:projectId/run-logs`

List run logs scoped to a specific project.

**Query params:** Same as above, minus `projectId`.

---

#### `GET /projects/:projectId/schedules/:scheduleId/run-logs`

List run logs for a specific schedule.

---

### Dashboard

#### `GET /dashboard/summary`

Returns aggregated stats for the dashboard including schedule runs and CVE scanner data.

**Response `200`:**

```json
{
  "activeProjects": 3,
  "inactiveProjects": 2,
  "totalSchedules": 8,
  "activeSchedules": 6,
  "successRuns": 9,
  "failedRuns": 3,
  "successRate": 75,
  "totalCveConfigs": 24,
  "activeCveMonitoring": 18,
  "totalVulnerabilities": 156,
  "secureConfigs": 8,
  "criticalVulns": 12,
  "highVulns": 34,
  "moderateVulns": 67,
  "lowVulns": 43
}
```

**Response Fields:**

| Field                   | Type      | Description                           |
| ----------------------  | --------- | ------------------------------------ |
| `activeProjects`       | `int`     | Number of active projects            |
| `inactiveProjects`     | `int`     | Number of inactive projects          |
| `totalSchedules`        | `int`     | Total schedule configs               |
| `activeSchedules`      | `int`     | Number of active schedules           |
| `successRuns`          | `int`     | Total successful schedule runs       |
| `failedRuns`           | `int`     | Total failed schedule runs           |
| `successRate`          | `float`   | Success rate percentage              |
| `totalCveConfigs`      | `int`     | Total CVE scanner configs (all projects) |
| `activeCveMonitoring`  | `int`     | Number of active CVE configs         |
| `totalVulnerabilities` | `int`     | Sum of all vulnerabilities found     |
| `secureConfigs`        | `int`     | CVE configs with 0 vulnerabilities   |
| `criticalVulns`       | `int`     | Critical severity vulnerabilities     |
| `highVulns`           | `int`     | High severity vulnerabilities        |
| `moderateVulns`       | `int`     | Moderate severity vulnerabilities      |
| `lowVulns`            | `int`     | Low severity vulnerabilities        |

---

#### `GET /dashboard/cve-recent-scans`

Get recent CVE scan results across all projects. Returns last 10 scans.

**Query params:**

| Param    | Type    | Default | Description |
| -------- | ------- | ------- | ------------|
| `limit`  | `int`   | 10      | Number of results (max 50) |

**Response `200`:**

```json
{
  "data": [
    {
      "id": 1,
      "configName": "Frontend Core",
      "configId": "uuid-123",
      "projectId": 1,
      "projectName": "Web App",
      "lastScan": "2026-04-17T10:30:00Z",
      "vulnCount": 3,
      "status": "failed"
    },
    {
      "id": 2,
      "configName": "Backend API",
      "configId": "uuid-456",
      "projectId": 1,
      "projectName": "Web App",
      "lastScan": "2026-04-17T09:00:00Z",
      "vulnCount": 0,
      "status": "success"
    }
  ],
  "total": 24
}
```

**Response Fields:**

| Field          | Type      | Description                          |
| -------------- | --------- | ------------------------------------ |
| `id`           | `int`     | Scan log ID (auto-increment)         |
| `configName`  | `string`  | CVE config name                      |
| `configId`     | `string`  | CVE config UUID                      |
| `projectId`    | `int`     | Project ID                           |
| `projectName`  | `string`  | Project name                         |
| `lastScan`     | `datetime`| Last scan timestamp                  |
| `vulnCount`    | `int`     | Number of vulnerabilities found     |
| `status`       | `string`  | `"success"` or `"failed"`            |

---

#### `GET /projects/:projectId/schedules/analysis`

Get analysis for all schedules in a project. Returns aggregated stats for each schedule.

**Path params:**

| Param       | Type     | Description |
| ----------- | -------- | ----------- |
| `projectId` | `number` | Project ID  |

**Response `200`:**

```json
{
  "data": [
    {
      "scheduleId": 1,
      "scheduleName": "Daily Standup",
      "status": "active",
      "lastRun": "2026-04-15T02:00:00Z",
      "lastStatus": "success",
      "totalRuns": 30,
      "successRuns": 28,
      "failedRuns": 2
    },
    {
      "scheduleId": 2,
      "scheduleName": "Weekly Report",
      "status": "paused",
      "lastRun": "2026-04-10T09:00:00Z",
      "lastStatus": "failed",
      "totalRuns": 15,
      "successRuns": 12,
      "failedRuns": 3
    }
  ],
  "total": 2
}
```

---

## Security Notes

1. **Admin token** — Issued on `/auth/login` with the gate passcode. Use JWT with short TTL (recommend 8h).
2. **Project secret key** — Per-project credential (`sk_proj_...`). Allows project-scoped access without admin privileges.
3. **Chatwork API key** — Stored encrypted at rest; never returned in plaintext in any GET response.
4. **HTTPS only** — Enforce TLS for all endpoints.
