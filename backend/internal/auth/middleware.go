package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"perchly-backend/internal/apierror"
	"perchly-backend/internal/requestlog"
)

// TokenVerifier is the subset of AccessTokenIssuer that Middleware needs,
// so tests/other callers can substitute a fake.
type TokenVerifier interface {
	Verify(token string) (AccessClaims, error)
}

// Middleware requires a valid `Authorization: Bearer <access token>`
// header and puts the token's user id and is_anonymous flag in the
// request context (see UserIDFromContext / IsAnonymousFromContext).
// Routes that don't need a signed-in user (e.g. /personas) should not
// use this. Only ever verifies access tokens — refresh tokens are opaque
// and go through AuthService.RefreshSession instead, never this
// middleware.
func Middleware(tokens TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				writeUnauthorized(w, r)
				return
			}

			claims, err := tokens.Verify(token)
			if err != nil {
				writeUnauthorized(w, r)
				return
			}

			ctx := WithUserID(r.Context(), claims.UserID)
			ctx = WithIsAnonymous(ctx, claims.IsAnonymous)
			requestlog.SetUserID(ctx, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalMiddleware behaves like Middleware when a valid bearer token
// is present — same user id/is_anonymous in context — but, unlike
// Middleware, never rejects the request for one being absent or
// invalid; it just proceeds without them. For routes that are public
// by default but have one specific, query-param-gated behavior that
// needs a signed-in user (see GET /personas?recommend=true) — the
// handler itself checks UserIDFromContext and decides whether that
// specific behavior requires rejecting the request.
func OptionalMiddleware(tokens TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token, ok := bearerToken(r); ok {
				if claims, err := tokens.Verify(token); err == nil {
					ctx := WithUserID(r.Context(), claims.UserID)
					ctx = WithIsAnonymous(ctx, claims.IsAnonymous)
					requestlog.SetUserID(ctx, claims.UserID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	locale := apierror.LocaleFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(apierror.New(apierror.CodeUnauthorized, apierror.Message(apierror.CodeUnauthorized, locale, "giriş gerekli")))
}
