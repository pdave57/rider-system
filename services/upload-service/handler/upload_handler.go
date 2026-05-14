package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/runns/shared/middleware"
	"github.com/runns/shared/utils"
	"github.com/runns/upload-service/service"
)

type UploadHandler struct {
	svc *service.CloudinaryService
}

func NewUploadHandler(svc *service.CloudinaryService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// POST /api/upload/avatar
// Content-Type: multipart/form-data
// Field name: "avatar"
func (h *UploadHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodPost) {
		return
	}

	userID := middleware.UserIDFrom(r.Context())

	// Parse multipart form — limit to 10MB in memory
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.BadRequest(w, "failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		utils.BadRequest(w, "avatar field is required (multipart/form-data, field name: avatar)")
		return
	}
	defer file.Close()

	result, err := h.svc.UploadAvatar(r.Context(), userID, file, header)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.OK(w, "avatar uploaded successfully", map[string]interface{}{
		"public_id":  result.PublicID,
		"secure_url": result.SecureURL,
		"width":      result.Width,
		"height":     result.Height,
		"format":     result.Format,
		"bytes":      result.Bytes,
	})
}

// DELETE /api/upload/avatar
func (h *UploadHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodDelete) {
		return
	}
	userID := middleware.UserIDFrom(r.Context())
	if err := h.svc.DeleteAvatar(r.Context(), userID); err != nil {
		utils.BadRequest(w, err.Error())
		return
	}
	utils.OK(w, "avatar deleted", nil)
}

// GET /api/upload/avatar/me
func (h *UploadHandler) GetMyAvatar(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) {
		return
	}
	userID := middleware.UserIDFrom(r.Context())
	user, err := h.svc.GetProfile(userID)
	if err != nil {
		utils.NotFound(w, err.Error())
		return
	}

	// Return multiple size variants using Cloudinary transformations
	utils.OK(w, "avatar retrieved", map[string]interface{}{
		"public_id":   user.AvatarPublicID,
		"original":    user.AvatarURL,
		"thumbnail":   h.svc.GetAvatarURL(user.AvatarPublicID, 100, 100),
		"medium":      h.svc.GetAvatarURL(user.AvatarPublicID, 400, 400),
		"large":       h.svc.GetAvatarURL(user.AvatarPublicID, 800, 800),
	})
}

// GET /api/upload/avatar/{userID}   — public, view any user's avatar
func (h *UploadHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	if !utils.AllowMethods(w, r, http.MethodGet) {
		return
	}
	raw := strings.TrimPrefix(r.URL.Path, "/api/upload/avatar/")
	id, err := strconv.ParseUint(strings.Trim(raw, "/"), 10, 64)
	if err != nil {
		utils.BadRequest(w, "invalid user id")
		return
	}

	user, err := h.svc.GetProfile(uint(id))
	if err != nil {
		utils.NotFound(w, err.Error())
		return
	}

	utils.OK(w, "avatar retrieved", map[string]interface{}{
		"user_id":   user.ID,
		"full_name": user.FullName,
		"public_id": user.AvatarPublicID,
		"thumbnail": h.svc.GetAvatarURL(user.AvatarPublicID, 100, 100),
		"medium":    h.svc.GetAvatarURL(user.AvatarPublicID, 400, 400),
		"large":     h.svc.GetAvatarURL(user.AvatarPublicID, 800, 800),
	})
}

func NewRouter(h *UploadHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)

	mux := http.NewServeMux()

	// Authenticated — own avatar
	mux.Handle("/api/upload/avatar/me", auth(http.HandlerFunc(h.GetMyAvatar)))
	mux.Handle("/api/upload/avatar", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.UploadAvatar(w, r)
		case http.MethodDelete:
			h.DeleteAvatar(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	// Public — view any user's avatar by ID
	mux.HandleFunc("/api/upload/avatar/", h.GetAvatar)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.OK(w, "ok", nil)
	})

	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}