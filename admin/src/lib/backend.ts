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

// --- Dashboard metrics ---

export type AdminMetrics = {
  total_users: number;
  new_users_today: number;
  messages_sent_today: number;
  active_conversations_today: number;
  error_rate_today: number;
};

export async function adminMetrics(sessionToken: string): Promise<AdminMetrics> {
  return authedFetch<AdminMetrics>("/admin/dashboard/metrics", sessionToken);
}

// --- Users ---

export type AppUser = {
  id: string;
  email: string | null;
  email_verified_at: string | null;
  display_name: string | null;
  is_anonymous: boolean;
  created_at: string;
  updated_at: string;
};

export type AppUsersPage = { users: AppUser[]; total: number };

export async function adminListUsers(
  sessionToken: string,
  params: { search?: string; limit?: number; offset?: number } = {},
): Promise<AppUsersPage> {
  const qs = new URLSearchParams();
  if (params.search) qs.set("search", params.search);
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.offset) qs.set("offset", String(params.offset));
  const query = qs.toString();
  return authedFetch<AppUsersPage>(
    `/admin/users${query ? `?${query}` : ""}`,
    sessionToken,
  );
}

export type UserQuota = {
  user_id: string;
  persona_id: string;
  message_count_today: number;
  daily_limit: number;
  last_reset_at: string;
};

export type AppUserDetail = {
  user: AppUser;
  quotas: UserQuota[];
  credits: number;
};

export async function adminGetUser(
  sessionToken: string,
  userId: string,
): Promise<AppUserDetail> {
  return authedFetch<AppUserDetail>(`/admin/users/${userId}`, sessionToken);
}

export async function adminSetQuota(
  sessionToken: string,
  userId: string,
  personaId: string,
  dailyLimit: number,
): Promise<UserQuota> {
  return authedFetch<UserQuota>(
    `/admin/users/${userId}/quota`,
    sessionToken,
    {
      method: "PATCH",
      ...jsonBody({ persona_id: personaId, daily_limit: dailyLimit }),
    },
  );
}

export async function adminSetCredits(
  sessionToken: string,
  userId: string,
  amount: number,
): Promise<{ credits: number }> {
  return authedFetch<{ credits: number }>(
    `/admin/users/${userId}/credits`,
    sessionToken,
    { method: "PATCH", ...jsonBody({ amount }) },
  );
}

// --- Personas ---

export type PersonaTraits = {
  warmth: number;
  humor: number;
  wisdom: number;
  directness: number;
  energy: number;
};

export type Persona = {
  id: string;
  slug: string;
  name: string;
  category: string;
  short_description: string;
  system_prompt: string;
  tone_description: string;
  avatar_url: string | null;
  accent_color: string;
  is_minor_appropriate: boolean;
  is_active: boolean;
  sort_order: number;
  default_traits: PersonaTraits;
  created_at: string;
  updated_at: string;
};

export type PersonaInput = Omit<Persona, "id" | "created_at" | "updated_at">;

export async function adminListPersonas(sessionToken: string): Promise<Persona[]> {
  return authedFetch<Persona[]>("/admin/personas", sessionToken);
}

export async function adminGetPersona(
  sessionToken: string,
  id: string,
): Promise<Persona> {
  return authedFetch<Persona>(`/admin/personas/${id}`, sessionToken);
}

export async function adminCreatePersona(
  sessionToken: string,
  input: PersonaInput,
): Promise<Persona> {
  return authedFetch<Persona>("/admin/personas", sessionToken, {
    method: "POST",
    ...jsonBody(input),
  });
}

export async function adminUpdatePersona(
  sessionToken: string,
  id: string,
  input: PersonaInput,
): Promise<Persona> {
  return authedFetch<Persona>(`/admin/personas/${id}`, sessionToken, {
    method: "PUT",
    ...jsonBody(input),
  });
}
