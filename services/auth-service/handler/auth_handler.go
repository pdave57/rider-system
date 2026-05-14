package handler

import (
	"net/http"
	"os"

	"github.com/runns/auth-service/dto"
	"github.com/runns/auth-service/service"
	"github.com/runns/shared/middleware"
	"github.com/runns/shared/utils"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.RegisterRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid request body"); return }
	resp, err := h.svc.Register(req)
	if err != nil { utils.BadRequest(w, err.Error()); return }
	utils.Created(w, "registration successful", resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.LoginRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid request body"); return }
	resp, err := h.svc.Login(req)
	if err != nil { utils.Unauthorized(w, err.Error()); return }
	utils.OK(w, "login successful", resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) { return }
	utils.OK(w, "authenticated", map[string]interface{}{
		"user_id": middleware.UserIDFrom(r.Context()),
		"role":    middleware.RoleFrom(r.Context()),
		"email":   middleware.EmailFrom(r.Context()),
	})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) { return }
	var req dto.ChangePasswordRequest
	if err := utils.DecodeJSON(r, &req); err != nil { utils.BadRequest(w, "invalid request body"); return }
	if err := h.svc.ChangePassword(middleware.UserIDFrom(r.Context()), req); err != nil {
		utils.BadRequest(w, err.Error()); return
	}
	utils.OK(w, "password changed", nil)
}

func NewRouter(h *AuthHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/register", h.Register)
	mux.HandleFunc("/api/auth/login", h.Login)
	mux.Handle("/api/auth/me", auth(http.HandlerFunc(h.Me)))
	mux.Handle("/api/auth/change-password", auth(http.HandlerFunc(h.ChangePassword)))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "ok", nil) })
	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}
