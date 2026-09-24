package model

import "time"

// RequestLog is one recorded API request: who made it (if authenticated),
// which endpoint, with which parameters, from which client, and how it
// resolved. Written by internal/requestlog's middleware.
type RequestLog struct {
	ID           int64          `json:"id"`
	UserID       *string        `json:"user_id"`
	Method       string         `json:"method"`
	RoutePattern string         `json:"route_pattern"`
	Path         string         `json:"path"`
	QueryParams  map[string]any `json:"query_params"`
	RouteParams  map[string]any `json:"route_params"`
	// Body is nil when the request had no body. Sensitive fields (code,
	// password, token, content, ...) are replaced with "[REDACTED]"
	// before this is set, unless redaction was explicitly disabled — see
	// internal/requestlog/redact.go.
	Body       any   `json:"body"`
	StatusCode int   `json:"status_code"`
	DurationMS int64 `json:"duration_ms"`
	// RequestHeaders/ResponseHeaders are every header on the request and
	// the response, keyed by header name — a single value is stored bare,
	// multiple values (e.g. repeated Set-Cookie) as an array. Also
	// subject to redaction (see internal/requestlog/redact.go); an
	// Authorization or Cookie header's value is always "[REDACTED]" when
	// redaction is on.
	RequestHeaders  map[string]any `json:"request_headers"`
	ResponseHeaders map[string]any `json:"response_headers"`
	// ResponseBody is the raw response as sent to the client, truncated
	// to a bounded size (see internal/requestlog/middleware.go). When it
	// parses as JSON, it's re-encoded after redaction (same sensitive
	// field names as Body/headers — this is how POST /auth/anonymous,
	// /admin/auth/login etc.'s access/session tokens stay out of the
	// log); a non-JSON response (an SSE chat stream, a plain-text error)
	// is stored as-is, unredacted, since there's no structured field to
	// target. Nil when the response had no body.
	ResponseBody  *string   `json:"response_body"`
	ClientVersion *string   `json:"client_version"`
	Platform      *string   `json:"platform"`
	OSVersion     *string   `json:"os_version"`
	DeviceModel   *string   `json:"device_model"`
	UserAgent     string    `json:"user_agent"`
	IPAddress     string    `json:"ip_address"`
	CreatedAt     time.Time `json:"created_at"`
}

// RequestLogEntry is a RequestLog with its user resolved to something a
// human can recognize — used only by the admin dashboard's log viewer,
// which needs to show "who" beyond a bare UUID without a second lookup
// per row. UserEmail/UserDisplayName are nil for an anonymous user or
// an unauthenticated request (UserID itself is nil then too).
type RequestLogEntry struct {
	RequestLog
	UserEmail       *string `json:"user_email"`
	UserDisplayName *string `json:"user_display_name"`
}
