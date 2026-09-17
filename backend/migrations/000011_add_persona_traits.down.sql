DROP TABLE user_persona_traits;

ALTER TABLE personas
    DROP COLUMN default_warmth,
    DROP COLUMN default_humor,
    DROP COLUMN default_wisdom,
    DROP COLUMN default_directness,
    DROP COLUMN default_energy;
