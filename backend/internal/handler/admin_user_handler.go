package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
	"perchly-backend/internal/service"
)

type adminUserLister interface {
	List(ctx context.Context, search string, limit, offset int) ([]model.User, int, error)
}

type adminUserGetter interface {
	Get(ctx context.Context, userID string) (service.AdminUserDetail, error)
}

type adminQuotaSetter interface {
	SetQuota(ctx context.Context, adminUserID, userID, personaID string, dailyLimit int) (model.UserQuota, error)
}

type adminCreditSetter interface {
	SetCredits(ctx context.Context, adminUserID, userID string, amount int) (int, error)
}

// AdminUserHandler exposes the admin dashboard's user endpoints —
// list/inspect app users and override their quota/credits.
type AdminUserHandler struct {
	list    adminUserLister
	get     adminUserGetter
	quota   adminQuotaSetter
	credits adminCreditSetter
}

func NewAdminUserHandler(users *service.AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{list: users, get: users, quota: users, credits: users}
}

type adminUsersListResponse struct {
	Users []model.User `json:"users"`
	Total int          `json:"total"`
}

// --- GET /admin/users ---

func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	users, total, err := h.list.List(r.Context(), q.Get("search"), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeAdminUsersListFailed, "kullanıcılar listelenemedi")
		return
	}
	writeJSON(w, http.StatusOK, adminUsersListResponse{Users: users, Total: total})
}

// --- GET /admin/users/{id} ---

func (h *AdminUserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidUserID, "geçersiz kullanıcı id")
		return
	}

	detail, err := h.get.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeUserNotFound, "kullanıcı bulunamadı")
			return
		}
		writeError(w, http.StatusInternalServerError, ErrCodeAdminUserFetchFailed, "kullanıcı getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// --- PATCH /admin/users/{id}/quota ---

type adminSetQuotaRequest struct {
	PersonaID  string `json:"persona_id"`
	DailyLimit int    `json:"daily_limit"`
}

func (h *AdminUserHandler) SetQuota(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(userID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidUserID, "geçersiz kullanıcı id")
		return
	}

	var req adminSetQuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}
	if _, err := uuid.Parse(req.PersonaID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidPersonaID, "geçersiz persona id")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	quota, err := h.quota.SetQuota(r.Context(), adminUserID, userID, req.PersonaID, req.DailyLimit)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, quota)
	case errors.Is(err, service.ErrInvalidDailyLimit):
		writeError(w, http.StatusBadRequest, ErrCodeInvalidDailyLimit, "günlük limit 0 veya daha büyük olmalı")
	case errors.Is(err, repository.ErrPersonaNotFound):
		writeError(w, http.StatusNotFound, ErrCodePersonaNotFound, "persona bulunamadı")
	default:
		writeError(w, http.StatusInternalServerError, ErrCodeAdminQuotaUpdateFailed, "kota güncellenemedi")
	}
}

// --- PATCH /admin/users/{id}/credits ---

type adminSetCreditsRequest struct {
	Amount int `json:"amount"`
}

func (h *AdminUserHandler) SetCredits(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(userID); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidUserID, "geçersiz kullanıcı id")
		return
	}

	var req adminSetCreditsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequestBody, "geçersiz istek gövdesi")
		return
	}

	adminUserID, _ := auth.AdminUserIDFromContext(r.Context())

	balance, err := h.credits.SetCredits(r.Context(), adminUserID, userID, req.Amount)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, map[string]int{"credits": balance})
	case errors.Is(err, service.ErrInvalidCreditAmount):
		writeError(w, http.StatusBadRequest, ErrCodeInvalidCreditAmount, "kredi miktarı 0 veya daha büyük olmalı")
	default:
		writeError(w, http.StatusInternalServerError, ErrCodeAdminCreditsUpdateFailed, "kredi güncellenemedi")
	}
}
