package service

import (
	"context"
	"time"

	"perchly-backend/internal/model"
)

type QuotaRepo interface {
	GetOrCreate(ctx context.Context, userID, personaID string, defaultDailyLimit int) (model.UserQuota, error)
	IncrementMessageCount(ctx context.Context, userID, personaID string) error
	GetExisting(ctx context.Context, userID, personaID string) (model.UserQuota, bool, error)
}

type CreditRepo interface {
	GetBalance(ctx context.Context, userID string) (int, error)
	SpendOne(ctx context.Context, userID string) (bool, error)
}

// ChatEnergyPersonaLister is the subset of PersonaRepository ChatEnergy
// needs — every active persona counts toward the "Günün sohbet
// enerjisi" figure, not just ones the user has already messaged.
type ChatEnergyPersonaLister interface {
	List(ctx context.Context) ([]model.Persona, error)
}

// ChatEnergyUserGetter is the subset of UserRepository ChatEnergy
// needs — only the user's IANA timezone, to judge whether a quota
// row's last_reset_at is from today or a previous local day.
type ChatEnergyUserGetter interface {
	GetByID(ctx context.Context, id string) (model.User, error)
}

type QuotaCheckResult struct {
	Allowed bool
	// UseCredit is true when the daily quota was already exhausted and
	// this message will spend one credit instead. RecordUsage needs this
	// to know which counter to touch after the message succeeds.
	UseCredit bool
}

// QuotaService enforces a per-user, per-persona daily message limit, with
// purchased credits as overage once that limit is hit:
//  1. under the daily limit -> allowed, free
//  2. at the daily limit but credit_amount > 0 -> allowed, spends a credit
//  3. neither -> rejected
//
// DailyLimit lives on each user_quotas row rather than a hardcoded
// constant, so a future admin dashboard can grant an individual
// user/persona pair a different limit just by updating that column.
type QuotaService struct {
	quotas            QuotaRepo
	credits           CreditRepo
	personas          ChatEnergyPersonaLister
	users             ChatEnergyUserGetter
	defaultDailyLimit int
}

func NewQuotaService(quotas QuotaRepo, credits CreditRepo, personas ChatEnergyPersonaLister, users ChatEnergyUserGetter, defaultDailyLimit int) *QuotaService {
	return &QuotaService{quotas: quotas, credits: credits, personas: personas, users: users, defaultDailyLimit: defaultDailyLimit}
}

// Check reports whether userID may send another message to personaID
// right now. It never mutates any counters — see RecordUsage, which is
// only called after a message actually succeeds, so a failed LLM call
// never costs the user part of their quota.
//
// Known tradeoff: Check and RecordUsage are two separate steps (checked
// before sending, recorded after success), not one atomic
// reserve-then-confirm. Two concurrent requests from the same user on
// their very last message could both pass Check before either calls
// RecordUsage, overrunning the limit by one. Given how unlikely
// concurrent double-sends in the same conversation are, and that the
// alternative (reserve now, refund on failure) is meaningfully more
// complex, this is an accepted tradeoff.
func (s *QuotaService) Check(ctx context.Context, userID, personaID string) (QuotaCheckResult, error) {
	quota, err := s.quotas.GetOrCreate(ctx, userID, personaID, s.defaultDailyLimit)
	if err != nil {
		return QuotaCheckResult{}, err
	}

	if quota.MessageCountToday < quota.DailyLimit {
		return QuotaCheckResult{Allowed: true}, nil
	}

	balance, err := s.credits.GetBalance(ctx, userID)
	if err != nil {
		return QuotaCheckResult{}, err
	}
	if balance > 0 {
		return QuotaCheckResult{Allowed: true, UseCredit: true}, nil
	}

	return QuotaCheckResult{Allowed: false}, nil
}

// RecordUsage is called once a message has successfully been answered:
// it increments today's free count, or spends one credit if useCredit
// (as decided by the Check that gated this request).
func (s *QuotaService) RecordUsage(ctx context.Context, userID, personaID string, useCredit bool) error {
	if useCredit {
		_, err := s.credits.SpendOne(ctx, userID)
		// If nothing was left to spend (e.g. raced with another request),
		// the message already succeeded — there's nothing left to charge,
		// and failing the response now would be wrong.
		return err
	}
	return s.quotas.IncrementMessageCount(ctx, userID, personaID)
}

// ChatEnergy is the read-only "Günün sohbet enerjisi" figure Keşfet
// shows as a single number, even though quota is tracked per persona.
type ChatEnergy struct {
	Remaining int
	Limit     int
	Used      int
	MoodLabel string
}

// ChatEnergy reports min(remaining) across every active persona — the
// persona closest to running out drives the number Keşfet shows, since
// showing anything more optimistic would misrepresent how soon the
// user will actually get rate-limited. It never creates or resets a
// quota row (unlike Check/GetOrCreate): a persona with no row yet is
// simply full, and a row from a previous local day is treated as
// unused without writing that reset back.
func (s *QuotaService) ChatEnergy(ctx context.Context, userID string) (ChatEnergy, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return ChatEnergy{}, err
	}
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	personas, err := s.personas.List(ctx)
	if err != nil {
		return ChatEnergy{}, err
	}

	best := ChatEnergy{Remaining: -1}
	for _, p := range personas {
		if !p.IsActive {
			continue
		}

		limit := s.defaultDailyLimit
		used := 0

		quota, ok, err := s.quotas.GetExisting(ctx, userID, p.ID)
		if err != nil {
			return ChatEnergy{}, err
		}
		if ok {
			limit = quota.DailyLimit
			if isSameLocalDay(quota.LastResetAt, now, loc) {
				used = quota.MessageCountToday
			}
		}

		remaining := limit - used
		if remaining < 0 {
			remaining = 0
		}

		if best.Remaining == -1 || remaining < best.Remaining {
			best = ChatEnergy{Remaining: remaining, Limit: limit, Used: used}
		}
	}

	if best.Remaining == -1 {
		// No active personas at all — nothing to be low on.
		best = ChatEnergy{Remaining: s.defaultDailyLimit, Limit: s.defaultDailyLimit, Used: 0}
	}

	best.MoodLabel = chatEnergyMoodLabel(best.Remaining, best.Limit)
	return best, nil
}

func isSameLocalDay(t, now time.Time, loc *time.Location) bool {
	t = t.In(loc)
	ty, tm, td := t.Date()
	ny, nm, nd := now.Date()
	return ty == ny && tm == nm && td == nd
}

func chatEnergyMoodLabel(remaining, limit int) string {
	if limit <= 0 {
		return "Yorgun"
	}
	ratio := float64(remaining) / float64(limit)
	switch {
	case ratio >= 0.7:
		return "Huzurlu"
	case ratio >= 0.3:
		return "Dengeli"
	default:
		return "Yorgun"
	}
}
