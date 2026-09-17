package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/service"
)

type anonymousSessionCreator interface {
	CreateAnonymousSession(ctx context.Context, deviceID, timezone string) (service.SessionTokens, model.User, error)
}

type emailCodeRequester interface {
	RequestEmailCode(ctx context.Context, email string) error
}

type emailCodeVerifier interface {
	VerifyEmailCode(ctx context.Context, email, code, anonymousAccessToken string) (service.SessionTokens, bool, error)
}

type registrationCompleter interface {
	CompleteRegistration(ctx context.Context, userID, displayName string) error
}

type sessionRefresher interface {
	RefreshSession(ctx context.Context, refreshToken string) (service.SessionTokens, error)
}

type sessionLogger interface {
	Logout(ctx context.Context, refreshToken string) error
}

// AuthHandler exposes every auth endpoint. All six share one AuthService
// underneath (see cmd/api/main.go); the interfaces above just narrow what
// each handler method admits it depends on.
type AuthHandler struct {
	anonymous    anonymousSessionCreator
	codeRequests emailCodeRequester
	codeVerify   emailCodeVerifier
	registration registrationCompleter
	refresher    sessionRefresher
	logout       sessionLogger
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		anonymous:    authService,
		codeRequests: authService,
		codeVerify:   authService,
		registration: authService,
		refresher:    authService,
		logout:       authService,
	}
}

type sessionResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	// HasCompletedOnboarding lets the client decide whether to show
	// onboarding without a separate round trip — see
	// service.SessionTokens.HasCompletedOnboarding for why this can't
	// just be a client-local flag.
	HasCompletedOnboarding bool `json:"has_completed_onboarding"`
}

// --- POST /auth/anonymous ---

type createAnonymousSessionRequest struct {
	DeviceID string `json:"device_id"`
	// IANA timezone (e.g. "Europe/Istanbul"), used to reset daily quotas
	// at the user's local midnight. Optional; defaults to UTC. Ignored
	// for a returning device_id — timezone is set once, at creation.
	Timezone string `json:"timezone"`
}

type createAnonymousSessionResponse struct {
	sessionResponse
	UserID      string `json:"user_id"`
	IsAnonymous bool   `json:"is_anonymous"`
}

func (h *AuthHandler) CreateAnonymousSession(w http.ResponseWriter, r *http.Request) {
	var req createAnonymousSessionRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // body is optional; zero value just means "no device id, UTC"

	tokens, user, err := h.anonymous.CreateAnonymousSession(r.Context(), req.DeviceID, req.Timezone)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "oturum oluşturulamadı")
		return
	}

	writeJSON(w, http.StatusCreated, createAnonymousSessionResponse{
		sessionResponse: sessionResponse{
			AccessToken:            tokens.AccessToken,
			RefreshToken:           tokens.RefreshToken,
			HasCompletedOnboarding: tokens.HasCompletedOnboarding,
		},
		UserID:      user.ID,
		IsAnonymous: user.IsAnonymous,
	})
}

// --- POST /auth/email/request-code ---

type requestEmailCodeRequest struct {
	Email string `json:"email"`
}

// RequestEmailCode always responds with the same generic success body
// regardless of whether the email is registered, was just rate-limited,
// or a code was actually sent — see AuthService.RequestEmailCode's doc
// for why that's the whole point, not an oversight. Only a malformed
// email or a genuine infrastructure failure get a different response.
func (h *AuthHandler) RequestEmailCode(w http.ResponseWriter, r *http.Request) {
	var req requestEmailCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}

	err := h.codeRequests.RequestEmailCode(r.Context(), req.Email)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case errors.Is(err, service.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "geçersiz e-posta adresi")
	default:
		writeError(w, http.StatusInternalServerError, "kod gönderilemedi, lütfen tekrar deneyin")
	}
}

// --- POST /auth/email/verify-code ---

type verifyEmailCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
	// AccessToken is the caller's current anonymous session's access
	// token, if any — see AuthService.VerifyEmailCode's Durum A/B split.
	AccessToken string `json:"access_token"`
}

type verifyEmailCodeResponse struct {
	sessionResponse
	IsNewRegistration bool `json:"is_new_registration"`
}

func (h *AuthHandler) VerifyEmailCode(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}

	tokens, isNew, err := h.codeVerify.VerifyEmailCode(r.Context(), req.Email, req.Code, req.AccessToken)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, verifyEmailCodeResponse{
			sessionResponse: sessionResponse{
				AccessToken:            tokens.AccessToken,
				RefreshToken:           tokens.RefreshToken,
				HasCompletedOnboarding: tokens.HasCompletedOnboarding,
			},
			IsNewRegistration: isNew,
		})
	case errors.Is(err, service.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "geçersiz e-posta adresi")
	case errors.Is(err, service.ErrTooManyAttempts):
		writeError(w, http.StatusTooManyRequests, "çok fazla hatalı deneme, yeni kod isteyin")
	case errors.Is(err, service.ErrCodeExpired):
		writeError(w, http.StatusUnauthorized, "kodun süresi doldu, yeni kod isteyin")
	case errors.Is(err, service.ErrInvalidCode):
		writeError(w, http.StatusUnauthorized, "kod geçersiz")
	default:
		writeError(w, http.StatusInternalServerError, "doğrulama başarısız oldu")
	}
}

// --- POST /auth/register/complete ---

type completeRegistrationRequest struct {
	DisplayName string `json:"display_name"`
}

func (h *AuthHandler) CompleteRegistration(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "giriş gerekli")
		return
	}

	var req completeRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}

	err := h.registration.CompleteRegistration(r.Context(), userID, req.DisplayName)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case errors.Is(err, service.ErrInvalidDisplayName):
		writeError(w, http.StatusBadRequest, "geçersiz görünen ad")
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "kullanıcı bulunamadı")
	default:
		writeError(w, http.StatusInternalServerError, "profil güncellenemedi")
	}
}

// --- POST /auth/refresh ---

type refreshSessionRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	var req refreshSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}

	tokens, err := h.refresher.RefreshSession(r.Context(), req.RefreshToken)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, sessionResponse{
			AccessToken:            tokens.AccessToken,
			RefreshToken:           tokens.RefreshToken,
			HasCompletedOnboarding: tokens.HasCompletedOnboarding,
		})
	case errors.Is(err, service.ErrInvalidRefreshToken), errors.Is(err, service.ErrRefreshTokenExpired):
		writeError(w, http.StatusUnauthorized, "oturum geçersiz, tekrar giriş yapın")
	default:
		writeError(w, http.StatusInternalServerError, "oturum yenilenemedi")
	}
}

// --- POST /auth/logout ---

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "geçersiz istek gövdesi")
		return
	}

	if err := h.logout.Logout(r.Context(), req.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "çıkış yapılamadı")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
