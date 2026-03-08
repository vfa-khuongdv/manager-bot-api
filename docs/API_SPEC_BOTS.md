# Bot Dashboard Hub — Bots & Bot Requests API Specification

> **Version:** 2.0.0
> **Base URL:** `https://api.your-domain.com/api/v2`
> **Auth:** All endpoints require the `Authorization: Bearer <admin_token>` header.

---

## Overview

This specification covers the **Bots** and **Bot Requests** modules of the Bot Dashboard Hub.

- **Bots**: Managed catalog of Chatwork bots. Each bot is stored in DB with an `apiToken`; profile data (name, accountId, avatarUrl, rooms count, etc.) is fetched **live** from the Chatwork API at request time.
- **Bot Requests**: Incoming friend requests received by each bot, fetched live from `GET /v2/incoming_requests` on the Chatwork API. No data is persisted in DB for requests.

---

## Data Models

### `ChatworkBot` (DB record — internal)

Stored in the `chatwork_bots` table. The `apiToken` field is **never exposed** in API responses.

| Field         | Type      | DB column     | Description                                    |
| ------------- | --------- | ------------- | ---------------------------------------------- |
| `id`          | `number`  | `id`          | Auto-increment primary key                     |
| `apiToken`    | `string`  | `api_token`   | Chatwork API token — used for all CW API calls |
| `email`       | `string?` | `email`       | Bot contact email (used for room invite flow)  |
| `description` | `string`  | `description` | Custom description managed by admin            |
| `createdAt`   | `string`  | `created_at`  | ISO 8601                                       |
| `updatedAt`   | `string`  | `updated_at`  | ISO 8601                                       |

---

### `BotDetail` (API response)

Returned by `GET /bots`. Combines DB fields with live data from Chatwork `GET /v2/me` and `GET /v2/rooms`.

| Field         | Type      | Source                    | Description                           |
| ------------- | --------- | ------------------------- | ------------------------------------- |
| `id`          | `number`  | DB                        | Internal bot ID                       |
| `accountId`   | `number`  | Chatwork `/me`            | Chatwork Account ID                   |
| `chatworkId`  | `string`  | Chatwork `/me`            | Chatwork handle (e.g. `reminder_bot`) |
| `name`        | `string`  | Chatwork `/me`            | Display name                          |
| `email`       | `string?` | DB                        | Contact email for room invites        |
| `avatarUrl`   | `string`  | Chatwork `/me`            | Avatar image URL                      |
| `description` | `string`  | DB                        | Admin-managed description             |
| `roomsCount`  | `number`  | Chatwork `/rooms` (count) | Number of rooms the bot is in         |

---

### `BotRequestItem` (API response)

Returned by `GET /bot-requests`. Fetched live from Chatwork `GET /v2/incoming_requests` across all bots.

| Field       | Type        | Description                                                                       |
| ----------- | ----------- | --------------------------------------------------------------------------------- |
| `id`        | `string`    | Composite ID: `"{dbBotID}_{cwRequestID}"` — used for accept/delete routing        |
| `botId`     | `number`    | Internal DB bot ID that received this request                                     |
| `botInfo`   | `BotDetail` | Nested bot details                                                                |
| `status`    | `string`    | Always `"pending"` — accepted/rejected requests are removed by Chatwork           |
| `createdAt` | `string`    | ISO 8601 timestamp (set at fetch time — Chatwork API does not return `createdAt`) |

---

## Endpoints

### Bots

#### `GET /bots`

List all bots with live-enriched profile data from Chatwork.

**Query params:**

| Param   | Type     | Default | Description    |
| ------- | -------- | ------- | -------------- |
| `page`  | `number` | `1`     | Page number    |
| `limit` | `number` | `999`   | Items per page |

**Response `200`:**

```json
{
  "data": [
    {
      "id": 1,
      "accountId": 1000001,
      "chatworkId": "reminder_bot",
      "name": "Daily Reminder Bot",
      "email": "bot_reminder_01@chatwork.com",
      "avatarUrl": "https://appdata.chatwork.com/avatar/...",
      "description": "Chuyên gửi thông báo Daily Meeting...",
      "roomsCount": 42
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 999
}
```

**Error responses:**

| Status | Code   | Description           |
| ------ | ------ | --------------------- |
| `500`  | `2001` | Database query failed |

---

#### `POST /bots`

Register a new Chatwork bot into the system.

**Request body:**

| Field         | Type     | Required | Description                    |
| ------------- | -------- | -------- | ------------------------------ |
| `apiToken`    | `string` | Yes      | Chatwork API token             |
| `email`       | `string` | No       | Contact email for room invites |
| `description` | `string` | No       | Custom description for the bot |

**Response `201`:**

Returns the newly created bot with its ID and live profile data.

```json
{
  "id": 4,
  "accountId": 1000004,
  "chatworkId": "new_bot",
  "name": "New Assistant Bot",
  "email": "bot_assistant_04@chatwork.com",
  "avatarUrl": "https://appdata.chatwork.com/avatar/...",
  "description": "Ready to help...",
  "roomsCount": 0
}
```

**Error responses:**

| Status | Code   | Description                  |
| ------ | ------ | ---------------------------- |
| `400`  | `4001` | Invalid API token or payload |
| `500`  | `2001` | Database insert failed       |

---

#### `DELETE /bots/:botId`

Remove a bot from the system. The bot's API token is deleted from DB; no action is taken on the Chatwork side.

**Path params:**

| Param   | Type     | Description          |
| ------- | -------- | -------------------- |
| `botId` | `number` | Internal DB bot ID   |

**Response `204`** — No body.

**Error responses:**

| Status | Code   | Description    |
| ------ | ------ | -------------- |
| `404`  | `1001` | Bot not found  |

---

### Bot Requests

> **Note:** Bot requests are fetched live from the Chatwork API. There is no database table for requests. The `status` is always `"pending"` because Chatwork removes accepted/rejected requests from `GET /v2/incoming_requests` automatically.

#### `GET /bot-requests`

Aggregate all incoming friend requests across every registered bot.

**Query params:**

| Param    | Type     | Description                                               |
| -------- | -------- | --------------------------------------------------------- |
| `status` | `string` | Optional. Only `"pending"` is meaningful (see note above) |

**Response `200`:**

```json
{
  "data": [
    {
      "id": "1_8001",
      "botId": 1,
      "botInfo": {
        "id": 1,
        "accountId": 1000001,
        "chatworkId": "reminder_bot",
        "name": "Daily Reminder Bot",
        "email": "bot_reminder_01@chatwork.com",
        "avatarUrl": "https://appdata.chatwork.com/avatar/...",
        "description": "Chuyên gửi thông báo Daily Meeting...",
        "roomsCount": 42
      },
      "status": "pending",
      "createdAt": "2026-03-08T12:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 1
}
```

**Error responses:**

| Status | Code   | Description           |
| ------ | ------ | --------------------- |
| `500`  | `1000` | Internal server error |

---

#### `POST /bot-requests/:requestId/accept`

Accept a pending friend request. `:requestId` is the composite ID `"{dbBotID}_{cwRequestID}"` from the list response.

**Path params:**

| Param       | Type     | Example    | Description                           |
| ----------- | -------- | ---------- | ------------------------------------- |
| `requestId` | `string` | `"1_8001"` | Composite ID from `GET /bot-requests` |

**Response `200`:**

```json
{
  "success": true,
  "message": "Friend request accepted"
}
```

**Error responses:**

| Status | Code   | Description                                         |
| ------ | ------ | --------------------------------------------------- |
| `400`  | `4001` | Invalid composite ID format or Chatwork API failure |

---

#### `DELETE /bot-requests/:requestId`

Reject/delete a pending friend request.

**Path params:**

| Param       | Type     | Example    | Description                           |
| ----------- | -------- | ---------- | ------------------------------------- |
| `requestId` | `string` | `"1_8001"` | Composite ID from `GET /bot-requests` |

**Response `204`** — No body.

**Error responses:**

| Status | Code   | Description                                         |
| ------ | ------ | --------------------------------------------------- |
| `400`  | `4001` | Invalid composite ID format or Chatwork API failure |

---

## Request ID Format

The `requestId` parameter uses a composite format to avoid storing request data in the database:

```
{dbBotID}_{cwRequestID}
```

| Part          | Description                                          |
| ------------- | ---------------------------------------------------- |
| `dbBotID`     | Internal DB ID of the bot (from `chatwork_bots.id`)  |
| `cwRequestID` | Chatwork's `request_id` from `/v2/incoming_requests` |

**Example:** `"1_8001"` → bot with DB id `1`, Chatwork request id `8001`.

The server parses this to look up the bot's `api_token` and routes the Chatwork API call accordingly.

---

## Flow: Inviting a Bot

1. **Discovery**: User browses `GET /bots` and finds a suitable bot.
2. **Copy email**: User copies the bot's `email` field.
3. **Chatwork side**: User searches for that email on Chatwork and sends a contact request.
4. **Approval**: Admin sees the pending request in `GET /bot-requests` and calls `POST /bot-requests/:requestId/accept`.
5. **Usage**: Once accepted, the bot appears in the user's Chatwork contacts and can be invited to any room.

---

## Architecture Notes

| Concern               | Decision                                                                 |
| --------------------- | ------------------------------------------------------------------------ |
| **DB storage**        | Only `chatwork_bots` table: `api_token`, `email`, `description`          |
| **Profile data**      | Fetched live from `GET /v2/me` per bot (name, accountId, avatarUrl)      |
| **Rooms count**       | Fetched live from `GET /v2/rooms` per bot (count of rooms)               |
| **Incoming requests** | Fetched live from `GET /v2/incoming_requests` per bot, then aggregated   |
| **Accept/Delete**     | Proxied directly to Chatwork `PUT/DELETE /v2/incoming_requests/{id}`     |
| **No request table**  | Accepted/rejected requests vanish from Chatwork API — no need to persist |
