# Bot Dashboard Hub — CVE Scanner API Specification

> **Version:** 1.1.0
> **Base URL:** `https://api.your-domain.com/api/v2`
> **Auth:** All endpoints require `X-Project-Key` header or JWT Bearer token.

---

## Overview

This specification covers the **CVE Scanner** module of the Bot Dashboard Hub.

- **CVE Config**: Configuration for scanning repositories for vulnerabilities using OSV (Open Source Vulnerabilities) API.
- **Scans**: Each config can have scheduled or manual scans. Each scan run creates a scan log.
- **Results**: Vulnerabilities are stored and linked to each scan log (not directly to config).

---

## Data Models

### `CveConfig` (DB record)

Stored in the `cve_configs` table.

| Field                  | Type        | DB column               | Description                                               |
| ---------------------- | ----------- | ----------------------- | --------------------------------------------------------- |
| `id`                   | `string`    | `id`                    | UUID primary key                                          |
| `projectId`            | `int`       | `project_id`            | Foreign key to project                                    |
| `name`                 | `string`    | `name`                  | Configuration name                                        |
| `repoUrl`              | `string`    | `repo_url`              | Git repository URL (optional)                             |
| `languages`            | `string`    | `languages`             | Comma-separated libraries: `npm:react@18,PyPI:django@4.2` |
| `cron`                 | `string`    | `cron`                  | Cron expression for scheduled scans                       |
| `status`               | `string`    | `status`                | `"active"` or `"paused"`                                  |
| `apiKey`               | `string?`   | `api_key`               | Chatwork API key for notification (optional)              |
| `botId`                | `number?`   | `bot_id`                | System bot ID for notification (optional)                 |
| `lastScan`             | `datetime?` | `last_scan`             | Timestamp of last scan                                    |
| `lastStatus`           | `string?`   | `last_status`           | `"success"`, `"failed"`, or `"no_scan"`                   |
| `vulnerabilitiesFound` | `number`    | `vulnerabilities_found` | Count of last scan vulnerabilities                        |
| `createdAt`            | `datetime`  | `created_at`            | ISO 8601                                                  |
| `updatedAt`            | `datetime`  | `updated_at`            | ISO 8601                                                  |

---

### `CveScanLog` (DB record)

Stored in the `cve_scan_logs` table. Each scan run creates a new log entry.

| Field            | Type        | DB column          | Description                             |
| ---------------- | ----------- | ------------------ | --------------------------------------- |
| `id`             | `int`       | `id`               | Auto-increment primary key              |
| `configId`       | `string`    | `config_id`        | Foreign key to cve_configs              |
| `projectId`      | `int`       | `project_id`       | Foreign key to project                  |
| `status`         | `string`    | `status`           | `"running"`, `"success"`, or `"failed"` |
| `vulnFoundCount` | `int`       | `vuln_found_count` | Number of vulnerabilities found         |
| `errorMessage`   | `string?`   | `error_message`    | Error message if scan failed            |
| `startedAt`      | `datetime`  | `started_at`       | Scan start timestamp                    |
| `finishedAt`     | `datetime?` | `finished_at`      | Scan finish timestamp                   |
| `createdAt`      | `datetime`  | `created_at`       | Record creation timestamp               |

---

### `Vulnerability` (DB record)

Stored in the `vulnerabilities` table. Linked to each scan log.

| Field       | Type       | DB column     | Description                                 |
| ----------- | ---------- | ------------- | ------------------------------------------- |
| `id`        | `string`   | `id`          | CVE ID (e.g., CVE-2024-1234)                |
| `scanLogId` | `int`      | `scan_log_id` | Foreign key to cve_scan_logs                |
| `configId`  | `string`   | `config_id`   | Foreign key to cve_configs                  |
| `cveId`     | `string`   | `cve_id`      | CVE identifier                              |
| `severity`  | `string`   | `severity`    | `"critical"`, `"high"`, `"medium"`, `"low"` |
| `package`   | `string`   | `package`     | Package name                                |
| `version`   | `string`   | `version`     | Package version                             |
| `summary`   | `string?`  | `summary`     | Vulnerability summary                       |
| `score`     | `number?`  | `score`       | CVSS score (0-10)                           |
| `createdAt` | `datetime` | `created_at`  | Record creation timestamp                   |

---

### `OsvQuery` (OSV API request)

Format for OSV (Open Source Vulnerabilities) API queries.

```json
{
  "queries": [
    {
      "package": { "name": "react", "ecosystem": "npm" },
      "version": "18.2.0"
    },
    {
      "package": { "name": "django", "ecosystem": "PyPI" },
      "version": "4.2.0"
    }
  ]
}
```

| Field     | Type     | Description                          |
| --------- | -------- | ------------------------------------ |
| `package` | `object` | Contains `name` and `ecosystem`      |
| `version` | `string` | Version to check for vulnerabilities |

**Supported Ecosystems:** `npm`, `PyPI`, `Maven`, `Go`, `crates.io`, `NuGet`, `RubyGems`, `Packagist`, `Pub`, `SwiftPM`, etc.

---

### `CveConfigResponse` (API response)

| Field                  | Type      | Description                        |
| ---------------------- | --------- | ---------------------------------- |
| `id`                   | `string`  | Config UUID                        |
| `projectId`            | `number`  | Project ID                         |
| `name`                 | `string`  | Configuration name                 |
| `repoUrl`              | `string`  | Repository URL                     |
| `languages`            | `string`  | Libraries string                   |
| `cron`                 | `string`  | Cron expression                    |
| `status`               | `string`  | `"active"` or `"paused"`           |
| `apiKey`               | `string?` | Chatwork API Key (optional)        |
| `botId`                | `number?` | Managed bot ID (optional)          |
| `notifyOnSuccess`      | `boolean` | Send notification on success       |
| `notifyOnFailure`      | `boolean` | Send notification on failure       |
| `notifyRoomId`         | `string?` | Chatwork Room ID for notifications |
| `notifyOnCritical`     | `boolean` | Notify on Critical severity        |
| `notifyOnHigh`         | `boolean` | Notify on High severity           |
| `notifyOnMedium`       | `boolean` | Notify on Medium severity         |
| `notifyOnLow`          | `boolean` | Notify on Low severity            |
| `lastScan`             | `string?` | Last scan timestamp                |
| `lastStatus`           | `string?` | Last scan status                   |
| `vulnerabilitiesFound` | `number`  | Count of last scan vulnerabilities |
| `createdAt`            | `string`  | Creation timestamp                 |

### `CveScanLogResponse` (API response)

| Field            | Type      | Description                          |
| ---------------- | --------- | ------------------------------------ |
| `id`             | `number`  | Log ID                               |
| `configId`       | `string`  | Config UUID                          |
| `projectId`      | `number`  | Project ID                           |
| `status`         | `string`  | `"running"`, `"success"`, `"failed"` |
| `vulnFoundCount` | `number`  | Vulnerabilities found                |
| `errorMessage`   | `string?` | Error message                        |
| `startedAt`      | `string`  | Scan start timestamp                 |
| `finishedAt`     | `string?` | Scan finish timestamp                |

---

## Endpoints

### CVE Configs

#### `GET /projects/:projectId/cve-configs`

List all CVE configurations for a project.

**Path params:**

| Param       | Type     | Description |
| ----------- | -------- | ----------- |
| `projectId` | `number` | Project ID  |

**Query params:**

| Param   | Type     | Default | Description    |
| ------- | -------- | ------- | -------------- |
| `page`  | `number` | `1`     | Page number    |
| `limit` | `number` | `10`    | Items per page |

**Response `200`:**

```json
{
  "data": [
    {
      "id": "cve-001",
      "projectId": 1,
      "name": "Frontend Core",
      "repoUrl": "https://github.com/company/frontend",
      "languages": "npm:react@18,PyPI:django@4.2",
      "cron": "0 0 * * 1",
      "status": "active",
      "lastScan": "2026-04-14T10:30:00Z",
      "lastStatus": "success",
      "vulnerabilitiesFound": 3,
      "createdAt": "2026-01-15T08:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

---

#### `POST /projects/:projectId/cve-configs`

Create a new CVE configuration.

**Path params:**

| Param       | Type     | Description |
| ----------- | -------- | ----------- |
| `projectId` | `number` | Project ID  |

**Request body:**

| Field             | Type      | Required | Description                                           |
| ----------------- | --------- | -------- | ----------------------------------------------------- |
| `name`            | `string`  | Yes      | Configuration name                                    |
| `repoUrl`         | `string`  | No       | Git repository URL                                    |
| `languages`       | `string`  | Yes      | Libraries format: `ecosystem:package@version,...`     |
| `cron`            | `string`  | Yes      | Cron expression (e.g., `0 0 * * 1`)                   |
| `status`          | `string`  | No       | `"active"` or `"paused"` (default: `"active"`)        |
| `apiKey`          | `string`  | No       | Chatwork API key for notifications                    |
| `botId`           | `number`  | No       | System bot ID for notifications                       |
| `notifyOnSuccess` | `boolean` | No       | Send notification when scan succeeds (default: false) |
| `notifyOnFailure` | `boolean` | No       | Send notification when scan fails (default: true)     |
| `notifyRoomId`    | `string`  | No       | Chatwork Room ID for notifications                    |
| `notifyOnCritical`| `boolean` | No       | Notify on Critical severity (default: true)           |
| `notifyOnHigh`    | `boolean` | No       | Notify on High severity (default: true)              |
| `notifyOnMedium`  | `boolean` | No       | Notify on Medium severity (default: false)           |
| `notifyOnLow`     | `boolean` | No       | Notify on Low severity (default: false)              |

**Example request:**

```json
{
  "name": "Frontend Core",
  "repoUrl": "https://github.com/company/frontend",
  "languages": "npm:react@18.2.0,PyPI:django@4.2.0",
  "cron": "0 0 * * 1",
  "status": "active"
}
```

**Response `201`:**

```json
{
  "id": "cve-002",
  "projectId": 1,
  "name": "Frontend Core",
  "repoUrl": "https://github.com/company/frontend",
  "languages": "npm:react@18.2.0,PyPI:django@4.2.0",
  "cron": "0 0 * * 1",
  "status": "active",
  "lastScan": null,
  "lastStatus": "no_scan",
  "vulnerabilitiesFound": 0,
  "createdAt": "2026-04-15T10:00:00Z"
}
```

---

#### `PUT /projects/:projectId/cve-configs/:configId`

Update an existing CVE configuration.

**Path params:**

| Param       | Type     | Description     |
| ----------- | -------- | --------------- |
| `projectId` | `number` | Project ID      |
| `configId`  | `string` | CVE config UUID |

**Request body:**

| Field       | Type     | Required | Description                                |
| ----------- | -------- | -------- | ------------------------------------------ |
| `name`      | `string` | No       | Configuration name                         |
| `repoUrl`   | `string` | No       | Git repository URL                         |
| `languages` | `string` | No       | Libraries format                           |
| `cron`      | `string` | No       | Cron expression                            |
| `status`    | `string` | No       | `"active"` or `"paused"`                   |
| `apiKey`    | `string` | No       | Chatwork API key (update only if provided) |
| `botId`     | `number` | No       | System bot ID                              |

**Response `200`:**

```json
{
  "id": "cve-002",
  "projectId": 1,
  "name": "Frontend Core - Updated",
  "repoUrl": "https://github.com/company/frontend",
  "languages": "npm:react@18.2.0,PyPI:django@4.2.0",
  "cron": "0 0 * * 0",
  "status": "active",
  "lastScan": "2026-04-15T10:00:00Z",
  "lastStatus": "success",
  "vulnerabilitiesFound": 3,
  "createdAt": "2026-04-15T10:00:00Z"
}
```

---

#### `DELETE /projects/:projectId/cve-configs/:configId`

Delete a CVE configuration.

**Path params:**

| Param       | Type     | Description     |
| ----------- | -------- | --------------- |
| `projectId` | `number` | Project ID      |
| `configId`  | `string` | CVE config UUID |

**Response `204`** — No body.

---

#### `POST /projects/:projectId/cve-configs/:configId/toggle`

Toggle CVE config status between active and paused.

**Path params:**

| Param       | Type     | Description     |
| ----------- | -------- | --------------- |
| `projectId` | `number` | Project ID      |
| `configId`  | `string` | CVE config UUID |

**Response `200`:**

```json
{
  "id": "cve-002",
  "status": "paused"
}
```

---

#### `POST /projects/:projectId/cve-configs/:configId/scan`

Trigger a manual CVE scan for a config.

**Path params:**

| Param       | Type     | Description     |
| ----------- | -------- | --------------- |
| `projectId` | `number` | Project ID      |
| `configId`  | `string` | CVE config UUID |

**Response `200`:**

```json
{
  "message": "Scan triggered successfully"
}
```

**Behavior:**

1. Create a scan log entry (status: "running")
2. Parse `languages` field into OSV query format
3. Call OSV batch API: `POST https://api.osv.dev/v1/querybatch`
4. Store vulnerabilities linked to the scan log
5. Update scan log status to "success" or "failed"
6. Update cve_config `lastScan`, `lastStatus`, `vulnerabilitiesFound`

---

#### `GET /projects/:projectId/cve-configs/:configId/logs`

Get scan logs for a config.

**Path params:**

| Param       | Type     | Description     |
| ----------- | -------- | --------------- |
| `projectId` | `number` | Project ID      |
| `configId`  | `string` | CVE config UUID |

**Query params:**

| Param   | Type     | Default | Description    |
| ------- | -------- | ------- | -------------- |
| `page`  | `number` | `1`     | Page number    |
| `limit` | `number` | `10`    | Items per page |

**Response `200`:**

```json
{
  "data": [
    {
      "id": 1,
      "configId": "cve-002",
      "projectId": 1,
      "status": "success",
      "vulnFoundCount": 3,
      "startedAt": "2026-04-15T10:00:00Z",
      "finishedAt": "2026-04-15T10:05:00Z"
    },
    {
      "id": 2,
      "configId": "cve-002",
      "projectId": 1,
      "status": "failed",
      "vulnFoundCount": 0,
      "errorMessage": "OSV API call failed",
      "startedAt": "2026-04-14T10:00:00Z",
      "finishedAt": "2026-04-14T10:01:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "limit": 10
}
```

---

#### `GET /projects/:projectId/cve-configs/:configId/vulnerabilities`

Get vulnerabilities found in the latest scan for a config.

**Path params:**

| Param       | Type     | Description     |
| ----------- | -------- | --------------- |
| `projectId` | `number` | Project ID      |
| `configId`  | `string` | CVE config UUID |

**Response `200`:**

```json
{
  "data": [
    {
      "id": "CVE-2024-1234",
      "scanLogId": 1,
      "configId": "cve-002",
      "cveId": "CVE-2024-1234",
      "severity": "critical",
      "package": "lodash",
      "version": "4.17.20",
      "summary": "Prototype pollution in lodash",
      "score": 9.8
    },
    {
      "id": "CVE-2024-5678",
      "scanLogId": 1,
      "configId": "cve-002",
      "cveId": "CVE-2024-5678",
      "severity": "high",
      "package": "axios",
      "version": "0.21.0",
      "summary": "Server-Side Request Forgery",
      "score": 8.2
    }
  ],
  "total": 2
}
```

**Note:** This returns vulnerabilities from the most recent scan. Use `/logs` endpoint to get historical scan data.

---

#### `POST /projects/:projectId/cve/test`

Manually test CVE scan with provided packages (no config saved, no database updated).

**Path params:**

| Param       | Type     | Description |
| ----------- | -------- | ----------- |
| `projectId` | `number` | Project ID  |

**Request body:**

| Field       | Type     | Required | Description                                       |
| ----------- | -------- | -------- | ------------------------------------------------- |
| `languages` | `string` | Yes      | Libraries format: `ecosystem:package@version,...` |

**Example request:**

```json
{
  "languages": "npm:react@18.2.0,PyPI:django@4.2.0"
}
```

**Response `200`:**

```json
{
  "data": [
    {
      "id": "CVE-2024-1234",
      "scanLogId": 0,
      "configId": "",
      "cveId": "CVE-2024-1234",
      "severity": "critical",
      "package": "lodash",
      "version": "4.17.20",
      "summary": "Prototype pollution in lodash",
      "score": 9.8
    }
  ],
  "total": 1
}
```

**Note:** This is a test-only endpoint. Results are not saved to the database.

---

#### `GET /projects/:projectId/cve/analysis`

Get analysis for all CVE configs in a project. Returns the latest scan for each config along with all vulnerabilities found.

**Path params:**

| Param       | Type     | Description |
| ----------- | -------- | ----------- |
| `projectId` | `number` | Project ID  |

**Response `200`:**

```json
{
  "data": [
    {
      "configId": "cve-001",
      "configName": "Frontend Core",
      "configStatus": "active",
      "lastScan": "2026-04-15T10:00:00Z",
      "lastStatus": "success",
      "vulnerabilities": [
        {
          "id": "CVE-2024-1234",
          "scanLogId": 1,
          "configId": "cve-001",
          "cveId": "CVE-2024-1234",
          "severity": "critical",
          "package": "lodash",
          "version": "4.17.20",
          "summary": "Prototype pollution in lodash",
          "score": 9.8
        },
        {
          "id": "CVE-2024-5678",
          "scanLogId": 1,
          "configId": "cve-001",
          "cveId": "CVE-2024-5678",
          "severity": "high",
          "package": "axios",
          "version": "0.21.0",
          "summary": "Server-Side Request Forgery",
          "score": 8.2
        }
      ]
    },
    {
      "configId": "cve-002",
      "configName": "Backend API",
      "configStatus": "paused",
      "lastScan": "2026-04-10T10:00:00Z",
      "lastStatus": "no_scan",
      "vulnerabilities": []
    }
  ],
  "total": 2
}
```

**Note:** This endpoint returns a summary of all CVE configs in the project with their latest scan results. Use this for project-level vulnerability dashboard.

---

## OSV API Integration

### Batch Query

**Endpoint:** `POST https://api.osv.dev/v1/querybatch`

**Request body:**

```json
{
  "queries": [
    {
      "package": { "name": "react", "ecosystem": "npm" },
      "version": "18.2.0"
    },
    {
      "package": { "name": "django", "ecosystem": "PyPI" },
      "version": "4.2.0"
    }
  ]
}
```

**Response:**

```json
{
  "results": [
    {
      "vulns": [
        {
          "id": "CVE-2024-1234",
          "summary": "Prototype pollution in lodash",
          "severity": "CRITICAL",
          "details": [
            { "type": "CVSS_V3", "score": 9.8, "severity": "CRITICAL" }
          ],
          "affected": [
            { "package": { "name": "lodash", "ecosystem": "npm" }, "ranges": [...] }
          ],
          "fixed_version": "4.17.21"
        }
      ]
    },
    {}
  ]
}
```

---

## Notification Flow

When vulnerabilities are found, send notification to Chatwork:

1. **Get room ID from config** (via `apiKey` or `botId`)
2. **Format message** with vulnerability summary:
   - Config name
   - Number of vulnerabilities
   - List of CVE IDs with severity
3. **Send via Chatwork API** `POST /v2/rooms/{roomId}/messages`

---

## Error Codes Summary

| Status | Code   | Description               |
| ------ | ------ | ------------------------- |
| `400`  | `4001` | Invalid payload           |
| `404`  | `1001` | Resource not found        |
| `500`  | `2001` | Database operation failed |
| `500`  | `2002` | External API call failed  |
