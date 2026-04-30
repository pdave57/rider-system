# Rydex System Architecture

Rydex is a **microservices-based delivery and logistics platform** built in Go with JWT authentication, PostgreSQL, Redis, WebSockets for realtime GPS tracking, and Nginx reverse proxy.

---

## High-Level Architecture

```
Clients (Mobile/Web)
       ↓
    Nginx (port 80)
       ↓
  API Gateway (port 8080)
       ↓
  ┌──────────────────────────────────────────────────┐
  │  Microservices (each with own DB/schema)         │
  ├──────────────────────────────────────────────────┤
  │  • auth-service      (8081) — registration, JWT  │
  │  • order-service     (8082) — order CRUD         │
  │  • dispatch-service  (8083) — rider assignment   │
  │  • payment-service   (8084) — wallet, billing    │
  │  • realtime-service  (8085) — WebSocket, GPS     │
  │  • rider-service     (8086) — rider profiles     │
  └──────────────────────────────────────────────────┘
       ↓                           ↓
  PostgreSQL (5432)           Redis (6379)
  (per-service schemas)       (GPS cache, pub/sub)
```

---

## Services

### 1. auth-service
- **Responsibilities**: User registration (client/rider/admin), login, JWT generation, password change
- **Database**: `rydex_auth` (users, rider_profiles, client_profiles)
- **Endpoints**:
  - `POST /api/auth/register`
  - `POST /api/auth/login`
  - `POST /api/auth/change-password` (authenticated)
  - `GET /api/auth/me` (authenticated)

### 2. order-service
- **Responsibilities**: Order creation, tracking, status transitions, cancellation
- **Database**: `rydex_orders` (orders, payments)
- **Endpoints**:
  - `POST /api/orders` (client)
  - `GET /api/orders` (role-based: client sees own, rider sees assigned, admin sees all pending)
  - `GET /api/orders/{id}`
  - `GET /api/orders/track/{code}` (public)
  - `PATCH /api/orders/{id}/status` (admin/rider)
  - `DELETE /api/orders/{id}` (client cancel)

### 3. dispatch-service
- **Responsibilities**: Assign orders to riders (manual or nearest auto-assign), accept/reject dispatch
- **Database**: `rydex_dispatch` (dispatches, rider_profiles)
- **Redis**: GEORADIUS for nearest rider lookup, GPS event cache
- **Endpoints**:
  - `POST /api/dispatch` (admin: assign order)
  - `POST /api/dispatch/{id}/respond` (rider: accept/reject)
  - `GET /api/dispatch/order/{orderID}`
  - `GET /api/dispatch/pending` (rider: my pending dispatches)

### 4. payment-service
- **Responsibilities**: Payment initiation (card/wallet/cash), wallet topup, verification
- **Database**: `rydex_payments` (payments, client_profiles for wallet balance)
- **Endpoints**:
  - `POST /api/payments` (client: initiate)
  - `POST /api/payments/verify` (admin)
  - `GET /api/payments/order/{orderID}`
  - `GET /api/payments/wallet` (client)
  - `POST /api/payments/wallet/topup` (client)

### 5. realtime-service
- **Responsibilities**: WebSocket hub, GPS event streaming, Redis pub/sub fan-out
- **Redis**: GPS cache (`gps:rider:{id}`), pub/sub channels (`gps:events`, `order:events`)
- **Endpoints**:
  - `GET /ws` (WebSocket upgrade, authenticated)
  - `GET /api/realtime/location/{riderID}` (latest GPS position)
- **Flow**: Riders send GPS updates via WS → cached in Redis → broadcast to watching clients

### 6. rider-service
- **Responsibilities**: Rider profile, availability toggle, location update
- **Database**: `rydex_riders` (rider_profiles)
- **Endpoints**:
  - `GET /api/riders/me` (rider)
  - `PATCH /api/riders/me/availability` (rider)
  - `PATCH /api/riders/me/location` (rider)
  - `GET /api/riders/nearby` (admin/client)

### 7. api-gateway
- **Responsibilities**: Reverse proxy, routes requests to services based on path prefix
- **Port**: 8080
- **Routes**:
  - `/api/auth/*` → auth-service
  - `/api/orders/*` → order-service
  - `/api/dispatch/*` → dispatch-service
  - `/api/payments/*` → payment-service
  - `/api/realtime/*` → realtime-service
  - `/ws` → realtime-service
  - `/api/riders/*` → rider-service

---

## Shared Layer

All services import `github.com/rydex/shared`:
- **models**: GORM models (User, Order, Payment, Dispatch, RiderProfile, ClientProfile, GPSEvent)
- **middleware**: JWT authentication, role guards, CORS, logging
- **utils**: JSON response helpers, ID generators, distance calculation, DB connection helpers

---

## Data Flow Examples

### 1. Client places order
1. Client → `POST /api/orders` (order-service)
2. order-service calculates delivery fee via Haversine distance
3. Creates order with `status=pending`, returns tracking code
4. Client → `POST /api/payments` (payment-service) to pay
5. If wallet: deduct immediately and mark `payment.status=success`
6. If card/cash: mark `payment.status=pending`, admin verifies later

### 2. Admin dispatches order
1. Admin → `POST /api/dispatch` with `order_id` (dispatch-service)
2. If `rider_id` provided: assign manually
3. If `rider_id` empty: query Redis GEORADIUS for nearest available rider
4. Create Dispatch record, update `order.status=assigned`, `order.rider_id=X`
5. Mark rider unavailable
6. Rider sees pending dispatch → `POST /api/dispatch/{id}/respond` to accept/reject

### 3. Real-time GPS tracking
1. Rider opens mobile app → WebSocket connects to `/ws`
2. Rider sends GPS events: `{"latitude": 6.5, "longitude": 3.4}`
3. realtime-service caches in Redis, updates GEORADIUS index, publishes to `gps:events` channel
4. All subscribed clients receive update via WebSocket broadcast

---

## Database Schema (Simplified)

Each service has its own database/schema for isolation:

**users** (auth-service)
- id, full_name, email, phone, password (bcrypt), role, is_active

**rider_profiles** (auth-service, rider-service)
- id, user_id, vehicle_type, vehicle_plate, license_number, is_available, current_latitude, current_longitude, rating, total_deliveries

**client_profiles** (auth-service, payment-service)
- id, user_id, default_address, latitude, longitude, wallet_balance

**orders** (order-service)
- id, tracking_code, client_id, rider_id, status, pickup/dropoff addresses & coords, package_desc, weight_kg, delivery_fee, notes, timestamps

**payments** (payment-service)
- id, order_id, client_id, amount, method, status, reference, gateway_ref, failure_reason, paid_at

**dispatches** (dispatch-service)
- id, order_id, rider_id, assigned_at, accepted_at, rejected_at

---

## Deployment

### Local dev (docker-compose)
```bash
cd deploy
docker-compose up --build
```

Services will be available:
- Nginx: http://localhost
- Gateway: http://localhost:8080
- Postgres: localhost:5432
- Redis: localhost:6379

### Environment
All config in `deploy/env/.env`:
- Database credentials
- Redis connection
- JWT secret
- Service URLs (for gateway routing)

---

## Security

- **JWT**: HS256 signed tokens with 24h expiry
- **Role-based access**: Middleware guards (RequireRoles) on sensitive endpoints
- **Password hashing**: bcrypt
- **Context injection**: User ID, role, email injected into request context after auth

---

## Scalability Notes

- Each service can scale horizontally (stateless except DB connections)
- Redis used for shared state (GPS cache, dispatch scoring)
- WebSocket connections are in-memory per realtime-service instance (use Redis pub/sub for cross-instance broadcast in production)
- Postgres per-service schemas allow independent migration and backup

---

## Future Enhancements

- **Service mesh**: Istio/Linkerd for mutual TLS, retries, circuit breaking
- **Message queue**: RabbitMQ/Kafka for async order processing, notifications
- **Monitoring**: Prometheus + Grafana for metrics, distributed tracing
- **API versioning**: `/v1/`, `/v2/` prefixes
- **Rate limiting**: Token bucket per user/IP in gateway
- **Multi-region**: Geo-replicated Postgres, Redis Sentinel for HA
