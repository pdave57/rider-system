# Rydex — Delivery & Logistics Microservices Platform

A **production-grade microservices system** for delivery and logistics, built in **Go** with PostgreSQL, Redis, WebSockets, JWT auth, and Docker Compose orchestration.

---

## Features

✅ **Multi-role authentication** — Client, Rider, Admin with JWT  
✅ **Order lifecycle** — Create, assign, track, deliver  
✅ **Smart dispatch** — Auto-assign nearest available rider via Redis GEORADIUS  
✅ **Real-time GPS tracking** — WebSocket streaming with Redis pub/sub  
✅ **Wallet system** — Client wallet topup, instant payment deduction  
✅ **Microservices architecture** — 7 independent services, each with own DB  
✅ **Reverse proxy** — Nginx + API Gateway for unified routing  
✅ **Docker orchestration** — Single-command deployment  

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Nginx (port 80) → API Gateway (8080)                       │
├─────────────────────────────────────────────────────────────┤
│  Services:                                                  │
│  • auth-service      (8081) — JWT, registration            │
│  • order-service     (8082) — Order CRUD, tracking         │
│  • dispatch-service  (8083) — Rider assignment, accept      │
│  • payment-service   (8084) — Wallet, billing, verify      │
│  • realtime-service  (8085) — WebSocket GPS streaming       │
│  • rider-service     (8086) — Rider profile, availability  │
├─────────────────────────────────────────────────────────────┤
│  Infrastructure:                                            │
│  • PostgreSQL (5432) — Per-service schemas                 │
│  • Redis (6379) — GPS cache, pub/sub, GEORADIUS           │
└─────────────────────────────────────────────────────────────┘
```

**Shared layer** (`shared/`) contains:
- **models**: GORM models (User, Order, Payment, Dispatch, GPS events)
- **middleware**: JWT auth, role guards, CORS, logging
- **utils**: Response helpers, ID generators, distance calc, DB connectors

---

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local dev)

### 1. Clone & Setup
```bash
git clone <repo>
cd rydex-system
```

### 2. Configure Environment
```bash
cp deploy/env/.env.example deploy/env/.env
# Edit JWT_SECRET and DB credentials
```

### 3. Start All Services
```bash
cd deploy
docker-compose up --build
```

Services will be available at:
- **Nginx**: http://localhost
- **API Gateway**: http://localhost:8080
- **Postgres**: localhost:5432
- **Redis**: localhost:6379

### 4. Test
```bash
# Register a client
curl -X POST http://localhost/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "full_name": "Jane Client",
    "email": "jane@example.com",
    "phone": "08012345678",
    "password": "pass123",
    "role": "client"
  }'

# Create an order (use token from register response)
curl -X POST http://localhost/api/orders \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{
    "pickup_address": "12 Adeola St, Lagos",
    "pickup_latitude": 6.4281,
    "pickup_longitude": 3.4219,
    "dropoff_address": "45 Allen Ave, Ikeja",
    "dropoff_latitude": 6.5958,
    "dropoff_longitude": 3.3678,
    "package_desc": "Documents",
    "weight_kg": 0.5
  }'
```

---

## Project Structure

```
rydex-system/
├── services/
│   ├── api-gateway/          # Reverse proxy, routes to services
│   ├── auth-service/         # Registration, login, JWT
│   ├── order-service/        # Order CRUD, tracking
│   ├── dispatch-service/     # Rider assignment, nearest-rider algo
│   ├── payment-service/      # Wallet, billing
│   ├── realtime-service/     # WebSocket hub, GPS streaming
│   └── rider-service/        # Rider profile, availability
├── shared/
│   ├── models/               # GORM models (User, Order, Payment, etc.)
│   ├── utils/                # Response helpers, DB connectors
│   └── middleware/           # JWT auth, role guards, CORS
├── infra/
│   ├── docker/               # Dockerfiles
│   ├── nginx/                # Nginx config
│   ├── postgres/             # (optional) init scripts
│   └── redis/                # (optional) redis.conf
├── deploy/
│   ├── docker-compose.yml    # Full system orchestration
│   └── env/
│       └── .env              # Environment variables
├── docs/
│   ├── architecture.md       # System design, data flow
│   └── api-spec.md           # API endpoints, examples
├── go.work                   # Go workspace (multi-module)
└── README.md
```

---

## Key Design Decisions

### 1. Database-per-Service
Each service has its own PostgreSQL schema (`rydex_auth`, `rydex_orders`, etc.). This allows independent scaling, migrations, and backups.

### 2. JWT Context Injection
Middleware validates the JWT and injects `UserID`, `Role`, `Email` into request context. Services read from context (never from query params) to prevent spoofing.

### 3. Redis for Geo + Realtime
- **GEORADIUS**: Dispatch service queries Redis for nearest available rider
- **Pub/Sub**: GPS events broadcast to all WebSocket clients via Redis channels
- **Cache**: Latest GPS position stored as JSON with 10min TTL

### 4. Minimal WebSocket Implementation
Realtime service uses a handcrafted WebSocket handler (no external lib). Frames are parsed manually. This keeps the binary small and dependencies minimal.

### 5. Shared Module
All domain models, middleware, and utils live in `shared/`. Services import via Go workspace. This ensures consistency (same JWT validation, same User model) without code duplication.

---

## API Examples

See [`docs/api-spec.md`](docs/api-spec.md) for full reference.

**Register a rider:**
```bash
curl -X POST http://localhost/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "full_name": "John Rider",
    "email": "john@example.com",
    "phone": "08098765432",
    "password": "pass123",
    "role": "rider",
    "vehicle_type": "motorcycle",
    "vehicle_plate": "LAG-456AB"
  }'
```

**Auto-assign nearest rider:**
```bash
curl -X POST http://localhost/api/dispatch \
  -H 'Authorization: Bearer <admin_token>' \
  -H 'Content-Type: application/json' \
  -d '{
    "order_id": 42
  }'
# If rider_id is omitted, dispatch-service queries Redis GEORADIUS for the nearest available rider
```

**WebSocket GPS tracking:**
```javascript
const ws = new WebSocket('ws://localhost/ws', {
  headers: { Authorization: 'Bearer <token>' }
});

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'gps_update') {
    console.log('Rider location:', msg.payload.latitude, msg.payload.longitude);
  }
};

// Rider sends GPS:
ws.send(JSON.stringify({ latitude: 6.5244, longitude: 3.3792, speed: 45.2 }));
```

---

## Development

### Run a single service locally
```bash
cd services/auth-service
go run main.go
# Runs on port 8081 (configurable via AUTH_SERVICE_PORT env var)
```

### Run tests
```bash
# TODO: Add unit tests for each service
go test ./...
```

### Add a new service
1. Create directory: `services/my-service/`
2. Add `go.mod` with `replace github.com/rydex/shared => ../../shared`
3. Import shared models/middleware
4. Add to `go.work` and `docker-compose.yml`
5. Route in `api-gateway/main.go`

---

## Production Considerations

- **TLS**: Terminate SSL at Nginx, use cert from Let's Encrypt
- **Secrets**: Store JWT_SECRET in HashiCorp Vault or AWS Secrets Manager
- **Service mesh**: Istio for mutual TLS, retries, circuit breaking
- **Horizontal scaling**: Run multiple instances of each service behind load balancer
- **Monitoring**: Prometheus + Grafana for metrics, Jaeger for tracing
- **Message queue**: RabbitMQ/Kafka for async order notifications, email dispatch
- **Rate limiting**: Implement token bucket in api-gateway per IP/user

---

## License

MIT

---

## Contributing

1. Fork the repo
2. Create a feature branch
3. Commit your changes
4. Open a pull request

---

## Support

For questions or issues, open a GitHub issue or reach out to the team.
