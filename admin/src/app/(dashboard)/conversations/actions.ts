"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";

import { requireSessionToken } from "@/lib/auth";
import { AdminApiError, adminDeleteMessage } from "@/lib/backend";

export async function deleteMessageAction(conversationId: string, messageId: string) {
  const token = await requireSessionToken();

  try {
    await adminDeleteMessage(token, conversationId, messageId);
  } catch (err) {
    const message = err instanceof AdminApiError ? err.message : "Mesaj silinemedi.";
    redirect(`/conversations/${conversationId}?error=${encodeURIComponent(message)}`);
  }

  revalidatePath(`/conversations/${conversationId}`);
  redirect(`/conversations/${conversationId}`);
}
