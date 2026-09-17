package service

import (
	"context"

	"perchly-backend/internal/model"
)

// PersonaTraitsRepo is the persistence dependency for PersonaTraitsService.
type PersonaTraitsRepo interface {
	GetEffective(ctx context.Context, userID, personaID string) (model.PersonaTraits, bool, error)
	Upsert(ctx context.Context, userID, personaID string, traits model.PersonaTraits) error
	Reset(ctx context.Context, userID, personaID string) error
}

// PersonaTraitsService lets a user tune (and later reset) their own
// personality-dial positions for a persona — see model.PersonaTraits.
type PersonaTraitsService struct {
	traits   PersonaTraitsRepo
	personas PersonaRepo
}

func NewPersonaTraitsService(traits PersonaTraitsRepo, personas PersonaRepo) *PersonaTraitsService {
	return &PersonaTraitsService{traits: traits, personas: personas}
}

// Get returns userID's current effective dial values for personaID
// (their own customization if any, else the persona's defaults) and
// whether they've actually customized it.
func (s *PersonaTraitsService) Get(ctx context.Context, userID, personaID string) (model.PersonaTraits, bool, error) {
	if _, err := s.personas.GetByID(ctx, personaID); err != nil {
		return model.PersonaTraits{}, false, err
	}
	return s.traits.GetEffective(ctx, userID, personaID)
}

// Set validates and saves userID's new dial values for personaID —
// effective starting with their very next message to it.
func (s *PersonaTraitsService) Set(ctx context.Context, userID, personaID string, traits model.PersonaTraits) error {
	if _, err := s.personas.GetByID(ctx, personaID); err != nil {
		return err
	}
	if err := traits.Validate(); err != nil {
		return err
	}
	return s.traits.Upsert(ctx, userID, personaID, traits)
}

// Reset clears userID's customization for personaID — the persona's
// own defaults apply again from their next message on.
func (s *PersonaTraitsService) Reset(ctx context.Context, userID, personaID string) error {
	if _, err := s.personas.GetByID(ctx, personaID); err != nil {
		return err
	}
	return s.traits.Reset(ctx, userID, personaID)
}
