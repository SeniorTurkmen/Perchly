package handler

import (
	"context"
	"net/http"

	"perchly-backend/internal/model"
	"perchly-backend/internal/service"
)

type adminOnboardingInsightsGetter interface {
	Insights(ctx context.Context) (model.AdminOnboardingInsights, error)
}

type AdminOnboardingHandler struct {
	insights adminOnboardingInsightsGetter
}

func NewAdminOnboardingHandler(onboarding *service.AdminOnboardingService) *AdminOnboardingHandler {
	return &AdminOnboardingHandler{insights: onboarding}
}

// --- GET /admin/onboarding/insights ---

func (h *AdminOnboardingHandler) Insights(w http.ResponseWriter, r *http.Request) {
	insights, err := h.insights.Insights(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, ErrCodeAdminOnboardingInsightsFailed, "onboarding analitiği getirilemedi")
		return
	}
	writeJSON(w, http.StatusOK, insights)
}
