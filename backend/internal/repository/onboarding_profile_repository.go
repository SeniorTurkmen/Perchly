package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrOnboardingProfileNotFound is returned when a user hasn't completed
// (or ever submitted) onboarding.
var ErrOnboardingProfileNotFound = errors.New("onboarding profile not found")

type OnboardingProfileRepository struct {
	pool *pgxpool.Pool
}

func NewOnboardingProfileRepository(pool *pgxpool.Pool) *OnboardingProfileRepository {
	return &OnboardingProfileRepository{pool: pool}
}

const onboardingProfileColumns = `
	user_id::text, age_range, is_minor, mood_preference,
	notifications_granted, selected_persona_id::text, preferred_name, skip_hitap,
	created_at, updated_at`

// Upsert saves a user's onboarding profile — there is only ever one per
// user, so a resubmission (e.g. redoing onboarding) replaces it rather
// than erroring. Callers decide PreferredName/SkipHitap up front (see
// OnboardingService.SaveProfile) — this just writes whatever the
// profile carries, verbatim.
func (r *OnboardingProfileRepository) Upsert(ctx context.Context, profile model.OnboardingProfile) (model.OnboardingProfile, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO onboarding_profiles (user_id, age_range, is_minor, mood_preference, notifications_granted, selected_persona_id, preferred_name, skip_hitap)
		VALUES ($1::uuid, $2, $3, $4, $5, $6::uuid, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			age_range = EXCLUDED.age_range,
			is_minor = EXCLUDED.is_minor,
			mood_preference = EXCLUDED.mood_preference,
			notifications_granted = EXCLUDED.notifications_granted,
			selected_persona_id = EXCLUDED.selected_persona_id,
			preferred_name = EXCLUDED.preferred_name,
			skip_hitap = EXCLUDED.skip_hitap,
			updated_at = now()
		RETURNING `+onboardingProfileColumns,
		profile.UserID, profile.AgeRange, profile.IsMinor, profile.MoodPreference,
		profile.NotificationsGranted, profile.SelectedPersonaID, profile.PreferredName, profile.SkipHitap,
	)
	return scanOnboardingProfile(row)
}

func (r *OnboardingProfileRepository) GetByUserID(ctx context.Context, userID string) (model.OnboardingProfile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+onboardingProfileColumns+`
		FROM onboarding_profiles
		WHERE user_id = $1::uuid
	`, userID)

	p, err := scanOnboardingProfile(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.OnboardingProfile{}, ErrOnboardingProfileNotFound
		}
		return model.OnboardingProfile{}, err
	}
	return p, nil
}

// Exists reports whether userID has a saved onboarding profile, without
// fetching its columns — a lighter query than GetByUserID for callers
// (see AuthService) that only need the yes/no.
func (r *OnboardingProfileRepository) Exists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM onboarding_profiles WHERE user_id = $1::uuid)
	`, userID).Scan(&exists)
	return exists, err
}

func scanOnboardingProfile(row rowScanner) (model.OnboardingProfile, error) {
	var p model.OnboardingProfile
	err := row.Scan(
		&p.UserID, &p.AgeRange, &p.IsMinor, &p.MoodPreference,
		&p.NotificationsGranted, &p.SelectedPersonaID, &p.PreferredName, &p.SkipHitap,
		&p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}
