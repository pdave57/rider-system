package handler

import (
	"net/http"
	"os"

	"github.com/runns/rider-service/service"
	"github.com/runns/shared/middleware"
	"github.com/runns/shared/models"
	"github.com/runns/shared/utils"
)

type RiderHandler struct{ svc *service.RiderService }

func NewRiderHandler(svc *service.RiderService) *RiderHandler { return &RiderHandler{svc: svc} }

func (h *RiderHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	user, err := h.svc.GetProfile(middleware.UserIDFrom(r.Context()))
	if err != nil { utils.NotFound(w, err.Error()); return }
	utils.OK(w, "profile retrieved", user)
}

func (h *RiderHandler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPatch) { return }
	var req struct { Available bool `json:"available"` }
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	if err := h.svc.UpdateAvailability(middleware.UserIDFrom(r.Context()), req.Available); err != nil {
		utils.BadRequest(w, err.Error()); return
	}
	utils.OK(w, "availability updated", nil)
}

func (h *RiderHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPatch) { return }
	var req struct { Latitude float64 `json:"latitude"`; Longitude float64 `json:"longitude"` }
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid body"); return }
	if err := h.svc.UpdateLocation(middleware.UserIDFrom(r.Context()), req.Latitude, req.Longitude); err != nil {
		utils.BadRequest(w, err.Error()); return
	}
	utils.OK(w, "location updated", nil)
}

func (h *RiderHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	riders, _ := h.svc.ListAvailable()
	utils.OK(w, "available riders", riders)
}

func NewRouter(h *RiderHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)
	rider := middleware.RequireRoles(models.RoleRider)
	mux := http.NewServeMux()
	mux.Handle("/api/riders/me", auth(rider(http.HandlerFunc(h.GetMyProfile))))
	mux.Handle("/api/riders/me/availability", auth(rider(http.HandlerFunc(h.UpdateAvailability))))
	mux.Handle("/api/riders/me/location", auth(rider(http.HandlerFunc(h.UpdateLocation))))
	mux.Handle("/api/riders/available", auth(http.HandlerFunc(h.ListAvailable)))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "ok", nil) })
	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}
