"use client";

import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { type MouseEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import type { Locale } from "@/i18n/locales";
import { localeLabels, locales } from "@/i18n/locales";
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
          const hasTranslation = byLocale.has(locale);
          const isSelected = locale === selected;
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
              {localeLabels[locale]}
              <span
                className={`h-1.5 w-1.5 rounded-full ${
                  hasTranslation
                    ? isSelected
                      ? "bg-primary-foreground"
                      : "bg-primary"
                    : "bg-transparent border border-current opacity-40"
                }`}
              />
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
            defaultValue={current?.name ?? ""}
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
            defaultValue={current?.short_description ?? ""}
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
            defaultValue={current?.tone_description ?? ""}
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
