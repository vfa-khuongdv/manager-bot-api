# Bot Dashboard Hub — Health Check API Specification

> **Version:** 1.0.0  
> **Base URL:** `https://api.your-domain.com/v2`  
> **Auth:** All endpoints require the `Authorization: Bearer <admin_token>` header.

---

## Overview

This API provides health check endpoints for monitoring system components:

- **Chatwork API** — External API connectivity and latency
- **Server** — Internal server metrics (CPU, memory, uptime)
- **Database** — Database connection pool and latency

All endpoints return JSON and support the same error format as the main API.

---

## Data Models

### `HealthStatus`

| Field       | Type                               | Description                                   |
| ------------ | ---------------------------------- | --------------------------------------------- |
| `status`     | `"healthy" \| "degraded" \| "unhealthy"` | Overall system health status            |
| `timestamp`  | `string`                           | ISO 8601 datetime of the check                |
| `checks`      | `object`                           | Individual service status                     |

### `ChatworkHealth`

| Field        | Type                              | Description                                |
| ------------- | --------------------------------- | ------------------------------------------ |
| `status`     | `"operational" \| "degraded" \| "down"` | Chatwork API connectivity status |
| `latency`     | `number`                          | Response time in milliseconds              |
| `apiVersion`  | `string`                          | Chatwork API version (e.g. `v2`)            |
| `lastChecked` | `string`                         | ISO 8601 datetime of last check             |

### `ServerHealth`

| Field       | Type                              | Description                                |
| ------------ | --------------------------------- | ------------------------------------------ |
| `status`    | `"healthy" \| "degraded" \| "unhealthy"` | Server health status              |
| `uptime`    | `number`                          | Server uptime in seconds                   |
| `cpu`       | `object`                          | CPU metrics                                 |
| `cpu.usage` | `number`                          | CPU usage percentage (0-100)               |
| `memory`    | `object`                          | Memory metrics                              |
| `memory.used` | `number`                        | Used memory in MB                          |
| `memory.total` | `number`                       | Total memory in MB                        |

### `DatabaseHealth`

| Field       | Type                              | Description                                |
| ------------ | --------------------------------- | ------------------------------------------ |
| `status`    | `"connected" \| "degraded" \| "disconnected"` | Database status              |
| `latency`   | `number`                          | Query latency in milliseconds              |
| `pool`      | `object`                          | Connection pool metrics                    |
| `pool.active` | `number`                         | Number of active connections               |
| `pool.idle` | `number`                          | Number of idle connections                 |
| `pool.total` | `number`                         | Total connection pool size                |

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
| `401` | Unauthorized — missing or invalid token |
| `500` | Internal Server Error |

---

## Endpoints

---

### Health Overview

#### `GET /health`

Returns overall system health status by checking all services.

**Response `200`:**

```json
{
  "status": "healthy",
  "timestamp": "2026-04-02T12:00:00Z",
  "checks": {
    "chatwork": "up",
    "server": "up",
    "database": "up"
  }
}
```

> **Status Logic:**
> - `healthy` — All services are `up`
> - `degraded` — At least one service is `degraded`, none are `down`
> - `unhealthy` — At least one service is `down`

---

### Chatwork API Health

#### `GET /health/chatwork`

Checks Chatwork API connectivity and returns latency metrics.

**Response `200`:**

```json
{
  "status": "operational",
  "latency": 120,
  "apiVersion": "v2",
  "lastChecked": "2026-04-02T12:00:00Z"
}
```

> **Implementation Notes:**
> - Perform a lightweight API call to Chatwork (e.g., `GET /me` or health endpoint)
> - Measure response time for latency
> - Consider rate limiting — cache results for 30-60 seconds

> **Status Logic:**
> - `operational` — API responds with latency < 500ms
> - `degraded` — API responds but latency >= 500ms
> - `down` — API request fails or times out

---

### Server Health

#### `GET /health/server`

Returns server resource metrics.

**Response `200`:**

```json
{
  "status": "healthy",
  "uptime": 86400,
  "cpu": {
    "usage": 45
  },
  "memory": {
    "used": 2048,
    "total": 4096
  }
}
```

> **Implementation Notes:**
> - `uptime` — Use system uptime (e.g., `process.uptime()` in Node.js)
> - `cpu.usage` — Use OS metrics (e.g., `os.loadavg()` or system monitoring library)
> - `memory` — Use `os.totalmem()` and `os.freemem()` to calculate

> **Status Logic:**
> - `healthy` — CPU < 80%, Memory < 80%
> - `degraded` — CPU 80-95%, or Memory 80-95%
> - `unhealthy` — CPU > 95%, or Memory > 95%

---

### Database Health

#### `GET /health/database`

Returns database connection pool status.

**Response `200`:**

```json
{
  "status": "connected",
  "latency": 15,
  "pool": {
    "active": 3,
    "idle": 5,
    "total": 10
  }
}
```

> **Implementation Notes:**
> - Use database driver pool metrics (e.g., `pool.active`, `pool.idle` from pg/Prisma)
> - Execute a simple query (e.g., `SELECT 1`) to measure latency
> - Ensure connection pool is properly configured

> **Status Logic:**
> - `connected` — Database responds, latency < 1000ms, pool not exhausted
> - `degraded` — Latency >= 1000ms, or pool usage > 80%
> - `disconnected` — Database connection failed

---

## Frontend Integration

The frontend polls these endpoints every 30 seconds:

```typescript
// Query keys used by React Query
["health"]              // GET /health
["health", "chatwork"] // GET /health/chatwork
["health", "server"]   // GET /health/server
["health", "database"] // GET /health/database
```

---

## Example Implementation (Node.js/Express)

```javascript
const express = require('express');
const os = require('os');
const axios = require('axios');
const { Pool } = require('pg');

const app = express();
const pool = new Pool({ connectionString: process.env.DATABASE_URL });

// GET /health
app.get('/health', async (req, res) => {
  const [chatwork, server, database] = await Promise.allSettled([
    checkChatwork(),
    checkServer(),
    checkDatabase()
  ]);

  const checks = {
    chatwork: chatwork.status === 'fulfilled' ? chatwork.value.status : 'down',
    server: server.status === 'fulfilled' ? server.value.status : 'down',
    database: database.status === 'fulfilled' ? database.value.status : 'down'
  };

  const statuses = Object.values(checks);
  let overallStatus = 'healthy';
  if (statuses.includes('down')) overallStatus = 'unhealthy';
  else if (statuses.includes('degraded')) overallStatus = 'degraded';

  res.json({
    status: overallStatus,
    timestamp: new Date().toISOString(),
    checks
  });
});

// GET /health/chatwork
async function checkChatwork() {
  const start = Date.now();
  await axios.get('https://api.chatwork.com/v2/me', {
    headers: { 'X-ChatWorkToken': process.env.CHATWORK_API_KEY },
    timeout: 5000
  });
  const latency = Date.now() - start;

  return {
    status: latency < 500 ? 'operational' : 'degraded',
    latency,
    apiVersion: 'v2',
    lastChecked: new Date().toISOString()
  };
}

// GET /health/server
function checkServer() {
  const totalMem = os.totalmem();
  const freeMem = os.freemem();
  const usedMem = (totalMem - freeMem) / (1024 * 1024);
  const totalMemMB = totalMem / (1024 * 1024);
  const memUsage = (usedMem / totalMemMB) * 100;

  const cpus = os.cpus();
  let totalIdle = 0, totalTick = 0;
  cpus.forEach(cpu => {
    for (let type in cpu.times) {
      totalTick += cpu.times[type];
    }
    totalIdle += cpu.times.idle;
  });
  const cpuUsage = 100 - (totalIdle / totalTick * 100);

  let status = 'healthy';
  if (cpuUsage > 95 || memUsage > 95) status = 'unhealthy';
  else if (cpuUsage > 80 || memUsage > 80) status = 'degraded';

  return {
    status,
    uptime: Math.floor(process.uptime()),
    cpu: { usage: Math.round(cpuUsage) },
    memory: { used: Math.round(usedMem), total: Math.round(totalMemMB) }
  };
}

// GET /health/database
async function checkDatabase() {
  const start = Date.now();
  const result = await pool.query('SELECT 1');
  const latency = Date.now() - start;

  const poolMetrics = {
    active: pool.totalCount - pool.idleCount,
    idle: pool.idleCount,
    total: pool.totalCount
  };

  const poolUsage = (poolMetrics.active / poolMetrics.total) * 100;
  let status = 'connected';
  if (latency >= 1000 || poolUsage > 80) status = 'degraded';
  else if (!result.rows.length) status = 'disconnected';

  return { status, latency, pool: poolMetrics };
}
```

---

## Security Notes

1. **Authentication** — All health endpoints require admin token (except may allow unauthenticated for load balancer health checks)
2. **Rate limiting** — Cache results server-side to avoid excessive polling
3. **Sensitive data** — Do not expose detailed error messages or stack traces
4. **HTTPS only** — Enforce TLS for all endpoints
