package dto

import (
	"github.com/runns/shared/models"
	"time"
)

type InitiatePaymentRequest struct {
	OrderID uint                 `json:"order_id"`
	Method  models.PaymentMethod `json:"method"`
}

type VerifyPaymentRequest struct {
	Reference  string `json:"reference"`
	GatewayRef string `json:"gateway_ref"`
}

type WalletTopupRequest struct {
	Amount float64 `json:"amount"`
}

type WalletTransferRequest struct {
	RiderID uint    `json:"rider_id"`
	Amount  float64 `json:"amount"`
}

type PaymentResponse struct {
	ID         uint                 `json:"id"`
	OrderID    uint                 `json:"order_id"`
	ClientID   uint                 `json:"client_id"`
	Amount     float64              `json:"amount"`
	Method     models.PaymentMethod `json:"method"`
	Status     models.PaymentStatus `json:"status"`
	Reference  string               `json:"reference"`
	GatewayRef string               `json:"gateway_ref,omitempty"`
}

type WalletResponse struct {
	ClientID      uint    `json:"client_id"`
	WalletBalance float64 `json:"wallet_balance"`
}

type WalletTransferResponse struct {
	Reference         string  `json:"reference"`
	Amount            float64 `json:"amount"`
	ClientNewBalance  float64 `json:"client_new_balance"`
	RiderNewBalance   float64 `json:"rider_new_balance"`
}

type RiderWalletResponse struct {
	RiderID        uint    `json:"rider_id"`
	Balance        float64 `json:"balance"`
	PendingBalance float64 `json:"pending_balance"`
}

type NINStatusResponse struct {
	RiderID     uint       `json:"rider_id"`
	NINVerified bool       `json:"nin_verified"`
	NIN         string     `json:"nin,omitempty"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
}

type WalletTransferEvent struct {
	PaymentID  uint      `json:"payment_id"`
	ClientID   uint      `json:"client_id"`
	RiderID    uint      `json:"rider_id"`
	Amount     float64   `json:"amount"`
	Reference  string    `json:"reference"`
	Timestamp  time.Time `json:"timestamp"`
}

func ToPaymentResponse(p *models.Payment) PaymentResponse {
	return PaymentResponse{ID: p.ID, OrderID: p.OrderID, ClientID: p.ClientID,
		Amount: p.Amount, Method: p.Method, Status: p.Status,
		Reference: p.Reference, GatewayRef: p.GatewayRef}
}
