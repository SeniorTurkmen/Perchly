package auth

import "context"

type contextKey string

const (
	userIDContextKey      contextKey = "userID"
	isAnonymousContextKey contextKey = "isAnonymous"
)

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// UserIDFromContext returns the authenticated user's id, set by
// Middleware. ok is false if called on a request Middleware never ran on.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
	return id, ok
}

func WithIsAnonymous(ctx context.Context, isAnonymous bool) context.Context {
	return context.WithValue(ctx, isAnonymousContextKey, isAnonymous)
}

// IsAnonymousFromContext reports whether the authenticated user's token
// was anonymous, set by Middleware. ok is false if called on a request
// Middleware never ran on.
func IsAnonymousFromContext(ctx context.Context) (bool, bool) {
	v, ok := ctx.Value(isAnonymousContextKey).(bool)
	return v, ok
}
