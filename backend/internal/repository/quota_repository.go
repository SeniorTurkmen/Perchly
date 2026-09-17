package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

type QuotaRepository struct {
	pool *pgxpool.Pool
}

func NewQuotaRepository(pool *pgxpool.Pool) *QuotaRepository {
	return &QuotaRepository{pool: pool}
}

// GetOrCreate returns the user's quota row for a persona, creating one
// with defaultDailyLimit if this is their first message to that persona.
//
// It also resets message_count_today first if the row's owner has
// crossed into a new local day since last_reset_at — the same
// comparison QuotaResetJob runs in the background, done here too so a
// quota check is always correct even between two ticks of that job (or
// if it hasn't started yet).
func (r *QuotaRepository) GetOrCreate(ctx context.Context, userID, personaID string, defaultDailyLimit int) (model.UserQuota, error) {
	if _, err := r.pool.Exec(ctx, `
		UPDATE user_quotas uq
		SET message_count_today = 0,
		    last_reset_at = now()
		FROM users u
		WHERE uq.user_id = $1::uuid
		  AND uq.persona_id = $2::uuid
		  AND uq.user_id = u.id
		  AND (now() AT TIME ZONE u.timezone)::date > (uq.last_reset_at AT TIME ZONE u.timezone)::date
	`, userID, personaID); err != nil {
		return model.UserQuota{}, err
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO user_quotas (user_id, persona_id, daily_limit)
		VALUES ($1::uuid, $2::uuid, $3)
		ON CONFLICT (user_id, persona_id) DO UPDATE SET user_id = user_quotas.user_id
		RETURNING user_id::text, persona_id::text, message_count_today, daily_limit, last_reset_at
	`, userID, personaID, defaultDailyLimit)

	var q model.UserQuota
	err := row.Scan(&q.UserID, &q.PersonaID, &q.MessageCountToday, &q.DailyLimit, &q.LastResetAt)
	return q, err
}

// IncrementMessageCount records one message against today's usage.
func (r *QuotaRepository) IncrementMessageCount(ctx context.Context, userID, personaID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE user_quotas
		SET message_count_today = message_count_today + 1
		WHERE user_id = $1::uuid AND persona_id = $2::uuid
	`, userID, personaID)
	return err
}

// ResetDueForNewLocalDay zeroes message_count_today for every quota row
// whose owning user has crossed into a new calendar day in their own
// timezone since last_reset_at, computed entirely in SQL via `AT TIME
// ZONE` so one query handles every user regardless of timezone. Returns
// how many rows were reset.
func (r *QuotaRepository) ResetDueForNewLocalDay(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE user_quotas uq
		SET message_count_today = 0,
		    last_reset_at = now()
		FROM users u
		WHERE uq.user_id = u.id
		  AND (now() AT TIME ZONE u.timezone)::date > (uq.last_reset_at AT TIME ZONE u.timezone)::date
	`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
