package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminAuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAdminAuditLogRepository(pool *pgxpool.Pool) *AdminAuditLogRepository {
	return &AdminAuditLogRepository{pool: pool}
}

// Create records one admin action. targetID and detail are optional
// (nil is fine — not every action, e.g. a login, has a single target
// resource or extra detail worth storing). detail is JSON-encoded here,
// not left to pgx, for the same reason as RequestLogRepository.Create —
// see marshalNullableJSON's doc.
func (r *AdminAuditLogRepository) Create(ctx context.Context, adminUserID, action, targetType string, targetID *string, detail map[string]any) error {
	detailJSON, err := marshalNullableJSON(detail)
	if err != nil {
		return fmt.Errorf("marshal detail: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO admin_audit_log (admin_user_id, action, target_type, target_id, detail)
		VALUES ($1::uuid, $2, $3, $4, $5::jsonb)
	`, adminUserID, action, targetType, targetID, detailJSON)
	return err
}
