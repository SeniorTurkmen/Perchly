package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"perchly-backend/internal/model"
)

// healthChecker is the subset of HealthService the handler depends on.
type healthChecker interface {
	CheckDatabase(ctx context.Context) error
}

type HealthHandler struct {
	health healthChecker
}

func NewHealthHandler(health healthChecker) *HealthHandler {
	return &HealthHandler{health: health}
}

// Health handles GET /health. It reports 200 when the database is
// reachable and 503 (with status "degraded") when it is not, so the
// server can start and be probed even while the database is unavailable.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := model.HealthResponse{Status: "ok", Database: "ok"}
	status := http.StatusOK

	if err := h.health.CheckDatabase(r.Context()); err != nil {
		resp.Status = "degraded"
		resp.Database = "unreachable"
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
