-- Internal admin-dashboard accounts. Deliberately separate from `users`
-- (anonymous/email-OTP app end users) so admin auth never mixes with the
-- public auth flow or its data.
CREATE TABLE admin_users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          TEXT NOT NULL UNIQUE,
    password_hash  TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Opaque, hashed admin session tokens — same scheme as refresh_tokens
-- (see internal/auth/refresh_token.go), scoped to admin_users instead of
-- users so a leaked table alone can't be replayed as a token.
CREATE TABLE admin_sessions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id  UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    token_hash     TEXT NOT NULL UNIQUE,
    expires_at     TIMESTAMPTZ NOT NULL,
    revoked_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX admin_sessions_admin_user_id_idx ON admin_sessions (admin_user_id);

-- Append-only trail of every mutating admin action, so changes to
-- production user data (quota overrides, persona edits, credit grants,
-- moderation) can always be traced back to who did what and when.
CREATE TABLE admin_audit_log (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id  UUID NOT NULL REFERENCES admin_users(id),
    action         TEXT NOT NULL,
    target_type    TEXT NOT NULL,
    target_id      TEXT,
    detail         JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX admin_audit_log_created_at_idx ON admin_audit_log (created_at DESC);
