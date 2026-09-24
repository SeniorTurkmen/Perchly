package model

// AdminOnboardingInsights is a snapshot of the onboarding funnel and
// what users answered during it, for the admin dashboard. Counts, not
// percentages — the frontend computes rates against TotalUsers /
// CompletedOnboarding so it can choose what to divide by per stat.
type AdminOnboardingInsights struct {
	TotalUsers                int `json:"total_users"`
	CompletedOnboarding       int `json:"completed_onboarding"`
	MinorCount                int `json:"minor_count"`
	NotificationsGrantedCount int `json:"notifications_granted_count"`
	PreferredNameSetCount     int `json:"preferred_name_set_count"`
	SkipHitapCount            int `json:"skip_hitap_count"`
	// AgeRangeCounts/MoodPreferenceCounts are keyed by the raw wire
	// value (see OnboardingProfile's doc comment for the exact enum
	// cases) — a missing mood_preference is grouped under the key
	// "none", not omitted.
	AgeRangeCounts       map[string]int          `json:"age_range_counts"`
	MoodPreferenceCounts map[string]int          `json:"mood_preference_counts"`
	TopSelectedPersonas  []PersonaSelectionCount `json:"top_selected_personas"`
}

// PersonaSelectionCount is how many onboarding profiles picked a given
// persona, newest-first isn't relevant here — this is always returned
// ordered by Count descending.
type PersonaSelectionCount struct {
	PersonaID   string `json:"persona_id"`
	PersonaName string `json:"persona_name"`
	Count       int    `json:"count"`
}
