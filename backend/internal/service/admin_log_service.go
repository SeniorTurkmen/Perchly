package service

import (
	"context"

	"perchly-backend/internal/model"
)

type AdminLogRepo interface {
	ListForAdmin(ctx context.Context, search string, statusMin int, userID *string, limit, offset int) ([]model.RequestLogEntry, int, error)
}

// AdminLogService backs the admin dashboard's request-log viewer —
// read-only, so no audit logging (nothing here mutates anything).
type AdminLogService struct {
	logs AdminLogRepo
}

func NewAdminLogService(logs AdminLogRepo) *AdminLogService {
	return &AdminLogService{logs: logs}
}

func (s *AdminLogService) List(ctx context.Context, search string, statusMin int, userID *string, limit, offset int) ([]model.RequestLogEntry, int, error) {
	return s.logs.ListForAdmin(ctx, search, statusMin, userID, limit, offset)
}
