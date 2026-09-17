package model

import "time"

// RequestLog is one recorded API request: who made it (if authenticated),
// which endpoint, with which parameters, from which client, and how it
// resolved. Written by internal/requestlog's middleware.
type RequestLog struct {
	ID           int64
	UserID       *string
	Method       string
	RoutePattern string
	Path         string
	QueryParams  map[string]any
	RouteParams  map[string]any
	// Body is nil when the request had no body. Sensitive fields (code,
	// password, token, content, ...) are replaced with "[REDACTED]"
	// before this is set, unless redaction was explicitly disabled — see
	// internal/requestlog/redact.go.
	Body          any
	StatusCode    int
	DurationMS    int64
	ClientVersion *string
	Platform      *string
	OSVersion     *string
	DeviceModel   *string
	UserAgent     string
	IPAddress     string
	CreatedAt     time.Time
}
