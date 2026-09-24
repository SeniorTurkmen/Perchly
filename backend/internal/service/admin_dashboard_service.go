package service

import (
	"context"

	"perchly-backend/internal/model"
)

type AdminMetricsRepo interface {
	Snapshot(ctx context.Context) (model.AdminMetrics, error)
}

// AdminDashboardService backs the admin dashboard's home page. Thin by
// design — Snapshot is one aggregate query with nothing to orchestrate
// — but kept as its own service rather than calling the repository
// straight from the handler, for the same handler → service →
// repository layering every other admin endpoint follows.
type AdminDashboardService struct {
	metrics AdminMetricsRepo
}

func NewAdminDashboardService(metrics AdminMetricsRepo) *AdminDashboardService {
	return &AdminDashboardService{metrics: metrics}
}

func (s *AdminDashboardService) Metrics(ctx context.Context) (model.AdminMetrics, error) {
	return s.metrics.Snapshot(ctx)
}
