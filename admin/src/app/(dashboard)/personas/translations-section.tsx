"use client";

import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { type MouseEvent, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import type { Locale } from "@/i18n/locales";
import { localeFlags, localeLabels, locales } from "@/i18n/locales";
import type { PersonaTranslation } from "@/lib/backend";

import { deletePersonaTranslationAction, savePersonaTranslationAction } from "./translation-actions";

// Every locale except Turkish — "tr" is the base content edited by the
// main PersonaForm above this section, not a translation row.
const TRANSLATABLE_LOCALES = locales.filter((locale) => locale !== "tr") as Exclude<
  Locale,
  "tr"
>[];

export function TranslationsSection({
  personaId,
  translations,
}: {
  personaId: string;
  translations: PersonaTranslation[];
}) {
  const t = useTranslations("personas");
  const searchParams = useSearchParams();
  const initialLocale = searchParams.get("t_locale");
  const [selected, setSelected] = useState<(typeof TRANSLATABLE_LOCALES)[number]>(
    (TRANSLATABLE_LOCALES.find((l) => l === initialLocale) ?? TRANSLATABLE_LOCALES[0]),
  );
  const error = searchParams.get("t_error");

  const byLocale = new Map(translations.map((tr) => [tr.locale, tr]));
  const current = byLocale.get(selected);

  // Tracks fields edited but not yet saved, per locale — reset below
  // whenever the saved content for `selected` changes (a successful
  // save) or the admin switches to a locale (a fresh, unedited view).
  // Adjusted during render rather than in an effect, per
  // https://react.dev/learn/you-might-not-need-an-effect#adjusting-some-state-when-a-prop-changes
  const [dirty, setDirty] = useState<Partial<Record<Locale, boolean>>>({});
  const nameRef = useRef<HTMLInputElement>(null);
  const shortRef = useRef<HTMLTextAreaElement>(null);
  const toneRef = useRef<HTMLTextAreaElement>(null);

  const savedSignature = `${selected}:${current?.name ?? ""}:${current?.short_description ?? ""}:${current?.tone_description ?? ""}`;
  const [lastSavedSignature, setLastSavedSignature] = useState(savedSignature);
  if (savedSignature !== lastSavedSignature) {
    setLastSavedSignature(savedSignature);
    setDirty((d) => ({ ...d, [selected]: false }));
  }

  function isIncomplete(locale: (typeof TRANSLATABLE_LOCALES)[number]) {
    const tr = byLocale.get(locale);
    if (!tr) return true;
    return !tr.name.trim() || !tr.short_description.trim() || !tr.tone_description.trim();
  }

  function needsAttention(locale: (typeof TRANSLATABLE_LOCALES)[number]) {
    return Boolean(dirty[locale]) || isIncomplete(locale);
  }

  function handleFieldChange() {
    const isDirty =
      (nameRef.current?.value ?? "").trim() !== (current?.name ?? "").trim() ||
      (shortRef.current?.value ?? "").trim() !== (current?.short_description ?? "").trim() ||
      (toneRef.current?.value ?? "").trim() !== (current?.tone_description ?? "").trim();
    setDirty((d) => ({ ...d, [selected]: isDirty }));
  }

  const boundSave = savePersonaTranslationAction.bind(null, personaId, selected);
  const boundDelete = deletePersonaTranslationAction.bind(null, personaId, selected);

  return (
    <div className="space-y-4 rounded-lg border p-4">
      <div>
        <h2 className="text-lg font-semibold">{t("translations.title")}</h2>
        <p className="text-sm text-muted-foreground">{t("translations.description")}</p>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="flex flex-wrap gap-2">
        {TRANSLATABLE_LOCALES.map((locale) => {
          const isSelected = locale === selected;
          const dotVisible = needsAttention(locale);
          return (
            <button
              key={locale}
              type="button"
              onClick={() => setSelected(locale)}
              className={`inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-sm ${
                isSelected
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-input bg-background"
              }`}
            >
              <span className="text-base leading-none">{localeFlags[locale]}</span>
              {localeLabels[locale]}
              {dotVisible && (
                <span
                  title={t("translations.needsAttention")}
                  className={`h-1.5 w-1.5 shrink-0 rounded-full ${
                    isSelected ? "bg-primary-foreground" : "bg-destructive"
                  }`}
                />
              )}
            </button>
          );
        })}
      </div>

      {!current && (
        <p className="text-sm text-muted-foreground">{t("translations.noneYet")}</p>
      )}

      <form key={selected} action={boundSave} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor={`tr_name_${selected}`}>{t("translations.nameLabel")}</Label>
          <Input
            id={`tr_name_${selected}`}
            name="name"
            ref={nameRef}
            defaultValue={current?.name ?? ""}
            onChange={handleFieldChange}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor={`tr_short_${selected}`}>
            {t("translations.shortDescriptionLabel")}
          </Label>
          <Textarea
            id={`tr_short_${selected}`}
            name="short_description"
            ref={shortRef}
            defaultValue={current?.short_description ?? ""}
            onChange={handleFieldChange}
            required
            rows={2}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor={`tr_tone_${selected}`}>
            {t("translations.toneDescriptionLabel")}
          </Label>
          <Textarea
            id={`tr_tone_${selected}`}
            name="tone_description"
            ref={toneRef}
            defaultValue={current?.tone_description ?? ""}
            onChange={handleFieldChange}
            required
            rows={2}
          />
        </div>

        <div className="flex items-center gap-2">
          <Button type="submit">{t("save")}</Button>
          {current && (
            <Button
              type="submit"
              variant="outline"
              className="text-destructive hover:text-destructive"
              formAction={boundDelete}
              onClick={(e: MouseEvent<HTMLButtonElement>) => {
                if (!confirm(t("translations.confirmReset"))) {
                  e.preventDefault();
                }
              }}
            >
              {t("translations.resetToTurkish")}
            </Button>
          )}
        </div>
      </form>
    </div>
  );
}
