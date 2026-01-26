# API Specification

## Authentication

All API endpoints require API key authentication. The API key must be included in every request.

### Authentication Methods

#### 1. Header Authentication (Recommended)
```
X-API-Key: your-secret-api-key-here
```

#### 2. Query Parameter Authentication
```
?api_key=your-secret-api-key-here
```

### Response Codes
- `200 OK` - Request successful
- `401 Unauthorized` - Invalid or missing API key
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

## Endpoints

### Health Check

#### GET /healthz
Health check endpoint (no authentication required)

**Response:**
```json
{
  "status": "healthy"
}
```

#### GET /readyz
Readiness check endpoint (no authentication required)

**Response:**
```json
{
  "status": "ready"
}
```

---

### Projects

#### GET /api/v1/projects
Get all projects

**Headers:**
```
X-API-Key: your-secret-api-key-here
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Project Name",
      "created_at": "2026-01-26T00:00:00Z",
      "updated_at": "2026-01-26T00:00:00Z"
    }
  ]
}
```

#### POST /api/v1/projects
Create a new project

**Headers:**
```
X-API-Key: your-secret-api-key-here
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Project Name"
}
```

#### GET /api/v1/projects/:id
Get project by ID

**Headers:**
```
X-API-Key: your-secret-api-key-here
```

#### PATCH /api/v1/projects/:id
Update project

**Headers:**
```
X-API-Key: your-secret-api-key-here
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Updated Project Name"
}
```

#### DELETE /api/v1/projects/:id
Delete project

**Headers:**
```
X-API-Key: your-secret-api-key-here
```

---

### Reminder Schedules

#### POST /api/v1/reminder-schedules
Create a reminder schedule

**Headers:**
```
X-API-Key: your-secret-api-key-here
Content-Type: application/json
```

**Request Body:**
```json
{
  "project_id": 1,
  "cron_expression": "0 9 * * 1-5",
  "chatwork_room_id": "123456789",
  "chatwork_api_token": "your-chatwork-token",
  "message": "Daily reminder message",
  "enabled": true
}
```

#### GET /api/v1/reminder-schedules/:id
Get reminder schedule by ID

**Headers:**
```
X-API-Key: your-secret-api-key-here
```

#### GET /api/v1/projects/:id/reminder-schedules
Get all reminder schedules for a project

**Headers:**
```
X-API-Key: your-secret-api-key-here
```

#### PATCH /api/v1/reminder-schedules/:id
Update reminder schedule

**Headers:**
```
X-API-Key: your-secret-api-key-here
Content-Type: application/json
```

#### DELETE /api/v1/reminder-schedules/:id
Delete reminder schedule

**Headers:**
```
X-API-Key: your-secret-api-key-here
```

---

### Webhooks

#### POST /api/v1/hooks/chatwork
Receive Discord webhooks and forward to Chatwork

**Headers:**
```
X-API-Key: your-secret-api-key-here
Content-Type: application/json
```

**Request Body:**
```json
{
  "embeds": [
    {
      "title": "Deployment Status",
      "description": "Application deployed successfully",
      "fields": [
        {
          "name": "Environment",
          "value": "Production",
          "inline": true
        }
      ],
      "footer": {
        "text": "Discord Notification"
      }
    }
  ]
}
```

#### POST /api/v1/hooks/slack
Receive Slack webhooks and forward to Chatwork

**Headers:**
```
X-API-Key: your-secret-api-key-here
Content-Type: application/json
```

**Request Body (New Format):**
```json
{
  "username": "Coolify",
  "attachments": [
    {
      "color": "#00ff00",
      "title": "Database backup successful",
      "text": "Database backup for mysql-database was successful.\n\n*Frequency:* daily",
      "footer": "Coolify"
    }
  ]
}
```

**Request Body (Old Format):**
```json
{
  "text": "Notification",
  "attachments": [
    {
      "color": "#00ff00",
      "blocks": [
        {
          "type": "section",
          "text": {
            "type": "mrkdwn",
            "text": "*Status:* Success\n*Message:* Deployment completed"
          }
        }
      ]
    }
  ]
}
```

---

### Dashboard

#### GET /api/v1/dashboard
Get dashboard statistics

**Headers:**
```
X-API-Key: your-secret-api-key-here
```

**Response:**
```json
{
  "total_projects": 10,
  "total_schedules": 25,
  "active_schedules": 20,
  "recent_logs": []
}
```

---

## Examples

### cURL Examples

```bash
# Get all projects
curl -H "X-API-Key: your-secret-api-key-here" \
  http://localhost:3000/api/v1/projects

# Create a project
curl -X POST \
  -H "X-API-Key: your-secret-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{"name":"New Project"}' \
  http://localhost:3000/api/v1/projects

# Create a reminder schedule
curl -X POST \
  -H "X-API-Key: your-secret-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": 1,
    "cron_expression": "0 9 * * 1-5",
    "chatwork_room_id": "123456789",
    "chatwork_api_token": "your-token",
    "message": "Daily standup reminder",
    "enabled": true
  }' \
  http://localhost:3000/api/v1/reminder-schedules
```
