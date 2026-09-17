# perchly-backend

## Local development

```bash
make dev   # boots colima+Docker, starts Postgres/pgvector, runs migrations, starts the API
```

See `.env.example` for all configuration.

## Gmail App Password (for SMTP_PASSWORD)

`SMTP_PASSWORD` must be a **Google App Password**, not your regular Gmail
password — Gmail rejects normal account passwords for SMTP. To generate
one:

1. Go to your [Google Account settings](https://myaccount.google.com/security).
2. Under "How you sign in to Google", turn on **2-Step Verification** if it isn't already on (App Passwords require it).
3. Go to the [App Passwords page](https://myaccount.google.com/apppasswords).
4. Enter a name for the app (e.g. "Perchly backend") and click **Create**.
5. Google shows a 16-character password in groups of four. Paste it quoted in `.env` (`SMTP_PASSWORD="xxxx xxxx xxxx xxxx"`); the backend strips spaces and copy-paste NBSPs at startup.
6. Set `SMTP_USERNAME` to the full Gmail address that generated the password.

If `SMTP_USERNAME`/`SMTP_PASSWORD` are left empty, the backend logs
verification codes to stdout instead of emailing them (see
`internal/email/console.go`) — useful for local development, not for
anything real.

## Request logging & device headers

Every request is logged to the `request_logs` table (see
`internal/requestlog`): method, route, query/route params, body,
status, duration, and the client's device info — asynchronously, so
logging never adds latency to the response.

Clients are expected to send these headers on every request:

| Header             | Example        | Meaning              |
|---------------------|----------------|-----------------------|
| `X-Client-Version`  | `2.3.1`        | App version           |
| `X-Platform`        | `ios`          | `ios`, `android`, etc.|
| `X-OS-Version`      | `26.1`         | OS version             |
| `X-Device-Model`    | `iPhone17,2`   | Device model identifier|

All are optional from the server's point of view — a missing header just
logs as `null`, requests are never rejected for lacking them. See
`internal/requestlog/device_info.go` for the exact header name constants.

Known-sensitive field values in logged bodies/query params (`code`,
`password`, anything containing `token`, `content`) are redacted to
`"[REDACTED]"` by default. Toggle this with `LOG_REDACT_SENSITIVE_FIELDS`
in `.env` — leave it `true` outside of short, deliberate local debugging
sessions, since turning it off means verification codes and chat message
content land in the database in plain text.
