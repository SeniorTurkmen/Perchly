-- Encrypted LLM provider credentials, managed from the admin dashboard.
-- The API key itself is never stored in plaintext: api_key_ciphertext /
-- api_key_nonce are AES-256-GCM output (see internal/crypto), encrypted
-- and decrypted only server-side using LLM_TOKEN_ENCRYPTION_KEY — the
-- admin API never returns the decrypted value, only a masked preview.
CREATE TABLE llm_credentials (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider            TEXT NOT NULL,
    label               TEXT NOT NULL,
    api_key_ciphertext  BYTEA NOT NULL,
    api_key_nonce       BYTEA NOT NULL,
    api_key_last4       TEXT NOT NULL,
    base_url            TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX llm_credentials_provider_idx ON llm_credentials (provider);

CREATE TRIGGER llm_credentials_set_updated_at
    BEFORE UPDATE ON llm_credentials
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
