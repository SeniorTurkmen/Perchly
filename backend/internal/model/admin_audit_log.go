package model

import "time"

// AdminAuditLog is one recorded admin action — who did what to which
// resource, and when. Written by every mutating admin handler so
// production data changes can be traced back to a specific admin.
// AdminEmail is only populated by AdminAuditLogRepository.ListForAdmin
// (a join, for the admin dashboard's own activity view) — never set by
// Create, which only has the id to write.
type AdminAuditLog struct {
	ID          string         `json:"id"`
	AdminUserID string         `json:"admin_user_id"`
	AdminEmail  string         `json:"admin_email,omitempty"`
	Action      string         `json:"action"`
	TargetType  string         `json:"target_type"`
	TargetID    *string        `json:"target_id"`
	Detail      map[string]any `json:"detail"`
	CreatedAt   time.Time      `json:"created_at"`
}
