package model

import "errors"

// ErrPersonaTraitOutOfRange is returned when a requested trait value
// isn't a valid dial position.
var ErrPersonaTraitOutOfRange = errors.New("persona trait values must be between 0 and 100")

// PersonaTraits are the five personality dials a user can tune for a
// given persona (Perchmate), each 0-100. They're read fresh at the
// start of every chat turn (see ChatService.SendMessage /
// ContextBuilder.Build) — never baked into the persona's own system
// prompt — so a change takes effect on the very next message, and one
// user's dial positions never bleed into another user's conversation
// with the same persona.
type PersonaTraits struct {
	// Warmth: how affectionate/close vs. distant and formal the persona is.
	Warmth int `json:"warmth"`
	// Humor: how often and readily it jokes vs. stays fully serious.
	Humor int `json:"humor"`
	// Wisdom: how deep/reflective vs. surface-level its responses are.
	Wisdom int `json:"wisdom"`
	// Directness: how blunt/unsoftened vs. gentle and hedged its
	// responses are — this is the dial behind "harsher answers".
	Directness int `json:"directness"`
	// Energy: how energetic/enthusiastic vs. calm and measured its tone is.
	Energy int `json:"energy"`
}

// Validate reports whether every dial is within [0, 100].
func (t PersonaTraits) Validate() error {
	for _, v := range [...]int{t.Warmth, t.Humor, t.Wisdom, t.Directness, t.Energy} {
		if v < 0 || v > 100 {
			return ErrPersonaTraitOutOfRange
		}
	}
	return nil
}
