package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/runns/order-service/dto"
	"github.com/runns/order-service/service"
	"github.com/runns/shared/middleware"
	"github.com/runns/shared/models"
	"github.com/runns/shared/utils"
)

//initiate dependency injection
type OrderHandler struct{ svc *service.OrderService }

//dependency injection
func NewOrderHandler(svc *service.OrderService) *OrderHandler { return &OrderHandler{svc: svc} }

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateOrderRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.Create(middleware.UserIDFrom(r.Context()), req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.Created(w, "order created", resp)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	var orders []dto.OrderResponse
	switch middleware.RoleFrom(r.Context()) {
	case models.RoleClient:
		orders, _ = h.svc.ListByClient(userID)
	case models.RoleRider:
		orders, _ = h.svc.ListByRider(userID)
	default:
		orders, _ = h.svc.ListPending()
	}
	utils.OK(w, "orders retrieved", orders)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := utils.PathID(r.URL.Path, "/api/orders/")
	if err != nil { utils.BadRequest(w, "invalid order id"); return }
	resp, err := h.svc.GetByID(id)
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "order retrieved", resp)
}

func (h *OrderHandler) Track(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/api/orders/track/")
	resp, err := h.svc.GetByTracking(code)
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "order found", resp)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/status")
	id, err := utils.PathID(path, "/api/orders/")
	if err != nil { utils.BadRequest(w, "invalid order id"); return }
	var req dto.UpdateStatusRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.UpdateStatus(id, req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "status updated", resp)
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	clientID := middleware.UserIDFrom(r.Context())
	id, err := utils.PathID(r.URL.Path, "/api/orders/")
	if err != nil { utils.BadRequest(w, "invalid order id"); return }
	if err := h.svc.Cancel(id, clientID); err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "order cancelled", nil)
}

func NewRouter(h *OrderHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)
	adminOrRider := middleware.RequireRoles(models.RoleAdmin, models.RoleRider)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders/track/", h.Track)
	mux.Handle("/api/orders", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.List(w, r)
		case http.MethodPost:
			middleware.RequireRoles(models.RoleClient)(http.HandlerFunc(h.Create)).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/orders/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/status") && r.Method == http.MethodPatch:
			adminOrRider(http.HandlerFunc(h.UpdateStatus)).ServeHTTP(w, r)
		case r.Method == http.MethodGet:
			h.GetByID(w, r)
		case r.Method == http.MethodDelete:
			middleware.RequireRoles(models.RoleClient)(http.HandlerFunc(h.Cancel)).ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "ok", nil) })
	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}
