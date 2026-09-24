package requestlog

import "strings"

// sensitiveFieldNames are body/query field names whose values get
// replaced with "[REDACTED]" before a log row is written, when redaction
// is enabled (see Middleware's redact parameter, LOG_REDACT_SENSITIVE_FIELDS
// in config). Matched as a case-insensitive substring, so "code",
// "verification_code", "access_token", "refresh_token", "password" etc.
// are all caught by this short list.
var sensitiveFieldNames = []string{
	"code",
	"password",
	"token",
	"secret",
	"content",
}

func isSensitiveField(name string) bool {
	lower := strings.ToLower(name)
	for _, s := range sensitiveFieldNames {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// sensitiveHeaderNames is a separate, narrower list for header redaction
// — reusing sensitiveFieldNames' "content" entry (there for body/query
// fields like message content) would also catch harmless headers like
// Content-Type or Content-Length, which carry no secret.
var sensitiveHeaderNames = []string{
	"authorization",
	"cookie",
}

func isSensitiveHeader(name string) bool {
	lower := strings.ToLower(name)
	for _, s := range sensitiveHeaderNames {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// redactHeaders replaces the value of any header matching
// isSensitiveHeader with "[REDACTED]" — a header map's values are
// always scalars (or an array of scalars for a repeated header, e.g.
// Set-Cookie), never nested objects, so unlike redactSensitiveFields
// this doesn't need to recurse.
func redactHeaders(headers map[string]any) map[string]any {
	result := make(map[string]any, len(headers))
	for k, v := range headers {
		if isSensitiveHeader(k) {
			result[k] = "[REDACTED]"
		} else {
			result[k] = v
		}
	}
	return result
}

// redactSensitiveFields walks a decoded JSON value (as produced by
// encoding/json's default unmarshal-into-any: map[string]any, []any, or
// a scalar) and replaces the value of any object key matching
// isSensitiveField with "[REDACTED]", recursively.
func redactSensitiveFields(v any) any {
	switch value := v.(type) {
	case map[string]any:
		result := make(map[string]any, len(value))
		for k, val := range value {
			if isSensitiveField(k) {
				result[k] = "[REDACTED]"
			} else {
				result[k] = redactSensitiveFields(val)
			}
		}
		return result
	case []any:
		result := make([]any, len(value))
		for i, val := range value {
			result[i] = redactSensitiveFields(val)
		}
		return result
	default:
		return value
	}
}
