package service

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

var (
	ErrInvalidAgeRange       = errors.New("invalid age range")
	ErrInvalidMoodPreference = errors.New("invalid mood preference")
	// ErrPreferredNameRequired is returned when the caller explicitly
	// opted into being addressed by name (skip_hitap: false) but didn't
	// actually provide one.
	ErrPreferredNameRequired = errors.New("preferred_name is required unless skip_hitap is true")
	// ErrPreferredNameInvalid is returned for a preferred_name outside
	// 1-40 unicode characters, or containing a control character.
	ErrPreferredNameInvalid = errors.New("preferred_name must be 1-40 characters with no control characters")
)

var validAgeRanges = map[string]bool{
	"under18":   true,
	"age18to24": true,
	"age25to34": true,
	"age35plus": true,
}

var validMoodPreferences = map[string]bool{
	"motivation": true,
	"dailyChat":  true,
	"hobbyTalk":  true,
	"skipped":    true,
}

type OnboardingProfileRepo interface {
	Upsert(ctx context.Context, profile model.OnboardingProfile) (model.OnboardingProfile, error)
	GetByUserID(ctx context.Context, userID string) (model.OnboardingProfile, error)
	// Exists reports whether userID has a saved onboarding profile at
	// all, without fetching its columns — used by AuthService to fill
	// in HasCompletedOnboarding on every token issuance.
	Exists(ctx context.Context, userID string) (bool, error)
}

// OnboardingService persists what a user answered during onboarding.
type OnboardingService struct {
	profiles OnboardingProfileRepo
}

func NewOnboardingService(profiles OnboardingProfileRepo) *OnboardingService {
	return &OnboardingService{profiles: profiles}
}

// SaveProfile validates and stores a user's onboarding answers.
//
// profile.IsMinor is deliberately ignored and recomputed from
// AgeRange here rather than trusted from the client: it's a
// safety-relevant field (it gates which personas ConversationService
// will allow — see IsMinorAppropriate), so the server must be the one
// deciding it, not whatever a client happened to send.
//
// preferredNameInput and skipHitapInput are the request's raw
// preferred_name/skip_hitap values — nil means the field was absent
// (or JSON null), which a *string/*bool can't otherwise distinguish
// from "explicitly empty/false". That distinction is exactly what
// decides how this is handled:
//
//   - skip_hitap: true            -> preferred_name forced to nil, stored as opted out.
//   - preferred_name given        -> trimmed, validated, stored (skip_hitap false).
//   - skip_hitap: false, no name  -> ErrPreferredNameRequired (a client bug: it opted
//     in to a name but didn't send one).
//   - neither field sent at all   -> a pre-hitap client; whatever was already on file
//     (nil/false for a brand new profile) is carried over untouched, never
//     silently cleared by a resubmission that doesn't know about this feature.
//
// The server never invents a nickname under any circumstance.
func (s *OnboardingService) SaveProfile(
	ctx context.Context,
	profile model.OnboardingProfile,
	preferredNameInput *string,
	skipHitapInput *bool,
) (model.OnboardingProfile, error) {
	if !validAgeRanges[profile.AgeRange] {
		return model.OnboardingProfile{}, ErrInvalidAgeRange
	}
	if profile.MoodPreference != nil && !validMoodPreferences[*profile.MoodPreference] {
		return model.OnboardingProfile{}, ErrInvalidMoodPreference
	}

	profile.IsMinor = profile.AgeRange == "under18"

	switch {
	case skipHitapInput != nil && *skipHitapInput:
		profile.PreferredName = nil
		profile.SkipHitap = true

	case preferredNameInput != nil:
		trimmed := strings.TrimSpace(*preferredNameInput)
		if trimmed == "" {
			return model.OnboardingProfile{}, ErrPreferredNameRequired
		}
		if err := validatePreferredName(trimmed); err != nil {
			return model.OnboardingProfile{}, err
		}
		profile.PreferredName = &trimmed
		profile.SkipHitap = false

	case skipHitapInput != nil && !*skipHitapInput:
		// Explicitly opted in to a hitap but didn't provide one.
		return model.OnboardingProfile{}, ErrPreferredNameRequired

	default:
		// Neither field was sent — a client that doesn't know about
		// this feature yet. Preserve whatever's already on file rather
		// than clobbering it to nil/false on every resubmission.
		existing, err := s.profiles.GetByUserID(ctx, profile.UserID)
		switch {
		case err == nil:
			profile.PreferredName = existing.PreferredName
			profile.SkipHitap = existing.SkipHitap
		case errors.Is(err, repository.ErrOnboardingProfileNotFound):
			profile.PreferredName = nil
			profile.SkipHitap = false
		default:
			return model.OnboardingProfile{}, err
		}
	}

	return s.profiles.Upsert(ctx, profile)
}

// validatePreferredName enforces 1-40 unicode characters with no
// control characters, on an already-trimmed name.
func validatePreferredName(name string) error {
	count := utf8.RuneCountInString(name)
	if count < 1 || count > 40 {
		return ErrPreferredNameInvalid
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return ErrPreferredNameInvalid
		}
	}
	return nil
}

func (s *OnboardingService) GetByUserID(ctx context.Context, userID string) (model.OnboardingProfile, error) {
	return s.profiles.GetByUserID(ctx, userID)
}
