import { cookies, headers } from "next/headers";

import { defaultLocale, isLocale, type Locale } from "@/i18n/locales";

// First-party cookie on the admin app's own domain, same pattern as
// session.ts's SESSION_COOKIE — an explicit choice always wins over
// anything inferred from the browser.
const LOCALE_COOKIE = "perchly_admin_locale";

// getLocale resolves the active locale for this request: the explicit
// cookie if one was ever set (via the language switcher), otherwise the
// browser's own Accept-Language header (so a first-time visitor sees
// their own language, not a hardcoded default), otherwise
// defaultLocale.
export async function getLocale(): Promise<Locale> {
  const store = await cookies();
  const fromCookie = store.get(LOCALE_COOKIE)?.value;
  if (fromCookie && isLocale(fromCookie)) {
    return fromCookie;
  }

  const acceptLanguage = (await headers()).get("accept-language");
  if (acceptLanguage) {
    const fromHeader = parseAcceptLanguage(acceptLanguage);
    if (fromHeader) return fromHeader;
  }

  return defaultLocale;
}

// Must be called from a Server Function or Route Handler — cookies()
// can't be mutated while rendering a Server Component (see
// session.ts's setSessionCookie for the same constraint).
export async function setLocaleCookie(locale: Locale): Promise<void> {
  const store = await cookies();
  store.set(LOCALE_COOKIE, locale, {
    // Not httpOnly: this is a display preference, not a secret — a
    // client-side language switcher (no full page reload) would need
    // to read it too, if one's ever added.
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 60 * 24 * 365,
  });
}

// parseAcceptLanguage mirrors backend/internal/apierror/locale.go's
// ParseAcceptLanguage (RFC 9110 q-weight negotiation, reduced to the
// primary language subtag) — kept as a small local reimplementation
// rather than a shared package, matching src/i18n/locales.ts's
// same reasoning.
function parseAcceptLanguage(header: string): Locale | null {
  let best: Locale | null = null;
  let bestQ = -1;

  for (const part of header.split(",")) {
    const trimmed = part.trim();
    if (!trimmed) continue;

    const [tagPart, ...params] = trimmed.split(";").map((s) => s.trim());
    const tag = tagPart;
    if (!tag || tag === "*") continue;

    let q = 1;
    for (const param of params) {
      const [key, value] = param.split("=").map((s) => s.trim());
      if (key === "q") {
        const parsed = Number.parseFloat(value);
        if (!Number.isNaN(parsed)) q = parsed;
      }
    }
    if (q <= 0) continue;

    const primary = tag.split("-")[0]?.toLowerCase() ?? "";
    if (!isLocale(primary)) continue;

    if (q > bestQ) {
      bestQ = q;
      best = primary;
    }
  }

  return best;
}
