package apierror

import (
	"context"
	"strconv"
	"strings"
)

// Locale is a stable, machine-readable language identifier — a bare
// IETF primary language subtag (e.g. "en" for "en-US", "zh" for
// "zh-Hans-CN"). Perchly serves one translation per primary subtag, not
// per region/script variant, so ParseAcceptLanguage always reduces to
// this.
type Locale string

const (
	LocaleTR Locale = "tr"
	LocaleEN Locale = "en"
	LocaleDE Locale = "de"
	LocaleAR Locale = "ar"
	LocaleES Locale = "es"
	LocaleFR Locale = "fr"
	LocaleRU Locale = "ru"
	LocaleZH Locale = "zh"
)

// DefaultLocale is what every handler's inline message is already
// written in — Turkish. It's also ParseAcceptLanguage's fallback when
// the header is absent, empty, or names nothing Perchly serves.
const DefaultLocale = LocaleTR

// RTLLocales are the locales that read right-to-left — clients (admin
// dashboard, iOS) use this to set dir="rtl" / layoutDirection; the
// backend itself has no layout, so this lives here only as the one
// shared source of truth for "which locales are RTL", not because
// apierror does anything with direction.
var RTLLocales = map[Locale]bool{
	LocaleAR: true,
}

var supportedLocales = map[Locale]bool{
	LocaleTR: true,
	LocaleEN: true,
	LocaleDE: true,
	LocaleAR: true,
	LocaleES: true,
	LocaleFR: true,
	LocaleRU: true,
	LocaleZH: true,
}

// IsSupported reports whether l is one of Perchly's served locales.
func IsSupported(l Locale) bool {
	return supportedLocales[l]
}

// ParseAcceptLanguage resolves an RFC 9110 Accept-Language header
// ("tr,en;q=0.8,de;q=0.5") to the best-matching supported Locale,
// honoring q-weights and preferring an earlier, higher-weighted tag
// over a later one of equal weight (input order is itself a preference
// signal per the RFC). A region/script subtag ("en-US", "zh-Hans-CN")
// is reduced to its primary language subtag before matching — Perchly
// has no per-region translations. An empty header, a header naming only
// unsupported languages, or a malformed entry (skipped, not fatal — one
// bad tag shouldn't sink the ones around it) all resolve to
// DefaultLocale.
func ParseAcceptLanguage(header string) Locale {
	best := Locale("")
	bestQ := -1.0

	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		tag, q := part, 1.0
		if i := strings.IndexByte(part, ';'); i != -1 {
			tag = strings.TrimSpace(part[:i])
			qPart := strings.TrimSpace(part[i+1:])
			if rest, ok := strings.CutPrefix(qPart, "q="); ok {
				if parsed, err := strconv.ParseFloat(strings.TrimSpace(rest), 64); err == nil {
					q = parsed
				}
			}
		}

		if tag == "*" || tag == "" || q <= 0 {
			continue
		}
		primary, _, _ := strings.Cut(tag, "-")
		locale := Locale(strings.ToLower(primary))
		if !IsSupported(locale) {
			continue
		}

		if q > bestQ {
			bestQ = q
			best = locale
		}
	}

	if best == "" {
		return DefaultLocale
	}
	return best
}

type localeContextKey struct{}

// WithLocale attaches l to ctx — see handler.LocaleMiddleware, the only
// place this is normally called from.
func WithLocale(ctx context.Context, l Locale) context.Context {
	return context.WithValue(ctx, localeContextKey{}, l)
}

// LocaleFromContext returns the locale LocaleMiddleware resolved for
// this request, or DefaultLocale if none was ever attached (e.g. a unit
// test calling a handler directly, or a code path reached before the
// middleware runs).
func LocaleFromContext(ctx context.Context) Locale {
	if l, ok := ctx.Value(localeContextKey{}).(Locale); ok {
		return l
	}
	return DefaultLocale
}
