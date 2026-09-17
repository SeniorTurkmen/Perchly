package service

import (
	"context"
	"errors"
	"testing"

	"perchly-backend/internal/model"
	"perchly-backend/internal/repository"
)

type fakeOnboardingProfileRepo struct {
	profiles map[string]model.OnboardingProfile
}

func newFakeOnboardingProfileRepo() *fakeOnboardingProfileRepo {
	return &fakeOnboardingProfileRepo{profiles: map[string]model.OnboardingProfile{}}
}

func (r *fakeOnboardingProfileRepo) Upsert(_ context.Context, p model.OnboardingProfile) (model.OnboardingProfile, error) {
	r.profiles[p.UserID] = p
	return p, nil
}

func (r *fakeOnboardingProfileRepo) GetByUserID(_ context.Context, userID string) (model.OnboardingProfile, error) {
	p, ok := r.profiles[userID]
	if !ok {
		return model.OnboardingProfile{}, repository.ErrOnboardingProfileNotFound
	}
	return p, nil
}

func (r *fakeOnboardingProfileRepo) Exists(_ context.Context, userID string) (bool, error) {
	_, ok := r.profiles[userID]
	return ok, nil
}

func ptr[T any](v T) *T { return &v }

func TestOnboardingService_SaveProfile_PreferredNameAndSkipHitap(t *testing.T) {
	baseProfile := func(userID string) model.OnboardingProfile {
		return model.OnboardingProfile{UserID: userID, AgeRange: "age25to34"}
	}

	t.Run("preferred name persists and round-trips", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		saved, err := svc.SaveProfile(context.Background(), baseProfile("u1"), ptr("  Deniz  "), ptr(false))
		if err != nil {
			t.Fatalf("SaveProfile() error = %v", err)
		}
		if saved.PreferredName == nil || *saved.PreferredName != "Deniz" {
			t.Fatalf("PreferredName = %v, want trimmed 'Deniz'", saved.PreferredName)
		}
		if saved.SkipHitap {
			t.Fatalf("SkipHitap = true, want false")
		}

		fetched, err := svc.GetByUserID(context.Background(), "u1")
		if err != nil {
			t.Fatalf("GetByUserID() error = %v", err)
		}
		if fetched.PreferredName == nil || *fetched.PreferredName != "Deniz" {
			t.Fatalf("round-trip PreferredName = %v, want 'Deniz'", fetched.PreferredName)
		}
	})

	t.Run("skip_hitap true forces preferred_name to nil even if a name was also sent", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		saved, err := svc.SaveProfile(context.Background(), baseProfile("u2"), ptr("Deniz"), ptr(true))
		if err != nil {
			t.Fatalf("SaveProfile() error = %v", err)
		}
		if saved.PreferredName != nil {
			t.Fatalf("PreferredName = %v, want nil when skip_hitap is true", saved.PreferredName)
		}
		if !saved.SkipHitap {
			t.Fatalf("SkipHitap = false, want true")
		}
	})

	t.Run("skip_hitap false with no name is rejected", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		_, err := svc.SaveProfile(context.Background(), baseProfile("u3"), nil, ptr(false))
		if !errors.Is(err, ErrPreferredNameRequired) {
			t.Fatalf("SaveProfile() error = %v, want ErrPreferredNameRequired", err)
		}
	})

	t.Run("empty preferred_name is rejected", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		_, err := svc.SaveProfile(context.Background(), baseProfile("u4"), ptr("   "), nil)
		if !errors.Is(err, ErrPreferredNameRequired) {
			t.Fatalf("SaveProfile() error = %v, want ErrPreferredNameRequired", err)
		}
	})

	t.Run("preferred_name over 40 unicode characters is rejected", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		long := ""
		for i := 0; i < 41; i++ {
			long += "a"
		}
		_, err := svc.SaveProfile(context.Background(), baseProfile("u5"), &long, nil)
		if !errors.Is(err, ErrPreferredNameInvalid) {
			t.Fatalf("SaveProfile() error = %v, want ErrPreferredNameInvalid", err)
		}
	})

	t.Run("preferred_name with a control character is rejected", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		bad := "Deniz"
		_, err := svc.SaveProfile(context.Background(), baseProfile("u6"), &bad, nil)
		if !errors.Is(err, ErrPreferredNameInvalid) {
			t.Fatalf("SaveProfile() error = %v, want ErrPreferredNameInvalid", err)
		}
	})

	t.Run("legacy client sending neither field leaves a brand-new profile at nil/false", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		saved, err := svc.SaveProfile(context.Background(), baseProfile("u7"), nil, nil)
		if err != nil {
			t.Fatalf("SaveProfile() error = %v", err)
		}
		if saved.PreferredName != nil || saved.SkipHitap {
			t.Fatalf("got PreferredName=%v SkipHitap=%v, want nil/false for a legacy submission", saved.PreferredName, saved.SkipHitap)
		}
	})

	t.Run("legacy client resubmission does not clobber a previously saved hitap preference", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		if _, err := svc.SaveProfile(context.Background(), baseProfile("u8"), ptr("Deniz"), ptr(false)); err != nil {
			t.Fatalf("initial SaveProfile() error = %v", err)
		}

		// A legacy resubmission (e.g. re-running onboarding on an old
		// client) sends only the original 4 fields.
		saved, err := svc.SaveProfile(context.Background(), baseProfile("u8"), nil, nil)
		if err != nil {
			t.Fatalf("SaveProfile() error = %v", err)
		}
		if saved.PreferredName == nil || *saved.PreferredName != "Deniz" {
			t.Fatalf("PreferredName = %v, want the previously saved 'Deniz' to survive a legacy resubmission", saved.PreferredName)
		}
	})

	t.Run("is_minor is still recomputed from age_range regardless of hitap fields", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		profile := model.OnboardingProfile{UserID: "u9", AgeRange: "under18", IsMinor: false} // client trying to lie
		saved, err := svc.SaveProfile(context.Background(), profile, ptr("Deniz"), ptr(false))
		if err != nil {
			t.Fatalf("SaveProfile() error = %v", err)
		}
		if !saved.IsMinor {
			t.Fatalf("IsMinor = false, want true for under18 regardless of anything else in the request")
		}
	})

	t.Run("invalid age range is still rejected before any hitap logic runs", func(t *testing.T) {
		repo := newFakeOnboardingProfileRepo()
		svc := NewOnboardingService(repo)

		_, err := svc.SaveProfile(context.Background(), model.OnboardingProfile{UserID: "u10", AgeRange: "not-a-real-range"}, ptr("Deniz"), ptr(false))
		if !errors.Is(err, ErrInvalidAgeRange) {
			t.Fatalf("SaveProfile() error = %v, want ErrInvalidAgeRange", err)
		}
	})
}
