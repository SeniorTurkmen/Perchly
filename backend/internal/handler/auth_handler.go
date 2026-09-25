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

// CreateAnonymousSession godoc
// @Summary Anonim oturum oluştur
// @Description device_id'ye göre yeni bir anonim kullanıcı oluşturur veya var olanı döner. Gövde opsiyoneldir.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body createAnonymousSessionRequest false "device_id ve opsiyonel IANA timezone"
// @Success 201 {object} createAnonymousSessionResponse
// @Failure 500 {object} errorResponse
// @Router /auth/anonymous [post]
func (h *AuthHandler) CreateAnonymousSession(w http.ResponseWriter, r *http.Request) {
	var req createAnonymousSessionRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // body is optional; zero value just means "no device id, UTC"

	tokens, user, err := h.anonymous.CreateAnonymousSession(r.Context(), req.DeviceID, req.Timezone)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodeSessionCreateFailed, "oturum oluşturulamadı")
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
// RequestEmailCode godoc
// @Summary E-posta doğrulama kodu isteği
// @Description E-postaya bir doğrulama kodu gönderir. Kayıtlı olup olmadığına bakılmaksızın her zaman aynı genel başarı cevabını döner.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body requestEmailCodeRequest true "email"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /auth/email/request-code [post]
func (h *AuthHandler) RequestEmailCode(w http.ResponseWriter, r *http.Request) {
	var req requestEmailCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	err := h.codeRequests.RequestEmailCode(r.Context(), req.Email)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case errors.Is(err, service.ErrInvalidEmail):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidEmail, "geçersiz e-posta adresi")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeEmailCodeSendFailed, "kod gönderilemedi, lütfen tekrar deneyin")
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

// VerifyEmailCode godoc
// @Summary E-posta doğrulama kodunu doğrula
// @Description Kodu doğrular; access_token dolu gönderilirse mevcut anonim oturumu aynı user_id ile e-postalı hesaba yükseltir (Durum B), boşsa/olmayan bir e-postaysa yeni hesap açar (Durum A).
// @Tags auth
// @Accept json
// @Produce json
// @Param body body verifyEmailCodeRequest true "email, code, opsiyonel access_token"
// @Success 200 {object} verifyEmailCodeResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 429 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /auth/email/verify-code [post]
func (h *AuthHandler) VerifyEmailCode(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
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
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidEmail, "geçersiz e-posta adresi")
	case errors.Is(err, service.ErrTooManyAttempts):
		writeError(w, r, http.StatusTooManyRequests, ErrCodeTooManyAttempts, "çok fazla hatalı deneme, yeni kod isteyin")
	case errors.Is(err, service.ErrCodeExpired):
		writeError(w, r, http.StatusUnauthorized, ErrCodeVerificationExpired, "kodun süresi doldu, yeni kod isteyin")
	case errors.Is(err, service.ErrInvalidCode):
		writeError(w, r, http.StatusUnauthorized, ErrCodeInvalidCode, "kod geçersiz")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeVerificationFailed, "doğrulama başarısız oldu")
	}
}

// --- POST /auth/register/complete ---

type completeRegistrationRequest struct {
	DisplayName string `json:"display_name"`
}

// CompleteRegistration godoc
// @Summary Kayıt sonrası görünen adı belirle
// @Description users.display_name'i yalnızca bu uç yazar — onboarding'deki preferred_name (hitap) ile karıştırılmaz.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body completeRegistrationRequest true "display_name"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /auth/register/complete [post]
func (h *AuthHandler) CompleteRegistration(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, ErrCodeUnauthorized, "giriş gerekli")
		return
	}

	var req completeRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	err := h.registration.CompleteRegistration(r.Context(), userID, req.DisplayName)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]bool{"success": true})
	case errors.Is(err, service.ErrInvalidDisplayName):
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidDisplayName, "geçersiz görünen ad")
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(w, r, http.StatusNotFound, ErrCodeUserNotFound, "kullanıcı bulunamadı")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeProfileUpdateFailed, "profil güncellenemedi")
	}
}

// --- POST /auth/refresh ---

type refreshSessionRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshSession godoc
// @Summary Oturumu yenile
// @Description Refresh token ile yeni bir access/refresh token çifti üretir.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body refreshSessionRequest true "refresh_token"
// @Success 200 {object} sessionResponse
// @Failure 400 {object} errorResponse
// @Failure 401 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	var req refreshSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
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
		writeError(w, r, http.StatusUnauthorized, ErrCodeInvalidRefreshToken, "oturum geçersiz, tekrar giriş yapın")
	default:
		writeError(w, r, http.StatusInternalServerError, ErrCodeSessionRefreshFailed, "oturum yenilenemedi")
	}
}

// --- POST /auth/logout ---

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout godoc
// @Summary Çıkış yap
// @Description Verilen refresh token'ı iptal eder.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body logoutRequest true "refresh_token"
// @Success 200 {object} map[string]bool
// @Failure 500 {object} errorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	if err := h.logout.Logout(r.Context(), req.RefreshToken); err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodeLogoutFailed, "çıkış yapılamadı")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
