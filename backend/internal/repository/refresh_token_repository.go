package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrRefreshTokenNotFound is returned when a presented refresh token
// doesn't match any stored hash.
var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (model.RefreshToken, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
		RETURNING id::text, user_id::text, token_hash, expires_at, revoked_at, created_at
	`, userID, tokenHash, expiresAt)
	return scanRefreshToken(row)
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (model.RefreshToken, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, user_id::text, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash)

	t, err := scanRefreshToken(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrRefreshTokenNotFound
		}
		return model.RefreshToken{}, err
	}
	return t, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now() WHERE id = $1::uuid AND revoked_at IS NULL
	`, id)
	return err
}

// RevokeAllForUser revokes every still-active refresh token for a user.
// Used as the security response when a refresh token is reused after
// already being rotated away (see AuthService.RefreshSession) — that
// pattern indicates the token may have leaked, so every session for the
// user is force-logged-out rather than just the one reused token.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1::uuid AND revoked_at IS NULL
	`, userID)
	return err
}

func scanRefreshToken(row rowScanner) (model.RefreshToken, error) {
	var t model.RefreshToken
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt)
	return t, err
}
