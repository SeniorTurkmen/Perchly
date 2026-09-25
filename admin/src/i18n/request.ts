import { getRequestConfig } from "next-intl/server";

import { defaultLocale, translatedLocales } from "@/i18n/locales";
import { getLocale } from "@/lib/locale";

// No URL-based locale routing (no /en/, /ar/ prefixes) — this is an
// internal admin tool, not a public/SEO-facing site, so the simpler
// cookie/header-resolved locale (see lib/locale.ts) is enough; every
// server component reads the same resolved locale via next-intl's
// getTranslations()/getLocale() instead of a route segment.
export default getRequestConfig(async () => {
  const locale = await getLocale();

  // All 8 locales have message files now; this fallback to Turkish
  // (defaultLocale, the always-complete source language) only matters
  // if a future locale is added to `locales` before its
  // messages/<locale>.json exists — the same fallback target the
  // backend's apierror.Message uses, so both clients degrade the same,
  // predictable way for an untranslated locale.
  const messagesLocale = translatedLocales.has(locale) ? locale : defaultLocale;

  return {
    locale,
    messages: (await import(`../../messages/${messagesLocale}.json`)).default,
  };
});
