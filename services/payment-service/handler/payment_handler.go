package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/rydex/payment-service/dto"
	"github.com/rydex/payment-service/service"
	"github.com/rydex/shared/middleware"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/utils"
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

func NewRouter(h *PaymentHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)
	client := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleClient)(next)) }
	admin := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleAdmin)(next)) }
	mux := http.NewServeMux()
	mux.Handle("/api/payments", client(http.HandlerFunc(h.Initiate)))
	mux.Handle("/api/payments/verify", admin(http.HandlerFunc(h.Verify)))
	mux.Handle("/api/payments/wallet/topup", client(http.HandlerFunc(h.TopupWallet)))
	mux.Handle("/api/payments/wallet", client(http.HandlerFunc(h.GetWallet)))
	mux.Handle("/api/payments/order/", auth(http.HandlerFunc(h.GetByOrder)))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "ok", nil) })
	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}
