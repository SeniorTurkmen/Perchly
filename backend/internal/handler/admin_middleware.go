package handler

import (
	"context"
	"net/http"
	"strings"

	"perchly-backend/internal/auth"
	"perchly-backend/internal/model"
)

type adminSessionVerifier interface {
	VerifySession(ctx context.Context, rawToken string) (model.AdminUser, error)
}

// AdminMiddleware requires a valid `Authorization: Bearer <admin session
// token>` header and puts the admin's id in the request context (see
// auth.AdminUserIDFromContext). Unlike auth.Middleware (which verifies a
// self-contained JWT), this always does a DB round-trip — admin sessions
// are opaque tokens, checked and revocable like refresh tokens.
func AdminMiddleware(sessions adminSessionVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := adminBearerToken(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, ErrCodeAdminUnauthorized, "admin girişi gerekli")
				return
			}

			admin, err := sessions.VerifySession(r.Context(), token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, ErrCodeAdminUnauthorized, "admin girişi gerekli")
				return
			}

			ctx := auth.WithAdminUserID(r.Context(), admin.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func adminBearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
