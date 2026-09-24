package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

type AdminMetricsRepository struct {
	pool *pgxpool.Pool
}

func NewAdminMetricsRepository(pool *pgxpool.Pool) *AdminMetricsRepository {
	return &AdminMetricsRepository{pool: pool}
}

// Snapshot computes the admin dashboard home page's numbers in one
// round trip — each figure is its own scalar subquery over an
// unrelated table, so there's no join to get wrong; only "today" (UTC)
// has to line up across them, which `current_date` guarantees.
func (r *AdminMetricsRepository) Snapshot(ctx context.Context) (model.AdminMetrics, error) {
	var m model.AdminMetrics
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM users),
			(SELECT count(*) FROM users WHERE created_at::date = current_date),
			(SELECT count(*) FROM messages WHERE created_at::date = current_date),
			(SELECT count(DISTINCT conversation_id) FROM messages WHERE created_at::date = current_date),
			(SELECT coalesce(
				count(*) FILTER (WHERE status_code >= 500)::float8 / nullif(count(*), 0), 0
			) FROM request_logs WHERE created_at::date = current_date)
	`).Scan(
		&m.TotalUsers, &m.NewUsersToday, &m.MessagesSentToday,
		&m.ActiveConversationsToday, &m.ErrorRateToday,
	)
	return m, err
}
