import { redirect } from "next/navigation";

import { adminMe, type AdminUser } from "@/lib/backend";
import { getSessionToken } from "@/lib/session";

// Guards every page under the (dashboard) route group. A missing or
// no-longer-valid session (expired, revoked, or the backend rejecting
// it for any reason) always sends the admin back to /login — there's
// no partial/stale-admin state to render around.
export async function requireAdmin(): Promise<AdminUser> {
  const token = await getSessionToken();
  if (!token) redirect("/login");

  try {
    return await adminMe(token);
  } catch {
    redirect("/login");
  }
}
