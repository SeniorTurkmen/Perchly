-- Append-only request log: which user hit which method/route with which
-- params, from which client. BIGSERIAL (not UUID, unlike the rest of the
-- schema) is deliberate here — a pure log table benefits from a cheap,
-- naturally chronological, sequential key more than from UUIDs.
CREATE TABLE request_logs (
    id             BIGSERIAL PRIMARY KEY,
    -- SET NULL (not CASCADE): deleting a user should never delete their
    -- audit trail, and requests with no authenticated user (personas,
    -- /auth/anonymous itself) always have a null user_id anyway.
    user_id        UUID REFERENCES users(id) ON DELETE SET NULL,
    method         TEXT NOT NULL,
    -- e.g. "/conversations/{id}/messages" — the chi route pattern, not
    -- the concrete path, so requests to the same endpoint group together.
    route_pattern  TEXT NOT NULL,
    path           TEXT NOT NULL,
    query_params   JSONB,
    route_params   JSONB,
    -- Redacted per LOG_REDACT_SENSITIVE_FIELDS at write time — see
    -- internal/requestlog. Null when the request had no body.
    body           JSONB,
    status_code    INTEGER NOT NULL,
    duration_ms    BIGINT NOT NULL,
    -- Device headers the client sends on every request — see
    -- internal/requestlog/device_info.go for the exact header names.
    client_version TEXT,
    platform       TEXT,
    os_version     TEXT,
    device_model   TEXT,
    user_agent     TEXT,
    ip_address     TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX request_logs_user_id_created_at_idx ON request_logs (user_id, created_at DESC);
CREATE INDEX request_logs_route_pattern_created_at_idx ON request_logs (route_pattern, created_at DESC);
CREATE INDEX request_logs_created_at_idx ON request_logs (created_at DESC);
