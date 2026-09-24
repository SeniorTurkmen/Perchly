package model

import "time"

// Persona is a chat persona users can talk to (e.g. a motivational coach or
// a book-club partner). SystemPrompt is the raw LLM instruction backing the
// persona; it is intentionally excluded from JSON so it is never leaked to
// clients over the public API — future chat orchestration code should read
// it directly from the repository instead.
type Persona struct {
	ID               string  `json:"id"`
	Slug             string  `json:"slug"`
	Name             string  `json:"name"`
	Category         string  `json:"category"`
	ShortDescription string  `json:"short_description"`
	SystemPrompt     string  `json:"-"`
	ToneDescription  string  `json:"tone_description"`
	AvatarURL        *string `json:"avatar_url"`
	AccentColor      string  `json:"accent_color"`
	// IsMinorAppropriate gates whether a user flagged as a minor (see
	// OnboardingProfile.IsMinor) may start a conversation with this
	// persona — enforced in ConversationService.Create.
	IsMinorAppropriate bool `json:"is_minor_appropriate"`
	IsActive           bool `json:"is_active"`
	SortOrder          int  `json:"sort_order"`
	// DefaultTraits are this persona's own personality dial positions —
	// what a user gets before they've ever customized anything for it.
	// See PersonaTraits and user_persona_traits.
	DefaultTraits PersonaTraits `json:"default_traits"`
	// LLMModelID optionally pins this persona to a specific stored
	// llm_models row (see internal/service.AdminLLMService) instead of
	// the process-wide LLM_PROVIDER/LLM_MODEL default. Nil means "use
	// the default". Excluded from JSON for the same reason as
	// SystemPrompt: it's an internal routing detail, never meant for the
	// public API — the admin dashboard's own response DTO
	// (adminPersonaResponse) exposes it explicitly instead.
	LLMModelID *string   `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
