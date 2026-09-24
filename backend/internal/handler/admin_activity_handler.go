package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"perchly-backend/internal/model"
	"perchly-backend/internal/service"
)

type adminActivityLister interface {
	List(ctx context.Context, adminUserID *string, limit, offset int) ([]model.AdminAuditLog, int, error)
}

// AdminActivityHandler exposes the admin dashboard's own activity log
// (admin_audit_log) — who among the admins did what, when. Distinct
// from AdminLogHandler, which lists app-user HTTP requests.
type AdminActivityHandler struct {
	list adminActivityLister
}

func NewAdminActivityHandler(activity *service.AdminActivityService) *AdminActivityHandler {
	return &AdminActivityHandler{list: activity}
}

type adminActivityListResponse struct {
	Entries []model.AdminAuditLog `json:"entries"`
	Total   int                   `json:"total"`
}

// --- GET /admin/activity ---

func (h *AdminActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	var adminUserID *string
	if raw := q.Get("admin_user_id"); raw != "" {
		if _, err := uuid.Parse(raw); err != nil {
			writeError(w, http.StatusBadRequest, ErrCodeInvalidUserID, "geçersiz admin id")
			return
		}
		adminUserID = &raw
	}

	entries, total, err := h.list.List(r.Context(), adminUserID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeAdminActivityListFailed, "admin işlem geçmişi listelenemedi")
		return
	}
	writeJSON(w, http.StatusOK, adminActivityListResponse{Entries: entries, Total: total})
}
