"use client";

import { useLocale, useTranslations } from "next-intl";
import { useTransition } from "react";

import { setLocaleAction } from "@/app/locale-actions";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { localeLabels, locales, type Locale } from "@/i18n/locales";

export function LanguageSwitcher() {
  const t = useTranslations("common");
  const locale = useLocale();
  const [isPending, startTransition] = useTransition();

  return (
    <Select
      value={locale}
      disabled={isPending}
      onValueChange={(value) => {
        startTransition(() => {
          void setLocaleAction(value as Locale);
        });
      }}
    >
      <SelectTrigger aria-label={t("language")} size="sm">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {locales.map((l) => (
          <SelectItem key={l} value={l}>
            {localeLabels[l]}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
