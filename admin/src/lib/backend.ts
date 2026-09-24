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

// --- Conversations & moderation ---

export type Message = {
  id: string;
  conversation_id: string;
  role: "user" | "assistant";
  content: string;
  reaction_emoji?: string | null;
  created_at: string;
};

export type ConversationPreview = {
  id: string;
  user_id: string;
  persona_id: string;
  created_at: string;
  updated_at: string;
  persona: Persona;
  last_message: Message | null;
};

export type ConversationsPage = {
  conversations: ConversationPreview[];
  total: number;
};

export async function adminListConversations(
  sessionToken: string,
  params: {
    userId?: string;
    personaId?: string;
    search?: string;
    limit?: number;
    offset?: number;
  } = {},
): Promise<ConversationsPage> {
  const qs = new URLSearchParams();
  if (params.userId) qs.set("user_id", params.userId);
  if (params.personaId) qs.set("persona_id", params.personaId);
  if (params.search) qs.set("search", params.search);
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.offset) qs.set("offset", String(params.offset));
  const query = qs.toString();
  return authedFetch<ConversationsPage>(
    `/admin/conversations${query ? `?${query}` : ""}`,
    sessionToken,
  );
}

export type Conversation = {
  id: string;
  user_id: string;
  persona_id: string;
  created_at: string;
  updated_at: string;
};

export type ConversationDetail = {
  conversation: Conversation;
  messages: Message[];
};

export async function adminGetConversation(
  sessionToken: string,
  id: string,
): Promise<ConversationDetail> {
  return authedFetch<ConversationDetail>(`/admin/conversations/${id}`, sessionToken);
}

export async function adminDeleteMessage(
  sessionToken: string,
  conversationId: string,
  messageId: string,
): Promise<void> {
  await authedFetch<{ success: boolean }>(
    `/admin/conversations/${conversationId}/messages/${messageId}`,
    sessionToken,
    { method: "DELETE" },
  );
}

// --- Request logs (app users) ---

export type RequestLog = {
  id: number;
  user_id: string | null;
  user_email: string | null;
  user_display_name: string | null;
  method: string;
  route_pattern: string;
  path: string;
  query_params: Record<string, unknown> | null;
  route_params: Record<string, unknown> | null;
  body: unknown;
  status_code: number;
  duration_ms: number;
  request_headers: Record<string, unknown> | null;
  response_headers: Record<string, unknown> | null;
  response_body: string | null;
  client_version: string | null;
  platform: string | null;
  os_version: string | null;
  device_model: string | null;
  user_agent: string;
  ip_address: string;
  created_at: string;
};

export type RequestLogsPage = { logs: RequestLog[]; total: number };

export async function adminListLogs(
  sessionToken: string,
  params: {
    search?: string;
    statusMin?: number;
    userId?: string;
    limit?: number;
    offset?: number;
  } = {},
): Promise<RequestLogsPage> {
  const qs = new URLSearchParams();
  if (params.search) qs.set("search", params.search);
  if (params.statusMin) qs.set("status_min", String(params.statusMin));
  if (params.userId) qs.set("user_id", params.userId);
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.offset) qs.set("offset", String(params.offset));
  const query = qs.toString();
  return authedFetch<RequestLogsPage>(
    `/admin/logs${query ? `?${query}` : ""}`,
    sessionToken,
  );
}

// --- Admin activity log (admins) ---

export type AdminActivityEntry = {
  id: string;
  admin_user_id: string;
  admin_email: string;
  action: string;
  target_type: string;
  target_id: string | null;
  detail: Record<string, unknown> | null;
  created_at: string;
};

export type AdminActivityPage = { entries: AdminActivityEntry[]; total: number };

export async function adminListActivity(
  sessionToken: string,
  params: { adminUserId?: string; limit?: number; offset?: number } = {},
): Promise<AdminActivityPage> {
  const qs = new URLSearchParams();
  if (params.adminUserId) qs.set("admin_user_id", params.adminUserId);
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.offset) qs.set("offset", String(params.offset));
  const query = qs.toString();
  return authedFetch<AdminActivityPage>(
    `/admin/activity${query ? `?${query}` : ""}`,
    sessionToken,
  );
}

// --- Onboarding insights ---

export type PersonaSelectionCount = {
  persona_id: string;
  persona_name: string;
  count: number;
};

export type AdminOnboardingInsights = {
  total_users: number;
  completed_onboarding: number;
  minor_count: number;
  notifications_granted_count: number;
  preferred_name_set_count: number;
  skip_hitap_count: number;
  age_range_counts: Record<string, number>;
  mood_preference_counts: Record<string, number>;
  top_selected_personas: PersonaSelectionCount[];
};

export async function adminGetOnboardingInsights(
  sessionToken: string,
): Promise<AdminOnboardingInsights> {
  return authedFetch<AdminOnboardingInsights>("/admin/onboarding/insights", sessionToken);
}

// --- LLM provider credentials & models ---

export type LLMProvider = "openai" | "anthropic" | "gemini" | "huggingface";

export type LLMCredential = {
  id: string;
  provider: LLMProvider;
  label: string;
  api_key_preview: string;
  base_url: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type CreateLLMCredentialInput = {
  provider: LLMProvider;
  label: string;
  api_key: string;
  base_url: string | null;
  is_active: boolean;
};

export type UpdateLLMCredentialInput = {
  label: string;
  api_key: string | null;
  base_url: string | null;
  is_active: boolean;
};

export async function adminListLLMCredentials(
  sessionToken: string,
): Promise<LLMCredential[]> {
  return authedFetch<LLMCredential[]>("/admin/llm/credentials", sessionToken);
}

export async function adminGetLLMCredential(
  sessionToken: string,
  id: string,
): Promise<LLMCredential> {
  return authedFetch<LLMCredential>(`/admin/llm/credentials/${id}`, sessionToken);
}

export async function adminCreateLLMCredential(
  sessionToken: string,
  input: CreateLLMCredentialInput,
): Promise<LLMCredential> {
  return authedFetch<LLMCredential>("/admin/llm/credentials", sessionToken, {
    method: "POST",
    ...jsonBody(input),
  });
}

export async function adminUpdateLLMCredential(
  sessionToken: string,
  id: string,
  input: UpdateLLMCredentialInput,
): Promise<LLMCredential> {
  return authedFetch<LLMCredential>(`/admin/llm/credentials/${id}`, sessionToken, {
    method: "PUT",
    ...jsonBody(input),
  });
}

export async function adminDeleteLLMCredential(
  sessionToken: string,
  id: string,
): Promise<void> {
  await authedFetch<{ success: boolean }>(`/admin/llm/credentials/${id}`, sessionToken, {
    method: "DELETE",
  });
}

export type LLMModel = {
  id: string;
  credential_id: string;
  model_name: string;
  display_name: string;
  is_default: boolean;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type CreateLLMModelInput = {
  credential_id: string;
  model_name: string;
  display_name: string;
  is_default: boolean;
  is_active: boolean;
};

export type UpdateLLMModelInput = {
  display_name: string;
  is_default: boolean;
  is_active: boolean;
};

export async function adminListLLMModels(
  sessionToken: string,
  credentialId?: string,
): Promise<LLMModel[]> {
  const query = credentialId ? `?credential_id=${credentialId}` : "";
  return authedFetch<LLMModel[]>(`/admin/llm/models${query}`, sessionToken);
}

export async function adminCreateLLMModel(
  sessionToken: string,
  input: CreateLLMModelInput,
): Promise<LLMModel> {
  return authedFetch<LLMModel>("/admin/llm/models", sessionToken, {
    method: "POST",
    ...jsonBody(input),
  });
}

export async function adminUpdateLLMModel(
  sessionToken: string,
  id: string,
  input: UpdateLLMModelInput,
): Promise<LLMModel> {
  return authedFetch<LLMModel>(`/admin/llm/models/${id}`, sessionToken, {
    method: "PUT",
    ...jsonBody(input),
  });
}

export async function adminDeleteLLMModel(
  sessionToken: string,
  id: string,
): Promise<void> {
  await authedFetch<{ success: boolean }>(`/admin/llm/models/${id}`, sessionToken, {
    method: "DELETE",
  });
}

// --- Backend health (public, no session token — same as the iOS app's
// own health check) ---

export type BackendHealth = { status: string; database: string };

export async function checkBackendHealth(): Promise<BackendHealth> {
  const res = await fetch(`${BACKEND_URL}/health`, { cache: "no-store" });
  if (!res.ok) await throwApiError(res);
  return res.json();
}
