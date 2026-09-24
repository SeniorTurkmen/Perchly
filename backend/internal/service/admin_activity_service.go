package service

import (
	"context"

	"perchly-backend/internal/model"
)

type AdminActivityRepo interface {
	ListForAdmin(ctx context.Context, adminUserID *string, limit, offset int) ([]model.AdminAuditLog, int, error)
}

// AdminActivityService backs the admin dashboard's own activity log —
// every admin.* action recorded to admin_audit_log by the other admin
// services, surfaced here read-only. This is the "which admin did
// what" view, distinct from AdminLogService's "which app user hit
// which endpoint" request-log view.
type AdminActivityService struct {
	activity AdminActivityRepo
}

func NewAdminActivityService(activity AdminActivityRepo) *AdminActivityService {
	return &AdminActivityService{activity: activity}
}

func (s *AdminActivityService) List(ctx context.Context, adminUserID *string, limit, offset int) ([]model.AdminAuditLog, int, error) {
	return s.activity.ListForAdmin(ctx, adminUserID, limit, offset)
}
