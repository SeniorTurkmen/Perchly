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

// Health godoc
// @Summary Sağlık kontrolü
// @Description Veritabanı bağlantısı dahil canlılık kontrolü. Veritabanına erişilemiyorsa 503 döner, sunucu çökmez.
// @Tags health
// @Produce json
// @Success 200 {object} model.HealthResponse
// @Failure 503 {object} model.HealthResponse
// @Router /health [get]
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
