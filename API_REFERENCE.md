# Complete API Reference

## Base URL
```
http://localhost:8080
```

## Authentication
Currently, the API does not implement authentication. In production, you should add JWT or OAuth2.

---

## Health Check

### Check System Health
```http
GET /health
```

**Response:**
- `200 OK` - All systems operational
- `503 Service Unavailable` - Database connection issues

---

## Organizations

### Create Organization
```http
POST /api/v1/orgs
Content-Type: application/json

{
  "id": "acme-corp",
  "name": "Acme Corporation"
}
```

### Get All Organizations
```http
GET /api/v1/orgs
```

---

## Groups

### Create Group
```http
POST /api/v1/orgs/{orgId}/groups
Content-Type: application/json

{
  "id": "engineering",
  "name": "Engineering Team"
}
```

### Get Organization Groups
```http
GET /api/v1/orgs/{orgId}/groups
```

---

## WebSocket

### Join Group (WebSocket)
```
ws://localhost:8080/ws/orgs/{orgId}/groups/{groupId}?clientId={clientId}
```

**Query Parameters:**
- `clientId` (required) - Unique identifier for the client

**Message Format:**
```json
{
  "content": "Hello, World!"
}
```

---

## Messaging

### Broadcast to Organization
```http
POST /api/v1/orgs/{orgId}/broadcast
Content-Type: application/json

{
  "content": "Company-wide announcement"
}
```

### Broadcast to Group
```http
POST /api/v1/orgs/{orgId}/groups/{groupId}/broadcast
Content-Type: application/json

{
  "content": "Team message",
  "client_id": "user-123"
}
```

### Get Message History
```http
GET /api/v1/orgs/{orgId}/groups/{groupId}/messages?limit=50
```

**Query Parameters:**
- `limit` (optional, default: 50) - Number of messages to retrieve

**Response:**
```json
{
  "messages": [
    {
      "id": "msg-uuid",
      "org_id": "acme-corp",
      "group_id": "engineering",
      "client_id": "user-123",
      "username": "john_doe",
      "content": "Hello!",
      "timestamp": "2025-12-01T10:30:00Z"
    }
  ],
  "count": 1
}
```

### Get Messages After Timestamp
```http
GET /api/v1/orgs/{orgId}/groups/{groupId}/messages/after?after=1733054400&limit=50
```

**Query Parameters:**
- `after` (required) - Unix timestamp
- `limit` (optional, default: 50)

### Get Messages Between Timestamps
```http
GET /api/v1/orgs/{orgId}/groups/{groupId}/messages/between?start=1733054400&end=1733140800&limit=50
```

**Query Parameters:**
- `start` (required) - Start Unix timestamp
- `end` (required) - End Unix timestamp
- `limit` (optional, default: 50)

### Get Message Count
```http
GET /api/v1/orgs/{orgId}/groups/{groupId}/messages/count
```

**Response:**
```json
{
  "count": 1523
}
```

---

## Users

### Create User
```http
POST /api/v1/users
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "full_name": "John Doe",
  "org_id": "acme-corp"
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john_doe",
  "email": "john@example.com",
  "full_name": "John Doe",
  "org_id": "acme-corp",
  "created_at": "2025-12-01T10:30:00Z",
  "updated_at": "2025-12-01T10:30:00Z"
}
```

### Get User by ID
```http
GET /api/v1/users/{id}
```

### Update User
```http
PUT /api/v1/users/{id}
Content-Type: application/json

{
  "full_name": "John Smith",
  "email": "john.smith@example.com"
}
```

### Delete User
```http
DELETE /api/v1/users/{id}
```

### Search User by Username
```http
GET /api/v1/users/search?username=john_doe
```

### Get Users in Organization
```http
GET /api/v1/orgs/{orgId}/users
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john_doe",
    "email": "john@example.com",
    "full_name": "John Doe",
    "org_id": "acme-corp",
    "created_at": "2025-12-01T10:30:00Z",
    "updated_at": "2025-12-01T10:30:00Z"
  }
]
```

---

## Tasks

### Create Task
```http
POST /api/v1/users/{userId}/tasks
Content-Type: application/json

{
  "title": "Complete project documentation",
  "description": "Write comprehensive API docs",
  "priority": "high",
  "due_date": "2025-12-15T17:00:00Z"
}
```

**Priority Options:** `low`, `medium`, `high`, `urgent`

**Response:**
```json
{
  "id": "650e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Complete project documentation",
  "description": "Write comprehensive API docs",
  "status": "pending",
  "priority": "high",
  "due_date": "2025-12-15T17:00:00Z",
  "created_at": "2025-12-01T10:30:00Z",
  "updated_at": "2025-12-01T10:30:00Z",
  "completed_at": null
}
```

### Get Task by ID
```http
GET /api/v1/tasks/{id}
```

### Get User Tasks
```http
GET /api/v1/users/{userId}/tasks?status=pending
```

**Query Parameters:**
- `status` (optional) - Filter by status: `pending`, `in_progress`, `completed`, `cancelled`

### Get Tasks Due Soon
```http
GET /api/v1/users/{userId}/tasks/due-soon?hours=24
```

**Query Parameters:**
- `hours` (optional, default: 24) - Tasks due within this many hours

### Update Task
```http
PUT /api/v1/tasks/{id}
Content-Type: application/json

{
  "status": "completed",
  "description": "Documentation is complete"
}
```

**Status Options:** `pending`, `in_progress`, `completed`, `cancelled`

### Delete Task
```http
DELETE /api/v1/tasks/{id}
```

---

## Error Responses

All endpoints return standard HTTP status codes:

### Success Codes
- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `204 No Content` - Deletion successful

### Client Error Codes
- `400 Bad Request` - Invalid request body or parameters
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists

### Server Error Codes
- `500 Internal Server Error` - Server-side error
- `503 Service Unavailable` - Service temporarily unavailable

**Error Response Format:**
```json
{
  "error": "Error message description"
}
```

---

## Rate Limiting

Currently not implemented. Consider adding rate limiting for production:
- Redis-based rate limiting
- Per-IP or per-user limits
- Different limits for different endpoint tiers

---

## WebSocket Protocol

### Connection Lifecycle

1. **Connect:** Client opens WebSocket connection with query parameter
2. **Registration:** Server registers client to the group
3. **Communication:** Bidirectional message exchange
4. **Heartbeat:** Automatic ping/pong every 54 seconds
5. **Disconnect:** Client disconnects or timeout occurs
6. **Cleanup:** Server removes client from group

### Message Types

**Outgoing (Client → Server):**
```json
{
  "content": "Message text"
}
```

**Incoming (Server → Client):**
```json
{
  "org_id": "acme-corp",
  "group_id": "engineering",
  "client_id": "user-123",
  "content": "Message text"
}
```

### Connection Parameters

- **Ping Interval:** 54 seconds
- **Pong Timeout:** 60 seconds
- **Max Message Size:** 512 bytes
- **Message Buffer:** 256 messages

---

## Data Persistence

### PostgreSQL
- **Users:** Permanent storage
- **Tasks:** Permanent storage
- **Retention:** Until explicitly deleted

### Redis
- **Chat Messages:** Time-limited storage
- **TTL:** 7 days (configurable)
- **Max Messages per Group:** 1000 (configurable)
- **Automatic Cleanup:** Old messages are automatically removed

---

## Complete Example Workflow

### 1. Setup
```bash
# Start databases
docker-compose up -d

# Start application
go run main.go
```

### 2. Create Organization and Group
```bash
# Create organization
curl -X POST http://localhost:8080/api/v1/orgs \
  -H "Content-Type: application/json" \
  -d '{"id":"acme","name":"Acme Corp"}'

# Create group
curl -X POST http://localhost:8080/api/v1/orgs/acme/groups \
  -H "Content-Type: application/json" \
  -d '{"id":"eng","name":"Engineering"}'
```

### 3. Create Users
```bash
# Create Alice
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username":"alice",
    "email":"alice@acme.com",
    "full_name":"Alice Johnson",
    "org_id":"acme"
  }'

# Save the returned user ID
ALICE_ID="<returned-id>"
```

### 4. Create Tasks
```bash
curl -X POST http://localhost:8080/api/v1/users/$ALICE_ID/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Review code",
    "description":"Review PR #123",
    "priority":"high"
  }'
```

### 5. Connect to Chat
```javascript
// In browser or Node.js
const ws = new WebSocket('ws://localhost:8080/ws/orgs/acme/groups/eng?clientId=alice');

ws.onopen = () => {
  console.log('Connected!');
  ws.send(JSON.stringify({content: 'Hello team!'}));
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log(`${msg.client_id}: ${msg.content}`);
};
```

### 6. Retrieve History
```bash
# Get recent messages
curl http://localhost:8080/api/v1/orgs/acme/groups/eng/messages?limit=10
```

---

## Testing Tools

### cURL Examples
See above for basic cURL commands.

### Postman Collection
Import the provided `postman_collection.json` (if available).

### WebSocket Clients
- **wscat:** `npm install -g wscat`
- **Browser DevTools:** Use browser console
- **Postman:** Supports WebSocket connections

---

## Performance Considerations

### Recommended Limits
- **Max Concurrent WebSocket Connections:** 10,000 per server
- **Message Rate:** 100 messages/second per group
- **API Rate Limit:** 1000 requests/minute per IP
- **Max Message Size:** 512 bytes (configurable)

### Scaling
- Use Redis pub/sub for horizontal scaling
- Load balance WebSocket connections
- Database read replicas for read-heavy workloads
- Consider message queue for high-volume scenarios
