-- Reworks the minimal "anonymous user with a timezone" model into full
-- auth: anonymous users identified by device_id, upgradeable in place to
-- an email-verified account. `timezone` is kept as-is — the quota reset
-- system already depends on it and it's unrelated to this rework.
ALTER TABLE users
    ADD COLUMN email              TEXT UNIQUE,
    ADD COLUMN email_verified_at  TIMESTAMPTZ,
    ADD COLUMN display_name       TEXT,
    ADD COLUMN is_anonymous       BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN device_id          TEXT UNIQUE,
    ADD COLUMN updated_at         TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TABLE email_verification_codes (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          TEXT NOT NULL,
    code           TEXT NOT NULL,
    expires_at     TIMESTAMPTZ NOT NULL,
    attempt_count  INTEGER NOT NULL DEFAULT 0,
    used_at        TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Powers both the request-code rate limiter (count rows for an email
-- created in the last minute/hour) and verify-code's "find the latest
-- unused code for this email" lookup.
CREATE INDEX email_verification_codes_email_created_at_idx
    ON email_verification_codes (email, created_at DESC);

CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Powers "revoke every refresh token for this user" on reuse-detection
-- (see AuthService.RefreshSession).
CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens (user_id);
