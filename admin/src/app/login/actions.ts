"use server";

import { redirect } from "next/navigation";

import { AdminApiError, adminLogin } from "@/lib/backend";
import { setSessionCookie } from "@/lib/session";

export type LoginState = { error?: string };

export async function loginAction(
  _prevState: LoginState,
  formData: FormData,
): Promise<LoginState> {
  const email = String(formData.get("email") ?? "").trim();
  const password = String(formData.get("password") ?? "");

  if (!email || !password) {
    return { error: "E-posta ve şifre gerekli." };
  }

  try {
    const session = await adminLogin(email, password);
    await setSessionCookie(session.session_token, session.expires_at);
  } catch (err) {
    if (err instanceof AdminApiError) {
      return { error: err.message };
    }
    return { error: "Giriş yapılamadı, backend'e ulaşılamıyor." };
  }

  redirect("/");
}
