CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- IANA timezone name (e.g. 'Europe/Istanbul'), used by the quota
    -- reset job to know when "today" rolls over for this user.
    timezone    TEXT NOT NULL DEFAULT 'UTC',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Conversations didn't belong to anyone until now; every conversation
-- must have an owner for quota enforcement and access control to work.
ALTER TABLE conversations
    ADD COLUMN user_id UUID NOT NULL REFERENCES users(id);

CREATE INDEX conversations_user_id_idx ON conversations (user_id);

CREATE TABLE user_quotas (
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    persona_id           UUID NOT NULL REFERENCES personas(id) ON DELETE CASCADE,
    message_count_today  INTEGER NOT NULL DEFAULT 0,
    -- Per user+persona so a future admin dashboard can grant an
    -- individual override just by updating this column; defaults to the
    -- app-wide default (DAILY_MESSAGE_LIMIT) when a row is first created.
    daily_limit          INTEGER NOT NULL DEFAULT 20,
    last_reset_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, persona_id)
);

CREATE TABLE credit_balances (
    user_id        UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    credit_amount  INTEGER NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
