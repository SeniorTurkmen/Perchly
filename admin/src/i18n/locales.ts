// Mirrors backend/internal/apierror/locale.go's Locale type — kept in
// sync by hand (small, stable list) rather than shared code, since
// there's no existing cross-language codegen in this repo and the set
// changes rarely.
export const locales = ["tr", "en", "de", "ar", "es", "fr", "ru", "zh"] as const;

export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = "tr";

export const rtlLocales: ReadonlySet<Locale> = new Set(["ar"]);

export function isLocale(value: string): value is Locale {
  return (locales as readonly string[]).includes(value);
}

// All 8 locales now have translated message files — see admin/messages/.
export const translatedLocales: ReadonlySet<Locale> = new Set(locales);

export const localeLabels: Record<Locale, string> = {
  tr: "Türkçe",
  en: "English",
  de: "Deutsch",
  ar: "العربية",
  es: "Español",
  fr: "Français",
  ru: "Русский",
  zh: "中文",
};

// Representative country flag per locale — a language doesn't map to
// exactly one country (ar and zh especially), so these are a visual
// shorthand for the switcher, not a claim about where a language is
// spoken.
export const localeFlags: Record<Locale, string> = {
  tr: "🇹🇷",
  en: "🇺🇸",
  de: "🇩🇪",
  ar: "🇸🇦",
  es: "🇪🇸",
  fr: "🇫🇷",
  ru: "🇷🇺",
  zh: "🇨🇳",
};
