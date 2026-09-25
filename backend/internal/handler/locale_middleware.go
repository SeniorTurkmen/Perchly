package handler

import (
	"net/http"

	"perchly-backend/internal/apierror"
)

// LocaleMiddleware resolves the caller's preferred language from the
// standard Accept-Language header (RFC 9110) and attaches it to the
// request context via apierror.WithLocale — every writeError call
// downstream, and internal/auth's own 401 writer, reads it back out via
// apierror.LocaleFromContext to translate its error text.
//
// Mount this first, before Recoverer and every other middleware — a
// rejection from auth.Middleware or QuotaMiddleware still needs the
// locale to translate its own error response, so this must run before
// any of them, not just before route handlers.
func LocaleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := apierror.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
		ctx := apierror.WithLocale(r.Context(), locale)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
