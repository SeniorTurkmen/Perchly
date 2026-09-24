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

// maxCapturedResponseBody bounds how much of a response body gets
// buffered for logging — enough to see the shape of a typical JSON
// response without holding an entire long SSE chat stream (see
// message_handler.go) in memory for the life of the request.
const maxCapturedResponseBody = 16 * 1024

// teeResponseWriter copies every byte written to the real response into
// a bounded in-memory buffer too, so buildRecord can log what was
// actually sent. Embeds the WrapResponseWriter interface (not the
// concrete *http.ResponseWriter) so Status()/BytesWritten() pass
// through unchanged; Flush is implemented explicitly because it isn't
// part of that interface and message_handler.go's SSE streaming
// type-asserts for it directly on whatever ResponseWriter it's given.
type teeResponseWriter struct {
	chimiddleware.WrapResponseWriter
	captured bytes.Buffer
}

func (t *teeResponseWriter) Write(b []byte) (int, error) {
	if remaining := maxCapturedResponseBody - t.captured.Len(); remaining > 0 {
		if remaining > len(b) {
			t.captured.Write(b)
		} else {
			t.captured.Write(b[:remaining])
		}
	}
	return t.WrapResponseWriter.Write(b)
}

func (t *teeResponseWriter) Flush() {
	if f, ok := t.WrapResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
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

			ww := &teeResponseWriter{WrapResponseWriter: chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)}
			start := time.Now()

			next.ServeHTTP(ww, r.WithContext(ctx))

			duration := time.Since(start)
			truncated := ww.BytesWritten() > ww.captured.Len()
			record := buildRecord(r, e, bodyBytes, ww.Status(), ww.Header(), ww.captured.Bytes(), truncated, duration, redact)

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

func buildRecord(r *http.Request, e *entry, bodyBytes []byte, status int, responseHeader http.Header, responseBodyBytes []byte, responseBodyTruncated bool, duration time.Duration, redact bool) model.RequestLog {
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

	requestHeaders := headersToMap(r.Header)
	responseHeaders := headersToMap(responseHeader)

	if redact {
		queryParams = redactSensitiveFields(queryParams).(map[string]any)
		if body != nil {
			body = redactSensitiveFields(body)
		}
		requestHeaders = redactHeaders(requestHeaders)
		responseHeaders = redactHeaders(responseHeaders)
	}

	pattern := routePattern(r)

	var responseBody *string
	if !skipResponseBodyRoutes[pattern] {
		responseBody = buildResponseBody(responseBodyBytes, responseBodyTruncated, redact)
	}

	return model.RequestLog{
		UserID:          e.userID,
		Method:          r.Method,
		RoutePattern:    pattern,
		Path:            r.URL.Path,
		QueryParams:     queryParams,
		RouteParams:     routeParams,
		Body:            body,
		ResponseBody:    responseBody,
		StatusCode:      status,
		DurationMS:      duration.Milliseconds(),
		RequestHeaders:  requestHeaders,
		ResponseHeaders: responseHeaders,
		ClientVersion:   device.clientVersion,
		Platform:        device.platform,
		OSVersion:       device.osVersion,
		DeviceModel:     device.deviceModel,
		UserAgent:       r.UserAgent(),
		IPAddress:       r.RemoteAddr,
	}
}

// headersToMap flattens an http.Header (map[string][]string) into the
// same shape queryParams already uses: a single value stored bare, more
// than one (e.g. repeated Set-Cookie) as an array — so both render the
// same way in the JSON column and the admin UI.
func headersToMap(h http.Header) map[string]any {
	m := make(map[string]any, len(h))
	for k, v := range h {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}
	return m
}

// skipResponseBodyRoutes are endpoints whose own response is never
// worth capturing as response_body — chiefly the admin dashboard's own
// log-listing endpoints, whose response is a page of other log rows.
// Without this, viewing /admin/logs gets logged too, its response_body
// embeds the previous page's rows (each carrying its own response_body
// string), and that recurses into an unreadable, ever-growing blob a
// few requests in.
var skipResponseBodyRoutes = map[string]bool{
	"/admin/logs":     true,
	"/admin/activity": true,
}

// buildResponseBody turns the captured (possibly truncated) response
// bytes into what gets stored. A JSON response is decoded, redacted the
// same way the request body is, and re-encoded — this is what keeps a
// session/access/refresh token out of the log for POST /auth/anonymous,
// /admin/auth/login, etc. Anything that isn't JSON (an SSE chat stream,
// a plain-text error) is stored verbatim, since there's no structured
// field to redact. Nil for an empty response.
func buildResponseBody(raw []byte, truncated bool, redact bool) *string {
	if len(raw) == 0 {
		return nil
	}

	var text string
	var parsed any
	if err := json.Unmarshal(raw, &parsed); err == nil {
		if redact {
			parsed = redactSensitiveFields(parsed)
		}
		encoded, err := json.Marshal(parsed)
		if err != nil {
			// Shouldn't happen for a value we just decoded, but fall
			// back to the raw bytes rather than lose the log entirely.
			text = string(raw)
		} else {
			text = string(encoded)
		}
	} else {
		text = string(raw)
	}

	if truncated {
		text += "...[TRUNCATED]"
	}
	return &text
}

func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if pattern := rctx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}
