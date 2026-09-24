import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const SESSION_COOKIE = "perchly_admin_session";

// Cheap cookie-presence redirect only — it can't tell an expired or
// revoked token from a valid one without a backend round trip on every
// request. The real check is GET /admin/auth/me, done once per
// navigation in the (dashboard) layout (see lib/auth.ts#requireAdmin).
// This just keeps a logged-out visitor off dashboard routes, and a
// logged-in one off /login, without that extra request.
export function proxy(request: NextRequest) {
  const hasSession = request.cookies.has(SESSION_COOKIE);
  const isLoginPage = request.nextUrl.pathname.startsWith("/login");

  if (!hasSession && !isLoginPage) {
    return NextResponse.redirect(new URL("/login", request.url));
  }
  if (hasSession && isLoginPage) {
    return NextResponse.redirect(new URL("/", request.url));
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
