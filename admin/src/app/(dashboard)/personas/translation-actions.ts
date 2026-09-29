"use server";

import { revalidatePath } from "next/cache";
import { getTranslations } from "next-intl/server";
import { redirect } from "next/navigation";

import { requireSessionToken } from "@/lib/auth";
import {
  AdminApiError,
  adminDeletePersonaTranslation,
  adminUpsertPersonaTranslation,
} from "@/lib/backend";

// Separate query params from the main form's own ?error= (see
// updatePersonaAction in actions.ts) so a translation-save failure
// doesn't get misread as a base-persona-form failure, or vice versa.
// ?t_locale is echoed back either way so the page reopens on the same
// locale tab the admin was editing.
function redirectTo(personaId: string, locale: string, error?: string) {
  const qs = new URLSearchParams({ t_locale: locale });
  if (error) qs.set("t_error", error);
  redirect(`/personas/${personaId}?${qs}`);
}

export async function savePersonaTranslationAction(
  personaId: string,
  locale: string,
  formData: FormData,
) {
  const token = await requireSessionToken();
  const input = {
    name: String(formData.get("name") ?? "").trim(),
    short_description: String(formData.get("short_description") ?? "").trim(),
    tone_description: String(formData.get("tone_description") ?? "").trim(),
  };

  try {
    await adminUpsertPersonaTranslation(token, personaId, locale, input);
  } catch (err) {
    const t = await getTranslations("personas");
    const message =
      err instanceof AdminApiError ? err.message : t("translations.saveFailed");
    redirectTo(personaId, locale, message);
  }

  revalidatePath(`/personas/${personaId}`);
  redirectTo(personaId, locale);
}

export async function deletePersonaTranslationAction(
  personaId: string,
  locale: string,
) {
  const token = await requireSessionToken();

  try {
    await adminDeletePersonaTranslation(token, personaId, locale);
  } catch (err) {
    const t = await getTranslations("personas");
    const message =
      err instanceof AdminApiError ? err.message : t("translations.deleteFailed");
    redirectTo(personaId, locale, message);
  }

  revalidatePath(`/personas/${personaId}`);
  redirectTo(personaId, locale);
}
