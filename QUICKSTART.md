# Runns — Quick Start Guide

Get the complete delivery platform running in 3 commands.

---

## Prerequisites

- Docker & Docker Compose
- `curl` (for API testing)

---

## 1️⃣ Start Everything

```bash
cd rider-system/deploy
docker-compose up --build
```

This spins up:
- PostgreSQL (localhost:5432)
- Redis (localhost:6379)
- Nginx (localhost:80)
- API Gateway (localhost:8080)
- 6 microservices (ports 8081–8086)

Wait for "service listening" messages. Then proceed.

---

## 2️⃣ Register Users

### Register a Client
```bash
curl -X POST http://localhost/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "full_name": "Lucky Ajidoku",
    "email": "luckyajidoku@example.com",
    "phone": "08064558643",
    "password": "pass123",
    "role": "client"
  }'
```

Response:
```json
{
  "success": true,
  "message": "registration successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "full_name": "Alice Client",
      "email": "alice@example.com",
      "phone": "08012345678",
      "role": "client",
      "is_active": true
    }
  }
}
```

**Save the token:**
```bash
export CLIENT_TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

### Register a Rider
```bash
curl -X POST http://localhost/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "full_name": "Bob Rider",
    "email": "bob@example.com",
    "phone": "08087654321",
    "password": "pass123",
    "role": "rider",
    "vehicle_type": "motorcycle",
    "vehicle_plate": "LAG-123XY"
  }'
```

Save token:
```bash
export RIDER_TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

### Register an Admin
```bash
curl -X POST http://localhost/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{
    "full_name": "Dave Admin",
    "email": "pdave@example.com",
    "phone": "08011112223",
    "password": "pass123",
    "role": "admin"
  }'
```

Save token:
```bash
export ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

---

## 3️⃣ Test the Workflow

### Step 1: Client Creates an Order

```bash
curl -X POST http://localhost/api/orders \
  -H "Authorization: Bearer $CLIENT_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "pickup_address": "12 Adeola Street, Ikoyi, Lagos",
    "pickup_latitude": 6.4281,
    "pickup_longitude": 3.4219,
    "dropoff_address": "45 Allen Avenue, Ikeja, Lagos",
    "dropoff_latitude": 6.5958,
    "dropoff_longitude": 3.3678,
    "package_desc": "Important documents",
    "weight_kg": 0.5,
    "notes": "Handle with care"
  }'
```

Response:
```json
{
  "success": true,
  "message": "order created",
  "data": {
    "id": 1,
    "tracking_code": "RYX-A3F2B1C4-5678",
    "client_id": 1,
    "status": "pending",
    "delivery_fee": 1200.0
  }
}
```

Save the order ID and tracking code:
```bash
export ORDER_ID=1
export TRACKING_CODE="RYX-A3F2B1C4-5678"
```

### Step 2: Client Pays (Wallet)

```bash
curl -X POST http://localhost/api/payments \
  -H "Authorization: Bearer $CLIENT_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "order_id": 1,
    "method": "wallet"
  }'
```

Response:
```json
{
  "success": true,
  "message": "payment initiated",
  "data": {
    "id": 1,
    "order_id": 1,
    "amount": 1200.0,
    "method": "wallet",
    "status": "success"
  }
}
```

### Step 3: Rider Sets Availability

```bash
curl -X PATCH http://localhost/api/riders/me/availability \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "available": true
  }'
```

Response:
```json
{
  "success": true,
  "message": "availability updated"
}
```

### Step 4: Admin Auto-Assigns Order to Nearest Rider

```bash
curl -X POST http://localhost/api/dispatch \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "order_id": 1
  }'
```

The system queries Redis GEORADIUS, finds the nearest available rider (Bob), and assigns the order.

Response:
```json
{
  "success": true,
  "message": "order dispatched",
  "data": {
    "id": 1,
    "order_id": 1,
    "rider_id": 2,
    "status": "pending"
  }
}
```

Save dispatch ID:
```bash
export DISPATCH_ID=1
```

### Step 5: Rider Accepts the Order

```bash
curl -X POST http://localhost/api/dispatch/$DISPATCH_ID/respond \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "accept": true
  }'
```

Response:
```json
{
  "success": true,
  "message": "dispatch response recorded",
  "data": {
    "status": "accepted"
  }
}
```

### Step 6: Rider Updates Location (GPS)

This can be done via WebSocket (real-time) or HTTP:

```bash
curl -X PATCH http://localhost/api/riders/me/location \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "latitude": 6.4500,
    "longitude": 3.4100
  }'
```

Response:
```json
{
  "success": true,
  "message": "location updated"
}
```

### Step 7: Rider Marks as Picked Up

```bash
curl -X PATCH http://localhost/api/orders/$ORDER_ID/status \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "status": "picked_up"
  }'
```

### Step 8: Rider Marks as Delivered

```bash
curl -X PATCH http://localhost/api/orders/$ORDER_ID/status \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "status": "delivered"
  }'
```

---

## 4️⃣ Real-Time Tracking (WebSocket)

Open a new terminal and connect as a rider to send GPS updates in real-time:

```bash
# Install wscat (Node.js required)
npm install -g wscat

# Connect
wscat -c "ws://localhost/ws" \
  --header "Authorization: Bearer $RIDER_TOKEN"
```

Once connected, send GPS events:
```json
{
  "latitude": 6.4500,
  "longitude": 3.4100,
  "speed": 45.5,
  "heading": 180.0
}
```

The client watching this order will receive updates in real-time.

---

## 5️⃣ Public Tracking

Anyone can track an order without authentication:

```bash
curl http://localhost/api/orders/track/$TRACKING_CODE
```

Response:
```json
{
  "success": true,
  "message": "order found",
  "data": {
    "id": 1,
    "tracking_code": "RYX-A3F2B1C4-5678",
    "status": "delivered",
    "pickup_address": "12 Adeola Street, Ikoyi, Lagos",
    "dropoff_address": "45 Allen Avenue, Ikeja, Lagos"
  }
}
```

---

## 🔍 Useful Endpoints

### Check System Health
```bash
curl http://localhost/health
```

### Get My Profile
```bash
curl -H "Authorization: Bearer $CLIENT_TOKEN" \
  http://localhost/api/auth/me
```

### Check Wallet Balance
```bash
curl -H "Authorization: Bearer $CLIENT_TOKEN" \
  http://localhost/api/payments/wallet
```

### List My Orders
```bash
curl -H "Authorization: Bearer $CLIENT_TOKEN" \
  http://localhost/api/orders
```

### Rider: Get Pending Dispatches
```bash
curl -H "Authorization: Bearer $RIDER_TOKEN" \
  http://localhost/api/dispatch/pending
```

---

## 🐛 Troubleshooting

### Services won't start
```bash
docker-compose logs postgres
docker-compose logs redis
```

### Token invalid
Make sure you copied the token without quotes:
```bash
export CLIENT_TOKEN=eyJhbGc...  # No quotes!
```

### Database locked
```bash
docker-compose down -v
docker-compose up --build
```

### Port already in use
```bash
# Change port in docker-compose.yml
# Or kill the process:
lsof -i :8080
kill -9 <PID>
```

---

## 📚 Next Steps

1. **Read the docs:**
   - `docs/architecture.md` — System design, data flows
   - `docs/api-spec.md` — Complete API reference
   - `MANIFEST.md` — Detailed file structure

2. **Extend the system:**
   - Add a new service (see Architecture doc)
   - Add a new endpoint
   - Integrate with external payment gateway

3. **Deploy to production:**
   - Use Kubernetes instead of Docker Compose
   - Add TLS at Nginx
   - Use managed Postgres & Redis
   - Set up monitoring (Prometheus + Grafana)

---

## 💡 Key Concepts

- **Microservices**: Each service is independent, has its own database
- **API Gateway**: Single entry point (port 80/8080), routes to services
- **JWT Auth**: Token-based, stateless authentication
- **WebSocket**: Real-time GPS streaming from riders to clients
- **Redis GEORADIUS**: Fast nearest-rider lookup for dispatch
- **Database-per-service**: Scales independently, no shared schema

---

## 🎯 Example Use Cases

1. **Food delivery** → Order (food items), Rider (delivery person), Payment (wallet or card)
2. **Parcel delivery** → Order (package), Rider (courier), Dispatch (auto-assign nearest)
3. **Ride-sharing** → Order (ride request), Rider (driver), Payment (wallet or card), Realtime (GPS tracking)
4. **Ambulance dispatch** → Order (emergency), Rider (ambulance), Dispatch (nearest station)

---

## 📞 Support

For issues or questions, check:
- `README.md` — Overview and project structure
- `docs/architecture.md` — System design details
- `docs/api-spec.md` — API documentation
- `MANIFEST.md` — Complete manifest and schema

Happy coding! 🚀
