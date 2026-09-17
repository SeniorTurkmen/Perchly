package model

// PersonaRecommendation wraps a Persona with a personalized
// recommendation, computed fresh at request time from the caller's
// onboarding mood preference — see GET /personas?recommend=true. Never
// persisted. Exactly one persona in a given response has Recommended
// true; MatchReason is a short human-readable Turkish string, or nil
// when the pick was a plain fallback (no mood on file) rather than an
// actual mood match — no fabricated match percentage.
type PersonaRecommendation struct {
	Persona
	Recommended bool    `json:"recommended"`
	MatchReason *string `json:"match_reason"`
}
