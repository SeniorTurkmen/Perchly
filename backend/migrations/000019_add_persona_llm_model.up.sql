-- Optional per-persona model override. NULL means "use the process-wide
-- LLM_PROVIDER/LLM_MODEL default" (see internal/config), so existing
-- personas keep working unchanged until an admin explicitly assigns one.
-- ON DELETE SET NULL: removing the model a persona was pointed at falls
-- it back to that same default rather than leaving a dangling reference.
ALTER TABLE personas
    ADD COLUMN llm_model_id UUID REFERENCES llm_models(id) ON DELETE SET NULL;
