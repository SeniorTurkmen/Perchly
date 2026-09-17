package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthService reports on the health of the service's dependencies.
type HealthService struct {
	pool *pgxpool.Pool
}

func NewHealthService(pool *pgxpool.Pool) *HealthService {
	return &HealthService{pool: pool}
}

// CheckDatabase pings the database to confirm it is reachable.
func (s *HealthService) CheckDatabase(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
