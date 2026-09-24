package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"perchly-backend/internal/model"
	"perchly-backend/internal/service"
)

type adminLogLister interface {
	List(ctx context.Context, search string, statusMin int, userID *string, limit, offset int) ([]model.RequestLogEntry, int, error)
}

type AdminLogHandler struct {
	list adminLogLister
}

func NewAdminLogHandler(logs *service.AdminLogService) *AdminLogHandler {
	return &AdminLogHandler{list: logs}
}

type adminLogsListResponse struct {
	Logs  []model.RequestLogEntry `json:"logs"`
	Total int                     `json:"total"`
}

// --- GET /admin/logs ---

func (h *AdminLogHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	statusMin, _ := strconv.Atoi(q.Get("status_min"))

	var userID *string
	if raw := q.Get("user_id"); raw != "" {
		if _, err := uuid.Parse(raw); err != nil {
			writeError(w, http.StatusBadRequest, ErrCodeInvalidUserID, "geçersiz kullanıcı id")
			return
		}
		userID = &raw
	}

	logs, total, err := h.list.List(r.Context(), q.Get("search"), statusMin, userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, ErrCodeAdminLogsListFailed, "loglar listelenemedi")
		return
	}
	writeJSON(w, http.StatusOK, adminLogsListResponse{Logs: logs, Total: total})
}
