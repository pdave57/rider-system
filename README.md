# Runns — Delivery & Logistics Microservices Platform

Production-grade microservices in Go: 7 independent services, PostgreSQL per-service, Redis geo-dispatch, WebSocket GPS streaming, JWT auth.

## Quick Start

```bash
cd deploy
cp env/.env.example env/.env   # edit JWT_SECRET and DB_PASSWORD
docker-compose up --build
```

System runs at http://localhost (Nginx → Gateway → Services).

## Architecture

```
Clients → Nginx (80) → API Gateway (8080)
                            ↓
  auth-service     (8081)   order-service    (8082)
  dispatch-service (8083)   payment-service  (8084)
  realtime-service (8085)   rider-service    (8086)
                            ↓
                   PostgreSQL (5432)  Redis (6379)
```

## Services

| Service | Port | Responsibility |
|---------|------|----------------|
| auth-service | 8081 | JWT registration, login |
| order-service | 8082 | Order CRUD, state machine, tracking |
| dispatch-service | 8083 | Rider assignment (manual or GEORADIUS) |
| payment-service | 8084 | Wallet, billing, verification |
| realtime-service | 8085 | WebSocket GPS streaming, Redis pub/sub |
| rider-service | 8086 | Rider profile, availability, location |
| api-gateway | 8080 | Reverse proxy, unified routing |

## API Examples

### Register
```bash
curl -X POST http://localhost/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Jane","email":"jane@ex.com","phone":"08012345678","password":"pass","role":"client"}'
```

### Create Order
```bash
curl -X POST http://localhost/api/orders \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"pickup_address":"12 St","pickup_latitude":6.43,"pickup_longitude":3.42,"dropoff_address":"45 Ave","dropoff_latitude":6.59,"dropoff_longitude":3.36,"weight_kg":0.5}'
```

### Track Order (public)
```bash
curl http://localhost/api/orders/track/RYX-XXXXXXXX-XXXX
```

### WebSocket GPS
```bash
wscat -c "ws://localhost/ws" --header "Authorization: Bearer <rider_token>"
# send: {"latitude":6.5,"longitude":3.3,"speed":45}
```

## Structure

```
rider-system/
├── go.work
├── shared/                  # Models, middleware, utils (imported by all services)
├── services/
│   ├── api-gateway/         # Reverse proxy
│   ├── auth-service/        # handler/ service/ dto/
│   ├── order-service/       # handler/ service/ dto/
│   ├── dispatch-service/    # handler/ service/ dto/
│   ├── payment-service/     # handler/ service/ dto/
│   ├── realtime-service/    # handler/ hub/
│   └── rider-service/       # handler/ service/
├── infra/
│   ├── docker/              # Dockerfiles
│   └── nginx/               # nginx.conf
├── deploy/
│   ├── docker-compose.yml
│   └── env/.env
└── docs/
    ├── architecture.md
    └── api-spec.md
```

## Workflow

1. Register client, rider, admin → get JWT tokens
2. Client creates order → order-service calculates fee via Haversine distance
3. Client pays via wallet/card/cash → payment-service
4. Admin dispatches → dispatch-service auto-assigns nearest rider (Redis GEORADIUS)
5. Rider accepts → PATCH /api/orders/{id}/status (picked_up → delivered)
6. Realtime GPS tracking → WebSocket + Redis pub/sub broadcasts to watching clients

## Roles

- **client**: create orders, pay, track, wallet top-up
- **rider**: accept/reject dispatch, update location, toggle availability
- **admin**: assign orders, verify payments, view all orders

## Stack

- Go 1.22 (net/http stdlib only — no external router)
- PostgreSQL 16 (GORM, per-service schemas)
- Redis 7 (GEORADIUS, pub/sub, GPS cache)
- JWT (HS256, golang-jwt/jwt)
- bcrypt (golang.org/x/crypto)
- WebSocket (RFC 6455, handcrafted — no gorilla/websocket)
- Docker & Docker Compose
- Nginx (reverse proxy + WebSocket upgrade)

## Local Dev (without Docker)

```bash
# Run one service at a time
cd services/auth-service
go run main.go

# Requires: Postgres and Redis running locally
# Configure in deploy/env/.env with DB_HOST=localhost, REDIS_HOST=localhost
```
