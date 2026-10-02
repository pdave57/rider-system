package models

import (
	"time"
	"gorm.io/gorm"
)

type Role string
const (
	RoleClient Role = "client"
	RoleRider  Role = "rider"
	RoleAdmin  Role = "admin"
)

type OrderStatus string
const (
	OrderPending   OrderStatus = "pending"
	OrderAssigned  OrderStatus = "assigned"
	OrderPickedUp  OrderStatus = "picked_up"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

type PaymentStatus string
const (
	PaymentPending  PaymentStatus = "pending"
	PaymentSuccess  PaymentStatus = "success"
	PaymentFailed   PaymentStatus = "failed"
	PaymentRefunded PaymentStatus = "refunded"
)

type PaymentMethod string
const (
	PaymentCard   PaymentMethod = "card"
	PaymentWallet PaymentMethod = "wallet"
	PaymentCash   PaymentMethod = "cash"
)

type VehicleType string
const (
	VehicleBike       VehicleType = "bike"
	VehicleMotorcycle VehicleType = "motorcycle"
	VehicleCar        VehicleType = "car"
	VehicleVan        VehicleType = "van"
)

type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	FullName  string         `gorm:"type:varchar(120);not null" json:"full_name"`
	Email     string         `gorm:"type:varchar(160);uniqueIndex;not null" json:"email"`
	Phone     string         `gorm:"type:varchar(25);uniqueIndex;not null" json:"phone"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Role      Role           `gorm:"type:varchar(20);not null;default:'client'" json:"role"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	AvatarURL string         `gorm:"type:varchar(255)" json:"avatar_url,omitempty"`
	AvatarPublicID string    `gorm:"type:varchar(255)" json:"avatar_public_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	RiderProfile  *RiderProfile  `gorm:"foreignKey:UserID" json:"rider_profile,omitempty"`
	ClientProfile *ClientProfile `gorm:"foreignKey:UserID" json:"client_profile,omitempty"`
}

type RiderProfile struct {
	ID               uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint        `gorm:"uniqueIndex;not null" json:"user_id"`
	VehicleType      VehicleType `gorm:"type:varchar(20)" json:"vehicle_type"`
	VehiclePlate     string      `gorm:"type:varchar(20)" json:"vehicle_plate"`
	LicenseNumber    string      `gorm:"type:varchar(50)" json:"license_number"`
	NIN              string      `gorm:"type:varchar(20);uniqueIndex" json:"nin,omitempty"`
	NINVerified      bool        `gorm:"default:false" json:"nin_verified"`
	NINVerifiedAt    *time.Time  `json:"nin_verified_at,omitempty"`
	IsAvailable      bool        `gorm:"default:false" json:"is_available"`
	IsVerified       bool        `gorm:"default:false" json:"is_verified"`
	CurrentLatitude  float64     `gorm:"type:decimal(10,8)" json:"current_latitude"`
	CurrentLongitude float64     `gorm:"type:decimal(11,8)" json:"current_longitude"`
	Rating           float64     `gorm:"type:decimal(3,2);default:0.00" json:"rating"`
	TotalDeliveries  int         `gorm:"default:0" json:"total_deliveries"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type RiderWallet struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance       float64   `gorm:"type:decimal(12,2);default:0.00" json:"balance"`
	PendingBalance float64  `gorm:"type:decimal(12,2);default:0.00" json:"pending_balance"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ShopForMeRequestStatus string
const (
	ShopForMePending   ShopForMeRequestStatus = "pending"
	ShopForMeMatched   ShopForMeRequestStatus = "matched"
	ShopForMeAccepted  ShopForMeRequestStatus = "accepted"
	ShopForMeOrdering  ShopForMeRequestStatus = "ordering"
	ShopForMeShopping  ShopForMeRequestStatus = "shopping"
	ShopForMeVerifying ShopForMeRequestStatus = "verifying"
	ShopForMePaid      ShopForMeRequestStatus = "paid"
	ShopForMeDispatched ShopForMeRequestStatus = "dispatched"
	ShopForMeDelivered ShopForMeRequestStatus = "delivered"
	ShopForMeCancelled ShopForMeRequestStatus = "cancelled"
)

type ShopForMeRequest struct {
	ID              uint                   `gorm:"primaryKey;autoIncrement" json:"id"`
	ClientID        uint                   `gorm:"not null;index" json:"client_id"`
	RiderID         *uint                  `gorm:"index" json:"rider_id,omitempty"`
	Status          ShopForMeRequestStatus `gorm:"type:varchar(30);not null;default:'pending'" json:"status"`
	PickupAddress   string                 `gorm:"type:text;not null" json:"pickup_address"`
	PickupLatitude  float64                `gorm:"type:decimal(10,8)" json:"pickup_latitude"`
	PickupLongitude float64                `gorm:"type:decimal(11,8)" json:"pickup_longitude"`
	DropoffAddress  string                 `gorm:"type:text;not null" json:"dropoff_address"`
	DropoffLatitude float64                `gorm:"type:decimal(10,8)" json:"dropoff_latitude"`
	DropoffLongitude float64               `gorm:"type:decimal(11,8)" json:"dropoff_longitude"`
	Notes           string                 `gorm:"type:text" json:"notes,omitempty"`
	TotalAmount     float64                `gorm:"type:decimal(12,2);default:0.00" json:"total_amount"`
	ServiceFee      float64                `gorm:"type:decimal(10,2);default:0.00" json:"service_fee"`
	MatchedAt       *time.Time             `json:"matched_at,omitempty"`
	AcceptedAt      *time.Time             `json:"accepted_at,omitempty"`
	OrderPlacedAt   *time.Time             `json:"order_placed_at,omitempty"`
	PaidAt          *time.Time             `json:"paid_at,omitempty"`
	DispatchedAt    *time.Time             `json:"dispatched_at,omitempty"`
	DeliveredAt     *time.Time             `json:"delivered_at,omitempty"`
	CancelledAt     *time.Time             `json:"cancelled_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	DeletedAt       gorm.DeletedAt         `gorm:"index" json:"-"`

	Client  *User              `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	Rider   *User              `gorm:"foreignKey:RiderID" json:"rider,omitempty"`
	Items   []ShopForMeItem    `gorm:"foreignKey:RequestID" json:"items,omitempty"`
	Order   *ShopForMeOrder    `gorm:"foreignKey:RequestID" json:"order,omitempty"`
}

type ShopForMeItem struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID   uint      `gorm:"not null;index" json:"request_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Quantity    int       `gorm:"default:1" json:"quantity"`
	UnitPrice   float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice  float64   `gorm:"type:decimal(12,2)" json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ShopForMeOrder struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestID      uint      `gorm:"uniqueIndex;not null" json:"request_id"`
	RiderID        uint      `gorm:"not null;index" json:"rider_id"`
	ClientID       uint      `gorm:"not null;index" json:"client_id"`
	TotalAmount    float64   `gorm:"type:decimal(12,2);not null" json:"total_amount"`
	ServiceFee     float64   `gorm:"type:decimal(10,2);not null" json:"service_fee"`
	WalletAmount   float64   `gorm:"type:decimal(12,2);not null" json:"wallet_amount"`
	PaymentRef     string    `gorm:"type:varchar(100)" json:"payment_ref,omitempty"`
	Status         string    `gorm:"type:varchar(30);not null;default:'pending'" json:"status"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ClientProfile struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	DefaultAddress string    `gorm:"type:text" json:"default_address"`
	Latitude       float64   `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude      float64   `gorm:"type:decimal(11,8)" json:"longitude"`
	WalletBalance  float64   `gorm:"type:decimal(12,2);default:0.00" json:"wallet_balance"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Order struct {
	ID               uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	TrackingCode     string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"tracking_code"`
	ClientID         uint           `gorm:"not null;index;constraint:Off" json:"client_id"`
	RiderID          *uint          `gorm:"index;constraint:Off" json:"rider_id,omitempty"`
	Status           OrderStatus    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	PickupAddress    string         `gorm:"type:text;not null" json:"pickup_address"`
	PickupLatitude   float64        `gorm:"type:decimal(10,8)" json:"pickup_latitude"`
	PickupLongitude  float64        `gorm:"type:decimal(11,8)" json:"pickup_longitude"`
	DropoffAddress   string         `gorm:"type:text;not null" json:"dropoff_address"`
	DropoffLatitude  float64        `gorm:"type:decimal(10,8)" json:"dropoff_latitude"`
	DropoffLongitude float64        `gorm:"type:decimal(11,8)" json:"dropoff_longitude"`
	PackageDesc      string         `gorm:"type:varchar(255)" json:"package_desc"`
	WeightKg         float64        `gorm:"type:decimal(6,2)" json:"weight_kg"`
	DeliveryFee      float64        `gorm:"type:decimal(10,2)" json:"delivery_fee"`
	Notes            string         `gorm:"type:text" json:"notes,omitempty"`
	Items            []OrderItem    `gorm:"foreignKey:OrderID;constraint:Off" json:"items,omitempty"`
	PickedUpAt       *time.Time     `json:"picked_up_at,omitempty"`
	DeliveredAt      *time.Time     `json:"delivered_at,omitempty"`
	CancelledAt      *time.Time     `json:"cancelled_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	Client  *User    `gorm:"foreignKey:ClientID;constraint:Off" json:"client,omitempty"`
	Rider   *User    `gorm:"foreignKey:RiderID;constraint:Off" json:"rider,omitempty"`
	Payment *Payment `gorm:"foreignKey:OrderID;constraint:Off" json:"payment,omitempty"`
}
type OrderItem struct {
	ID          uint `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID     uint `gorm:"not null" json:"order_id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description,omitempty"`
	Quantity    int       `gorm:"default:1" json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Dispatch struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID    uint       `gorm:"uniqueIndex;not null" json:"order_id"`
	RiderID    uint       `gorm:"not null;index" json:"rider_id"`
	AssignedAt time.Time  `json:"assigned_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	RejectedAt *time.Time `json:"rejected_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Rider *User  `gorm:"foreignKey:RiderID" json:"rider,omitempty"`
}

type Payment struct {
	ID            uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID       uint          `gorm:"uniqueIndex;not null;constraint:Off" json:"order_id"`
	ClientID      uint          `gorm:"not null;index;constraint:Off" json:"client_id"`
	Amount        float64       `gorm:"type:decimal(12,2);not null" json:"amount"`
	Method        PaymentMethod `gorm:"type:varchar(20);not null" json:"method"`
	Status        PaymentStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Reference     string        `gorm:"type:varchar(100);uniqueIndex" json:"reference"`
	GatewayRef    string        `gorm:"type:varchar(200)" json:"gateway_ref,omitempty"`
	FailureReason string        `gorm:"type:varchar(255)" json:"failure_reason,omitempty"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}

type GPSEvent struct {
	RiderID   uint      `json:"rider_id"`
	OrderID   *uint     `json:"order_id,omitempty"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Speed     float64   `json:"speed,omitempty"`
	Heading   float64   `json:"heading,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
