package model

import "time"

// OnboardingProfile is what a user answered during onboarding, tied to
// their user id — collected even for a purely anonymous session (see
// the auth system's device-id-pinned anonymous users), so nothing is
// lost if they later link an email: an anonymous user upgrades in place
// (same id), and this row upgrades with it.
type OnboardingProfile struct {
	UserID string `json:"user_id"`
	// One of "under18", "age18to24", "age25to34", "age35plus" — matches
	// the iOS OnboardingProfile.AgeRange enum's case names exactly.
	AgeRange string `json:"age_range"`
	// Always recomputed server-side from AgeRange (see
	// OnboardingService.SaveProfile) — never trusted verbatim from the
	// client, since persona age-gating depends on it.
	IsMinor bool `json:"is_minor"`
	// One of "motivation", "dailyChat", "hobbyTalk", "skipped", or nil.
	MoodPreference       *string   `json:"mood_preference"`
	NotificationsGranted bool      `json:"notifications_granted"`
	SelectedPersonaID    *string   `json:"selected_persona_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
