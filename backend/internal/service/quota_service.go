package service

import (
	"context"

	"perchly-backend/internal/model"
)

type QuotaRepo interface {
	GetOrCreate(ctx context.Context, userID, personaID string, defaultDailyLimit int) (model.UserQuota, error)
	IncrementMessageCount(ctx context.Context, userID, personaID string) error
}

type CreditRepo interface {
	GetBalance(ctx context.Context, userID string) (int, error)
	SpendOne(ctx context.Context, userID string) (bool, error)
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
	defaultDailyLimit int
}

func NewQuotaService(quotas QuotaRepo, credits CreditRepo, defaultDailyLimit int) *QuotaService {
	return &QuotaService{quotas: quotas, credits: credits, defaultDailyLimit: defaultDailyLimit}
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
