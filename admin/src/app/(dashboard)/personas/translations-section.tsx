"use client";

import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { type MouseEvent, useState } from "react";

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

type TranslatableLocale = (typeof TRANSLATABLE_LOCALES)[number];

type Draft = { name: string; short_description: string; tone_description: string };

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
  const [selected, setSelected] = useState<TranslatableLocale>(
    (TRANSLATABLE_LOCALES.find((l) => l === initialLocale) ?? TRANSLATABLE_LOCALES[0]),
  );
  const error = searchParams.get("t_error");

  const byLocale = new Map(translations.map((tr) => [tr.locale, tr]));
  const current = byLocale.get(selected);

  // Per-locale edit buffers that survive switching tabs — switching to
  // another locale and back used to remount the form with defaultValue
  // from the *saved* row, silently discarding whatever was typed but
  // not yet saved. Only onChange writes here now; a locale with no
  // buffer yet just falls back to its saved content.
  const [drafts, setDrafts] = useState<Partial<Record<Locale, Draft>>>({});

  function savedValues(locale: TranslatableLocale): Draft {
    const tr = byLocale.get(locale);
    return {
      name: tr?.name ?? "",
      short_description: tr?.short_description ?? "",
      tone_description: tr?.tone_description ?? "",
    };
  }

  function fieldValues(locale: TranslatableLocale): Draft {
    return drafts[locale] ?? savedValues(locale);
  }

  const values = fieldValues(selected);

  function updateField(field: keyof Draft, value: string) {
    setDrafts((d) => ({ ...d, [selected]: { ...fieldValues(selected), [field]: value } }));
  }

  function isIncomplete(locale: TranslatableLocale) {
    const tr = byLocale.get(locale);
    if (!tr) return true;
    return !tr.name.trim() || !tr.short_description.trim() || !tr.tone_description.trim();
  }

  function isDirty(locale: TranslatableLocale) {
    const draft = drafts[locale];
    if (!draft) return false;
    const saved = savedValues(locale);
    return (
      draft.name.trim() !== saved.name.trim() ||
      draft.short_description.trim() !== saved.short_description.trim() ||
      draft.tone_description.trim() !== saved.tone_description.trim()
    );
  }

  function needsAttention(locale: TranslatableLocale) {
    return isDirty(locale) || isIncomplete(locale);
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
          // Only shown on the other tabs — while a tab is selected its
          // fields are right there on screen, so the reminder dot would
          // be redundant (and, in the selected pill's own color, easy to
          // mistake for a rendering glitch).
          const dotVisible = !isSelected && needsAttention(locale);
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
                  className="h-1.5 w-1.5 shrink-0 rounded-full bg-destructive"
                />
              )}
            </button>
          );
        })}
      </div>

      {!current && (
        <p className="text-sm text-muted-foreground">{t("translations.noneYet")}</p>
      )}

      <form action={boundSave} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor={`tr_name_${selected}`}>{t("translations.nameLabel")}</Label>
          <Input
            id={`tr_name_${selected}`}
            name="name"
            value={values.name}
            onChange={(e) => updateField("name", e.target.value)}
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
            value={values.short_description}
            onChange={(e) => updateField("short_description", e.target.value)}
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
            value={values.tone_description}
            onChange={(e) => updateField("tone_description", e.target.value)}
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
