-- Each persona's default personality dial values (0-100). A user's own
-- current values for a persona, if they've customized any, live in
-- user_persona_traits below and take precedence over these.
ALTER TABLE personas
    ADD COLUMN default_warmth SMALLINT NOT NULL DEFAULT 50 CHECK (default_warmth BETWEEN 0 AND 100),
    ADD COLUMN default_humor SMALLINT NOT NULL DEFAULT 50 CHECK (default_humor BETWEEN 0 AND 100),
    ADD COLUMN default_wisdom SMALLINT NOT NULL DEFAULT 50 CHECK (default_wisdom BETWEEN 0 AND 100),
    ADD COLUMN default_directness SMALLINT NOT NULL DEFAULT 50 CHECK (default_directness BETWEEN 0 AND 100),
    ADD COLUMN default_energy SMALLINT NOT NULL DEFAULT 50 CHECK (default_energy BETWEEN 0 AND 100);

UPDATE personas SET
    default_warmth = 70, default_humor = 40, default_wisdom = 60, default_directness = 75, default_energy = 80
    WHERE slug = 'motivational-coach';
UPDATE personas SET
    default_warmth = 90, default_humor = 60, default_wisdom = 50, default_directness = 30, default_energy = 60
    WHERE slug = 'daily-companion';
UPDATE personas SET
    default_warmth = 65, default_humor = 70, default_wisdom = 65, default_directness = 40, default_energy = 55
    WHERE slug = 'hobby-book-partner';

-- A user's own current dial values for one persona, adjustable any time
-- (see PUT /personas/{id}/traits) and applied to every message sent to
-- that persona from then on, until changed again. No row means the
-- user hasn't customized this persona yet — the persona's own defaults
-- above apply.
CREATE TABLE user_persona_traits (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    persona_id  UUID NOT NULL REFERENCES personas(id) ON DELETE CASCADE,
    warmth      SMALLINT NOT NULL CHECK (warmth BETWEEN 0 AND 100),
    humor       SMALLINT NOT NULL CHECK (humor BETWEEN 0 AND 100),
    wisdom      SMALLINT NOT NULL CHECK (wisdom BETWEEN 0 AND 100),
    directness  SMALLINT NOT NULL CHECK (directness BETWEEN 0 AND 100),
    energy      SMALLINT NOT NULL CHECK (energy BETWEEN 0 AND 100),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, persona_id)
);
