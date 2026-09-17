package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrUserNotFound is returned when no user matches the given lookup.
var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

const userColumns = `id::text, email, email_verified_at, display_name, is_anonymous, device_id, timezone, created_at, updated_at`

func (r *UserRepository) GetByID(ctx context.Context, id string) (model.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1::uuid`, id)
	return scanUserOrNotFound(row)
}

func (r *UserRepository) GetByDeviceID(ctx context.Context, deviceID string) (model.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE device_id = $1`, deviceID)
	return scanUserOrNotFound(row)
}

// GetVerifiedByEmail returns the user owning email, but only if that
// email has actually been verified — an unverified/abandoned row (which
// shouldn't normally exist, since verification is what sets email in the
// first place) is treated as not found.
func (r *UserRepository) GetVerifiedByEmail(ctx context.Context, email string) (model.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+userColumns+` FROM users
		WHERE email = $1 AND email_verified_at IS NOT NULL
	`, email)
	return scanUserOrNotFound(row)
}

// CreateAnonymous creates a new anonymous user. deviceID may be empty
// (no device-based lookup will ever find it, so a fresh anonymous
// session is created every time for that caller — acceptable, since the
// client is expected to always send its device id).
func (r *UserRepository) CreateAnonymous(ctx context.Context, deviceID, timezone string) (model.User, error) {
	timezone = normalizeTimezone(timezone)

	var deviceIDArg any
	if deviceID != "" {
		deviceIDArg = deviceID
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (device_id, timezone, is_anonymous)
		VALUES ($1, $2, true)
		RETURNING `+userColumns, deviceIDArg, timezone)
	return scanUser(row)
}

// CreateVerifiedEmailUser creates a brand-new, already-verified user for
// an email with no prior anonymous session to upgrade (Durum B without
// an anonymous token).
func (r *UserRepository) CreateVerifiedEmailUser(ctx context.Context, email, timezone string) (model.User, error) {
	timezone = normalizeTimezone(timezone)

	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (email, email_verified_at, is_anonymous, timezone)
		VALUES ($1, now(), false, $2)
		RETURNING `+userColumns, email, timezone)
	return scanUser(row)
}

// UpgradeAnonymousToVerifiedEmail turns an existing anonymous user into a
// verified account in place (Durum B with an anonymous token) — the
// user's id, and everything tied to it (conversations, quota, credits),
// carries over untouched.
func (r *UserRepository) UpgradeAnonymousToVerifiedEmail(ctx context.Context, userID, email string) (model.User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET email = $2, email_verified_at = now(), is_anonymous = false
		WHERE id = $1::uuid
		RETURNING `+userColumns, userID, email)
	return scanUserOrNotFound(row)
}

func (r *UserRepository) UpdateDisplayName(ctx context.Context, userID, displayName string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET display_name = $2 WHERE id = $1::uuid`, userID, displayName)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func normalizeTimezone(timezone string) string {
	if _, err := time.LoadLocation(timezone); timezone == "" || err != nil {
		return "UTC"
	}
	return timezone
}

func scanUserOrNotFound(row rowScanner) (model.User, error) {
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}
	return u, nil
}

func scanUser(row rowScanner) (model.User, error) {
	var u model.User
	err := row.Scan(
		&u.ID, &u.Email, &u.EmailVerifiedAt, &u.DisplayName, &u.IsAnonymous, &u.DeviceID,
		&u.Timezone, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}
