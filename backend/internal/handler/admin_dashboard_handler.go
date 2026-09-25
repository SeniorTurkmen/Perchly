package handler

import (
	"context"
	"net/http"

	"perchly-backend/internal/model"
	"perchly-backend/internal/service"
)

type adminMetricsGetter interface {
	Metrics(ctx context.Context) (model.AdminMetrics, error)
}

type AdminDashboardHandler struct {
	metrics adminMetricsGetter
}

func NewAdminDashboardHandler(dashboard *service.AdminDashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{metrics: dashboard}
}

// --- GET /admin/dashboard/metrics ---

func (h *AdminDashboardHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.metrics.Metrics(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminMetricsFailed, "metrikler getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}
