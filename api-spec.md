# Rydex API Specification

All requests go through Nginx → API Gateway → Service.

Base URL: `http://localhost` (local) or `https://api.rydex.com` (production)

---

## Authentication

Most endpoints require JWT Bearer token:
```
Authorization: Bearer <token>
```

Get token via `/api/auth/login` or `/api/auth/register`.

---

## Common Response Format

```json
{
  "success": true,
  "message": "operation successful",
  "data": { ... }
}
```

Error:
```json
{
  "success": false,
  "error": "error message"
}
```

---

## Endpoints

### Auth Service

#### Register
```http
POST /api/auth/register
Content-Type: application/json

{
  "full_name": "John Doe",
  "email": "john@example.com",
  "phone": "08012345678",
  "password": "secret123",
  "role": "client",  // "client" | "rider" | "admin"
  "vehicle_type": "motorcycle",  // required if role=rider
  "vehicle_plate": "LAG-123XY"   // required if role=rider
}
```

Response (201):
```json
{
  "success": true,
  "message": "registration successful",
  "data": {
    "token": "eyJhbGciOi...",
    "user": {
      "id": 1,
      "full_name": "John Doe",
      "email": "john@example.com",
      "phone": "08012345678",
      "role": "client",
      "is_active": true
    }
  }
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "secret123"
}
```

---

### Order Service

#### Create Order (Client only)
```http
POST /api/orders
Authorization: Bearer <token>
Content-Type: application/json

{
  "pickup_address": "12 Adeola Street, Lagos",
  "pickup_latitude": 6.4281,
  "pickup_longitude": 3.4219,
  "dropoff_address": "45 Allen Avenue, Ikeja",
  "dropoff_latitude": 6.5958,
  "dropoff_longitude": 3.3678,
  "package_desc": "Documents",
  "weight_kg": 0.5,
  "notes": "Fragile"
}
```

Response (201):
```json
{
  "success": true,
  "message": "order created",
  "data": {
    "id": 42,
    "tracking_code": "RYX-A3F2B1C4-5678",
    "client_id": 1,
    "status": "pending",
    "pickup_address": "12 Adeola Street, Lagos",
    "dropoff_address": "45 Allen Avenue, Ikeja",
    "package_desc": "Documents",
    "weight_kg": 0.5,
    "delivery_fee": 1200.0,
    "notes": "Fragile"
  }
}
```

#### Track Order (Public)
```http
GET /api/orders/track/RYX-A3F2B1C4-5678
```

#### List Orders (Authenticated)
```http
GET /api/orders
Authorization: Bearer <token>
```

- Client sees own orders
- Rider sees assigned orders
- Admin sees all pending orders

#### Update Status (Admin/Rider)
```http
PATCH /api/orders/42/status
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "picked_up"  // "assigned" | "picked_up" | "delivered" | "cancelled"
}
```

---

### Dispatch Service

#### Assign Order to Rider (Admin)
```http
POST /api/dispatch
Authorization: Bearer <token>
Content-Type: application/json

{
  "order_id": 42,
  "rider_id": 5  // optional: if empty, auto-assigns nearest rider
}
```

#### Rider Respond (Rider only)
```http
POST /api/dispatch/12/respond
Authorization: Bearer <token>
Content-Type: application/json

{
  "accept": true  // true = accept, false = reject
}
```

#### List Pending Dispatches (Rider)
```http
GET /api/dispatch/pending
Authorization: Bearer <token>
```

---

### Payment Service

#### Initiate Payment (Client)
```http
POST /api/payments
Authorization: Bearer <token>
Content-Type: application/json

{
  "order_id": 42,
  "method": "wallet"  // "card" | "wallet" | "cash"
}
```

Response:
- Wallet: immediately deducted, `status=success`
- Card/Cash: `status=pending`, admin verifies later

#### Check Wallet Balance (Client)
```http
GET /api/payments/wallet
Authorization: Bearer <token>
```

#### Topup Wallet (Client)
```http
POST /api/payments/wallet/topup
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": 5000.0
}
```

---

### Realtime Service

#### WebSocket Connection (Authenticated)
```http
GET /ws
Upgrade: websocket
Authorization: Bearer <token>
```

After connection, riders send GPS updates:
```json
{
  "latitude": 6.5244,
  "longitude": 3.3792,
  "speed": 45.2,
  "heading": 180.0
}
```

Clients receive GPS updates:
```json
{
  "type": "gps_update",
  "payload": {
    "rider_id": 5,
    "order_id": 42,
    "latitude": 6.5244,
    "longitude": 3.3792,
    "speed": 45.2,
    "heading": 180.0,
    "timestamp": "2026-04-20T14:30:00Z"
  }
}
```

#### Get Latest Rider Location (HTTP fallback)
```http
GET /api/realtime/location/5
Authorization: Bearer <token>
```

---

### Rider Service

#### Get My Profile (Rider)
```http
GET /api/riders/me
Authorization: Bearer <token>
```

#### Toggle Availability (Rider)
```http
PATCH /api/riders/me/availability
Authorization: Bearer <token>
Content-Type: application/json

{
  "available": true
}
```

#### Update Location (Rider)
```http
PATCH /api/riders/me/location
Authorization: Bearer <token>
Content-Type: application/json

{
  "latitude": 6.5244,
  "longitude": 3.3792
}
```

---

## Status Codes

- `200` OK
- `201` Created
- `400` Bad Request (validation error)
- `401` Unauthorized (missing/invalid token)
- `403` Forbidden (insufficient role)
- `404` Not Found
- `409` Conflict (duplicate email, etc.)
- `500` Internal Server Error

---

## Role-Based Access

| Endpoint | Client | Rider | Admin |
|---|---|---|---|
| `POST /api/orders` | ✅ | ❌ | ❌ |
| `GET /api/orders` | Own | Assigned | All pending |
| `PATCH /api/orders/{id}/status` | ❌ | ✅ | ✅ |
| `POST /api/dispatch` | ❌ | ❌ | ✅ |
| `POST /api/dispatch/{id}/respond` | ❌ | ✅ | ❌ |
| `POST /api/payments` | ✅ | ❌ | ❌ |
| `POST /api/payments/verify` | ❌ | ❌ | ✅ |
| `GET /api/riders/me` | ❌ | ✅ | ❌ |

---

## Example Workflow

1. **Client registers**
   ```
   POST /api/auth/register → get JWT token
   ```

2. **Client creates order**
   ```
   POST /api/orders → order created, status=pending
   ```

3. **Client pays**
   ```
   POST /api/payments (method=wallet) → payment deducted
   ```

4. **Admin dispatches**
   ```
   POST /api/dispatch (rider_id=5) → order.status=assigned
   ```

5. **Rider accepts**
   ```
   POST /api/dispatch/12/respond (accept=true)
   ```

6. **Rider picks up**
   ```
   PATCH /api/orders/42/status (status=picked_up)
   ```

7. **Rider delivers**
   ```
   PATCH /api/orders/42/status (status=delivered)
   ```

8. **Client tracks in realtime**
   ```
   WS /ws → receives GPS updates from rider
   ```
