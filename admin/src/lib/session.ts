import { cookies } from "next/headers";

// First-party cookie on the admin app's own domain — never the
// backend's. It holds the opaque admin session token issued by POST
// /admin/auth/login; every server-side call to the backend reads it
// back and forwards it as a Bearer token (see lib/backend.ts).
const SESSION_COOKIE = "perchly_admin_session";

export async function getSessionToken(): Promise<string | undefined> {
  const store = await cookies();
  return store.get(SESSION_COOKIE)?.value;
}

// Must be called from a Server Function or Route Handler — cookies()
// can't be mutated while rendering a Server Component.
export async function setSessionCookie(
  token: string,
  expiresAt: string,
): Promise<void> {
  const store = await cookies();
  store.set(SESSION_COOKIE, token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    expires: new Date(expiresAt),
  });
}

export async function clearSessionCookie(): Promise<void> {
  const store = await cookies();
  store.delete(SESSION_COOKIE);
}
