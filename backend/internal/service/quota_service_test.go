package service

import (
	"context"
	"testing"
	"time"

	"perchly-backend/internal/model"
)

type fakeQuotaRepoForEnergy struct {
	rows map[string]model.UserQuota // keyed by personaID
}

func (r *fakeQuotaRepoForEnergy) GetOrCreate(context.Context, string, string, int) (model.UserQuota, error) {
	panic("not used by ChatEnergy tests")
}

func (r *fakeQuotaRepoForEnergy) IncrementMessageCount(context.Context, string, string) error {
	panic("not used by ChatEnergy tests")
}

func (r *fakeQuotaRepoForEnergy) GetExisting(_ context.Context, _, personaID string) (model.UserQuota, bool, error) {
	q, ok := r.rows[personaID]
	return q, ok, nil
}

type fakePersonaListerForEnergy struct {
	personas []model.Persona
}

func (l fakePersonaListerForEnergy) List(context.Context) ([]model.Persona, error) {
	return l.personas, nil
}

type fakeUserGetterForEnergy struct {
	user model.User
}

func (g fakeUserGetterForEnergy) GetByID(context.Context, string) (model.User, error) {
	return g.user, nil
}

func TestQuotaService_ChatEnergy(t *testing.T) {
	istanbul, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	now := time.Now().In(istanbul)
	user := model.User{ID: "u1", Timezone: "Europe/Istanbul"}

	personas := []model.Persona{
		{ID: "ada", IsActive: true},
		{ID: "mira", IsActive: true},
		{ID: "kerem", IsActive: false}, // inactive: must never affect the result
	}

	t.Run("min(remaining) across active personas wins, inactive personas ignored", func(t *testing.T) {
		quotas := &fakeQuotaRepoForEnergy{rows: map[string]model.UserQuota{
			"ada":  {MessageCountToday: 2, DailyLimit: 10, LastResetAt: now},
			"mira": {MessageCountToday: 8, DailyLimit: 10, LastResetAt: now},
			// kerem has no row at all, and is inactive anyway.
		}}
		svc := NewQuotaService(quotas, nil, fakePersonaListerForEnergy{personas: personas}, fakeUserGetterForEnergy{user: user}, 10)

		energy, err := svc.ChatEnergy(context.Background(), "u1")
		if err != nil {
			t.Fatalf("ChatEnergy() error = %v", err)
		}
		if energy.Remaining != 2 || energy.Limit != 10 || energy.Used != 8 {
			t.Fatalf("energy = %+v, want the more-depleted persona (mira: 2 remaining, used 8)", energy)
		}
	})

	t.Run("a persona with no quota row yet is treated as fully unused", func(t *testing.T) {
		quotas := &fakeQuotaRepoForEnergy{rows: map[string]model.UserQuota{
			"ada": {MessageCountToday: 9, DailyLimit: 10, LastResetAt: now},
			// mira has never been messaged: no row.
		}}
		svc := NewQuotaService(quotas, nil, fakePersonaListerForEnergy{personas: personas}, fakeUserGetterForEnergy{user: user}, 10)

		energy, err := svc.ChatEnergy(context.Background(), "u1")
		if err != nil {
			t.Fatalf("ChatEnergy() error = %v", err)
		}
		// mira (no row, full 10 remaining) is NOT the min; ada (1 remaining) is.
		if energy.Remaining != 1 {
			t.Fatalf("Remaining = %d, want 1 (ada's depleted quota, not mira's untouched one)", energy.Remaining)
		}
	})

	t.Run("a stale row from a previous local day counts as unused, without mutating it", func(t *testing.T) {
		yesterday := now.AddDate(0, 0, -1)
		quotas := &fakeQuotaRepoForEnergy{rows: map[string]model.UserQuota{
			"ada":  {MessageCountToday: 10, DailyLimit: 10, LastResetAt: yesterday}, // stale: looks exhausted but rolled over
			"mira": {MessageCountToday: 1, DailyLimit: 10, LastResetAt: now},
		}}
		svc := NewQuotaService(quotas, nil, fakePersonaListerForEnergy{personas: personas}, fakeUserGetterForEnergy{user: user}, 10)

		energy, err := svc.ChatEnergy(context.Background(), "u1")
		if err != nil {
			t.Fatalf("ChatEnergy() error = %v", err)
		}
		// ada's stale row should read as 0 used / 10 remaining, so mira (9 remaining) is the min.
		if energy.Remaining != 9 || energy.Used != 1 {
			t.Fatalf("energy = %+v, want mira's 9 remaining (ada's stale row must not count as exhausted)", energy)
		}
	})

	t.Run("mood label thresholds", func(t *testing.T) {
		tests := []struct {
			remaining, limit int
			want              string
		}{
			{10, 10, "Huzurlu"},
			{7, 10, "Huzurlu"},
			{6, 10, "Dengeli"},
			{3, 10, "Dengeli"},
			{2, 10, "Yorgun"},
			{0, 10, "Yorgun"},
		}
		for _, tt := range tests {
			if got := chatEnergyMoodLabel(tt.remaining, tt.limit); got != tt.want {
				t.Errorf("chatEnergyMoodLabel(%d, %d) = %q, want %q", tt.remaining, tt.limit, got, tt.want)
			}
		}
	})

	t.Run("no active personas at all falls back to the full default, not zero", func(t *testing.T) {
		quotas := &fakeQuotaRepoForEnergy{rows: map[string]model.UserQuota{}}
		onlyInactive := []model.Persona{{ID: "kerem", IsActive: false}}
		svc := NewQuotaService(quotas, nil, fakePersonaListerForEnergy{personas: onlyInactive}, fakeUserGetterForEnergy{user: user}, 10)

		energy, err := svc.ChatEnergy(context.Background(), "u1")
		if err != nil {
			t.Fatalf("ChatEnergy() error = %v", err)
		}
		if energy.Remaining != 10 || energy.Limit != 10 {
			t.Fatalf("energy = %+v, want the full default limit when there's nothing to be low on", energy)
		}
	})
}
