CREATE TABLE personas (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug               TEXT NOT NULL UNIQUE,
    name               TEXT NOT NULL,
    category           TEXT NOT NULL,
    short_description  TEXT NOT NULL,
    system_prompt      TEXT NOT NULL,
    tone_description   TEXT NOT NULL,
    avatar_url         TEXT,
    accent_color       TEXT NOT NULL,
    is_active          BOOLEAN NOT NULL DEFAULT true,
    sort_order         INTEGER NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX personas_category_idx ON personas (category);
CREATE INDEX personas_active_sort_idx ON personas (is_active, sort_order);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER personas_set_updated_at
    BEFORE UPDATE ON personas
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
