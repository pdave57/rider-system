// shared/rabbitmq/events.go
package rabbitmq

import "time"

// Event routing keys
const (
	// Order events
	RoutingKeyOrderCreated          = "order.created"
	RoutingKeyOrderStatusChanged    = "order.status.changed"
	RoutingKeyOrderReadyForDispatch = "order.ready_for_dispatch"

	// Payment events
	RoutingKeyPaymentConfirmed = "payment.confirmed"
	RoutingKeyPaymentFailed    = "payment.failed"

	// Dispatch events
	RoutingKeyDispatchAssigned = "dispatch.assigned"
	RoutingKeyDispatchAccepted = "dispatch.accepted"

	// Rider events
	RoutingKeyRiderAvailability = "rider.availability"

	// Auth events
	RoutingKeyUserRegistered = "user.registered"
	RoutingKeyUserRoleChanged = "user.role_changed"
	RoutingKeyUserDeactivated = "user.deactivated"
	RoutingKeyUserValidated  = "user.validated"
)

// BaseEvent is embedded in all events
type BaseEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
}

// UserRegisteredEvent published when a new user registers
type UserRegisteredEvent struct {
	BaseEvent
	Data UserRegisteredData `json:"data"`
}

type UserRegisteredData struct {
	UserID       uint      `json:"user_id"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	RegisteredAt time.Time `json:"registered_at"`

	// Rider-specific fields
	VehicleType   string `json:"vehicle_type,omitempty"`
	VehiclePlate  string `json:"vehicle_plate,omitempty"`
	LicenseNumber string `json:"license_number,omitempty"`

	// Client-specific fields
	DefaultAddress string  `json:"default_address,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
}

// UserRoleChangedEvent published when user role changes
type UserRoleChangedEvent struct {
	BaseEvent
	Data UserRoleChangedData `json:"data"`
}

type UserRoleChangedData struct {
	UserID    uint      `json:"user_id"`
	OldRole   string    `json:"old_role"`
	NewRole   string    `json:"new_role"`
	ChangedBy uint      `json:"changed_by"`
	ChangedAt time.Time `json:"changed_at"`
}

// UserDeactivatedEvent published when user is deactivated
type UserDeactivatedEvent struct {
	BaseEvent
	Data UserDeactivatedData `json:"data"`
}

type UserDeactivatedData struct {
	UserID        uint      `json:"user_id"`
	Reason        string    `json:"reason"`
	DeactivatedBy uint      `json:"deactivated_by"`
	DeactivatedAt time.Time `json:"deactivated_at"`
}

// Order Events
type OrderCreatedEvent struct {
    OrderID       uint    `json:"order_id"`
    TrackingCode  string  `json:"tracking_code"`
    ClientID      uint    `json:"client_id"`
    PickupLat     float64 `json:"pickup_latitude"`
    PickupLng     float64 `json:"pickup_longitude"`
    DropoffLat    float64 `json:"dropoff_latitude"`
    DropoffLng    float64 `json:"dropoff_longitude"`
    DeliveryFee   float64 `json:"delivery_fee"`
    WeightKg      float64 `json:"weight_kg"`
    Timestamp     time.Time `json:"timestamp"`
}

type OrderStatusChangedEvent struct {
    OrderID      uint      `json:"order_id"`
    TrackingCode string    `json:"tracking_code"`
    OldStatus    string    `json:"old_status"`
    NewStatus    string    `json:"new_status"`
    RiderID      *uint     `json:"rider_id,omitempty"`
    UpdatedBy    uint      `json:"updated_by"`
    Timestamp    time.Time `json:"timestamp"`
}

// Payment Events
type PaymentConfirmedEvent struct {
    PaymentID     uint      `json:"payment_id"`
    OrderID       uint      `json:"order_id"`
    ClientID      uint      `json:"client_id"`
    Amount        float64   `json:"amount"`
    Method        string    `json:"method"`
    Status        string    `json:"status"`
    Reference     string    `json:"reference"`
    ConfirmedAt   time.Time `json:"confirmed_at"`
}

type PaymentFailedEvent struct {
    PaymentID     uint      `json:"payment_id"`
    OrderID       uint      `json:"order_id"`
    ClientID      uint      `json:"client_id"`
    Amount        float64   `json:"amount"`
    Method        string    `json:"method"`
    Reason        string    `json:"reason"`
    Timestamp     time.Time `json:"timestamp"`
}

// Dispatch Events
type OrderReadyForDispatchEvent struct {
    OrderID       uint      `json:"order_id"`
    TrackingCode  string    `json:"tracking_code"`
    PickupLat     float64   `json:"pickup_latitude"`
    PickupLng     float64   `json:"pickup_longitude"`
    DropoffLat    float64   `json:"dropoff_latitude"`
    DropoffLng    float64   `json:"dropoff_longitude"`
    Timestamp     time.Time `json:"timestamp"`
}

type DispatchAssignedEvent struct {
    DispatchID    uint      `json:"dispatch_id"`
    OrderID       uint      `json:"order_id"`
    RiderID       uint      `json:"rider_id"`
    AssignedAt    time.Time `json:"assigned_at"`
}

type DispatchAcceptedEvent struct {
    DispatchID    uint      `json:"dispatch_id"`
    OrderID       uint      `json:"order_id"`
    RiderID       uint      `json:"rider_id"`
    AcceptedAt    time.Time `json:"accepted_at"`
}

// Rider Events
type RiderAvailabilityEvent struct {
    RiderID       uint      `json:"rider_id"`
    IsAvailable   bool      `json:"is_available"`
    Latitude      float64   `json:"latitude,omitempty"`
    Longitude     float64   `json:"longitude,omitempty"`
    Timestamp     time.Time `json:"timestamp"`
}

// Auth Events
type UserValidationEvent struct {
    UserID        uint      `json:"user_id"`
    Email         string    `json:"email"`
    Role          string    `json:"role"`
    IsActive      bool      `json:"is_active"`
    ValidatedAt   time.Time `json:"validated_at"`
}