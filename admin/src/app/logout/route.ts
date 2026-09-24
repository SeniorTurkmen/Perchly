import { type NextRequest, NextResponse } from "next/server";

import { clearSessionCookie, getSessionToken } from "@/lib/session";
import { adminLogout } from "@/lib/backend";

// A plain Route Handler, not a Server Function — deliberately, so the
// logout form's action is resolved by ordinary URL routing instead of
// the Server Function action-id dispatch mechanism. Every dashboard
// page also renders a page-level Server Function (saving a quota,
// creating a persona, ...); this Next.js version doesn't reliably
// resolve which bound action to invoke when the logout button's form
// and a page's own form are both Server Functions present at once —
// see the (dashboard) layout for where this is mounted.
export async function POST(request: NextRequest) {
  const token = await getSessionToken();
  if (token) await adminLogout(token);
  await clearSessionCookie();
  return NextResponse.redirect(new URL("/login", request.url));
}
