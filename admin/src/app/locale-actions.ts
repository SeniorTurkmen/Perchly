"use server";

import { revalidatePath } from "next/cache";

import type { Locale } from "@/i18n/locales";
import { setLocaleCookie } from "@/lib/locale";

export async function setLocaleAction(locale: Locale) {
  await setLocaleCookie(locale);
  // Every server component reads the locale fresh on next render
  // (see i18n/request.ts) — revalidate the whole tree so the switch is
  // visible immediately instead of only after the next real navigation.
  revalidatePath("/", "layout");
}
