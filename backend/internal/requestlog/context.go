package requestlog

import "context"

// entry is the mutable, request-scoped record Middleware builds up over
// the life of a request. It's stashed in the request context as a
// pointer (see withEntry/entryFromContext) specifically so code deeper in
// the handler chain — auth.Middleware, namely — can enrich it (SetUserID)
// after Middleware has already called next.ServeHTTP: a plain
// context.WithValue further down the chain would be invisible to
// Middleware's own stack frame once control returns to it, since
// c.WithValue returns a new context rather than mutating the one
// Middleware is holding. A shared pointer's pointee, unlike the context
// wrapping it, is visible from both ends.
type entry struct {
	userID *string
}

type contextKey string

const entryContextKey contextKey = "requestlog.entry"

func withEntry(ctx context.Context, e *entry) context.Context {
	return context.WithValue(ctx, entryContextKey, e)
}

func entryFromContext(ctx context.Context) *entry {
	e, _ := ctx.Value(entryContextKey).(*entry)
	return e
}

// SetUserID records the authenticated user for the current request's log
// entry, if request logging is active on this request (see Middleware).
// Safe to call unconditionally, including on requests Middleware never
// ran on — it's a no-op then.
func SetUserID(ctx context.Context, userID string) {
	if e := entryFromContext(ctx); e != nil {
		e.userID = &userID
	}
}
