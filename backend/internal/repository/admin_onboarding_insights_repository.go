package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

type AdminOnboardingInsightsRepository struct {
	pool *pgxpool.Pool
}

func NewAdminOnboardingInsightsRepository(pool *pgxpool.Pool) *AdminOnboardingInsightsRepository {
	return &AdminOnboardingInsightsRepository{pool: pool}
}

// Snapshot computes every onboarding-insights number in one round trip:
// the scalar counts as subqueries (same pattern as
// AdminMetricsRepository.Snapshot), and the three distributions
// (age range, mood preference, selected persona) folded into JSON via
// json_object_agg/json_agg so pgx can decode them straight into Go
// maps/slices without a query per distribution.
func (r *AdminOnboardingInsightsRepository) Snapshot(ctx context.Context) (model.AdminOnboardingInsights, error) {
	var insights model.AdminOnboardingInsights
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM users),
			(SELECT count(*) FROM onboarding_profiles),
			(SELECT count(*) FROM onboarding_profiles WHERE is_minor),
			(SELECT count(*) FROM onboarding_profiles WHERE notifications_granted),
			(SELECT count(*) FROM onboarding_profiles WHERE preferred_name IS NOT NULL),
			(SELECT count(*) FROM onboarding_profiles WHERE skip_hitap),
			(SELECT coalesce(json_object_agg(age_range, cnt), '{}'::json) FROM (
				SELECT age_range, count(*) AS cnt FROM onboarding_profiles GROUP BY age_range
			) age_counts),
			(SELECT coalesce(json_object_agg(coalesce(mood_preference, 'none'), cnt), '{}'::json) FROM (
				SELECT mood_preference, count(*) AS cnt FROM onboarding_profiles GROUP BY mood_preference
			) mood_counts),
			(SELECT coalesce(json_agg(persona_counts), '[]'::json) FROM (
				SELECT p.id::text AS persona_id, p.name AS persona_name, count(*) AS count
				FROM onboarding_profiles op
				JOIN personas p ON p.id = op.selected_persona_id
				GROUP BY p.id, p.name
				ORDER BY count(*) DESC
			) persona_counts)
	`).Scan(
		&insights.TotalUsers, &insights.CompletedOnboarding, &insights.MinorCount,
		&insights.NotificationsGrantedCount, &insights.PreferredNameSetCount, &insights.SkipHitapCount,
		&insights.AgeRangeCounts, &insights.MoodPreferenceCounts, &insights.TopSelectedPersonas,
	)
	return insights, err
}
