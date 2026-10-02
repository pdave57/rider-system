package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/runns/payment-service/dto"
	"github.com/runns/payment-service/service"
	"github.com/runns/shared/middleware"
	"github.com/runns/shared/models"
	"github.com/runns/shared/utils"
)

type PaymentHandler struct{ svc *service.PaymentService }

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler { return &PaymentHandler{svc: svc} }

func (h *PaymentHandler) Initiate(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.InitiatePaymentRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.Initiate(middleware.UserIDFrom(r.Context()), req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.Created(w, "payment initiated", resp)
}

func (h *PaymentHandler) Verify(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.VerifyPaymentRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.Verify(req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "payment verified", resp)
}

func (h *PaymentHandler) GetByOrder(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	raw := strings.TrimPrefix(r.URL.Path, "/api/payments/order/")
	id, err := utils.ParseUint(raw, nil)
	if err != nil { utils.BadRequest(w, "invalid order id"); return }
	resp, err := h.svc.GetByOrder(id)
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "payment retrieved", resp)
}

func (h *PaymentHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	resp, err := h.svc.GetWallet(middleware.UserIDFrom(r.Context()))
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "wallet balance", resp)
}

func (h *PaymentHandler) TopupWallet(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.WalletTopupRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.TopupWallet(middleware.UserIDFrom(r.Context()), req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "wallet topped up", resp)
}

func (h *PaymentHandler) GetRiderWallet(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	riderID := middleware.UserIDFrom(r.Context())
	resp, err := h.svc.GetRiderWallet(riderID)
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "rider wallet balance", resp)
}

func (h *PaymentHandler) TransferToRiderWallet(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.WalletTransferRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.TransferToRiderWallet(middleware.UserIDFrom(r.Context()), req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "wallet transfer successful", resp)
}

func (h *PaymentHandler) VerifyNIN(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req struct {
		NIN string `json:"nin"`
	}
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	if req.NIN == "" { utils.BadRequest(w, "NIN required"); return }
	err := h.svc.VerifyNIN(middleware.UserIDFrom(r.Context()), req.NIN)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "NIN verified successfully", nil)
}

func (h *PaymentHandler) GetNINStatus(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	riderID := middleware.UserIDFrom(r.Context())
	resp, err := h.svc.GetNINStatus(riderID)
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "NIN status", resp)
}

func NewRouter(h *PaymentHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)
	client := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleClient)(next)) }
	rider := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleRider)(next)) }
	admin := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleAdmin)(next)) }
	mux := http.NewServeMux()
	mux.Handle("/api/payments", client(http.HandlerFunc(h.Initiate)))
	mux.Handle("/api/payments/verify", admin(http.HandlerFunc(h.Verify)))
	mux.Handle("/api/payments/wallet/topup", client(http.HandlerFunc(h.TopupWallet)))
	mux.Handle("/api/payments/wallet", client(http.HandlerFunc(h.GetWallet)))
	mux.Handle("/api/payments/wallet/transfer", client(http.HandlerFunc(h.TransferToRiderWallet)))
	mux.Handle("/api/payments/rider/wallet", rider(http.HandlerFunc(h.GetRiderWallet)))
	mux.Handle("/api/payments/rider/nin/verify", rider(http.HandlerFunc(h.VerifyNIN)))
	mux.Handle("/api/payments/rider/nin/status", rider(http.HandlerFunc(h.GetNINStatus)))
	mux.Handle("/api/payments/order/", auth(http.HandlerFunc(h.GetByOrder)))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "ok", nil) })
	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}
