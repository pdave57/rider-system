package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/rydex/dispatch-service/dto"
	"github.com/rydex/dispatch-service/service"
	"github.com/rydex/shared/middleware"
	"github.com/rydex/shared/models"
	"github.com/rydex/shared/utils"
)

type DispatchHandler struct{ svc *service.DispatchService }

func NewDispatchHandler(svc *service.DispatchService) *DispatchHandler {
	return &DispatchHandler{svc: svc}
}

func (h *DispatchHandler) Assign(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.AssignRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.Assign(req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.Created(w, "order dispatched", resp)
}

func (h *DispatchHandler) Respond(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	riderID := middleware.UserIDFrom(r.Context())
	path := strings.TrimSuffix(r.URL.Path, "/respond")
	id, err := utils.PathID(path, "/api/dispatch/")
	if err != nil { utils.BadRequest(w, "invalid dispatch id"); return }
	var req dto.RespondRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.Respond(id, riderID, req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "response recorded", resp)
}

func (h *DispatchHandler) GetByOrder(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	id, err := utils.PathID(r.URL.Path, "/api/dispatch/order/")
	if err != nil { utils.BadRequest(w, "invalid order id"); return }
	resp, err := h.svc.GetByOrder(id)
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "dispatch retrieved", resp)
}

func (h *DispatchHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	items, _ := h.svc.ListPendingForRider(middleware.UserIDFrom(r.Context()))
	utils.OK(w, "pending dispatches", items)
}

func NewRouter(h *DispatchHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)
	admin := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleAdmin)(next)) }
	rider := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleRider)(next)) }
	mux := http.NewServeMux()
	mux.Handle("/api/dispatch", admin(http.HandlerFunc(h.Assign)))
	mux.Handle("/api/dispatch/pending", rider(http.HandlerFunc(h.ListPending)))
	mux.Handle("/api/dispatch/order/", auth(http.HandlerFunc(h.GetByOrder)))
	mux.Handle("/api/dispatch/", rider(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/respond") {
			h.Respond(w, r)
		} else {
			utils.NotFound(w, "not found")
		}
	})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "ok", nil) })
	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}
