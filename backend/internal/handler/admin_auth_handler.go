package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/service"
)

type adminLogger interface {
	Login(ctx context.Context, email, password string) (string, model.AdminUser, time.Time, error)
}

type adminSessionLogger interface {
	Logout(ctx context.Context, rawToken string) error
}

type adminGetter interface {
	CurrentAdmin(ctx context.Context, adminUserID string) (model.AdminUser, error)
}

// AdminAuthHandler exposes the admin dashboard's own login/logout/me
// endpoints — entirely separate from AuthHandler (public app auth). The
// admin/ web app is expected to store the session token itself (its own
// first-party cookie) and send it back as a Bearer token, the same way
// the iOS app does with access tokens.
type AdminAuthHandler struct {
	login  adminLogger
	logout adminSessionLogger
	get    adminGetter
}

func NewAdminAuthHandler(authService *service.AdminAuthService) *AdminAuthHandler {
	return &AdminAuthHandler{login: authService, logout: authService, get: authService}
}

type adminSessionResponse struct {
	SessionToken string          `json:"session_token"`
	ExpiresAt    time.Time       `json:"expires_at"`
	Admin        adminMeResponse `json:"admin"`
}

type adminMeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// --- POST /admin/auth/login ---

type adminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req adminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	token, admin, expiresAt, err := h.login.Login(r.Context(), req.Email, req.Password)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, adminSessionResponse{
			SessionToken: token,
			ExpiresAt:    expiresAt,
			Admin:        adminMeResponse{ID: admin.ID, Email: admin.Email},
		})
	case errors.Is(err, service.ErrInvalidAdminCredentials):
		writeError(w, r, http.StatusUnauthorized, ErrCodeAdminInvalidCredentials, "e-posta veya şifre hatalı")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLoginFailed, "giriş yapılamadı")
	}
}

// --- POST /admin/auth/logout ---

type adminLogoutRequest struct {
	SessionToken string `json:"session_token"`
}

func (h *AdminAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req adminLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	if err := h.logout.Logout(r.Context(), req.SessionToken); err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminLogoutFailed, "çıkış yapılamadı")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// --- GET /admin/auth/me ---

// Me lets the dashboard confirm who's logged in (and that its stored
// token is still valid) without hardcoding admin identity client-side —
// needed because a Next.js server component has no in-memory session
// state of its own between requests.
func (h *AdminAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := auth.AdminUserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, ErrCodeAdminUnauthorized, "admin girişi gerekli")
		return
	}

	admin, err := h.get.CurrentAdmin(r.Context(), adminUserID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminMeFailed, "admin bilgisi alınamadı")
		return
	}

	writeJSON(w, http.StatusOK, adminMeResponse{ID: admin.ID, Email: admin.Email})
}
