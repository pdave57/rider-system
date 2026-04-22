package dto

import "github.com/rydex/shared/models"

type AssignRequest struct {
	OrderID uint `json:"order_id"`
	RiderID uint `json:"rider_id"`
}

type RespondRequest struct {
	Accept bool `json:"accept"`
}

type DispatchResponse struct {
	ID      uint   `json:"id"`
	OrderID uint   `json:"order_id"`
	RiderID uint   `json:"rider_id"`
	Status  string `json:"status"`
}

func ToDispatchResponse(d *models.Dispatch) DispatchResponse {
	status := "pending"
	if d.AcceptedAt != nil { status = "accepted" }
	if d.RejectedAt != nil { status = "rejected" }
	return DispatchResponse{ID: d.ID, OrderID: d.OrderID, RiderID: d.RiderID, Status: status}
}
