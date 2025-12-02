# Go Realtime Workspace

A production-ready, scalable WebSocket-based real-time messaging and task management system for multi-tenant organizations with PostgreSQL and Redis integration.

## 🌟 Features

### Real-Time Messaging
✅ Multi-tenant organization support
✅ Group-based messaging within organizations  
✅ WebSocket real-time communication
✅ Message history with 7-day retention (Redis)
✅ Broadcast to entire organization or specific groups
✅ Automatic ping/pong heartbeat mechanism

### Data Persistence
✅ PostgreSQL for user and task management
✅ Redis for message history and caching
✅ Automatic timestamps and audit trails
✅ Connection pooling and health checks

### Task Management
✅ User task creation and tracking
✅ Status management (pending, in_progress, completed, cancelled)
✅ Priority levels (low, medium, high, urgent)
✅ Due date tracking with reminders
✅ Task filtering and search

### Production Ready
✅ Thread-safe concurrent operations
✅ Graceful shutdown handling
✅ Docker support with docker-compose
✅ Comprehensive error handling
✅ Health check endpoints

## 🏗️ Architecture Overview

```
OrgHub (Top Level)
  └── Organization 1
      ├── Group 1 (GroupHub)
      │   ├── Client 1 (WebSocket)
      │   ├── Client 2 (WebSocket)
      │   └── Messages → Redis
      └── Group 2
          ├── Clients...
          └── Messages → Redis
  
PostgreSQL
  ├── Users (Organization members)
  └── Tasks (User tasks)

Redis
  └── Message History (7-day TTL)
```

## 📚 Documentation

- **[Quick Start Guide](./quickstart.sh)** - Get up and running in minutes
- **[Database Setup](./DATABASE_SETUP.md)** - PostgreSQL and Redis installation
- **[API Reference](./API_REFERENCE.md)** - Complete API documentation
- **[Database Integration](./DATABASE_INTEGRATION.md)** - Integration details
- **[API Testing Guide](./API_TESTING.md)** - Testing examples

## 🚀 Quick Start

### Prerequisites

- Go 1.23.1 or higher
- Docker and docker-compose (recommended)
- OR PostgreSQL 14+ and Redis 7+ (local installation)

### Option 1: Using Docker (Recommended)

```bash
# 1. Clone the repository
git clone <repository-url>
cd go-org-hub-architecture-main

# 2. Run the quick start script
chmod +x quickstart.sh
./quickstart.sh

# 3. Install Go dependencies
go mod download

# 4. Start the application
go run main.go

# 5. Test the health endpoint
curl http://localhost:8080/health
# Should return: OK
```

### Option 2: Manual Setup

See [DATABASE_SETUP.md](./DATABASE_SETUP.md) for detailed instructions.

## 🔌 API Endpoints

### Organization Management

#### Create Organization
```bash
POST /api/v1/orgs
Content-Type: application/json

{
  "id": "org-123",
  "name": "My Organization"
}
```

#### Get All Organizations
```bash
GET /api/v1/orgs
```

### Group Management

#### Create Group
```bash
POST /api/v1/orgs/{orgId}/groups
Content-Type: application/json

{
  "id": "group-456",
  "name": "Engineering Team"
}
```

#### Get Organization Groups
```bash
GET /api/v1/orgs/{orgId}/groups
```

### WebSocket Connection

#### Join Group
```bash
WebSocket: ws://localhost:8080/ws/orgs/{orgId}/groups/{groupId}?clientId=client-789
```

Once connected, clients can send and receive messages in real-time.

### Broadcasting

#### Broadcast to Organization
```bash
POST /api/v1/orgs/{orgId}/broadcast
Content-Type: application/json

{
  "content": "Message to all groups in organization"
}
```

#### Broadcast to Group
```bash
POST /api/v1/orgs/{orgId}/groups/{groupId}/broadcast
Content-Type: application/json

{
  "content": "Message to specific group"
}
```

### Health Check
```bash
GET /health
```

## Message Format

WebSocket messages follow this JSON structure:

```json
{
  "org_id": "org-123",
  "group_id": "group-456",
  "client_id": "client-789",
  "content": "Hello, World!"
}
```

## Getting Started

### Prerequisites

- Go 1.23.1 or higher
- Dependencies:
  - `github.com/gorilla/mux` - HTTP router
  - `github.com/gorilla/websocket` - WebSocket implementation

### Installation

1. Clone the repository
```bash
git clone <repository-url>
cd go-org-hub-architecture-main
```

2. Install dependencies
```bash
go mod download
```

3. Run the server
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## Usage Example

### 1. Create an Organization
```bash
curl -X POST http://localhost:8080/api/v1/orgs \
  -H "Content-Type: application/json" \
  -d '{"id":"org1","name":"Acme Corp"}'
```

### 2. Create a Group
```bash
curl -X POST http://localhost:8080/api/v1/orgs/org1/groups \
  -H "Content-Type: application/json" \
  -d '{"id":"group1","name":"Engineering"}'
```

### 3. Connect via WebSocket
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/orgs/org1/groups/group1?clientId=user123');

ws.onopen = () => {
  console.log('Connected!');
  ws.send(JSON.stringify({
    content: 'Hello from client!'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
```

### 4. Broadcast Message
```bash
curl -X POST http://localhost:8080/api/v1/orgs/org1/groups/group1/broadcast \
  -H "Content-Type: application/json" \
  -d '{"content":"Team meeting in 5 minutes!"}'
```

## Configuration

Configuration can be customized in `config/config.go`. Default values:

- **Server Port**: 8080
- **Read/Write Timeout**: 15 seconds
- **WebSocket Buffer Size**: 1024 bytes
- **Message Buffer**: 256 messages
- **Ping Interval**: 54 seconds
- **Pong Wait**: 60 seconds

## Project Structure

```
.
├── main.go                 # Application entry point
├── go.mod                  # Go module dependencies
├── README.md              # This file
├── config/
│   └── config.go          # Configuration management
├── handlers/
│   └── websocket_handler.go  # HTTP/WebSocket handlers
└── hub/
    ├── client.go          # WebSocket client implementation
    ├── group_hub.go       # Group management
    └── org_hub.go         # Organization management
```

## Concurrency & Thread Safety

- All map operations are protected with `sync.RWMutex`
- Non-blocking channel sends to prevent deadlocks
- Buffered channels for message queuing
- Proper goroutine cleanup on disconnection

## Graceful Shutdown

The server handles `SIGINT` and `SIGTERM` signals gracefully:
- Stops accepting new connections
- Allows existing connections to complete
- 30-second timeout for shutdown

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go best practices
- Thread-safe operations
- Proper error handling
- Documentation for public APIs

## License

[Add your license here]

## Future Improvements

- [ ] Authentication and authorization
- [ ] Message persistence
- [ ] Redis pub/sub for horizontal scaling
- [ ] Metrics and monitoring
- [ ] Rate limiting
- [ ] Message encryption
- [ ] Client reconnection handling
- [ ] Presence tracking
