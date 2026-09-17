package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrVerificationCodeNotFound is returned when an email has never had a
// code requested for it.
var ErrVerificationCodeNotFound = errors.New("verification code not found")

type VerificationCodeRepository struct {
	pool *pgxpool.Pool
}

func NewVerificationCodeRepository(pool *pgxpool.Pool) *VerificationCodeRepository {
	return &VerificationCodeRepository{pool: pool}
}

// CountSince counts codes requested for email since the given time —
// used by AuthService to enforce the per-minute / per-hour request-code
// rate limits.
func (r *VerificationCodeRepository) CountSince(ctx context.Context, email string, since time.Time) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM email_verification_codes
		WHERE email = $1 AND created_at > $2
	`, email, since).Scan(&count)
	return count, err
}

func (r *VerificationCodeRepository) Create(ctx context.Context, email, code string, expiresAt time.Time) (model.EmailVerificationCode, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO email_verification_codes (email, code, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, code, expires_at, attempt_count, used_at, created_at
	`, email, code, expiresAt)
	return scanVerificationCode(row)
}

// GetLatest returns the most recently requested code for email,
// regardless of whether it has expired or already been used — the
// caller (AuthService) decides what those states mean, so it can return
// a specific, actionable error ("expired", "already used", "too many
// attempts") instead of a generic "not found".
func (r *VerificationCodeRepository) GetLatest(ctx context.Context, email string) (model.EmailVerificationCode, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, email, code, expires_at, attempt_count, used_at, created_at
		FROM email_verification_codes
		WHERE email = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, email)

	code, err := scanVerificationCode(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EmailVerificationCode{}, ErrVerificationCodeNotFound
		}
		return model.EmailVerificationCode{}, err
	}
	return code, nil
}

func (r *VerificationCodeRepository) IncrementAttempt(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE email_verification_codes SET attempt_count = attempt_count + 1 WHERE id = $1::uuid
	`, id)
	return err
}

func (r *VerificationCodeRepository) MarkUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE email_verification_codes SET used_at = now() WHERE id = $1::uuid
	`, id)
	return err
}

func scanVerificationCode(row rowScanner) (model.EmailVerificationCode, error) {
	var c model.EmailVerificationCode
	err := row.Scan(&c.ID, &c.Email, &c.Code, &c.ExpiresAt, &c.AttemptCount, &c.UsedAt, &c.CreatedAt)
	return c, err
}
