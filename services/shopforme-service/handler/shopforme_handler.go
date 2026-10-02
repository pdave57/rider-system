package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/runns/shopforme-service/dto"
	"github.com/runns/shopforme-service/service"
	"github.com/runns/shared/middleware"
	"github.com/runns/shared/models"
	"github.com/runns/shared/utils"
)

type ShopForMeHandler struct{ svc *service.ShopForMeService }

func NewShopForMeHandler(svc *service.ShopForMeService) *ShopForMeHandler { return &ShopForMeHandler{svc: svc} }

func (h *ShopForMeHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.CreateShopForMeRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.CreateRequest(middleware.UserIDFrom(r.Context()), req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.Created(w, "shop-for-me request created", resp)
}

func (h *ShopForMeHandler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	id, err := utils.PathID(r.URL.Path, "/api/shopforme/")
	if err != nil { utils.BadRequest(w, "invalid request id"); return }
	id = utils.ParseUintPath(strings.TrimPrefix(r.URL.Path, "/api/shopforme/"), "accept")
	if id == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "accept" && i > 0 {
				id, _ = utils.ParseUint(parts[i-1], nil)
				break
			}
		}
	}
	if id == 0 { utils.BadRequest(w, "invalid request id"); return }

	resp, err := h.svc.AcceptRequest(middleware.UserIDFrom(r.Context()), id)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "request accepted", resp)
}

func (h *ShopForMeHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	id, err := utils.PathID(r.URL.Path, "/api/shopforme/")
	if err != nil { utils.BadRequest(w, "invalid request id"); return }
	id = utils.ParseUintPath(strings.TrimPrefix(r.URL.Path, "/api/shopforme/"), "order")
	if id == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "order" && i > 0 {
				id, _ = utils.ParseUint(parts[i-1], nil)
				break
			}
		}
	}
	if id == 0 { utils.BadRequest(w, "invalid request id"); return }

	var req dto.PlaceOrderRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.PlaceOrder(middleware.UserIDFrom(r.Context()), id, req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.Created(w, "order placed", resp)
}

func (h *ShopForMeHandler) VerifyItems(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	id, err := utils.PathID(r.URL.Path, "/api/shopforme/")
	if err != nil { utils.BadRequest(w, "invalid request id"); return }
	id = utils.ParseUintPath(strings.TrimPrefix(r.URL.Path, "/api/shopforme/"), "verify")
	if id == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "verify" && i > 0 {
				id, _ = utils.ParseUint(parts[i-1], nil)
				break
			}
		}
	}
	if id == 0 { utils.BadRequest(w, "invalid request id"); return }

	var req dto.VerifyItemsRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	resp, err := h.svc.VerifyItems(middleware.UserIDFrom(r.Context()), id, req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "items verified", resp)
}

func (h *ShopForMeHandler) PayForOrder(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	id, err := utils.PathID(r.URL.Path, "/api/shopforme/")
	if err != nil { utils.BadRequest(w, "invalid request id"); return }
	id = utils.ParseUintPath(strings.TrimPrefix(r.URL.Path, "/api/shopforme/"), "pay")
	if id == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "pay" && i > 0 {
				id, _ = utils.ParseUint(parts[i-1], nil)
				break
			}
		}
	}
	if id == 0 { utils.BadRequest(w, "invalid request id"); return }

	resp, err := h.svc.PayForOrder(middleware.UserIDFrom(r.Context()), id)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "payment successful", resp)
}

func (h *ShopForMeHandler) DispatchOrder(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	id, err := utils.PathID(r.URL.Path, "/api/shopforme/")
	if err != nil { utils.BadRequest(w, "invalid request id"); return }
	id = utils.ParseUintPath(strings.TrimPrefix(r.URL.Path, "/api/shopforme/"), "dispatch")
	if id == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "dispatch" && i > 0 {
				id, _ = utils.ParseUint(parts[i-1], nil)
				break
			}
		}
	}
	if id == 0 { utils.BadRequest(w, "invalid request id"); return }

	resp, err := h.svc.DispatchOrder(middleware.UserIDFrom(r.Context()), id)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "order dispatched", resp)
}

func (h *ShopForMeHandler) DeliverOrder(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	id, err := utils.PathID(r.URL.Path, "/api/shopforme/")
	if err != nil { utils.BadRequest(w, "invalid request id"); return }
	id = utils.ParseUintPath(strings.TrimPrefix(r.URL.Path, "/api/shopforme/"), "deliver")
	if id == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "deliver" && i > 0 {
				id, _ = utils.ParseUint(parts[i-1], nil)
				break
			}
		}
	}
	if id == 0 { utils.BadRequest(w, "invalid request id"); return }

	resp, err := h.svc.DeliverOrder(middleware.UserIDFrom(r.Context()), id)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "order delivered", resp)
}

func (h *ShopForMeHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	id, err := utils.PathID(r.URL.Path, "/api/shopforme/")
	if err != nil { utils.BadRequest(w, "invalid request id"); return }
	id = utils.ParseUintPath(strings.TrimPrefix(r.URL.Path, "/api/shopforme/"), "cancel")
	if id == 0 {
		parts := strings.Split(r.URL.Path, "/")
		for i, p := range parts {
			if p == "cancel" && i > 0 {
				id, _ = utils.ParseUint(parts[i-1], nil)
				break
			}
		}
	}
	if id == 0 { utils.BadRequest(w, "invalid request id"); return }

	resp, err := h.svc.CancelRequest(middleware.UserIDFrom(r.Context()), id, middleware.RoleFrom(r.Context()) == models.RoleClient)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "request cancelled", resp)
}

func (h *ShopForMeHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	id, err := utils.PathID(r.URL.Path, "/api/shopforme/")
	if err != nil { utils.BadRequest(w, "invalid request id"); return }
	resp, err := h.svc.GetRequest(id)
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "request retrieved", resp)
}

func (h *ShopForMeHandler) ListMyRequests(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	userID := middleware.UserIDFrom(r.Context())
	var requests []dto.ShopForMeResponse
	var err error
	switch middleware.RoleFrom(r.Context()) {
	case models.RoleClient:
		requests, err = h.svc.ListByClient(userID)
	case models.RoleRider:
		requests, err = h.svc.ListByRider(userID)
	default:
		utils.BadRequest(w, "unauthorized"); return
	}
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.OK(w, "requests retrieved", requests)
}

func NewRouter(h *ShopForMeHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)
	client := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleClient)(next)) }
	rider := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleRider)(next)) }
	clientOrRider := func(next http.Handler) http.Handler { return auth(middleware.RequireRoles(models.RoleClient, models.RoleRider)(next)) }
	mux := http.NewServeMux()
	mux.Handle("/api/shopforme", client(http.HandlerFunc(h.CreateRequest)))
	mux.Handle("/api/shopforme/my", clientOrRider(http.HandlerFunc(h.ListMyRequests)))
	mux.Handle("/api/shopforme/", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/accept") && r.Method == http.MethodPost:
			rider(http.HandlerFunc(h.AcceptRequest)).ServeHTTP(w, r)
		case strings.HasSuffix(path, "/order") && r.Method == http.MethodPost:
			client(http.HandlerFunc(h.PlaceOrder)).ServeHTTP(w, r)
		case strings.HasSuffix(path, "/verify") && r.Method == http.MethodPost:
			rider(http.HandlerFunc(h.VerifyItems)).ServeHTTP(w, r)
		case strings.HasSuffix(path, "/pay") && r.Method == http.MethodPost:
			client(http.HandlerFunc(h.PayForOrder)).ServeHTTP(w, r)
		case strings.HasSuffix(path, "/dispatch") && r.Method == http.MethodPost:
			rider(http.HandlerFunc(h.DispatchOrder)).ServeHTTP(w, r)
		case strings.HasSuffix(path, "/deliver") && r.Method == http.MethodPost:
			rider(http.HandlerFunc(h.DeliverOrder)).ServeHTTP(w, r)
		case strings.HasSuffix(path, "/cancel") && r.Method == http.MethodPost:
			clientOrRider(http.HandlerFunc(h.CancelRequest)).ServeHTTP(w, r)
		case r.Method == http.MethodGet:
			h.GetRequest(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "ok", nil) })
	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}