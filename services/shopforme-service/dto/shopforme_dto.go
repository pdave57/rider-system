package dto

import (
	"github.com/runns/shared/models"
	"time"
)

type CreateShopForMeRequest struct {
	PickupAddress   string  `json:"pickup_address"`
	PickupLatitude  float64 `json:"pickup_latitude"`
	PickupLongitude float64 `json:"pickup_longitude"`
	DropoffAddress  string  `json:"dropoff_address"`
	DropoffLatitude float64 `json:"dropoff_latitude"`
	DropoffLongitude float64 `json:"dropoff_longitude"`
	Notes           string  `json:"notes,omitempty"`
}

type PlaceOrderRequest struct {
	Items []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
}

type VerifyItemsRequest struct {
	Verified bool   `json:"verified"`
	Notes    string `json:"notes,omitempty"`
}

type ShopForMeResponse struct {
	ID              uint                       `json:"id"`
	ClientID        uint                       `json:"client_id"`
	RiderID         *uint                      `json:"rider_id,omitempty"`
	Status          models.ShopForMeRequestStatus `json:"status"`
	PickupAddress   string                     `json:"pickup_address"`
	PickupLatitude  float64                    `json:"pickup_latitude"`
	PickupLongitude float64                    `json:"pickup_longitude"`
	DropoffAddress  string                     `json:"dropoff_address"`
	DropoffLatitude float64                    `json:"dropoff_latitude"`
	DropoffLongitude float64                   `json:"dropoff_longitude"`
	Notes           string                     `json:"notes,omitempty"`
	TotalAmount     float64                    `json:"total_amount"`
	ServiceFee      float64                    `json:"service_fee"`
	Items           []ShopForMeItemResponse    `json:"items,omitempty"`
	Order           *ShopForMeOrderResponse    `json:"order,omitempty"`
	MatchedAt       *time.Time                 `json:"matched_at,omitempty"`
	AcceptedAt      *time.Time                 `json:"accepted_at,omitempty"`
	OrderPlacedAt   *time.Time                 `json:"order_placed_at,omitempty"`
	PaidAt          *time.Time                 `json:"paid_at,omitempty"`
	DispatchedAt    *time.Time                 `json:"dispatched_at,omitempty"`
	DeliveredAt     *time.Time                 `json:"delivered_at,omitempty"`
	CancelledAt     *time.Time                 `json:"cancelled_at,omitempty"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
}

type ShopForMeItemResponse struct {
	ID          uint    `json:"id"`
	RequestID   uint    `json:"request_id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ShopForMeOrderResponse struct {
	ID           uint      `json:"id"`
	RequestID    uint      `json:"request_id"`
	RiderID      uint      `json:"rider_id"`
	ClientID     uint      `json:"client_id"`
	TotalAmount  float64   `json:"total_amount"`
	ServiceFee   float64   `json:"service_fee"`
	WalletAmount float64   `json:"wallet_amount"`
	PaymentRef   string    `json:"payment_ref,omitempty"`
	Status       string    `json:"status"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type RiderMatchResponse struct {
	RiderID   uint    `json:"rider_id"`
	Name      string  `json:"name"`
	Vehicle   string  `json:"vehicle"`
	Rating    float64 `json:"rating"`
	Distance  float64 `json:"distance_km"`
}

func ToShopForMeResponse(r *models.ShopForMeRequest) ShopForMeResponse {
	items := make([]ShopForMeItemResponse, len(r.Items))
	for i, item := range r.Items {
		items[i] = ShopForMeItemResponse{
			ID:          item.ID,
			RequestID:   item.RequestID,
			Name:        item.Name,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	var orderResp *ShopForMeOrderResponse
	if r.Order != nil {
		orderResp = &ShopForMeOrderResponse{
			ID:           r.Order.ID,
			RequestID:    r.Order.RequestID,
			RiderID:      r.Order.RiderID,
			ClientID:     r.Order.ClientID,
			TotalAmount:  r.Order.TotalAmount,
			ServiceFee:   r.Order.ServiceFee,
			WalletAmount: r.Order.WalletAmount,
			PaymentRef:   r.Order.PaymentRef,
			Status:       r.Order.Status,
			VerifiedAt:   r.Order.VerifiedAt,
			PaidAt:       r.Order.PaidAt,
			CreatedAt:    r.Order.CreatedAt,
			UpdatedAt:    r.Order.UpdatedAt,
		}
	}

	return ShopForMeResponse{
		ID:              r.ID,
		ClientID:        r.ClientID,
		RiderID:         r.RiderID,
		Status:          r.Status,
		PickupAddress:   r.PickupAddress,
		PickupLatitude:  r.PickupLatitude,
		PickupLongitude: r.PickupLongitude,
		DropoffAddress:  r.DropoffAddress,
		DropoffLatitude: r.DropoffLatitude,
		DropoffLongitude: r.DropoffLongitude,
		Notes:           r.Notes,
		TotalAmount:     r.TotalAmount,
		ServiceFee:      r.ServiceFee,
		Items:           items,
		Order:           orderResp,
		MatchedAt:       r.MatchedAt,
		AcceptedAt:      r.AcceptedAt,
		OrderPlacedAt:   r.OrderPlacedAt,
		PaidAt:          r.PaidAt,
		DispatchedAt:    r.DispatchedAt,
		DeliveredAt:     r.DeliveredAt,
		CancelledAt:     r.CancelledAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}