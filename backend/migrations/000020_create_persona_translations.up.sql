-- Per-locale display copy for a persona (name/short_description/tone_description
-- only — system_prompt and category are never translated, see the admin
-- CRUD handlers). Turkish is never stored here: it's always read straight
-- off the base personas columns, so a persona with no rows in this table
-- (or none for a given locale) simply reads as Turkish everywhere.
CREATE TABLE persona_translations (
    persona_id          UUID NOT NULL REFERENCES personas(id) ON DELETE CASCADE,
    locale               TEXT NOT NULL,
    name                 TEXT NOT NULL,
    short_description    TEXT NOT NULL,
    tone_description      TEXT NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (persona_id, locale)
);

CREATE TRIGGER persona_translations_set_updated_at
    BEFORE UPDATE ON persona_translations
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
