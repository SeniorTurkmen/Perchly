package service

import (
	"context"

	"perchly-backend/internal/model"
)

type AdminOnboardingRepo interface {
	Snapshot(ctx context.Context) (model.AdminOnboardingInsights, error)
}

// AdminOnboardingService backs the admin dashboard's onboarding
// insights page — read-only, same reasoning as AdminDashboardService.
type AdminOnboardingService struct {
	insights AdminOnboardingRepo
}

func NewAdminOnboardingService(insights AdminOnboardingRepo) *AdminOnboardingService {
	return &AdminOnboardingService{insights: insights}
}

func (s *AdminOnboardingService) Insights(ctx context.Context) (model.AdminOnboardingInsights, error) {
	return s.insights.Snapshot(ctx)
}
