"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requireSessionToken } from "@/lib/auth";
import {
  AdminApiError,
  adminCreatePersona,
  adminUpdatePersona,
  type PersonaInput,
} from "@/lib/backend";

function readPersonaInput(formData: FormData): PersonaInput {
  const trait = (name: string) => Number(formData.get(name) ?? 50);
  const avatarUrl = String(formData.get("avatar_url") ?? "").trim();

  return {
    slug: String(formData.get("slug") ?? "").trim(),
    name: String(formData.get("name") ?? "").trim(),
    category: String(formData.get("category") ?? "").trim(),
    short_description: String(formData.get("short_description") ?? "").trim(),
    system_prompt: String(formData.get("system_prompt") ?? "").trim(),
    tone_description: String(formData.get("tone_description") ?? "").trim(),
    avatar_url: avatarUrl || null,
    accent_color: String(formData.get("accent_color") ?? "").trim(),
    is_minor_appropriate: formData.get("is_minor_appropriate") === "on",
    is_active: formData.get("is_active") === "on",
    sort_order: Number(formData.get("sort_order") ?? 0),
    default_traits: {
      warmth: trait("warmth"),
      humor: trait("humor"),
      wisdom: trait("wisdom"),
      directness: trait("directness"),
      energy: trait("energy"),
    },
  };
}

// Errors are reported via a redirect back to the same page with
// ?error=... (not useActionState) — this page's dashboard layout
// always renders a second server action (the logout button's form),
// and this Next.js version doesn't reliably resolve which bound action
// to invoke when a useActionState-wrapped action shares a page with
// another one. Plain actions + redirect-based error display (the same
// pattern users/[id]/actions.ts already uses successfully) sidesteps it.
export async function createPersonaAction(formData: FormData) {
  const token = await requireSessionToken();
  const input = readPersonaInput(formData);

  let created;
  try {
    created = await adminCreatePersona(token, input);
  } catch (err) {
    const message =
      err instanceof AdminApiError ? err.message : "Persona oluşturulamadı.";
    redirect(`/personas/new?error=${encodeURIComponent(message)}`);
  }

  revalidatePath("/personas");
  redirect(`/personas/${created.id}`);
}

export async function updatePersonaAction(personaId: string, formData: FormData) {
  const token = await requireSessionToken();
  const input = readPersonaInput(formData);

  try {
    await adminUpdatePersona(token, personaId, input);
  } catch (err) {
    const message =
      err instanceof AdminApiError ? err.message : "Persona güncellenemedi.";
    redirect(`/personas/${personaId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath("/personas");
  redirect(`/personas/${personaId}`);
}
