package auth

import (
	"encoding/json"
	"net/http"
	"strings"

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
				writeUnauthorized(w)
				return
			}

			claims, err := tokens.Verify(token)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := WithUserID(r.Context(), claims.UserID)
			ctx = WithIsAnonymous(ctx, claims.IsAnonymous)
			requestlog.SetUserID(ctx, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
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

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "giriş gerekli"})
}
