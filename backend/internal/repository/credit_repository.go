package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreditRepository struct {
	pool *pgxpool.Pool
}

func NewCreditRepository(pool *pgxpool.Pool) *CreditRepository {
	return &CreditRepository{pool: pool}
}

// GetBalance returns a user's credit balance. A user with no
// credit_balances row (never granted any credits) has a balance of 0,
// which is not an error.
func (r *CreditRepository) GetBalance(ctx context.Context, userID string) (int, error) {
	var amount int
	err := r.pool.QueryRow(ctx, `
		SELECT credit_amount FROM credit_balances WHERE user_id = $1::uuid
	`, userID).Scan(&amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return amount, nil
}

// SetBalance overrides a user's credit balance directly — used by the
// admin dashboard to grant (or correct) credits. Creates the row if the
// user never had one (credit_balances only gets a row lazily, the
// first time credits are involved).
func (r *CreditRepository) SetBalance(ctx context.Context, userID string, amount int) (int, error) {
	var balance int
	err := r.pool.QueryRow(ctx, `
		INSERT INTO credit_balances (user_id, credit_amount)
		VALUES ($1::uuid, $2)
		ON CONFLICT (user_id) DO UPDATE SET credit_amount = excluded.credit_amount, updated_at = now()
		RETURNING credit_amount
	`, userID, amount).Scan(&balance)
	return balance, err
}

// SpendOne atomically decrements one credit if the user has at least one
// available, reporting whether it actually spent one (false if the
// balance was already 0, e.g. spent by a concurrent request first).
func (r *CreditRepository) SpendOne(ctx context.Context, userID string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE credit_balances
		SET credit_amount = credit_amount - 1, updated_at = now()
		WHERE user_id = $1::uuid AND credit_amount > 0
	`, userID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
