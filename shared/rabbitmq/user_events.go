package rabbitmq

import "time"

// Event types for routing keys
const (
    UserRegistered   = "user.registered"
    UserUpdated      = "user.updated"
    UserRoleChanged  = "user.role_changed"
    UserDeactivated  = "user.deactivated"
    UserActivated    = "user.activated"
    PasswordChanged  = "user.password_changed"
)

// BaseEvent is embedded in all events
type BaseEvent struct {
    EventID     string    `json:"event_id"`
    EventType   string    `json:"event_type"`
    Timestamp   time.Time `json:"timestamp"`
    Service     string    `json:"service"`
    Version     string    `json:"version"`
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