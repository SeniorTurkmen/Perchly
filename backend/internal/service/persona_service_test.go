package service

import (
	"context"
	"testing"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type fakePersonaRepoForRecommendation struct {
	personas []model.Persona
}

func (r fakePersonaRepoForRecommendation) List(context.Context) ([]model.Persona, error) {
	return r.personas, nil
}

func (r fakePersonaRepoForRecommendation) GetByID(_ context.Context, id string) (model.Persona, error) {
	for _, p := range r.personas {
		if p.ID == id {
			return p, nil
		}
	}
	return model.Persona{}, repository.ErrPersonaNotFound
}

func TestPersonaService_ListWithRecommendation(t *testing.T) {
	ada := model.Persona{ID: "ada", Category: "motivational_coach", SortOrder: 1, IsMinorAppropriate: true}
	mira := model.Persona{ID: "mira", Category: "daily_companion", SortOrder: 2, IsMinorAppropriate: true}
	kerem := model.Persona{ID: "kerem", Category: "hobby_book_club", SortOrder: 3, IsMinorAppropriate: false}
	personas := []model.Persona{ada, mira, kerem}

	newService := func(profiles *fakeOnboardingProfileRepo) *PersonaService {
		return NewPersonaService(fakePersonaRepoForRecommendation{personas: personas}, profiles)
	}

	countRecommended := func(t *testing.T, recs []model.PersonaRecommendation) int {
		t.Helper()
		n := 0
		for _, r := range recs {
			if r.Recommended {
				n++
			}
		}
		return n
	}

	t.Run("mood preference recommends the matching category", func(t *testing.T) {
		profiles := newFakeOnboardingProfileRepo()
		profiles.profiles["u1"] = model.OnboardingProfile{UserID: "u1", MoodPreference: ptr("hobbyTalk")}

		recs, err := newService(profiles).ListWithRecommendation(context.Background(), "u1")
		if err != nil {
			t.Fatalf("ListWithRecommendation() error = %v", err)
		}
		if got := countRecommended(t, recs); got != 1 {
			t.Fatalf("recommended count = %d, want exactly 1", got)
		}
		for _, r := range recs {
			if r.Recommended && r.ID != "kerem" {
				t.Fatalf("recommended persona = %s, want kerem for hobbyTalk", r.ID)
			}
		}
	})

	t.Run("minor never gets a not-minor-appropriate persona recommended", func(t *testing.T) {
		profiles := newFakeOnboardingProfileRepo()
		profiles.profiles["u2"] = model.OnboardingProfile{
			UserID: "u2", IsMinor: true, MoodPreference: ptr("hobbyTalk"), // kerem matches the mood but isn't minor-appropriate
		}

		recs, err := newService(profiles).ListWithRecommendation(context.Background(), "u2")
		if err != nil {
			t.Fatalf("ListWithRecommendation() error = %v", err)
		}
		if got := countRecommended(t, recs); got != 1 {
			t.Fatalf("recommended count = %d, want exactly 1", got)
		}
		for _, r := range recs {
			if r.Recommended && !r.IsMinorAppropriate {
				t.Fatalf("recommended persona %s is not minor-appropriate", r.ID)
			}
		}
	})

	t.Run("skipped mood falls back to the first active persona by sort order, no fabricated reason", func(t *testing.T) {
		profiles := newFakeOnboardingProfileRepo()
		profiles.profiles["u3"] = model.OnboardingProfile{UserID: "u3", MoodPreference: ptr("skipped")}

		recs, err := newService(profiles).ListWithRecommendation(context.Background(), "u3")
		if err != nil {
			t.Fatalf("ListWithRecommendation() error = %v", err)
		}
		if got := countRecommended(t, recs); got != 1 {
			t.Fatalf("recommended count = %d, want exactly 1", got)
		}
		if !recs[0].Recommended {
			t.Fatalf("expected the first persona (by sort order) to be recommended, got: %+v", recs)
		}
		if recs[0].MatchReason != nil {
			t.Fatalf("MatchReason = %v, want nil for a plain positional fallback", recs[0].MatchReason)
		}
	})

	t.Run("no onboarding profile at all is permissive, still recommends exactly one", func(t *testing.T) {
		profiles := newFakeOnboardingProfileRepo()

		recs, err := newService(profiles).ListWithRecommendation(context.Background(), "unknown-user")
		if err != nil {
			t.Fatalf("ListWithRecommendation() error = %v", err)
		}
		if got := countRecommended(t, recs); got != 1 {
			t.Fatalf("recommended count = %d, want exactly 1", got)
		}
	})
}
