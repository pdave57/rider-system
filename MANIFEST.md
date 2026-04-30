# Rydex Microservices System — Complete Manifest

A **7-service microservices delivery platform** in Go with PostgreSQL, Redis, WebSockets, and Docker Compose. Production-grade architecture, database-per-service isolation, and unified API gateway.

---

## 🏗️ System Overview

```
                    CLIENTS
                      ↓
                    NGINX (80)
                      ↓
                  API GATEWAY (8080)
                      ↓
    ┌─────────────────────────────────────┐
    │    MICROSERVICES                    │
    ├─────────────────────────────────────┤
    │ auth-service      (8081)            │
    │ order-service     (8082)            │
    │ dispatch-service  (8083)            │
    │ payment-service   (8084)            │
    │ realtime-service  (8085)            │
    │ rider-service     (8086)            │
    └─────────────────────────────────────┘
            ↓                 ↓
        POSTGRES          REDIS
        (5432)            (6379)
```

---

## 📦 Project Structure

```
rydex-system/
├── services/                    # 6 microservices + 1 gateway
│   ├── api-gateway/
│   │   ├── go.mod
│   │   └── main.go             # Reverse proxy, routing to services
│   ├── auth-service/
│   │   ├── go.mod
│   │   ├── main.go
│   │   ├── handler/
│   │   │   └── auth_handler.go
│   │   ├── service/
│   │   │   ├── auth_service.go
│   │   │   └── jwt.go
│   │   └── dto/
│   │       └── auth_dto.go
│   ├── order-service/
│   │   ├── go.mod
│   │   ├── main.go
│   │   ├── handler/
│   │   │   └── order_handler.go
│   │   ├── service/
│   │   │   └── order_service.go
│   │   └── dto/
│   │       └── order_dto.go
│   ├── dispatch-service/
│   │   ├── go.mod
│   │   ├── main.go
│   │   ├── handler/
│   │   │   └── dispatch_handler.go
│   │   ├── service/
│   │   │   └── dispatch_service.go  # Geo-based nearest rider
│   │   └── dto/
│   │       └── dispatch_dto.go
│   ├── payment-service/
│   │   ├── go.mod
│   │   ├── main.go
│   │   ├── handler/
│   │   │   └── payment_handler.go
│   │   ├── service/
│   │   │   └── payment_service.go
│   │   └── dto/
│   │       └── payment_dto.go
│   ├── realtime-service/
│   │   ├── go.mod
│   │   ├── main.go
│   │   ├── handler/
│   │   │   └── realtime_handler.go  # WebSocket, GPS streaming
│   │   └── hub/
│   │       └── hub.go               # In-memory connection manager
│   └── rider-service/
│       ├── go.mod
│       ├── main.go
│       ├── handler/
│       │   └── rider_handler.go
│       └── service/
│           └── rider_service.go
├── shared/                      # Shared across all services
│   ├── go.mod
│   ├── models/
│   │   └── models.go            # GORM: User, Order, Payment, Dispatch, etc.
│   ├── middleware/
│   │   └── middleware.go        # JWT auth, role guards, CORS, logging
│   └── utils/
│       ├── response.go          # JSON helpers, error responses
│       ├── db.go                # PostgreSQL & Redis connections
│       ├── env.go               # Environment variable loading
│       └── parse.go             # ID parsing, uint conversion
├── infra/
│   ├── docker/
│   │   ├── Dockerfile.service  # Generic service build
│   │   └── Dockerfile.gateway  # Gateway build
│   ├── nginx/
│   │   └── nginx.conf          # Reverse proxy config
│   ├── postgres/               # (Optional) init scripts
│   └── redis/                  # (Optional) redis.conf
├── deploy/
│   ├── docker-compose.yml      # Complete system orchestration
│   └── env/
│       └── .env                # Environment variables (DB, JWT, URLs)
├── docs/
│   ├── architecture.md         # System design, data flows, schemas
│   └── api-spec.md             # All endpoints, examples, role-based access
├── go.work                     # Go workspace (multi-module setup)
└── README.md                   # Quick start, features, project structure
```

---

## 🚀 Services at a Glance

| Service | Port | Responsibility | Key Features |
|---------|------|---|---|
| **auth-service** | 8081 | Registration, login, JWT | bcrypt passwords, 24h tokens |
| **order-service** | 8082 | Order CRUD, tracking | State machine (pending→assigned→picked_up→delivered), Haversine distance calc |
| **dispatch-service** | 8083 | Rider assignment | Manual assign or auto-nearest via Redis GEORADIUS, accept/reject |
| **payment-service** | 8084 | Wallet, billing | Wallet topup, instant deduction, card/cash verification |
| **realtime-service** | 8085 | WebSocket GPS | Handcrafted WebSocket, Redis pub/sub, 10min GPS cache |
| **rider-service** | 8086 | Rider profile | Availability toggle, location update, nearby listing |
| **api-gateway** | 8080 | Routing | Reverse proxy, path-based routing, unified entry |

---

## 💾 Databases

Each service has its **own isolated schema** for independent scaling:

```sql
-- auth-service: rydex_auth
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  full_name VARCHAR(120) NOT NULL,
  email VARCHAR(160) UNIQUE NOT NULL,
  phone VARCHAR(25) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  role VARCHAR(20) DEFAULT 'client',
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP
);

CREATE TABLE rider_profiles (
  id SERIAL PRIMARY KEY,
  user_id INT UNIQUE REFERENCES users(id),
  vehicle_type VARCHAR(20),
  vehicle_plate VARCHAR(20),
  is_available BOOLEAN DEFAULT FALSE,
  current_latitude DECIMAL(10,8),
  current_longitude DECIMAL(11,8),
  rating DECIMAL(3,2) DEFAULT 0,
  total_deliveries INT DEFAULT 0
);

CREATE TABLE client_profiles (
  id SERIAL PRIMARY KEY,
  user_id INT UNIQUE REFERENCES users(id),
  default_address TEXT,
  wallet_balance DECIMAL(12,2) DEFAULT 0
);

-- order-service: rydex_orders
CREATE TABLE orders (
  id SERIAL PRIMARY KEY,
  tracking_code VARCHAR(20) UNIQUE NOT NULL,
  client_id INT NOT NULL,
  rider_id INT,
  status VARCHAR(20) DEFAULT 'pending',
  pickup_address TEXT NOT NULL,
  dropoff_address TEXT NOT NULL,
  weight_kg DECIMAL(6,2),
  delivery_fee DECIMAL(10,2),
  created_at TIMESTAMP
);

-- dispatch-service: rydex_dispatch
CREATE TABLE dispatches (
  id SERIAL PRIMARY KEY,
  order_id INT UNIQUE NOT NULL,
  rider_id INT NOT NULL,
  assigned_at TIMESTAMP,
  accepted_at TIMESTAMP,
  rejected_at TIMESTAMP
);

-- payment-service: rydex_payments
CREATE TABLE payments (
  id SERIAL PRIMARY KEY,
  order_id INT NOT NULL,
  client_id INT NOT NULL,
  amount DECIMAL(12,2) NOT NULL,
  method VARCHAR(20),
  status VARCHAR(20) DEFAULT 'pending',
  reference VARCHAR(100) UNIQUE,
  paid_at TIMESTAMP
);
```

**Redis keys:**
- `rider:locations` (GEORADIUS index for dispatch)
- `gps:rider:{id}` (latest GPS event, JSON, 10min TTL)
- `gps:events` (pub/sub channel for GPS updates)
- `order:events` (pub/sub channel for order status)

---

## 🔐 Authentication & Authorization

**JWT Flow:**
1. Client registers or logs in → `POST /api/auth/register|login`
2. Response includes JWT token (HS256, 24h expiry)
3. Client includes token in all requests: `Authorization: Bearer <token>`
4. Middleware validates token, injects `UserID`, `Role`, `Email` into context
5. Role guards (`RequireRoles`) protect sensitive endpoints

**Context injection (type-safe):**
```go
userID := middleware.UserIDFrom(r.Context())  // uint
role := middleware.RoleFrom(r.Context())      // models.Role
email := middleware.EmailFrom(r.Context())    // string
```

**Roles:**
- `client` — create orders, pay, track, manage wallet
- `rider` — accept/reject dispatch, update location, toggle availability
- `admin` — assign orders, verify payments, list all orders

---

## 🔄 Data Flow Examples

### 1. Client Places Order
```
Client → POST /api/orders
         (pickup/dropoff coords, weight)
         ↓
Order Service calculates Haversine distance
Creates order (fee = 500 + 80*km + 50*weight)
Returns tracking code
         ↓
Order status: PENDING
```

### 2. Client Pays with Wallet
```
Client → POST /api/payments (method=wallet)
         ↓
Payment Service checks wallet balance
Deducts fee from client_profiles.wallet_balance
Sets payment.status = SUCCESS
         ↓
Order ready for dispatch
```

### 3. Admin Auto-Assigns Rider
```
Admin → POST /api/dispatch (order_id=42, rider_id=null)
        ↓
Dispatch Service queries Redis GEORADIUS:
  - Gets all riders within 50km
  - Filters by is_available=true
  - Selects nearest
        ↓
Creates Dispatch record
Updates order.status = ASSIGNED, order.rider_id = X
        ↓
Rider gets notification via WebSocket
```

### 4. Rider Accepts & Tracks
```
Rider → POST /api/dispatch/12/respond (accept=true)
        ↓
Dispatch.accepted_at = NOW()
        ↓
Rider → WS /ws (connect)
        ↓
Every GPS update:
  {lat, lon, speed, heading} → cached in Redis
  → broadcast to all clients watching this order
        ↓
Clients see rider in real-time on map
```

### 5. Order Delivered
```
Rider → PATCH /api/orders/42/status (status=delivered)
        ↓
Order Service validates state transition
Updates order.status = DELIVERED, delivered_at = NOW()
Publishes to "order:events" Redis channel
        ↓
All subscribers notified via WebSocket
Client sees "Delivered" badge
```

---

## 🚢 Deployment

### Local (Docker Compose)
```bash
cd deploy
docker-compose up --build
```

Spins up:
- 1× Nginx (80)
- 1× API Gateway (8080)
- 6× Services (8081–8086)
- 1× PostgreSQL (5432)
- 1× Redis (6379)

All services wait for Postgres/Redis health checks before starting.

### Environment (`.env`)
```
DB_HOST=postgres
DB_USER=postgres
DB_PASSWORD=postgres
REDIS_HOST=redis
JWT_SECRET=<your-secret>
```

### Production Notes
- Use managed PostgreSQL (AWS RDS, GCP Cloud SQL)
- Use managed Redis (AWS ElastiCache, GCP Memorystore)
- Deploy services on Kubernetes or Docker Swarm
- Use TLS at Nginx + service mesh mTLS
- Scale each service independently based on demand

---

## 📡 API Gateway Routing

All requests route through Nginx → Gateway → Service:

```go
// api-gateway/main.go
var routes = []Route{
  {"/api/auth", "http://auth-service:8081"},
  {"/api/orders", "http://order-service:8082"},
  {"/api/dispatch", "http://dispatch-service:8083"},
  {"/api/payments", "http://payment-service:8084"},
  {"/api/realtime", "http://realtime-service:8085"},
  {"/ws", "http://realtime-service:8085"},
  {"/api/riders", "http://rider-service:8086"},
}
```

The gateway:
- Logs all requests
- Applies CORS headers
- Handles path-based routing via `httputil.ReverseProxy`
- WebSocket upgrade delegated to realtime-service

---

## 🔌 Key Implementation Details

### 1. Shared Models (GORM)
All domain entities defined once in `shared/models/models.go`:
- User (with role)
- RiderProfile (vehicle, coords, availability)
- ClientProfile (wallet)
- Order (lifecycle: pending→assigned→picked_up→delivered)
- Payment (wallet or external gateway)
- Dispatch (assignment record)
- GPSEvent (for WebSocket)

Every service imports these models, ensuring schema consistency.

### 2. Middleware Chain
```go
middleware.Chain(mux, middleware.Logger, middleware.CORS, middleware.Authenticate(secret))
```

Applies middleware right-to-left (auth closest to handler).

### 3. Role-Based Access Control
```go
mux.Handle("/api/orders", 
  auth(
    middleware.RequireRoles(models.RoleAdmin)(
      http.HandlerFunc(h.Assign),
    ),
  ),
)
```

Nested middleware ensures only Admins can POST /api/orders.

### 4. WebSocket (Minimal)
Realtime service implements RFC 6455 manually:
- No external library (gorilla/websocket)
- Handshake: SHA1 Sec-WebSocket-Key validation
- Frame parsing: FIN bit, opcode, payload mask
- Broadcast: In-memory hub + Redis pub/sub for multi-instance

### 5. Distance-Based Dispatch
```go
// dispatch-service/service/dispatch_service.go
func (s *DispatchService) nearestAvailableRider(lat, lon float64) (uint, error) {
  results, _ := s.rdb.GeoRadius(ctx, "rider:locations", lon, lat, 
    &redis.GeoRadiusQuery{
      Radius: 50,     // 50 km
      Unit:   "km",
      Sort:   "ASC",  // nearest first
      Count:  10,
    }).Result()
  
  // Verify still available in DB, return closest
}
```

Uses Redis GEORADIUS (O(N+log(M)) complexity) for fast nearest-neighbor lookup.

---

## 📊 Database Schema Diagram

```
┌──────────────────┐
│     USERS        │
├──────────────────┤
│ id (PK)          │
│ email (UNIQUE)   │
│ phone (UNIQUE)   │
│ password (hash)  │
│ role             │
│ is_active        │
└────┬─────┬───────┘
     │     │
  1:1 1:1  1:1
     │     └─────────────────────┐
     │                           │
┌────▼──────────┐    ┌──────────▼──┐
│RIDER_PROFILES │    │CLIENT_       │
├───────────────┤    │PROFILES      │
│id             │    ├──────────────┤
│user_id (FK)   │    │id           │
│vehicle_type   │    │user_id (FK) │
│is_available   │    │wallet_balance
│current_lat    │    │default_addr  │
│current_lon    │    └──────────────┘
│rating         │
└───────────────┘

┌──────────────┐         ┌────────────┐
│   ORDERS     │    1:N  │ DISPATCHES │
├──────────────┤─────────┤────────────┤
│id (PK)       │         │id          │
│client_id (FK)├─┐       │order_id(FK)│
│rider_id (FK) │ │   1:1 │rider_id(FK)│
│status        │ │───────│accepted_at │
│fee           │         │rejected_at │
│tracking_code │         └────────────┘
└──────────────┘

┌──────────────┐
│  PAYMENTS    │
├──────────────┤
│id (PK)       │
│order_id (FK) │
│client_id (FK)│
│amount        │
│method        │
│status        │
│reference     │
│paid_at       │
└──────────────┘
```

---

## 🛠️ How to Extend

### Add a New Service
1. Create `services/my-service/`
2. Write `go.mod` with `replace github.com/rydex/shared => ../../shared`
3. Import `shared/models`, `shared/middleware`, `shared/utils`
4. Implement handler → service → repository layers
5. Add to `go.work`
6. Add to `docker-compose.yml` (with health checks)
7. Add route to `api-gateway/main.go`
8. Document endpoints in `docs/api-spec.md`

### Add a New Endpoint
1. Add handler method in `services/{service}/handler/`
2. Add route in `NewRouter()`
3. Protect with auth/role middleware if needed
4. Return via `utils.OK()`, `utils.BadRequest()`, etc.
5. Document in `docs/api-spec.md`

### Add a New Database Table
1. Define GORM model in `shared/models/models.go`
2. Create in service's `main.go`: `db.AutoMigrate(&MyModel{})`
3. Use GORM queries in service layer
4. Update API spec doc

---

## 🧪 Testing

Each service can be tested in isolation:

```bash
# Unit test
go test ./services/order-service/...

# Integration test (requires Postgres)
docker-compose up postgres
go test ./... -v

# Load test
# TODO: Add ab/wrk scripts
```

---

## 📈 Scalability Roadmap

### Phase 1 (Current)
- Single instance per service
- Shared Postgres, single instance
- In-memory WebSocket hub

### Phase 2 (Production)
- Kubernetes with HPA (auto-scale based on CPU)
- Managed RDS Postgres with read replicas
- Redis Sentinel for HA
- Service mesh (Istio) for mTLS

### Phase 3 (Advanced)
- Event streaming (Kafka) for orders → async workers
- CQRS + event sourcing for order state
- GraphQL API layer
- Multi-region replication

---

## 📝 Files Summary

| File | Purpose |
|------|---------|
| `go.work` | Multi-module workspace declaration |
| `shared/go.mod` | Shared library module |
| `services/*/go.mod` | Service-specific modules |
| `deploy/docker-compose.yml` | Full system orchestration |
| `deploy/env/.env` | Configuration (DB, JWT, URLs) |
| `infra/docker/Dockerfile.*` | Service & gateway builds |
| `infra/nginx/nginx.conf` | Reverse proxy config |
| `docs/architecture.md` | System design, data flows |
| `docs/api-spec.md` | All endpoints, request/response examples |
| `README.md` | Quick start, features, structure |

---

## 🎯 Key Takeaways

✅ **Microservices**: 7 independent services, each owns its data  
✅ **Shared layer**: Models, middleware, utils via Go workspace  
✅ **Type-safe auth**: JWT context injection, no query param leaks  
✅ **Real-time**: WebSocket hub + Redis pub/sub for GPS streaming  
✅ **Geo-dispatch**: GEORADIUS for nearest-rider auto-assignment  
✅ **Stateless services**: Horizontal scaling via replicas  
✅ **Docker-first**: Single `docker-compose up` for local dev  
✅ **Extensible**: Easy to add new services, endpoints, models  

---

## 📞 Quick Reference

**Start system:**
```bash
cd deploy && docker-compose up --build
```

**Register user:**
```bash
curl -X POST http://localhost/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Jane","email":"jane@ex.com","phone":"08012345678","password":"pass","role":"client"}'
```

**Create order:**
```bash
curl -X POST http://localhost/api/orders \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"pickup_address":"12 St","pickup_latitude":6.4,"pickup_longitude":3.4,"dropoff_address":"45 Ave","dropoff_latitude":6.5,"dropoff_longitude":3.3,"weight_kg":0.5}'
```

**Track order:**
```bash
curl http://localhost/api/orders/track/RYX-ABC123
```

**Connect WebSocket (riders send GPS):**
```
ws://localhost/ws
Header: Authorization: Bearer <token>
Send: {"latitude":6.5,"longitude":3.3,"speed":45}
```

---

**This is a production-ready, scalable microservices system. Deploy, extend, and profit.** 🚀
