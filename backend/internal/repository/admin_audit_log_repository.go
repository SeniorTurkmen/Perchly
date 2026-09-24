package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
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

// ListForAdmin returns a page of admin activity, newest first, joined
// with the acting admin's email — this is the admin dashboard's own
// audit trail of itself. adminUserID narrows to one admin's actions;
// nil means every admin. Also returns the total row count matching the
// filter, for pagination.
func (r *AdminAuditLogRepository) ListForAdmin(ctx context.Context, adminUserID *string, limit, offset int) ([]model.AdminAuditLog, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	const where = `$1::uuid IS NULL OR a.admin_user_id = $1::uuid`

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM admin_audit_log a WHERE `+where, adminUserID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT a.id::text, a.admin_user_id::text, u.email, a.action, a.target_type, a.target_id, a.detail, a.created_at
		FROM admin_audit_log a
		JOIN admin_users u ON u.id = a.admin_user_id
		WHERE `+where+`
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`, adminUserID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entries := make([]model.AdminAuditLog, 0)
	for rows.Next() {
		var e model.AdminAuditLog
		if err := rows.Scan(&e.ID, &e.AdminUserID, &e.AdminEmail, &e.Action, &e.TargetType, &e.TargetID, &e.Detail, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}
