-- Individual models available under a given credential (a single API
-- key/token can usually call several models, e.g. one OpenAI key for
-- both gpt-4.1 and gpt-4o-mini) — personas will later reference a row
-- here (see llm_model_id on personas) instead of the process-wide
-- LLM_PROVIDER/LLM_MODEL env vars.
CREATE TABLE llm_models (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id  UUID NOT NULL REFERENCES llm_credentials(id) ON DELETE CASCADE,
    model_name     TEXT NOT NULL,
    display_name   TEXT NOT NULL,
    is_default     BOOLEAN NOT NULL DEFAULT false,
    is_active      BOOLEAN NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (credential_id, model_name)
);

CREATE INDEX llm_models_credential_id_idx ON llm_models (credential_id);

CREATE TRIGGER llm_models_set_updated_at
    BEFORE UPDATE ON llm_models
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
