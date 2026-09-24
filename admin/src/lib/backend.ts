import { BACKEND_URL } from "@/lib/config";

// Thrown for any non-2xx response from the backend admin API, carrying
// the {error, code} envelope every perchly-backend endpoint returns
// (see backend/internal/apierror) so callers can show err.message
// directly or switch on err.code.
export class AdminApiError extends Error {
  status: number;
  code?: string;

  constructor(message: string, status: number, code?: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function throwApiError(res: Response): Promise<never> {
  let message = `İstek başarısız oldu (${res.status})`;
  let code: string | undefined;
  try {
    const body = await res.json();
    if (typeof body?.error === "string") message = body.error;
    if (typeof body?.code === "string") code = body.code;
  } catch {
    // Body wasn't JSON — keep the generic message.
  }
  throw new AdminApiError(message, res.status, code);
}

export type AdminUser = {
  id: string;
  email: string;
};

export type AdminSession = {
  session_token: string;
  expires_at: string;
  admin: AdminUser;
};

// Every call here runs server-side only (Server Functions, route
// handlers, server components) — the browser never talks to
// BACKEND_URL directly, so the backend needs no CORS configuration for
// this app. See admin/README or the Phase 0 plan for why.

export async function adminLogin(
  email: string,
  password: string,
): Promise<AdminSession> {
  const res = await fetch(`${BACKEND_URL}/admin/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
    cache: "no-store",
  });
  if (!res.ok) await throwApiError(res);
  return res.json();
}

export async function adminLogout(sessionToken: string): Promise<void> {
  await fetch(`${BACKEND_URL}/admin/auth/logout`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ session_token: sessionToken }),
    cache: "no-store",
  });
}

// authedFetch is the shared plumbing for every admin endpoint below
// that needs a session token — attaches the Bearer header, always
// no-stores (this dashboard shows live operational data, never stale
// cached data), and throws AdminApiError on a non-2xx response so
// callers only have to handle the success shape.
async function authedFetch<T>(
  path: string,
  sessionToken: string,
  init?: RequestInit,
): Promise<T> {
  const res = await fetch(`${BACKEND_URL}${path}`, {
    ...init,
    headers: {
      ...init?.headers,
      Authorization: `Bearer ${sessionToken}`,
    },
    cache: "no-store",
  });
  if (!res.ok) await throwApiError(res);
  return res.json();
}

function jsonBody(body: unknown): RequestInit {
  return {
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  };
}

export async function adminMe(sessionToken: string): Promise<AdminUser> {
  return authedFetch<AdminUser>("/admin/auth/me", sessionToken);
}
