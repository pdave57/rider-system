================================================================================
                    Rider MICROSERVICES SYSTEM
             Complete Delivery & Logistics Platform in Go
================================================================================

PROJECT OVERVIEW
================================================================================
A production-grade microservices delivery platform with 7 independent Go services,
PostgreSQL per-service isolation, Redis geo-indexing, WebSocket GPS streaming,
JWT authentication, and unified API Gateway via Nginx.

Designed to be:
  ✓ Scalable (stateless services, independent databases)
  ✓ Extensible (easy to add new services/endpoints)
  ✓ Type-safe (shared models, interfaces)
  ✓ Containerized (single docker-compose up --build)

ARCHITECTURE DIAGRAM
================================================================================

        Clients (mobile/web)
              ↓
          NGINX (80)
              ↓
      API GATEWAY (8080)
              ↓
  ┌────────────────────────────────────┐
  │      MICROSERVICES                 │
  ├────────────────────────────────────┤
  │  auth-service        (8081)        │  JWT, registration
  │  order-service       (8082)        │  Order CRUD, tracking
  │  dispatch-service    (8083)        │  Rider assignment, geo-lookup
  │  payment-service     (8084)        │  Wallet, billing
  │  realtime-service    (8085)        │  WebSocket, GPS streaming
  │  rider-service       (8086)        │  Rider profile, availability
  └────────────────────────────────────┘
          ↓              ↓
      PostgreSQL      Redis
      (5432)          (6379)
      (per-service    (geo, pub/sub,
       schemas)       GPS cache)

DIRECTORY STRUCTURE
================================================================================

rydex-system/
├── services/
│   ├── api-gateway/             Reverse proxy, routing
│   ├── auth-service/            Registration, JWT, login
│   ├── order-service/           Order CRUD, tracking, status FSM
│   ├── dispatch-service/        Rider assignment, GEORADIUS
│   ├── payment-service/         Wallet, billing, verification
│   ├── realtime-service/        WebSocket hub, GPS streaming
│   └── rider-service/           Rider profile, location, availability
├── shared/
│   ├── models/                  GORM: User, Order, Payment, Dispatch
│   ├── middleware/              JWT auth, role guards, CORS, logging
│   └── utils/                   Helpers, DB connectors, distance calc
├── infra/
│   ├── docker/                  Dockerfiles (service + gateway)
│   └── nginx/                   Reverse proxy config
├── deploy/
│   ├── docker-compose.yml       Full system orchestration
│   └── env/.env                 Environment variables
├── docs/
│   ├── architecture.md          System design, data flows, schemas
│   ├── api-spec.md              All endpoints, role-based access
│   ├── MANIFEST.md              Complete file structure + details
│   └── QUICKSTART.md            Step-by-step workflow examples
├── go.work                      Multi-module workspace
└── README.md                    Project overview

KEY FILES
================================================================================

1. docker-compose.yml
   - Defines all 8 containers (postgres, redis, nginx, gateway, 6 services)
   - Health checks, environment injection, volume mounting
   - Single command: docker-compose up --build

2. shared/models/models.go
   - GORM models: User, Order, Payment, Dispatch, RiderProfile, etc.
   - Enums: Role (client/rider/admin), OrderStatus, PaymentStatus
   - Shared across all services

3. shared/middleware/middleware.go
   - JWT authentication (HS256, 24h tokens)
   - Role guards (RequireRoles)
   - CORS, logging, context injection
   - Type-safe context access (UserIDFrom, RoleFrom, etc.)

4. {service}/service/{service}_service.go
   - Business logic layer (no HTTP concerns)
   - Repository pattern for data access
   - Error handling, validation

5. {service}/handler/{service}_handler.go
   - HTTP handlers, routing, middleware chains
   - Request/response marshaling
   - Status code logic

WORKFLOW EXAMPLES
================================================================================

1. CLIENT PLACES ORDER
   ├─ POST /api/orders (with pickup/dropoff coords, weight)
   ├─ order-service calculates Haversine distance
   ├─ Fee = 500 + 80*km + 50*weight_kg (in naira)
   └─ Returns tracking code, status=PENDING

2. CLIENT PAYS WITH WALLET
   ├─ POST /api/payments (method=wallet)
   ├─ payment-service deducts from client_profiles.wallet_balance
   └─ payment.status=SUCCESS, order ready for dispatch

3. ADMIN AUTO-ASSIGNS RIDER
   ├─ POST /api/dispatch (order_id=42, rider_id=null)
   ├─ dispatch-service queries Redis GEORADIUS for nearest available rider
   ├─ Creates Dispatch record, updates order.status=ASSIGNED
   └─ Rider receives notification

4. RIDER ACCEPTS & UPDATES LOCATION
   ├─ POST /api/dispatch/{id}/respond (accept=true)
   ├─ GET /ws (WebSocket connection)
   ├─ Send GPS: {lat, lon, speed, heading} every N seconds
   └─ Cached in Redis, broadcast to all watching clients (real-time)

5. ORDER DELIVERED
   ├─ PATCH /api/orders/{id}/status (status=delivered)
   ├─ order-service validates state machine transition
   └─ Publishes to "order:events" Redis channel

AUTHENTICATION & AUTHORIZATION
================================================================================

JWT Flow:
  1. Register/Login → POST /api/auth/register or /api/auth/login
  2. Response includes JWT token (HS256, 24h expiry)
  3. Client includes token: Authorization: Bearer <token>
  4. Middleware validates, injects UserID/Role/Email into context
  5. Role guards (RequireRoles) protect sensitive endpoints

Roles:
  ├─ client   (create orders, pay, track, manage wallet)
  ├─ rider    (accept/reject dispatch, update location)
  └─ admin    (assign orders, verify payments, list all)

Context Injection (type-safe):
  ├─ userID := middleware.UserIDFrom(r.Context())
  ├─ role := middleware.RoleFrom(r.Context())
  └─ email := middleware.EmailFrom(r.Context())

DATABASE ISOLATION
================================================================================

Each service owns its database schema:

  auth-service      → rydex_auth
    ├─ users
    ├─ rider_profiles
    └─ client_profiles

  order-service     → rydex_orders
    └─ orders

  dispatch-service  → rydex_dispatch
    └─ dispatches

  payment-service   → rydex_payments
    └─ payments

  rider-service     → rydex_riders
    └─ rider_profiles (synced with auth-service)

Benefits:
  ✓ Independent schema evolution
  ✓ Per-service backup/restore
  ✓ Horizontal scaling without bottleneck
  ✓ Clear data ownership

REDIS USAGE
================================================================================

Keys:
  ├─ rider:locations        (GEORADIUS index for dispatch geo-lookup)
  ├─ gps:rider:{id}         (latest GPS event, JSON, 10min TTL)
  └─ pub/sub channels
      ├─ gps:events         (GPS updates broadcast)
      └─ order:events       (Order status updates broadcast)

Operations:
  ├─ GeoAdd             (update rider location in GEORADIUS)
  ├─ GeoRadius          (find nearest riders within radius)
  ├─ Get/Set            (cache GPS position)
  ├─ Publish/Subscribe  (broadcast events to WebSocket clients)

Performance:
  ├─ GEORADIUS: O(N + log(M)) for N riders, M within radius
  ├─ Get/Set: O(1)
  └─ Pub/Sub: O(1) per subscriber

GETTING STARTED
================================================================================

1. Start the system
   $ cd deploy
   $ docker-compose up --build
   
   This spins up:
   - Nginx (80)
   - API Gateway (8080)
   - PostgreSQL (5432)
   - Redis (6379)
   - 6 microservices (8081-8086)

2. Register a client
   $ curl -X POST http://localhost/api/auth/register \
     -H 'Content-Type: application/json' \
     -d '{
       "full_name": "Jane Client",
       "email": "jane@ex.com",
       "phone": "08012345678",
       "password": "pass123",
       "role": "client"
     }'

3. Create an order
   $ curl -X POST http://localhost/api/orders \
     -H "Authorization: Bearer <token>" \
     -H 'Content-Type: application/json' \
     -d '{
       "pickup_address": "12 St, Lagos",
       "pickup_latitude": 6.4281,
       "pickup_longitude": 3.4219,
       "dropoff_address": "45 Ave, Ikeja",
       "dropoff_latitude": 6.5958,
       "dropoff_longitude": 3.3678,
       "weight_kg": 0.5
     }'

4. Track order (public)
   $ curl http://localhost/api/orders/track/RYX-ABC123

5. WebSocket GPS streaming
   $ wscat -c "ws://localhost/ws" \
     --header "Authorization: Bearer <rider_token>"
   
   Send GPS: {"latitude": 6.5, "longitude": 3.3, "speed": 45}

See QUICKSTART.md for complete workflow with all steps.

SCALING STRATEGY
================================================================================

Stateless Services:
  ├─ Each service has no in-memory state
  ├─ Only database/Redis used for persistent data
  └─ Replicas can be spun up/down without data loss

Database Per Service:
  ├─ Independent Postgres instances (or managed RDS)
  ├─ No cross-service schema dependencies
  └─ Schema migrations isolated per service

Horizontal Scaling:
  ├─ Auth: stateless, scale freely
  ├─ Order: stateless, scale freely
  ├─ Dispatch: stateless (geo-lookup in Redis), scale freely
  ├─ Payment: stateless, scale freely
  ├─ Realtime: in-memory hub (use Redis pub/sub for multi-instance)
  └─ Rider: stateless, scale freely

Production Architecture:
  ├─ Kubernetes with HPA (auto-scale by CPU)
  ├─ AWS RDS PostgreSQL (multi-AZ)
  ├─ AWS ElastiCache Redis (replication, Sentinel)
  ├─ Service mesh (Istio) for mTLS
  ├─ API rate limiting in gateway
  └─ Monitoring: Prometheus + Grafana

IMPLEMENTATION HIGHLIGHTS
================================================================================

1. Minimal WebSocket (no external library)
   - Handcrafted RFC 6455 implementation
   - SHA1 key validation
   - Frame parsing (FIN, opcode, mask)
   - In-memory hub + Redis pub/sub

2. Distance Calculation (Haversine)
   - Accurate to ~0.5% for short distances
   - Used for fee calculation and dispatch scoring
   - shared/utils/response.go: HaversineKm()

3. Role-Based Middleware Chaining
   - Composable: Authenticate → RequireRoles → Handler
   - Right-to-left application
   - No third-party router needed (stdlib net/http)

4. Type-Safe Context
   - Typed context keys prevent key collision
   - Middleware injects UserID/Role/Email as uint/enum/string
   - Services extract with type-safe helpers

5. GORM AutoMigrate
   - Each service runs AutoMigrate on startup
   - Schemas created automatically
   - No separate migration tool needed

6. Shared Module via go.work
   - Go 1.18+ workspace feature
   - All services import shared/
   - Single source of truth for models/middleware/utils

TESTING
================================================================================

Unit Tests:
  $ go test ./services/order-service/...

Integration Tests (requires Postgres):
  $ docker-compose up postgres
  $ go test ./... -v

Load Testing:
  $ ab -n 1000 -c 10 http://localhost/api/orders
  $ wrk -t4 -c100 -d30s http://localhost/api/orders

Mock Tests:
  - Repository pattern allows easy mocking
  - See shared/models for interface examples

EXTENDING THE SYSTEM
================================================================================

Add a New Service:
  1. mkdir services/my-service/{handler,service,dto}
  2. Create go.mod with:
     replace github.com/rydex/shared => ../../shared
  3. Implement handler → service → repository
  4. Add to go.work
  5. Add to docker-compose.yml
  6. Add route to api-gateway/main.go

Add a New Endpoint:
  1. Add handler method in {service}/handler/
  2. Add route in NewRouter()
  3. Protect with auth/role middleware if needed
  4. Return via utils.OK(), utils.BadRequest(), etc.
  5. Document in docs/api-spec.md

Add a New Data Model:
  1. Define GORM struct in shared/models/models.go
  2. Add AutoMigrate call in service's main.go
  3. Use GORM in service layer
  4. Document in MANIFEST.md

DOCUMENTATION FILES
================================================================================

README.md
  ├─ Project overview
  ├─ Quick start
  ├─ Architecture diagram
  ├─ Project structure
  └─ Key design decisions

QUICKSTART.md
  ├─ Prerequisites
  ├─ Step-by-step workflow (register → order → dispatch → deliver)
  ├─ API endpoint examples with curl
  ├─ WebSocket connection example
  ├─ Troubleshooting
  └─ Use case examples

MANIFEST.md
  ├─ Complete project structure
  ├─ Services at a glance
  ├─ Database schemas (SQL)
  ├─ Data flow examples
  ├─ Authentication & authorization
  ├─ Implementation details
  ├─ How to extend
  ├─ Files summary
  └─ Key takeaways

docs/architecture.md
  ├─ System overview
  ├─ Service descriptions
  ├─ Shared layer details
  ├─ Data flow examples
  ├─ Database schema
  ├─ Deployment
  ├─ Security notes
  ├─ Future enhancements
  └─ Scalability roadmap

docs/api-spec.md
  ├─ All endpoints (POST/GET/PATCH/DELETE)
  ├─ Request/response examples
  ├─ Status codes
  ├─ Role-based access table
  ├─ Example workflow
  └─ Common response format

SUMMARY
================================================================================

What You Get:
  ✓ 7 production-grade microservices (1,500+ lines of Go)
  ✓ Shared models, middleware, utilities
  ✓ Docker Compose for local dev
  ✓ Complete API specification
  ✓ Database-per-service architecture
  ✓ Real-time WebSocket GPS tracking
  ✓ Nearest-rider auto-assignment (Redis GEORADIUS)
  ✓ JWT authentication with role-based access
  ✓ Wallet system with instant payment deduction

Key Technologies:
  ├─ Go 1.22+
  ├─ PostgreSQL 16
  ├─ Redis 7
  ├─ Nginx
  ├─ Docker & Compose
  ├─ GORM (ORM)
  ├─ JWT (golang-jwt/jwt)
  ├─ Crypto (bcrypt, sha1)
  └─ Native stdlib (net/http, context, encoding/json)

Ready to Deploy:
  ├─ Local: docker-compose up --build
  ├─ Cloud: Deploy to Kubernetes
  ├─ Scale: Add replicas per service
  ├─ Monitor: Integrate Prometheus + Grafana
  └─ Extend: Add new services/endpoints easily

================================================================================
See README.md, QUICKSTART.md, and docs/ for detailed documentation.
Happy coding! 🚀
================================================================================
