package dto

import "github.com/runns/shared/models"

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

func ToPaymentResponse(p *models.Payment) PaymentResponse {
	return PaymentResponse{ID: p.ID, OrderID: p.OrderID, ClientID: p.ClientID,
		Amount: p.Amount, Method: p.Method, Status: p.Status,
		Reference: p.Reference, GatewayRef: p.GatewayRef}
}
