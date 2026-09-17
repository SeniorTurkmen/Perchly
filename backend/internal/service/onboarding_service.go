package service

import (
	"context"
	"errors"

	"perchly-backend/internal/model"
)

var (
	ErrInvalidAgeRange       = errors.New("invalid age range")
	ErrInvalidMoodPreference = errors.New("invalid mood preference")
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
func (s *OnboardingService) SaveProfile(ctx context.Context, profile model.OnboardingProfile) (model.OnboardingProfile, error) {
	if !validAgeRanges[profile.AgeRange] {
		return model.OnboardingProfile{}, ErrInvalidAgeRange
	}
	if profile.MoodPreference != nil && !validMoodPreferences[*profile.MoodPreference] {
		return model.OnboardingProfile{}, ErrInvalidMoodPreference
	}

	profile.IsMinor = profile.AgeRange == "under18"

	return s.profiles.Upsert(ctx, profile)
}

func (s *OnboardingService) GetByUserID(ctx context.Context, userID string) (model.OnboardingProfile, error) {
	return s.profiles.GetByUserID(ctx, userID)
}
