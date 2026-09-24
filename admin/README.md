# perchly-admin

Internal dashboard for operating Perchly (users, personas, quotas,
moderation, logs) and tracking engineering work — see the root
`README.md`'s "Repository layout" for how this fits alongside
`backend/` and `ios/`.

Next.js (App Router, TypeScript, Tailwind, shadcn/ui). Talks to
`backend/`'s `/admin/*` API entirely server-side (server components,
Server Functions, route handlers) — the browser never calls the
backend directly, so no CORS setup is needed there.

## Local development

```bash
cp .env.example .env.local   # defaults to a locally running backend on :8080
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). You'll be
redirected to `/login`; sign in with an account created via the
backend's `make admin-create` (see `backend/README.md`).

## Auth model

Login (`POST /admin/auth/login` on the backend) returns an opaque
session token, stored here as a first-party, `HttpOnly` cookie
(`perchly_admin_session` — see `src/lib/session.ts`). Every subsequent
call to the backend forwards it as `Authorization: Bearer <token>`
(`src/lib/backend.ts`), the same way the iOS app sends its access
token. `src/proxy.ts` does a cheap cookie-presence redirect; the real
check — is the token still valid — happens once per navigation via
`GET /admin/auth/me` in `src/app/(dashboard)/layout.tsx`
(`src/lib/auth.ts#requireAdmin`).

## Status

Phase 0: auth + navigation shell. Users, Personas, Conversations,
Quotas, Logs, and Engineering are placeholder pages, filled in over the
next phases (see the project's dashboard plan).
