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

// Faz 0 (i18n altyapısı): only tr/en have translated message files —
// see admin/messages/. The rest are valid, selectable locales (and are
// already accepted end-to-end by the backend's Accept-Language
// handling) but fall back to English strings until Faz 1's translation
// pass fills in their messages/<locale>.json.
export const translatedLocales: ReadonlySet<Locale> = new Set(["tr", "en"]);

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
