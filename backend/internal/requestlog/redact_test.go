package requestlog

import (
	"encoding/json"
	"testing"
)

func TestRedactSensitiveFields(t *testing.T) {
	input := `{
		"email": "user@example.com",
		"code": "482913",
		"content": "kişisel bir mesaj",
		"password": "s3cr3t",
		"access_token": "abc.def.ghi",
		"persona_id": "8a870ee3-e627-4871-8764-f310cb3251c4",
		"nested": {"refresh_token": "xyz", "display_name": "Ada"}
	}`

	var decoded any
	if err := json.Unmarshal([]byte(input), &decoded); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	redacted, ok := redactSensitiveFields(decoded).(map[string]any)
	if !ok {
		t.Fatalf("expected redactSensitiveFields to return a map")
	}

	wantRedacted := []string{"code", "content", "password", "access_token"}
	for _, field := range wantRedacted {
		if redacted[field] != "[REDACTED]" {
			t.Errorf("field %q = %v, want [REDACTED]", field, redacted[field])
		}
	}

	wantUntouched := map[string]string{
		"email":      "user@example.com",
		"persona_id": "8a870ee3-e627-4871-8764-f310cb3251c4",
	}
	for field, want := range wantUntouched {
		if redacted[field] != want {
			t.Errorf("field %q = %v, want %v (untouched)", field, redacted[field], want)
		}
	}

	nested, ok := redacted["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested to remain a map")
	}
	if nested["refresh_token"] != "[REDACTED]" {
		t.Errorf("nested.refresh_token = %v, want [REDACTED]", nested["refresh_token"])
	}
	if nested["display_name"] != "Ada" {
		t.Errorf("nested.display_name = %v, want untouched \"Ada\"", nested["display_name"])
	}
}

func TestRedactSensitiveFields_NonObjectValuesPassThrough(t *testing.T) {
	if got := redactSensitiveFields("plain string"); got != "plain string" {
		t.Errorf("scalar input = %v, want passthrough", got)
	}
	if got := redactSensitiveFields(nil); got != nil {
		t.Errorf("nil input = %v, want nil", got)
	}
}
