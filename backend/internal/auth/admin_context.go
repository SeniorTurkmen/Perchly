package auth

import "context"

type adminContextKey string

const adminUserIDContextKey adminContextKey = "adminUserID"

func WithAdminUserID(ctx context.Context, adminUserID string) context.Context {
	return context.WithValue(ctx, adminUserIDContextKey, adminUserID)
}

// AdminUserIDFromContext returns the authenticated admin's id, set by
// handler.AdminMiddleware. ok is false if called on a request that
// middleware never ran on.
func AdminUserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(adminUserIDContextKey).(string)
	return id, ok
}
