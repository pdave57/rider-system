package dto

import "github.com/rydex/shared/models"

type CreateOrderRequest struct {
	PickupAddress    string  `json:"pickup_address"`
	PickupLatitude   float64 `json:"pickup_latitude"`
	PickupLongitude  float64 `json:"pickup_longitude"`
	DropoffAddress   string  `json:"dropoff_address"`
	DropoffLatitude  float64 `json:"dropoff_latitude"`
	DropoffLongitude float64 `json:"dropoff_longitude"`
	PackageDesc      string  `json:"package_desc"`
	WeightKg         float64 `json:"weight_kg"`
	Notes            string  `json:"notes"`
}

type UpdateStatusRequest struct {
	Status models.OrderStatus `json:"status"`
}

type OrderResponse struct {
	ID             uint               `json:"id"`
	TrackingCode   string             `json:"tracking_code"`
	ClientID       uint               `json:"client_id"`
	RiderID        *uint              `json:"rider_id,omitempty"`
	Status         models.OrderStatus `json:"status"`
	PickupAddress  string             `json:"pickup_address"`
	DropoffAddress string             `json:"dropoff_address"`
	PackageDesc    string             `json:"package_desc"`
	WeightKg       float64            `json:"weight_kg"`
	DeliveryFee    float64            `json:"delivery_fee"`
	Notes          string             `json:"notes"`
}

func ToOrderResponse(o *models.Order) OrderResponse {
	return OrderResponse{
		ID: o.ID, TrackingCode: o.TrackingCode, ClientID: o.ClientID,
		RiderID: o.RiderID, Status: o.Status, PickupAddress: o.PickupAddress,
		DropoffAddress: o.DropoffAddress, PackageDesc: o.PackageDesc,
		WeightKg: o.WeightKg, DeliveryFee: o.DeliveryFee, Notes: o.Notes,
	}
}
