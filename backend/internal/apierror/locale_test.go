package apierror

import "testing"

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   Locale
	}{
		{"empty header falls back to default", "", DefaultLocale},
		{"simple supported tag", "en", LocaleEN},
		{"region subtag reduces to primary language", "en-US", LocaleEN},
		{"script+region subtag reduces to primary language", "zh-Hans-CN", LocaleZH},
		{"unsupported language falls back to default", "fi", DefaultLocale},
		{"picks the highest q-weight", "en;q=0.5,de;q=0.9", LocaleDE},
		{"equal weights prefer the earlier tag", "es;q=0.8,fr;q=0.8", LocaleES},
		{"unsupported tags are skipped, not fatal", "fi,xx-YY;q=0.9,ru;q=0.7", LocaleRU},
		{"wildcard is ignored", "*,tr;q=0.5", LocaleTR},
		{"malformed q value falls back to 1.0 for that tag", "ar;q=notanumber", LocaleAR},
		{"whitespace around tags is trimmed", " en , de;q=0.5 ", LocaleEN},
		{"zero-or-negative q excludes the tag", "en;q=0,de;q=0.1", LocaleDE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseAcceptLanguage(tt.header); got != tt.want {
				t.Errorf("ParseAcceptLanguage(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestLocaleContext(t *testing.T) {
	ctx := WithLocale(t.Context(), LocaleDE)
	if got := LocaleFromContext(ctx); got != LocaleDE {
		t.Errorf("LocaleFromContext() = %q, want %q", got, LocaleDE)
	}
}

func TestLocaleFromContext_DefaultsWhenAbsent(t *testing.T) {
	if got := LocaleFromContext(t.Context()); got != DefaultLocale {
		t.Errorf("LocaleFromContext() on a bare context = %q, want DefaultLocale %q", got, DefaultLocale)
	}
}

func TestMessage(t *testing.T) {
	const fallback = "geçersiz istek gövdesi"

	tests := []struct {
		name   string
		code   Code
		locale Locale
		want   string
	}{
		{"turkish locale always uses the fallback", CodeInvalidRequestBody, LocaleTR, fallback},
		{"english locale uses the translation", CodeInvalidRequestBody, LocaleEN, "invalid request body"},
		{"a locale with no translation yet falls back", CodeInvalidRequestBody, LocaleDE, fallback},
		{"an unknown code falls back regardless of locale", Code("made_up_code"), LocaleEN, fallback},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Message(tt.code, tt.locale, fallback); got != tt.want {
				t.Errorf("Message(%q, %q, fallback) = %q, want %q", tt.code, tt.locale, got, tt.want)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	for _, l := range []Locale{LocaleTR, LocaleEN, LocaleDE, LocaleAR, LocaleES, LocaleFR, LocaleRU, LocaleZH} {
		if !IsSupported(l) {
			t.Errorf("IsSupported(%q) = false, want true", l)
		}
	}
	if IsSupported(Locale("xx")) {
		t.Error("IsSupported(\"xx\") = true, want false")
	}
}
