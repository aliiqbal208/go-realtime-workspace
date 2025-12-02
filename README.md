git clone <repository-url>
git clone <repository-url>
go mod download
go run main.go
# go-realtime-workspace

Lightweight, multi-tenant real‑time messaging (org + group + direct messages) over WebSockets with Redis history and PostgreSQL persistence.

## ✨ Core Features
* Organizations → Groups → Clients hierarchy
* Group chat + direct (client ↔ client) messaging
* Broadcast to org or single group
* Redis 7‑day message history (TTL)
* PostgreSQL users / tasks (extensible domain layer)
* Clean hub architecture, thread‑safe maps
* Graceful shutdown & health checks

## 🚀 Quick Start
```bash
git clone <repo-url>
cd go-realtime-workspace

# Start dependencies (adjust if you renamed docker-compose file)


# Run the server
go run main.go

# Health
curl http://localhost:8080/api/v1/health
```

## 🔌 WebSocket Usage
Group channel:
```
ws://localhost:8080/ws/orgs/{orgId}/groups/{groupId}?clientId={clientId}
```
Direct messaging channel:
```
ws://localhost:8080/ws/dm?clientId={clientId}
```
Send (client → group or direct):
```json
{
  "content": "Hello team",
  "recipient_id": "<clientId-optional-for-dm>"
}
```

## 🧾 Key HTTP Endpoints (minimal)
| Method | Path | Purpose |
| ------ | ---- | ------- |
| POST | /api/v1/orgs | Create organization |
| POST | /api/v1/orgs/{orgId}/groups | Create group |
| GET  | /api/v1/orgs | List orgs |
| GET  | /api/v1/orgs/{orgId}/groups | List groups in org |
| POST | /api/v1/orgs/{orgId}/broadcast | Org broadcast |
| POST | /api/v1/orgs/{orgId}/groups/{groupId}/broadcast | Group broadcast |
| GET  | /api/v1/health | Liveness & dependency check |

## 🗂 Structure
```
hub/         core hubs (org, group, client)
handlers/    HTTP + WebSocket handlers
middleware/  logging, recovery, CORS, rate limit, security, validation
config/      config & initialization
```

## ⚙️ Environment (set via vars or .env)
| Variable | Default | Purpose |
| -------- | ------- | ------- |
| APP_PORT | 8080 | HTTP / WS listen port |
| PG_HOST  | localhost | PostgreSQL host |
| PG_DB    | realtime_workspace | Database name |
| REDIS_ADDR | localhost:6379 | Redis address |

## 🛣 Roadmap (next)
* Auth (JWT / OAuth) & per‑org access control
* Horizontal scaling (Redis Pub/Sub fanout)
* Metrics / tracing (Prometheus + OpenTelemetry)
* Presence & typing indicators
* Rate limiting refinement / quota per org

## 🧪 Simple Browser Test
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/orgs/org1/groups/group1?clientId=alice');
ws.onopen = () => ws.send(JSON.stringify({ content: 'Hello world' }));
ws.onmessage = e => console.log('Received', e.data);
```

## 📄 License
Add your chosen license (MIT/Apache-2.0). Replace this section.

---
Concise by design. For extended docs (API details, DB schema, scaling notes) create `docs/` as the project evolves.
See [DATABASE_SETUP.md](./DATABASE_SETUP.md) for detailed instructions.
