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

func (r fakePersonaRepoForRecommendation) ListLocalized(context.Context, string) ([]model.Persona, error) {
	return r.personas, nil
}

func (r fakePersonaRepoForRecommendation) GetByIDLocalized(_ context.Context, id, _ string) (model.Persona, error) {
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

		recs, err := newService(profiles).ListWithRecommendation(context.Background(), "u1", "tr")
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

		recs, err := newService(profiles).ListWithRecommendation(context.Background(), "u2", "tr")
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

		recs, err := newService(profiles).ListWithRecommendation(context.Background(), "u3", "tr")
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

		recs, err := newService(profiles).ListWithRecommendation(context.Background(), "unknown-user", "tr")
		if err != nil {
			t.Fatalf("ListWithRecommendation() error = %v", err)
		}
		if got := countRecommended(t, recs); got != 1 {
			t.Fatalf("recommended count = %d, want exactly 1", got)
		}
	})
}

// fakeLocaleRecordingPersonaRepo records the locale it was called with,
// so tests can confirm PersonaService threads its locale argument through
// to the repository rather than dropping it.
type fakeLocaleRecordingPersonaRepo struct {
	lastListLocale   string
	lastGetIDLocale  string
	persona          model.Persona
	translatedResult []model.Persona
}

func (r *fakeLocaleRecordingPersonaRepo) ListLocalized(_ context.Context, locale string) ([]model.Persona, error) {
	r.lastListLocale = locale
	return r.translatedResult, nil
}

func (r *fakeLocaleRecordingPersonaRepo) GetByIDLocalized(_ context.Context, _ string, locale string) (model.Persona, error) {
	r.lastGetIDLocale = locale
	return r.persona, nil
}

func TestPersonaService_List_ThreadsLocaleToRepo(t *testing.T) {
	repo := &fakeLocaleRecordingPersonaRepo{translatedResult: []model.Persona{{ID: "ada", Name: "Ada"}}}
	svc := NewPersonaService(repo, newFakeOnboardingProfileRepo())

	if _, err := svc.List(context.Background(), "en"); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repo.lastListLocale != "en" {
		t.Fatalf("repo received locale %q, want %q", repo.lastListLocale, "en")
	}
}

func TestPersonaService_GetByID_ThreadsLocaleToRepo(t *testing.T) {
	repo := &fakeLocaleRecordingPersonaRepo{persona: model.Persona{ID: "ada", Name: "Ada"}}
	svc := NewPersonaService(repo, newFakeOnboardingProfileRepo())

	if _, err := svc.GetByID(context.Background(), "ada", "de"); err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if repo.lastGetIDLocale != "de" {
		t.Fatalf("repo received locale %q, want %q", repo.lastGetIDLocale, "de")
	}
}
