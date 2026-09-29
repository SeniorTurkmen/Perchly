package model

import "time"

// PersonaTranslation is one persona's display copy in a non-Turkish
// locale — name/short_description/tone_description only. Turkish is
// never stored as a row here; it's always read straight off the base
// Persona columns, mirroring apierror.Message's "Turkish is the
// zero-duplication fallback" pattern. system_prompt and category are
// deliberately absent: the former is an internal LLM instruction never
// sent to clients, the latter is a stable machine key the client maps
// to its own display copy, not free text to translate.
type PersonaTranslation struct {
	PersonaID        string    `json:"persona_id"`
	Locale           string    `json:"locale"`
	Name             string    `json:"name"`
	ShortDescription string    `json:"short_description"`
	ToneDescription  string    `json:"tone_description"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
