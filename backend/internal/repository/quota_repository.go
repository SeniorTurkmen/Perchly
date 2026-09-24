package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// GetExisting returns the raw quota row for a user/persona pair without
// creating one and without resetting it for a new local day — unlike
// GetOrCreate, this is strictly read-only, for reporting (e.g. chat
// energy) where a stale row's rollover must be accounted for by the
// caller (who knows the user's timezone) rather than mutated here.
// ok is false when no row exists yet, meaning that persona's quota is
// still fully unused.
func (r *QuotaRepository) GetExisting(ctx context.Context, userID, personaID string) (model.UserQuota, bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id::text, persona_id::text, message_count_today, daily_limit, last_reset_at
		FROM user_quotas
		WHERE user_id = $1::uuid AND persona_id = $2::uuid
	`, userID, personaID)

	var q model.UserQuota
	if err := row.Scan(&q.UserID, &q.PersonaID, &q.MessageCountToday, &q.DailyLimit, &q.LastResetAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserQuota{}, false, nil
		}
		return model.UserQuota{}, false, err
	}
	return q, true, nil
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

// ListForUser returns every persona this user has a quota row for —
// used by the admin dashboard's user detail view. A persona the user
// has never messaged has no row and so doesn't appear here.
func (r *QuotaRepository) ListForUser(ctx context.Context, userID string) ([]model.UserQuota, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT user_id::text, persona_id::text, message_count_today, daily_limit, last_reset_at
		FROM user_quotas
		WHERE user_id = $1::uuid
		ORDER BY last_reset_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotas := make([]model.UserQuota, 0)
	for rows.Next() {
		var q model.UserQuota
		if err := rows.Scan(&q.UserID, &q.PersonaID, &q.MessageCountToday, &q.DailyLimit, &q.LastResetAt); err != nil {
			return nil, err
		}
		quotas = append(quotas, q)
	}
	return quotas, rows.Err()
}

// SetDailyLimit overrides a user's daily message limit for one persona
// — the exact operation user_quotas.daily_limit's own doc comment
// anticipates ("a future admin dashboard can grant an individual
// override by updating this column directly"). Creates the row if the
// user has never messaged that persona yet, since there's nothing to
// override before that. Returns ErrPersonaNotFound if personaID
// doesn't exist.
func (r *QuotaRepository) SetDailyLimit(ctx context.Context, userID, personaID string, dailyLimit int) (model.UserQuota, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO user_quotas (user_id, persona_id, daily_limit)
		VALUES ($1::uuid, $2::uuid, $3)
		ON CONFLICT (user_id, persona_id) DO UPDATE SET daily_limit = excluded.daily_limit
		RETURNING user_id::text, persona_id::text, message_count_today, daily_limit, last_reset_at
	`, userID, personaID, dailyLimit)

	var q model.UserQuota
	err := row.Scan(&q.UserID, &q.PersonaID, &q.MessageCountToday, &q.DailyLimit, &q.LastResetAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign_key_violation
			return model.UserQuota{}, ErrPersonaNotFound
		}
		return model.UserQuota{}, err
	}
	return q, nil
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
