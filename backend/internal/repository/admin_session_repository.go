package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrAdminSessionNotFound is returned when a presented admin session
// token doesn't match any stored hash.
var ErrAdminSessionNotFound = errors.New("admin session not found")

type AdminSessionRepository struct {
	pool *pgxpool.Pool
}

func NewAdminSessionRepository(pool *pgxpool.Pool) *AdminSessionRepository {
	return &AdminSessionRepository{pool: pool}
}

const adminSessionColumns = `id::text, admin_user_id::text, token_hash, expires_at, revoked_at, created_at`

func (r *AdminSessionRepository) Create(ctx context.Context, adminUserID, tokenHash string, expiresAt time.Time) (model.AdminSession, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO admin_sessions (admin_user_id, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
		RETURNING `+adminSessionColumns, adminUserID, tokenHash, expiresAt)
	return scanAdminSession(row)
}

func (r *AdminSessionRepository) GetByHash(ctx context.Context, tokenHash string) (model.AdminSession, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+adminSessionColumns+` FROM admin_sessions WHERE token_hash = $1`, tokenHash)

	s, err := scanAdminSession(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AdminSession{}, ErrAdminSessionNotFound
		}
		return model.AdminSession{}, err
	}
	return s, nil
}

func (r *AdminSessionRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE admin_sessions SET revoked_at = now() WHERE id = $1::uuid AND revoked_at IS NULL
	`, id)
	return err
}

func scanAdminSession(row rowScanner) (model.AdminSession, error) {
	var s model.AdminSession
	err := row.Scan(&s.ID, &s.AdminUserID, &s.TokenHash, &s.ExpiresAt, &s.RevokedAt, &s.CreatedAt)
	return s, err
}
