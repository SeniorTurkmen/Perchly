package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

// ErrAdminUserNotFound is returned when no admin user matches the given
// lookup.
var ErrAdminUserNotFound = errors.New("admin user not found")

type AdminUserRepository struct {
	pool *pgxpool.Pool
}

func NewAdminUserRepository(pool *pgxpool.Pool) *AdminUserRepository {
	return &AdminUserRepository{pool: pool}
}

const adminUserColumns = `id::text, email, password_hash, created_at`

func (r *AdminUserRepository) GetByEmail(ctx context.Context, email string) (model.AdminUser, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+adminUserColumns+` FROM admin_users WHERE email = $1`, email)
	return scanAdminUserOrNotFound(row)
}

func (r *AdminUserRepository) GetByID(ctx context.Context, id string) (model.AdminUser, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+adminUserColumns+` FROM admin_users WHERE id = $1::uuid`, id)
	return scanAdminUserOrNotFound(row)
}

// Upsert creates a new admin user or, if the email already exists,
// resets its password — used by the createadmin CLI (cmd/createadmin),
// the only way admin accounts are provisioned today.
func (r *AdminUserRepository) Upsert(ctx context.Context, email, passwordHash string) (model.AdminUser, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO admin_users (email, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET password_hash = excluded.password_hash
		RETURNING `+adminUserColumns, email, passwordHash)
	return scanAdminUser(row)
}

func scanAdminUserOrNotFound(row rowScanner) (model.AdminUser, error) {
	u, err := scanAdminUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AdminUser{}, ErrAdminUserNotFound
		}
		return model.AdminUser{}, err
	}
	return u, nil
}

func scanAdminUser(row rowScanner) (model.AdminUser, error) {
	var u model.AdminUser
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	return u, err
}
