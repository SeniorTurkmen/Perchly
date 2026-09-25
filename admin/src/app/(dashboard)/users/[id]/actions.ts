"use server";

import { revalidatePath } from "next/cache";
import { getTranslations } from "next-intl/server";
import { redirect } from "next/navigation";

import { requireSessionToken } from "@/lib/auth";
import { AdminApiError, adminSetCredits, adminSetQuota } from "@/lib/backend";

export async function setQuotaAction(userId: string, formData: FormData) {
  const token = await requireSessionToken();
  const personaId = String(formData.get("persona_id") ?? "");
  const dailyLimit = Number(formData.get("daily_limit"));

  try {
    await adminSetQuota(token, userId, personaId, dailyLimit);
  } catch (err) {
    const t = await getTranslations("users");
    const message = err instanceof AdminApiError ? err.message : t("errors.quotaUpdateFailed");
    redirect(`/users/${userId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath(`/users/${userId}`);
}

export async function setCreditsAction(userId: string, formData: FormData) {
  const token = await requireSessionToken();
  const amount = Number(formData.get("amount"));

  try {
    await adminSetCredits(token, userId, amount);
  } catch (err) {
    const t = await getTranslations("users");
    const message = err instanceof AdminApiError ? err.message : t("errors.creditsUpdateFailed");
    redirect(`/users/${userId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath(`/users/${userId}`);
}
