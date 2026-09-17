package requestlog

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"perchly-backend/internal/model"
)

// Writer persists a completed request's log row. Satisfied by
// *repository.RequestLogRepository.
type Writer interface {
	Create(ctx context.Context, entry model.RequestLog) error
}

// Middleware logs every request: method, route, query/route params,
// (optionally redacted) body, device headers, status, and duration.
// Persisting is asynchronous — on its own background context, in a
// goroutine — so a slow or unavailable database never adds latency to
// the response, and a logging failure never fails the request.
//
// Mount this OUTSIDE (registered before) chi's Recoverer: that way, if a
// handler panics, Recoverer's recover() runs in a frame between the
// panic and this middleware, and unwinds normally back up through this
// middleware's `next.ServeHTTP` call — letting Middleware still observe
// and log the final (500) status, instead of the panic skipping past it
// entirely. Mounting it after Recoverer would mean a panicking request
// is never logged.
//
// redact controls whether known-sensitive field values (see
// isSensitiveField) are replaced with "[REDACTED]" before the row is
// written. Toggle it via config (LOG_REDACT_SENSITIVE_FIELDS) — leave it
// on by default, and only turn it off for a short, deliberate debugging
// session, never in a shared or production environment.
func Middleware(writer Writer, redact bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			e := &entry{}
			ctx := withEntry(r.Context(), e)

			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes)) // restore for downstream handlers

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r.WithContext(ctx))

			duration := time.Since(start)
			record := buildRecord(r, e, bodyBytes, ww.Status(), duration, redact)

			go func() {
				bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := writer.Create(bgCtx, record); err != nil {
					log.Printf("requestlog: failed to persist entry: %v", err)
				}
			}()
		})
	}
}

func buildRecord(r *http.Request, e *entry, bodyBytes []byte, status int, duration time.Duration, redact bool) model.RequestLog {
	device := parseDeviceInfo(r)

	queryParams := map[string]any{}
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			queryParams[k] = v[0]
		} else {
			queryParams[k] = v
		}
	}

	routeParams := map[string]any{}
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		for i, key := range rctx.URLParams.Keys {
			routeParams[key] = rctx.URLParams.Values[i]
		}
	}

	var body any
	if len(bodyBytes) > 0 {
		var parsed any
		if err := json.Unmarshal(bodyBytes, &parsed); err == nil {
			body = parsed
		}
	}

	if redact {
		queryParams = redactSensitiveFields(queryParams).(map[string]any)
		if body != nil {
			body = redactSensitiveFields(body)
		}
	}

	return model.RequestLog{
		UserID:        e.userID,
		Method:        r.Method,
		RoutePattern:  routePattern(r),
		Path:          r.URL.Path,
		QueryParams:   queryParams,
		RouteParams:   routeParams,
		Body:          body,
		StatusCode:    status,
		DurationMS:    duration.Milliseconds(),
		ClientVersion: device.clientVersion,
		Platform:      device.platform,
		OSVersion:     device.osVersion,
		DeviceModel:   device.deviceModel,
		UserAgent:     r.UserAgent(),
		IPAddress:     r.RemoteAddr,
	}
}

func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if pattern := rctx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}
