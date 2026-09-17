package service

import (
	"context"
	"errors"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

// PersonaRepo is the persistence dependency PersonaService needs.
type PersonaRepo interface {
	List(ctx context.Context) ([]model.Persona, error)
	GetByID(ctx context.Context, id string) (model.Persona, error)
}

// moodToCategory mirrors the iOS client's own mood -> persona category
// mapping (see OnboardingProfile.MoodPreference.personaCategory) — kept
// server-side too so GET /personas?recommend=true agrees with it
// without the two ever being able to drift apart silently.
var moodToCategory = map[string]string{
	"motivation": "motivational_coach",
	"dailyChat":  "daily_companion",
	"hobbyTalk":  "hobby_book_club",
}

// moodMatchReason returns the short Turkish explanation shown for a
// mood-driven recommendation, or nil for a mood with no such copy
// (skipped/unknown) — those fall back to a plain positional pick with
// no reason attached, rather than an invented one.
func moodMatchReason(mood string) *string {
	var reason string
	switch mood {
	case "motivation":
		reason = "Motive olmak istediğin için önerdik."
	case "dailyChat":
		reason = "Gündelik sohbet aradığın için önerdik."
	case "hobbyTalk":
		reason = "Hobi paylaşmak istediğin için önerdik."
	default:
		return nil
	}
	return &reason
}

type PersonaService struct {
	repo               PersonaRepo
	onboardingProfiles OnboardingProfileRepo
}

func NewPersonaService(repo PersonaRepo, onboardingProfiles OnboardingProfileRepo) *PersonaService {
	return &PersonaService{repo: repo, onboardingProfiles: onboardingProfiles}
}

func (s *PersonaService) List(ctx context.Context) ([]model.Persona, error) {
	return s.repo.List(ctx)
}

func (s *PersonaService) GetByID(ctx context.Context, id string) (model.Persona, error) {
	return s.repo.GetByID(ctx, id)
}

// ListWithRecommendation is GET /personas?recommend=true's data: every
// active persona, with exactly one flagged Recommended — the one
// matching userID's onboarding mood preference (see moodToCategory),
// or the first persona by sort order if there's no mood on file or no
// active persona in that category. A minor never gets a
// not-appropriate-for-them persona recommended, mirroring the same
// gate ConversationService.Create enforces when actually starting a
// chat — this is presentation only, so the full list is still
// returned either way.
func (s *PersonaService) ListWithRecommendation(ctx context.Context, userID string) ([]model.PersonaRecommendation, error) {
	personas, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := s.onboardingProfiles.GetByUserID(ctx, userID)
	isMinor := false
	var moodCategory string
	switch {
	case err == nil:
		isMinor = profile.IsMinor
		if profile.MoodPreference != nil {
			moodCategory = moodToCategory[*profile.MoodPreference]
		}
	case errors.Is(err, repository.ErrOnboardingProfileNotFound):
		// No profile yet — permissive, same as ConversationService's
		// own fail-open default for an unknown age/mood.
	default:
		return nil, err
	}

	appropriate := func(p model.Persona) bool {
		return !isMinor || p.IsMinorAppropriate
	}

	recommendedIndex := -1
	if moodCategory != "" {
		for i, p := range personas {
			if p.Category == moodCategory && appropriate(p) {
				recommendedIndex = i
				break
			}
		}
	}
	if recommendedIndex == -1 {
		for i, p := range personas {
			if appropriate(p) {
				recommendedIndex = i
				break
			}
		}
	}

	result := make([]model.PersonaRecommendation, len(personas))
	for i, p := range personas {
		rec := model.PersonaRecommendation{Persona: p, Recommended: i == recommendedIndex}
		if i == recommendedIndex && moodCategory != "" && p.Category == moodCategory {
			rec.MatchReason = moodMatchReason(*profile.MoodPreference)
		}
		result[i] = rec
	}
	return result, nil
}
