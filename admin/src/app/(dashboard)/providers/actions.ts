"use server";

import { revalidatePath } from "next/cache";
import { getTranslations } from "next-intl/server";
import { redirect } from "next/navigation";

import { requireSessionToken } from "@/lib/auth";
import {
  AdminApiError,
  adminCreateLLMCredential,
  adminCreateLLMModel,
  adminDeleteLLMCredential,
  adminDeleteLLMModel,
  adminUpdateLLMCredential,
  adminUpdateLLMModel,
  type LLMProvider,
} from "@/lib/backend";

// Errors are reported via a redirect back to the same page with
// ?error=... (not useActionState) — same reasoning as
// personas/actions.ts: this page's dashboard layout always renders a
// second server action (the logout button's form), and plain actions +
// redirect-based error display sidesteps that reliably.

export async function createCredentialAction(formData: FormData) {
  const token = await requireSessionToken();

  const baseUrl = String(formData.get("base_url") ?? "").trim();

  let created;
  try {
    created = await adminCreateLLMCredential(token, {
      provider: String(formData.get("provider") ?? "") as LLMProvider,
      label: String(formData.get("label") ?? "").trim(),
      api_key: String(formData.get("api_key") ?? "").trim(),
      base_url: baseUrl || null,
      is_active: formData.get("is_active") === "on",
    });
  } catch (err) {
    const message =
      err instanceof AdminApiError ? err.message : (await getTranslations("providers"))("errors.credentialCreateFailed");
    redirect(`/providers/new?error=${encodeURIComponent(message)}`);
  }

  revalidatePath("/providers");
  redirect(`/providers/${created.id}`);
}

export async function updateCredentialAction(credentialId: string, formData: FormData) {
  const token = await requireSessionToken();

  const baseUrl = String(formData.get("base_url") ?? "").trim();
  const apiKey = String(formData.get("api_key") ?? "").trim();

  try {
    await adminUpdateLLMCredential(token, credentialId, {
      label: String(formData.get("label") ?? "").trim(),
      api_key: apiKey || null,
      base_url: baseUrl || null,
      is_active: formData.get("is_active") === "on",
    });
  } catch (err) {
    const message =
      err instanceof AdminApiError ? err.message : (await getTranslations("providers"))("errors.credentialUpdateFailed");
    redirect(`/providers/${credentialId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath("/providers");
  revalidatePath(`/providers/${credentialId}`);
  redirect(`/providers/${credentialId}`);
}

export async function deleteCredentialAction(credentialId: string) {
  const token = await requireSessionToken();

  try {
    await adminDeleteLLMCredential(token, credentialId);
  } catch (err) {
    const message =
      err instanceof AdminApiError ? err.message : (await getTranslations("providers"))("errors.credentialDeleteFailed");
    redirect(`/providers/${credentialId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath("/providers");
  redirect("/providers");
}

export async function createModelAction(credentialId: string, formData: FormData) {
  const token = await requireSessionToken();

  try {
    await adminCreateLLMModel(token, {
      credential_id: credentialId,
      model_name: String(formData.get("model_name") ?? "").trim(),
      display_name: String(formData.get("display_name") ?? "").trim(),
      is_default: formData.get("is_default") === "on",
      is_active: formData.get("is_active") === "on",
    });
  } catch (err) {
    const message = err instanceof AdminApiError ? err.message : (await getTranslations("providers"))("errors.modelCreateFailed");
    redirect(`/providers/${credentialId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath(`/providers/${credentialId}`);
  redirect(`/providers/${credentialId}`);
}

export async function updateModelAction(
  credentialId: string,
  modelId: string,
  formData: FormData,
) {
  const token = await requireSessionToken();

  try {
    await adminUpdateLLMModel(token, modelId, {
      display_name: String(formData.get("display_name") ?? "").trim(),
      is_default: formData.get("is_default") === "on",
      is_active: formData.get("is_active") === "on",
    });
  } catch (err) {
    const message = err instanceof AdminApiError ? err.message : (await getTranslations("providers"))("errors.modelUpdateFailed");
    redirect(`/providers/${credentialId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath(`/providers/${credentialId}`);
  redirect(`/providers/${credentialId}`);
}

export async function deleteModelAction(credentialId: string, modelId: string) {
  const token = await requireSessionToken();

  try {
    await adminDeleteLLMModel(token, modelId);
  } catch (err) {
    const message = err instanceof AdminApiError ? err.message : (await getTranslations("providers"))("errors.modelDeleteFailed");
    redirect(`/providers/${credentialId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath(`/providers/${credentialId}`);
  redirect(`/providers/${credentialId}`);
}
